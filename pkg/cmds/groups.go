package cmds

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.

type ParameterLayer struct {
	Name        string                 `yaml:"name"`
	Slug        string                 `yaml:"slug"`
	Description string                 `yaml:"description"`
	Flags       []*ParameterDefinition `yaml:"flags,omitempty"`
}

func NewParameterLayerFromYAML(s []byte) (*ParameterLayer, error) {
	ret := &ParameterLayer{}

	err := yaml.Unmarshal(s, ret)
	if err != nil {
		return nil, err
	}

	for _, p := range ret.Flags {
		err := p.CheckParameterDefaultValueValidity()
		if err != nil {
			panic(errors.Wrap(err, "Failed to check parameter default value validity"))
		}
	}

	return ret, nil
}

func (p *ParameterLayer) AddFlag(flag *ParameterDefinition) {
	p.Flags = append(p.Flags, flag)
}

// GetParameters returns a map that maps all parameters (flags and arguments) to their name.
// I'm not sure if this is worth caching, but if we hook this up like something like
// a lambda that might become more relevant.
func (p *ParameterLayer) GetParameters() map[string]*ParameterDefinition {
	ret := map[string]*ParameterDefinition{}
	for _, f := range p.Flags {
		ret[f.Name] = f
	}
	return ret
}

func (p *ParameterLayer) InitializeStructFromDefaults(s interface{}) error {
	parameters := p.GetParameters()
	err := InitializeStructFromParameterDefinitions(s, parameters)
	return err
}

func (p *ParameterLayer) AddFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error {
	parameters, err := CloneParameterDefinitionsWithDefaultsStruct(p.Flags, defaults)
	if err != nil {
		return err
	}

	// NOTE(manuel, 2023-02-21) Do we need to allow flags that are not "persistent"?
	err = AddFlagsToCobraCommand(cmd.PersistentFlags(), parameters)
	if err != nil {
		return err
	}

	AddFlagGroupToCobraCommand(cmd, p.Slug, p.Name, parameters)

	return nil
}

func (p *ParameterLayer) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	return GatherFlagsFromCobraCommand(cmd, p.Flags, false)
}
