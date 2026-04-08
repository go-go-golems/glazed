package server

import (
	"net/http"
)

// APIPathPrefix is the URL prefix under which the API handler is mounted.
// All API routes are relative to this prefix. The constant is exported so
// callers assembling a combined handler can prefix routes correctly.
const APIPathPrefix = "/api"

// NewCORSHandler returns a middleware that appends CORS headers to every response,
// allowing any origin (matching the design decision for a local dev tool).
//
// For production deployments that need stricter CORS, callers can wrap this
// with a custom middleware or replace it with a library such as
// github.com/rs/cors.
func NewCORSHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
