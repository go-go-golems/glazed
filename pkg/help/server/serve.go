package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	stdpath "path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	helploader "github.com/go-go-golems/glazed/pkg/help/loader"
	"github.com/go-go-golems/glazed/pkg/help/store"
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
	Address        string   `glazed:"address"`
	Paths          []string `glazed:"paths"`
	FromJSON       []string `glazed:"from-json"`
	FromSQLite     []string `glazed:"from-sqlite"`
	FromSQLiteDir  []string `glazed:"from-sqlite-dir"`
	FromGlazedCmd  []string `glazed:"from-glazed-cmd"`
	WithEmbedded   bool     `glazed:"with-embedded"`
	ReloadInterval string   `glazed:"reload-interval"`
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
recursively. When no external sources are given, the server serves the built-in
Glazed documentation already loaded into the help system. When one or more
external sources are given, the serve command clears any preloaded sections by
default and serves only the sections discovered from those explicit sources.
Use --with-embedded to merge external sources with the built-in documentation.

External sources can be JSON exports, SQLite exports, or Glazed-compatible
binaries loaded through --from-glazed-cmd. For example:
  glaze serve --from-glazed-cmd pinocchio,sqleton
  glaze serve --from-json ./help.json --from-sqlite ./help.db
  glaze serve --from-sqlite-dir ./help-dbs

--from-sqlite-dir scans recursively for package/version layouts:
  X.db       -> package X, no version
  X/X.db     -> package X, no version
  X/Y/X.db   -> package X, version Y

Use --reload-interval with external sources to periodically reload them. This is
intended for directory-backed deployments such as docs.yolo.scapegoat.dev where
a registry process writes new SQLite package versions into a shared directory.
Example: --reload-interval 30s.

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
				fields.New(
					"from-json",
					fields.TypeStringList,
					fields.WithHelp("JSON help export files to load; use - for stdin"),
				),
				fields.New(
					"from-sqlite",
					fields.TypeStringList,
					fields.WithHelp("SQLite help export databases to load"),
				),
				fields.New(
					"from-sqlite-dir",
					fields.TypeStringList,
					fields.WithHelp("Directories to recursively scan for X.db, X/X.db, and X/Y/X.db SQLite help exports"),
				),
				fields.New(
					"from-glazed-cmd",
					fields.TypeStringList,
					fields.WithHelp("Glazed binaries to load by running '<binary> help export --output json'"),
				),
				fields.New(
					"with-embedded",
					fields.TypeBool,
					fields.WithHelp("Include embedded docs when external sources are provided"),
					fields.WithDefault(false),
				),
				fields.New(
					"reload-interval",
					fields.TypeString,
					fields.WithHelp("Periodically reload external sources, for example 30s or 5m; disabled by default"),
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

	loaders := buildServeLoaders(s)
	if len(loaders) == 0 || s.WithEmbedded {
		if err := hs.Store.SetDefaultPackage(ctx, "glazed", ""); err != nil {
			return fmt.Errorf("assigning embedded package metadata: %w", err)
		}
	}
	var reloadMu sync.Mutex
	if len(loaders) > 0 {
		if err := loadServeSources(ctx, hs, loaders, !s.WithEmbedded, &reloadMu); err != nil {
			return err
		}
	}

	if err := startServeReloadLoop(ctx, hs, loaders, s, &reloadMu); err != nil {
		return err
	}

	count, err := hs.Store.Count(ctx)
	if err != nil {
		return fmt.Errorf("counting help sections: %w", err)
	}
	log.Info().Int64("sections", count).Msg("Loaded help sections")

	deps := HandlerDeps{Store: hs.Store}
	handler := NewServeHandler(deps, sc.spaHandler)
	return serveHTTP(s.Address, handler)
}

func loadServeSources(ctx context.Context, hs *help.HelpSystem, loaders []helploader.ContentLoader, clearBeforeLoad bool, mu *sync.Mutex) error {
	if mu != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	if clearBeforeLoad {
		staging := help.NewHelpSystem()
		if err := loadIntoHelpSystem(ctx, staging, loaders); err != nil {
			_ = staging.Store.Close()
			return err
		}
		defer func() { _ = staging.Store.Close() }()
		if err := replaceStoreSections(ctx, hs, staging); err != nil {
			return err
		}
		return nil
	}
	return loadIntoHelpSystem(ctx, hs, loaders)
}

func loadIntoHelpSystem(ctx context.Context, hs *help.HelpSystem, loaders []helploader.ContentLoader) error {
	for _, l := range loaders {
		log.Info().Str("source", l.String()).Msg("Loading help source")
		if err := l.Load(ctx, hs); err != nil {
			return fmt.Errorf("loading %s: %w", l.String(), err)
		}
	}
	return nil
}

func replaceStoreSections(ctx context.Context, target, source *help.HelpSystem) error {
	sections, err := source.Store.Find(ctx, func(*store.QueryCompiler) {})
	if err != nil {
		return fmt.Errorf("reading staged sections: %w", err)
	}
	if err := target.Store.Clear(ctx); err != nil {
		return fmt.Errorf("clearing preloaded sections: %w", err)
	}
	for _, section := range sections {
		if err := target.Store.Upsert(ctx, section); err != nil {
			return fmt.Errorf("replacing staged section %q: %w", section.Slug, err)
		}
	}
	return nil
}

func startServeReloadLoop(ctx context.Context, hs *help.HelpSystem, loaders []helploader.ContentLoader, s *ServeSettings, mu *sync.Mutex) error {
	if strings.TrimSpace(s.ReloadInterval) == "" {
		return nil
	}
	if len(loaders) == 0 {
		return fmt.Errorf("--reload-interval requires at least one external source")
	}
	interval, err := time.ParseDuration(s.ReloadInterval)
	if err != nil {
		return fmt.Errorf("parsing --reload-interval: %w", err)
	}
	if interval <= 0 {
		return fmt.Errorf("--reload-interval must be greater than zero")
	}
	clearBeforeLoad := !s.WithEmbedded
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := loadServeSources(ctx, hs, loaders, clearBeforeLoad, mu); err != nil {
					log.Error().Err(err).Msg("Reloading help sources failed")
					continue
				}
				count, err := hs.Store.Count(ctx)
				if err != nil {
					log.Error().Err(err).Msg("Counting help sections after reload failed")
					continue
				}
				log.Info().Int64("sections", count).Msg("Reloaded help sources")
			}
		}
	}()
	return nil
}

func buildServeLoaders(s *ServeSettings) []helploader.ContentLoader {
	var loaders []helploader.ContentLoader
	if len(helploader.NormalizeStringList(s.Paths)) > 0 {
		loaders = append(loaders, &helploader.MarkdownPathLoader{Paths: s.Paths})
	}
	if len(helploader.NormalizeStringList(s.FromJSON)) > 0 {
		loaders = append(loaders, &helploader.JSONFileLoader{Paths: s.FromJSON})
	}
	if len(helploader.NormalizeStringList(s.FromSQLite)) > 0 {
		loaders = append(loaders, &helploader.SQLiteLoader{Paths: s.FromSQLite})
	}
	if len(helploader.NormalizeStringList(s.FromSQLiteDir)) > 0 {
		loaders = append(loaders, &helploader.SQLiteDirLoader{Roots: s.FromSQLiteDir})
	}
	if len(helploader.NormalizeStringList(s.FromGlazedCmd)) > 0 {
		loaders = append(loaders, &helploader.GlazedCommandLoader{Binaries: s.FromGlazedCmd})
	}
	return loaders
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
