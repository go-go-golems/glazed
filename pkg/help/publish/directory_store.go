package publish

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const CatalogFileName = "catalog.json"

type DirectoryPackageStore struct {
	Root string
	Now  func() time.Time
}

type PackageCatalog struct {
	UpdatedAt time.Time          `json:"updatedAt"`
	Packages  []PublishedPackage `json:"packages"`
}

func NewDirectoryPackageStore(root string) *DirectoryPackageStore {
	return &DirectoryPackageStore{Root: root}
}

func (s *DirectoryPackageStore) Publish(ctx context.Context, packageName, version, dbPath string, result *SQLiteValidationResult, identity *PublisherIdentity) (*PublishedPackage, error) {
	if err := ValidatePackageVersion(packageName, version); err != nil {
		return nil, err
	}
	if s.Root == "" {
		return nil, fmt.Errorf("package root must not be empty")
	}
	rel, err := PackageVersionDBPath(packageName, version)
	if err != nil {
		return nil, err
	}
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return nil, err
	}
	target := filepath.Join(root, rel)
	if !isPathUnderRoot(root, target) {
		return nil, fmt.Errorf("target path escapes package root")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return nil, fmt.Errorf("create package directory: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(target), ".upload-*.db")
	if err != nil {
		return nil, fmt.Errorf("create temp package db: %w", err)
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			_ = os.Remove(tmpPath)
		}
	}()

	src, err := os.Open(dbPath)
	if err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("open validated db: %w", err)
	}
	h := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(tmp, h), src)
	_ = src.Close()
	if copyErr != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("copy package db: %w", copyErr)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return nil, fmt.Errorf("sync package db: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return nil, fmt.Errorf("close package db: %w", err)
	}
	if err := os.Rename(tmpPath, target); err != nil {
		return nil, fmt.Errorf("publish package db: %w", err)
	}
	removeTmp = false
	_ = syncDir(filepath.Dir(target))

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	published := PublishedPackage{PackageName: packageName, Version: version, SectionCount: result.SectionCount, SlugCount: result.SlugCount, PublishedAt: now, Path: filepath.ToSlash(rel), SHA256: hex.EncodeToString(h.Sum(nil))}
	if identity != nil {
		published.PublishedBy = identity.Subject
	}
	if err := s.upsertCatalog(ctx, root, published); err != nil {
		return nil, err
	}
	return &published, nil
}

func (s *DirectoryPackageStore) List(ctx context.Context) ([]PublishedPackage, error) {
	root, err := filepath.Abs(s.Root)
	if err != nil {
		return nil, err
	}
	catalog, err := readCatalog(root)
	if os.IsNotExist(err) {
		return []PublishedPackage{}, nil
	}
	if err != nil {
		return nil, err
	}
	return catalog.Packages, nil
}

func (s *DirectoryPackageStore) upsertCatalog(ctx context.Context, root string, pkg PublishedPackage) error {
	catalog, err := readCatalog(root)
	if os.IsNotExist(err) {
		catalog = PackageCatalog{}
	} else if err != nil {
		return err
	}
	found := false
	for i := range catalog.Packages {
		if catalog.Packages[i].PackageName == pkg.PackageName && catalog.Packages[i].Version == pkg.Version {
			catalog.Packages[i] = pkg
			found = true
			break
		}
	}
	if !found {
		catalog.Packages = append(catalog.Packages, pkg)
	}
	sort.Slice(catalog.Packages, func(i, j int) bool {
		if catalog.Packages[i].PackageName != catalog.Packages[j].PackageName {
			return catalog.Packages[i].PackageName < catalog.Packages[j].PackageName
		}
		return catalog.Packages[i].Version > catalog.Packages[j].Version
	})
	catalog.UpdatedAt = pkg.PublishedAt
	return writeCatalog(root, catalog)
}

func readCatalog(root string) (PackageCatalog, error) {
	var catalog PackageCatalog
	data, err := os.ReadFile(filepath.Join(root, CatalogFileName))
	if err != nil {
		return catalog, err
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		return catalog, fmt.Errorf("decode catalog: %w", err)
	}
	return catalog, nil
}

func writeCatalog(root string, catalog PackageCatalog) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create package root: %w", err)
	}
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(root, ".catalog-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	remove := true
	defer func() {
		if remove {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, filepath.Join(root, CatalogFileName)); err != nil {
		return err
	}
	remove = false
	_ = syncDir(root)
	return nil
}

func isPathUnderRoot(root, target string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != "." && rel != "" && !startsWithDotDot(rel)
}
func startsWithDotDot(rel string) bool { return rel == ".." || len(rel) > 3 && rel[:3] == "../" }
func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Sync()
}
