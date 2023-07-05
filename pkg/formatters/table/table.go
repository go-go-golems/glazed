package table

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
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
	OutputFileTemplate  string
	OutputMultipleFiles bool
	TableFormat         string
	TableStyle          table.Style
	TableStyleFile      string
	OutputFile          string
	PrintTableStyle     bool
	hasOutputHeaders    bool
}

func (tof *OutputFormatter) Close(ctx context.Context) error {
	return nil
}

func (tof *OutputFormatter) RegisterTableMiddlewares(mw *middlewares.TableProcessor) error {
	mw.AddRowMiddlewareInFront(row.NewFlattenObjectMiddleware())
	return nil
}

func (tof *OutputFormatter) RegisterRowMiddlewares(mw *middlewares.TableProcessor) error {
	mw.AddRowMiddlewareInFront(row.NewFlattenObjectMiddleware())
	return nil
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

func WithTableStyleFile(tableStyleFile string) OutputFormatterOption {
	return func(f *OutputFormatter) {
		f.TableStyleFile = tableStyleFile
	}
}

func WithPrintTableStyle(printTableStyle bool) OutputFormatterOption {
	return func(f *OutputFormatter) {
		if printTableStyle {
			f.PrintTableStyle = printTableStyle
		}
	}
}

func NewOutputFormatter(tableFormat string, opts ...OutputFormatterOption) *OutputFormatter {
	f := &OutputFormatter{
		TableFormat: tableFormat,
		TableStyle:  table.StyleDefault,
	}

	// avoid setting everything to uppercase
	f.TableStyle.Format.Header = text.FormatDefault
	f.TableStyle.Format.Footer = text.FormatDefault

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (tof *OutputFormatter) ContentType() string {
	switch tof.TableFormat {
	case "csv":
		return "text/csv"
	case "html":
		return "text/html"
	default:
		return "text/plain"
	}
}

func (tof *OutputFormatter) OutputTable(ctx context.Context, table_ *types.Table, w io.Writer) error {
	if tof.OutputMultipleFiles {
		if tof.OutputFileTemplate == "" && tof.OutputFile == "" {
			return fmt.Errorf("neither output file or output file template is set")
		}

		for i, row := range table_.Rows {
			outputFileName, err := formatters.ComputeOutputFilename(tof.OutputFile, tof.OutputFileTemplate, row, i)
			if err != nil {
				return err
			}

			f_, err := os.Create(outputFileName)
			if err != nil {
				return err
			}

			err = tof.makeTable(table_, []types.Row{row}, f_)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(w, "Wrote output to %s\n", outputFileName)
		}

		return nil
	}

	if tof.OutputFile != "" {
		f_, err := os.Create(tof.OutputFile)
		if err != nil {
			return err
		}
		err = tof.makeTable(table_, table_.Rows, f_)
		if err != nil {
			return err
		}

		_, _ = fmt.Fprintf(w, "Wrote output to %s\n", tof.OutputFile)
		return nil
	}

	err := tof.makeTable(table_, table_.Rows, w)
	if err != nil {
		return err
	}

	return nil
}

func (tof *OutputFormatter) OutputRow(ctx context.Context, row_ types.Row, w io.Writer) error {
	switch tof.TableFormat {
	case "html":
		return tof.outputHTMLRow(row_, w)
	case "markdown":
		return tof.outputMarkdownRow(row_, w)
	default:
		return errors.New("unsupported table format")
	}
}

func (tof *OutputFormatter) outputMarkdownRow(row_ types.Row, w io.Writer) error {
	if !tof.hasOutputHeaders {
		fields := types.GetFields(row_)

		for _, field := range fields {
			_, err := fmt.Fprintf(w, "| %s ", field)
			if err != nil {
				return err
			}
		}
		_, err := fmt.Fprintf(w, "|\n")
		if err != nil {
			return err
		}

		for range fields {
			_, err = fmt.Fprintf(w, "| --- ")
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintf(w, "|\n")
		if err != nil {
			return err
		}

		tof.hasOutputHeaders = true
	}

	for pair := row_.Oldest(); pair != nil; pair = pair.Next() {
		_, err := fmt.Fprintf(w, "| %s ", pair.Value)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "|\n")
	if err != nil {
		return err
	}

	return nil
}

func (tof *OutputFormatter) outputHTMLRow(row_ types.Row, w io.Writer) error {
	if !tof.hasOutputHeaders {
		fields := types.GetFields(row_)

		_, err := fmt.Fprintf(w, "<tr>")
		if err != nil {
			return err
		}
		for _, field := range fields {
			_, err = fmt.Fprintf(w, "<th>%s</th>", field)
			if err != nil {
				return err
			}
		}

		_, err = fmt.Fprintf(w, "</tr>\n")
		if err != nil {
			return err
		}

		tof.hasOutputHeaders = true
	}

	_, err := fmt.Fprintf(w, "<tr>")
	if err != nil {
		return err
	}

	for pair := row_.Oldest(); pair != nil; pair = pair.Next() {
		_, err = fmt.Fprintf(w, "<td>%s</td>", valueToString(pair.Value))
		if err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(w, "</tr>\n")
	if err != nil {
		return err
	}
	return nil
}

func (tof *OutputFormatter) makeTable(table_ *types.Table, rows []types.Row, w io.Writer) error {
	t := table.NewWriter()

	headers, _ := cast.CastList[interface{}](table_.Columns)

	t.AppendHeader(headers)
	for _, row := range rows {
		var row_ []interface{}
		for _, column := range table_.Columns {
			s := ""
			if v, ok := row.Get(column); ok {
				s = valueToString(v)
			}
			row_ = append(row_, s)
		}
		t.AppendRow(row_)
	}

	if tof.TableFormat == "markdown" {
		s := t.RenderMarkdown()
		_, err := w.Write([]byte(s))
		if err != nil {
			return err
		}
		return nil
	} else if tof.TableFormat == "html" {
		html := t.RenderHTML()
		_, err := w.Write([]byte(html))
		if err != nil {
			return err
		}
		return nil
	} else {
		if tof.TableStyleFile != "" {
			f, err := os.Open(tof.TableStyleFile)
			if err != nil {
				return err
			}
			style, err := styleFromYAML(f)
			if err != nil {
				return err
			}
			t.SetStyle(*style)
		} else {
			t.SetStyle(tof.TableStyle)

			t.Style().Format.Footer = text.FormatDefault
			t.Style().Format.Header = text.FormatDefault
		}
		if tof.PrintTableStyle {
			err := styleToYAML(w, prettyStyleToStyle(t.Style()))
			if err != nil {
				return err
			}
			return nil
		}
		render := t.Render() + "\n"
		_, err := w.Write([]byte(render))
		if err != nil {
			return err
		}
		return nil
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
