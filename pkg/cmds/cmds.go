package cmds

import (
	"context"
	"io"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/rs/zerolog/log"
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
	Layout         []*layout.Section      `yaml:"layout,omitempty"`
	Layers         *schema.Schema         `yaml:"layers,omitempty"`
	AdditionalData map[string]interface{} `yaml:"additionalData,omitempty"`
	Type           string                 `yaml:"type,omitempty"`
	Tags           []string               `yaml:"tags,omitempty"`
	Metadata       map[string]interface{} `yaml:"metadata,omitempty"`

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

func WithLayersList(ls ...schema.Section) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for _, l := range ls {
			c.Layers.Set(l.GetSlug(), l)
		}
	}
}

func WithLayers(ls *schema.Schema) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layers.Merge(ls)
	}
}

// WithSchema is an alias for WithLayers that accepts a schema.Schema.
// It provides a more intuitive name when working with the schema package.
func WithSchema(s *schema.Schema) CommandDescriptionOption {
	return WithLayers((*schema.Schema)(s))
}

// WithLayersMap registers layers using explicit slugs from the provided map.
// The map key is used as the registration slug. If a layer's internal slug
// (returned by l.GetSlug()) differs from the map key, this function will try
// to align them when possible so that runtime parsing and lookups are
// consistent:
//   - Prefer cloning the layer and overriding the slug on the clone when the
//     clone is a *schema.SectionImpl (common for wrapper types that embed
//     SectionImpl and whose Clone returns a SectionImpl).
//   - Otherwise, the layer is registered under the provided key as-is.
//
// Note: If a non-SectionImpl is registered under a key that differs
// from its internal slug, middlewares that derive parsed layer slugs from the
// layer's GetSlug() may use the internal slug instead of the registration key.
// Prefer using matching slugs or SectionImpl when you need explicit
// remapping.
func WithLayersMap(m map[string]schema.Section) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for slug, l := range m {
			if l.GetSlug() != slug {
				// Try a generic clone: many wrapper types embed ParameterLayerImpl,
				// whose Clone returns *ParameterLayerImpl. If so, set the slug.
				cloned := l.Clone()
				if impl, ok := cloned.(*schema.SectionImpl); ok {
					impl.Slug = slug
					c.Layers.Set(slug, impl)
					log.Debug().Str("slug", slug).Str("internalSlug", l.GetSlug()).Msg("WithLayersMap: cloned layer and set overridden slug")
					continue
				}
				// Fallback: keep original layer but register under provided key.
				// Parsed layers may still use the internal slug when indexing.
				log.Warn().Str("slug", slug).Str("internalSlug", l.GetSlug()).Msg("WithLayersMap: registering layer with mismatched internal slug; parsed layers may use internal slug")
			}
			c.Layers.Set(slug, l)
		}
	}
}

// WithFlags is a convenience function to add arguments to the default layer, useful
// to make the transition from explicit flags and arguments to a default layer a bit easier.
func WithFlags(
	flags ...*fields.Definition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		layer, ok := c.GetDefaultLayer()
		var err error
		if !ok {
			layer, err = schema.NewSection(schema.DefaultSlug, "Flags")
			if err != nil {
				panic(err)
			}
			c.Layers.Set(layer.GetSlug(), layer)
			err = c.Layers.MoveToFront(layer.GetSlug())
			if err != nil {
				panic(err)
			}
		}
		layer.AddFields(flags...)
	}
}

// WithArguments is a convenience function to add arguments to the default layer, useful
// to make the transition from explicit flags and arguments to a default layer a bit easier.
func WithArguments(
	arguments ...*fields.Definition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		layer, ok := c.GetDefaultLayer()
		var err error
		if !ok {
			layer, err = schema.NewSection(schema.DefaultSlug, "Arguments")
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
		layer.AddFields(arguments...)
	}
}

func WithLayout(l *layout.Layout) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layout = l.Sections
	}
}

func WithReplaceLayers(layers_ ...schema.Section) CommandDescriptionOption {
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
		Layers: schema.NewSchema(),
	}

	for _, o := range options {
		o(ret)
	}

	return ret
}

// NewCommandDefinition creates a new command definition with the given name and options.
// It is an alias for NewCommandDescription that provides clearer vocabulary.
func NewCommandDefinition(name string, options ...CommandDescriptionOption) *CommandDescription {
	return NewCommandDescription(name, options...)
}

func (cd *CommandDescription) FullPath() string {
	if len(cd.Parents) == 0 {
		return cd.Name
	}
	return strings.Join(cd.Parents, "/") + "/" + cd.Name
}

func (cd *CommandDescription) GetDefaultLayer() (schema.Section, bool) {
	return cd.GetLayer(schema.DefaultSlug)
}

func (cd *CommandDescription) GetDefaultFlags() *fields.Definitions {
	l, ok := cd.GetDefaultLayer()
	if !ok {
		return fields.NewDefinitions()
	}
	return l.GetDefinitions().GetFlags()
}

func (cd *CommandDescription) GetDefaultArguments() *fields.Definitions {
	l, ok := cd.GetDefaultLayer()
	if !ok {
		return fields.NewDefinitions()
	}

	return l.GetDefinitions().GetArguments()
}

// GetDefaultsMap returns a map of parameter names to their default values
// by combining the default flags and arguments.
func (cd *CommandDescription) GetDefaultsMap() (map[string]interface{}, error) {
	flags := cd.GetDefaultFlags()
	arguments := cd.GetDefaultArguments()

	params, err := flags.FieldValuesFromDefaults()
	if err != nil {
		return nil, err
	}

	argsParams, err := arguments.FieldValuesFromDefaults()
	if err != nil {
		return nil, err
	}

	paramsMap := params.ToMap()
	for k, v := range argsParams.ToMap() {
		paramsMap[k] = v
	}

	return paramsMap, nil
}

func (cd *CommandDescription) GetLayer(name string) (schema.Section, bool) {
	return cd.Layers.Get(name)
}

func (cd *CommandDescription) Clone(cloneLayers bool, options ...CommandDescriptionOption) *CommandDescription {
	// clone flags
	layers_ := schema.NewSchema()
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

func (cd *CommandDescription) SetLayers(layers ...schema.Section) {
	for _, l := range layers {
		cd.Layers.Set(l.GetSlug(), l)
	}
}

type Command interface {
	Description() *CommandDescription
	ToYAML(w io.Writer) error
}

type CommandWithMetadata interface {
	Command
	Metadata(ctx context.Context, parsedLayers *values.Values) (map[string]interface{}, error)
}

// NOTE(manuel, 2023-03-17) Future types of commands that we could need
// - async emitting command (just strings, for example)
// - async emitting structured log
//   - async emitting of glaze rows (useful in general, and could be done with a special TableOutputFormatter, really)
// - no output (just do it yourself)
// - typed generic output structure (with error)

type BareCommand interface {
	Command
	Run(ctx context.Context, parsedLayers *values.Values) error
}

type WriterCommand interface {
	Command
	RunIntoWriter(ctx context.Context, parsedLayers *values.Values, w io.Writer) error
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
	RunIntoGlazeProcessor(ctx context.Context, parsedLayers *values.Values, gp middlewares.Processor) error
}

type ExitWithoutGlazeError struct{}

func (e *ExitWithoutGlazeError) Error() string {
	return "Exit without glaze"
}
