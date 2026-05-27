package publish

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultMaxUploadBytes int64 = 64 << 20 // 64 MiB

// PublishedPackage is a lightweight registry catalog entry.
type PublishedPackage struct {
	PackageName  string    `json:"packageName"`
	Version      string    `json:"version"`
	SectionCount int       `json:"sectionCount"`
	SlugCount    int       `json:"slugCount"`
	Path         string    `json:"path,omitempty"`
	SHA256       string    `json:"sha256,omitempty"`
	PublishedBy  string    `json:"publishedBy,omitempty"`
	PublishedAt  time.Time `json:"publishedAt,omitempty"`
}

// PackageStore persists a validated package docs DB and can list published packages.
type PackageStore interface {
	Publish(ctx context.Context, packageName, version, dbPath string, result *SQLiteValidationResult, identity *PublisherIdentity) (*PublishedPackage, error)
	List(ctx context.Context) ([]PublishedPackage, error)
}

// RegistryHandler serves the Phase 1 docs publishing registry API.
type RegistryHandler struct {
	Auth                    PublisherAuth
	Store                   PackageStore
	MaxUploadBytes          int64
	TempDir                 string
	MaxConcurrentUploads    int
	RateLimitRequestsPerMin int
	RateLimitBurst          int
	Metrics                 *RegistryMetrics

	publishSem chan struct{}
}

type registryError struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type healthResponse struct {
	OK bool `json:"ok"`
}

type listPackagesResponse struct {
	Packages []PublishedPackage `json:"packages"`
}

type publishResponse struct {
	OK      bool                    `json:"ok"`
	Package PublishedPackage        `json:"package"`
	Result  *SQLiteValidationResult `json:"validation"`
	Actor   *PublisherIdentity      `json:"actor,omitempty"`
}

func NewRegistryHandler(auth PublisherAuth, store PackageStore) *RegistryHandler {
	return &RegistryHandler{Auth: auth, Store: store, MaxUploadBytes: defaultMaxUploadBytes}
}

func (h *RegistryHandler) acquirePublishSlot() bool {
	if h.MaxConcurrentUploads <= 0 {
		return true
	}
	if h.publishSem == nil {
		return true
	}
	select {
	case h.publishSem <- struct{}{}:
		return true
	default:
		return false
	}
}

func (h *RegistryHandler) releasePublishSlot() {
	if h.MaxConcurrentUploads <= 0 || h.publishSem == nil {
		return
	}
	select {
	case <-h.publishSem:
	default:
	}
}

func (h *RegistryHandler) Handler() http.Handler {
	if h.MaxConcurrentUploads > 0 && h.publishSem == nil {
		h.publishSem = make(chan struct{}, h.MaxConcurrentUploads)
	}
	if h.Metrics == nil {
		h.Metrics = NewRegistryMetrics()
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.handleHealth)
	mux.Handle("GET /metrics", h.Metrics)
	mux.HandleFunc("GET /v1/packages", h.handleListPackages)
	mux.HandleFunc("PUT /v1/packages/{package}/versions/{version}/sqlite", h.handlePublishSQLite)

	var handler http.Handler = mux
	handler = withRateLimit(handler, NewSimpleRateLimiter(h.RateLimitRequestsPerMin, h.RateLimitBurst))
	handler = withAccessLog(handler, h.Metrics)
	handler = withRequestID(handler)
	return handler
}

func (h *RegistryHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeRegistryJSON(w, http.StatusOK, healthResponse{OK: true})
}

func (h *RegistryHandler) handleListPackages(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		writeRegistryJSON(w, http.StatusOK, listPackagesResponse{Packages: []PublishedPackage{}})
		return
	}
	packages, err := h.Store.List(r.Context())
	if err != nil {
		writeRegistryError(w, http.StatusInternalServerError, "list_failed", "failed to list packages")
		return
	}
	writeRegistryJSON(w, http.StatusOK, listPackagesResponse{Packages: packages})
}

