package formatters

import (
	"fmt"
	middlewares2 "github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
	"strings"
	"text/template"
)

type SelectOutputFormatter struct {
	Table          *types.Table
	middlewares    []middlewares2.TableMiddleware
	Field          string
	IsTemplate     bool
	QuoteCharacter string
	JoinCharacter  string
}

func NewSelectOutputFormatter(field string) *SelectOutputFormatter {
	return &SelectOutputFormatter{
		Table:          types.NewTable(),
		middlewares:    []middlewares2.TableMiddleware{},
		Field:          field,
		IsTemplate:     false,
		QuoteCharacter: "",
		JoinCharacter:  " ",
	}
}

func (f *SelectOutputFormatter) SetIsTemplate(isTemplate bool) {
	f.IsTemplate = isTemplate
}

func (f *SelectOutputFormatter) SetQuoteCharacter(quoteCharacter string) {
	f.QuoteCharacter = quoteCharacter
}

func (f *SelectOutputFormatter) SetJoinCharacter(joinCharacter string) {
	f.JoinCharacter = joinCharacter
}

func (f *SelectOutputFormatter) AddTableMiddleware(m middlewares2.TableMiddleware) {
	f.middlewares = append(f.middlewares, m)
}

func (f *SelectOutputFormatter) AddRow(row types.Row) {
	f.Table.Rows = append(f.Table.Rows, row)
}

func (f *SelectOutputFormatter) Output() (string, error) {
	for _, middleware := range f.middlewares {
		newTable, err := middleware.Process(f.Table)
		if err != nil {
			return "", err
		}
		f.Table = newTable
	}

	var t *template.Template
	var err error
	if f.IsTemplate {
		t = template.New("select")
		t, err = t.Parse(f.Field)
		if err != nil {
			return "", err
		}
	}

	outputs := make([]string, 0)

	for _, row := range f.Table.Rows {
		values := row.GetValues()
		if f.IsTemplate {
			sb := strings.Builder{}
			sb.WriteString(f.QuoteCharacter)
			err := t.Execute(&sb, values)
			if err != nil {
				return "", err
			}
			sb.WriteString(f.QuoteCharacter)
			outputs = append(outputs, sb.String())
		} else {
			v, ok := values[f.Field]
			if ok {
				outputs = append(outputs, fmt.Sprintf("%s%v%s", f.QuoteCharacter, v, f.QuoteCharacter))
			}
		}
	}

	return strings.Join(outputs, f.JoinCharacter), nil
}
