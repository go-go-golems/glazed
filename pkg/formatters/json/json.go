package json

import (
	"bytes"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"os"
)

type OutputFormatter struct {
	OutputIndividualRows bool
	OutputFile           string
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
			log.Debug().Str("file", f.OutputFile).Msg("Writing output to file")
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
