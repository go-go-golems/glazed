package cli

import (
	"context"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type structuredOutputTestCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*structuredOutputTestCommand)(nil)

func (c *structuredOutputTestCommand) RunIntoGlazeProcessor(
	context.Context,
	*values.Values,
	middlewares.Processor,
) error {
	return nil
}

func newStructuredOutputTestCommand(name string, flags ...*fields.Definition) *structuredOutputTestCommand {
	return &structuredOutputTestCommand{CommandDescription: cmds.NewCommandDescription(
		name,
		cmds.WithSource("test"),
		cmds.WithFlags(flags...),
	)}
}

func TestGlazeCommandMountsOnlyMinimalStructuredOutputFlags(t *testing.T) {
	cmd, err := BuildCobraCommandFromCommand(newStructuredOutputTestCommand("rows"))
	require.NoError(t, err)

	for _, name := range []string{"format", "output-fields", "max-output-rows"} {
		assert.NotNil(t, cmd.Flags().Lookup(name), name)
	}
	for _, name := range []string{
		"output", "output-file", "fields", "filter", "regex-fields",
		"sort-columns", "remove-nulls", "rename", "replace-file", "select",
		"template", "jq", "sort-by", "glazed-skip", "glazed-limit",
		"stream", "table-style", "sheet-name", "sql-table-name",
	} {
		assert.Nil(t, cmd.Flags().Lookup(name), name)
	}
}

func TestGlazeCommandAllowsFormerGenericFlagNames(t *testing.T) {
	applicationFlags := []string{
		"output", "output-file", "fields", "filter", "template", "select", "stream", "sort-by",
	}
	definitions := make([]*fields.Definition, 0, len(applicationFlags))
	for _, name := range applicationFlags {
		definitions = append(definitions, fields.New(name, fields.TypeString))
	}

	cmd, err := BuildCobraCommandFromCommand(newStructuredOutputTestCommand("rows", definitions...))
	require.NoError(t, err)
	for _, name := range applicationFlags {
		assert.NotNil(t, cmd.Flags().Lookup(name), name)
	}
	assert.NotNil(t, cmd.Flags().Lookup("format"))
}

func TestAddCommandsToRootIsAtomicOnStructuredOutputCollision(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	valid := newStructuredOutputTestCommand("valid")
	invalid := newStructuredOutputTestCommand(
		"invalid",
		fields.New("format", fields.TypeString, fields.WithHelp("application format")),
	)

	err := AddCommandsToRootCommand(root, []cmds.Command{valid, invalid}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not build command invalid from test")
	assert.Empty(t, root.Commands())
}
