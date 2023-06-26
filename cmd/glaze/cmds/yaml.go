package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	yaml2 "github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/go-go-golems/glazed/pkg/processor"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"strings"
)

type YamlCommand struct {
	description *cmds.CommandDescription
}

func NewYamlCommand() (*YamlCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
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
				parameters.NewParameterDefinition(
					"sanitize",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Sanitize input (very hacky, meant for LLM cleanup)"),
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
	_ map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp processor.Processor,
) error {
	inputIsArray, ok := ps["input-is-array"].(bool)
	if !ok {
		return fmt.Errorf("input-is-array flag is not a bool")
	}

	sanitize, ok := ps["sanitize"].(bool)
	if !ok {
		return fmt.Errorf("sanitize flag is not a bool")
	}

	inputFiles, ok := ps["input-files"].([]string)
	if !ok {
		return fmt.Errorf("input-files is not a string list")
	}

	for _, arg := range inputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		var f io.Reader
		var err error

		if sanitize {
			// read in file
			data, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			cleanData := yaml2.Clean(string(data))
			f = strings.NewReader(cleanData)
		} else {
			f, err = os.Open(arg)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", arg, err)
				os.Exit(1)
			}
			defer func(file *os.File) {
				_ = file.Close()
			}(f.(*os.File))
		}

		if inputIsArray {
			// TODO(manuel, 2023-06-25) We should implement an unmarshaller for maprow from yaml
			// See https://github.com/go-go-golems/glazed/issues/305
			data := make([]map[string]interface{}, 0)
			err = yaml.NewDecoder(f).Decode(&data)
			if err != nil {
				// check for EOF
				if err == io.EOF {
					return nil
				}
				return errors.Wrapf(err, "Error decoding file %s as array", arg)
			}

			i := 1
			for _, d := range data {
				err = gp.ProcessInputObject(ctx, types.NewMapRowFromMap(d))
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
				// check for EOF
				if err == io.EOF {
					return nil
				}
				return errors.Wrapf(err, "Error decoding file %s as object", arg)
			}
			err = gp.ProcessInputObject(ctx, types.NewMapRowFromMap(data))
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
