package cliopatra

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// Field describes a cliopatra field, which can be either a flag or an argument.
// It does mirror the definition of fields.Definition, but here we only
// have a Value, and a Short description (which should actually describe which value we chose).
//
// The Flag makes it possible to override the flag used on the CLI, if necessary.
// The Raw field makes it possible to pass a raw string to override the value being rendered
// out. This is useful to for example test invalid value for flags.
type Field struct {
	Name       string             `yaml:"name"`
	Flag       string             `yaml:"flag,omitempty"`
	Short      string             `yaml:"short"`
	Type       fields.Type        `yaml:"type"`
	Value      interface{}        `yaml:"value"`
	Raw        string             `yaml:"raw,omitempty"`
	NoValue    bool               `yaml:"noValue,omitempty"`
	IsArgument bool               `yaml:"isArgument,omitempty"`
	Log        []fields.ParseStep `yaml:"log,omitempty"`
}

// NOTE(manuel, 2023-03-16) What about sandboxing the execution of the command, especially if it outputs files
// NOTE(manuel, 2023-03-16) It would be interesting to provide some more tests on the output (say, as shell scripts)
// NOTE(manuel, 2023-03-16) What about measuring profiling regression

func (p *Field) Clone() *Field {
	p_ := *p
	return &p_
}

// Program describes a program to be executed by cliopatra.
//
// This can be used for golden tests by providing
// and ExpectedStdout, ExpectedError, ExpectedStatusCode and ExpectedFiles.
type Program struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path,omitempty"`
	Verbs       []string `yaml:"verbs,omitempty"`
	Description string   `yaml:"description"`
	// Env makes it possible to specify environment variables to set manually
	Env map[string]string `yaml:"env,omitempty"`

	// TODO(manuel, 2023-03-16) Probably add RawFlags here, when we say quickly want to record a run.
	// Of course, if we are using Command, we could have that render a more precisely described
	// cliopatra file. But just capturing normal calls is nice too.
	RawFlags []string `yaml:"rawFlags,omitempty"`

	// These Flags will be passed to the CLI tool. This allows us to register
	// flags with a type to cobra itself, when exposing this command again.
	Flags []*Field `yaml:"flags,omitempty"`
	// Args is an ordered list of fields. The Flag field is ignored.
	Args []*Field `yaml:"args,omitempty"`
	// Stdin makes it possible to pass data into stdin. If empty, no data is passed.
	Stdin string `yaml:"stdin,omitempty"`

	// These fields are useful for golden testing.
	ExpectedStdout     string            `yaml:"expectedStdout,omitempty"`
	ExpectedError      string            `yaml:"expectedError,omitempty"`
	ExpectedStatusCode int               `yaml:"expectedStatusCode,omitempty"`
	ExpectedFiles      map[string]string `yaml:"expectedFiles,omitempty"`
}

//var _ glazedcmds.WriterCommand = (*Program)(nil)

type ProgramOption func(*Program)

func WithName(name string) ProgramOption {
	return func(p *Program) {
		p.Name = name
	}
}

func WithPath(path string) ProgramOption {
	return func(p *Program) {
		p.Path = path
	}
}

func WithVerbs(verbs ...string) ProgramOption {
	return func(p *Program) {
		p.Verbs = verbs
	}
}

func WithDescription(description string) ProgramOption {
	return func(p *Program) {
		p.Description = description
	}
}

func WithEnv(env map[string]string) ProgramOption {
	return func(p *Program) {
		p.Env = env
	}
}

func WithAddEnv(key, value string) ProgramOption {
	return func(p *Program) {
		if p.Env == nil {
			p.Env = make(map[string]string)
		}
		p.Env[key] = value
	}
}

func WithRawFlags(flags ...string) ProgramOption {
	return func(p *Program) {
		p.RawFlags = flags
	}
}

func WithAddRawFlags(flags ...string) ProgramOption {
	return func(p *Program) {
		p.RawFlags = append(p.RawFlags, flags...)
	}
}

