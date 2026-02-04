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
	ReplaceFile string            `glazed:"replace-file"`
	AddFields   map[string]string `glazed:"add-fields"`
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

type ReplaceSection struct {
	*schema.SectionImpl `yaml:",inline"`
}

//go:embed "flags/replace.yaml"
var replaceFlagsYaml []byte

func NewReplaceSection(options ...schema.SectionOption) (*ReplaceSection, error) {
	ret := &ReplaceSection{}
	section, err := schema.NewSectionFromYAML(replaceFlagsYaml, options...)
	if err != nil {
		return nil, err
	}
	ret.SectionImpl = section

	return ret, nil
}
func (f *ReplaceSection) Clone() schema.Section {
	return &ReplaceSection{
		SectionImpl: f.SectionImpl.Clone().(*schema.SectionImpl),
	}
}

func NewReplaceSettingsFromValues(glazedValues *values.SectionValues) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := glazedValues.Fields.DecodeInto(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from fields")
	}
	return s, nil
}
