package publish

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestRegistryHealthAndListPackages(t *testing.T) {
	store := &fakePackageStore{packages: []PublishedPackage{{PackageName: "pinocchio", Version: "v1", SectionCount: 1, SlugCount: 1}}}
	h := NewRegistryHandler(newRegistryTestAuth(t), store).Handler()

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("health status = %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/packages", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", rr.Code, rr.Body.String())
	}
	var payload listPackagesResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(payload.Packages) != 1 || payload.Packages[0].PackageName != "pinocchio" {
		t.Fatalf("unexpected packages: %#v", payload.Packages)
	}
}

func TestRegistryPublishSQLiteSuccess(t *testing.T) {
	store := &fakePackageStore{}
	h := NewRegistryHandler(newRegistryTestAuth(t), store)
	h.TempDir = t.TempDir()
	server := h.Handler()

	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if store.publishCalls != 1 {
		t.Fatalf("expected 1 publish call, got %d", store.publishCalls)
	}
	if store.lastPackage != "pinocchio" || store.lastVersion != "v1" {
		t.Fatalf("unexpected package/version: %s %s", store.lastPackage, store.lastVersion)
	}
	var payload publishResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal publish response: %v", err)
	}
	if !payload.OK || payload.Package.SectionCount != 1 {
		t.Fatalf("unexpected response: %#v", payload)
	}
}

func TestRegistryPublishSQLiteForbiddenPackage(t *testing.T) {
	h := NewRegistryHandler(newRegistryTestAuth(t), &fakePackageStore{}).Handler()
	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/glazed/versions/v1/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegistryPublishAuditDoesNotLogBearerToken(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	h := NewRegistryHandler(newRegistryTestAuth(t), &fakePackageStore{}).Handler()
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader([]byte("not sqlite")))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	req.Header.Set(requestIDHeader, "audit-test-request")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	logText := logs.String()
	if !strings.Contains(logText, "docs registry publish") || !strings.Contains(logText, "audit-test-request") {
		t.Fatalf("publish audit log missing expected fields: %s", logText)
	}
	if strings.Contains(logText, "pinocchio-token") || strings.Contains(logText, "Authorization") {
		t.Fatalf("publish audit log leaked auth material: %s", logText)
	}
}

func TestRegistryPublishSQLiteInvalidDB(t *testing.T) {
	store := &fakePackageStore{}
	h := NewRegistryHandler(newRegistryTestAuth(t), store).Handler()
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader([]byte("not sqlite")))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if store.publishCalls != 0 {
		t.Fatalf("store should not be called for invalid DB")
	}
}

func TestRegistryPublishSQLiteOversized(t *testing.T) {
	store := &fakePackageStore{}
	h := NewRegistryHandler(newRegistryTestAuth(t), store)
	h.MaxUploadBytes = 4
	server := h.Handler()
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader([]byte("too large")))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if store.publishCalls != 0 {
		t.Fatalf("store should not be called for oversized upload")
	}
}

func TestRegistryPublishSQLiteUnknownLengthOversized(t *testing.T) {
	store := &fakePackageStore{}
	h := NewRegistryHandler(newRegistryTestAuth(t), store)
	h.MaxUploadBytes = 4
	server := h.Handler()
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader([]byte("too large")))
	req.ContentLength = -1
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	server.ServeHTTP(rr, req)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if store.publishCalls != 0 {
		t.Fatalf("store should not be called for oversized upload")
	}
}

func TestRegistryRequestIDHeader(t *testing.T) {
	h := NewRegistryHandler(newRegistryTestAuth(t), &fakePackageStore{}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(requestIDHeader, "test-request-id")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if got := rr.Header().Get(requestIDHeader); got != "test-request-id" {
		t.Fatalf("request id header = %q", got)
	}
}

func TestRegistryRateLimit(t *testing.T) {
	h := NewRegistryHandler(newRegistryTestAuth(t), &fakePackageStore{})
	h.RateLimitRequestsPerMin = 1
	h.RateLimitBurst = 1
	server := h.Handler()

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/packages", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	server.ServeHTTP(first, req)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d body=%s", first.Code, first.Body.String())
	}

	second := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/packages", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	server.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d body=%s", second.Code, second.Body.String())
	}
}

