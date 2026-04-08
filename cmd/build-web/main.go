// build-web is a Dagger-based build tool for the React frontend.
// It runs entirely inside a container and does not require Node.js to be installed locally.
//
// The program walks up from the current working directory to find the repo root
// (by locating go.mod), builds web/ with pnpm inside a node:22 container,
// and copies the dist/ output to pkg/web/dist/ for embedding via //go:embed
// in pkg/web. Generation is triggered by `go generate ./pkg/web`.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"dagger.io/dagger"
)

func main() {
	pnpmVersion := getenv("WEB_PNPM_VERSION", "10.15.0")
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
	outPath := filepath.Join(repoRoot, "pkg", "web", "dist")

	log.Printf("repo root: %s", repoRoot)
	log.Printf("web source: %s", webPath)
	log.Printf("build output: %s", outPath)

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

	base := client.Container().From(builderImage)
	webDir := client.Host().Directory(webPath)

	ctr := base.
		WithWorkdir("/src").
		WithMountedDirectory("/src", webDir).
		WithEnvVariable("PNPM_HOME", "/pnpm")

	ctr = ctr.WithExec([]string{
		"sh", "-lc",
		fmt.Sprintf("corepack enable && corepack prepare pnpm@%s --activate", pnpmVersion),
	})

	ctr = ctr.
		WithExec([]string{"sh", "-lc", "pnpm --version"}).
		WithExec([]string{"sh", "-lc", "yes | pnpm install --reporter=append-only"}).
		WithExec([]string{"sh", "-lc", "pnpm build"})

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
		return fmt.Errorf("remove dist: %v", err)
	}

	for _, cmdStr := range []string{"yes | pnpm install", "pnpm build"} {
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
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
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

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
