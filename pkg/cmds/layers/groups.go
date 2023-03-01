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

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string
}

// ParsedParameterLayer is the result of "parsing" input data using a ParameterLayer
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query parameters.
type ParsedParameterLayer struct {
	Layer      ParameterLayer
	Parameters map[string]interface{}
}

// Clone returns a copy of the parsedParameterLayer with a fresh Parameters map.
// However, neither the Layer nor the Parameters are deep copied.
func (ppl *ParsedParameterLayer) Clone() *ParsedParameterLayer {
	ret := &ParsedParameterLayer{
		Layer:      ppl.Layer,
		Parameters: make(map[string]interface{}),
	}
	for k, v := range ppl.Parameters {
		ret.Parameters[k] = v
	}
	return ret
}

type ParameterLayerParserFunc func() (*ParsedParameterLayer, error)

type ParameterLayerParser interface {
	RegisterParameterLayer(ParameterLayer) (ParameterLayerParserFunc, error)
}

// ParameterLayerImpl is a straight forward simple implementation of ParameterLayer
// that can easily be reused in more complex implementations.
type ParameterLayerImpl struct {
	Name        string                            `yaml:"name"`
	Slug        string                            `yaml:"slug"`
	Description string                            `yaml:"description"`
	Prefix      string                            `yaml:"prefix"`
	Flags       []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
}

func (p *ParameterLayerImpl) GetName() string {
	return p.Name
}

func (p *ParameterLayerImpl) GetSlug() string {
	return p.Slug
}

func (p *ParameterLayerImpl) GetDescription() string {
	return p.Description
}

func (p *ParameterLayerImpl) GetPrefix() string {
	return p.Prefix
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

type ParameterLayerOptions func(*ParameterLayerImpl) error

func NewParameterLayer(slug string, name string, options ...ParameterLayerOptions) (*ParameterLayerImpl, error) {
	ret := &ParameterLayerImpl{
		Slug: slug,
		Name: name,
	}

	for _, o := range options {
		err := o(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func WithPrefix(prefix string) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		p.Prefix = prefix
		return nil
	}
}

func WithDescription(description string) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		p.Description = description
		return nil
	}
}

func WithDefaults(s interface{}) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		// if s is a map[string]interface{} then we can just use that

		if m, ok := s.(map[string]interface{}); ok {
			return p.InitializeParameterDefaultsFromParameters(m)
		} else {
			return p.InitializeParameterDefaultsFromStruct(s)
		}
	}
}

// TODO(manuel, 2023-02-27) Might be worth making a struct defaults middleware

func WithFlags(flags ...*parameters.ParameterDefinition) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		p.Flags = flags
		return nil
	}
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

func NewParameterLayerFromYAML(s []byte, options ...ParameterLayerOptions) (*ParameterLayerImpl, error) {
	ret := &ParameterLayerImpl{}
	err := ret.LoadFromYAML(s)
	if err != nil {
		return nil, err
	}

	for _, o := range options {
		err = o(ret)
		if err != nil {
			return nil, err
		}
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

// InitializeParameterDefaultsFromStruct initializes the `ParameterDefinition` of the layer,
// which are often defined at compile time and loaded from a YAML file, with fresh
// ones from the struct.
// This is in some ways the opposite of `InitializeStructFromParameterDefaults`.
// The struct fields of `defaults` with a struct tag of `glazed.parameter` are used
// to initialize the `ParameterDefinition` with a matching name. If no matching
// `ParameterDefinition` is found, an error is returned.
func (p *ParameterLayerImpl) InitializeParameterDefaultsFromStruct(defaults interface{}) error {
	// check if defaults is a nil pointer
	if defaults == nil {
		return nil
	}
	ps := p.GetParameterDefinitions()
	err := parameters.InitializeParameterDefinitionsFromStruct(ps, defaults)
	return err
}

func (p *ParameterLayerImpl) InitializeParameterDefaultsFromParameters(
	ps map[string]interface{},
) error {
	pds := p.GetParameterDefinitions()
	err := parameters.InitializeParameterDefaultsFromParameters(pds, ps)
	return err

}

func (p *ParameterLayerImpl) InitializeStructFromParameterDefaults(s interface{}) error {
	if s == nil {
		return nil
	}
	ps := p.GetParameterDefinitions()
	err := parameters.InitializeStructFromParameterDefinitions(s, ps)
	return err
}

func (p *ParameterLayerImpl) AddFlagsToCobraCommand(cmd *cobra.Command) error {
	// NOTE(manuel, 2023-02-21) Do we need to allow flags that are not "persistent"?
	err := parameters.AddFlagsToCobraCommand(cmd.Flags(), p.Flags, p.Prefix)
	if err != nil {
		return err
	}

	AddFlagGroupToCobraCommand(cmd, p.Slug, p.Name, p.Flags)

	return nil
}

func (p *ParameterLayerImpl) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	return parameters.GatherFlagsFromCobraCommand(cmd, p.Flags, false, p.Prefix)
}
