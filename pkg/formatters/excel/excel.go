package excel

import (
	"fmt"
	strings2 "github.com/go-go-golems/glazed/pkg/helpers/strings"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/xuri/excelize/v2"
	"strings"
)

type ExcelOutputFormatter struct {
	SheetName   string
	OutputFile  string
	Table       *types.Table
	middlewares []middlewares.TableMiddleware
}

func (E *ExcelOutputFormatter) GetTable() (*types.Table, error) {
	return E.Table, nil
}

func (E *ExcelOutputFormatter) AddRow(row types.Row) {
	E.Table.Rows = append(E.Table.Rows, row)
}

func (E *ExcelOutputFormatter) SetColumnOrder(columns []types.FieldName) {
	E.Table.Columns = columns
}

func (E *ExcelOutputFormatter) AddTableMiddleware(mw middlewares.TableMiddleware) {
	E.middlewares = append(E.middlewares, mw)
}

func (E *ExcelOutputFormatter) AddTableMiddlewareInFront(mw middlewares.TableMiddleware) {
	E.middlewares = append([]middlewares.TableMiddleware{mw}, E.middlewares...)
}

func (E *ExcelOutputFormatter) AddTableMiddlewareAtIndex(i int, mw middlewares.TableMiddleware) {
	E.middlewares = append(E.middlewares[:i], append([]middlewares.TableMiddleware{mw}, E.middlewares[i:]...)...)
}

func (E *ExcelOutputFormatter) Output() (string, error) {
	E.Table.Finalize()

	for _, middleware := range E.middlewares {
		newTable, err := middleware.Process(E.Table)
		if err != nil {
			return "", err
		}
		E.Table = newTable
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetIndex, err := f.NewSheet(E.SheetName)
	if err != nil {
		return "", err
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
		return "", err
	}

	for j, col := range E.Table.Columns {
		colIndex := strings2.ToAlphaString(j + 1)
		cellIndex := colIndex + "1"
		rowKeyToColumn[col] = colIndex
		err = f.SetCellValue(E.SheetName, cellIndex, col)
		if err != nil {
			return "", err
		}

		err = f.SetCellStyle(E.SheetName, cellIndex, cellIndex, headerStyle)
		if err != nil {
			return "", err
		}
	}

	for i, row := range E.Table.Rows {
		vals := row.GetValues()
		for _, j := range E.Table.Columns {
			val := vals[j]
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
				return "", err
			}
		}
	}

	f.SetActiveSheet(sheetIndex)

	if err := f.SaveAs(E.OutputFile); err != nil {
		return "", err
	}

	return fmt.Sprintf("Output file created successfully at %s", E.OutputFile), nil
}

func NewExcelOutputFormatter(sheetName string, outputFile string) *ExcelOutputFormatter {
	return &ExcelOutputFormatter{
		SheetName:   sheetName,
		OutputFile:  outputFile,
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
	}
}
