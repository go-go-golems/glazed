package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/pkg/errors"
	"os"
)

type ReplaceSettings struct {
	ReplaceFile string            `glazed.parameter:"replace-file"`
	AddFields   map[string]string `glazed.parameter:"add-fields"`
}

func (rs *ReplaceSettings) AddMiddlewares(of *middlewares.TableProcessor) error {
	if rs.ReplaceFile != "" {
		b, err := os.ReadFile(rs.ReplaceFile)
		if err != nil {
			return err
		}

		mw, err := row.NewReplaceMiddlewareFromYAML(b)
		if err != nil {
			return err
		}

		of.AddRowMiddleware(mw)
	}

	if len(rs.AddFields) > 0 {
		mw := row.NewAddFieldMiddleware(rs.AddFields)
		of.AddRowMiddleware(mw)
	}

	return nil
}

type ReplaceParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
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
func (f *ReplaceParameterLayer) Clone() layers.ParameterLayer {
	return &ReplaceParameterLayer{
		ParameterLayerImpl: f.ParameterLayerImpl.Clone().(*layers.ParameterLayerImpl),
	}
}

func NewReplaceSettingsFromParameters(glazedLayer *layers.ParsedLayer) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from parameters")
	}
	return s, nil
}
