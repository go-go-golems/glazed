package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"os"
)

type CsvCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*CsvCommand)(nil)

func NewCsvCommand() (*CsvCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &CsvCommand{
		CommandDescription: cmds.NewCommandDescription(
			"csv",
			cmds.WithShort("Format CSV files"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input-files",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"delimiter",
					parameters.ParameterTypeString,
					parameters.WithHelp("delimiter to use"),
					parameters.WithDefault(","),
				),
				parameters.NewParameterDefinition(
					"comment",
					parameters.ParameterTypeString,
					parameters.WithHelp("comment character to use"),
					parameters.WithDefault("#"),
				),
				parameters.NewParameterDefinition(
					"fields-per-record",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("number of fields per record (negative to disable)"),
					parameters.WithDefault(0),
				),
				parameters.NewParameterDefinition(
					"trim-leading-space",
					parameters.ParameterTypeBool,
					parameters.WithHelp("trim leading space"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"lazy-quotes",
					parameters.ParameterTypeBool,
					parameters.WithHelp("allow lazy quotes"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

type CsvSettings struct {
	InputFiles       []string `glazed.parameter:"input-files"`
	Delimiter        string   `glazed.parameter:"delimiter"`
	Comment          string   `glazed.parameter:"comment"`
	FieldsPerRecord  int      `glazed.parameter:"fields-per-record"`
	TrimLeadingSpace bool     `glazed.parameter:"trim-leading-space"`
	LazyQuotes       bool     `glazed.parameter:"lazy-quotes"`
}

func (c *CsvCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &CsvSettings{}
	err := parsedLayers.InitializeStructFromLayer(layers.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "failed to initialize csv settings from parameters")
	}

	commaRune := rune(s.Delimiter[0])

	commentRune := rune(s.Comment[0])

	options := []csv.ParseCSVOption{
		csv.WithComma(commaRune),
		csv.WithComment(commentRune),
		csv.WithFieldsPerRecord(s.FieldsPerRecord),
		csv.WithTrimLeadingSpace(s.TrimLeadingSpace),
		csv.WithLazyQuotes(s.LazyQuotes),
	}

	for _, arg := range s.InputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}

		// open arg and create a reader
		f, err := os.Open(arg)
		if err != nil {
			return errors.Wrap(err, "could not open file")
		}
		defer func(f *os.File) {
			_ = f.Close()
		}(f)

		header, s, err := csv.ParseCSV(f, options...)
		if err != nil {
			return errors.Wrap(err, "could not parse CSV file")
		}

		for _, row := range s {
			err = gp.AddRow(ctx, types.NewRowFromMapWithColumns(row, header))
			if err != nil {
				return errors.Wrap(err, "could not process CSV row")
			}
		}
	}

	return nil
}
