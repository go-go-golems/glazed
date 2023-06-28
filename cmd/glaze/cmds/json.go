package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/processor"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"os"
)

type JsonCommand struct {
	description *cmds.CommandDescription
}

func NewJsonCommand() (*JsonCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &JsonCommand{
		description: cmds.NewCommandDescription(
			"json",
			cmds.WithShort("Format JSON data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-is-array",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Input is an array of objects (multiple files will be concatenated)"),
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

func (j *JsonCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp processor.Processor,
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
			return errors.Wrapf(err, "Error opening file %s", arg)
		}

		if inputIsArray {
			data := make([]types.Row, 0)
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as array", arg)
			}

			i := 1
			for _, d := range data {
				err = gp.ProcessInputObject(ctx, d)
				if err != nil {
					return errors.Wrapf(err, "Error processing row %d of file %s as object", i, arg)
				}
				i++
			}
		} else {
			// read json file
			data := types.NewMapRow()
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as object", arg)
			}

			err = gp.ProcessInputObject(ctx, data)
			if err != nil {
				return errors.Wrapf(err, "Error processing file %s as object", arg)
			}
		}
	}

	return nil
}

func (j *JsonCommand) Description() *cmds.CommandDescription {
	return j.description
}