func (h *RegistryHandler) handlePublishSQLite(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	packageName := r.PathValue("package")
	version := r.PathValue("version")
	audit := publishAuditEvent{
		RequestID:     requestIDFromContext(r.Context()),
		PackageName:   packageName,
		Version:       version,
		ContentLength: r.ContentLength,
		Status:        http.StatusInternalServerError,
		Outcome:       "failed",
		ErrorCode:     "unknown",
	}
	defer func() {
		audit.Duration = time.Since(started)
		logPublishAudit(r, audit)
		h.Metrics.RecordPublish(audit)
	}()

	if !h.acquirePublishSlot() {
		audit.Status = http.StatusTooManyRequests
		audit.Outcome = "rejected"
		audit.ErrorCode = "too_many_concurrent_uploads"
		writeRegistryError(w, http.StatusTooManyRequests, "too_many_concurrent_uploads", "too many concurrent uploads")
		return
	}
	defer h.releasePublishSlot()

	if h.Auth == nil {
		audit.Status = http.StatusServiceUnavailable
		audit.Outcome = "rejected"
		audit.ErrorCode = "auth_not_configured"
		writeRegistryError(w, http.StatusServiceUnavailable, "auth_not_configured", "publisher auth is not configured")
		return
	}
	if h.Store == nil {
		audit.Status = http.StatusServiceUnavailable
		audit.Outcome = "rejected"
		audit.ErrorCode = "store_not_configured"
		writeRegistryError(w, http.StatusServiceUnavailable, "store_not_configured", "package store is not configured")
		return
	}

	identity, err := h.Auth.AuthorizePublish(r.Context(), bearerToken(r), PublishRequest{PackageName: packageName, Version: version})
	if err != nil {
		audit.Status, audit.ErrorCode = authErrorStatusAndCode(err)
		audit.Outcome = "rejected"
		writeAuthError(w, err)
		return
	}
	audit.Identity = identity

	tmpPath, err := h.receiveUpload(r)
	if err != nil {
		var maxErr *maxUploadError
		if errors.As(err, &maxErr) {
			audit.Status = http.StatusRequestEntityTooLarge
			audit.Outcome = "rejected"
			audit.ErrorCode = "upload_too_large"
			writeRegistryError(w, http.StatusRequestEntityTooLarge, "upload_too_large", maxErr.Error())
			return
		}
		audit.Status = http.StatusBadRequest
		audit.Outcome = "rejected"
		audit.ErrorCode = "invalid_upload"
		writeRegistryError(w, http.StatusBadRequest, "invalid_upload", err.Error())
		return
	}
	defer func() { _ = os.Remove(tmpPath) }()
	if info, err := os.Stat(tmpPath); err == nil {
		audit.UploadBytes = info.Size()
	}

	result, err := ValidateSQLiteHelpDB(r.Context(), tmpPath, SQLiteValidationOptions{PackageName: packageName, Version: version})
	if err != nil {
		audit.Status = http.StatusBadRequest
		audit.Outcome = "rejected"
		audit.ErrorCode = "invalid_help_db"
		writeRegistryError(w, http.StatusBadRequest, "invalid_help_db", err.Error())
		return
	}
	audit.SectionCount = result.SectionCount
	audit.SlugCount = result.SlugCount

	published, err := h.Store.Publish(r.Context(), packageName, version, tmpPath, result, identity)
	if err != nil {
		audit.Status, audit.ErrorCode = publishErrorStatusAndCode(err)
		audit.Outcome = "rejected"
		writePublishError(w, err)
		return
	}
	audit.Status = http.StatusOK
	audit.Outcome = "success"
	audit.ErrorCode = ""
	audit.SHA256 = published.SHA256
	writeRegistryJSON(w, http.StatusOK, publishResponse{OK: true, Package: *published, Result: result, Actor: identity})
}

func (h *RegistryHandler) receiveUpload(r *http.Request) (string, error) {
	maxBytes := h.MaxUploadBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxUploadBytes
	}
	if r.ContentLength > maxBytes {
		return "", &maxUploadError{Max: maxBytes}
	}

	tmpDir := h.TempDir
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}
	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		return "", fmt.Errorf("create upload temp dir: %w", err)
	}
	file, err := os.CreateTemp(tmpDir, "docs-upload-*.db")
	if err != nil {
		return "", fmt.Errorf("create temp upload: %w", err)
	}
	path := file.Name()
	defer func() { _ = file.Close() }()

	reader := io.LimitReader(r.Body, maxBytes+1)
	written, err := io.Copy(file, reader)
	if err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("read upload body: %w", err)
	}
	if written > maxBytes {
		_ = os.Remove(path)
		return "", &maxUploadError{Max: maxBytes}
	}
	if written == 0 {
		_ = os.Remove(path)
		return "", errors.New("upload body is empty")
	}
	if err := file.Sync(); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("sync upload: %w", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return abs, nil
}

type maxUploadError struct {
	Max int64
}

func (e *maxUploadError) Error() string {
	return fmt.Sprintf("upload exceeds maximum size of %d bytes", e.Max)
}

func bearerToken(r *http.Request) string {
	value := r.Header.Get("Authorization")
	if value == "" {
		return ""
	}
	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func publishErrorStatusAndCode(err error) (int, string) {
	switch {
	case errors.Is(err, ErrVersionAlreadyExists):
		return http.StatusConflict, "version_already_exists"
	case errors.Is(err, ErrPackageQuotaExceeded):
		return http.StatusInsufficientStorage, "quota_exceeded"
	case errors.Is(err, ErrPackageVersionQuotaExceeded):
		return http.StatusConflict, "version_quota_exceeded"
	default:
		return http.StatusInternalServerError, "publish_failed"
	}
}

func writePublishError(w http.ResponseWriter, err error) {
	status, code := publishErrorStatusAndCode(err)
	if code == "publish_failed" {
		writeRegistryError(w, status, code, "failed to publish package")
		return
	}
	writeRegistryError(w, status, code, err.Error())
}

func authErrorStatusAndCode(err error) (int, string) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "forbidden"
	default:
		return http.StatusForbidden, "forbidden"
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	status, code := authErrorStatusAndCode(err)
	switch code {
	case "unauthorized":
		writeRegistryError(w, status, code, "missing or invalid publish token")
	case "forbidden":
		if errors.Is(err, ErrForbidden) {
			writeRegistryError(w, status, code, "publish token is not allowed for this package")
			return
		}
		writeRegistryError(w, status, code, err.Error())
	}
}

func writeRegistryError(w http.ResponseWriter, status int, code, message string) {
	writeRegistryJSON(w, status, registryError{Error: code, Message: message})
}

func writeRegistryJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
