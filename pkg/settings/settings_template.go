package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
)

type TemplateSettings struct {
	RenameSeparator string
	UseRowTemplates bool `glazed:"use-row-templates"`
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
	Template        string            `glazed:"template"`
	TemplateField   map[string]string `glazed:"template-field"`
	UseRowTemplates bool              `glazed:"use-row-templates"`
}

func NewTemplateFlagsDefaults() *TemplateFlagsDefaults {
	return &TemplateFlagsDefaults{
		UseRowTemplates: false,
	}
}

type TemplateParameterLayer struct {
	*schema.SectionImpl `yaml:",inline"`
}

const GlazedTemplateLayerSlug = "glazed-template"

func NewTemplateParameterLayer(options ...schema.SectionOption) (*TemplateParameterLayer, error) {
	ret := &TemplateParameterLayer{}
	layer, err := schema.NewSectionFromYAML(templateFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create template parameter layer")
	}
	ret.SectionImpl = layer

	return ret, nil
}

func (f *TemplateParameterLayer) Clone() schema.Section {
	return &TemplateParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func NewTemplateSettings(layer *values.SectionValues) (*TemplateSettings, error) {
	//TODO(manuel, 2024-01-05) This could better be done with a InitializeStruct I think

	// templates get applied before flattening
	templates := map[types.FieldName]string{}

	templateArgument, ok := layer.Parameters.GetValue("template").(string)
	if ok && templateArgument != "" {
		templates["_0"] = templateArgument
	} else {
		v := layer.Parameters.GetValue("template-field")
		templateFields, err := cast.ConvertMapToInterfaceMap(v)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to convert template-field to map[string]interface{}")
		}
		for k, v := range templateFields {
			vString, ok := v.(string)
			if !ok {
				return nil, errors.Errorf("template-field %s is not a string", k)
			}
			templates[k] = vString
		}
	}

	useRowTemplates, ok := layer.Parameters.GetValue("use-row-templates").(bool)
	if !ok {
		useRowTemplates = false
	}

	return &TemplateSettings{
		Templates:       templates,
		UseRowTemplates: useRowTemplates,
		RenameSeparator: "_",
	}, nil
}
