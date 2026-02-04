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
	Schema         *schema.Schema         `yaml:"schema,omitempty"`
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

func WithSections(sections ...schema.Section) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for _, section := range sections {
			c.Schema.Set(section.GetSlug(), section)
		}
	}
}

func WithSchema(schema_ *schema.Schema) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Schema.Merge(schema_)
	}
}

// WithSectionsMap registers sections using explicit slugs from the provided map.
// The map key is used as the registration slug. If a section's internal slug
// (returned by section.GetSlug()) differs from the map key, this function will try
// to align them when possible so that runtime parsing and lookups are
// consistent:
//   - Prefer cloning the section and overriding the slug on the clone when the
//     clone is a *schema.SectionImpl (common for wrapper types that embed
//     SectionImpl and whose Clone returns a SectionImpl).
//   - Otherwise, the section is registered under the provided key as-is.
//
// Note: If a non-SectionImpl is registered under a key that differs
// from its internal slug, middlewares that derive resolved section slugs from
// the section's GetSlug() may use the internal slug instead of the registration
// key. Prefer using matching slugs or SectionImpl when you need explicit remapping.
func WithSectionsMap(m map[string]schema.Section) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for slug, section := range m {
			if section.GetSlug() != slug {
				// Try a generic clone: many wrapper types embed SectionImpl,
				// whose Clone returns *SectionImpl. If so, set the slug.
				cloned := section.Clone()
				if impl, ok := cloned.(*schema.SectionImpl); ok {
					impl.Slug = slug
					c.Schema.Set(slug, impl)
					log.Debug().Str("slug", slug).Str("internalSlug", section.GetSlug()).Msg("WithSectionsMap: cloned section and set overridden slug")
					continue
				}
				// Fallback: keep original section but register under provided key.
				// Resolved values may still use the internal slug when indexing.
				log.Warn().Str("slug", slug).Str("internalSlug", section.GetSlug()).Msg("WithSectionsMap: registering section with mismatched internal slug; resolved values may use internal slug")
			}
			c.Schema.Set(slug, section)
		}
	}
}

// WithFlags is a convenience function to add arguments to the default section, useful
// to make the transition from explicit flags and arguments to a default section a bit easier.
func WithFlags(
	flags ...*fields.Definition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		section, ok := c.GetDefaultSection()
		var err error
		if !ok {
			section, err = schema.NewSection(schema.DefaultSlug, "Flags")
			if err != nil {
				panic(err)
			}
			c.Schema.Set(section.GetSlug(), section)
			err = c.Schema.MoveToFront(section.GetSlug())
			if err != nil {
				panic(err)
			}
		}
		section.AddFields(flags...)
	}
}

// WithArguments is a convenience function to add arguments to the default section, useful
// to make the transition from explicit flags and arguments to a default section a bit easier.
func WithArguments(
	arguments ...*fields.Definition,
) CommandDescriptionOption {
	return func(c *CommandDescription) {
		section, ok := c.GetDefaultSection()
		var err error
		if !ok {
			section, err = schema.NewSection(schema.DefaultSlug, "Arguments")
			if err != nil {
				panic(err)
			}
			c.Schema.Set(section.GetSlug(), section)
			err = c.Schema.MoveToFront(section.GetSlug())
			if err != nil {
				panic(err)
			}
		}

		for _, arg := range arguments {
			arg.IsArgument = true
		}
		section.AddFields(arguments...)
	}
}

func WithLayout(l *layout.Layout) CommandDescriptionOption {
	return func(c *CommandDescription) {
		c.Layout = l.Sections
	}
}

func WithReplaceSections(sections ...schema.Section) CommandDescriptionOption {
	return func(c *CommandDescription) {
		for _, section := range sections {
			c.Schema.Set(section.GetSlug(), section)
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
		Schema: schema.NewSchema(),
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

func (cd *CommandDescription) GetDefaultSection() (schema.Section, bool) {
	return cd.GetSection(schema.DefaultSlug)
}

func (cd *CommandDescription) GetDefaultFlags() *fields.Definitions {
	section, ok := cd.GetDefaultSection()
	if !ok {
		return fields.NewDefinitions()
	}
	return section.GetDefinitions().GetFlags()
}

func (cd *CommandDescription) GetDefaultArguments() *fields.Definitions {
	section, ok := cd.GetDefaultSection()
	if !ok {
		return fields.NewDefinitions()
	}

	return section.GetDefinitions().GetArguments()
}

// GetDefaultsMap returns a map of field names to their default values
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

func (cd *CommandDescription) GetSection(name string) (schema.Section, bool) {
	return cd.Schema.Get(name)
}

func (cd *CommandDescription) Clone(cloneSchema bool, options ...CommandDescriptionOption) *CommandDescription {
	// clone flags
	schema_ := schema.NewSchema()
	if cloneSchema {
		schema_ = cd.Schema.Clone()
	}

	// copy parents
	parents := make([]string, len(cd.Parents))
	copy(parents, cd.Parents)

	ret := &CommandDescription{
		Name:    cd.Name,
		Short:   cd.Short,
		Long:    cd.Long,
		Schema:  schema_,
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

func (cd *CommandDescription) SetSections(sections ...schema.Section) {
	for _, section := range sections {
		cd.Schema.Set(section.GetSlug(), section)
	}
}

type Command interface {
	Description() *CommandDescription
	ToYAML(w io.Writer) error
}

type CommandWithMetadata interface {
	Command
	Metadata(ctx context.Context, parsedValues *values.Values) (map[string]interface{}, error)
}

// NOTE(manuel, 2023-03-17) Future types of commands that we could need
// - async emitting command (just strings, for example)
// - async emitting structured log
//   - async emitting of glaze rows (useful in general, and could be done with a special TableOutputFormatter, really)
// - no output (just do it yourself)
// - typed generic output structure (with error)

type BareCommand interface {
	Command
	Run(ctx context.Context, parsedValues *values.Values) error
}

type WriterCommand interface {
	Command
	RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error
}

type GlazeCommand interface {
	Command
	// RunIntoGlazeProcessor is called to actually execute the command.
	//
	// NOTE(manuel, 2023-02-27) We can probably simplify this to only take resolved values
	//
	// The ps and GlazeProcessor calls could be replaced by a GlazeCommand specific section,
	// which would allow the GlazeCommand to parse into a specific struct. The GlazeProcessor
	// is just something created by the passed in GlazeSection anyway.
	//
	// When we are just left with building a convenience wrapper for Glaze based commands,
	// instead of forcing it into the upstream interface.
	//
	// https://github.com/go-go-golems/glazed/issues/217
	// https://github.com/go-go-golems/glazed/issues/216
	// See https://github.com/go-go-golems/glazed/issues/173
	RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error
}

type ExitWithoutGlazeError struct{}

func (e *ExitWithoutGlazeError) Error() string {
	return "Exit without glaze"
}
