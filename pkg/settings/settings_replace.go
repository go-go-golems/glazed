package settings

import (
	_ "embed"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/pkg/errors"
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
	*schema.SectionImpl `yaml:",inline"`
}

//go:embed "flags/replace.yaml"
var replaceFlagsYaml []byte

func NewReplaceParameterLayer(options ...schema.SectionOption) (*ReplaceParameterLayer, error) {
	ret := &ReplaceParameterLayer{}
	layer, err := schema.NewSectionFromYAML(replaceFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret.SectionImpl = layer

	return ret, nil
}
func (f *ReplaceParameterLayer) Clone() schema.Section {
	return &ReplaceParameterLayer{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func NewReplaceSettingsFromParameters(glazedLayer *values.SectionValues) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := glazedLayer.Parameters.InitializeStruct(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from parameters")
	}
	return s, nil
}