func WithFlags(flags ...*Field) ProgramOption {
	return func(p *Program) {
		p.Flags = flags
	}
}

func WithAddFlags(flags ...*Field) ProgramOption {
	return func(p *Program) {
		p.Flags = append(p.Flags, flags...)
	}
}

func WithReplaceFlags(flags ...*Field) ProgramOption {
	return func(p *Program) {
		for _, flag := range flags {
			found := false
			for i, existing := range p.Flags {
				if existing.Name == flag.Name {
					p.Flags[i] = flag
					found = true
					break
				}
			}
			if !found {
				p.Flags = append(p.Flags, flag)
			}
		}
	}
}

func WithArgs(args ...*Field) ProgramOption {
	return func(p *Program) {
		p.Args = args
	}
}

func WithAddArgs(args ...*Field) ProgramOption {
	return func(p *Program) {
		p.Args = append(p.Args, args...)
	}
}

func WithReplaceArgs(args ...*Field) ProgramOption {
	return func(p *Program) {
		for _, arg := range args {
			found := false
			for i, existing := range p.Args {
				if existing.Name == arg.Name {
					p.Args[i] = arg
					found = true
					break
				}
			}
			if !found {
				p.Args = append(p.Args, arg)
			}
		}
	}
}

func WithStdin(stdin string) ProgramOption {
	return func(p *Program) {
		p.Stdin = stdin
	}
}

func WithExpectedStdout(stdout string) ProgramOption {
	return func(p *Program) {
		p.ExpectedStdout = stdout
	}
}

func WithExpectedError(err string) ProgramOption {
	return func(p *Program) {
		p.ExpectedError = err
	}
}

func WithExpectedStatusCode(statusCode int) ProgramOption {
	return func(p *Program) {
		p.ExpectedStatusCode = statusCode
	}
}

func WithExpectedFiles(files map[string]string) ProgramOption {
	return func(p *Program) {
		p.ExpectedFiles = files
	}
}

