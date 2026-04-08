package server

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
)

// SPAHandler returns a function that builds an http.Handler serving a React SPA
// from an embedded filesystem (typically populated by go:embed in the calling package).
//
// The returned handler tries to serve the requested path as a static file first.
// If the file is not found, it delegates to next (the API handler) — allowing
// /api/* routes to be handled there. If neither matches, it serves index.html
// so that client-side routing works.
//
// Usage:
//
//	spaHandler := server.SPAHandler(spaFS, "dist")
//	handler := spaHandler(apiHandler)  // apiHandler = server.NewHandler(deps)
//
// The indexFS subdirectory (e.g. "dist") is relative to the root of fsys.
func SPAHandler(fsys embed.FS, indexFS string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		subFS, err := fs.Sub(fsys, indexFS)
		if err != nil {
			// No valid subdirectory — fall back to API-only mode.
			return next
		}
		return &spaFileHandler{
			staticFS:   subFS,
			embeddedFS: fsys,
			indexRel:   filepath.Join(indexFS, "index.html"),
			next:       next,
		}
	}
}

type spaFileHandler struct {
	staticFS   fs.FS        // subdirectory contents (assets)
	embeddedFS embed.FS     // root filesystem (for index.html at indexRel)
	indexRel   string       // path to index.html within embeddedFS
	next       http.Handler // API handler
}

func (h *spaFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// For the root path, serve index.html directly.
	if r.URL.Path == "/" {
		h.serveIndex(w)
		return
	}

	// Try to serve as a static asset from the subdirectory.
	filePath := stripSlash(r.URL.Path)
	if _, err := fs.Stat(h.staticFS, filePath); err == nil {
		http.FileServerFS(h.staticFS).ServeHTTP(w, r)
		return
	}

	// Not a static file — delegate to next (API handler).
	// If next wrote a response, we're done; otherwise fall back to index.html.
	rw := &responseWriter{ResponseWriter: w, wroteHeader: false}
	h.next.ServeHTTP(rw, r)
	if rw.wroteHeader {
		return
	}
	h.serveIndex(w)
}

func stripSlash(p string) string {
	if len(p) > 0 && p[0] == '/' {
		return p[1:]
	}
	return p
}

func (h *spaFileHandler) serveIndex(w http.ResponseWriter) {
	data, err := h.embeddedFS.ReadFile(h.indexRel)
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// responseWriter wraps http.ResponseWriter to detect whether a handler already
// wrote a response, enabling the fallback chain to short-circuit.
type responseWriter struct {
	http.ResponseWriter
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(status)
}
