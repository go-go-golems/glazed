package sql

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
	"strings"
)

type OutputFormatter struct {
	TableName string
	UseUpsert bool
	// if 0, output all rows as a single INSERT statement, otherwise make a new statement every n rows
	SplitByRows int
	curIdx      int
	columns     []types.FieldName
	printEnd    bool
}

func valToSQL(i interface{}) string {
	var result string
	switch v := i.(type) {
	case string:
		// Escape single quotes with another single quote in string type
		result = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
	case nil:
		result = "NULL"
	default:
		// For the rest, let fmt.Sprintf handle their string conversion
		result = fmt.Sprintf("%v", v)
	}
	return result
}

func (f *OutputFormatter) printInsertBegin(w io.Writer) error {
	_, err := fmt.Fprintf(
		w,
		"INSERT INTO %s (%s) VALUES\n",
		f.TableName,
		strings.Join(f.columns, ", "))
	if err != nil {
		return err
	}
	f.printEnd = true
	return nil
}

func (f *OutputFormatter) printInsertEnd(w io.Writer) error {
	if !f.printEnd {
		return nil
	}
	if f.UseUpsert {
		_, err := fmt.Fprintf(w, "ON DUPLICATE KEY UPDATE\n")
		for i, col := range f.columns {
			if i > 0 {
				_, err = fmt.Fprintf(w, ",\n")
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "%s = VALUES(%s)", col, col)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, ";\n")
	f.printEnd = false
	return err
}

func (f *OutputFormatter) Close(ctx context.Context, w io.Writer) error {
	return f.printInsertEnd(w)
}

func (f *OutputFormatter) ContentType() string {
	return "application/sql"
}

func (f *OutputFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

func (f *OutputFormatter) RegisterRowMiddlewares(mw *middlewares.TableProcessor) error {
	return nil
}

type OutputFormatterOption func(*OutputFormatter)

func WithTableName(tableName string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.TableName = tableName
	}
}

func WithUseUpsert(useUpsert bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.UseUpsert = useUpsert
	}
}

func WithSplitByRows(splitByRows int) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.SplitByRows = splitByRows
	}
}

func NewOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		TableName:   "output",
		UseUpsert:   false,
		SplitByRows: 0,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *OutputFormatter) OutputRow(ctx context.Context, row types.Row, w io.Writer) error {
	if f.TableName == "" {
		return fmt.Errorf("table name is empty")
	}

	if row.Len() == 0 {
		return nil
	}

	if len(f.columns) == 0 {
		f.columns = []types.FieldName{}
		for pair := row.Oldest(); pair != nil; pair = pair.Next() {
			f.columns = append(f.columns, pair.Key)
		}
	}

	printInsert := false
	if f.curIdx == 0 {
		printInsert = true
	}

	if f.SplitByRows > 0 && f.curIdx == f.SplitByRows {
		err := f.printInsertEnd(w)
		if err != nil {
			return err
		}
		printInsert = true
		f.curIdx = 0
	}

	if printInsert {
		err := f.printInsertBegin(w)
		if err != nil {
			return err
		}
	}

	colIdx := 0
	if row.Len() > 0 {
		if !printInsert {
			_, err := fmt.Fprintf(w, ", ")
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(w, "(")
		if err != nil {
			return err
		}
		for _, col := range f.columns {
			v, ok := row.Get(col)
			if !ok {
				v = nil
			}

			_, err := fmt.Fprintf(w, "%s", valToSQL(v))
			if err != nil {
				return err
			}

			if colIdx < row.Len()-1 {
				_, err = fmt.Fprintf(w, ", ")
			}
			if err != nil {
				return err
			}
			colIdx++
		}
		_, err = fmt.Fprintf(w, ")")
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\n")
	if err != nil {
		return err
	}

	f.curIdx++

	return nil
}