func NewProgram(opts ...ProgramOption) *Program {
	p := &Program{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func NewProgramFromYAML(s io.Reader, opts ...ProgramOption) (*Program, error) {
	var program Program
	if err := yaml.NewDecoder(s).Decode(&program); err != nil {
		return nil, errors.Wrap(err, "could not decode program")
	}
	for _, opt := range opts {
		opt(&program)
	}
	return &program, nil
}

func (p *Program) Clone() *Program {
	clone := *p

	clone.RawFlags = make([]string, len(p.RawFlags))
	copy(clone.RawFlags, p.RawFlags)
	clone.Flags = make([]*Field, len(p.Flags))
	for i, f := range p.Flags {
		clone.Flags[i] = f.Clone()
	}
	clone.Args = make([]*Field, len(p.Args))
	for i, a := range p.Args {
		clone.Args[i] = a.Clone()
	}
	clone.Env = make(map[string]string, len(p.Env))
	for k, v := range p.Env {
		clone.Env[k] = v
	}

	clone.ExpectedFiles = make(map[string]string, len(p.ExpectedFiles))
	for k, v := range p.ExpectedFiles {
		clone.ExpectedFiles[k] = v
	}

	return &clone
}

func (p *Program) SetFlagValue(name string, value interface{}) error {
	for _, f := range p.Flags {
		if f.Name == name {
			f.Value = value
			return nil
		}
	}

	return errors.Errorf("could not find flag %s", name)
}

func (p *Program) SetFlagRaw(name string, raw string) error {
	for _, f := range p.Flags {
		if f.Name == name {
			f.Raw = raw
			return nil
		}
	}

	return errors.Errorf("could not find flag %s", name)
}

func (p *Program) SetArgValue(name string, value interface{}) error {
	for _, a := range p.Args {
		if a.Name == name {
			a.Value = value
			return nil
		}
	}

	return errors.Errorf("could not find arg %s", name)
}

func (p *Program) SetArgRaw(name string, raw string) error {
	for _, a := range p.Args {
		if a.Name == name {
			a.Raw = raw
			return nil
		}
	}

	return errors.Errorf("could not find arg %s", name)
}

func (p *Program) AddRawFlag(raw ...string) {
	p.RawFlags = append(p.RawFlags, raw...)
}

func (p *Program) RunIntoWriter(
	ctx context.Context,
	parsedValues *values.Values,
	w io.Writer,
) error {
	var err error
	path := p.Path
	if path == "" {
		path, err = exec.LookPath(p.Name)
		if err != nil {
			return errors.Wrapf(err, "could not find executable %s", p.Name)
		}
	}

	ps := parsedValues.AllFieldValues()

	args, err2 := p.ComputeArgs(ps)
	if err2 != nil {
		return err2
	}

	log.Debug().Str("path", path).Strs("args", args).Msg("running program")

	// Validate the path to prevent command injection
	// Use filepath.Clean to normalize the path
	cleanPath := filepath.Clean(path)

	// Check if the path doesn't try to escape with "../" or is not absolute
	if !filepath.IsAbs(cleanPath) {
		// If it's not absolute, resolve it through PATH lookup
		var err error
		cleanPath, err = exec.LookPath(cleanPath)
		if err != nil {
			return errors.Wrapf(err, "invalid program path: %s", path)
		}
	}

	// Check if the file exists and is executable
	info, err := os.Stat(cleanPath)
	if err != nil {
		return errors.Wrapf(err, "cannot access program path: %s", cleanPath)
	}

	// Check if it's a regular file and is executable (on Unix systems)
	if info.IsDir() {
		return errors.Errorf("program path is a directory: %s", cleanPath)
	}

	// On Unix, check file mode for executable bit
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return errors.Errorf("program is not executable: %s", cleanPath)
		}
	}

	cmd := exec.CommandContext(ctx, cleanPath, args...)
	cmd.Env = []string{}
	// copy current environment
	cmd.Env = append(cmd.Env, os.Environ()...)
	for k, v := range p.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	log.Trace().Strs("env", cmd.Env).Msg("environment")

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

func (p *Program) ComputeArgs(ps *fields.FieldValues) ([]string, error) {
	var err error

	args := []string{}

	args = append(args, p.Verbs...)

	// I'm not sure how useful this raw flags mixed with the other stuff is at all.
	// I don't think both together make much sense, maybe we should differentiate
	// at a higher level, so that it is either RawFlags, or all the rest
	args = append(args, p.RawFlags...)

	for _, flag := range p.Flags {
		flag_ := flag.Flag
		if flag_ == "" {
			flag_ = "--" + flag.Name
		}
		if flag.NoValue {
			if flag.Type != fields.TypeBool {
				return nil, errors.Errorf("flag %s is not a bool flag, only bool flags can be noValue", flag.Name)
			}
			if flag.Value.(bool) {
				args = append(args, flag_)
			}
			continue
		}

		fieldValue, ok := ps.Get(flag.Name)
		value_ := ""
		if !ok {
			value_ = flag.Raw
		} else {
			value_, err = fieldValue.RenderValue()
			if err != nil {
				return nil, errors.Wrapf(err, "could not render flag %s", flag.Name)
			}
		}

		if value_ == "" {
			value_, err = fields.RenderValue(flag.Type, flag.Value)
			if err != nil {
				return nil, errors.Wrapf(err, "could not render flag %s", flag.Name)
			}
		}
		args = append(args, flag_)
		args = append(args, value_)
	}

	for _, arg := range p.Args {
		value, ok := ps.Get(arg.Name)
		value_ := ""
		if !ok {
			value_ = arg.Raw
		} else {
			value_, err = fields.RenderValue(arg.Type, value.Value)
			if err != nil {
				return nil, errors.Wrapf(err, "could not render arg %s", arg.Name)
			}
		}

		if value_ == "" {
			value_, err = fields.RenderValue(arg.Type, arg.Value)
			if err != nil {
				return nil, errors.Wrapf(err, "could not render arg %s", arg.Name)
			}
		}
		args = append(args, value_)
	}
	return args, nil
}
