package sources

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

const VaultSettingsSlug = "vault-settings"

type VaultTokenSource string

const (
	VaultTokenSourceAuto   VaultTokenSource = "auto"
	VaultTokenSourceEnv    VaultTokenSource = "env"
	VaultTokenSourceFile   VaultTokenSource = "file"
	VaultTokenSourceLookup VaultTokenSource = "lookup"
)

type VaultSettings struct {
	VaultAddr        string `glazed:"vault-addr"`
	VaultToken       string `glazed:"vault-token"`
	VaultTokenSource string `glazed:"vault-token-source"`
	VaultTokenFile   string `glazed:"vault-token-file"`
	SecretPath       string `glazed:"secret-path"`
}

func NewVaultSettingsSection() (schema.Section, error) {
	return schema.NewSection(
		VaultSettingsSlug,
		"Vault settings",
		schema.WithFields(
			fields.New(
				"vault-addr",
				fields.TypeString,
				fields.WithHelp("Vault server address"),
				fields.WithDefault("http://127.0.0.1:8200"),
			),
			fields.New(
				"vault-token",
				fields.TypeSecret,
				fields.WithHelp("Vault token"),
				fields.WithDefault(""),
			),
			fields.New(
				"vault-token-source",
				fields.TypeChoice,
				fields.WithHelp("Vault token source: auto|env|file|lookup"),
				fields.WithChoices(
					string(VaultTokenSourceAuto),
					string(VaultTokenSourceEnv),
					string(VaultTokenSourceFile),
					string(VaultTokenSourceLookup),
				),
				fields.WithDefault(string(VaultTokenSourceAuto)),
			),
			fields.New(
				"vault-token-file",
				fields.TypeString,
				fields.WithHelp("Path to the Vault token file (defaults to ~/.vault-token when token source is file)"),
				fields.WithDefault(""),
			),
			fields.New(
				"secret-path",
				fields.TypeString,
				fields.WithHelp("Vault secret path to hydrate TypeSecret fields from"),
				fields.WithDefault(""),
			),
		),
	)
}

func GetVaultSettings(parsed *values.Values) (*VaultSettings, error) {
	var settings VaultSettings
	if err := parsed.DecodeSectionInto(VaultSettingsSlug, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse vault settings: %w", err)
	}
	return &settings, nil
}
