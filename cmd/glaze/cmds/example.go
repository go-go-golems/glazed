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
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}, nil
}

// Run is called to actually execute the command.
//
// parsedLayers contains the result of parsing each layer that has been
// registered with the command description. These layers can be glazed structured data
// flags, database connection parameters, application specification parameters.
//
// ps is a convenience map containing *all* parsed flags.
//
// gp is a GlazeProcessor that can be used to emit rows. Each row is an ordered map.
func (c *ExampleCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	gp middlewares.Processor,
) error {
	d, ok := parsedLayers["default"]
	if !ok {
		return errors.New("no default layer")
	}
	count := d.Parameters["count"].(int)
	test := d.Parameters["test"].(bool)

	for i := 0; i < count; i++ {
		row := types.NewRow(
			types.MRP("id", i),
			types.MRP("name", "foobar-"+strconv.Itoa(i)),
		)

		if test {
			row.Set("test", rand.Intn(100)+1)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
