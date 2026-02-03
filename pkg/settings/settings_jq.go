package settings

import (
	_ "embed"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
)

type JqSettings struct {
	JqExpression       string            `glazed:"jq"`
	JqFile             string            `glazed:"jq-file"`
	JqFieldExpressions map[string]string `glazed:"field-jq"`
}

//go:embed "flags/jq.yaml"
var jqFlagsYaml []byte

type JqParameterLayer struct {
	*schema.SectionImpl `yaml:",inline"`
}

func NewJqParameterLayer(options ...schema.SectionOption) (*JqParameterLayer, error) {
	ret := &JqParameterLayer{}
	layer, err := schema.NewSectionFromYAML(jqFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create jq parameter layer")
	}
	ret.SectionImpl = layer

	return ret, nil
}
func (f *JqParameterLayer) Clone() schema.Section {
	return &JqParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func NewJqSettingsFromParameters(glazedLayer *values.SectionValues) (*JqSettings, error) {
	s := &JqSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize jq settings from parameters")
	}

	return s, nil
}

func NewJqMiddlewaresFromSettings(settings *JqSettings) (*middlewares.JqObjectMiddleware, *middlewares.JqTableMiddleware, error) {
	var jqObjectMiddleware *middlewares.JqObjectMiddleware
	var jqTableMiddleware *middlewares.JqTableMiddleware
	var err error

	jqExpression := settings.JqExpression
	if jqExpression == "" {
		jqExpression = settings.JqFile
	}

	if jqExpression != "" {
		jqObjectMiddleware, err = middlewares.NewJqObjectMiddleware(settings.JqExpression)

		if err != nil {
			return nil, nil, err
		}
	}

	if len(settings.JqFieldExpressions) > 0 {
		jqTableMiddleware, err = middlewares.NewJqTableMiddleware(settings.JqFieldExpressions)
		if err != nil {
			return nil, nil, err
		}
	}

	return jqObjectMiddleware, jqTableMiddleware, nil
}
