package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SectionImpl is a straight forward simple implementation of Section
// that can easily be reused in more complex implementations.
type SectionImpl struct {
	Name        string              `yaml:"name"`
	Slug        string              `yaml:"slug"`
	Description string              `yaml:"description"`
	Prefix      string              `yaml:"prefix"`
	Definitions *fields.Definitions `yaml:"flags,omitempty"`
	ChildLayers []Section           `yaml:"childLayers,omitempty"`
}

var _ Section = &SectionImpl{}
var _ CobraSection = &SectionImpl{}

func (p *SectionImpl) GetName() string {
	return p.Name
}

func (p *SectionImpl) GetSlug() string {
	return p.Slug
}

func (p *SectionImpl) GetDescription() string {
	return p.Description
}

func (p *SectionImpl) GetPrefix() string {
	return p.Prefix
}

func (p *SectionImpl) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var raw struct {
		Name        string              `yaml:"name"`
		Slug        string              `yaml:"slug"`
		Description string              `yaml:"description"`
		Flags       *fields.Definitions `yaml:"flags,omitempty"`
	}
	raw.Flags = fields.NewDefinitions()
	if err := unmarshal(&raw); err != nil {
		return err
	}
	p.Name = raw.Name
	p.Slug = raw.Slug
	p.Description = raw.Description
	p.Definitions = raw.Flags
	return nil
}

type SectionOption func(*SectionImpl) error

