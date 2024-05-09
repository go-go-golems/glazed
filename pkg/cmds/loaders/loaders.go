package loaders

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CommandLoader is an interface that describes the most generic loader type,
// which is then used to load commands and command aliases from embedded queries
// and from "repository" directories used by glazed.
//
// Examples of this pattern are used in sqleton, escuse-me and pinocchio.
type CommandLoader interface {
	LoadCommands(
		f fs.FS, entryName string,
		options []cmds.CommandDescriptionOption,
		aliasOptions []alias.Option,
	) ([]cmds.Command, error)
	IsFileSupported(f fs.FS, fileName string) bool
}

type LoadReaderCommandFunc func(
	r io.Reader,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error)

func LoadCommandOrAliasFromReader(
	r io.Reader,
	rawLoadCommand LoadReaderCommandFunc,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	br := strings.NewReader(string(bytes))
	cmds_, err := rawLoadCommand(br, options, aliasOptions)
	if err != nil {
		br = strings.NewReader(string(bytes))
		aliases, errAlias := LoadCommandAliasFromYAML(br, aliasOptions...)
		if errAlias != nil {
			// return the first error, as it's probably more salient
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

func LoadCommandAliasFromYAML(s io.Reader, options ...alias.Option) ([]*alias.CommandAlias, error) {
	alias_, err := alias.NewCommandAliasFromYAML(s, options...)
	if err != nil {
		return nil, err
	}

	return []*alias.CommandAlias{alias_}, nil
}

// LoadCommandsFromFS walks the FS and loads all commands and command aliases found.
//
// TODO(manuel, 2023-03-16) Add loading of helpsystem files
// See https://github.com/go-go-golems/glazed/issues/55
// See https://github.com/go-go-golems/glazed/issues/218
func LoadCommandsFromFS(
	f fs.FS,
	dir string,
	source string,
	loader CommandLoader,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	var commands []cmds.Command

	entries, err := fs.ReadDir(f, dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		// skip hidden files
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		fileName := filepath.Join(dir, entry.Name())

		// NOTE(2023-02-07, manuel) This might benefit from being made more generic than just loading from YAML
		//
		// One problem with the "commands from YAML" pattern being defined in glazed
		// is that is actually not great for a more complex application like pinocchio which
		// would benefit from loading applications from entire directories.
		//
		// This can of course be solved by providing a CommandLoader for directories.
		//
		// Similarly, we might want to store applications in a database, or generate them on the
		// fly using some resources on the disk.
		//
		// See https://github.com/go-go-golems/glazed/issues/116
		if loader.IsFileSupported(f, fileName) {
			fromDir := GetParentsFromDir(dir)
			commands_, err := func() ([]cmds.Command, error) {
				log.Debug().Str("file", fileName).Msg("Loading command from file")
				options_ := append([]cmds.CommandDescriptionOption{
					cmds.WithSource(source + "/" + fileName),
					cmds.WithParents(fromDir...),
				}, options...)
				aliasOptions_ := append([]alias.Option{
					alias.WithSource(source + "/" + fileName),
					alias.WithParents(fromDir...),
				}, aliasOptions...)
				commands_, err_ := loader.LoadCommands(f, fileName, options_, aliasOptions_)
				if err_ != nil {
					log.Debug().Err(err_).Str("file", fileName).Msg("Could not load command from file")
					return nil, err_
				}

				return commands_, err_
			}()
			if err != nil {
				log.Warn().Err(err).Str("file", fileName).Msg("Could not load command from file")
				continue
			}

			commands = append(commands, commands_...)
			continue
		}

		if entry.IsDir() {
			subCommands, err := LoadCommandsFromFS(f, fileName, source, loader, options, aliasOptions)
			if err != nil {
				return nil, err
			}
			commands = append(commands, subCommands...)
			continue
		}
	}

	return commands, nil
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

func FileNameToFsFilePath(fileName string) (fs.FS, string, error) {
	// get absolute path from config_.File
	fs_ := os.DirFS("/")

	cleanedPath := filepath.Clean(fileName)

	var filePath string
	switch {
	case strings.HasPrefix(fileName, "/"):
		filePath = fileName[1:]
	case strings.HasPrefix(fileName, "./"):
		fs_ = os.DirFS(".")
		filePath = fileName[2:]
	case strings.HasPrefix(fileName, "../"):
		var upCount int
		relPath := cleanedPath
		for strings.HasPrefix(relPath, "../") {
			upCount++
			relPath = strings.TrimPrefix(relPath, "../")
		}

		// Walk up the directory tree as needed
		currentDir, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}

		for i := 0; i < upCount; i++ {
			currentDir = filepath.Dir(currentDir)
		}

		fileSystem := os.DirFS(currentDir)
		return fileSystem, relPath, nil
	default:
		fs_ = os.DirFS(".")
		filePath = fileName
	}
	return fs_, filePath, nil
}
