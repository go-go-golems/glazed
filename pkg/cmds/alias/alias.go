package alias

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
)

type Option func(*CommandAlias)

// CommandAlias defines a struct that should be able to define generic aliases
// for any kind of command line applications, by providing overrides for certain
// flags (prepopulating them with certain flags and arguments, basically)
type CommandAlias struct {
	Name      string            `yaml:"name"`
	AliasFor  string            `yaml:"aliasFor"`
	Short     string            `yaml:"short,omitempty"`
	Long      string            `yaml:"long,omitempty"`
	Flags     map[string]string `yaml:"flags,omitempty"`
	Arguments []string          `yaml:"arguments,omitempty"`
	Layout    []*layout.Section `yaml:"layout,omitempty"`

	AliasedCommand cmds.Command `yaml:",omitempty"`
	Parents        []string     `yaml:",omitempty"`
	Source         string       `yaml:",omitempty"`
}

var _ cmds.Command = (*CommandAlias)(nil)

func WithName(name string) Option {
	return func(a *CommandAlias) {
		a.Name = name
	}
}

func WithAliasFor(aliasFor string) Option {
	return func(a *CommandAlias) {
		a.AliasFor = aliasFor
	}
}

func WithFlags(flags map[string]string) Option {
	return func(a *CommandAlias) {
		for k, v := range flags {
			a.Flags[k] = v
		}
	}
}

func WithArguments(arguments []string) Option {
	return func(a *CommandAlias) {
		a.Arguments = arguments
	}
}

func WithParents(parents ...string) Option {
	return func(a *CommandAlias) {
		a.Parents = parents
	}
}

func WithStripParentsPrefix(prefixes []string) Option {
	return func(a *CommandAlias) {
		toRemove := 0
		for i, p := range a.Parents {
			if i < len(prefixes) && p == prefixes[i] {
				toRemove = i + 1
			}
		}
		a.Parents = a.Parents[toRemove:]
	}
}

func WithSource(source string) Option {
	return func(a *CommandAlias) {
		a.Source = source
	}
}

func WithPrependSource(source string) Option {
	return func(a *CommandAlias) {
		a.Source = source + a.Source
	}
}

func NewCommandAlias(options ...Option) *CommandAlias {
	a := &CommandAlias{
		Flags: make(map[string]string),
	}
	for _, option := range options {
		option(a)
	}
	return a
}

func NewCommandAliasFromYAML(s io.Reader, options ...Option) (*CommandAlias, error) {
	a := NewCommandAlias()
	if err := yaml.NewDecoder(s).Decode(a); err != nil {
		return nil, err
	}
	if !a.IsValid() {
		return nil, errors.New("Invalid command alias")
	}

	for _, option := range options {
		option(a)
	}

	return a, nil
}

var _ cmds.GlazeCommand = (*CommandAlias)(nil)
var _ cmds.WriterCommand = (*CommandAlias)(nil)

func (a *CommandAlias) String() string {
	return fmt.Sprintf("CommandAlias{Name: %s, AliasFor: %s, Parents: %v, Source: %s}",
		a.Name, a.AliasFor, a.Parents, a.Source)
}

func (a *CommandAlias) ToYAML(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	defer func(enc *yaml.Encoder) {
		_ = enc.Close()
	}(enc)

	return enc.Encode(a)
}

func (a *CommandAlias) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	glazeCommand, ok := a.AliasedCommand.(cmds.GlazeCommand)
	if !ok {
		return errors.New("aliased command is not a GlazeCommand")
	}
	return glazeCommand.RunIntoGlazeProcessor(ctx, parsedLayers, gp)
}

func (a *CommandAlias) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	writerCommand, ok := a.AliasedCommand.(cmds.WriterCommand)
	if !ok {
		return errors.New("aliased command is not a GlazeCommand")
	}
	return writerCommand.RunIntoWriter(ctx, parsedLayers, w)
}

func (a *CommandAlias) IsValid() bool {
	return a.Name != "" && a.AliasFor != ""
}

// Description returns the CommandDescription of an alias.
// It computes it at runtime by loading the aliased command's Description() and
// making copies of its flags and arguments.
// This is necessary because they get mutated at runtime with various defaults,
// depending on where they come from.
func (a *CommandAlias) Description() *cmds.CommandDescription {
	if a.AliasedCommand == nil {
		return nil
	}
	s := a.AliasedCommand.Description()
	layout_ := a.Layout
	if layout_ == nil {
		layout_ = s.Layout
	}

	newLayers := s.Layers.Clone()

	ret := &cmds.CommandDescription{
		Name:    a.Name,
		Short:   s.Short,
		Long:    s.Long,
		Layout:  layout_,
		Layers:  newLayers,
		Parents: a.Parents,
		Source:  a.Source,
	}

	return ret
}
