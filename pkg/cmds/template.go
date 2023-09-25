package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/layout"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"gopkg.in/yaml.v3"
	"io"
	"text/template"

	"github.com/pkg/errors"
)

type TemplateCommand struct {
	*CommandDescription
	Template string
}

type TemplateCommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Layout    []*layout.Section                 `yaml:"layout,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	Layers    []layers.ParameterLayer           `yaml:"layers,omitempty"`
	Template  string                            `yaml:"template"`
}

func (t *TemplateCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	w io.Writer,
) error {
	tmpl, err := template.New("template").Parse(t.Template)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}

	err = tmpl.Execute(w, ps)
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

	if t.CommandDescription.Name == "" || t.CommandDescription.Short == "" {
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

	options_ := []CommandDescriptionOption{
		WithShort(tcd.Short),
		WithLong(tcd.Long),
		WithFlags(tcd.Flags...),
		WithArguments(tcd.Arguments...),
		WithLayers(tcd.Layers...),
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
