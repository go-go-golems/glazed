package settings

import (
	_ "embed"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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

type GlazedParameterLayers struct {
	FieldsFiltersParameterLayer *FieldsFiltersParameterLayer `yaml:"fieldsFiltersParameterLayer"`
	OutputParameterLayer        *OutputParameterLayer        `yaml:"outputParameterLayer"`
	RenameParameterLayer        *RenameParameterLayer        `yaml:"renameParameterLayer"`
	ReplaceParameterLayer       *ReplaceParameterLayer       `yaml:"replaceParameterLayer"`
	SelectParameterLayer        *SelectParameterLayer        `yaml:"selectParameterLayer"`
	TemplateParameterLayer      *TemplateParameterLayer      `yaml:"templateParameterLayer"`
	JqParameterLayer            *JqParameterLayer            `yaml:"jqParameterLayer"`
	SortParameterLayer          *SortParameterLayer          `yaml:"sortParameterLayer"`
	SkipLimitParameterLayer     *SkipLimitParameterLayer     `yaml:"skipLimitParameterLayer"`
}

const GlazedSlug = "glazed"

var _ schema.Section = (*GlazedParameterLayers)(nil)
var _ layers.CobraParameterLayer = (*GlazedParameterLayers)(nil)

// NewGlazedSchema creates a new glazed schema section containing all glazed output/formatting settings.
// It wraps NewGlazedParameterLayers for the schema facade package.
func NewGlazedSchema(options ...GlazeParameterLayerOption) (schema.Section, error) {
	return NewGlazedParameterLayers(options...)
}

func (g *GlazedParameterLayers) Clone() schema.Section {
	return &GlazedParameterLayers{
		FieldsFiltersParameterLayer: g.FieldsFiltersParameterLayer.Clone().(*FieldsFiltersParameterLayer),
		OutputParameterLayer:        g.OutputParameterLayer.Clone().(*OutputParameterLayer),
		RenameParameterLayer:        g.RenameParameterLayer.Clone().(*RenameParameterLayer),
		ReplaceParameterLayer:       g.ReplaceParameterLayer.Clone().(*ReplaceParameterLayer),
		SelectParameterLayer:        g.SelectParameterLayer.Clone().(*SelectParameterLayer),
		TemplateParameterLayer:      g.TemplateParameterLayer.Clone().(*TemplateParameterLayer),
		JqParameterLayer:            g.JqParameterLayer.Clone().(*JqParameterLayer),
		SortParameterLayer:          g.SortParameterLayer.Clone().(*SortParameterLayer),
		SkipLimitParameterLayer:     g.SkipLimitParameterLayer.Clone().(*SkipLimitParameterLayer),
	}
}

func (g *GlazedParameterLayers) MarshalYAML() (interface{}, error) {
	return &schema.SectionImpl{
		Name:        g.GetName(),
		Slug:        g.GetSlug(),
		Description: g.GetDescription(),
		Prefix:      g.GetPrefix(),
		ChildLayers: []schema.Section{
			g.FieldsFiltersParameterLayer,
			g.OutputParameterLayer,
			g.RenameParameterLayer,
			g.ReplaceParameterLayer,
			g.SelectParameterLayer,
			g.TemplateParameterLayer,
			g.JqParameterLayer,
			g.SortParameterLayer,
		},
	}, nil
}

func (g *GlazedParameterLayers) GetName() string {
	return "Glazed Flags"
}

func (g *GlazedParameterLayers) GetSlug() string {
	return GlazedSlug
}

func (g *GlazedParameterLayers) GetDescription() string {
	return "Glazed flags"
}

func (g *GlazedParameterLayers) GetPrefix() string {
	return g.FieldsFiltersParameterLayer.GetPrefix()
}

func (g *GlazedParameterLayers) AddFlags(...*fields.Definition) {
	panic("not supported me")
}

func (g *GlazedParameterLayers) GetParameterDefinitions() *fields.Definitions {
	ret := fields.NewDefinitions()
	ret.Merge(g.OutputParameterLayer.GetParameterDefinitions()).
		Merge(g.FieldsFiltersParameterLayer.GetParameterDefinitions()).
		Merge(g.SelectParameterLayer.GetParameterDefinitions()).
		Merge(g.TemplateParameterLayer.GetParameterDefinitions()).
		Merge(g.RenameParameterLayer.GetParameterDefinitions()).
		Merge(g.ReplaceParameterLayer.GetParameterDefinitions()).
		Merge(g.JqParameterLayer.GetParameterDefinitions()).
		Merge(g.SortParameterLayer.GetParameterDefinitions()).
		Merge(g.SkipLimitParameterLayer.GetParameterDefinitions())

	return ret
}

func (g *GlazedParameterLayers) AddLayerToCobraCommand(cmd *cobra.Command) error {
	layers := []layers.CobraParameterLayer{
		g.OutputParameterLayer,
		g.FieldsFiltersParameterLayer,
		g.SelectParameterLayer,
		g.TemplateParameterLayer,
		g.RenameParameterLayer,
		g.ReplaceParameterLayer,
		g.JqParameterLayer,
		g.SortParameterLayer,
		g.SkipLimitParameterLayer,
	}

	for _, layer := range layers {
		if err := layer.AddLayerToCobraCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (g *GlazedParameterLayers) ParseLayerFromCobraCommand(
	cmd *cobra.Command,
	options ...parameters.ParseStepOption,
) (*values.SectionValues, error) {
	res := &values.SectionValues{
		Layer: g,
	}
	ps := parameters.NewParsedParameters()

	layers := []layers.CobraParameterLayer{
		g.OutputParameterLayer,
		g.SelectParameterLayer,
		g.RenameParameterLayer,
		g.TemplateParameterLayer,
		g.FieldsFiltersParameterLayer,
		g.ReplaceParameterLayer,
		g.JqParameterLayer,
		g.SortParameterLayer,
		g.SkipLimitParameterLayer,
	}

	for _, layer := range layers {
		l, err := layer.ParseLayerFromCobraCommand(cmd, options...)
		if err != nil {
			return nil, err
		}
		if _, err = ps.Merge(l.Parameters); err != nil {
			return nil, err
		}
	}

	res.Parameters = ps
	return res, nil
}

func (g *GlazedParameterLayers) GatherParametersFromMap(
	m map[string]interface{}, onlyProvided bool,
	options ...parameters.ParseStepOption,
) (*parameters.ParsedParameters, error) {
	ps := parameters.NewParsedParameters()

	layers := []schema.Section{
		g.OutputParameterLayer,
		g.SelectParameterLayer,
		g.RenameParameterLayer,
		g.TemplateParameterLayer,
		g.FieldsFiltersParameterLayer,
		g.ReplaceParameterLayer,
		g.JqParameterLayer,
		g.SortParameterLayer,
		g.SkipLimitParameterLayer,
	}

	for _, layer := range layers {
		ps_, err := layer.GetParameterDefinitions().GatherParametersFromMap(m, onlyProvided, options...)
		if err != nil {
			return nil, err
		}
		if _, err = ps.Merge(ps_); err != nil {
			return nil, err
		}
	}

	return ps, nil
}

func (g *GlazedParameterLayers) InitializeParameterDefaultsFromStruct(s interface{}) error {
	layers := []schema.Section{
		g.OutputParameterLayer,
		g.FieldsFiltersParameterLayer,
		g.SelectParameterLayer,
		g.TemplateParameterLayer,
		g.RenameParameterLayer,
		g.ReplaceParameterLayer,
		g.JqParameterLayer,
		g.SortParameterLayer,
		g.SkipLimitParameterLayer,
	}

	for _, layer := range layers {
		if err := layer.InitializeParameterDefaultsFromStruct(s); err != nil {
			return err
		}
	}
	return nil
}

type GlazeParameterLayerOption func(*GlazedParameterLayers) error

func WithOutputParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.OutputParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSelectParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SelectParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithTemplateParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.TemplateParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithRenameParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.RenameParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithReplaceParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.ReplaceParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithFieldsFiltersParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.FieldsFiltersParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithJqParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.JqParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSortParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SortParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSkipLimitParameterLayerOptions(options ...schema.SectionOption) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SkipLimitParameterLayer.SectionImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func NewGlazedParameterLayers(options ...GlazeParameterLayerOption) (*GlazedParameterLayers, error) {
	fieldsFiltersParameterLayer, err := NewFieldsFiltersParameterLayer()
	if err != nil {
		return nil, err
	}
	outputParameterLayer, err := NewOutputParameterLayer()
	if err != nil {
		return nil, err
	}
	renameParameterLayer, err := NewRenameParameterLayer()
	if err != nil {
		return nil, err
	}
	replaceParameterLayer, err := NewReplaceParameterLayer()
	if err != nil {
		return nil, err
	}
	selectParameterLayer, err := NewSelectParameterLayer()
	if err != nil {
		return nil, err
	}
	templateParameterLayer, err := NewTemplateParameterLayer()
	if err != nil {
		return nil, err
	}
	jqParameterLayer, err := NewJqParameterLayer()
	if err != nil {
		return nil, err
	}
	sortParameterLayer, err := NewSortParameterLayer()
	if err != nil {
		return nil, err
	}
	skipLimitParameterLayer, err := NewSkipLimitParameterLayer()
	if err != nil {
		return nil, err
	}
	ret := &GlazedParameterLayers{
		FieldsFiltersParameterLayer: fieldsFiltersParameterLayer,
		OutputParameterLayer:        outputParameterLayer,
		RenameParameterLayer:        renameParameterLayer,
		ReplaceParameterLayer:       replaceParameterLayer,
		SelectParameterLayer:        selectParameterLayer,
		TemplateParameterLayer:      templateParameterLayer,
		JqParameterLayer:            jqParameterLayer,
		SortParameterLayer:          sortParameterLayer,
		SkipLimitParameterLayer:     skipLimitParameterLayer,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func SetupRowOutputFormatter(glazedLayer *values.SectionValues) (formatters.RowOutputFormatter, error) {
	outputSettings, err := NewOutputFormatterSettings(glazedLayer)
	if err != nil {
		return nil, err
	}

	of, err := outputSettings.CreateRowOutputFormatter()
	if err != nil {
		return nil, err
	}

	return of, nil
}

func SetupTableOutputFormatter(glazedLayer *values.SectionValues) (formatters.TableOutputFormatter, error) {
	selectSettings, err := NewSelectSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}

	outputSettings, err := NewOutputFormatterSettings(glazedLayer)
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
// DO(manuel, 2023-06-30) It would be good to used a parsedLayer here, if we ever refactor that part
func SetupTableProcessor(
	glazedLayer *values.SectionValues,
	options ...middlewares.TableProcessorOption,
) (*middlewares.TableProcessor, error) {
	// TODO(manuel, 2023-03-06): This is where we should check that flags that are mutually incompatible don't clash
	//
	// See: https://github.com/go-go-golems/glazed/issues/199
	templateSettings, err := NewTemplateSettings(glazedLayer)
	if err != nil {
		return nil, err
	}
	selectSettings, err := NewSelectSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}
	renameSettings, err := NewRenameSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}
	fieldsFilterSettings, err := NewFieldsFilterSettings(glazedLayer)
	if err != nil {
		return nil, err
	}
	replaceSettings, err := NewReplaceSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}
	jqSettings, err := NewJqSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}
	sortSettings, err := NewSortSettingsFromParameters(glazedLayer)
	if err != nil {
		return nil, err
	}
	outputSettings, err := NewOutputFormatterSettings(glazedLayer)
	if err != nil {
		return nil, err
	}
	skipLimitSettings, err := NewSkipLimitSettingsFromParameters(glazedLayer)
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
	// to the API that we currently use (which is a unordered hashmap, and parsed layers that lose the positioning)
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
	glazedLayer *values.SectionValues,
	w io.Writer,
) (formatters.OutputFormatter, error) {
	// first, try to get a row updater
	rowOf, err := SetupRowOutputFormatter(glazedLayer)

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

		of, err := SetupTableOutputFormatter(glazedLayer)
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
