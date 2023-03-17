package cmds

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CommandDescription contains the necessary information for registering
// a command with cobra. Because a command gets registered in a verb tree,
// a full list of Parents all the way to the root needs to be provided.
type CommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	Layers    []layers.ParameterLayer           `yaml:"layers,omitempty"`

	Parents []string `yaml:",omitempty"`
	// Source indicates where the command was loaded from, to make debugging easier.
	Source string `yaml:",omitempty"`
}

// Steal the builder API from https://github.com/bbkane/warg

type CommandDescriptionOption func(*CommandDescription)

func NewCommandDescription(name string, options ...CommandDescriptionOption) *CommandDescription {
	ret := &CommandDescription{
		Name: name,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (c *CommandDescription) Clone(cloneFlagsAndArguments bool, options ...CommandDescriptionOption) *CommandDescription {
	// clone flags
	flags := make([]*parameters.ParameterDefinition, len(c.Flags))
	for i, f := range c.Flags {
		if !cloneFlagsAndArguments {
			flags[i] = f
		} else {
			flags[i] = f.Copy()
		}
	}

	// clone arguments
	arguments := make([]*parameters.ParameterDefinition, len(c.Arguments))
	for i, a := range c.Arguments {
		if !cloneFlagsAndArguments {
			arguments[i] = a
		} else {
			arguments[i] = a.Copy()
		}
	}

	// clone layers
	layers_ := make([]layers.ParameterLayer, len(c.Layers))
	copy(layers_, c.Layers)

	// copy parents
	parents := make([]string, len(c.Parents))
	copy(parents, c.Parents)

	ret := &CommandDescription{
		Name:      c.Name,
		Short:     c.Short,
		Long:      c.Long,
		Flags:     flags,
		Arguments: arguments,
		Layers:    layers_,
		Parents:   parents,
		Source:    c.Source,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func WithName(s string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Name = s
	}
}

func WithShort(s string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Short = s
	}
}

func WithLong(s string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Long = s
	}
}

func WithFlags(f ...*parameters.ParameterDefinition) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Flags = append(c.Flags, f...)
	}
}

func WithArguments(a ...*parameters.ParameterDefinition) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Arguments = append(c.Arguments, a...)
	}
}

func WithLayers(l ...layers.ParameterLayer) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layers = append(c.Layers, l...)
	}
}

func WithReplaceLayers(layers_ ...layers.ParameterLayer) CommandDescriptionOption {
	return func(c *CommandDescription) {
	outerLoop:
		for _, l := range layers_ {
			for i, ll := range c.Layers {
				if ll.GetSlug() == l.GetSlug() {
					c.Layers[i] = l
					continue outerLoop
				}
			}
		}
	}
}

func WithReplaceFlags(flags ...*parameters.ParameterDefinition) CommandDescriptionOption {
	return func(c *CommandDescription) {
	outerLoop:
		for _, f := range flags {
			for i, ff := range c.Flags {
				if ff.Name == f.Name {
					c.Flags[i] = f
					continue outerLoop
				}
			}
		}
	}
}

func WithReplaceArguments(args ...*parameters.ParameterDefinition) CommandDescriptionOption {
	return func(c *CommandDescription) {
	outerLoop:
		for _, a := range args {
			for i, aa := range c.Arguments {
				if aa.Name == a.Name {
					c.Arguments[i] = a
					continue outerLoop
				}
			}
		}
	}
}

func WithParents(p ...string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Parents = p
	}
}

func WithPrependParents(p ...string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Parents = append(p, c.Parents...)
	}
}

func WithStripParentsPrefix(prefixes []string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		toRemove := 0
		for i, p := range c.Parents {
			if i < len(prefixes) && p == prefixes[i] {
				toRemove = i + 1
			}
		}
		c.Parents = c.Parents[toRemove:]
	}
}

func WithSource(s string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Source = s
	}
}

func WithPrependSource(s string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Source = s + c.Source
	}
}

type Command interface {
	Description() *CommandDescription
}

type WriterCommand interface {
	Command
	RunIntoWriter(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
		w io.Writer,
	) error
}

type GlazeCommand interface {
	Command
	// Run is called to actually execute the command.
	//
	// NOTE(manuel, 2023-02-27) We can probably simplify this to only take parsed layers
	//
	// The ps and GlazeProcessor calls could be replaced by a GlazeCommand specific layer,
	// which would allow the GlazeCommand to parse into a specific struct. The GlazeProcessor
	// is just something created by the passed in GlazeLayer anyway.
	//
	// When we are just left with building a convenience wrapper for Glaze based commands,
	// instead of forcing it into the upstream interface.
	//
	// https://github.com/go-go-golems/glazed/issues/217
	// https://github.com/go-go-golems/glazed/issues/216
	// See https://github.com/go-go-golems/glazed/issues/173
	Run(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
		ps map[string]interface{},
		gp Processor,
	) error
}

type ExitWithoutGlazeError struct{}

func (e *ExitWithoutGlazeError) Error() string {
	return "Exit without glaze"
}

// YAMLCommandLoader is an interface that allows an application using the glazed
// library to loader commands from YAML files.
type YAMLCommandLoader interface {
	LoadCommandFromYAML(s io.Reader, options ...CommandDescriptionOption) ([]Command, error)
	LoadCommandAliasFromYAML(s io.Reader) ([]*CommandAlias, error)
}

// TODO(2023-02-09, manuel) We can probably implement the directory walking part in a couple of lines
//
// Currently, we walk the directory in both the yaml loader below, and in the elastic search directory
// command loader in escuse-me.

