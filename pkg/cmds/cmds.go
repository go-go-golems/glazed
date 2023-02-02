package cmds

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/tj/go-naturaldate"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

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

type CommandDescription struct {
	Name      string       `yaml:"name"`
	Short     string       `yaml:"short"`
	Long      string       `yaml:"long,omitempty"`
	Flags     []*Parameter `yaml:"flags,omitempty"`
	Arguments []*Parameter `yaml:"arguments,omitempty"`

	Parents []string `yaml:",omitempty"`
	Source  string   `yaml:",omitempty"`
}

type Command interface {
	Run(map[string]interface{}) error
	Description() *CommandDescription
	// XXX(manuel, 2023-01-25) what about parents and source to load inside a cobra command tree
}

type CommandLoader interface {
	LoadCommandFromYAML(s io.Reader) ([]Command, error)
	LoadCommandAliasFromYAML(s io.Reader) ([]*CommandAlias, error)
}

// TODO(2022-12-21, manuel): Add list of choices as a type
// what about list of dates? list of bools?
// should list just be a flag?

type ParameterType string

const (
	ParameterTypeString         ParameterType = "string"
	ParameterTypeStringFromFile ParameterType = "stringFromFile"
	// load structure from json/yaml/csv file
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

func LoadCommandsFromEmbedFS(loader CommandLoader, f embed.FS, dir string, cmdRoot string) ([]Command, []*CommandAlias, error) {
	var commands []Command
	var aliases []*CommandAlias

	entries, err := f.ReadDir(dir)
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
			subCommands, _, err := LoadCommandsFromEmbedFS(loader, f, fileName, cmdRoot)
			if err != nil {
				return nil, nil, err
			}
			commands = append(commands, subCommands...)
		} else {
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
					commands, err := loader.LoadCommandFromYAML(file)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not load command from file %s", fileName)
					}
					if len(commands) != 1 {
						return nil, errors.New("Expected exactly one command")
					}
					command := commands[0]
					command.Description().Source = "embed:" + fileName

					pathToFile := strings.TrimPrefix(dir+"/", cmdRoot)
					parents := strings.Split(pathToFile, "/")
					if len(parents) > 0 && parents[len(parents)-1] == "" {
						parents = parents[:len(parents)-1]
					}
					command.Description().Parents = parents

					return command, err
				}()
				if err != nil {
					alias, err := func() (*CommandAlias, error) {
						file, err := f.Open(fileName)
						if err != nil {
							return nil, errors.Wrapf(err, "Could not open file %s", fileName)
						}
						defer func() {
							_ = file.Close()
						}()

						log.Debug().Str("file", fileName).Msg("Loading alias from file")
						aliases, err := loader.LoadCommandAliasFromYAML(file)
						if err != nil {
							return nil, err
						}
						if len(aliases) != 1 {
							return nil, errors.New("Expected exactly one alias")
						}
						alias := aliases[0]
						alias.Source = "embed:" + fileName

						pathToFile := strings.TrimPrefix(dir, cmdRoot)
						alias.Parents = strings.Split(pathToFile, "/")

						return alias, err
					}()

					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Could not load command or alias from file %s: %s\n", fileName, err)
						continue
					} else {
						aliases = append(aliases, alias)
					}
				} else {
					commands = append(commands, command)
				}
			}
		}
	}

	return commands, aliases, nil
}

func LoadCommandsFromDirectory(loader CommandLoader, dir string, cmdRoot string) ([]Command, []*CommandAlias, error) {
	var commands []Command
	var aliases []*CommandAlias

	entries, err := os.ReadDir(dir)
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
			subCommands, subAliases, err := LoadCommandsFromDirectory(loader, fileName, cmdRoot)
			if err != nil {
				return nil, nil, err
			}
			commands = append(commands, subCommands...)
			aliases = append(aliases, subAliases...)
		} else {
			if strings.HasSuffix(entry.Name(), ".yml") ||
				strings.HasSuffix(entry.Name(), ".yaml") {
				command, err := func() (Command, error) {
					file, err := os.Open(fileName)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not open file %s", fileName)
					}
					defer func() {
						_ = file.Close()
					}()

					log.Debug().Str("file", fileName).Msg("Loading command from file")
					commands, err := loader.LoadCommandFromYAML(file)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not load command from file %s", fileName)
					}
					if len(commands) != 1 {
						return nil, errors.New("Expected exactly one command")
					}
					command := commands[0]

					pathToFile := strings.TrimPrefix(dir, cmdRoot)
					pathToFile = strings.TrimPrefix(pathToFile, "/")
					command.Description().Parents = strings.Split(pathToFile, "/")

					command.Description().Source = "file:" + fileName

					return command, err
				}()
				if err != nil {
					alias, err := func() (*CommandAlias, error) {
						file, err := os.Open(fileName)
						if err != nil {
							return nil, errors.Wrapf(err, "Could not open file %s", fileName)
						}
						defer func() {
							_ = file.Close()
						}()

						log.Debug().Str("file", fileName).Msg("Loading alias from file")
						aliases, err := loader.LoadCommandAliasFromYAML(file)
						if err != nil {
							return nil, err
						}
						if len(aliases) != 1 {
							return nil, errors.New("Expected exactly one alias")
						}
						alias := aliases[0]

						alias.Source = "file:" + fileName

						pathToFile := strings.TrimPrefix(dir, cmdRoot)
						pathToFile = strings.TrimPrefix(pathToFile, "/")
						alias.Parents = strings.Split(pathToFile, "/")

						return alias, err
					}()

					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Could not load command or alias from file %s: %s\n", fileName, err)
						continue

					} else {
						aliases = append(aliases, alias)
					}
				} else {
					commands = append(commands, command)
				}
			}
		}
	}

	return commands, aliases, nil
}
