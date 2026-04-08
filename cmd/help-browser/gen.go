//go:generate go run ../build-web

package main

// Build the React frontend and embed it as a static asset FS.
//
// Run from the glazed/ repo root with:
//
//	LEFTHOOK=0 go generate ./cmd/help-browser
//
// This runs the Dagger builder in cmd/build-web/ (or falls back to a local
// pnpm build if Dagger is unavailable). The resulting web/dist/ is copied
// to pkg/web/dist/ and embedded by pkg/web via //go:embed.
