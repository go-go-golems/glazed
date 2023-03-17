package cliopatra

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"os/exec"
	"strings"
)

// Parameter describes a cliopatra parameter, which can be either a flag or an argument.
// It does mirror the definition of parameters.ParameterDefinition, but here we only
// have a Value, and a Short description (which should actually describe which value we chose).
//
// The Flag makes it possible to override the flag used on the CLI, if necessary.
// The Raw field makes it possible to pass a raw string to override the value being rendered
// out. This is useful to for example test invalid value for flags.
type Parameter struct {
	Name    string                   `yaml:"name"`
	Flag    string                   `yaml:"flag,omitempty"`
	Short   string                   `yaml:"short"`
	Type    parameters.ParameterType `yaml:"type"`
	Value   interface{}              `yaml:"value"`
	Raw     string                   `yaml:"raw,omitempty"`
	NoValue bool                     `yaml:"noValue,omitempty"`
}

// NOTE(manuel, 2023-03-16) What about sandboxing the execution of the command, especially if it outputs files
// NOTE(manuel, 2023-03-16) It would be interesting to provide some more tests on the output (say, as shell scripts)
// NOTE(manuel, 2023-03-16) What about measuring profiling regression

// Program describes a program to be executed by cliopatra.
//
// This can be used for golden tests by providing the
type Program struct {
	Name        string `yaml:"name"`
	Path        string `yaml:"path,omitempty"`
	Description string `yaml:"description"`
	// Env makes it possible to specify environment variables to set manually
	Env map[string]string `yaml:"env,omitempty"`
	// These Flags will be passed to the CLI tool
	Flags []*Parameter `yaml:"flags,omitempty"`
	// Args is an ordered list of Parameters. The Flag field is ignored.
	Args []*Parameter `yaml:"args,omitempty"`
	// Stdin makes it possible to pass data into stdin. If empty, no data is passed.
	Stdin string `yaml:"stdin,omitempty"`

	// These fields are useful for golden testing.
	ExpectedStdout     string            `yaml:"expectedStdout,omitempty"`
	ExpectedError      string            `yaml:"expectedError,omitempty"`
	ExpectedStatusCode int               `yaml:"expectedStatusCode,omitempty"`
	ExpectedFiles      map[string]string `yaml:"expectedFiles,omitempty"`
}

func NewProgramFromYAML(s io.Reader) (*Program, error) {
	var program Program
	if err := yaml.NewDecoder(s).Decode(&program); err != nil {
		return nil, errors.Wrap(err, "could not decode program")
	}
	return &program, nil
}

func (p *Program) RunIntoWriter(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	w io.Writer) error {
	var err error
	path := p.Path
	if path == "" {
		path, err = exec.LookPath(p.Name)
		if err != nil {
			return errors.Wrapf(err, "could not find executable %s", p.Name)
		}
	}

	args := []string{}
	for _, flag := range p.Flags {
		flag_ := flag.Flag
		if flag_ == "" {
			flag_ = "--" + flag.Name
		}
		if flag.NoValue {
			args = append(args, flag_)
			continue
		}

		value, ok := ps[flag.Name]
		value_ := ""
		if !ok {
			value_ = flag.Raw
		} else {
			value_, err = RenderParameter(flag.Type, value)
			if err != nil {
				return errors.Wrapf(err, "could not render flag %s", flag.Name)
			}
		}

		if value_ == "" {
			value_, err = RenderParameter(flag.Type, flag.Value)
			if err != nil {
				return errors.Wrapf(err, "could not render flag %s", flag.Name)
			}
		}
		args = append(args, flag_)
		args = append(args, value_)
	}

	for _, arg := range p.Args {
		value, ok := ps[arg.Name]
		value_ := ""
		if !ok {
			value_ = arg.Raw
		} else {
			value_, err = RenderParameter(arg.Type, value)
			if err != nil {
				return errors.Wrapf(err, "could not render arg %s", arg.Name)
			}
		}

		if value_ == "" {
			value_, err = RenderParameter(arg.Type, arg.Value)
			if err != nil {
				return errors.Wrapf(err, "could not render arg %s", arg.Name)
			}
		}
		args = append(args, value_)
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = []string{}
	for k, v := range p.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	if p.Stdin != "" {
		cmd.Stdin = strings.NewReader(p.Stdin)
	}
	cmd.Stdout = w
	cmd.Stderr = w
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "could not run %s", p.Name)
	}

	return nil
}

func RenderParameter(type_ parameters.ParameterType, value interface{}) (string, error) {
	switch type_ {
	case parameters.ParameterTypeString:
		fallthrough
	case parameters.ParameterTypeStringFromFile:
		fallthrough
	case parameters.ParameterTypeObjectListFromFile:
		fallthrough
	case parameters.ParameterTypeObjectFromFile:
		fallthrough
	case parameters.ParameterTypeStringListFromFile:
		fallthrough
	case parameters.ParameterTypeDate:
		fallthrough
	case parameters.ParameterTypeChoice:
		s, ok := value.(string)
		if !ok {
			return "", errors.Errorf("expected string, got %T", value)
		}
		return s, nil

	case parameters.ParameterTypeKeyValue:
		m, ok := value.(map[string]string)
		if !ok {
			return "", errors.Errorf("expected map[string]string, got %T", value)
		}
		s := []string{}
		for k, v := range m {
			s = append(s, k+":"+v)
		}
		return strings.Join(s, ","), nil

	case parameters.ParameterTypeInteger:
		return fmt.Sprintf("%d", value), nil

	case parameters.ParameterTypeFloat:
		return fmt.Sprintf("%f", value), nil

	case parameters.ParameterTypeBool:
		v, ok := value.(bool)
		if !ok {
			return "", errors.Errorf("expected bool, got %T", value)
		}
		if v {
			return "true", nil
		}
		return "false", nil

	case parameters.ParameterTypeStringList:
		l, ok := value.([]string)
		if !ok {
			return "", errors.Errorf("expected []string, got %T", value)
		}
		return strings.Join(l, ","), nil

	case parameters.ParameterTypeIntegerList:
		v, ok := value.([]interface{})
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		l, ok := helpers.CastInterfaceListToIntList[int64](v)
		if !ok {
			return "", errors.Errorf("expected []int64, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%d", i))
		}
		return strings.Join(s, ","), nil

	case parameters.ParameterTypeFloatList:
		v, ok := value.([]interface{})
		if !ok {
			return "", errors.Errorf("expected []interface{}, got %T", value)
		}
		l, ok := helpers.CastInterfaceListToFloatList[float64](v)
		if !ok {
			return "", errors.Errorf("expected []float64, got %T", value)
		}
		s := []string{}
		for _, i := range l {
			s = append(s, fmt.Sprintf("%f", i))
		}
		return strings.Join(s, ","), nil
	}

	return "", errors.Errorf("unknown type %s", type_)
}
