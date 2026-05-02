//go:build !embed

package web

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing/fstest"
)

// PublicFS serves frontend assets in development and tests when the binary was
// not built with -tags embed. It prefers generated files on disk and falls back
// to a tiny placeholder SPA so ordinary `go test`/`go build` do not require
// Node, Dagger, or committed generated assets.
var PublicFS = findPublicFS()

func findPublicFS() fs.FS {
	if root, err := findRepoRootForWeb(); err == nil {
		for _, rel := range []string{
			filepath.Join("pkg", "web", "embed", "public"),
			filepath.Join("web", "dist"),
		} {
			dir := filepath.Join(root, rel)
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				return os.DirFS(dir)
			}
		}
	}
	return fallbackPublicFS()
}

func findRepoRootForWeb() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", wd)
		}
	}
}

func fallbackPublicFS() fs.FS {
	return fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<!doctype html>
<html lang="en">
  <head><meta charset="UTF-8" /><title>Glazed Help Browser</title></head>
  <body><div id="root">Glazed Help Browser assets have not been generated. Run go generate ./pkg/web for the full browser.</div></body>
</html>
`)},
		"site-config.js": &fstest.MapFile{Data: []byte(`window.__GLAZED_HELP_CONFIG__ = {};`)},
	}
}
