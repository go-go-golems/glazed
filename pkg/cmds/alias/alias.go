package alias

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
	Flags     map[string]string `yaml:"flags,omitempty"`
	Arguments []string          `yaml:"arguments,omitempty"`
	Layout    []*layout.Section `yaml:"layout,omitempty"`

	AliasedCommand cmds.Command `yaml:",omitempty"`
	Parents        []string     `yaml:",omitempty"`
	Source         string       `yaml:",omitempty"`
}

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

func (a *CommandAlias) String() string {
	return fmt.Sprintf("CommandAlias{Name: %s, AliasFor: %s, Parents: %v, Source: %s}",
		a.Name, a.AliasFor, a.Parents, a.Source)
}

func (a *CommandAlias) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp cmds.Processor,
) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	glazeCommand, ok := a.AliasedCommand.(cmds.GlazeCommand)
	if !ok {
		return errors.New("aliased command is not a GlazeCommand")
	}
	return glazeCommand.Run(ctx, parsedLayers, ps, gp)
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
	s := a.AliasedCommand.Description()
	layout_ := a.Layout
	if layout_ == nil {
		layout_ = s.Layout.Sections
	}
	ret := &cmds.CommandDescription{
		Name:      a.Name,
		Short:     s.Short,
		Long:      s.Long,
		Flags:     []*parameters.ParameterDefinition{},
		Arguments: []*parameters.ParameterDefinition{},
		Layout: &layout.Layout{
			Sections: layout_,
		},
		Layers:  s.Layers,
		Parents: a.Parents,
		Source:  a.Source,
	}

	for _, flag := range s.Flags {
		newFlag := flag.Copy()
		// newFlag.Required = false
		ret.Flags = append(ret.Flags, newFlag)
	}

	for _, argument := range s.Arguments {
		newArgument := argument.Copy()

		// ## Parsing the overloaded strings to actual types to store as flag defaults
		//
		// NOTE(2023-04-20) We can't easily return overloaded flags and arguments as defaults in the CommandDescription
		//
		// This was created before layers being a thing, so that the overloads are not really type specific.
		// This is a problem already when capturing the aliases, but it should be much easier now.
		//
		// For now, we still use strings, and as such need the overloading of an alias to be caught at the primitive
		// parsing step (cobra for CLI, HTTP parsers for parka).
		//
		// See https://github.com/go-go-golems/glazed/issues/287
		//
		// For now, parka handling takes an explicit list of defaults in its parser functions,
		// which might not be the worst idea for overloading things at registration time either.

		// ## Handling argument count
		//
		// See also the note in glazed_layer.go about checking the argument count. This might all
		// refer to overloading arguments, and not just flags. This seems to make sense given the
		// talk about argument counts.
		//
		// ---
		//
		// TODO(2022-12-22, manuel) this needs to be handled, overriding arguments and figuring out which order
		// is a bitch
		//
		// so iN command.go in cobra, prerun is run before the arg validation is done
		// so that we could potentially override the args here
		//
		// the args are gotten from c.Flags().Args()
		//
		// it looks like in prerun, we could check if args is empty,
		// and if so, pass in our arguments  by calling Parse() a second time,
		// and then going over the newly set arg?
		//
		// It's of course going to be relying on cobra internals a bit,
		// by assuming that calling parse a second time is not going to interfere with already set flags
		// so maybe the best solution is really just to interleave the flags at the outset
		// by doing our own little scanning, which is probably useful anyway if done in glazed
		// so that we can handle different types of arg parsing.
		//
		// if defaultValue, ok := a.ArgumentDefaults[argument.Name]; ok {
		//	newArgument.Default = defaultValue
		// }
		// newArgument.Required = false
		ret.Arguments = append(ret.Arguments, newArgument)
	}

	return ret
}
