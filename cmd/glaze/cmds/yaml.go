package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	yaml2 "github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*YamlCommand)(nil)

func NewYamlCommand() (*YamlCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &YamlCommand{
		CommandDescription: cmds.NewCommandDescription(
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
				parameters.NewParameterDefinition(
					"from-markdown",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Input is markdown"),
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

type YamlSettings struct {
	InputIsArray bool     `glazed.parameter:"input-is-array"`
	Sanitize     bool     `glazed.parameter:"sanitize"`
	FromMarkdown bool     `glazed.parameter:"from-markdown"`
	InputFiles   []string `glazed.parameter:"input-files"`
}

func (y *YamlCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	d := parsedLayers.GetDefaultParameterLayer()

	s := &YamlSettings{}
	err := d.Parameters.InitializeStruct(s)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize yaml settings from parameters")
	}

	for _, arg := range s.InputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		var f io.Reader
		var err error

		if s.Sanitize || s.FromMarkdown {
			// read in file
			data, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			cleanData := yaml2.Clean(string(data), s.FromMarkdown)
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

		if s.InputIsArray {
			// TODO(manuel, 2023-06-25) We should implement an unmarshaller for maprow from yaml
			// See https://github.com/go-go-golems/glazed/issues/305
			data := make([]types.Row, 0)
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
				err = gp.AddRow(ctx, d)
				if err != nil {
					return errors.Wrapf(err, "Error processing row %d of file %s as object", i, arg)
				}
				i++
			}
		} else {
			// read json file
			data := types.NewRow()
			err = yaml.NewDecoder(f).Decode(&data)
			if err != nil {
				// check for EOF
				if err == io.EOF {
					return nil
				}
				return errors.Wrapf(err, "Error decoding file %s as object", arg)
			}
			err = gp.AddRow(ctx, data)
			if err != nil {
				return errors.Wrapf(err, "Error processing file %s as object", arg)
			}
		}
	}

	return nil
}
