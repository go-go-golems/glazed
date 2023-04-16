package loaders

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// YAMLCommandLoader is an interface that allows an application using the glazed
// library to loader commands from YAML files.
type YAMLCommandLoader interface {
	LoadCommandFromYAML(s io.Reader, options ...cmds.CommandDescriptionOption) ([]cmds.Command, error)
	LoadCommandAliasFromYAML(s io.Reader, options ...alias.Option) ([]*alias.CommandAlias, error)
}

type ReaderCommandLoader interface {
	LoadCommandsFromReader(r io.Reader, options []cmds.CommandDescriptionOption, aliasOptions []alias.Option) ([]cmds.Command, error)
}

type YAMLReaderCommandLoader struct {
	YAMLCommandLoader
}

func YAMLReaderCommandLoaderFromYAMLCommandLoader(loader YAMLCommandLoader) *YAMLReaderCommandLoader {
	return &YAMLReaderCommandLoader{
		YAMLCommandLoader: loader,
	}
}

func (l *YAMLReaderCommandLoader) LoadCommandsFromReader(r io.Reader, options []cmds.CommandDescriptionOption, aliasOptions []alias.Option) ([]cmds.Command, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	br := strings.NewReader(string(bytes))
	cmds_, err := l.LoadCommandFromYAML(br, options...)
	if err != nil {
		br = strings.NewReader(string(bytes))
		aliases, err := l.LoadCommandAliasFromYAML(br, aliasOptions...)
		if err != nil {
			return nil, err
		}
		aliases_, b := cast.CastList[cmds.Command](aliases)
		if !b {
			return nil, errors.New("could not cast aliases to commands")
		}
		return aliases_, nil
	}

	return cmds_, nil
}

// FSCommandLoader is an interface that describes the most generic loader type,
// which is then used to load commands and command aliases from embedded queries
// and from "repository" directories used by glazed.
//
// Examples of this pattern are used in sqleton, escuse-me and pinocchio.
type FSCommandLoader interface {
	LoadCommandsFromFS(f fs.FS, dir string, options []cmds.CommandDescriptionOption, aliasOptions []alias.Option) ([]cmds.Command, []*alias.CommandAlias, error)
}

func LoadCommandAliasFromYAML(s io.Reader, options ...alias.Option) ([]*alias.CommandAlias, error) {
	alias_, err := alias.NewCommandAliasFromYAML(s, options...)
	if err != nil {
		return nil, err
	}

	return []*alias.CommandAlias{alias_}, nil
}

// YAMLFSCommandLoader walks a FS and finds all yaml files, loading them using the passed
// YAMLCommandLoader.
//
// It handles the following generic functionality:
// - recursive FS walking
// - setting SourceName for each command
// - setting Parents for each command
type YAMLFSCommandLoader struct {
	loader YAMLCommandLoader
}

func NewYAMLFSCommandLoader(
	loader YAMLCommandLoader,
) *YAMLFSCommandLoader {
	return &YAMLFSCommandLoader{
		loader: loader,
	}
}

// LoadCommandsFromFS walks the FS and loads all commands and command aliases found.
//
// TODO(manuel, 2023-03-16) Add loading of helpsystem files
// See https://github.com/go-go-golems/glazed/issues/55
// See https://github.com/go-go-golems/glazed/issues/218
func (l *YAMLFSCommandLoader) LoadCommandsFromFS(f fs.FS, dir string, options []cmds.CommandDescriptionOption, aliasOptions []alias.Option) ([]cmds.Command, []*alias.CommandAlias, error) {
	var commands []cmds.Command
	var aliases []*alias.CommandAlias

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
			subCommands, subAliases, err := l.LoadCommandsFromFS(f, fileName, options, aliasOptions)
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
			commands_, err := func() ([]cmds.Command, error) {
				file, err := f.Open(fileName)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not open file %s", fileName)
				}
				defer func() {
					_ = file.Close()
				}()

				log.Debug().Str("file", fileName).Msg("Loading command from file")
				options_ := append([]cmds.CommandDescriptionOption{
					cmds.WithSource(fileName),
					cmds.WithParents(GetParentsFromDir(dir)...),
				}, options...)
				commands_, err := l.loader.LoadCommandFromYAML(file, options_...)
				if err != nil {
					log.Debug().Err(err).Str("file", fileName).Msg("Could not load command from file")
					return nil, err
				}
				if len(commands_) != 1 {
					return nil, errors.New("Expected exactly one command")
				}

				return commands_, err
			}()

			if err != nil {
				// If the error was a yaml parsing error, then we try to load the YAML file
				// again, but as an alias this time around. YAML / JSON parsing in golang
				// definitely is a bit of an adventure.
				if _, ok := err.(*yaml.TypeError); ok {
					aliases_, err := func() ([]*alias.CommandAlias, error) {
						file, err := f.Open(fileName)
						if err != nil {
							return nil, errors.Wrapf(err, "Could not open file %s", fileName)
						}
						defer func() {
							_ = file.Close()
						}()

						options_ := append(
							[]alias.Option{
								alias.WithSource(fileName),
								alias.WithParents(GetParentsFromDir(dir)...),
								alias.WithParents(GetParentsFromDir(dir)...),
							},
							aliasOptions...,
						)
						log.Debug().Str("file", fileName).Msg("Loading alias from file")
						aliases_, err := l.loader.LoadCommandAliasFromYAML(file, options_...)
						if err != nil {
							log.Debug().Err(err).Str("file", fileName).Msg("Could not load alias from file")
							return nil, err
						}
						if len(aliases_) != 1 {
							return nil, errors.New("Expected exactly one alias")
						}

						return aliases_, err
					}()
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Could not load command or alias from file %s: %s\n", fileName, err)
						continue
					} else {
						aliases = append(aliases, aliases_...)
					}
				}
				continue
			}

			commands = append(commands, commands_...)
		}
	}

	return commands, aliases, nil
}

// GetParentsFromDir is a helper function to simply return a list of parent verbs
// for applications loaded from declarative yaml files.
// The directory structure mirrors the verb structure in cobra.
func GetParentsFromDir(dir string) []string {
	// make sure both dir and rootDirectory have a trailing slash
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	pathToFile := dir
	parents := strings.Split(pathToFile, "/")
	if len(parents) > 0 && parents[len(parents)-1] == "" {
		parents = parents[:len(parents)-1]
	}
	return parents
}
