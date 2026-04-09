package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// FS contains the generated Vite frontend assets copied to pkg/web/dist by
// cmd/build-web.
//
//go:embed dist
var FS embed.FS

// NewSPAHandler returns a handler that serves the embedded frontend assets with
// SPA fallback to index.html for unknown paths.
func NewSPAHandler() (http.Handler, error) {
	sub, err := fs.Sub(FS, "dist")
	if err != nil {
		return nil, err
	}

	indexBytes, err := fs.ReadFile(sub, "index.html")
	if err != nil {
		return nil, err
	}

	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		cleanPath := path.Clean("/" + r.URL.Path)
		if cleanPath == "/" {
			serveIndex(w, r, indexBytes)
			return
		}

		assetPath := strings.TrimPrefix(cleanPath, "/")
		if _, err := fs.Stat(sub, assetPath); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		serveIndex(w, r, indexBytes)
	}), nil
}

func serveIndex(w http.ResponseWriter, r *http.Request, indexBytes []byte) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write(indexBytes)
	}
}
