package cmds

import (
	"fmt"
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
	Name      string       `yaml:"name"`
	Short     string       `yaml:"short"`
	Long      string       `yaml:"long,omitempty"`
	Flags     []*Parameter `yaml:"flags,omitempty"`
	Arguments []*Parameter `yaml:"arguments,omitempty"`

	Parents []string `yaml:",omitempty"`
	// Source indicates where the command was loaded from, to make debugging easier.
	Source string `yaml:",omitempty"`
}

type Command interface {
	// NOTE(2023-02-07, manuel) This is not actually used either by sqleton or pinocchio
	//
	// The reason for this is that they implement CobraCommand, which calls
	// RunFromCobra(cmd), and thus there is no need to actually implement Run() itself.
	// All they use is the Description() call, so there might be a reason to split the
	// interface into DescribedCommand and RunnableCommand, or so.
	// I don't really feel fluent with golang interface architecturing yet.

	Run(parameters map[string]interface{}, gp *GlazeProcessor) error
	Description() *CommandDescription
}

// YAMLCommandLoader is an interface that allows an application using the glazed
// library to loader commands from YAML files.
type YAMLCommandLoader interface {
	LoadCommandFromYAML(s io.Reader) ([]Command, error)
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
	LoadCommandsFromFS(f fs.FS, dir string) ([]Command, []*CommandAlias, error)
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

type YAMLFSCommandLoader struct {
	loader     YAMLCommandLoader
	sourceName string
	cmdRoot    string
}

func NewYAMLFSCommandLoader(
	loader YAMLCommandLoader,
	sourceName string,
	cmdRoot string,
) *YAMLFSCommandLoader {
	return &YAMLFSCommandLoader{
		loader:     loader,
		sourceName: sourceName,
		cmdRoot:    cmdRoot,
	}
}

func (l *YAMLFSCommandLoader) LoadCommandsFromFS(f fs.FS, dir string) ([]Command, []*CommandAlias, error) {
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
			subCommands, subAliases, err := l.LoadCommandsFromFS(f, fileName)
			if err != nil {
				return nil, nil, err
			}
			commands = append(commands, subCommands...)
			aliases = append(aliases, subAliases...)
		} else {
			// NOTE(2023-02-07, manuel) This might benefit from being made more generic than just loading from YAML
			//
			// One problem with the "commands from YAML" pattern being defined in glazed
			// is that is actually not great for a more complex application like pinocchio which
			// would benefit from loading applications from entire directories.
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
					commands, err := l.loader.LoadCommandFromYAML(file)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not load command from file %s", fileName)
					}
					if len(commands) != 1 {
						return nil, errors.New("Expected exactly one command")
					}
					command := commands[0]

					command.Description().Parents = getParentsFromDir(dir, l.cmdRoot)
					command.Description().Source = l.sourceName + ":" + fileName

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

							alias.Parents = getParentsFromDir(dir, l.cmdRoot)

							return alias, err
						}()

						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "Could not load command or alias from file %s: %s\n", fileName, err)
							continue
						} else {
							aliases = append(aliases, alias)
						}
					}
				} else {
					commands = append(commands, command)
				}
			}
		}
	}

	return commands, aliases, nil
}

// getParentsFromDir is a helper function to simply return a list of parent verbs
// for applications loaded from declarative yaml files.
// The directory structure mirrors the verb structure in cobra.
func getParentsFromDir(dir string, cmdRoot string) []string {
	// make sure both dir and cmdRoot have a trailing slash
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
