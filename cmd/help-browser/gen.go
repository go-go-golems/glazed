//go:generate go run ../build-web

// Build the React frontend and embed it as a static asset FS.
//
// Run from the glazed/ repo root with:
//
//	LEFTHOOK=0 go generate ./cmd/help-browser
//
// This runs the Dagger builder in cmd/build-web/ (or falls back to a local
// pnpm build if Dagger is unavailable). The resulting web/dist/ is exported
// to cmd/help-browser/dist/ and embedded at compile time via embed.go.
package main
