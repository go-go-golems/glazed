package table

import (
	"fmt"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"gopkg.in/yaml.v3"
	"io"
)

type Style struct {
	Name    string        `yaml:"name"`
	Box     BoxStyle      `yaml:"box"`
	Color   ColorOptions  `yaml:"color"`
	Format  FormatOptions `yaml:"format"`
	Options Options       `yaml:"options"`
	Title   TitleOptions  `yaml:"title"`
}

type TitleOptions struct {
	Align  string `yaml:"align,omitempty"`
	Colors Colors `yaml:"colors,omitempty"`
	Format string `yaml:"format,omitempty"`
}

var alignStringToAlign = map[string]text.Align{
	"center":  text.AlignCenter,
	"left":    text.AlignLeft,
	"right":   text.AlignRight,
	"justify": text.AlignJustify,
	"default": text.AlignDefault,
}

type BoxStyle struct {
	BottomLeft       string `yaml:"bottom-left,omitempty,flow"`
	BottomRight      string `yaml:"bottom-right,omitempty,flow"`
	BottomSeparator  string `yaml:"bottom-separator,omitempty,flow"`
	Left             string `yaml:"left,omitempty,flow"`
	LeftSeparator    string `yaml:"left-separator,omitempty,flow"`
	MiddleHorizontal string `yaml:"middle-horizontal,omitempty,flow"`
	MiddleSeparator  string `yaml:"middle-separator,omitempty,flow"`
	MiddleVertical   string `yaml:"middle-vertical,omitempty,flow"`
	PaddingLeft      string `yaml:"padding-left,omitempty,flow"`
	PaddingRight     string `yaml:"padding-right,omitempty,flow"`
	PageSeparator    string `yaml:"page-separator,omitempty,flow,flow"`
	Right            string `yaml:"right,omitempty,flow"`
	RightSeparator   string `yaml:"right-separator,omitempty,flow"`
	TopLeft          string `yaml:"top-left,omitempty,flow"`
	TopRight         string `yaml:"top-right,omitempty,flow"`
	TopSeparator     string `yaml:"top-separator,omitempty,flow"`
	UnfinishedRow    string `yaml:"unfinished-row,omitempty,flow"`
}

type FormatOptions struct {
	Footer string `yaml:"footer,omitempty,flow"`
	Header string `yaml:"header,omitempty,flow"`
	Row    string `yaml:"row,omitempty,flow"`
}

type Options struct {
	DrawBorder      bool `yaml:"draw-border,omitempty"`
	SeparateColumns bool `yaml:"separate-columns,omitempty"`
	SeparateFooter  bool `yaml:"separate-footer,omitempty"`
	SeparateHeader  bool `yaml:"separate-header,omitempty"`
	SeparateRows    bool `yaml:"separate-rows,omitempty"`
}

var formatStringToFormat = map[string]text.Format{
	"default": text.FormatDefault,
	"lower":   text.FormatLower,
	"title":   text.FormatTitle,
	"upper":   text.FormatUpper,
}

type ColorOptions struct {
	IndexColumn  Colors `yaml:"index-column,omitempty"`
	Footer       Colors `yaml:"footer,omitempty"`
	Header       Colors `yaml:"header,omitempty"`
	Row          Colors `yaml:"row,omitempty"`
	RowAlternate Colors `yaml:"row-alternate,omitempty"`
}

type Colors []Color
type Color string

func colorStringListToColors(colorStrings []Color) (text.Colors, bool) {
	if len(colorStrings) == 0 {
		return nil, true
	}
	ret := make(text.Colors, 0)
	for _, colorString := range colorStrings {
		color, ok := colorStringToColor[colorString]
		if !ok {
			return nil, false
		}
		ret = append(ret, color)
	}
	return ret, true
}

