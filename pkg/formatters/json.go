package formatters

import (
	"bytes"
	"encoding/json"
	"glazed/pkg/middlewares"
	"glazed/pkg/types"
)

type JSONOutputFormatter struct {
	OutputIndividualRows bool
	Table                *types.Table
	middlewares          []middlewares.TableMiddleware
}

func (J *JSONOutputFormatter) AddRow(row types.Row) {
	J.Table.Rows = append(J.Table.Rows, row)
}

func (J *JSONOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	J.middlewares = append(J.middlewares, mw)
}

func (J *JSONOutputFormatter) Output() (string, error) {
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
				panic(err)
			}
			buf.Write(jsonBytes)
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
			panic(err)
		}
		return string(jsonBytes), nil
	}
}

func NewJSONOutputFormatter(outputAsObjects bool) *JSONOutputFormatter {
	return &JSONOutputFormatter{
		OutputIndividualRows: outputAsObjects,
		Table:                types.NewTable(),
		middlewares:          []middlewares.TableMiddleware{},
	}
}
