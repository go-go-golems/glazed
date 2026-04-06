package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const envPrefix = "GLAZED_VAULT_SMOKE_TEST"

// AppSettings contains the fields whose precedence we want to make visible.
type AppSettings struct {
	Host     string `glazed:"host"`
	Password string `glazed:"password"`
	APIKey   string `glazed:"api-key"`
}

type VaultSmokeTestCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &VaultSmokeTestCommand{}

func NewVaultSmokeTestCommand() (*VaultSmokeTestCommand, error) {
	appSection, err := schema.NewSection(
		"app",
		"Application settings",
		schema.WithFields(
			fields.New(
				"host",
				fields.TypeString,
				fields.WithDefault("from-default-host"),
				fields.WithHelp("Non-secret field that should never hydrate from Vault"),
			),
			fields.New(
				"password",
				fields.TypeSecret,
				fields.WithDefault("from-default-password"),
				fields.WithHelp("Secret field that can hydrate from Vault"),
			),
			fields.New(
				"api-key",
				fields.TypeSecret,
				fields.WithDefault("from-default-api-key"),
				fields.WithHelp("Second secret field that can hydrate from Vault"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	vaultSection, err := sources.NewVaultSettingsSection()
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"vault-smoke-test",
		cmds.WithShort("Exercise Glazed Vault secret hydration against a real CLI command"),
		cmds.WithLong(`This example exists to prove the real Vault wiring end to end.

Expected precedence:
  defaults -> config -> vault -> env -> args -> cobra

Useful runs:
  go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml
  GLAZED_VAULT_SMOKE_TEST_PASSWORD=from-env go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml
  go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml --password from-flag
  go run ./cmd/examples/vault-smoke-test --config-file ./config.yaml --print-parsed-fields
`),
		cmds.WithSections(appSection, vaultSection),
	)

	return &VaultSmokeTestCommand{CommandDescription: desc}, nil
}

func (c *VaultSmokeTestCommand) Run(_ context.Context, vals *values.Values) error {
	settings := &AppSettings{}
	if err := vals.DecodeSectionInto("app", settings); err != nil {
		return errors.Wrap(err, "failed to decode app settings")
	}

	vaultSettings, err := sources.GetVaultSettings(vals)
	if err != nil {
		return errors.Wrap(err, "failed to decode vault settings")
	}

	hostSource, err := fieldSource(vals, "app", "host")
	if err != nil {
		return err
	}
	passwordSource, err := fieldSource(vals, "app", "password")
	if err != nil {
		return err
	}
	apiKeySource, err := fieldSource(vals, "app", "api-key")
	if err != nil {
		return err
	}

	fmt.Printf("env_prefix=%s\n", envPrefix)
	fmt.Printf("vault_addr=%s\n", vaultSettings.VaultAddr)
	fmt.Printf("secret_path=%s\n", vaultSettings.SecretPath)
	fmt.Printf("host=%s\n", settings.Host)
	fmt.Printf("host_source=%s\n", hostSource)
	// lgtm [go/clear-text-logging] -- This example intentionally prints resolved secret values to prove real precedence in the smoke harness.
	fmt.Printf("password=%s\n", settings.Password)
	// lgtm [go/clear-text-logging] -- This example intentionally prints the winning source for the resolved password in the smoke harness.
	fmt.Printf("password_source=%s\n", passwordSource)
	// lgtm [go/clear-text-logging] -- This example intentionally prints resolved secret values to prove real precedence in the smoke harness.
	fmt.Printf("api_key=%s\n", settings.APIKey)
	// lgtm [go/clear-text-logging] -- This example intentionally prints the winning source for the resolved API key in the smoke harness.
	fmt.Printf("api_key_source=%s\n", apiKeySource)

	return nil
}

func fieldSource(vals *values.Values, sectionName, fieldName string) (string, error) {
	fieldValue, ok := vals.GetField(sectionName, fieldName)
	if !ok {
		return "", errors.Errorf("field %s.%s not found", sectionName, fieldName)
	}
	if len(fieldValue.Log) == 0 {
		return "", nil
	}
	return fieldValue.Log[len(fieldValue.Log)-1].Source, nil
}

func resolveConfigFiles(parsedCommandSections *values.Values) ([]string, error) {
	commandSettings := &cli.CommandSettings{}
	if err := parsedCommandSections.DecodeSectionInto(cli.CommandSettingsSlug, commandSettings); err != nil {
		return nil, errors.Wrap(err, "failed to decode command settings")
	}

	configFile := strings.TrimSpace(commandSettings.ConfigFile)
	if configFile == "" {
		return nil, nil
	}

	return []string{configFile}, nil
}

func buildVaultSmokeMiddlewares(parsedCommandSections *values.Values, cmd *cobra.Command, args []string) ([]sources.Middleware, error) {
	configFiles, err := resolveConfigFiles(parsedCommandSections)
	if err != nil {
		return nil, err
	}

	vaultSettings, err := sources.BootstrapVaultSettings(configFiles, []string{envPrefix}, cmd)
	if err != nil {
		return nil, errors.Wrap(err, "failed to bootstrap vault settings")
	}

	middlewares := []sources.Middleware{
		sources.FromCobra(cmd, fields.WithSource("cobra")),
		sources.FromArgs(args, fields.WithSource("arguments")),
		sources.FromEnv(envPrefix, fields.WithSource("env")),
		sources.FromVaultSettings(vaultSettings),
	}
	if len(configFiles) > 0 {
		middlewares = append(middlewares,
			sources.FromFiles(configFiles, sources.WithParseOptions(fields.WithSource("config"))),
		)
	}
	middlewares = append(middlewares,
		sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
	)

	return middlewares, nil
}

func main() {
	command, err := NewVaultSmokeTestCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		command,
		cli.WithParserConfig(cli.CobraParserConfig{
			MiddlewaresFunc: buildVaultSmokeMiddlewares,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build cobra command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		return logging.InitLoggerFromCobra(cmd)
	}
	if err := logging.AddLoggingSectionToRootCommand(cobraCmd, "vault-smoke-test"); err != nil {
		fmt.Fprintf(os.Stderr, "failed to add logging flags: %v\n", err)
		os.Exit(1)
	}

	if err := cobraCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
