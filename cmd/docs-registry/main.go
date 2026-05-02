package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help/publish"
	"github.com/spf13/cobra"
)

var version = "dev"

type RegistryCommand struct {
	*cmds.CommandDescription
}

type settings struct {
	Address          string `glazed:"address"`
	MaxUploadBytes   int64  `glazed:"max-upload-bytes"`
	TempDir          string `glazed:"temp-dir"`
	PackageRoot      string `glazed:"package-root"`
	PublisherCatalog string `glazed:"publisher-catalog"`
}

var _ cmds.BareCommand = (*RegistryCommand)(nil)

func NewRegistryCommand() (*RegistryCommand, error) {
	return &RegistryCommand{CommandDescription: cmds.NewCommandDescription(
		"docs-registry",
		cmds.WithShort("Accept versioned Glazed help database uploads"),
		cmds.WithLong(`docs-registry is the Phase 1 upload service for docs.yolo.scapegoat.dev.

It authorizes package publishers, validates uploaded Glazed help SQLite
databases, and publishes them into a package/version storage backend.`),
		cmds.WithFlags(
			fields.New("address", fields.TypeString, fields.WithHelp("HTTP listen address"), fields.WithDefault(":8090")),
			fields.New("max-upload-bytes", fields.TypeInteger, fields.WithHelp("Maximum SQLite upload size in bytes"), fields.WithDefault(64<<20)),
			fields.New("temp-dir", fields.TypeString, fields.WithHelp("Directory for temporary uploads"), fields.WithDefault("")),
			fields.New("package-root", fields.TypeString, fields.WithHelp("Root directory where package/version SQLite DBs are published"), fields.WithRequired(true)),
			fields.New("publisher-catalog", fields.TypeString, fields.WithHelp("JSON file with static publisher token hashes"), fields.WithRequired(true)),
		),
	)}, nil
}

func (c *RegistryCommand) Run(ctx context.Context, parsedValues *values.Values) error {
	s := &settings{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	return run(ctx, s)
}

func newRootCommand() *cobra.Command {
	registryCommand, err := NewRegistryCommand()
	if err != nil {
		cobra.CheckErr(err)
	}
	cmd, err := cli.BuildCobraCommandFromCommand(registryCommand)
	if err != nil {
		cobra.CheckErr(err)
	}
	cmd.Use = "docs-registry"
	cmd.Version = version
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromCobra(cmd)
	}
	if err := logging.AddLoggingSectionToRootCommand(cmd, "docs-registry"); err != nil {
		cobra.CheckErr(err)
	}
	return cmd
}

func run(ctx context.Context, s *settings) error {
	catalog := publish.NewReloadablePublisherCatalog(publish.FilePublisherCatalogSource{Path: s.PublisherCatalog})
	if err := catalog.Reload(ctx); err != nil {
		return fmt.Errorf("load publisher catalog: %w", err)
	}
	store := publish.NewDirectoryPackageStore(s.PackageRoot)
	h := publish.NewRegistryHandler(catalog, store)
	h.MaxUploadBytes = s.MaxUploadBytes
	h.TempDir = s.TempDir

	srv := &http.Server{
		Addr:              s.Address,
		Handler:           h.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("docs registry listening", "address", s.Address)
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
	if err := newRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
