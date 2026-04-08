package main

import (
	"embed"
	"io/fs"
)

// staticFS embeds the Vite build output produced by go:generate (cmd/build-web).
// The directory "dist" must exist at package build time (after running go generate).
//
//go:embed dist
var staticFS embed.FS

// verify that dist is a valid subdirectory of staticFS at compile time.
var _ fs.FS = staticFS
