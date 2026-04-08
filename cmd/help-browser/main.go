// help-browser is a standalone HTTP server that serves Glazed help Markdown files
// as a browsable web application. It loads sections from one or more files and/or
// directories passed as positional arguments and serves them via the HTTP API
// defined in pkg/help/server.
//
// Example usage:
//
//	help-browser ./docs
//	help-browser ./docs /tmp/more-sections
//	help-browser ./docs/README.md ./examples/*.md
//
// When built with //go:embed dist (see gen.go), it also serves the React SPA
// as a fallback for non-API routes. Without the embed, only the API is available.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/server"
)

const usage = `help-browser: serve Glazed help files as a browsable web application

Usage:
  help-browser [flags] [paths...]

Paths can be files (*.md) or directories. If a directory is given, all *.md files
inside it (recursively) are loaded.

Flags:
  --address  Address to listen on (default :8088)
  --help     Show this help message
`

func main() {
	flags := flag.NewFlagSet("help-browser", flag.ContinueOnError)
	flags.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	addr := flags.String("address", server.DefaultAddr, "address to listen on")

	// Parse flags first, then use remaining args as paths.
	// This mirrors how the rest of glazed parses flags.
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	paths := flags.Args()
	if len(paths) == 0 {
		flags.Usage()
		fmt.Fprintln(os.Stderr, "error: no paths given; pass one or more files or directories")
		os.Exit(1)
	}

	logger := slog.Default()
	ctx := context.Background()

	hs := help.NewHelpSystem()
	defer func() { _ = hs.Store.Close() }()

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			logger.Warn("skipping inaccessible path", "path", path, "error", err)
			continue
		}

		if info.IsDir() {
			if err := loadDir(ctx, hs, path); err != nil {
				logger.Warn("error loading directory", "path", path, "error", err)
			}
		} else if strings.HasSuffix(strings.ToLower(path), ".md") {
			if err := loadFile(ctx, hs, path); err != nil {
				logger.Warn("error loading file", "path", path, "error", err)
			}
		} else {
			logger.Warn("skipping non-markdown file", "path", path)
		}
	}

	logger.Info("starting server", "address", *addr)
	srv := server.NewServer(hs.Store,
		server.WithAddr(*addr),
		server.WithSlogger(logger),
		server.WithSPA(staticFS),
	)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server exited", "error", err)
		os.Exit(1)
	}
}

// loadDir loads all *.md files from dir recursively into hs.
func loadDir(ctx context.Context, hs *help.HelpSystem, dir string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		return loadFile(ctx, hs, path)
	})
}

// loadFile reads a single .md file and adds it as a section to hs.
func loadFile(ctx context.Context, hs *help.HelpSystem, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	section, err := help.LoadSectionFromMarkdown(data)
	if err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}
	return hs.Store.Upsert(ctx, section.Section)
}
