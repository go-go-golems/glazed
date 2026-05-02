// build-web is a Dagger-based build tool for the React frontend.
// It runs inside a container when Dagger is available and falls back to local
// pnpm when requested or when the Dagger engine is unavailable.
//
// The program walks up from the current working directory to find the repo root
// (by locating go.mod), builds web/ with pnpm, and copies the dist/ output to
// pkg/web/embed/public/ for embedding in production builds via `go build -tags embed`.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

const defaultPNPMVersion = "10.15.0"

func main() {
	builderImage := getenv("WEB_BUILDER_IMAGE", "node:22")

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}
	repoRoot, err := findRepoRoot(wd)
	if err != nil {
		log.Fatalf("find repo root: %v", err)
	}

	webPath := filepath.Join(repoRoot, "web")
	outPath := filepath.Join(repoRoot, "pkg", "web", "embed", "public")
	pnpmVersion := getenv("WEB_PNPM_VERSION", packageManagerPNPMVersion(webPath, defaultPNPMVersion))

	log.Printf("repo root: %s", repoRoot)
	log.Printf("web source: %s", webPath)
	log.Printf("build output: %s", outPath)
	log.Printf("pnpm version: %s", pnpmVersion)

	if getenv("BUILD_WEB_LOCAL", "") == "1" {
		if err := buildLocal(webPath, outPath); err != nil {
			log.Fatalf("local build failed: %v", err)
		}
		log.Printf("exported web dist to %s", outPath)
		return
	}

	if err := buildWithDagger(webPath, outPath, pnpmVersion, builderImage); err != nil {
		log.Printf("Dagger build failed (%v), falling back to local pnpm", err)
		if err := buildLocal(webPath, outPath); err != nil {
			log.Fatalf("local build also failed: %v", err)
		}
	}
	log.Printf("exported web dist to %s", outPath)
}

func buildWithDagger(webPath, outPath, pnpmVersion, builderImage string) error {
	ctx := context.Background()
	client, err := dagger.Connect(ctx)
	if err != nil {
		return fmt.Errorf("connect dagger: %v", err)
	}
	defer func() { _ = client.Close() }()

	storeCache := client.CacheVolume("glazed-help-browser-pnpm-store")
	webDir := client.Host().Directory(webPath)

	ctr := client.Container().From(builderImage).
		WithWorkdir("/src").
		WithMountedDirectory("/src", webDir).
		WithMountedCache("/pnpm/store", storeCache).
		WithEnvVariable("PNPM_HOME", "/pnpm").
		WithEnvVariable("COREPACK_ENABLE_DOWNLOAD_PROMPT", "0")

	ctr = ctr.WithExec([]string{
		"sh", "-lc",
		fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate && pnpm config set store-dir /pnpm/store", pnpmVersion),
	})

	ctr = ctr.
		WithExec([]string{"sh", "-lc", "pnpm --version"}).
		WithExec([]string{"sh", "-lc", "pnpm install --frozen-lockfile --reporter=append-only"}).
		WithExec([]string{"sh", "-lc", "pnpm build"})

	if err := os.RemoveAll(outPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove old embedded assets: %v", err)
	}
	dist := ctr.Directory("/src/dist")
	if _, err := dist.Export(ctx, outPath); err != nil {
		return fmt.Errorf("export dist to %s: %v", outPath, err)
	}
	return nil
}

func buildLocal(webPath, outPath string) error {
	if _, err := exec.LookPath("pnpm"); err != nil {
		return fmt.Errorf("pnpm not found in PATH: %v", err)
	}

	if err := os.RemoveAll(outPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove embedded assets: %v", err)
	}

	for _, cmdStr := range []string{"pnpm install --frozen-lockfile", "pnpm build"} {
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Dir = webPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %v", cmdStr, err)
		}
	}

	srcDist := filepath.Join(webPath, "dist")
	if err := copyDir(srcDist, outPath); err != nil {
		return fmt.Errorf("copy dist to %s: %v", outPath, err)
	}
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		// #nosec G122 -- src is a repo-local build output directory created by this tool; we intentionally walk and copy it.
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		// #nosec G703 -- dstPath is derived from filepath.Rel(src, path) under the same controlled copy operation.
		return os.WriteFile(dstPath, data, info.Mode())
	})
}

func findRepoRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", start)
		}
		dir = parent
	}
}

func packageManagerPNPMVersion(webPath, fallback string) string {
	data, err := os.ReadFile(filepath.Join(webPath, "package.json"))
	if err != nil {
		return fallback
	}
	var pkg struct {
		PackageManager string `json:"packageManager"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fallback
	}
	pm := strings.TrimSpace(pkg.PackageManager)
	if version, ok := strings.CutPrefix(pm, "pnpm@"); ok && strings.TrimSpace(version) != "" {
		return strings.TrimSpace(version)
	}
	return fallback
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
