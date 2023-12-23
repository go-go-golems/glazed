package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
)

type JqSettings struct {
	JqExpression       string            `glazed.parameter:"jq"`
	JqFile             string            `glazed.parameter:"jq-file"`
	JqFieldExpressions map[string]string `glazed.parameter:"field-jq"`
}

//go:embed "flags/jq.yaml"
var jqFlagsYaml []byte

type JqParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

func NewJqParameterLayer(options ...layers.ParameterLayerOptions) (*JqParameterLayer, error) {
	ret := &JqParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(jqFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create jq parameter layer")
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func NewJqSettingsFromParameters(glazedLayer *layers.ParsedLayer) (*JqSettings, error) {
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
