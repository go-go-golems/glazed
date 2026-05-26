package publish

import "errors"

var (
	ErrVersionAlreadyExists        = errors.New("version already exists")
	ErrPackageQuotaExceeded        = errors.New("package storage quota exceeded")
	ErrPackageVersionQuotaExceeded = errors.New("package version quota exceeded")
)

type VersionAlreadyExistsError struct {
	PackageName string
	Version     string
}

func (e *VersionAlreadyExistsError) Error() string {
	return e.PackageName + "@" + e.Version + " is already published with different content"
}

func (e *VersionAlreadyExistsError) Unwrap() error { return ErrVersionAlreadyExists }

type PackageQuotaExceededError struct {
	PackageName string
	MaxBytes    int64
	Projected   int64
}

func (e *PackageQuotaExceededError) Error() string {
	return e.PackageName + " would exceed configured storage quota"
}

func (e *PackageQuotaExceededError) Unwrap() error { return ErrPackageQuotaExceeded }

type PackageVersionQuotaExceededError struct {
	PackageName string
	MaxVersions int
	Projected   int
}

func (e *PackageVersionQuotaExceededError) Error() string {
	return e.PackageName + " would exceed configured version quota"
}

func (e *PackageVersionQuotaExceededError) Unwrap() error { return ErrPackageVersionQuotaExceeded }