var colorStringToColor = map[Color]text.Color{
	// base colors
	"reset":         text.Reset,
	"bold":          text.Bold,
	"faint":         text.Faint,
	"italic":        text.Italic,
	"Underline":     text.Underline,
	"blink-slow":    text.BlinkSlow,
	"blink-rapid":   text.BlinkRapid,
	"reverse-video": text.ReverseVideo,
	"concealed":     text.Concealed,
	"crossed-out":   text.CrossedOut,

	// Foreground colors
	"fg-black":   text.FgBlack,
	"fg-red":     text.FgRed,
	"fg-green":   text.FgGreen,
	"fg-yellow":  text.FgYellow,
	"fg-blue":    text.FgBlue,
	"fg-magenta": text.FgMagenta,
	"fg-cyan":    text.FgCyan,
	"fg-white":   text.FgWhite,

	// Foreground Hi-Intensity colors
	"fg-hi-black":   text.FgHiBlack,
	"fg-hi-red":     text.FgHiRed,
	"fg-hi-green":   text.FgHiGreen,
	"fg-hi-yellow":  text.FgHiYellow,
	"fg-hi-blue":    text.FgHiBlue,
	"fg-hi-magenta": text.FgHiMagenta,
	"fg-hi-cyan":    text.FgHiCyan,
	"fg-hi-white":   text.FgHiWhite,

	// Background colors
	"bg-black":   text.BgBlack,
	"bg-red":     text.BgRed,
	"bg-green":   text.BgGreen,
	"bg-yellow":  text.BgYellow,
	"bg-blue":    text.BgBlue,
	"bg-magenta": text.BgMagenta,
	"bg-cyan":    text.BgCyan,
	"bg-white":   text.BgWhite,

	// Background Hi-Intensity colors
	"bg-hi-black":   text.BgHiBlack,
	"bg-hi-red":     text.BgHiRed,
	"bg-hi-green":   text.BgHiGreen,
	"bg-hi-yellow":  text.BgHiYellow,
	"bg-hi-blue":    text.BgHiBlue,
	"bg-hi-magenta": text.BgHiMagenta,
	"bg-hi-cyan":    text.BgHiCyan,
	"bg-hi-white":   text.BgHiWhite,
}

func styleFromYAML(r io.Reader) (*table.Style, error) {
	var style Style
	err := yaml.NewDecoder(r).Decode(&style)
	if err != nil {
		return nil, err
	}

	indexColumnColors, ok := colorStringListToColors(style.Color.IndexColumn)
	if !ok {
		return nil, fmt.Errorf("invalid index column colors: %v", style.Color.IndexColumn)
	}
	footerColors, ok := colorStringListToColors(style.Color.Footer)
	if !ok {
		return nil, fmt.Errorf("invalid footer colors: %v", style.Color.Footer)
	}
	headerColors, ok := colorStringListToColors(style.Color.Header)
	if !ok {
		return nil, fmt.Errorf("invalid header colors: %v", style.Color.Header)
	}
	rowColors, ok := colorStringListToColors(style.Color.Row)
	if !ok {
		return nil, fmt.Errorf("invalid row colors: %v", style.Color.Row)
	}
	rowAlternateColors, ok := colorStringListToColors(style.Color.RowAlternate)
	if !ok {
		return nil, fmt.Errorf("invalid row alternate colors: %v", style.Color.RowAlternate)
	}

	footerFormat, ok := formatStringToFormat[style.Format.Footer]
	if !ok {
		return nil, fmt.Errorf("invalid footer format: %v", style.Format.Footer)
	}
	headerFormat, ok := formatStringToFormat[style.Format.Header]
	if !ok {
		return nil, fmt.Errorf("invalid header format: %v", style.Format.Header)
	}
	rowFormat, ok := formatStringToFormat[style.Format.Row]
	if !ok {
		return nil, fmt.Errorf("invalid row format: %v", style.Format.Row)
	}

	titleAlign, ok := alignStringToAlign[style.Title.Align]
	if !ok {
		return nil, fmt.Errorf("invalid title align: %v", style.Title.Align)
	}
	titleColors, ok := colorStringListToColors(style.Title.Colors)
	if !ok {
		return nil, fmt.Errorf("invalid title colors: %v", style.Title.Colors)
	}
	titleFormat, ok := formatStringToFormat[style.Title.Format]
	if !ok {
		return nil, fmt.Errorf("invalid title format: %v", style.Title.Format)
	}

	ret := &table.Style{
		Name: style.Name,
		Box: table.BoxStyle{
			BottomLeft:       style.Box.BottomLeft,
			BottomRight:      style.Box.BottomRight,
			BottomSeparator:  style.Box.BottomSeparator,
			Left:             style.Box.Left,
			LeftSeparator:    style.Box.LeftSeparator,
			MiddleHorizontal: style.Box.MiddleHorizontal,
			MiddleSeparator:  style.Box.MiddleSeparator,
			MiddleVertical:   style.Box.MiddleVertical,
			PaddingLeft:      style.Box.PaddingLeft,
			PaddingRight:     style.Box.PaddingRight,
			PageSeparator:    style.Box.PageSeparator,
			Right:            style.Box.Right,
			RightSeparator:   style.Box.RightSeparator,
			TopLeft:          style.Box.TopLeft,
			TopRight:         style.Box.TopRight,
			TopSeparator:     style.Box.TopSeparator,
			UnfinishedRow:    style.Box.UnfinishedRow,
		},
		Color: table.ColorOptions{
			Footer:       footerColors,
			Header:       headerColors,
			IndexColumn:  indexColumnColors,
			Row:          rowColors,
			RowAlternate: rowAlternateColors,
		},
		Format: table.FormatOptions{
			Footer: footerFormat,
			Header: headerFormat,
			Row:    rowFormat,
		},
		Options: table.Options{
			DrawBorder:      style.Options.DrawBorder,
			SeparateColumns: style.Options.SeparateColumns,
			SeparateFooter:  style.Options.SeparateFooter,
			SeparateHeader:  style.Options.SeparateHeader,
			SeparateRows:    style.Options.SeparateRows,
		},
		Title: table.TitleOptions{
			Align:  titleAlign,
			Colors: titleColors,
			Format: titleFormat,
		},
	}

	return ret, nil
}