// FSCommandLoader is an interface that describes the most generic loader type,
// which is then used to load commands and command aliases from embedded queries
// and from "repository" directories used by glazed.
//
// Examples of this pattern are used in sqleton, escuse-me and pinocchio.
type FSCommandLoader interface {
	LoadCommandsFromFS(f fs.FS, dir string, options ...CommandDescriptionOption) ([]Command, []*CommandAlias, error)
}

func LoadCommandAliasFromYAML(s io.Reader) ([]*CommandAlias, error) {
	var alias CommandAlias
	err := yaml.NewDecoder(s).Decode(&alias)
	if err != nil {
		return nil, err
	}

	if !alias.IsValid() {
		return nil, errors.New("Invalid command alias")
	}

	return []*CommandAlias{&alias}, nil
}

// TODO(2022-12-21, manuel): Add list of choices as a type
// what about list of dates? list of bools?
// should list just be a flag?
//
// See https://github.com/go-go-golems/glazed/issues/117

// YAMLFSCommandLoader walks a FS and finds all yaml files, loading them using the passed
// YAMLCommandLoader.
//
// It handles the following generic functionality:
// - recursive FS walking
// - setting SourceName for each command
// - setting Parents for each command
type YAMLFSCommandLoader struct {
	loader YAMLCommandLoader
	// sourceName is a prefix prepended to give information about where each command comes from
	sourceName string
	// rootDirectory is the root directory these commands will be loaded from
	rootDirectory string
}

func NewYAMLFSCommandLoader(
	loader YAMLCommandLoader,
	sourceName string,
	cmdRoot string,
) *YAMLFSCommandLoader {
	return &YAMLFSCommandLoader{
		loader:        loader,
		sourceName:    sourceName,
		rootDirectory: cmdRoot,
	}
}

// LoadCommandsFromFS walks the FS and loads all commands and command aliases found.
//
// TODO(manuel, 2023-03-16) Add loading of helpsystem files
// See https://github.com/go-go-golems/glazed/issues/55
// See https://github.com/go-go-golems/glazed/issues/218
func (l *YAMLFSCommandLoader) LoadCommandsFromFS(
	f fs.FS,
	dir string,
	options ...CommandDescriptionOption,
) ([]Command, []*CommandAlias, error) {
	var commands []Command
	var aliases []*CommandAlias

	entries, err := fs.ReadDir(f, dir)
	if err != nil {
		return nil, nil, err
	}
	for _, entry := range entries {
		// skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fileName := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			subCommands, subAliases, err := l.LoadCommandsFromFS(f, fileName, options...)
			if err != nil {
				return nil, nil, err
			}
			commands = append(commands, subCommands...)
			aliases = append(aliases, subAliases...)
			continue
		}
		// NOTE(2023-02-07, manuel) This might benefit from being made more generic than just loading from YAML
		//
		// One problem with the "commands from YAML" pattern being defined in glazed
		// is that is actually not great for a more complex application like pinocchio which
		// would benefit from loading applications from entire directories.
		//
		// This can of course be solved by providing a FSCommandLoader for directories.
		//
		// Similarly, we might want to store applications in a database, or generate them on the
		// fly using some resources on the disk.
		//
		// See https://github.com/go-go-golems/glazed/issues/116
		if strings.HasSuffix(entry.Name(), ".yml") ||
			strings.HasSuffix(entry.Name(), ".yaml") {
			command, err := func() (Command, error) {
				file, err := f.Open(fileName)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not open file %s", fileName)
				}
				defer func() {
					_ = file.Close()
				}()

				log.Debug().Str("file", fileName).Msg("Loading command from file")
				options_ := append([]CommandDescriptionOption{
					WithSource(l.sourceName + ":" + fileName),
					WithParents(GetParentsFromDir(dir, l.rootDirectory)...),
				}, options...)
				commands, err := l.loader.LoadCommandFromYAML(file, options_...)
				if err != nil {
					return nil, err
				}
				if len(commands) != 1 {
					return nil, errors.New("Expected exactly one command")
				}
				command := commands[0]

				return command, err
			}()

			if err != nil {
				// If the error was a yaml parsing error, then we try to load the YAML file
				// again, but as an alias this time around. YAML / JSON parsing in golang
				// definitely is a bit of an adventure.
				if _, ok := err.(*yaml.TypeError); ok {
					alias, err := func() (*CommandAlias, error) {
						file, err := f.Open(fileName)
						if err != nil {
							return nil, errors.Wrapf(err, "Could not open file %s", fileName)
						}
						defer func() {
							_ = file.Close()
						}()

						log.Debug().Str("file", fileName).Msg("Loading alias from file")
						aliases, err := l.loader.LoadCommandAliasFromYAML(file)
						if err != nil {
							return nil, err
						}
						if len(aliases) != 1 {
							return nil, errors.New("Expected exactly one alias")
						}
						alias := aliases[0]
						alias.Source = l.sourceName + ":" + fileName

						alias.Parents = GetParentsFromDir(dir, l.rootDirectory)

						return alias, err
					}()
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Could not load command or alias from file %s: %s\n", fileName, err)
						continue
					} else {
						aliases = append(aliases, alias)
					}
				}
				continue
			}

			commands = append(commands, command)
		}
	}

	return commands, aliases, nil
}

// GetParentsFromDir is a helper function to simply return a list of parent verbs
// for applications loaded from declarative yaml files.
// The directory structure mirrors the verb structure in cobra.
func GetParentsFromDir(dir string, cmdRoot string) []string {
	// make sure both dir and rootDirectory have a trailing slash
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	if !strings.HasSuffix(cmdRoot, "/") {
		cmdRoot += "/"
	}
	pathToFile := strings.TrimPrefix(dir, cmdRoot)
	parents := strings.Split(pathToFile, "/")
	if len(parents) > 0 && parents[len(parents)-1] == "" {
		parents = parents[:len(parents)-1]
	}
	return parents
}
