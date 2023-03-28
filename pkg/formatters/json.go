package formatters

import (
	"bytes"
	"encoding/json"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"os"
)

type JSONOutputFormatter struct {
	OutputIndividualRows bool
	OutputFile           string
	Table                *types.Table
	middlewares          []middlewares.TableMiddleware
}

func (J *JSONOutputFormatter) GetTable() (*types.Table, error) {
	return J.Table, nil
}

func (J *JSONOutputFormatter) AddRow(row types.Row) {
	J.Table.Rows = append(J.Table.Rows, row)
}

func (f *JSONOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (J *JSONOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	J.middlewares = append(J.middlewares, mw)
}

func (J *JSONOutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	J.middlewares = append([]middlewares.TableMiddleware{mw}, J.middlewares...)
}

func (J *JSONOutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	J.middlewares = append(J.middlewares[:i], append([]middlewares.TableMiddleware{mw}, J.middlewares[i:]...)...)
}

func (J *JSONOutputFormatter) Output() (string, error) {
	J.Table.Finalize()

	for _, middleware := range J.middlewares {
		newTable, err := middleware.Process(J.Table)
		if err != nil {
			return "", err
		}
		J.Table = newTable
	}

	if J.OutputIndividualRows {
		var buf bytes.Buffer
		for _, row := range J.Table.Rows {
			jsonBytes, err := json.MarshalIndent(row.GetValues(), "", "  ")
			if err != nil {
				return "", err
			}
			buf.Write(jsonBytes)
		}

		if J.OutputFile != "" {
			log.Debug().Str("file", J.OutputFile).Msg("Writing output to file")
			err := os.WriteFile(J.OutputFile, buf.Bytes(), 0644)
			if err != nil {
				return "", err
			}
			return "", nil
		}

		return buf.String(), nil
	} else {
		// TODO(manuel, 2022-11-21) We should build a custom JSONMarshal for Table
		var rows []map[string]interface{}
		for _, row := range J.Table.Rows {
			rows = append(rows, row.GetValues())
		}
		jsonBytes, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return "", err
		}

		if J.OutputFile != "" {
			log.Debug().Str("file", J.OutputFile).Msg("Writing output to file")
			err := os.WriteFile(J.OutputFile, jsonBytes, 0644)
			if err != nil {
				return "", err
			}
			return "", nil
		}

		return string(jsonBytes), nil
	}
}

func NewJSONOutputFormatter(outputAsObjects bool, outputFile string) *JSONOutputFormatter {
	return &JSONOutputFormatter{
		OutputIndividualRows: outputAsObjects,
		Table:                types.NewTable(),
		OutputFile:           outputFile,
		middlewares:          []middlewares.TableMiddleware{},
	}
}
