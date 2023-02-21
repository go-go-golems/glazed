package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/formatters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// Helpers for cobra commands

type FlagsDefaults struct {
	Select   *SelectFlagsDefaults
	Template *TemplateFlagsDefaults
}

type GlazedParameterLayers struct {
	FieldsFiltersParameterLayer *FieldsFiltersParameterLayer
	OutputParameterLayer        *OutputParameterLayer
	RenameParameterLayer        *RenameParameterLayer
	ReplaceParameterLayer       *ReplaceParameterLayer
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
	return &GlazedParameterLayers{
		FieldsFiltersParameterLayer: fieldsFiltersParameterLayer,
		OutputParameterLayer:        outputParameterLayer,
		RenameParameterLayer:        renameParameterLayer,
		ReplaceParameterLayer:       replaceParameterLayer,
	}, nil
}

func (g *GlazedParameterLayers) NewFlagsDefaults() *FlagsDefaults {
	return &FlagsDefaults{
		Select:   NewSelectFlagsDefaults(),
		Template: NewTemplateFlagsDefaults(),
	}
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
func (g *GlazedParameterLayers) AddFlags(cmd *cobra.Command, defaults *FlagsDefaults) error {
	err := g.OutputParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = AddSelectFlags(cmd, defaults.Select)
	if err != nil {
		return err
	}
	err = g.RenameParameterLayer.AddFlags(cmd)
	if err != nil {
		return err
	}
	err = AddTemplateFlags(cmd, defaults.Template)
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

	cmds.SetFlagGroupOrder(cmd, []string{
		"glazed-output",
		"glazed-select",
		"glazed-template",
		"glazed-fields-filter",
		"glazed-rename",
		"glazed-replace",
	})

	return nil
}

func SetupProcessor(cmd *cobra.Command) (*cmds.GlazeProcessor, formatters.OutputFormatter, error) {
	g, err := NewGlazedParameterLayers()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating glazed parameter layers")
	}

	err = g.OutputParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	err = g.FieldsFiltersParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	selectSettings, err := ParseSelectFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing select flags")
	}
	outputSettings := g.OutputParameterLayer.Settings
	outputSettings.UpdateWithSelectSettings(selectSettings)
	g.FieldsFiltersParameterLayer.Settings.UpdateWithSelectSettings(selectSettings)
	templateSettings.UpdateWithSelectSettings(selectSettings)

	err = g.RenameParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing rename flags")
	}

	err = g.ReplaceParameterLayer.ParseFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing replace flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = g.RenameParameterLayer.Settings.AddMiddlewares(of)
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
	g.FieldsFiltersParameterLayer.Settings.AddMiddlewares(of)

	err = g.ReplaceParameterLayer.Settings.AddMiddlewares(of)
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