func NewSection(slug string, name string, options ...SectionOption) (*SectionImpl, error) {
	ret := &SectionImpl{
		Slug:        slug,
		Name:        name,
		Definitions: fields.NewDefinitions(),
	}

	for _, o := range options {
		err := o(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func WithPrefix(prefix string) SectionOption {
	return func(p *SectionImpl) error {
		p.Prefix = prefix
		return nil
	}
}

func WithName(name string) SectionOption {
	return func(p *SectionImpl) error {
		p.Name = name
		return nil
	}
}

func WithDescription(description string) SectionOption {
	return func(p *SectionImpl) error {
		p.Description = description
		return nil
	}
}

func WithDefaults(s interface{}) SectionOption {
	return func(p *SectionImpl) error {
		// if s is a map[string]interface{} then we can just use that

		if m, ok := s.(map[string]interface{}); ok {
			return p.InitializeDefaultsFromParameters(m)
		} else {
			return p.InitializeDefaultsFromStruct(s)
		}
	}
}

func WithFields(parameterDefinitions ...*fields.Definition) SectionOption {
	return func(p *SectionImpl) error {
		for _, f := range parameterDefinitions {
			p.Definitions.Set(f.Name, f)
		}
		return nil
	}
}

func WithArguments(arguments ...*fields.Definition) SectionOption {
	return func(p *SectionImpl) error {
		for _, a := range arguments {
			a.IsArgument = true
			p.Definitions.Set(a.Name, a)
		}
		return nil
	}
}

func (p *SectionImpl) LoadFromYAML(s []byte) error {
	err := yaml.Unmarshal(s, p)
	if err != nil {
		return err
	}

	for f_ := p.Definitions.Oldest(); f_ != nil; f_ = f_.Next() {
		_, err := f_.Value.CheckDefaultValueValidity()
		if err != nil {
			return err
		}
	}

	return nil
}

func NewSectionFromYAML(s []byte, options ...SectionOption) (*SectionImpl, error) {
	ret := &SectionImpl{}
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

func (p *SectionImpl) AddFields(flag ...*fields.Definition) {
	for _, f := range flag {
		p.Definitions.Set(f.Name, f)
	}
}

// GetDefinitions returns a map that maps all parameters (flags and arguments) to their name.
// I'm not sure if this is worth caching, but if we hook this up like something like
// a lambda that might become more relevant.
func (p *SectionImpl) GetDefinitions() *fields.Definitions {
	ret := fields.NewDefinitions()
	for f := p.Definitions.Oldest(); f != nil; f = f.Next() {
		ret.Set(f.Key, f.Value)
	}
	return ret
}

// InitializeDefaultsFromStruct initializes the `ParameterDefinition` of the layer,
// which are often defined at compile time and loaded from a YAML file, with fresh
// ones from the struct.
// This is in some ways the opposite of `InitializeStructFromParameterDefaults`.
// The struct fields of `defaults` with a struct tag of `glazed` are used
// to initialize the `ParameterDefinition` with a matching name. If no matching
// `ParameterDefinition` is found, an error is returned.
func (p *SectionImpl) InitializeDefaultsFromStruct(defaults interface{}) error {
	// check if defaults is a nil pointer
	if defaults == nil {
		return nil
	}
	ps := p.GetDefinitions()
	err := ps.InitializeDefaultsFromStruct(defaults)
	return err
}

// InitializeDefaultsFromParameters initializes the parameter definitions
// of the layer from the given map of parameter values. The parameter definitions
// are updated in-place.
func (p *SectionImpl) InitializeDefaultsFromParameters(
	ps map[string]interface{},
) error {
	pds := p.GetDefinitions()
	err := pds.InitializeDefaultsFromMap(ps)
	return err
}

func (p *SectionImpl) InitializeStructFromParameterDefaults(s interface{}) error {
	if s == nil {
		return nil
	}
	pds := p.GetDefinitions()
	err := pds.InitializeStructFromDefaults(s)
	return err
}

// AddSectionToCobraCommand adds all flags of the section to the given Cobra command.
// It also creates a flag group representing the layer and adds it to the command.
// If the layer has a prefix, the flags are added with that prefix.
func (p *SectionImpl) AddSectionToCobraCommand(cmd *cobra.Command) error {
	err := p.Definitions.AddFieldsToCobraCommand(cmd, p.Prefix)
	if err != nil {
		return err
	}

	AddFlagGroupToCobraCommand(cmd, p.Slug, p.Name, p.Definitions, p.Prefix)

	return nil
}

// ParseLayerFromCobraCommand parses the flags of the layer from the given Cobra command.
// If the layer has a prefix, the flags are parsed with that prefix (meaning, the prefix
// is stripped from the flag names before they are added to the returned map).
//
// This will return a map containing the value (or default value) of each flag
// of the layer.
func (p *SectionImpl) ParseLayerFromCobraCommand(
	cmd *cobra.Command,
	options ...fields.ParseOption,
) (*values.SectionValues, error) {
	ps, err := p.Definitions.GatherFlagsFromCobraCommand(
		// TODO(manuel, 2024-01-05) We probably need to move the required check to a higher level middleware, because
		// we are not relying on cobra so much anymore since we introduced middlewares
		//
		// NOTE(manuel, 2024-01-17) I'm moving onlyProvided back to false because we need the default values when adding flags to individual manual commands.
		// See MD WHITE (2) p.26
		// NOTE(manuel, 2024-01-17) I'm setting it back to true, because we want each middleware (including ParseFromCobraCommand) to only override the defaults
		// because defaults are now set through a middleware as well.
		cmd, true, false, p.Prefix,
		options...,
	)
	if err != nil {
		return nil, err
	}

	return &values.SectionValues{
		Section: p,
		Fields:  ps,
	}, nil
}

func (p *SectionImpl) GatherFieldsFromMap(
	m map[string]interface{}, onlyProvided bool,
	options ...fields.ParseOption,
) (*fields.FieldValues, error) {
	return p.Definitions.GatherFieldsFromMap(m, onlyProvided, options...)
}

func (p *SectionImpl) Clone() Section {
	ret := &SectionImpl{
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Prefix:      p.Prefix,
		Definitions: fields.NewDefinitions(),
	}
	for v := p.Definitions.Oldest(); v != nil; v = v.Next() {
		ret.Definitions.Set(v.Key, v.Value.Clone())
	}
	return ret

}
