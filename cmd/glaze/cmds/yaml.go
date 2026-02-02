package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	yaml2 "github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type YamlCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*YamlCommand)(nil)

func NewYamlCommand() (*YamlCommand, error) {
	glazedLayer, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &YamlCommand{
		CommandDescription: cmds.NewCommandDescription(
			"yaml",
			cmds.WithShort("Format YAML data"),
			cmds.WithFlags(
				fields.New(
					"input-is-array",
					fields.TypeBool,
					fields.WithHelp("Input is an array of objects"),
					fields.WithDefault(false),
				),
				fields.New(
					"sanitize",
					fields.TypeBool,
					fields.WithHelp("Sanitize input (very hacky, meant for LLM cleanup)"),
					fields.WithDefault(false),
				),
				fields.New(
					"from-markdown",
					fields.TypeBool,
					fields.WithHelp("Input is markdown"),
					fields.WithDefault(false),
				),
			),
			cmds.WithArguments(
				fields.New(
					"input-files",
					fields.TypeStringList,
					fields.WithRequired(true),
				),
			),
			cmds.WithLayersList(
				glazedLayer,
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

func (y *YamlCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &YamlSettings{}
	err := values.DecodeSectionInto(vals, schema.DefaultSlug, s)
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
