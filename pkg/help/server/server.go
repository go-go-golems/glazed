package server

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/help/store"
)

// DefaultAddr is the TCP address the server listens on when no --address flag is supplied.
const DefaultAddr = ":8088"

// Server is a convenience wrapper around an http.Server that wires the API handler,
// optional SPA handler, and CORS middleware together. It also handles graceful shutdown.
//
// For callers that want to mount the help API as part of a larger server, prefer
// using NewHandler directly. Server is intended for the standalone `glaze serve` command.
type Server struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	deps HandlerDeps

	// staticFS is non-nil when the SPA is embedded.
	staticFS embed.FS

	handler http.Handler
}

// ServerOption is a functional option for NewServer.
type ServerOption func(*Server)

// WithAddr overrides the default listen address.
func WithAddr(addr string) ServerOption {
	return func(s *Server) { s.Addr = addr }
}

// WithSlogger sets the logger used for request logging.
func WithSlogger(logger *slog.Logger) ServerOption {
	return func(s *Server) { s.deps.Logger = logger }
}

// WithReadTimeout sets the per-request read timeout. Default is 10 seconds.
func WithReadTimeout(d time.Duration) ServerOption {
	return func(s *Server) { s.ReadTimeout = d }
}

// WithWriteTimeout sets the per-request write timeout. Default is 30 seconds.
func WithWriteTimeout(d time.Duration) ServerOption {
	return func(s *Server) { s.WriteTimeout = d }
}

// NewServer builds a Server wired to the given store. The store must already be
// initialised and populated (e.g. via HelpSystem.LoadSectionsFromFS).
//
// If a Vite build is embedded (via //go:embed dist in the calling package),
// pass it with WithSPA. If not, omit it — only the API will be served.
//
// The returned Server is not started; call ListenAndServe.
func NewServer(st *store.Store, opts ...ServerOption) *Server {
	if st == nil {
		panic("server.NewServer: st must not be nil")
	}

	srv := &Server{
		Addr:         DefaultAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		deps: HandlerDeps{
			Store:  st,
			Logger: slog.Default(),
		},
	}
	for _, opt := range opts {
		opt(srv)
	}

	srv.handler = srv.buildHandler()
	return srv
}

// WithSPA embeds a Vite build filesystem and serves it as a fallback for
// non-API routes. The fsys argument must be an embedded filesystem created
// with //go:embed and should include a "dist" subdirectory containing the
// Vite output (index.html + assets/).
//
// Example usage in cmd/help-browser:
//
//	//go:embed dist
//	var staticFS embed.FS
//
//	srv := server.NewServer(st,
//	  server.WithSPA(staticFS, "dist"),
//	)
func WithSPA(fsys embed.FS) ServerOption {
	return func(s *Server) { s.staticFS = fsys }
}

// buildHandler assembles the final http.Handler:
//
//	CORS middleware (outermost)
//	  -> API handler at /api/*
//	    -> SPA handler for all other paths (serves index.html, only if SPA configured)
func (s *Server) buildHandler() http.Handler {
	apiHandler := NewHandler(s.deps)

	h := apiHandler

	// If the SPA filesystem is set, wrap the API handler in the SPA fallback.
	// SPAHandler returns a middleware that tries to serve static files first,
	// then delegates to next (the API handler), and finally falls back to
	// index.html for client-side routing.
	if s.staticFS != (embed.FS{}) {
		h = SPAHandler(s.staticFS, "dist")(h)
	}

	return NewCORSHandler(h)
}

// ListenAndServe starts the HTTP server and blocks until the process is interrupted.
// On SIGINT or SIGTERM it performs a graceful shutdown with a 5-second timeout,
// returning nil on clean exit or the shutdown error.
func (s *Server) ListenAndServe() error {
	httpSrv := &http.Server{
		Addr:         s.Addr,
		Handler:      s.handler,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		s.deps.Logger.Info("starting help browser server", "addr", s.Addr)
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
		s.deps.Logger.Info("shutting down", "signal", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpSrv.Shutdown(ctx)
	}
}
