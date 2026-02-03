package settings

import (
	_ "embed"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/formatters/simple"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/middlewares/object"
	"github.com/go-go-golems/glazed/pkg/middlewares/row"
	"github.com/go-go-golems/glazed/pkg/middlewares/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Helpers for cobra commands

type GlazedSection struct {
	FieldsFiltersSection *FieldsFiltersSection `yaml:"fieldsFiltersSection"`
	OutputSection        *OutputSection        `yaml:"outputSection"`
	RenameSection        *RenameSection        `yaml:"renameSection"`
	ReplaceSection       *ReplaceSection       `yaml:"replaceSection"`
	SelectSection        *SelectSection        `yaml:"selectSection"`
	TemplateSection      *TemplateSection      `yaml:"templateSection"`
	JqSection            *JqSection            `yaml:"jqSection"`
	SortSection          *SortSection          `yaml:"sortSection"`
	SkipLimitSection     *SkipLimitSection     `yaml:"skipLimitSection"`
}

const GlazedSlug = "glazed"

var _ schema.Section = (*GlazedSection)(nil)
var _ schema.CobraSection = (*GlazedSection)(nil)

// NewGlazedSchema creates a new glazed schema section containing all glazed output/formatting settings.
// It wraps NewGlazedSection for the schema facade package.
func NewGlazedSchema(options ...GlazeSectionOption) (schema.Section, error) {
	return NewGlazedSection(options...)
}

func (g *GlazedSection) Clone() schema.Section {
	return &GlazedSection{
		FieldsFiltersSection: g.FieldsFiltersSection.Clone().(*FieldsFiltersSection),
		OutputSection:        g.OutputSection.Clone().(*OutputSection),
		RenameSection:        g.RenameSection.Clone().(*RenameSection),
		ReplaceSection:       g.ReplaceSection.Clone().(*ReplaceSection),
		SelectSection:        g.SelectSection.Clone().(*SelectSection),
		TemplateSection:      g.TemplateSection.Clone().(*TemplateSection),
		JqSection:            g.JqSection.Clone().(*JqSection),
		SortSection:          g.SortSection.Clone().(*SortSection),
		SkipLimitSection:     g.SkipLimitSection.Clone().(*SkipLimitSection),
	}
}

func (g *GlazedSection) MarshalYAML() (interface{}, error) {
	return &schema.SectionImpl{
		Name:        g.GetName(),
		Slug:        g.GetSlug(),
		Description: g.GetDescription(),
		Prefix:      g.GetPrefix(),
		ChildSections: []schema.Section{
			g.FieldsFiltersSection,
			g.OutputSection,
			g.RenameSection,
			g.ReplaceSection,
			g.SelectSection,
			g.TemplateSection,
			g.JqSection,
			g.SortSection,
		},
	}, nil
}

func (g *GlazedSection) GetName() string {
	return "Glazed Flags"
}

func (g *GlazedSection) GetSlug() string {
	return GlazedSlug
}

func (g *GlazedSection) GetDescription() string {
	return "Glazed flags"
}

func (g *GlazedSection) GetPrefix() string {
	return g.FieldsFiltersSection.GetPrefix()
}

func (g *GlazedSection) AddFields(...*fields.Definition) {
	panic("not supported me")
}

func (g *GlazedSection) GetDefinitions() *fields.Definitions {
	ret := fields.NewDefinitions()
	ret.Merge(g.OutputSection.GetDefinitions()).
		Merge(g.FieldsFiltersSection.GetDefinitions()).
		Merge(g.SelectSection.GetDefinitions()).
		Merge(g.TemplateSection.GetDefinitions()).
		Merge(g.RenameSection.GetDefinitions()).
		Merge(g.ReplaceSection.GetDefinitions()).
		Merge(g.JqSection.GetDefinitions()).
		Merge(g.SortSection.GetDefinitions()).
		Merge(g.SkipLimitSection.GetDefinitions())

	return ret
}

func (g *GlazedSection) AddSectionToCobraCommand(cmd *cobra.Command) error {
	sections := []schema.CobraSection{
		g.OutputSection,
		g.FieldsFiltersSection,
		g.SelectSection,
		g.TemplateSection,
		g.RenameSection,
		g.ReplaceSection,
		g.JqSection,
		g.SortSection,
		g.SkipLimitSection,
	}

	for _, section := range sections {
		if err := section.AddSectionToCobraCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (g *GlazedSection) ParseSectionFromCobraCommand(
	cmd *cobra.Command,
	options ...fields.ParseOption,
) (*values.SectionValues, error) {
	res := &values.SectionValues{
		Section: g,
	}
	ps := fields.NewFieldValues()

	sections := []schema.CobraSection{
		g.OutputSection,
		g.SelectSection,
		g.RenameSection,
		g.TemplateSection,
		g.FieldsFiltersSection,
		g.ReplaceSection,
		g.JqSection,
		g.SortSection,
		g.SkipLimitSection,
	}

	for _, section := range sections {
		l, err := section.ParseSectionFromCobraCommand(cmd, options...)
		if err != nil {
			return nil, err
		}
		if _, err = ps.Merge(l.Fields); err != nil {
			return nil, err
		}
	}

	res.Fields = ps
	return res, nil
}

func (g *GlazedSection) GatherFieldsFromMap(
	m map[string]interface{}, onlyProvided bool,
	options ...fields.ParseOption,
) (*fields.FieldValues, error) {
	ps := fields.NewFieldValues()

	sections := []schema.Section{
		g.OutputSection,
		g.SelectSection,
		g.RenameSection,
		g.TemplateSection,
		g.FieldsFiltersSection,
		g.ReplaceSection,
		g.JqSection,
		g.SortSection,
		g.SkipLimitSection,
	}

	for _, section := range sections {
		ps_, err := section.GetDefinitions().GatherFieldsFromMap(m, onlyProvided, options...)
		if err != nil {
			return nil, err
		}
		if _, err = ps.Merge(ps_); err != nil {
			return nil, err
		}
	}

	return ps, nil
}

func (g *GlazedSection) InitializeDefaultsFromStruct(s interface{}) error {
	sections := []schema.Section{
		g.OutputSection,
		g.FieldsFiltersSection,
		g.SelectSection,
		g.TemplateSection,
		g.RenameSection,
		g.ReplaceSection,
		g.JqSection,
		g.SortSection,
		g.SkipLimitSection,
	}

	for _, section := range sections {
		if err := section.InitializeDefaultsFromStruct(s); err != nil {
			return err
		}
	}
	return nil
}

type GlazeSectionOption func(*GlazedSection) error

func WithOutputSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.OutputSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSelectSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.SelectSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithTemplateSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.TemplateSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithRenameSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.RenameSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithReplaceSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.ReplaceSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithFieldsFiltersSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.FieldsFiltersSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithJqSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.JqSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSortSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.SortSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSkipLimitSectionOptions(options ...schema.SectionOption) GlazeSectionOption {
	return func(g *GlazedSection) error {
		for _, option := range options {
			err := option(g.SkipLimitSection.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func NewGlazedSection(options ...GlazeSectionOption) (*GlazedSection, error) {
	fieldsFiltersSection, err := NewFieldsFiltersSection()
	if err != nil {
		return nil, err
	}
	outputSection, err := NewOutputSection()
	if err != nil {
		return nil, err
	}
	renameSection, err := NewRenameSection()
	if err != nil {
		return nil, err
	}
	replaceSection, err := NewReplaceSection()
	if err != nil {
		return nil, err
	}
	selectSection, err := NewSelectSection()
	if err != nil {
		return nil, err
	}
	templateSection, err := NewTemplateSection()
	if err != nil {
		return nil, err
	}
	jqSection, err := NewJqSection()
	if err != nil {
		return nil, err
	}
	sortSection, err := NewSortSection()
	if err != nil {
		return nil, err
	}
	skipLimitSection, err := NewSkipLimitSection()
	if err != nil {
		return nil, err
	}
	ret := &GlazedSection{
		FieldsFiltersSection: fieldsFiltersSection,
		OutputSection:        outputSection,
		RenameSection:        renameSection,
		ReplaceSection:       replaceSection,
		SelectSection:        selectSection,
		TemplateSection:      templateSection,
		JqSection:            jqSection,
		SortSection:          sortSection,
		SkipLimitSection:     skipLimitSection,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func SetupRowOutputFormatter(glazedValues *values.SectionValues) (formatters.RowOutputFormatter, error) {
	outputSettings, err := NewOutputFormatterSettings(glazedValues)
	if err != nil {
		return nil, err
	}

	of, err := outputSettings.CreateRowOutputFormatter()
	if err != nil {
		return nil, err
	}

	return of, nil
}

func SetupTableOutputFormatter(glazedValues *values.SectionValues) (formatters.TableOutputFormatter, error) {
	selectSettings, err := NewSelectSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}

	outputSettings, err := NewOutputFormatterSettings(glazedValues)
	if err != nil {
		return nil, err
	}

	var of formatters.TableOutputFormatter
	if selectSettings.SelectField != "" {
		of = simple.NewSingleColumnFormatter(
			selectSettings.SelectField,
			simple.WithSeparator(selectSettings.SelectSeparator),
			simple.WithOutputFile(outputSettings.OutputFile),
			simple.WithOutputMultipleFiles(outputSettings.OutputMultipleFiles),
			simple.WithOutputFileTemplate(outputSettings.OutputFileTemplate),
		)
	} else {
		of, err = outputSettings.CreateTableOutputFormatter()
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating output formatter")
		}
	}
	return of, nil

}

func SetupSimpleTableProcessor(
	output string,
	tableFormat string,
	options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, error) {
	gp := middlewares.NewTableProcessor(options...)

	return gp, nil
}

// SetupTableProcessor processes all the glazed flags out of ps and returns a TableProcessor
// configured with all the necessary middlewares except for the output formatter.
//
// DO(manuel, 2023-06-30) It would be good to use parsed values here, if we ever refactor that part
func SetupTableProcessor(
	glazedValues *values.SectionValues,
	options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, error) {
	// TODO(manuel, 2023-03-06): This is where we should check that flags that are mutually incompatible don't clash
	//
	// See: https://github.com/go-go-golems/glazed/issues/199
	templateSettings, err := NewTemplateSettings(glazedValues)
	if err != nil {
		return nil, err
	}
	selectSettings, err := NewSelectSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}
	renameSettings, err := NewRenameSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}
	fieldsFilterSettings, err := NewFieldsFilterSettings(glazedValues)
	if err != nil {
		return nil, err
	}
	replaceSettings, err := NewReplaceSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}
	jqSettings, err := NewJqSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}
	sortSettings, err := NewSortSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}
	outputSettings, err := NewOutputFormatterSettings(glazedValues)
	if err != nil {
		return nil, err
	}
	skipLimitSettings, err := NewSkipLimitSettingsFromValues(glazedValues)
	if err != nil {
		return nil, err
	}

	templateSettings.UpdateWithSelectSettings(selectSettings)

	gp := middlewares.NewTableProcessor(options...)

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := row.NewFlattenObjectMiddleware()
		gp.AddRowMiddlewareInFront(mw)
	}
	fieldsFilterSettings.AddMiddlewares(gp)

	err = replaceSettings.AddMiddlewares(gp)
	if err != nil {
		return nil, errors.Wrapf(err, "Error adding replace middlewares")
	}

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := object.NewTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	jqObjectMiddleware, jqTableMiddleware, err := NewJqMiddlewaresFromSettings(jqSettings)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not create jq middlewares")
	}

	if jqObjectMiddleware != nil {
		middlewares_ = append(middlewares_, jqObjectMiddleware)
	}

	if jqTableMiddleware != nil {
		gp.AddTableMiddleware(jqTableMiddleware)
	}

	// NOTE(manuel, 2023-03-20): We need to figure out how to order middlewares, on the command line.
	// This is not possible with cobra, which doesn't have ordering of flags, and adding that
	// to the API that we currently use (which is a unordered hashmap, and parsed values that lose the positioning)
	// is not trivial.
	sortSettings.AddMiddlewares(gp)

	if skipLimitSettings.Skip != 0 || skipLimitSettings.Limit != 0 {
		gp.AddRowMiddleware(&row.SkipLimitMiddleware{
			Skip:  skipLimitSettings.Skip,
			Limit: skipLimitSettings.Limit,
		})
	}

	gp.AddObjectMiddleware(middlewares_...)

	return gp, nil
}

// SetupProcessorOutput creates a new Output middleware (either row or table, depending on the format
// and the stream flag set in ps) and adds it to the TableProcessor. Additional middlewares required by ]
// the chosen output format might be added as well (for example, flattening rows when using table-oriented
// output formats).
//
// It also returns the output formatter that was created.
func SetupProcessorOutput(
	gp *middlewares.TableProcessor,
	glazedValues *values.SectionValues,
	w io.Writer,
) (formatters.OutputFormatter, error) {
	// first, try to get a row updater
	rowOf, err := SetupRowOutputFormatter(glazedValues)

	if rowOf != nil {
		err = rowOf.RegisterRowMiddlewares(gp)
		if err != nil {
			return nil, err
		}
		gp.AddRowMiddleware(row.NewOutputMiddleware(rowOf, w))
		return rowOf, nil
	} else {
		if _, ok := err.(*ErrorRowFormatUnsupported); !ok {
			return nil, err
		}

		of, err := SetupTableOutputFormatter(glazedValues)
		if err != nil {
			return nil, err
		}
		err = of.RegisterTableMiddlewares(gp)
		if err != nil {
			return nil, err
		}

		gp.AddTableMiddleware(table.NewOutputMiddleware(of, w))

		return of, nil
	}
}
