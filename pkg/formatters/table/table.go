package table

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/jedib0t/go-pretty/table"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

var (
	TableStyles = map[string]table.Style{
		"default": table.StyleDefault,
		"bold":    table.StyleBold,

		"bright":           table.StyleColoredBright,
		"dark":             table.StyleColoredDark,
		"black-on-blue":    table.StyleColoredBlackOnBlueWhite,
		"black-on-cyan":    table.StyleColoredBlackOnCyanWhite,
		"black-on-green":   table.StyleColoredBlackOnGreenWhite,
		"black-on-magenta": table.StyleColoredBlackOnMagentaWhite,
		"black-on-yellow":  table.StyleColoredBlackOnYellowWhite,
		"black-on-red":     table.StyleColoredBlackOnRedWhite,
		"blue-on-black":    table.StyleColoredBlueWhiteOnBlack,
		"cyan-on-black":    table.StyleColoredCyanWhiteOnBlack,
		"green-on-black":   table.StyleColoredGreenWhiteOnBlack,
		"magenta-on-black": table.StyleColoredMagentaWhiteOnBlack,
		"red-on-black":     table.StyleColoredRedWhiteOnBlack,
		"yellow-on-black":  table.StyleColoredYellowWhiteOnBlack,
		"double":           table.StyleDouble,
		"light":            table.StyleLight,
		"rounded":          table.StyleRounded,
	}
)

type OutputFormatter struct {
	Table               *types.Table
	OutputFileTemplate  string
	OutputMultipleFiles bool
	middlewares         []middlewares.TableMiddleware
	TableFormat         string
	TableStyle          table.Style
	OutputFile          string
}

type OutputFormatterOption func(*OutputFormatter)

func WithOutputFile(outputFile string) OutputFormatterOption {
	return func(tof *OutputFormatter) {
		tof.OutputFile = outputFile
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

func WithTableStyle(tableStyle string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		if style, ok := TableStyles[tableStyle]; ok {
			f.TableStyle = style
		} else {
			log.Warn().Msgf("Table style %s not found, using default", tableStyle)
		}
	}
}

func NewOutputFormatter(tableFormat string, opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		Table:       types.NewTable(),
		middlewares: []middlewares.TableMiddleware{},
		TableFormat: tableFormat,
		TableStyle:  table.StyleDefault,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (tof *OutputFormatter) GetTable() (*types.Table, error) {
	return tof.Table, nil
}

func (tof *OutputFormatter) Output() (string, error) {
	tof.Table.Finalize()

	for _, middleware := range tof.middlewares {
		newTable, err := middleware.Process(tof.Table)
		if err != nil {
			return "", err
		}
		tof.Table = newTable
	}

	if tof.OutputMultipleFiles {
		if tof.OutputFileTemplate == "" && tof.OutputFile == "" {
			return "", fmt.Errorf("neither output file or output file template is set")
		}

		s := ""

		for i, row := range tof.Table.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(tof.OutputFile, tof.OutputFileTemplate, row, i)
			if err != nil {
				return "", err
			}

			s_ := tof.makeTable([]types.Row{row})

			err = os.WriteFile(outputFileName, []byte(s_), 0644)
			if err != nil {
				return "", err
			}

			s += fmt.Sprintf("Wrote output to %s\n", outputFileName)
		}

		return s, nil
	}

	s := tof.makeTable(tof.Table.Rows)

	if tof.OutputFile != "" {
		log.Debug().Str("file", tof.OutputFile).Msg("Writing output to file")
		err := os.WriteFile(tof.OutputFile, []byte(s), 0644)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return s, nil
}

func (tof *OutputFormatter) makeTable(rows []types.Row) string {
	t := table.NewWriter()

	headers, _ := cast.CastList[interface{}](tof.Table.Columns)
	t.AppendHeader(headers)
	for _, row := range rows {
		values := row.GetValues()
		var row_ []interface{}
		for _, column := range tof.Table.Columns {
			s := ""
			if v, ok := values[column]; ok {
				s = valueToString(v)
			}
			row_ = append(row_, s)
		}
		t.AppendRow(row_)
	}

	if tof.TableFormat == "markdown" {
		return t.RenderMarkdown()
	} else if tof.TableFormat == "html" {
		return t.RenderHTML()
	} else {
		t.SetStyle(tof.TableStyle)
		return t.Render()
	}
}

func valueToString(v types.GenericCellValue) string {
	var s string
	if v_, ok := v.([]interface{}); ok {
		var elms []string
		for _, elm := range v_ {
			elms = append(elms, fmt.Sprintf("%v", elm))
		}
		s = strings.Join(elms, ", ")
	} else if v_, ok := v.(map[string]interface{}); ok {
		var elms []string
		for k, v__ := range v_ {
			elms = append(elms, fmt.Sprintf("%v:%v", k, v__))
		}
		s = strings.Join(elms, ",")
	} else {
		s = fmt.Sprintf("%v", v)
	}
	return s
}

func (tof *OutputFormatter) AddTableMiddleware(m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares, m)
}

func (tof *OutputFormatter) AddTableMiddlewareInFront(m middlewares.TableMiddleware) {
	tof.middlewares = append([]middlewares.TableMiddleware{m}, tof.middlewares...)
}

func (tof *OutputFormatter) AddTableMiddlewareAtIndex(i int, m middlewares.TableMiddleware) {
	tof.middlewares = append(tof.middlewares[:i], append([]middlewares.TableMiddleware{m}, tof.middlewares[i:]...)...)
}

func (tof *OutputFormatter) AddRow(row types.Row) {
	tof.Table.Rows = append(tof.Table.Rows, row)
}

func (tof *OutputFormatter) SetColumnOrder(columnOrder []types.FieldName) {
	tof.Table.Columns = columnOrder
}
