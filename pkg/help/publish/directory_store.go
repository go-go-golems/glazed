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
	"strings"
	"time"
)

const CatalogFileName = "catalog.json"

type DirectoryPackageStore struct {
	Root                  string
	Now                   func() time.Time
	AllowOverwrite        bool
	MaxPackageBytes       int64
	MaxVersionsPerPackage int
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
	targetDir := filepath.Dir(target)
	if !isPathUnderRoot(root, targetDir) {
		return nil, fmt.Errorf("target directory escapes package root")
	}
	srcInfo, err := os.Stat(dbPath)
	if err != nil {
		return nil, fmt.Errorf("stat validated db: %w", err)
	}
	// #nosec G703 -- target is derived from validated package/version path segments and checked to remain under the configured package root.
	existingInfo, existingErr := os.Stat(target)
	if existingErr == nil {
		if same, err := sameFileSHA256(dbPath, target); err != nil {
			return nil, err
		} else if same {
			return s.publishedPackageForExisting(ctx, root, rel, packageName, version, result, identity)
		}
		if !s.AllowOverwrite {
			return nil, &VersionAlreadyExistsError{PackageName: packageName, Version: version}
		}
	} else if !os.IsNotExist(existingErr) {
		return nil, fmt.Errorf("stat existing package db: %w", existingErr)
	}
	if err := s.checkQuota(root, packageName, version, srcInfo.Size(), existingInfo, existingErr == nil); err != nil {
		return nil, err
	}

	// #nosec G703 -- targetDir is derived from validated package/version path segments and checked to remain under the configured package root.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("create package directory: %w", err)
	}

	tmp, err := os.CreateTemp(targetDir, ".upload-*.db")
	if err != nil {
		return nil, fmt.Errorf("create temp package db: %w", err)
	}
	tmpPath := tmp.Name()
	removeTmp := true
	defer func() {
		if removeTmp {
			// #nosec G703 -- tmpPath is returned by os.CreateTemp in targetDir, which is checked to remain under the package root.
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
	// #nosec G703 -- tmpPath and target are both constrained to the validated package/version directory under the package root.
	if err := os.Rename(tmpPath, target); err != nil {
		return nil, fmt.Errorf("publish package db: %w", err)
	}
	removeTmp = false
	_ = syncDir(targetDir)

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

func (s *DirectoryPackageStore) publishedPackageForExisting(ctx context.Context, root, rel, packageName, version string, result *SQLiteValidationResult, identity *PublisherIdentity) (*PublishedPackage, error) {
	catalog, err := readCatalog(root)
	if err == nil {
		for i := range catalog.Packages {
			if catalog.Packages[i].PackageName == packageName && catalog.Packages[i].Version == version {
				return &catalog.Packages[i], nil
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	target := filepath.Join(root, rel)
	sha, err := fileSHA256(target)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	published := PublishedPackage{PackageName: packageName, Version: version, SectionCount: result.SectionCount, SlugCount: result.SlugCount, PublishedAt: now, Path: filepath.ToSlash(rel), SHA256: sha}
	if identity != nil {
		published.PublishedBy = identity.Subject
	}
	if err := s.upsertCatalog(ctx, root, published); err != nil {
		return nil, err
	}
	return &published, nil
}

func (s *DirectoryPackageStore) checkQuota(root, packageName, version string, newSize int64, existingInfo os.FileInfo, replacing bool) error {
	if s.MaxPackageBytes <= 0 && s.MaxVersionsPerPackage <= 0 {
		return nil
	}
	usage, err := scanPackageUsage(root, packageName)
	if err != nil {
		return err
	}
	projectedBytes := usage.totalBytes + newSize
	if replacing && existingInfo != nil {
		projectedBytes -= existingInfo.Size()
	}
	if s.MaxPackageBytes > 0 && projectedBytes > s.MaxPackageBytes {
		return &PackageQuotaExceededError{PackageName: packageName, MaxBytes: s.MaxPackageBytes, Projected: projectedBytes}
	}
	projectedVersions := usage.versionCount
	if !usage.versions[version] {
		projectedVersions++
	}
	if s.MaxVersionsPerPackage > 0 && projectedVersions > s.MaxVersionsPerPackage {
		return &PackageVersionQuotaExceededError{PackageName: packageName, MaxVersions: s.MaxVersionsPerPackage, Projected: projectedVersions}
	}
	return nil
}

type packageUsage struct {
	totalBytes   int64
	versionCount int
	versions     map[string]bool
}

func scanPackageUsage(root, packageName string) (packageUsage, error) {
	usage := packageUsage{versions: map[string]bool{}}
	packageDir := filepath.Join(root, packageName)
	// #nosec G703 -- packageDir is derived from the configured root and a validated package name.
	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return usage, nil
	} else if err != nil {
		return usage, err
	}
	// #nosec G703 -- packageDir is derived from the configured root and a validated package name.
	err := filepath.WalkDir(packageDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == packageDir {
			return nil
		}
		rel, err := filepath.Rel(packageDir, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) > 0 && parts[0] != "." && parts[0] != "" {
			usage.versions[parts[0]] = true
		}
		if d.Type().IsRegular() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			usage.totalBytes += info.Size()
		}
		return nil
	})
	if err != nil {
		return usage, err
	}
	usage.versionCount = len(usage.versions)
	return usage, nil
}

func sameFileSHA256(a, b string) (bool, error) {
	aSHA, err := fileSHA256(a)
	if err != nil {
		return false, err
	}
	bSHA, err := fileSHA256(b)
	if err != nil {
		return false, err
	}
	return aSHA == bSHA, nil
}

func fileSHA256(path string) (string, error) {
	// #nosec G304,G703 -- path is supplied by callers after validation or from a trusted temporary upload path.
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
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
	// #nosec G703 -- callers pass directories derived from configured package roots and validated paths.
	f, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return f.Sync()
}
