package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.

type ParameterLayer struct {
	Name        string                            `yaml:"name"`
	Slug        string                            `yaml:"slug"`
	Description string                            `yaml:"description"`
	Flags       []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
}

type ParameterLayerOptions func(*ParameterLayer)

func NewParameterLayer(slug string, name string, options ...ParameterLayerOptions) *ParameterLayer {
	ret := &ParameterLayer{
		Slug: slug,
		Name: name,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (p *ParameterLayer) LoadFromYAML(s []byte) error {
	err := yaml.Unmarshal(s, p)
	if err != nil {
		return err
	}

	for _, p := range p.Flags {
		err := p.CheckParameterDefaultValueValidity()
		if err != nil {
			panic(errors.Wrap(err, "Failed to check parameter default value validity"))
		}
	}

	return nil
}

func NewParameterLayerFromYAML(s []byte) (*ParameterLayer, error) {
	ret := &ParameterLayer{}
	err := ret.LoadFromYAML(s)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *ParameterLayer) AddFlag(flag *parameters.ParameterDefinition) {
	p.Flags = append(p.Flags, flag)
}

// GetParameters returns a map that maps all parameters (flags and arguments) to their name.
// I'm not sure if this is worth caching, but if we hook this up like something like
// a lambda that might become more relevant.
func (p *ParameterLayer) GetParameters() map[string]*parameters.ParameterDefinition {
	ret := map[string]*parameters.ParameterDefinition{}
	for _, f := range p.Flags {
		ret[f.Name] = f
	}
	return ret
}

func (p *ParameterLayer) InitializeStructFromDefaults(s interface{}) error {
	ps := p.GetParameters()
	err := parameters.InitializeStructFromParameterDefinitions(s, ps)
	return err
}

func (p *ParameterLayer) AddFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error {
	ps, err := parameters.CloneParameterDefinitionsWithDefaultsStruct(p.Flags, defaults)
	if err != nil {
		return err
	}

	// NOTE(manuel, 2023-02-21) Do we need to allow flags that are not "persistent"?
	err = parameters.AddFlagsToCobraCommand(cmd.PersistentFlags(), ps)
	if err != nil {
		return err
	}

	AddFlagGroupToCobraCommand(cmd, p.Slug, p.Name, ps)

	return nil
}

func (p *ParameterLayer) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	return parameters.GatherFlagsFromCobraCommand(cmd, p.Flags, false)
}
