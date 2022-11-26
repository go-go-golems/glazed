package cmds

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"glazed/pkg/cli"
	"glazed/pkg/formatters"
	"glazed/pkg/middlewares"
)

func SetupProcessor(cmd *cobra.Command) (*cli.GlazeProcessor, formatters.OutputFormatter, error) {
	outputSettings, err := cli.ParseOutputFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := cli.ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	fieldsFilterSettings, err := cli.ParseFieldsFilterFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
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

	var middlewares_ []middlewares.ObjectMiddleware
	if templateSettings.UseRowTemplates {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := cli.NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}
