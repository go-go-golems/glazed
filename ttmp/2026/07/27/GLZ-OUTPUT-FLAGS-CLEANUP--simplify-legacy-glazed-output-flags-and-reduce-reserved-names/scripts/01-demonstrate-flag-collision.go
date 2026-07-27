package main

import (
	"context"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
)

type collisionCommand struct {
	description *cmds.CommandDescription
}

func (c *collisionCommand) Description() *cmds.CommandDescription {
	return c.description
}

func (c *collisionCommand) ToYAML(w io.Writer) error {
	return c.description.ToYAML(w)
}

func (c *collisionCommand) RunIntoGlazeProcessor(
	_ context.Context,
	_ *values.Values,
	_ middlewares.Processor,
) error {
	return nil
}

func main() {
	command := &collisionCommand{
		description: cmds.NewCommandDescription(
			"collision",
			cmds.WithFlags(fields.New(
				"output",
				fields.TypeString,
				fields.WithHelp("Business output destination"),
			)),
		),
	}

	_, err := cli.BuildCobraCommandFromCommand(command)
	fmt.Printf("BuildCobraCommandFromCommand error: %v\n", err)
}
