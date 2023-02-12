package cli

import (
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type TemplateSettings struct {
	RenameSeparator string `glazed.parameter:"rename-separator"`
	UseRowTemplates bool   `glazed.parameter:"use-row-templates"`
	Templates       map[types.FieldName]string
}

func (tf *TemplateSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if tf.UseRowTemplates && len(tf.Templates) > 0 {
		middleware, err := middlewares.NewRowGoTemplateMiddleware(tf.Templates, tf.RenameSeparator)
		if err != nil {
			return err
		}
		of.AddTableMiddleware(middleware)
	}

	return nil
}
