package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ParameterLayerImpl is a straight forward simple implementation of ParameterLayer
// that can easily be reused in more complex implementations.
type ParameterLayerImpl struct {
	Name        string                            `yaml:"name"`
	Slug        string                            `yaml:"slug"`
	Description string                            `yaml:"description"`
	Prefix      string                            `yaml:"prefix"`
	Flags       []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	ChildLayers []ParameterLayer                  `yaml:"childLayers,omitempty"`
}

var _ ParameterLayer = &ParameterLayerImpl{}
var _ CobraParameterLayer = &ParameterLayerImpl{}

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

func WithName(name string) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		p.Name = name
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

func WithParameters(flags ...*parameters.ParameterDefinition) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		p.Flags = append(p.Flags, flags...)
		return nil
	}
}

func WithArguments(arguments ...*parameters.ParameterDefinition) ParameterLayerOptions {
	return func(p *ParameterLayerImpl) error {
		for _, a := range arguments {
			a.IsArgument = true
		}
		p.Flags = append(p.Flags, arguments...)
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

func (p *ParameterLayerImpl) AddFlags(flag ...*parameters.ParameterDefinition) {
	p.Flags = append(p.Flags, flag...)
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

func (p *ParameterLayerImpl) GetParameterValuesFromMap(m map[string]interface{}, onlyProvided bool) (map[string]interface{}, error) {
	ps := p.GetParameterDefinitions()
	return parameters.GatherParametersFromMap(m, ps, onlyProvided)
}

// AddLayerToCobraCommand adds all flags of the layer to the given Cobra command.
// It also creates a flag group representing the layer and adds it to the command.
// If the layer has a prefix, the flags are added with that prefix.
func (p *ParameterLayerImpl) AddLayerToCobraCommand(cmd *cobra.Command) error {
	// NOTE(manuel, 2023-02-21) Do we need to allow flags that are not "persistent"?
	err := parameters.AddParametersToCobraCommand(cmd, p.Flags, p.Prefix)
	if err != nil {
		return err
	}

	AddFlagGroupToCobraCommand(cmd, p.Slug, p.Name, p.Flags, p.Prefix)

	return nil
}

// ParseLayerFromCobraCommand parses the flags of the layer from the given Cobra command.
// If the layer has a prefix, the flags are parsed with that prefix (meaning, the prefix
// is stripped from the flag names before they are added to the returned map).
//
// This will return a map containing the value (or default value) of each flag
// of the layer.
func (p *ParameterLayerImpl) ParseLayerFromCobraCommand(cmd *cobra.Command) (*ParsedParameterLayer, error) {
	var flags = []*parameters.ParameterDefinition{}
	for _, pd := range p.Flags {
		if !pd.IsArgument {
			flags = append(flags, pd)
		}
	}
	ps, err := parameters.GatherFlagsFromCobraCommand(cmd, flags, false, false, p.Prefix)
	if err != nil {
		return nil, err
	}

	return &ParsedParameterLayer{
		Layer:      p,
		Parameters: ps,
	}, nil
}

func (p *ParameterLayerImpl) ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (map[string]interface{}, error) {
	return parameters.GatherParametersFromMap(m, p.GetParameterDefinitions(), onlyProvided)
}

func (p *ParameterLayerImpl) Clone() ParameterLayer {
	ret := &ParameterLayerImpl{
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Prefix:      p.Prefix,
		Flags:       []*parameters.ParameterDefinition{},
	}
	for _, f := range p.Flags {
		ret.Flags = append(ret.Flags, f.Clone())
	}
	return ret

}
