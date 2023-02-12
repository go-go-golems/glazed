package cmds

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CommandAlias defines a struct that should be able to define generic aliases
// for any kind of command line applications, by providing overrides for certain
// flags (prepopulating them with certain flags and arguments, basically)
type CommandAlias struct {
	Name      string            `yaml:"name"`
	AliasFor  string            `yaml:"aliasFor"`
	Flags     map[string]string `yaml:"flags,omitempty"`
	Arguments []string          `yaml:"arguments,omitempty"`

	AliasedCommand Command  `yaml:",omitempty"`
	Parents        []string `yaml:",omitempty"`
	Source         string   `yaml:",omitempty"`
}

func (a *CommandAlias) Run(parameters map[string]interface{}, gp *GlazeProcessor) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	return a.AliasedCommand.Run(parameters, gp)
}

func (a *CommandAlias) IsValid() bool {
	return a.Name != "" && a.AliasFor != ""
}

// Description returns the CommandDescription of an alias.
// It computes it at runtime by loading the aliased command's Description() and
// making copies of its flags and arguments.
// This is necessary because they get mutated at runtime with various defaults,
// depending on where they come from.
func (a *CommandAlias) Description() *CommandDescription {
	s := a.AliasedCommand.Description()
	ret := &CommandDescription{
		Name:      a.Name,
		Short:     s.Short,
		Long:      s.Long,
		Flags:     []*ParameterDefinition{},
		Arguments: []*ParameterDefinition{},
	}

	for _, flag := range s.Flags {
		newFlag := flag.Copy()
		// newFlag.Required = false
		ret.Flags = append(ret.Flags, newFlag)
	}

	for _, argument := range s.Arguments {
		newArgument := argument.Copy()

		// NOTE(2023-02-07, manuel) I don't fully understand what this is referring to anymore,
		// but I remember struggling with this in the context of setting and overriding default values.
		// Say, if an alias defines --fields id,name and then the user passes in --fields foo,bla
		// on top, I remember there being some kind of conflict.
		//
		// See also the note in cobra.go about checking the argument count. This might all
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

func (a *CommandAlias) RunFromCobra(cmd *cobra.Command, args []string) error {
	return a.AliasedCommand.(CobraCommand).RunFromCobra(cmd, args)
}

func (a *CommandAlias) BuildCobraCommand() (*cobra.Command, error) {
	return NewCobraCommand(a)
}
