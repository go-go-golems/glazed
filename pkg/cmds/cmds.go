package cmds

import (
	"context"
	"io"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"gopkg.in/yaml.v3"
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
	Layers         *layers.ParameterLayers `yaml:"layers,omitempty"`
	AdditionalData map[string]interface{}  `yaml:"additionalData,omitempty"`
	Type           string                  `yaml:"type,omitempty"`
	Tags           []string                `yaml:"tags,omitempty"`
	Metadata       map[string]interface{}  `yaml:"metadata,omitempty"`

	Parents []string `yaml:",omitempty"`
	// Source indicates where the command was loaded from, to make debugging easier.
	Source string `yaml:",omitempty"`
}

// Steal the builder API from https://github.com/bbkane/warg

type CommandDescriptionOption func(*CommandDescription)

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

func WithType(t string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Type = t
	}
}

func WithTags(tags ...string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Tags = tags
	}
}

func WithMetadata(metadata map[string]interface{}) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Metadata = metadata
	}
}

func WithLayersList(ls ...layers.ParameterLayer) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for _, l := range ls {
			c.Layers.Set(l.GetSlug(), l)
		}
	}
}

func WithLayers(ls *layers.ParameterLayers) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layers.Merge(ls)
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
			layer, err = layers.NewParameterLayer(layers.DefaultSlug, "Flags")
			if err != nil {
				panic(err)
			}
			c.Layers.Set(layer.GetSlug(), layer)
			err = c.Layers.MoveToFront(layer.GetSlug())
			if err != nil {
				panic(err)
			}
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
			layer, err = layers.NewParameterLayer(layers.DefaultSlug, "Arguments")
			if err != nil {
				panic(err)
			}
			c.Layers.Set(layer.GetSlug(), layer)
			err = c.Layers.MoveToFront(layer.GetSlug())
			if err != nil {
				panic(err)
			}
		}

		for _, arg := range arguments {
			arg.IsArgument = true
		}
		layer.AddFlags(arguments...)
	}
}

func WithLayout(l *layout.Layout) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layout = l.Sections
	}
}

func WithReplaceLayers(layers_ ...layers.ParameterLayer) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for _, l := range layers_ {
			c.Layers.Set(l.GetSlug(), l)
		}
	}
}

func WithParents(p ...string) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Parents = p
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

func NewCommandDescription(name string, options ...CommandDescriptionOption) *CommandDescription {
	ret := &CommandDescription{
		Name:   name,
		Layers: layers.NewParameterLayers(),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

func (cd *CommandDescription) FullPath() string {
	if len(cd.Parents) == 0 {
		return cd.Name
	}
	return strings.Join(cd.Parents, "/") + "/" + cd.Name
}

func (cd *CommandDescription) GetDefaultLayer() (layers.ParameterLayer, bool) {
	return cd.GetLayer(layers.DefaultSlug)
}

func (cd *CommandDescription) GetDefaultFlags() *parameters.ParameterDefinitions {
	l, ok := cd.GetDefaultLayer()
	if !ok {
		return parameters.NewParameterDefinitions()
	}
	return l.GetParameterDefinitions().GetFlags()
}

func (cd *CommandDescription) GetDefaultArguments() *parameters.ParameterDefinitions {
	l, ok := cd.GetDefaultLayer()
	if !ok {
		return parameters.NewParameterDefinitions()
	}

	return l.GetParameterDefinitions().GetArguments()
}

func (cd *CommandDescription) GetLayer(name string) (layers.ParameterLayer, bool) {
	return cd.Layers.Get(name)
}

func (cd *CommandDescription) Clone(cloneLayers bool, options ...CommandDescriptionOption) *CommandDescription {
	// clone flags
	layers_ := layers.NewParameterLayers()
	if cloneLayers {
		layers_ = cd.Layers.Clone()
	}

	// copy parents
	parents := make([]string, len(cd.Parents))
	copy(parents, cd.Parents)

	ret := &CommandDescription{
		Name:    cd.Name,
		Short:   cd.Short,
		Long:    cd.Long,
		Layers:  layers_,
		Parents: parents,
		Source:  cd.Source,
	}

	for _, o := range options {
		o(ret)
	}

	return ret
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
	Metadata(ctx context.Context, parsedLayers *layers.ParsedLayers) (map[string]interface{}, error)
}

// NOTE(manuel, 2023-03-17) Future types of commands that we could need
// - async emitting command (just strings, for example)
// - async emitting structured log
//   - async emitting of glaze rows (useful in general, and could be done with a special TableOutputFormatter, really)
// - no output (just do it yourself)
// - typed generic output structure (with error)

type BareCommand interface {
	Command
	Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error
}

type WriterCommand interface {
	Command
	RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error
}

type GlazeCommand interface {
	Command
	// RunIntoGlazeProcessor is called to actually execute the command.
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
	RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error
}

type ExitWithoutGlazeError struct{}

func (e *ExitWithoutGlazeError) Error() string {
	return "Exit without glaze"
}
