package cli

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

const CreateCommandSettingsSlug = "create-command-settings"

func NewCreateCommandSettingsLayer() (schema.Section, error) {
	createCommandSettingsLayer, err := schema.NewSection(
		CreateCommandSettingsSlug,
		"Create command settings",
		schema.WithFields(
			fields.New(
				"create-command",
				fields.TypeString,
				fields.WithHelp("Create a new command for the query, with the defaults updated"),
			),
			fields.New(
				"create-alias",
				fields.TypeString,
				fields.WithHelp("Create a CLI alias for the query"),
			),
			fields.New(
				"create-cliopatra",
				fields.TypeString,
				fields.WithHelp("Print the CLIopatra YAML for the command"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return createCommandSettingsLayer, nil
}

type CreateCommandSettings struct {
	CreateCommand   string `glazed:"create-command"`
	CreateAlias     string `glazed:"create-alias"`
	CreateCliopatra string `glazed:"create-cliopatra"`
}

type ProfileSettings struct {
	Profile     string `glazed:"profile"`
	ProfileFile string `glazed:"profile-file"`
}

const ProfileSettingsSlug = "profile-settings"

func NewProfileSettingsLayer() (schema.Section, error) {
	profileSettingsLayer, err := schema.NewSection(
		ProfileSettingsSlug,
		"Profile settings",
		schema.WithFields(
			fields.New(
				"profile",
				fields.TypeString,
				fields.WithHelp("Load the profile"),
			),
			fields.New(
				"profile-file",
				fields.TypeString,
				fields.WithHelp("Load the profile from a file"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return profileSettingsLayer, nil
}

// GlazedMinimalCommandSettings contains a subset of the most commonly used settings
type CommandSettings struct {
	PrintYAML              bool   `glazed:"print-yaml"`
	PrintParsedParameters  bool   `glazed:"print-parsed-parameters"`
	LoadParametersFromFile string `glazed:"load-parameters-from-file"`
	PrintSchema            bool   `glazed:"print-schema"`
	ConfigFile             string `glazed:"config-file"`
}

const CommandSettingsSlug = "command-settings"

func NewCommandSettingsLayer() (schema.Section, error) {
	glazedMinimalCommandLayer, err := schema.NewSection(
		CommandSettingsSlug,
		"General purpose command options",
		schema.WithFields(
			fields.New(
				"print-yaml",
				fields.TypeBool,
				fields.WithHelp("Print the command's YAML"),
			),
			fields.New(
				"print-parsed-parameters",
				fields.TypeBool,
				fields.WithHelp("Print the parsed parameters"),
			),
			// Deprecated: legacy per-command parameter file injection (removed from default flow)
			fields.New(
				"print-schema",
				fields.TypeBool,
				fields.WithHelp("Print the command's schema"),
			),
			fields.New(
				"config-file",
				fields.TypeString,
				fields.WithHelp("Explicit config file path to load via middlewares"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return glazedMinimalCommandLayer, nil
}
