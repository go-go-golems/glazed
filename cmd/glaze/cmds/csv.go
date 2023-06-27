package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
	"github.com/go-go-golems/glazed/pkg/processor"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"os"
)

type CsvCommand struct {
	description *cmds.CommandDescription
}

func NewCsvCommand() (*CsvCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	return &CsvCommand{
		description: cmds.NewCommandDescription(
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
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *CsvCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *CsvCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp processor.Processor,
) error {
	inputFiles, ok := ps["input-files"].([]string)
	if !ok {
		return errors.New("input-files argument is not a string list")
	}

	comma, _ := ps["delimiter"].(string)
	if len(comma) != 1 {
		return errors.New("delimiter must be a single character")
	}
	commaRune := rune(comma[0])

	comment, _ := ps["comment"].(string)
	if len(comment) != 1 {
		return errors.New("comment must be a single character")
	}
	commentRune := rune(comment[0])

	fieldsPerRecord, _ := ps["fields-per-record"].(int)
	trimLeadingSpace, _ := ps["trim-leading-space"].(bool)
	lazyQuotes, _ := ps["lazy-quotes"].(bool)

	options := []csv.ParseCSVOption{
		csv.WithComma(commaRune),
		csv.WithComment(commentRune),
		csv.WithFieldsPerRecord(fieldsPerRecord),
		csv.WithTrimLeadingSpace(trimLeadingSpace),
		csv.WithLazyQuotes(lazyQuotes),
	}

	finalHeaders := []string{}
	seenHeaders := map[string]interface{}{}

	for _, arg := range inputFiles {
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
			err = gp.ProcessInputObject(ctx, &types.SimpleRow{Hash: types.NewMapRowFromMapWithColumns(row, header)})
			if err != nil {
				return errors.Wrap(err, "could not process CSV row")
			}
		}

		// append headers to finalHeaders, only if they don't exist yet
		for _, h := range header {
			if _, ok := seenHeaders[h]; !ok {
				finalHeaders = append(finalHeaders, h)
				seenHeaders[h] = nil
			}
		}
	}

	err := gp.Processor().FinalizeTable(ctx)
	if err != nil {
		return errors.Wrap(err, "could not finalize table")
	}
	table := gp.Processor().GetTable()
	table.Columns = finalHeaders

	return nil
}
