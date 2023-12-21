package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"gopkg.in/yaml.v3"
	"io"
)

// CommandDescription contains the necessary information for registering
// a command with cobra. Because a command gets registered in a verb tree,
// a full list of Parents all the way to the root needs to be provided.
type CommandDescription struct {
	Name  string `yaml:"name"`
	Short string `yaml:"short"`
	Long  string `yaml:"long,omitempty"`
	// TODO(manuel, 2023-12-21) Does this need to be a list of pointers? Could it just be a list of struct?
	Layout         []*layout.Section       `yaml:"layout,omitempty"`
	Layers         []layers.ParameterLayer `yaml:"layers,omitempty"`
	AdditionalData map[string]interface{}  `yaml:"additionalData,omitempty"`

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

func (c *CommandDescription) GetDefaultLayer() (layers.ParameterLayer, bool) {
	return c.GetLayer("default")
}

func (c *CommandDescription) GetDefaultFlags() []*parameters.ParameterDefinition {
	l, ok := c.GetDefaultLayer()
	if !ok {
		return nil
	}
	var flags []*parameters.ParameterDefinition
	for _, f := range l.GetParameterDefinitions() {
		if !f.IsArgument {
			flags = append(flags, f)
		}
	}
	return flags
}

func (c *CommandDescription) GetDefaultArguments() []*parameters.ParameterDefinition {
	l, ok := c.GetDefaultLayer()
	if !ok {
		return nil
	}
	var arguments []*parameters.ParameterDefinition
	for _, f := range l.GetParameterDefinitions() {
		if f.IsArgument {
			arguments = append(arguments, f)
		}
	}
	return arguments
}

func (c *CommandDescription) GetLayer(name string) (layers.ParameterLayer, bool) {
	for _, l := range c.Layers {
		if l.GetSlug() == name {
			return l, true
		}
	}
	return nil, false
}

func (c *CommandDescription) Clone(cloneLayers bool, options ...CommandDescriptionOption) *CommandDescription {
	// clone flags
	var layers_ []layers.ParameterLayer
	if cloneLayers {
		for _, l := range c.Layers {
			layers_ = append(layers_, l.Clone())
		}
	}

	// copy parents
	parents := make([]string, len(c.Parents))
	copy(parents, c.Parents)

	ret := &CommandDescription{
		Name:    c.Name,
		Short:   c.Short,
		Long:    c.Long,
		Layers:  layers_,
		Parents: parents,
		Source:  c.Source,
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

func WithLayers(l ...layers.ParameterLayer) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layers = append(c.Layers, l...)
	}
}

// WithFlags is a convenience function to add arguments to the default layer, useful
// to make the transition from explicit flags and arguments to a default layer a bit easier.
func WithFlags(
	flags ...*parameters.ParameterDefinition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		layer, ok := c.GetDefaultLayer()
		var err error
		if !ok {
			layer, err = layers.NewParameterLayer("default", "Default")
			if err != nil {
				panic(err)
			}
			c.Layers = append(c.Layers, layer)
		}
		layer.AddFlags(flags...)
	}
}

// WithArguments is a convenience function to add arguments to the default layer, useful
// to make the transition from explicit flags and arguments to a default layer a bit easier.
func WithArguments(
	arguments ...*parameters.ParameterDefinition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		layer, ok := c.GetDefaultLayer()
		var err error
		if !ok {
			layer, err = layers.NewParameterLayer("default", "Default")
			if err != nil {
				panic(err)
			}
			c.Layers = append(c.Layers, layer)
		}

		for _, arg := range arguments {
			arg.IsArgument = true
		}
		layer.AddFlags(arguments...)
	}
}

func WithDefaultLayer(
	flags []*parameters.ParameterDefinition,
	arguments []*parameters.ParameterDefinition,
) CommandDescriptionOption {

	for _, arg := range arguments {
		arg.IsArgument = true
	}
	return func(c *CommandDescription) {
		layer, err := layers.NewParameterLayer(
			"default",
			"Default",
			layers.WithFlags(flags...),
			layers.WithFlags(arguments...),
		)
		if err != nil {
			panic(err)
		}
		c.Layers = append(c.Layers, layer)
	}
}

func WithLayout(l *layout.Layout) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layout = l.Sections
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

func (cd *CommandDescription) ToYAML(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	defer func(enc *yaml.Encoder) {
		_ = enc.Close()
	}(enc)

	return enc.Encode(cd)
}

func (cd *CommandDescription) Description() *CommandDescription {
	return cd
}

type Command interface {
	Description() *CommandDescription
	ToYAML(w io.Writer) error
}

type CommandWithMetadata interface {
	Command
	Metadata(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
	) (map[string]interface{}, error)
}

// NOTE(manuel, 2023-03-17) Future types of commands that we could need
// - async emitting command (just strings, for example)
// - async emitting structured log
//   - async emitting of glaze rows (useful in general, and could be done with a special TableOutputFormatter, really)
// - no output (just do it yourself)
// - typed generic output structure (with error)

type BareCommand interface {
	Command
	Run(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
	) error
}

type WriterCommand interface {
	Command
	RunIntoWriter(
		ctx context.Context,
		parsedLayers map[string]*layers.ParsedParameterLayer,
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
		gp middlewares.Processor,
	) error
}

type ExitWithoutGlazeError struct{}

func (e *ExitWithoutGlazeError) Error() string {
	return "Exit without glaze"
}
