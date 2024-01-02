package cmds

import (
	"context"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/pkg/errors"
	"math/rand"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

type ExampleCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ExampleCommand)(nil)

type ExampleSettings struct {
	Count int  `glazed.parameter:"count"`
	Test  bool `glazed.parameter:"test"`
}

func NewExampleCommand() (*ExampleCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ExampleCommand{
		CommandDescription: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Example command"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of rows to output"),
					parameters.WithDefault(10),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"test",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Whether to add a test column"),
					parameters.WithDefault(false),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

// RunIntoGlazeProcessor is called to actually execute the command.
//
// parsedLayers contains the result of parsing each layer that has been
// registered with the command description. These layers can be glazed structured data
// flags, database connection parameters, application specification parameters.
//
// ps is a convenience map containing *all* parsed flags.
//
// gp is a GlazeProcessor that can be used to emit rows. Each row is an ordered map.
func (c *ExampleCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
	s := &ExampleSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "failed to initialize example settings from parameters")
	}

	for i := 0; i < s.Count; i++ {
		row := types.NewRow(
			types.MRP("id", i),
			types.MRP("name", "foobar-"+strconv.Itoa(i)),
		)

		if s.Test {
			row.Set("test", rand.Intn(100)+1)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
