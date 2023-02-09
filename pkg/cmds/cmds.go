package cmds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Parameter is a declarative way of describing a command line parameter.
// A Parameter can be either a Flag or an Argument.
// Along with metadata (Name, Help) that is useful for help,
// it also specifies a Type, a Default value and if it is Required.
type Parameter struct {
	Name      string        `yaml:"name"`
	ShortFlag string        `yaml:"shortFlag,omitempty"`
	Type      ParameterType `yaml:"type"`
	Help      string        `yaml:"help,omitempty"`
	Default   interface{}   `yaml:"default,omitempty"`
	Choices   []string      `yaml:"choices,omitempty"`
	Required  bool          `yaml:"required,omitempty"`
}

func (p *Parameter) Copy() *Parameter {
	return &Parameter{
		Name:      p.Name,
		ShortFlag: p.ShortFlag,
		Type:      p.Type,
		Help:      p.Help,
		Default:   p.Default,
		Choices:   p.Choices,
		Required:  p.Required,
	}
}

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

	Run(map[string]interface{}) error
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

type ParameterType string

const (
	ParameterTypeString         ParameterType = "string"
	ParameterTypeStringFromFile ParameterType = "stringFromFile"

	// TODO (2023-02-07) It would be great to have "list of objects from file" here
	// See https://github.com/go-go-golems/glazed/issues/117

	// ParameterTypeObjectFromFile - loader structure from json/yaml/csv file
	ParameterTypeObjectFromFile ParameterType = "objectFromFile"
	ParameterTypeInteger        ParameterType = "int"
	ParameterTypeFloat          ParameterType = "float"
	ParameterTypeBool           ParameterType = "bool"
	ParameterTypeDate           ParameterType = "date"
	ParameterTypeStringList     ParameterType = "stringList"
	ParameterTypeIntegerList    ParameterType = "intList"
	ParameterTypeFloatList      ParameterType = "floatList"
	ParameterTypeChoice         ParameterType = "choice"
)

func (p *Parameter) CheckParameterDefaultValueValidity() error {
	// we can have no default
	if p.Default == nil {
		return nil
	}

	switch p.Type {
	case ParameterTypeString:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeStringFromFile:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeObjectFromFile:
		_, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

	case ParameterTypeInteger:
		_, ok := p.Default.(int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, p.Default)
		}

	case ParameterTypeFloat:
		_, ok := p.Default.(int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer: %v", p.Name, p.Default)
		}

	case ParameterTypeBool:
		_, ok := p.Default.(bool)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a bool: %v", p.Name, p.Default)
		}

	case ParameterTypeDate:
		defaultValue, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

		_, err2 := parseDate(defaultValue)
		if err2 != nil {
			return errors.Wrapf(err2, "Default value for parameter %s is not a valid date: %v", p.Name, p.Default)
		}

	case ParameterTypeStringList:
		_, ok := p.Default.([]string)
		if !ok {
			defaultValue, ok := p.Default.([]interface{})
			if !ok {
				return errors.Errorf("Default value for parameter %s is not a string list: %v", p.Name, p.Default)
			}

			// convert to string list
			fixedDefault, err := convertToStringList(defaultValue)
			if err != nil {
				return errors.Wrapf(err, "Could not convert default value for parameter %s to string list: %v", p.Name, p.Default)
			}
			p.Default = fixedDefault
		}

	case ParameterTypeIntegerList:
		_, ok := p.Default.([]int)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not an integer list: %v", p.Name, p.Default)
		}

	case ParameterTypeFloatList:
		_, ok := p.Default.([]float32)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a float list: %v", p.Name, p.Default)
		}

	case ParameterTypeChoice:
		if len(p.Choices) == 0 {
			return errors.Errorf("Parameter %s is a choice parameter but has no choices", p.Name)
		}

		defaultValue, ok := p.Default.(string)
		if !ok {
			return errors.Errorf("Default value for parameter %s is not a string: %v", p.Name, p.Default)
		}

		found := false
		for _, choice := range p.Choices {
			if choice == defaultValue {
				found = true
			}
		}
		if !found {
			return errors.Errorf("Default value for parameter %s is not a valid choice: %v", p.Name, p.Default)
		}
	}

	return nil
}

func (p *Parameter) ParseParameter(v []string) (interface{}, error) {
	if len(v) == 0 {
		if p.Required {
			return nil, errors.Errorf("Argument %s not found", p.Name)
		} else {
			return p.Default, nil
		}
	}

	switch p.Type {
	case ParameterTypeString:
		return v[0], nil
	case ParameterTypeInteger:
		i, err := strconv.Atoi(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
		}
		return i, nil
	case ParameterTypeStringList:
		return v, nil
	case ParameterTypeIntegerList:
		ints := make([]int, 0)
		for _, arg := range v {
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			ints = append(ints, i)
		}
		return ints, nil

	case ParameterTypeBool:
		b, err := strconv.ParseBool(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as bool", p.Name)
		}
		return b, nil

	case ParameterTypeChoice:
		choice := v[0]
		found := false
		for _, c := range p.Choices {
			if c == choice {
				found = true
			}
		}
		if !found {
			return nil, errors.Errorf("Argument %s has invalid choice %s", p.Name, choice)
		}
		return choice, nil

	case ParameterTypeDate:
		parsedDate, err := parseDate(v[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse argument %s as date", p.Name)
		}
		return parsedDate, nil

	case ParameterTypeObjectFromFile:
		fileName := v[0]
		f, err := os.Open(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}

		object := interface{}(nil)
		if strings.HasSuffix(fileName, ".json") {
			err = json.NewDecoder(f).Decode(&object)
		} else if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
			err = yaml.NewDecoder(f).Decode(&object)
		} else {
			return nil, errors.Errorf("Could not parse file %s: unknown file type", fileName)
		}

		if err != nil {
			return nil, errors.Wrapf(err, "Could not parse file %s", fileName)
		}

		return object, nil

	case ParameterTypeStringFromFile:
		fileName := v[0]
		if fileName == "-" {
			var b bytes.Buffer
			_, err := io.Copy(&b, os.Stdin)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not read from stdin")
			}
			return b.String(), nil
		}

		bs, err := os.ReadFile(fileName)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not read file %s", v[0])
		}
		return string(bs), nil

	case ParameterTypeFloatList:
		floats := make([]float64, 0)
		for _, arg := range v {
			// parse to float
			f, err := strconv.ParseFloat(arg, 64)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not parse argument %s as integer", p.Name)
			}
			floats = append(floats, f)
		}
		return floats, nil
	}

	return nil, errors.Errorf("Unknown parameter type %s", p.Type)
}

func convertToStringList(value []interface{}) ([]string, error) {
	stringList := make([]string, len(value))
	for i, v := range value {
		s, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("Not a string: %v", v)
		}
		stringList[i] = s
	}
	return stringList, nil
}

func parseDate(value string) (time.Time, error) {
	parsedDate, err := dateparse.ParseAny(value)
	if err != nil {
		refTime_ := time.Now()
		if refTime != nil {
			refTime_ = *refTime
		}
		parsedDate, err = naturaldate.Parse(value, refTime_)
		if err != nil {
			return time.Time{}, errors.Wrapf(err, "Could not parse date: %s", value)
		}
	}

	return parsedDate, nil
}

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
