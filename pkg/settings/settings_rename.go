package settings

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type RenameSettings struct {
	RenameFields  map[types.FieldName]string
	RenameRegexps table.RegexpReplacements
	YamlFile      string
}

func (rs *RenameSettings) AddMiddlewares(processor *middlewares.Processor) error {
	if len(rs.RenameFields) > 0 || len(rs.RenameRegexps) > 0 {
		processor.AddTableMiddleware(table.NewRenameColumnMiddleware(rs.RenameFields, rs.RenameRegexps))
	}

	if rs.YamlFile != "" {
		f, err := os.Open(rs.YamlFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(f)

		mw, err := table.NewRenameColumnMiddlewareFromYAML(decoder)
		if err != nil {
			return err
		}

		processor.AddTableMiddleware(mw)
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
	*layers.ParameterLayerImpl
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

func NewRenameSettingsFromParameters(ps map[string]interface{}) (*RenameSettings, error) {
	if ps["rename"] == nil {
		return &RenameSettings{
			RenameFields:  map[types.FieldName]string{},
			RenameRegexps: table.RegexpReplacements{},
		}, nil
	}

	renameFields, ok := cast.CastList2[string, interface{}](ps["rename"])
	if !ok {
		return nil, errors.Errorf("Invalid rename fields %s", ps["rename"])
	}
	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[parts[0]] = parts[1]
	}

	regexpReplacements := table.RegexpReplacements{}
	renameRegexpFields, ok := ps["rename-regexp"].(map[string]interface{})
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
			&table.RegexpReplacement{Regexp: re, Replacement: replacement_})
	}

	renameYaml, ok := ps["rename-yaml"].(string)
	if !ok {
		return nil, errors.Errorf("Invalid rename yaml")
	}

	return &RenameSettings{
		RenameFields:  renamesFieldsMap,
		RenameRegexps: regexpReplacements,
		YamlFile:      renameYaml,
	}, nil

}
