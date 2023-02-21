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

var replaceParameterLayer *cmds.ParameterLayer

func init() {
	var err error
	replaceParameterLayer, err = cmds.NewParameterLayerFromYAML(replaceFlagsYaml)
	if err != nil {
		panic(errors.Wrap(err, "Failed to initialize replace parameter layer"))
	}
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
	err := replaceParameterLayer.InitializeStructFromDefaults(ret)
	if err != nil {
		panic(err)
	}

	return ret
}

func AddReplaceFlags(cmd *cobra.Command, defaults *ReplaceFlagsDefaults) error {
	return replaceParameterLayer.AddFlagsToCobraCommand(cmd, defaults)
}

func ParseReplaceFlags(cmd *cobra.Command) (*ReplaceSettings, error) {
	parameters, err := replaceParameterLayer.ParseFlagsFromCobraCommand(cmd)
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