func TestRegistryPublishConcurrencyLimit(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	store := &fakePackageStore{entered: entered, release: release}
	h := NewRegistryHandler(newRegistryTestAuth(t), store)
	h.MaxConcurrentUploads = 1
	server := h.Handler()
	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer pinocchio-token")
		rr := httptest.NewRecorder()
		server.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Errorf("first status = %d body=%s", rr.Code, rr.Body.String())
		}
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("first publish did not enter store")
	}

	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v2/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()
	server.ServeHTTP(rr, req)
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d body=%s", rr.Code, rr.Body.String())
	}

	close(release)
	wg.Wait()
}

func TestRegistryPublishSQLiteVersionAlreadyExists(t *testing.T) {
	store := &fakePackageStore{publishErr: &VersionAlreadyExistsError{PackageName: "pinocchio", Version: "v1"}}
	h := NewRegistryHandler(newRegistryTestAuth(t), store).Handler()
	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegistryPublishSQLitePackageQuotaExceeded(t *testing.T) {
	store := &fakePackageStore{publishErr: &PackageQuotaExceededError{PackageName: "pinocchio", MaxBytes: 1, Projected: 2}}
	h := NewRegistryHandler(newRegistryTestAuth(t), store).Handler()
	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInsufficientStorage {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestRegistryPublishSQLiteStoreFailure(t *testing.T) {
	store := &fakePackageStore{publishErr: errors.New("boom")}
	h := NewRegistryHandler(newRegistryTestAuth(t), store).Handler()
	body := readFileBytes(t, createRegistryHelpDB(t, "intro"))
	req := httptest.NewRequest(http.MethodPut, "/v1/packages/pinocchio/versions/v1/sqlite", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer pinocchio-token")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
}

func newRegistryTestAuth(t *testing.T) *StaticTokenAuth {
	t.Helper()
	auth, err := NewStaticTokenAuth([]StaticPublisherToken{{PackageName: "pinocchio", TokenHash: HashPublishToken("pinocchio-token")}})
	if err != nil {
		t.Fatalf("NewStaticTokenAuth: %v", err)
	}
	return auth
}

type fakePackageStore struct {
	packages     []PublishedPackage
	publishErr   error
	publishCalls int
	lastPackage  string
	lastVersion  string
	entered      chan struct{}
	release      chan struct{}
}

func (s *fakePackageStore) Publish(ctx context.Context, packageName, version, dbPath string, result *SQLiteValidationResult, identity *PublisherIdentity) (*PublishedPackage, error) {
	if s.entered != nil {
		s.entered <- struct{}{}
	}
	if s.release != nil {
		<-s.release
	}
	if s.publishErr != nil {
		return nil, s.publishErr
	}
	s.publishCalls++
	s.lastPackage = packageName
	s.lastVersion = version
	pkg := PublishedPackage{PackageName: packageName, Version: version, SectionCount: result.SectionCount, SlugCount: result.SlugCount, PublishedAt: time.Now()}
	s.packages = append(s.packages, pkg)
	return &pkg, nil
}

func (s *fakePackageStore) List(ctx context.Context) ([]PublishedPackage, error) {
	ret := make([]PublishedPackage, len(s.packages))
	copy(ret, s.packages)
	return ret, nil
}

func createRegistryHelpDB(t *testing.T, slug string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "help.db")
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.Exec(`
		CREATE TABLE sections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			slug TEXT,
			title TEXT NOT NULL
		);
		INSERT INTO sections (slug, title) VALUES (?, 'Intro');
	`, slug)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	return path
}

func readFileBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return data
}
