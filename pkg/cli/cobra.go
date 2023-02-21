package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
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

type CobraParameterLayer interface {
	AddFlags(cmd *cobra.Command) error
	ParseFlags(cmd *cobra.Command) error
}

// AddFlags adds all the glazed processing layer flags to a cobra.Command
//
// TODO(manuel, 2023-02-21) Interfacing directly with cobra to do glazed flags handling is deprecated
// As we are moving towards #150 we should do all of this through the parameter definitions instead.
//
// This could probably be modelled as something of a "Layer" class that can be added to a command.
//
// NOTE(manuel, 2023-02-21) It looks like this is basically a CobraParameterLayer too
// So maybe there is something there, that entire libraries can just very easily provide collections of layers.
// I don't think I need to do anything special here, actually.
//
// Actually, that is not true, I can to the SetupProcessor independent of the cobra parsing
// by collecting all the settings necessary to create it from the results of calling
// ParseFlags on the GlazedParameterLayers.
func (g *GlazedParameterLayers) AddFlags(cmd *cobra.Command) error {
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

func (g *GlazedParameterLayers) ParseFlags(cmd *cobra.Command) error {
	err := g.OutputParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	err = g.SelectParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	err = g.TemplateParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	err = g.FieldsFiltersParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	err = g.ReplaceParameterLayer.ParseFlags(cmd)
	if err != nil {
		return err
	}
	return nil

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

	err = gpl.ParseFlags(cmd)
	if err != nil {
		return nil, nil, err
	}

	return gpl.CreateProcessor()
}
