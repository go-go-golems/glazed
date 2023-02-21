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

func (g *GlazedParameterLayers) AddFlag(flag *parameters.ParameterDefinition) {
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

func (g *GlazedParameterLayers) InitializeStructFromDefaults(s interface{}) error {
	panic("implement me")
}

func (g *GlazedParameterLayers) AddFlagsToCobraCommand(cmd *cobra.Command, defaults interface{}) error {
	err := g.OutputParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = g.SelectParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}

	layers.SetFlagGroupOrder(cmd, []string{
		"glazed-output",
		"glazed-select",
		"glazed-template",
		"glazed-fields-filter",
		"glazed-rename",
		"glazed-replace",
	})

	return nil
}

func (g *GlazedParameterLayers) ParseFlagsFromCobraCommand(cmd *cobra.Command) (map[string]interface{}, error) {
	ps, err := g.OutputParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	ps_, err := g.SelectParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.RenameParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.TemplateParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.FieldsFiltersParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	ps_, err = g.ReplaceParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, err
	}
	for k, v := range ps_ {
		ps[k] = v
	}
	return ps, nil
}

func NewGlazedParameterLayers() (*GlazedParameterLayers, error) {
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
	return &GlazedParameterLayers{
		FieldsFiltersParameterLayer: fieldsFiltersParameterLayer,
		OutputParameterLayer:        outputParameterLayer,
		RenameParameterLayer:        renameParameterLayer,
		ReplaceParameterLayer:       replaceParameterLayer,
		SelectParameterLayer:        selectParameterLayer,
		TemplateParameterLayer:      templateParameterLayer,
	}, nil
}

func (g *GlazedParameterLayers) CreateProcessor() (
	*cmds.GlazeProcessor,
	formatters.OutputFormatter,
	error,
) {
	outputSettings := g.OutputParameterLayer.Settings
	selectSettings := g.SelectParameterLayer.Settings
	templateSettings := g.TemplateParameterLayer.Settings
	fieldsFilterSettings := g.FieldsFiltersParameterLayer.Settings
	renameSettings := g.RenameParameterLayer.Settings
	replaceSettings := g.ReplaceParameterLayer.Settings

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

func SetupProcessor(cmd *cobra.Command) (*cmds.GlazeProcessor, formatters.OutputFormatter, error) {
	gpl, err := NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, err
	}

	// TODO(manuel, 2023-02-21) Remove this weird SetupProcessor logic
	//
	// See #150
	//
	// this is abit wonk, because we create the settings structures while parsing the flags,
	// and the ncreate the processor from the settings, when really we should be creating the
	// processor from the parameters themselves, which we could fairly easily do
	// except for the weird regexp replace leftover
	_, err = gpl.ParseFlagsFromCobraCommand(cmd)
	if err != nil {
		return nil, nil, err
	}

	return gpl.CreateProcessor()
}
