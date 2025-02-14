package cli

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const CreateCommandSettingsSlug = "create-command-settings"

func NewCreateCommandSettingsLayer() (layers.ParameterLayer, error) {
	createCommandSettingsLayer, err := layers.NewParameterLayer(
		CreateCommandSettingsSlug,
		"General purpose command options",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"create-command",
				parameters.ParameterTypeString,
				parameters.WithHelp("Create a new command for the query, with the defaults updated"),
			),
			parameters.NewParameterDefinition(
				"create-alias",
				parameters.ParameterTypeString,
				parameters.WithHelp("Create a CLI alias for the query"),
			),
			parameters.NewParameterDefinition(
				"create-cliopatra",
				parameters.ParameterTypeString,
				parameters.WithHelp("Print the CLIopatra YAML for the command"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return createCommandSettingsLayer, nil
}

type CreateCommandSettings struct {
	CreateCommand   string `glazed.parameter:"create-command"`
	CreateAlias     string `glazed.parameter:"create-alias"`
	CreateCliopatra string `glazed.parameter:"create-cliopatra"`
}

type ProfileSettings struct {
	Profile     string `glazed.parameter:"profile"`
	ProfileFile string `glazed.parameter:"profile-file"`
}

const ProfileSettingsSlug = "profile-settings"

func NewProfileSettingsLayer() (layers.ParameterLayer, error) {
	profileSettingsLayer, err := layers.NewParameterLayer(
		ProfileSettingsSlug,
		"Profile settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"profile",
				parameters.ParameterTypeString,
				parameters.WithHelp("Load the profile"),
			),
			parameters.NewParameterDefinition(
				"profile-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Load the profile from a file"),
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
	PrintYAML              bool   `glazed.parameter:"print-yaml"`
	PrintParsedParameters  bool   `glazed.parameter:"print-parsed-parameters"`
	LoadParametersFromFile string `glazed.parameter:"load-parameters-from-file"`
	PrintSchema            bool   `glazed.parameter:"print-schema"`
}

const CommandSettingsSlug = "command-settings"

func NewCommandSettingsLayer() (layers.ParameterLayer, error) {
	glazedMinimalCommandLayer, err := layers.NewParameterLayer(
		CommandSettingsSlug,
		"Minimal set of general purpose command options",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"print-yaml",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Print the command's YAML"),
			),
			parameters.NewParameterDefinition(
				"print-parsed-parameters",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Print the parsed parameters"),
			),
			parameters.NewParameterDefinition(
				"load-parameters-from-file",
				parameters.ParameterTypeString,
				parameters.WithHelp("Load the command's flags from file"),
			),
			parameters.NewParameterDefinition(
				"print-schema",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Print the command's schema"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return glazedMinimalCommandLayer, nil
}
