package main

import (
	"context"
	"errors"
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
	Address                 string `glazed:"address"`
	MaxUploadBytes          int64  `glazed:"max-upload-bytes"`
	TempDir                 string `glazed:"temp-dir"`
	PackageRoot             string `glazed:"package-root"`
	AuthMode                string `glazed:"auth-mode"`
	PublisherCatalog        string `glazed:"publisher-catalog"`
	JWTIssuer               string `glazed:"jwt-issuer"`
	JWTClientID             string `glazed:"jwt-client-id"`
	MaxConcurrentUploads    int    `glazed:"max-concurrent-uploads"`
	RateLimitRequestsPerMin int    `glazed:"rate-limit-requests-per-minute"`
	RateLimitBurst          int    `glazed:"rate-limit-burst"`
	AllowOverwrite          bool   `glazed:"allow-overwrite"`
	MaxPackageBytes         int64  `glazed:"max-package-bytes"`
	MaxVersionsPerPackage   int    `glazed:"max-versions-per-package"`
}

var _ cmds.BareCommand = (*RegistryCommand)(nil)

func NewRegistryCommand() (*RegistryCommand, error) {
	return &RegistryCommand{CommandDescription: cmds.NewCommandDescription(
		"docs-registry",
		cmds.WithShort("Accept versioned Glazed help database uploads"),
		cmds.WithLong(`docs-registry is the upload service for docs.yolo.scapegoat.dev.

It authorizes package publishers, validates uploaded Glazed help SQLite
databases, and publishes them into a package/version storage backend.

The registry supports two publisher auth modes:
  static-catalog  — read package-scoped static token hashes from publishers.json
  vault-oidc-jwt  — validate Vault Identity/OIDC publish JWTs via discovery/JWKS`),
		cmds.WithFlags(
			fields.New("address", fields.TypeString, fields.WithHelp("HTTP listen address"), fields.WithDefault(":8090")),
			fields.New("max-upload-bytes", fields.TypeInteger, fields.WithHelp("Maximum SQLite upload size in bytes"), fields.WithDefault(64<<20)),
			fields.New("temp-dir", fields.TypeString, fields.WithHelp("Directory for temporary uploads"), fields.WithDefault("")),
			fields.New("package-root", fields.TypeString, fields.WithHelp("Root directory where package/version SQLite DBs are published"), fields.WithRequired(true)),
			fields.New("auth-mode", fields.TypeString, fields.WithHelp("Publisher auth mode: static-catalog or vault-oidc-jwt"), fields.WithDefault("static-catalog")),
			fields.New("publisher-catalog", fields.TypeString, fields.WithHelp("JSON file with static publisher token hashes; required for --auth-mode static-catalog"), fields.WithDefault("")),
			fields.New("jwt-issuer", fields.TypeString, fields.WithHelp("Expected OIDC issuer URL for --auth-mode vault-oidc-jwt, for example https://vault.example/v1/identity/oidc"), fields.WithDefault("")),
			fields.New("jwt-client-id", fields.TypeString, fields.WithHelp("Expected JWT audience/client ID for --auth-mode vault-oidc-jwt"), fields.WithDefault("")),
			fields.New("max-concurrent-uploads", fields.TypeInteger, fields.WithHelp("Maximum concurrent publish uploads; 0 disables the limit"), fields.WithDefault(2)),
			fields.New("rate-limit-requests-per-minute", fields.TypeInteger, fields.WithHelp("Per-client per-route request rate limit; 0 disables rate limiting"), fields.WithDefault(60)),
			fields.New("rate-limit-burst", fields.TypeInteger, fields.WithHelp("Per-client per-route rate limit burst size; ignored when rate limiting is disabled"), fields.WithDefault(10)),
			fields.New("allow-overwrite", fields.TypeBool, fields.WithHelp("Allow publishing different bytes over an existing package version; disabled by default"), fields.WithDefault(false)),
			fields.New("max-package-bytes", fields.TypeInteger, fields.WithHelp("Maximum total stored bytes per package; 0 disables the quota"), fields.WithDefault(0)),
			fields.New("max-versions-per-package", fields.TypeInteger, fields.WithHelp("Maximum stored versions per package; 0 disables the quota"), fields.WithDefault(0)),
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
	auth, err := buildPublisherAuth(ctx, s)
	if err != nil {
		return err
	}
	store := publish.NewDirectoryPackageStore(s.PackageRoot)
	store.AllowOverwrite = s.AllowOverwrite
	store.MaxPackageBytes = s.MaxPackageBytes
	store.MaxVersionsPerPackage = s.MaxVersionsPerPackage
	h := publish.NewRegistryHandler(auth, store)
	h.MaxUploadBytes = s.MaxUploadBytes
	h.TempDir = s.TempDir
	h.MaxConcurrentUploads = s.MaxConcurrentUploads
	h.RateLimitRequestsPerMin = s.RateLimitRequestsPerMin
	h.RateLimitBurst = s.RateLimitBurst

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

func buildPublisherAuth(ctx context.Context, s *settings) (publish.PublisherAuth, error) {
	switch s.AuthMode {
	case "", "static-catalog":
		if s.PublisherCatalog == "" {
			return nil, errors.New("--publisher-catalog is required for --auth-mode static-catalog")
		}
		catalog := publish.NewReloadablePublisherCatalog(publish.FilePublisherCatalogSource{Path: s.PublisherCatalog})
		if err := catalog.Reload(ctx); err != nil {
			return nil, fmt.Errorf("load publisher catalog: %w", err)
		}
		return catalog, nil
	case "vault-oidc-jwt":
		if s.JWTIssuer == "" {
			return nil, errors.New("--jwt-issuer is required for --auth-mode vault-oidc-jwt")
		}
		if s.JWTClientID == "" {
			return nil, errors.New("--jwt-client-id is required for --auth-mode vault-oidc-jwt")
		}
		auth, err := publish.NewJWTPublisherAuth(ctx, s.JWTIssuer, s.JWTClientID)
		if err != nil {
			return nil, fmt.Errorf("configure JWT publisher auth: %w", err)
		}
		return auth, nil
	default:
		return nil, fmt.Errorf("unknown --auth-mode %q", s.AuthMode)
	}
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
