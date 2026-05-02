package publish

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
}

func (s *fakePackageStore) Publish(ctx context.Context, packageName, version, dbPath string, result *SQLiteValidationResult, identity *PublisherIdentity) (*PublishedPackage, error) {
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
