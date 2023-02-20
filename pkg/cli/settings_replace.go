package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

//go:embed "flags/replace.yaml"
var replaceFlagsYaml []byte

var replaceFlagsParameters map[string]*cmds.ParameterDefinition
var replaceFlagsParametersList []*cmds.ParameterDefinition

func init() {
	replaceFlagsParameters, replaceFlagsParametersList = cmds.InitFlagsFromYaml(replaceFlagsYaml)
}

type ReplaceSettings struct {
	ReplaceFile string `glazed.parameter:"replace-file"`
}

func (rs *ReplaceSettings) AddMiddlewares(of formatters.OutputFormatter) error {
	if rs.ReplaceFile != "" {
		b, err := os.ReadFile(rs.ReplaceFile)
		if err != nil {
			return err
		}

		mw, err := middlewares.NewReplaceMiddlewareFromYAML(b)
		if err != nil {
			return err
		}

		of.AddTableMiddleware(mw)
	}

	return nil
}

type ReplaceFlagsDefaults struct {
	// currently, only support loading replacements from a file
	ReplaceFile string `glazed.parameter:"replace-file"`
}

func NewReplaceFlagsDefaults() *ReplaceFlagsDefaults {
	ret := &ReplaceFlagsDefaults{}
	err := cmds.InitializeStructFromParameterDefinitions(ret, replaceFlagsParameters)
	if err != nil {
		panic(err)
	}

	return ret
}

func AddReplaceFlags(cmd *cobra.Command, defaults *ReplaceFlagsDefaults) error {
	parameters, err := cmds.CloneParameterDefinitionsWithDefaultsStruct(replaceFlagsParametersList, defaults)
	if err != nil {
		return errors.Wrap(err, "failed to clone replace parameter definitions")
	}

	err = cmds.AddFlagsToCobraCommand(cmd.PersistentFlags(), parameters)
	if err != nil {
		return errors.Wrap(err, "failed to add replace flags to cobra command")
	}

	cmds.AddFlagGroupToCobraCommand(cmd, "replace", "Glazed replacing", parameters)
	return nil
}

func ParseReplaceFlags(cmd *cobra.Command) (*ReplaceSettings, error) {
	parameters, err := cmds.GatherFlagsFromCobraCommand(cmd, selectFlagsParametersList, false)
	if err != nil {
		return nil, err
	}

	return NewReplaceSettingsFromParameters(parameters)

}

func NewReplaceSettingsFromParameters(parameters map[string]interface{}) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := cmds.InitializeStructFromParameters(s, parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from parameters")
	}
	return s, nil
}
