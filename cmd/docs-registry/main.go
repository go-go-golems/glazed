package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-go-golems/glazed/pkg/help/publish"
	"github.com/spf13/cobra"
)

var version = "dev"

type settings struct {
	address          string
	maxUploadBytes   int64
	tempDir          string
	packageRoot      string
	publisherCatalog string
}

func newRootCommand() *cobra.Command {
	s := &settings{}
	cmd := &cobra.Command{
		Use:     "docs-registry",
		Short:   "Accept versioned Glazed help database uploads",
		Version: version,
		Long: `docs-registry is the Phase 1 upload service for docs.yolo.scapegoat.dev.

It authorizes package publishers, validates uploaded Glazed help SQLite
databases, and publishes them into a package/version storage backend. The
initial skeleton exposes health and package listing endpoints; upload storage
is completed in later Phase 1 tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), s)
		},
	}
	cmd.Flags().StringVar(&s.address, "address", ":8090", "HTTP listen address")
	cmd.Flags().Int64Var(&s.maxUploadBytes, "max-upload-bytes", 64<<20, "Maximum SQLite upload size in bytes")
	cmd.Flags().StringVar(&s.tempDir, "temp-dir", "", "Directory for temporary uploads")
	cmd.Flags().StringVar(&s.packageRoot, "package-root", "", "Root directory where package/version SQLite DBs are published")
	cmd.Flags().StringVar(&s.publisherCatalog, "publisher-catalog", "", "JSON file with static publisher token hashes")
	return cmd
}

func run(ctx context.Context, s *settings) error {
	if s.packageRoot == "" {
		return fmt.Errorf("--package-root is required")
	}
	if s.publisherCatalog == "" {
		return fmt.Errorf("--publisher-catalog is required")
	}
	catalog := publish.NewReloadablePublisherCatalog(publish.FilePublisherCatalogSource{Path: s.publisherCatalog})
	if err := catalog.Reload(ctx); err != nil {
		return fmt.Errorf("load publisher catalog: %w", err)
	}
	store := publish.NewDirectoryPackageStore(s.packageRoot)
	h := publish.NewRegistryHandler(catalog, store)
	h.MaxUploadBytes = s.maxUploadBytes
	h.TempDir = s.tempDir

	srv := &http.Server{
		Addr:              s.address,
		Handler:           h.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("docs registry listening", "address", s.address)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cmd := newRootCommand()
	cmd.SetContext(ctx)
	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
