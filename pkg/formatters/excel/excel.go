package excel

import (
	"context"
	"fmt"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"io"
	"strings"
)

type OutputFormatter struct {
	SheetName  string
	OutputFile string

	f              *excelize.File
	sheetIndex     int
	headerStyle    int
	colIndex       int
	rowIndex       int
	rowKeyToColumn map[string]string
}

func (E *OutputFormatter) Close(ctx context.Context, w io.Writer) error {
	if E.f != nil {
		E.f.SetActiveSheet(E.sheetIndex)
		if err := E.f.SaveAs(E.OutputFile); err != nil {
			return err
		}

		err := E.f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (E *OutputFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	mw.AddRowMiddlewareInFront(row.NewFlattenObjectMiddleware())
	return nil
}

func (E *OutputFormatter) openFile() error {
	var err error
	if E.f == nil {
		E.f = excelize.NewFile()

		E.sheetIndex, err = E.f.NewSheet(E.SheetName)
		if err != nil {
			return err
		}

		// Set the headers in bold
		E.headerStyle, err = E.f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Pattern: 1,
				Color:   []string{"DDDDDD"},
			},
		})
		if err != nil {
			return err
		}

		E.rowKeyToColumn = make(map[string]string)
	}
	return nil
}

func (E *OutputFormatter) addColumns(fields []types.FieldName) error {
	for _, col := range fields {
		if _, present := E.rowKeyToColumn[col]; present {
			continue
		}
		colIndex := strings2.ToAlphaString(E.colIndex + 1)
		E.colIndex++
		cellIndex := colIndex + "1"
		E.rowKeyToColumn[col] = colIndex
		err := E.f.SetCellValue(E.SheetName, cellIndex, col)
		if err != nil {
			return err
		}

		err = E.f.SetCellStyle(E.SheetName, cellIndex, cellIndex, E.headerStyle)
		if err != nil {
			return err
		}
	}

	return nil
}

func (E *OutputFormatter) OutputRow(_ context.Context, row types.Row, w io.Writer) error {
	if E.f == nil {
		err := E.openFile()
		if err != nil {
			return err
		}
	}

	fields := types.GetFields(row)
	err := E.addColumns(fields)
	if err != nil {
		return err
	}

	for pair := row.Oldest(); pair != nil; pair = pair.Next() {
		colIndex, present := E.rowKeyToColumn[pair.Key]
		if !present {
			return errors.Errorf("column %s not found", pair.Key)
		}

		cellIndex := colIndex + fmt.Sprint(E.rowIndex+2)

		// Format val as a comma-separated list if it is a list
		if list, ok := pair.Value.([]interface{}); ok {
			valStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(list)), ","), "[]")
			err = E.f.SetCellValue(E.SheetName, cellIndex, valStr)
		} else {
			err = E.f.SetCellValue(E.SheetName, cellIndex, pair.Value)
		}

		if err != nil {
			return err
		}
	}

	E.rowIndex++

	return nil
}

func (f *OutputFormatter) ContentType() string {
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

type OutputFormatterOption func(*OutputFormatter)

func WithSheetName(sheetName string) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.SheetName = sheetName
	}
}

func WithOutputFile(outputFile string) OutputFormatterOption {
	return func(formatter *OutputFormatter) {
		formatter.OutputFile = outputFile
	}
}

func NewOutputFormatter(opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{}

	for _, opt := range opts {
		opt(f)
	}

	return f
}
