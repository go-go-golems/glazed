package cmds

import (
	"context"
	"crypto/rand"
	"math/big"
	"strconv"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
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
	glazedLayer, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ExampleCommand{
		CommandDescription: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Example command"),
			cmds.WithFlags(
				fields.New(
					"count",
					fields.TypeInteger,
					fields.WithHelp("Number of rows to output"),
					fields.WithDefault(10),
				),
			),
			cmds.WithArguments(
				fields.New(
					"test",
					fields.TypeBool,
					fields.WithHelp("Whether to add a test column"),
					fields.WithDefault(false),
				),
			),
			cmds.WithLayersList(
				glazedLayer,
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
func (c *ExampleCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &ExampleSettings{}
	err := values.DecodeSectionInto(vals, schema.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "failed to initialize example settings from parameters")
	}

	for i := 0; i < s.Count; i++ {
		row := types.NewRow(
			types.MRP("id", i),
			types.MRP("name", "foobar-"+strconv.Itoa(i)),
		)

		if s.Test {
			// Generate a secure random number between 1 and 100
			n, err := rand.Int(rand.Reader, big.NewInt(100))
			if err != nil {
				return errors.Wrap(err, "failed to generate random number")
			}
			row.Set("test", int(n.Int64())+1)
		}

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
