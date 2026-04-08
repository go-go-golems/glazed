//go:generate go run ../../cmd/build-web

package web

// Build the React frontend and embed it as a static asset FS.
//
// Run from the glazed/ repo root with:
//
//	GOWORK=off go generate ./pkg/web
//
// This runs the Dagger builder in cmd/build-web/ (or falls back to a local
// pnpm build if Dagger is unavailable). The resulting web/dist/ is copied
// to pkg/web/dist/ and embedded by this package via //go:embed.
