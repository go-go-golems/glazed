package cmds

import (
	"context"
	"io"
	"text/template"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type TemplateCommand struct {
	*CommandDescription
	Template string
}

type TemplateCommandDescription struct {
	Name      string               `yaml:"name"`
	Short     string               `yaml:"short"`
	Long      string               `yaml:"long,omitempty"`
	Layout    []*layout.Section    `yaml:"layout,omitempty"`
	Flags     []*fields.Definition `yaml:"flags,omitempty"`
	Arguments []*fields.Definition `yaml:"arguments,omitempty"`
	Schema    schema.Schema        `yaml:"schema,omitempty"`
	Template  string               `yaml:"template"`
}

func NewTemplateCommand(name string, template string, options ...CommandDescriptionOption) *TemplateCommand {
	tc := &TemplateCommand{
		CommandDescription: NewCommandDescription(name),
		Template:           template,
	}

	for _, option := range options {
		option(tc.Description())
	}

	return tc
}

var _ WriterCommand = (*TemplateCommand)(nil)

func (t *TemplateCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	tmpl, err := template.New("template").Parse(t.Template)
	if err != nil {
		log.Warn().Err(err).Str("template", t.Template).Msg("failed to parse template")
		return errors.Wrap(err, "failed to parse template")
	}

	err = tmpl.Execute(w, parsedValues.GetDataMap())
	if err != nil {
		return errors.Wrap(err, "failed to execute template")
	}

	return nil
}

func (t *TemplateCommand) IsValid() bool {
	if t.Template == "" {
		return false
	}

	if t.CommandDescription == nil {
		return false
	}

	if t.Name == "" || t.Short == "" {
		return false
	}

	return true
}

type TemplateCommandLoader struct{}

func (tcl *TemplateCommandLoader) LoadCommandFromYAML(
	s io.Reader,
	options ...CommandDescriptionOption,
) ([]Command, error) {
	tcd := &TemplateCommandDescription{}
	err := yaml.NewDecoder(s).Decode(tcd)
	if err != nil {
		return nil, err
	}

	for _, argument := range tcd.Arguments {
		argument.IsArgument = true
	}

	defaultSection, err := schema.NewSection(schema.DefaultSlug, "Default",
		schema.WithFields(append(tcd.Flags, tcd.Arguments...)...))
	if err != nil {
		return nil, err
	}

	options_ := []CommandDescriptionOption{
		WithShort(tcd.Short),
		WithLong(tcd.Long),
		WithSections(tcd.Schema.AsList()...),
		WithSections(defaultSection),
		WithLayout(&layout.Layout{
			Sections: tcd.Layout,
		}),
	}
	options_ = append(options_, options...)

	tc := &TemplateCommand{
		CommandDescription: NewCommandDescription(tcd.Name),
		Template:           tcd.Template,
	}

	for _, option := range options_ {
		option(tc.Description())
	}

	if !tc.IsValid() {
		return nil, errors.New("Invalid command")
	}

	return []Command{tc}, nil
}

func (t *TemplateCommand) ToYAML(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	defer func(enc *yaml.Encoder) {
		_ = enc.Close()
	}(enc)

	return enc.Encode(t)
}
