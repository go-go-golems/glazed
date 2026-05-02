package publish

import "testing"

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "glazed"},
		{name: "go-go-golems"},
		{name: "package_name"},
		{name: "package.name"},
		{name: "pkg123"},
		{name: "", wantErr: true},
		{name: " package", wantErr: true},
		{name: "package ", wantErr: true},
		{name: "../package", wantErr: true},
		{name: "package/name", wantErr: true},
		{name: `package\name`, wantErr: true},
		{name: ".", wantErr: true},
		{name: "..", wantErr: true},
		{name: "-package", wantErr: true},
		{name: "package+plus", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.name)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		version string
		wantErr bool
	}{
		{version: "v1.2.3"},
		{version: "2026-05-02"},
		{version: "main"},
		{version: "v1.2.3+build.1"},
		{version: "", wantErr: true},
		{version: " version", wantErr: true},
		{version: "../v1", wantErr: true},
		{version: "v1/v2", wantErr: true},
		{version: ".", wantErr: true},
		{version: "..", wantErr: true},
		{version: "-v1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			err := ValidateVersion(tt.version)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPackageVersionDBPath(t *testing.T) {
	path, err := PackageVersionDBPath("pinocchio", "v1.2.3")
	if err != nil {
		t.Fatalf("PackageVersionDBPath: %v", err)
	}
	if path != "pinocchio/v1.2.3/pinocchio.db" {
		t.Fatalf("unexpected path: %s", path)
	}
}
