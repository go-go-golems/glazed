package publish

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	maxPackageNameLength = 128
	maxVersionLength     = 128
)

var (
	packageNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	versionPattern     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]*$`)
)

// ValidatePackageName checks whether name is safe to use as a docs package
// identity and as one path segment below the package publication root.
func ValidatePackageName(name string) error {
	return validatePathSegment("package", name, maxPackageNameLength, packageNamePattern)
}

// ValidateVersion checks whether version is safe to use as a package version
// identity and as one path segment below the package publication root.
func ValidateVersion(version string) error {
	return validatePathSegment("version", version, maxVersionLength, versionPattern)
}

// ValidatePackageVersion validates the package/version pair used by publish
// APIs and filesystem materialization.
func ValidatePackageVersion(packageName, version string) error {
	if err := ValidatePackageName(packageName); err != nil {
		return err
	}
	if err := ValidateVersion(version); err != nil {
		return err
	}
	return nil
}

// DBFileName returns the canonical SQLite export file name for packageName.
func DBFileName(packageName string) (string, error) {
	if err := ValidatePackageName(packageName); err != nil {
		return "", err
	}
	return packageName + ".db", nil
}

// PackageVersionDir returns the relative directory for a package/version pair.
func PackageVersionDir(packageName, version string) (string, error) {
	if err := ValidatePackageVersion(packageName, version); err != nil {
		return "", err
	}
	return filepath.Join(packageName, version), nil
}

// PackageVersionDBPath returns the relative canonical DB path for a package
// version, e.g. pinocchio/v1.2.3/pinocchio.db.
func PackageVersionDBPath(packageName, version string) (string, error) {
	dir, err := PackageVersionDir(packageName, version)
	if err != nil {
		return "", err
	}
	fileName, err := DBFileName(packageName)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, fileName), nil
}

func validatePathSegment(kind, value string, maxLength int, pattern *regexp.Regexp) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", kind)
	}
	if strings.TrimSpace(value) != value {
		return fmt.Errorf("%s must not have leading or trailing whitespace", kind)
	}
	if len(value) > maxLength {
		return fmt.Errorf("%s %q is too long: max %d bytes", kind, value, maxLength)
	}
	if value == "." || value == ".." {
		return fmt.Errorf("%s must not be %q", kind, value)
	}
	if strings.ContainsAny(value, `/\\`) {
		return fmt.Errorf("%s %q must not contain path separators", kind, value)
	}
	if !pattern.MatchString(value) {
		return fmt.Errorf("%s %q must start with an alphanumeric character and contain only alphanumerics, dot, underscore, dash%s", kind, value, plusSuffix(kind))
	}
	return nil
}

func plusSuffix(kind string) string {
	if kind == "version" {
		return ", or plus"
	}
	return ""
}
