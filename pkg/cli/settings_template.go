package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type TemplateSettings struct {
	RenameSeparator string
	UseRowTemplates bool `glazed.parameter:"use-row-templates"`
	Templates       map[types.FieldName]string
}

//go:embed "flags/template.yaml"
var templateFlagsYaml []byte

var templateFlagsParameters map[string]*cmds.ParameterDefinition
var templateFlagsParametersList []*cmds.ParameterDefinition

func init() {
	templateFlagsParameters, templateFlagsParametersList = cmds.InitFlagsFromYaml(templateFlagsYaml)
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

type TemplateFlagsDefaults struct {
	Template        string   `glazed.parameter:"template"`
	TemplateField   []string `glazed.parameter:"template-field"`
	UseRowTemplates bool     `glazed.parameter:"use-row-templates"`
}

func NewTemplateFlagsDefaults() *TemplateFlagsDefaults {
	s := &TemplateFlagsDefaults{}
	err := cmds.InitializeStructFromParameterDefinitions(s, templateFlagsParameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize template flags defaults"))
	}

	return s
}

func AddTemplateFlags(cmd *cobra.Command, defaults *TemplateFlagsDefaults) error {
	parameters, err := cmds.CloneParameterDefinitionsWithDefaultsStruct(templateFlagsParametersList, defaults)
	if err != nil {
		return errors.Wrap(err, "Failed to clone template flags parameters")
	}
	err = cmds.AddFlagsToCobraCommand(cmd, parameters)
	if err != nil {
		return errors.Wrap(err, "Failed to add template flags to cobra command")
	}

	return nil
}

func NewTemplateSettings(parameters map[string]interface{}) (*TemplateSettings, error) {
	// templates get applied before flattening
	templates := map[types.FieldName]string{}

	templateArgument, ok := parameters["template"].(string)
	if ok && templateArgument != "" {
		templates["_0"] = templateArgument
	} else {
		templateFields, ok := parameters["template-field"].([]string)
		if ok && len(templateFields) > 0 {
			p := templateFlagsParameters["template-field"]
			v, ok := helpers.CastMapMember[[]string](parameters, p.Name)
			if !ok {
				return nil, errors.Errorf("template-field parameter not found")
			}
			kv, err := p.ParseParameter(*v)
			if err != nil {
				return nil, err
			}
			for k, v := range kv.(map[string]interface{}) {
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

func ParseTemplateFlags(cmd *cobra.Command) (*TemplateSettings, error) {
	parameters, err := cmds.GatherFlagsFromCobraCommand(cmd, templateFlagsParametersList, false)
	if err != nil {
		return nil, err
	}

	return NewTemplateSettings(parameters)
}
