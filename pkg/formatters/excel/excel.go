package excel

import (
	"context"
	"fmt"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/xuri/excelize/v2"
	"io"
	"strings"
)

type OutputFormatter struct {
	SheetName  string
	OutputFile string
}

func (E *OutputFormatter) RegisterMiddlewares(mw *middlewares.TableProcessor) error {
	mw.AddRowMiddlewareInFront(row.NewFlattenObjectMiddleware())
	return nil
}

func (E *OutputFormatter) Output(_ context.Context, table_ *types.Table, w io.Writer) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetIndex, err := f.NewSheet(E.SheetName)
	if err != nil {
		return err
	}

	rowKeyToColumn := make(map[string]string)

	// Set the headers in bold
	headerStyle, err := f.NewStyle(&excelize.Style{
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

	for j, col := range table_.Columns {
		colIndex := strings2.ToAlphaString(j + 1)
		cellIndex := colIndex + "1"
		rowKeyToColumn[col] = colIndex
		err = f.SetCellValue(E.SheetName, cellIndex, col)
		if err != nil {
			return err
		}

		err = f.SetCellStyle(E.SheetName, cellIndex, cellIndex, headerStyle)
		if err != nil {
			return err
		}
	}

	for i, row := range table_.Rows {
		for _, j := range table_.Columns {
			val, present := row.Get(j)
			if !present {
				continue
			}

			colIndex := rowKeyToColumn[j]
			cellIndex := colIndex + fmt.Sprint(i+2)

			// Format val as a comma-separated list if it is a list
			if list, ok := val.([]interface{}); ok {
				valStr := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(list)), ","), "[]")
				err = f.SetCellValue(E.SheetName, cellIndex, valStr)
			} else {
				err = f.SetCellValue(E.SheetName, cellIndex, val)
			}

			if err != nil {
				return err
			}
		}
	}

	f.SetActiveSheet(sheetIndex)

	if err := f.SaveAs(E.OutputFile); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Output file created successfully at %s", E.OutputFile)

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
