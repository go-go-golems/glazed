package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
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

//go:embed "flags/rename.yaml"
var renameFlagsYaml []byte

var renameParameterLayer *cmds.ParameterLayer

func init() {
	var err error
	renameParameterLayer, err = cmds.NewParameterLayerFromYAML(renameFlagsYaml)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize rename parameter layer"))
	}
}

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

func NewRenameFlagsDefaults() *RenameFlagsDefaults {
	s := &RenameFlagsDefaults{}
	err := renameParameterLayer.InitializeStructFromDefaults(s)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize rename flags defaults"))
	}

	return s
}

func AddRenameFlags(cmd *cobra.Command, defaults *RenameFlagsDefaults) error {
	return renameParameterLayer.AddFlagsToCobraCommand(cmd, defaults)
}

func ParseRenameFlags(cmd *cobra.Command) (*RenameSettings, error) {
	renameFields, _ := cmd.Flags().GetStringSlice("rename")
	renameRegexpFields, _ := cmd.Flags().GetStringSlice("rename-regexp")
	renameYaml, _ := cmd.Flags().GetString("rename-yaml")

	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[types.FieldName(parts[0])] = types.FieldName(parts[1])
	}

	regexpReplacements := middlewares.RegexpReplacements{}
	for _, renameRegexpField := range renameRegexpFields {
		parts := strings.Split(renameRegexpField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename-regexp field: %s", renameRegexpField)
		}
		re, err := regexp.Compile(parts[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid regexp: %s", parts[0])
		}
		regexpReplacements = append(regexpReplacements,
			&middlewares.RegexpReplacement{Regexp: re, Replacement: parts[1]})
	}

	return &RenameSettings{
		RenameFields:  renamesFieldsMap,
		RenameRegexps: regexpReplacements,
		YamlFile:      renameYaml,
	}, nil
}
