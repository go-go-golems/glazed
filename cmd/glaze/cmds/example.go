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
	description *cmds.CommandDescription
}

func NewExampleCommand() (*ExampleCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ExampleCommand{
		description: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Example command"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"count",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Number of rows to output"),
					parameters.WithDefault(10),
				),
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

func (c *ExampleCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *ExampleCommand) Run(
	ctx context.Context,
	parsedLayers map[string]*layers.ParsedParameterLayer,
	ps map[string]interface{},
	gp middlewares.Processor,
) error {
	count := ps["count"].(int)
	test := ps["test"].(bool)

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
