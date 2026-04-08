package server

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

// NewServeCommand returns a Cobra command that starts the help browser HTTP server.
// It discovers Glazed Markdown files from the given file/directory arguments and
// serves them over HTTP with an optional embedded React SPA.
func NewServeCommand(hs *help.HelpSystem, embedFS embed.FS) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve [flags] <path> [<path>...]",
		Short: "Serve help documentation as a web browser application",
		Long: `Discover Glazed Markdown files from the given paths and start an HTTP
server that serves them with an optional React SPA frontend.

Paths can be individual .md files or directories. Directories are walked
recursively.

The server listens on the address specified by --address (default :8088) and
serves:
  GET /api/*   — REST API for section listing and retrieval
  GET /*       — React SPA (serves index.html for all other paths, if configured)

The resulting handler is also mountable under prefixes such as /help or /docs
using MountPrefix or NewMountedHandler.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServe(cmd, args, hs, embedFS)
		},
	}

	cmd.Flags().String("address", DefaultAddr, "Address to listen on")
	return cmd
}

// NewServeHandler composes the API handler and optional SPA handler for use at
// the server root (/). The returned handler already includes CORS because
// NewHandler applies it internally.
func NewServeHandler(deps HandlerDeps, embedFS embed.FS) http.Handler {
	h := NewHandler(deps)
	if embedFS != (embed.FS{}) {
		h = SPAHandler(embedFS, "dist")(h)
	}
	return h
}

// MountPrefix adapts a root-mounted handler so it can be exposed under a prefix
// such as /help or /docs in an existing HTTP server.
//
// Example:
//
//	mux := http.NewServeMux()
//	h := server.NewServeHandler(deps, web.FS)
//	mux.Handle("/help/", server.MountPrefix("/help", h))
//	mux.Handle("/help", server.MountPrefix("/help", h))
func MountPrefix(prefix string, h http.Handler) http.Handler {
	prefix = normalizePrefix(prefix)
	if prefix == "/" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := path.Clean("/" + r.URL.Path)
		if cleanPath != prefix && !strings.HasPrefix(cleanPath, prefix+"/") {
			http.NotFound(w, r)
			return
		}

		r2 := r.Clone(r.Context())
		urlCopy := *r.URL
		r2.URL = &urlCopy
		r2.URL.Path = strings.TrimPrefix(cleanPath, prefix)
		if r2.URL.Path == "" {
			r2.URL.Path = "/"
		}
		h.ServeHTTP(w, r2)
	})
}

// NewMountedHandler builds a root handler and adapts it for mounting under a
// prefix in an existing HTTP server.
func NewMountedHandler(prefix string, deps HandlerDeps, embedFS embed.FS) http.Handler {
	return MountPrefix(prefix, NewServeHandler(deps, embedFS))
}

func runServe(cmd *cobra.Command, args []string, hs *help.HelpSystem, embedFS embed.FS) error {
	addr, err := cmd.Flags().GetString("address")
	if err != nil {
		return fmt.Errorf("address flag: %w", err)
	}
	if hs.Store == nil {
		return errors.New("HelpSystem.Store is nil")
	}

	ctx := context.Background()
	if err := loadPaths(ctx, hs, args); err != nil {
		return err
	}

	deps := HandlerDeps{Store: hs.Store}
	handler := NewServeHandler(deps, embedFS)
	return serveHTTP(addr, handler)
}

func loadPaths(ctx context.Context, hs *help.HelpSystem, paths []string) error {
	for _, input := range paths {
		info, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("stat %q: %w", input, err)
		}
		if info.IsDir() {
			if err := loadDir(ctx, hs, input); err != nil {
				return fmt.Errorf("loading directory %q: %w", input, err)
			}
			continue
		}
		if err := loadFile(ctx, hs, input); err != nil {
			return fmt.Errorf("loading file %q: %w", input, err)
		}
	}
	return nil
}

func loadDir(ctx context.Context, hs *help.HelpSystem, dir string) error {
	return filepath.WalkDir(dir, func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if name == "readme.md" || !strings.HasSuffix(name, ".md") {
			return nil
		}
		return loadFile(ctx, hs, filePath)
	})
}

func loadFile(ctx context.Context, hs *help.HelpSystem, filePath string) error {
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		return nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	section, err := help.LoadSectionFromMarkdown(data)
	if err != nil {
		return err
	}
	return hs.Store.Upsert(ctx, section.Section)
}

func serveHTTP(addr string, handler http.Handler) error {
	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpSrv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigCh:
		_, _ = fmt.Fprintf(os.Stderr, "received %v, shutting down...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(ctx)
	}
}

func normalizePrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = path.Clean(prefix)
	if prefix == "." {
		return "/"
	}
	return strings.TrimSuffix(prefix, "/")
}
