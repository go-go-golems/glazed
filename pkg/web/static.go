package web

import "embed"

// FS contains the generated Vite frontend assets copied to pkg/web/dist by
// cmd/build-web.
//
//go:embed dist
var FS embed.FS
