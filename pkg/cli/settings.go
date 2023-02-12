package cli

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"gopkg.in/yaml.v3"
	"os"
)

type FieldsFilterSettings struct {
	Filters        []string
	Fields         []string
	SortColumns    bool
	ReorderColumns []string
}

func (fff *FieldsFilterSettings) AddMiddlewares(of formatters.OutputFormatter) {
	of.AddTableMiddleware(middlewares.NewFieldsFilterMiddleware(fff.Fields, fff.Filters))
	if fff.SortColumns {
		of.AddTableMiddleware(middlewares.NewSortColumnsMiddleware())
	}
	if len(fff.ReorderColumns) > 0 {
		of.AddTableMiddleware(middlewares.NewReorderColumnOrderMiddleware(fff.ReorderColumns))
	}
}

type SelectSettings struct {
	SelectField    string
	SelectTemplate string
}

func (ofs *OutputFormatterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" || ss.SelectTemplate != "" {
		ofs.Output = "table"
		ofs.TableFormat = "tsv"
		ofs.FlattenObjects = true
		ofs.WithHeaders = false
	}
}

func (ffs *FieldsFilterSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectField != "" {
		ffs.Fields = []string{ss.SelectField}
	}
}

func (tf *TemplateSettings) UpdateWithSelectSettings(ss *SelectSettings) {
	if ss.SelectTemplate != "" {
		tf.Templates = map[types.FieldName]string{
			"_0": ss.SelectTemplate,
		}
	}
}

type RenameSettings struct {
	RenameFields  map[types.FieldName]string
	RenameRegexps middlewares.RegexpReplacements
	YamlFile      string
}

func (rs *RenameSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if len(rs.RenameFields) > 0 || len(rs.RenameRegexps) > 0 {
		of.AddTableMiddleware(middlewares.NewRenameColumnMiddleware(rs.RenameFields, rs.RenameRegexps))
	}

	if rs.YamlFile != "" {
		f, err := os.Open(rs.YamlFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(f)

		mw, err := middlewares.NewRenameColumnMiddlewareFromYAML(decoder)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
	}

	return nil
}

type ReplaceSettings struct {
	ReplaceFile string
}

func (rs *ReplaceSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if rs.ReplaceFile != "" {
		b, err := os.ReadFile(rs.ReplaceFile)
		if err != nil {
			return err
		}

		mw, err := middlewares.NewReplaceMiddlewareFromYAML(b)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
	}

	return nil
}
