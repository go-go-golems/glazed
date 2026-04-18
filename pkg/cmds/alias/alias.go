package alias

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Option func(*CommandAlias)

type AliasTarget []string

func NewAliasTarget(parts ...string) AliasTarget {
	ret := AliasTarget{}
	for _, part := range parts {
		ret = append(ret, splitAliasTargetPart(part)...)
	}
	return ret
}

func NewAliasTargetFromString(aliasFor string) AliasTarget {
	return NewAliasTarget(aliasFor)
}

func (t AliasTarget) IsZero() bool {
	return len(t) == 0
}

func (t AliasTarget) Segments() []string {
	return append([]string(nil), t...)
}

func (t AliasTarget) String() string {
	return strings.Join(t, " ")
}

func (t *AliasTarget) UnmarshalYAML(node *yaml.Node) error {
	if node == nil {
		*t = nil
		return nil
	}

	ret := AliasTarget{}
	switch node.Kind {
	case yaml.ScalarNode:
		ret = append(ret, splitAliasTargetPart(node.Value)...)
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if child.Kind != yaml.ScalarNode {
				return errors.Errorf("aliasFor entries must be scalar, got YAML kind %d", child.Kind)
			}
			ret = append(ret, splitAliasTargetPart(child.Value)...)
		}
	case yaml.DocumentNode, yaml.MappingNode, yaml.AliasNode:
		return errors.Errorf("aliasFor must be a scalar or sequence, got YAML kind %d", node.Kind)
	default:
		return errors.Errorf("aliasFor has unsupported YAML kind %d", node.Kind)
	}

	if len(ret) == 0 {
		return errors.New("aliasFor must not be empty")
	}
	*t = ret
	return nil
}

func (t AliasTarget) MarshalYAML() (interface{}, error) {
	if len(t) == 0 {
		return nil, nil
	}
	if len(t) == 1 {
		return t[0], nil
	}
	return []string(t), nil
}

func splitAliasTargetPart(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '/' || r == '\\' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	ret := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		ret = append(ret, part)
	}
	return ret
}

// CommandAlias defines a struct that should be able to define generic aliases
// for any kind of command line applications, by providing overrides for certain
// flags (prepopulating them with certain flags and arguments, basically)
type CommandAlias struct {
	Name      string            `yaml:"name"`
	AliasFor  AliasTarget       `yaml:"aliasFor"`
	Short     string            `yaml:"short,omitempty"`
	Long      string            `yaml:"long,omitempty"`
	Flags     map[string]string `yaml:"flags,omitempty"`
	Arguments []string          `yaml:"arguments,omitempty"`
	Layout    []*layout.Section `yaml:"layout,omitempty"`

	AliasedCommand cmds.Command `yaml:",omitempty"`
	Parents        []string     `yaml:",omitempty"`
	Source         string       `yaml:",omitempty"`
}

var _ cmds.Command = (*CommandAlias)(nil)

func WithName(name string) Option {
	return func(a *CommandAlias) {
		a.Name = name
	}
}

func WithAliasFor(aliasFor string) Option {
	return func(a *CommandAlias) {
		a.AliasFor = NewAliasTargetFromString(aliasFor)
	}
}

func WithAliasForPath(parts ...string) Option {
	return func(a *CommandAlias) {
		a.AliasFor = NewAliasTarget(parts...)
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

var _ cmds.GlazeCommand = (*CommandAlias)(nil)
var _ cmds.WriterCommand = (*CommandAlias)(nil)

func (a *CommandAlias) String() string {
	return fmt.Sprintf("CommandAlias{Name: %s, AliasFor: %s, Parents: %v, Source: %s}",
		a.Name, a.AliasFor.String(), a.Parents, a.Source)
}

func (a *CommandAlias) ResolveAliasedCommandPath() []string {
	segments := a.AliasFor.Segments()
	if len(segments) <= 1 {
		return append(append([]string{}, a.Parents...), segments...)
	}
	return segments
}

func (a *CommandAlias) ToYAML(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	defer func(enc *yaml.Encoder) {
		_ = enc.Close()
	}(enc)

	return enc.Encode(a)
}

func (a *CommandAlias) RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	glazeCommand, ok := a.AliasedCommand.(cmds.GlazeCommand)
	if !ok {
		return errors.New("aliased command is not a GlazeCommand")
	}
	return glazeCommand.RunIntoGlazeProcessor(ctx, parsedValues, gp)
}

func (a *CommandAlias) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	if a.AliasedCommand == nil {
		return errors.New("no aliased command")
	}
	writerCommand, ok := a.AliasedCommand.(cmds.WriterCommand)
	if !ok {
		return errors.New("aliased command is not a GlazeCommand")
	}
	return writerCommand.RunIntoWriter(ctx, parsedValues, w)
}

func (a *CommandAlias) IsValid() bool {
	return a.Name != "" && !a.AliasFor.IsZero()
}

// Description returns the CommandDescription of an alias.
// It computes it at runtime by loading the aliased command's Description() and
// making copies of its flags and arguments.
// This is necessary because they get mutated at runtime with various defaults,
// depending on where they come from.
func (a *CommandAlias) Description() *cmds.CommandDescription {
	if a.AliasedCommand == nil {
		return nil
	}
	s := a.AliasedCommand.Description()
	layout_ := a.Layout
	if layout_ == nil {
		layout_ = s.Layout
	}

	newSchema := s.Schema.Clone()

	ret := &cmds.CommandDescription{
		Name:           a.Name,
		Short:          s.Short,
		Long:           s.Long,
		Layout:         layout_,
		Schema:         newSchema,
		Parents:        a.Parents,
		Source:         a.Source,
		Type:           s.Type,
		Tags:           s.Tags,
		Metadata:       s.Metadata,
		AdditionalData: s.AdditionalData,
	}

	return ret
}
