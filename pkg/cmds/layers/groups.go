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

type ParameterLayer interface {
	AddFlag(flag *parameters.ParameterDefinition)
	GetParameterDefinitions() map[string]*parameters.ParameterDefinition
	InitializeStructFromDefaults(s interface{}) error
	AddFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error
	ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error)
}

type ParameterLayerImpl struct {
	Name        string                            `yaml:"name"`
	Slug        string                            `yaml:"slug"`
	Description string                            `yaml:"description"`
	Flags       []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
}

func (p *ParameterLayerImpl) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		Name        string                            `yaml:"name"`
		Slug        string                            `yaml:"slug"`
		Description string                            `yaml:"description"`
		Flags       []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	}
	if err := unmarshal(&raw); err != nil {
		return err
	}
	p.Name = raw.Name
	p.Slug = raw.Slug
	p.Description = raw.Description
	p.Flags = raw.Flags
	return nil
}

type ParameterLayerOptions func(*ParameterLayerImpl)

func NewParameterLayer(slug string, name string, options ...ParameterLayerOptions) *ParameterLayerImpl {
	ret := &ParameterLayerImpl{
		Slug: slug,
		Name: name,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (p *ParameterLayerImpl) LoadFromYAML(s []byte) error {
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

func NewParameterLayerFromYAML(s []byte) (*ParameterLayerImpl, error) {
	ret := &ParameterLayerImpl{}
	err := ret.LoadFromYAML(s)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (p *ParameterLayerImpl) AddFlag(flag *parameters.ParameterDefinition) {
	p.Flags = append(p.Flags, flag)
}

// GetParameterDefinitions returns a map that maps all parameters (flags and arguments) to their name.
// I'm not sure if this is worth caching, but if we hook this up like something like
// a lambda that might become more relevant.
func (p *ParameterLayerImpl) GetParameterDefinitions() map[string]*parameters.ParameterDefinition {
	ret := map[string]*parameters.ParameterDefinition{}
	for _, f := range p.Flags {
		ret[f.Name] = f
	}
	return ret
}

func (p *ParameterLayerImpl) InitializeStructFromDefaults(s interface{}) error {
	ps := p.GetParameterDefinitions()
	err := parameters.InitializeStructFromParameterDefinitions(s, ps)
	return err
}

func (p *ParameterLayerImpl) AddFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error {
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

func (p *ParameterLayerImpl) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	return parameters.GatherFlagsFromCobraCommand(cmd, p.Flags, false)
}