func prettyStyleToStyle(style *table.Style) *Style {
	return &Style{
		Name: style.Name,
		Box: BoxStyle{
			BottomLeft:       style.Box.BottomLeft,
			BottomRight:      style.Box.BottomRight,
			BottomSeparator:  style.Box.BottomSeparator,
			Left:             style.Box.Left,
			LeftSeparator:    style.Box.LeftSeparator,
			MiddleHorizontal: style.Box.MiddleHorizontal,
			MiddleSeparator:  style.Box.MiddleSeparator,
			MiddleVertical:   style.Box.MiddleVertical,
			PaddingLeft:      style.Box.PaddingLeft,
			PaddingRight:     style.Box.PaddingRight,
			PageSeparator:    style.Box.PageSeparator,
			Right:            style.Box.Right,
			RightSeparator:   style.Box.RightSeparator,
			TopLeft:          style.Box.TopLeft,
			TopRight:         style.Box.TopRight,
			TopSeparator:     style.Box.TopSeparator,
			UnfinishedRow:    style.Box.UnfinishedRow,
		},
		Color: ColorOptions{
			Footer:       colorsToColorStringList(style.Color.Footer),
			Header:       colorsToColorStringList(style.Color.Header),
			IndexColumn:  colorsToColorStringList(style.Color.IndexColumn),
			Row:          colorsToColorStringList(style.Color.Row),
			RowAlternate: colorsToColorStringList(style.Color.RowAlternate),
		},
		Format: FormatOptions{
			Footer: formatToFormatString(style.Format.Footer),
			Header: formatToFormatString(style.Format.Header),
			Row:    formatToFormatString(style.Format.Row),
		},
		Options: Options{
			DrawBorder:      style.Options.DrawBorder,
			SeparateColumns: style.Options.SeparateColumns,
			SeparateFooter:  style.Options.SeparateFooter,
			SeparateHeader:  style.Options.SeparateHeader,
			SeparateRows:    style.Options.SeparateRows,
		},
		Title: TitleOptions{
			Align:  alignToAlignString(style.Title.Align),
			Colors: colorsToColorStringList(style.Title.Colors),
			Format: formatToFormatString(style.Title.Format),
		},
	}
}

func styleToYAML(w io.Writer, style *Style) error {
	enc := yaml.NewEncoder(w)
	return enc.Encode(style)
}

func alignToAlignString(align text.Align) string {
	for k, v := range alignStringToAlign {
		if v == align {
			return k
		}
	}
	return ""
}

func formatToFormatString(format text.Format) string {
	for k, v := range formatStringToFormat {
		if v == format {
			return k
		}
	}
	return ""
}

func colorsToColorStringList(colors text.Colors) Colors {
	if colors == nil {
		return nil
	}
	ret := make(Colors, len(colors))
	for i, c := range colors {
		ret[i] = colorToColorString(c)
	}
	return ret
}

func colorToColorString(color text.Color) Color {
	for k, v := range colorStringToColor {
		if v == color {
			return k
		}
	}
	return ""
}
