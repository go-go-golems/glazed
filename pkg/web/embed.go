//go:build embed

package web

import (
	"embed"
	"io/fs"
)

// embeddedFS contains generated Vite frontend assets copied to
// pkg/web/embed/public by cmd/build-web.
//
//go:embed embed/public
var embeddedFS embed.FS

// PublicFS is the generated frontend asset filesystem used by NewSPAHandler in
// production builds.
var PublicFS = mustSubFS(embeddedFS, "embed/public")

func mustSubFS(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
