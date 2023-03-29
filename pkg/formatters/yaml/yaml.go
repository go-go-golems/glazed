package yaml

import (
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"os"
)

type YAMLOutputFormatter struct {
	Table       *types.Table
	OutputFile  string
	middlewares []middlewares.TableMiddleware
}

func (Y *YAMLOutputFormatter) GetTable() (*types.Table, error) {
	return Y.Table, nil
}

func (Y *YAMLOutputFormatter) AddRow(row types.Row) {
	Y.Table.Rows = append(Y.Table.Rows, row)
}

func (f *YAMLOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	f.Table.Columns = columns
}

func (Y *YAMLOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares, mw)
}

func (Y *YAMLOutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	Y.middlewares = append([]middlewares.TableMiddleware{mw}, Y.middlewares...)
}

func (Y *YAMLOutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	Y.middlewares = append(Y.middlewares[:i], append([]middlewares.TableMiddleware{mw}, Y.middlewares[i:]...)...)
}

func (Y *YAMLOutputFormatter) Output() (string, error) {
	Y.Table.Finalize()

	for _, middleware := range Y.middlewares {
		newTable, err := middleware.Process(Y.Table)
		if err != nil {
			return "", err
		}
		Y.Table = newTable
	}

	var rows []map[string]interface{}
	for _, row := range Y.Table.Rows {
		rows = append(rows, row.GetValues())

	}

	d, err := yaml.Marshal(rows)
	if err != nil {
		return "", err
	}

	if Y.OutputFile != "" {
		log.Debug().Str("file", Y.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(Y.OutputFile, d, 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return string(d), nil
}

func NewYAMLOutputFormatter(outputFile string) *YAMLOutputFormatter {
	return &YAMLOutputFormatter{
		Table:       types.NewTable(),
		OutputFile:  outputFile,
		middlewares: []middlewares.TableMiddleware{},
	}
}
