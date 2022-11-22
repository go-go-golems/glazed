package cli

import (
	"dd-cli/pkg/formatters"
	"dd-cli/pkg/middlewares"
	"dd-cli/pkg/types"
)

type OutputFormatterSettings struct {
	Output          string
	TableFormat     string
	OutputFormatter formatters.OutputFormatter
	OutputAsObjects bool
	FlattenObjects  bool
}

type TemplateSettings struct {
	UseRowTemplates bool
	Templates       map[types.FieldName]string
}

func (tf *TemplateSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if tf.UseRowTemplates {
		middleware, err := middlewares.NewRowGoTemplateMiddleware(tf.Templates)
		if err != nil {
			return err
		}
		of.AddTableMiddleware(middleware)
	} else {
	}

	return nil
}

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
