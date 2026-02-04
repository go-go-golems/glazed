package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/helpers/csv"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/settings"
)

type CsvCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*CsvCommand)(nil)

func NewCsvCommand() (*CsvCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	return &CsvCommand{
		CommandDescription: cmds.NewCommandDescription(
			"csv",
			cmds.WithShort("Format CSV files"),
			cmds.WithArguments(
				fields.New(
					"input-files",
					fields.TypeStringList,
					fields.WithRequired(true),
				),
			),
			cmds.WithFlags(
				fields.New(
					"delimiter",
					fields.TypeString,
					fields.WithHelp("delimiter to use"),
					fields.WithDefault(","),
				),
				fields.New(
					"comment",
					fields.TypeString,
					fields.WithHelp("comment character to use"),
					fields.WithDefault("#"),
				),
				fields.New(
					"fields-per-record",
					fields.TypeInteger,
					fields.WithHelp("number of fields per record (negative to disable)"),
					fields.WithDefault(0),
				),
				fields.New(
					"trim-leading-space",
					fields.TypeBool,
					fields.WithHelp("trim leading space"),
					fields.WithDefault(false),
				),
				fields.New(
					"lazy-quotes",
					fields.TypeBool,
					fields.WithHelp("allow lazy quotes"),
					fields.WithDefault(false),
				),
			),
			cmds.WithSections(
				glazedSection,
			),
		),
	}, nil
}

type CsvSettings struct {
	InputFiles       []string `glazed:"input-files"`
	Delimiter        string   `glazed:"delimiter"`
	Comment          string   `glazed:"comment"`
	FieldsPerRecord  int      `glazed:"fields-per-record"`
	TrimLeadingSpace bool     `glazed:"trim-leading-space"`
	LazyQuotes       bool     `glazed:"lazy-quotes"`
}

func (c *CsvCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &CsvSettings{}
	err := vals.DecodeSectionInto(schema.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "failed to initialize csv settings from fields")
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
