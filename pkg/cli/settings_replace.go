package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

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

type ReplaceParameterLayer struct {
	layers.ParameterLayer
	Settings *ReplaceSettings
	Defaults *ReplaceFlagsDefaults
}

//go:embed "flags/replace.yaml"
var replaceFlagsYaml []byte

func NewReplaceParameterLayer() (*ReplaceParameterLayer, error) {
	ret := &ReplaceParameterLayer{}
	err := ret.LoadFromYAML(replaceFlagsYaml)
	if err != nil {
		return nil, err
	}
	ret.Defaults = &ReplaceFlagsDefaults{}
	err = ret.InitializeStructFromDefaults(ret.Defaults)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (r *ReplaceParameterLayer) AddFlags(cmd *cobra.Command) error {
	return r.AddFlagsToCobraCommand(cmd, r.Defaults)
}

func (r *ReplaceParameterLayer) ParseFlags(cmd *cobra.Command) error {
	parameters, err := r.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return err
	}

	r.Settings, err = NewReplaceSettingsFromParameters(parameters)
	if err != nil {
		return err
	}
	return nil
}

func NewReplaceSettingsFromParameters(ps map[string]interface{}) (*ReplaceSettings, error) {
	s := &ReplaceSettings{}
	err := parameters.InitializeStructFromParameters(s, ps)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize replace settings from parameters")
	}
	return s, nil
}
