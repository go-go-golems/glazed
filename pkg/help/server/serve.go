package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	stdpath "path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog/log"
)

// DefaultAddr is the TCP address used by the serve command when no --address is supplied.
const DefaultAddr = ":8088"

// ServeCommand implements cmds.BareCommand to start the help browser HTTP server.
type ServeCommand struct {
	*cmds.CommandDescription
	helpSystem *help.HelpSystem
	spaHandler http.Handler
}

var _ cmds.BareCommand = (*ServeCommand)(nil)

// ServeSettings holds the parsed flag values for the serve command.
type ServeSettings struct {
	Address string   `glazed:"address"`
	Paths   []string `glazed:"paths"`
}

// NewServeCommand creates a BareCommand that starts the help browser HTTP server.
func NewServeCommand(hs *help.HelpSystem, spaHandler http.Handler) (*ServeCommand, error) {
	return &ServeCommand{
		CommandDescription: cmds.NewCommandDescription(
			"serve",
			cmds.WithShort("Serve help documentation as a web browser application"),
			cmds.WithLong(`Discover Glazed Markdown files from the given paths and start an HTTP
server that serves them with an optional React SPA frontend.

Paths can be individual .md files or directories. Directories are walked
recursively. When no paths are given, the server loads the built-in Glazed
documentation from the embedded filesystem.

The server listens on the address specified by --address (default :8088) and
serves:
  GET /api/*   — REST API for section listing and retrieval
  GET /*       — React SPA (serves index.html for all other paths, if configured)

The resulting handler is also mountable under prefixes such as /help or /docs
using MountPrefix or NewMountedHandler.`),
			cmds.WithFlags(
				fields.New(
					"address",
					fields.TypeString,
					fields.WithHelp("Address to listen on"),
					fields.WithDefault(DefaultAddr),
				),
			),
			cmds.WithArguments(
				fields.New(
					"paths",
					fields.TypeStringList,
					fields.WithHelp("Markdown files or directories to load (default: embedded docs)"),
				),
			),
		),
		helpSystem: hs,
		spaHandler: spaHandler,
	}, nil
}

// Run starts the HTTP server. Implements cmds.BareCommand.
func (sc *ServeCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &ServeSettings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to decode serve settings: %w", err)
	}

	hs := sc.helpSystem
	if hs.Store == nil {
		return errors.New("HelpSystem.Store is nil")
	}

	// When no paths are given, the help system was already loaded with the
	// embedded documentation (e.g. via doc.AddDocToHelpSystem). Just use that.
	if len(s.Paths) > 0 {
		if err := loadPaths(ctx, hs, s.Paths); err != nil {
			return err
		}
	}

	count, _ := hs.Store.Count(ctx)
	log.Info().Int64("sections", count).Msg("Loaded help sections")

	deps := HandlerDeps{Store: hs.Store}
	handler := NewServeHandler(deps, sc.spaHandler)
	return serveHTTP(s.Address, handler)
}

// NewServeHandler composes the API handler and optional SPA handler for use at
// the server root (/). The returned handler already includes CORS because
// NewHandler applies it internally.
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
	apiHandler := NewHandler(deps)
	if spaHandler == nil {
		return apiHandler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := stdpath.Clean("/" + r.URL.Path)
		if cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/") {
			apiHandler.ServeHTTP(w, r)
			return
		}
		spaHandler.ServeHTTP(w, r)
	})
}

// MountPrefix adapts a root-mounted handler so it can be exposed under a prefix
// such as /help or /docs in an existing HTTP server.
//
// Example:
//
//	mux := http.NewServeMux()
//	h := server.NewServeHandler(deps, spa)
//	mux.Handle("/help/", server.MountPrefix("/help", h))
//	mux.Handle("/help", server.MountPrefix("/help", h))
func MountPrefix(prefix string, h http.Handler) http.Handler {
	prefix = normalizePrefix(prefix)
	if prefix == "/" {
		return h
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := stdpath.Clean("/" + r.URL.Path)
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
func NewMountedHandler(prefix string, deps HandlerDeps, spaHandler http.Handler) http.Handler {
	return MountPrefix(prefix, NewServeHandler(deps, spaHandler))
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

	log.Info().Str("address", addr).Msg("Help browser listening")

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
		log.Info().Str("signal", sig.String()).Msg("Shutting down")
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
	prefix = stdpath.Clean(prefix)
	if prefix == "." {
		return "/"
	}
	return strings.TrimSuffix(prefix, "/")
}
