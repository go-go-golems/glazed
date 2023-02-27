package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

type YamlCommand struct {
	description *cmds.CommandDescription
}

func NewYamlCommand() (*YamlCommand, error) {
	glazedParameterLayer, err := cli.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &YamlCommand{
		description: cmds.NewCommandDescription(
			"yaml",
			cmds.WithShort("Format YAML data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-is-array",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Input is an array of objects"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input-files",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (y *YamlCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp *cmds.GlazeProcessor,
) error {
	inputIsArray, ok := ps["input-is-array"].(bool)
	if !ok {
		return fmt.Errorf("input-is-array flag is not a bool")
	}

	inputFiles, ok := ps["input-files"].([]string)
	if !ok {
		return fmt.Errorf("input-files is not a string list")
	}

	for _, arg := range inputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		f, err := os.Open(arg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", arg, err)
			os.Exit(1)
		}

		if inputIsArray {
			data := make([]map[string]interface{}, 0)
			err = yaml.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as array", arg)
			}

			i := 1
			for _, d := range data {
				err = gp.ProcessInputObject(d)
				if err != nil {
					return errors.Wrapf(err, "Error processing row %d of file %s as object", i, arg)
				}
				i++
			}
		} else {
			// read json file
			data := make(map[string]interface{})
			err = yaml.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as object", arg)
			}
			err = gp.ProcessInputObject(data)
			if err != nil {
				return errors.Wrapf(err, "Error processing file %s as object", arg)
			}
		}
	}

	return nil
}

func (y *YamlCommand) Description() *cmds.CommandDescription {
	return y.description
}
