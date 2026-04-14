package config

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanResolvesInLayerOrderAndDedupe(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()

	a := filepath.Join(tmp, "a.yaml")
	b := filepath.Join(tmp, "b.yaml")
	c := filepath.Join(tmp, "c.yaml")
	for _, p := range []string{a, b, c} {
		if err := os.WriteFile(p, []byte("x: 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	plan := NewPlan(
		WithLayerOrder(LayerSystem, LayerUser, LayerRepo, LayerCWD, LayerExplicit),
		WithDedupePaths(),
	).Add(
		SourceSpec{
			Name:  "cwd",
			Layer: LayerCWD,
			Discover: func(context.Context) ([]string, error) {
				return []string{b, c}, nil
			},
		},
		SourceSpec{
			Name:  "explicit",
			Layer: LayerExplicit,
			Discover: func(context.Context) ([]string, error) {
				return []string{b}, nil
			},
		},
		SourceSpec{
			Name:  "repo",
			Layer: LayerRepo,
			Discover: func(context.Context) ([]string, error) {
				return []string{b, a}, nil
			},
		},
	)

	files, report, err := plan.Resolve(ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	got := report.Paths()
	want := []string{normalizePath(b), normalizePath(a), normalizePath(c)}
	if len(got) != len(want) {
		t.Fatalf("paths len mismatch: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("path[%d] mismatch: got=%q want=%q", i, got[i], want[i])
		}
		if files[i].Index != i {
			t.Fatalf("index[%d] mismatch: got=%d want=%d", i, files[i].Index, i)
		}
	}

	if len(report.Sources) != 3 {
		t.Fatalf("expected 3 resolved sources, got %d", len(report.Sources))
	}
	if report.Sources[0].Name != "repo" {
		t.Fatalf("expected repo source first after ordering, got %q", report.Sources[0].Name)
	}
	if report.Sources[1].Name != "cwd" {
		t.Fatalf("expected cwd source second after ordering, got %q", report.Sources[1].Name)
	}
	if len(report.Sources[1].DedupedPaths) != 1 || report.Sources[1].DedupedPaths[0] != normalizePath(b) {
		t.Fatalf("expected cwd source to dedupe %q, got %#v", normalizePath(b), report.Sources[1].DedupedPaths)
	}
}

func TestWorkingDirFileFindsLocalFile(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	file := filepath.Join(tmp, ".pinocchio-profile.yml")
	if err := os.WriteFile(file, []byte("profile: local\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldGetwd := getwdFunc
	getwdFunc = func() (string, error) { return tmp, nil }
	t.Cleanup(func() { getwdFunc = oldGetwd })

	src := WorkingDirFile(".pinocchio-profile.yml")
	files, err := src.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(files) != 1 || normalizePath(files[0]) != normalizePath(file) {
		t.Fatalf("working dir file mismatch: got=%v want=%q", files, normalizePath(file))
	}
}

func TestGitRootFileFindsRepoRootFileFromNestedDirectory(t *testing.T) {
	ctx := context.Background()
	repo := t.TempDir()
	rootFile := filepath.Join(repo, ".pinocchio-profile.yml")
	if err := os.WriteFile(rootFile, []byte("profile: repo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldGitRoot := gitRootFunc
	gitRootFunc = func(context.Context) (string, error) { return repo, nil }
	t.Cleanup(func() { gitRootFunc = oldGitRoot })

	src := GitRootFile(".pinocchio-profile.yml")
	files, err := src.Discover(ctx)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(files) != 1 || normalizePath(files[0]) != normalizePath(rootFile) {
		t.Fatalf("git root file mismatch: got=%v want=%q", files, normalizePath(rootFile))
	}
}

func TestXDGAndHomeAppConfigDiscoverFiles(t *testing.T) {
	ctx := context.Background()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	xdg := filepath.Join(tmp, "xdg")
	if err := os.MkdirAll(filepath.Join(home, ".pinocchio"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(xdg, "pinocchio"), 0o755); err != nil {
		t.Fatal(err)
	}
	homeFile := filepath.Join(home, ".pinocchio", "config.yaml")
	xdgFile := filepath.Join(xdg, "pinocchio", "config.yaml")
	for _, p := range []string{homeFile, xdgFile} {
		if err := os.WriteFile(p, []byte("x: 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	oldHomeDir := userHomeDirFunc
	oldConfigDir := userConfigDirFunc
	userHomeDirFunc = func() (string, error) { return home, nil }
	userConfigDirFunc = func() (string, error) { return xdg, nil }
	t.Cleanup(func() {
		userHomeDirFunc = oldHomeDir
		userConfigDirFunc = oldConfigDir
	})

	homeFiles, err := HomeAppConfig("pinocchio").Discover(ctx)
	if err != nil {
		t.Fatalf("HomeAppConfig discover failed: %v", err)
	}
	if len(homeFiles) != 1 || normalizePath(homeFiles[0]) != normalizePath(homeFile) {
		t.Fatalf("home config mismatch: got=%v want=%q", homeFiles, normalizePath(homeFile))
	}

	xdgFiles, err := XDGAppConfig("pinocchio").Discover(ctx)
	if err != nil {
		t.Fatalf("XDGAppConfig discover failed: %v", err)
	}
	if len(xdgFiles) != 1 || normalizePath(xdgFiles[0]) != normalizePath(xdgFile) {
		t.Fatalf("xdg config mismatch: got=%v want=%q", xdgFiles, normalizePath(xdgFile))
	}
}

func TestPlanReportString(t *testing.T) {
	report := &PlanReport{
		Sources: []ResolvedSource{{
			Name:       "repo",
			Layer:      LayerRepo,
			AddedPaths: []string{"/tmp/repo/.pinocchio-profile.yml"},
		}, {
			Name:          "cwd",
			Layer:         LayerCWD,
			DedupedPaths:  []string{"/tmp/repo/.pinocchio-profile.yml"},
			SkippedReason: "all discovered paths were duplicates",
		}},
	}
	out := report.String()
	for _, needle := range []string{"Config resolution plan:", "repo", "layer=repo", "cwd", "deduped"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("report missing %q: %s", needle, out)
		}
	}
}
