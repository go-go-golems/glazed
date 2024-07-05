package cmds

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	"time"
)

type JsonCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*JsonCommand)(nil)

type JsonSettings struct {
	InputIsArray bool     `glazed.parameter:"input-is-array"`
	Sanitize     bool     `glazed.parameter:"sanitize"`
	FromMarkdown bool     `glazed.parameter:"from-markdown"`
	TailMode     bool     `glazed.parameter:"tail"`
	InputFiles   []string `glazed.parameter:"input-files"`
}

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
				parameters.NewParameterDefinition(
					"tail",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Tail mode: read one JSON object per line"),
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
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (j *JsonCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &JsonSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize json settings from parameters")
	}

	for _, arg := range s.InputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		var err error
		var f io.Reader
		if s.Sanitize || s.FromMarkdown {
			b, err := os.ReadFile(arg)
			if err != nil {
				return errors.Wrapf(err, "Error reading file %s", arg)
			}
			s := json2.SanitizeJSONString(string(b), s.FromMarkdown)
			f = bytes.NewReader([]byte(s))
		} else {
			f, err = os.Open(arg)
			if err != nil {
				return errors.Wrapf(err, "Error opening file %s", arg)
			}
		}

		if s.TailMode {
			err = processTailMode(ctx, f, gp, arg)
		} else if s.InputIsArray {
			err = processArrayMode(ctx, f, gp, arg)
		} else {
			err = processObjectMode(ctx, f, gp, arg)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func processTailMode(ctx context.Context, f io.Reader, gp middlewares.Processor, arg string) error {
	file, ok := f.(*os.File)
	if ok {
		_, err := file.Seek(0, io.SeekEnd)
		if err != nil {
			return err
		}
	}

	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return errors.Wrap(err, "Error reading line from file")
		}

		if len(line) > 0 {
			var data types.Row
			err := json.Unmarshal(line, &data)
			if err != nil {
				return errors.Wrap(err, "Error decoding line as object")
			}
			err = gp.AddRow(ctx, data)
			if err != nil {
				return errors.Wrap(err, "Error processing line as object")
			}
		}

		if err == io.EOF {
			// Check if we should continue or exit
			select {
			case <-ctx.Done():
				return nil
			default:
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}
	}
}

func processArrayMode(ctx context.Context, f io.Reader, gp middlewares.Processor, arg string) error {
	data := make([]types.Row, 0)
	err := json.NewDecoder(f).Decode(&data)
	if err != nil {
		return errors.Errorf("Error decoding file %s as array", arg)
	}
	for i, d := range data {
		err = gp.AddRow(ctx, d)
		if err != nil {
			return errors.Wrapf(err, "Error processing row %d of file %s as object", i+1, arg)
		}
	}
	return nil
}

func processObjectMode(ctx context.Context, f io.Reader, gp middlewares.Processor, arg string) error {
	data := types.NewRow()
	err := json.NewDecoder(f).Decode(&data)
	if err != nil {
		return errors.Wrapf(err, "Error decoding file %s as object", arg)
	}
	err = gp.AddRow(ctx, data)
	if err != nil {
		return errors.Wrapf(err, "Error processing file %s as object", arg)
	}
	return nil
}
