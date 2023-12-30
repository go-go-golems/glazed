package cli

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func NewGlazedCommandLayer() (layers.ParameterLayer, error) {
	glazedCommandLayer, err := layers.NewParameterLayer(
		"glazed-command",
		"General purpose Command options",
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
			parameters.NewParameterDefinition(
				"print-yaml",
				parameters.ParameterTypeBool,
				parameters.WithHelp("Print the command's YAML"),
			),
			parameters.NewParameterDefinition(
				"load-parameters-from-json",
				parameters.ParameterTypeString,
				parameters.WithHelp("Load the command's flags from JSON"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return glazedCommandLayer, nil
}

type GlazedCommandSettings struct {
	CreateCommand          string `glazed.parameter:"create-command"`
	CreateAlias            string `glazed.parameter:"create-alias"`
	CreateCliopatra        string `glazed.parameter:"create-cliopatra"`
	PrintYAML              bool   `glazed.parameter:"print-yaml"`
	LoadParametersFromFile string `glazed.parameter:"load-parameters-from-file"`
}
