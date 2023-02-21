package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
	"strings"
)

type RenameSettings struct {
	RenameFields  map[types.FieldName]string
	RenameRegexps middlewares.RegexpReplacements
	YamlFile      string
}

func (rs *RenameSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if len(rs.RenameFields) > 0 || len(rs.RenameRegexps) > 0 {
		of.AddTableMiddleware(middlewares.NewRenameColumnMiddleware(rs.RenameFields, rs.RenameRegexps))
	}

	if rs.YamlFile != "" {
		f, err := os.Open(rs.YamlFile)
		if err != nil {
			return err
		}
		decoder := yaml.NewDecoder(f)

		mw, err := middlewares.NewRenameColumnMiddlewareFromYAML(decoder)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
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
	layers.ParameterLayerImpl
	Defaults *RenameFlagsDefaults
}

func NewRenameParameterLayer() (*RenameParameterLayer, error) {
	ret := &RenameParameterLayer{}
	err := ret.LoadFromYAML(renameFlagsYaml)
	if err != nil {
		return nil, err
	}
	ret.Defaults = &RenameFlagsDefaults{}
	err = ret.InitializeStructFromDefaults(ret.Defaults)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize rename flags defaults"))
	}

	return ret, nil
}

func (r *RenameParameterLayer) AddFlagsToCobraCommand(cmd *cobra.Command, s interface{}) error {
	if s == nil {
		s = r.Defaults
	}
	return r.ParameterLayerImpl.AddFlagsToCobraCommand(cmd, s)
}

func NewRenameSettingsFromParameters(ps map[string]interface{}) (*RenameSettings, error) {
	renameFields, ok := ps["rename"].([]string)
	if !ok {
		return nil, errors.Errorf("Invalid rename fields")
	}
	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[parts[0]] = parts[1]
	}

	regexpReplacements := middlewares.RegexpReplacements{}
	renameRegexpFields, ok := ps["rename-regexp"].(map[string]string)
	if !ok {
		return nil, errors.Errorf("Invalid rename regexp fields")
	}
	for regex, replacement := range renameRegexpFields {
		re, err := regexp.Compile(regex)
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid regexp: %s", regex)
		}
		regexpReplacements = append(regexpReplacements,
			&middlewares.RegexpReplacement{Regexp: re, Replacement: replacement})
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
