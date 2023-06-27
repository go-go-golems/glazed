package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
	"os"
)

type ReplaceSettings struct {
	ReplaceFile string            `glazed.parameter:"replace-file"`
	AddFields   map[string]string `glazed.parameter:"add-fields"`
}

func (rs *ReplaceSettings) AddMiddlewares(of *middlewares.Processor) error {
	if rs.ReplaceFile != "" {
		b, err := os.ReadFile(rs.ReplaceFile)
		if err != nil {
			return err
		}

		mw, err := table.NewReplaceMiddlewareFromYAML(b)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
	}

	if len(rs.AddFields) > 0 {
		mw := table.NewAddFieldMiddleware(rs.AddFields)
		of.AddTableMiddleware(mw)
	}

	return nil
}

type ReplaceParameterLayer struct {
	*layers.ParameterLayerImpl
}

//go:embed "flags/replace.yaml"
var replaceFlagsYaml []byte

func NewReplaceParameterLayer(options ...layers.ParameterLayerOptions) (*ReplaceParameterLayer, error) {
	ret := &ReplaceParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(replaceFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret.ParameterLayerImpl = layer

	return ret, nil
}

func NewReplaceSettingsFromParameters(ps map[string]interface{}) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from parameters")
	}
	return s, nil
}
