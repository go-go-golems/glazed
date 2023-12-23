package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	json2 "github.com/go-go-golems/glazed/pkg/helpers/json"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"io"
	"os"
)

type JsonCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*JsonCommand)(nil)

func NewJsonCommand() (*JsonCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &JsonCommand{
		CommandDescription: cmds.NewCommandDescription(
			"json",
			cmds.WithShort("Format JSON data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-is-array",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Input is an array of objects (multiple files will be concatenated)"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"sanitize",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Sanitize JSON input"),
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

func (j *JsonCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	d := parsedLayers.GetDefaultParameterLayer()

	inputIsArray, ok := d.Parameters.GetValue("input-is-array").(bool)
	if !ok {
		return fmt.Errorf("input-is-array flag is not a bool")
	}

	inputFiles, ok := d.Parameters.GetValue("input-files").([]string)
	if !ok {
		return fmt.Errorf("input-files is not a string list")
	}

	sanitizeInput, ok := d.Parameters.GetValue("sanitize").(bool)
	if !ok {
		return fmt.Errorf("sanitize flag is not a bool")
	}

	fromMarkdown, ok := d.Parameters.GetValue("from-markdown").(bool)
	if !ok {
		return fmt.Errorf("from-markdown flag is not a bool")
	}

	for _, arg := range inputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		var err error
		var f io.Reader

		if sanitizeInput || fromMarkdown {
			b, err := os.ReadFile(arg)
			if err != nil {
				return errors.Wrapf(err, "Error reading file %s", arg)
			}

			s := json2.SanitizeJSONString(string(b), fromMarkdown)

			f = bytes.NewReader([]byte(s))
		} else {
			f, err = os.Open(arg)
			if err != nil {
				return errors.Wrapf(err, "Error opening file %s", arg)
			}
		}

		if inputIsArray {
			data := make([]types.Row, 0)
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
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
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
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
