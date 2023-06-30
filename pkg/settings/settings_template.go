package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type TemplateSettings struct {
	RenameSeparator string
	UseRowTemplates bool `glazed.parameter:"use-row-templates"`
	Templates       map[types.FieldName]string
}

//go:embed "flags/template.yaml"
var templateFlagsYaml []byte

func (tf *TemplateSettings) AddMiddlewares(p_ *middlewares.TableProcessor) error {
	if tf.UseRowTemplates && len(tf.Templates) > 0 {
		middleware, err := row.NewTemplateMiddleware(tf.Templates, tf.RenameSeparator)
		if err != nil {
			return err
		}
		p_.AddRowMiddleware(middleware)
	}

	return nil
}

type TemplateFlagsDefaults struct {
	Template        string            `glazed.parameter:"template"`
	TemplateField   map[string]string `glazed.parameter:"template-field"`
	UseRowTemplates bool              `glazed.parameter:"use-row-templates"`
}

func NewTemplateFlagsDefaults() *TemplateFlagsDefaults {
	return &TemplateFlagsDefaults{
		UseRowTemplates: false,
	}
}

type TemplateParameterLayer struct {
	*layers.ParameterLayerImpl
}

func NewTemplateParameterLayer(options ...layers.ParameterLayerOptions) (*TemplateParameterLayer, error) {
	ret := &TemplateParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(templateFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create template parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func NewTemplateSettings(parameters map[string]interface{}) (*TemplateSettings, error) {
	// templates get applied before flattening
	templates := map[types.FieldName]string{}

	templateArgument, ok := parameters["template"].(string)
	if ok && templateArgument != "" {
		templates["_0"] = templateArgument
	} else {
		templateFields, ok := parameters["template-field"].(map[string]interface{})
		if ok && len(templateFields) > 0 {
			for k, v := range templateFields {
				vString, ok := v.(string)
				if !ok {
					return nil, errors.Errorf("template-field %s is not a string", k)
				}
				templates[k] = vString
			}
		}
	}

	useRowTemplates, ok := parameters["use-row-templates"].(bool)
	if !ok {
		useRowTemplates = false
	}

	return &TemplateSettings{
		Templates:       templates,
		UseRowTemplates: useRowTemplates,
		RenameSeparator: "_",
	}, nil
}
