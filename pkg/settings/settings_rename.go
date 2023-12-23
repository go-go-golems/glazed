package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type RenameSettings struct {
	RenameFields  map[types.FieldName]string
	RenameRegexps row.RegexpReplacements
	YamlFile      string
}

func (rs *RenameSettings) AddMiddlewares(processor *middlewares.TableProcessor) error {
	if len(rs.RenameFields) > 0 || len(rs.RenameRegexps) > 0 {
		processor.AddRowMiddleware(row.NewRenameColumnMiddleware(rs.RenameFields, rs.RenameRegexps))
	}

	if rs.YamlFile != "" {
		f, err := os.Open(rs.YamlFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(f)

		mw, err := row.NewRenameColumnMiddlewareFromYAML(decoder)
		if err != nil {
			return err
		}

		processor.AddRowMiddleware(mw)
	}

	return nil
}

type RenameFlagsDefaults struct {
	Rename       []string          `glazed.parameter:"rename"`
	RenameRegexp map[string]string `glazed.parameter:"rename-regexp"`
	RenameYaml   string            `glazed.parameter:"rename-yaml"`
}

//go:embed "flags/rename.yaml"
var renameFlagsYaml []byte

type RenameParameterLayer struct {
	*layers.ParameterLayerImpl `yaml:",inline"`
}

func NewRenameParameterLayer(options ...layers.ParameterLayerOptions) (*RenameParameterLayer, error) {
	ret := &RenameParameterLayer{}
	layer, err := layers.NewParameterLayerFromYAML(renameFlagsYaml, options...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create rename parameter layer")
	}
	ret.ParameterLayerImpl = layer
	return ret, nil
}

func NewRenameSettingsFromParameters(glazedLayer *layers.ParsedLayer) (*RenameSettings, error) {
	ps := glazedLayer.Parameters
	rename := ps.GetValue("rename")
	if rename == nil {
		return &RenameSettings{
			RenameFields:  map[types.FieldName]string{},
			RenameRegexps: row.RegexpReplacements{},
		}, nil
	}

	renameFields, ok := cast.CastList2[string, interface{}](rename)
	if !ok {
		return nil, errors.Errorf("Invalid rename fields %s", rename)
	}
	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[parts[0]] = parts[1]
	}

	regexpReplacements := row.RegexpReplacements{}
	renameRegexpFields, ok := ps.GetValue("rename-regexp").(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("Invalid rename regexp fields")
	}
	for regex, replacement := range renameRegexpFields {
		replacement_, ok := replacement.(string)
		if !ok {
			return nil, errors.Errorf("Invalid rename regexp replacement: %s", replacement)
		}
		re, err := regexp.Compile(regex)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid regexp: %s", regex)
		}
		regexpReplacements = append(regexpReplacements,
			&row.RegexpReplacement{Regexp: re, Replacement: replacement_})
	}

	renameYaml, ok := ps.GetValue("rename-yaml").(string)
	if !ok {
		return nil, errors.Errorf("Invalid rename yaml")
	}

	return &RenameSettings{
		RenameFields:  renamesFieldsMap,
		RenameRegexps: regexpReplacements,
		YamlFile:      renameYaml,
	}, nil

}
