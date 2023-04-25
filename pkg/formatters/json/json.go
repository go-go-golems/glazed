package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"os"
)

type OutputFormatter struct {
	OutputIndividualRows bool
	OutputFile           string
	OutputFileTemplate   string
	OutputMultipleFiles  bool
	Table                *types.Table
	middlewares          []middlewares.TableMiddleware
}

func (f *OutputFormatter) GetTable() (*types.Table, error) {
	return f.Table, nil
}

func (f *OutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
}

func (f *OutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

// TODO(manuel: 2023-04-25) This could actually all be cleaned up with OutputFormatterOption

func (f *OutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	f.middlewares = append(f.middlewares, mw)
}

func (f *OutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	f.middlewares = append([]middlewares.TableMiddleware{mw}, f.middlewares...)
}

func (f *OutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	f.middlewares = append(f.middlewares[:i], append([]middlewares.TableMiddleware{mw}, f.middlewares[i:]...)...)
}

func (f *OutputFormatter) Output() (string, error) {
	f.Table.Finalize()

	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return "", err
		}
		f.Table = newTable
	}

	if f.OutputMultipleFiles {
		if f.OutputFileTemplate == "" && f.OutputFile == "" {
			return "", fmt.Errorf("neither output file or output file template is set")
		}

		s := ""

		for i, row := range f.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(f.OutputFile, f.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			jsonBytes, err := json.MarshalIndent(row.GetValues(), "", "  ")
			if err != nil {
				return "", err
			}
			err = os.WriteFile(outputFileName, jsonBytes, 0644)
			if err != nil {
				return "", err
			}
			s += fmt.Sprintf("Wrote output to %s\n", outputFileName)
		}

		return s, nil
	}

	if f.OutputIndividualRows {
		var buf bytes.Buffer
		for _, row := range f.Table.Rows {
			jsonBytes, err := json.MarshalIndent(row.GetValues(), "", "  ")
			if err != nil {
				return "", err
			}
			buf.Write(jsonBytes)
		}

		if f.OutputFile != "" {
			err := os.WriteFile(f.OutputFile, buf.Bytes(), 0644)
			if err != nil {
				return "", err
			}
			return "", nil
		}

		return buf.String(), nil
	} else {
		// TODO(manuel, 2022-11-21) We should build a custom JSONMarshal for Table
		var rows []map[string]interface{}
		for _, row := range f.Table.Rows {
			rows = append(rows, row.GetValues())
		}
		jsonBytes, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return "", err
		}

		if f.OutputFile != "" {
			log.Debug().Str("file", f.OutputFile).Msg("Writing output to file")
			err := os.WriteFile(f.OutputFile, jsonBytes, 0644)
			if err != nil {
				return "", err
			}
			return "", nil
		}

		return string(jsonBytes), nil
	}
}

type OutputFormatterOption func(*OutputFormatter)

func WithOutputIndividualRows(outputIndividualRows bool) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.OutputIndividualRows = outputIndividualRows
	}
}

func WithOutputFile(file string) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.OutputFile = file
	}
}

func WithOutputFileTemplate(outputFileTemplate string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputFileTemplate = outputFileTemplate
	}
}

func WithOutputMultipleFiles(outputMultipleFiles bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.OutputMultipleFiles = outputMultipleFiles
	}
}

func WithTable(table *types.Table) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.Table = table
	}
}

func NewOutputFormatter(options ...OutputFormatterOption) *OutputFormatter {
	ret := &OutputFormatter{
		OutputIndividualRows: false,
		Table:                types.NewTable(),
		OutputFile:           "",
		middlewares:          []middlewares.TableMiddleware{},
	}

	for _, option := range options {
		option(ret)
	}

	return ret
}
