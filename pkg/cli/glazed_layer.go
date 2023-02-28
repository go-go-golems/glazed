package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Helpers for cobra commands

type GlazedParameterLayers struct {
	FieldsFiltersParameterLayer *FieldsFiltersParameterLayer
	OutputParameterLayer        *OutputParameterLayer
	RenameParameterLayer        *RenameParameterLayer
	ReplaceParameterLayer       *ReplaceParameterLayer
	SelectParameterLayer        *SelectParameterLayer
	TemplateParameterLayer      *TemplateParameterLayer
}

func (g *GlazedParameterLayers) GetName() string {
	return "Glazed Flags"
}

func (g *GlazedParameterLayers) GetSlug() string {
	return "glazed"
}

func (g *GlazedParameterLayers) GetDescription() string {
	return "Glazed flags"
}

func (g *GlazedParameterLayers) GetPrefix() string {
	return g.FieldsFiltersParameterLayer.GetPrefix()
}

func (g *GlazedParameterLayers) AddFlag(*parameters.ParameterDefinition) {
	panic("not supported me")
}

func (g *GlazedParameterLayers) GetParameterDefinitions() map[string]*parameters.ParameterDefinition {
	ret := make(map[string]*parameters.ParameterDefinition)
	for k, v := range g.RenameParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.OutputParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.SelectParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.TemplateParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.FieldsFiltersParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	for k, v := range g.ReplaceParameterLayer.GetParameterDefinitions() {
		ret[k] = v
	}

	return ret
}

func (g *GlazedParameterLayers) AddFlagsToCobraCommand(cmd *cobra.Command) error {
	err := g.OutputParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.SelectParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.AddFlagsToCobraCommand(cmd)
	if err != nil {
		return err
	}

	return nil
}

func (g *GlazedParameterLayers) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	ps, err := g.OutputParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	ps_, err := g.SelectParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.RenameParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.TemplateParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.FieldsFiltersParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.ReplaceParameterLayer.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	return ps, nil
}

func (g *GlazedParameterLayers) InitializeParameterDefaultsFromStruct(s interface{}) error {
	err := g.OutputParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}

	err = g.SelectParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.InitializeParameterDefaultsFromStruct(s)
	if err != nil {
		return err
	}
	return nil
}

type GlazeParameterLayerOption func(*GlazedParameterLayers) error

func WithOutputParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.OutputParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithSelectParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.SelectParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithTemplateParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.TemplateParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithRenameParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.RenameParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithReplaceParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.ReplaceParameterLayer.ParameterLayerImpl)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func WithFieldsFiltersParameterLayerOptions(options ...layers.ParameterLayerOptions) GlazeParameterLayerOption {
	return func(g *GlazedParameterLayers) error {
		for _, option := range options {
			err := option(g.FieldsFiltersParameterLayer.ParameterLayerImpl)
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
	ret := &GlazedParameterLayers{
		FieldsFiltersParameterLayer: fieldsFiltersParameterLayer,
		OutputParameterLayer:        outputParameterLayer,
		RenameParameterLayer:        renameParameterLayer,
		ReplaceParameterLayer:       replaceParameterLayer,
		SelectParameterLayer:        selectParameterLayer,
		TemplateParameterLayer:      templateParameterLayer,
	}

	for _, option := range options {
		err := option(ret)
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func SetupProcessor(ps map[string]interface{}) (
	*cmds.GlazeProcessor,
	formatters.OutputFormatter,
	error,
) {
	templateSettings, err := NewTemplateSettings(ps)
	if err != nil {
		return nil, nil, err
	}
	outputSettings, err := NewOutputFormatterSettings(ps)
	if err != nil {
		return nil, nil, err
	}
	selectSettings, err := NewSelectSettingsFromParameters(ps)
	if err != nil {
		return nil, nil, err
	}
	renameSettings, err := NewRenameSettingsFromParameters(ps)
	if err != nil {
		return nil, nil, err
	}
	fieldsFilterSettings, err := NewFieldsFilterSettings(ps)
	if err != nil {
		return nil, nil, err
	}
	replaceSettings, err := NewReplaceSettingsFromParameters(ps)
	if err != nil {
		return nil, nil, err
	}

	outputSettings.UpdateWithSelectSettings(selectSettings)
	fieldsFilterSettings.UpdateWithSelectSettings(selectSettings)
	templateSettings.UpdateWithSelectSettings(selectSettings)

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := middlewares.NewFlattenObjectMiddleware()
		of.AddTableMiddleware(mw)
	}
	fieldsFilterSettings.AddMiddlewares(of)

	err = replaceSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding replace middlewares")
	}

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := cmds.NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}
