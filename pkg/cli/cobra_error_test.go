package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	cmdalias "github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cobraErrorTestError struct {
	what string
}

func (e *cobraErrorTestError) Error() string { return "not found: " + e.what }

type cobraErrorBareCommand struct {
	*cmds.CommandDescription
	err error
}

func (c *cobraErrorBareCommand) Run(context.Context, *values.Values) error {
	return c.err
}

type cobraErrorGlazeCommand struct {
	*cmds.CommandDescription
	err error
}

func (c *cobraErrorGlazeCommand) RunIntoGlazeProcessor(
	context.Context,
	*values.Values,
	middlewares.Processor,
) error {
	return c.err
}

func TestBuiltCobraCommandsPropagateCommandErrorsToExecute(t *testing.T) {
	tests := []struct {
		name    string
		command cmds.Command
	}{
		{
			name: "bare command",
			command: &cobraErrorBareCommand{
				CommandDescription: cmds.NewCommandDescription("fail"),
				err:                &cobraErrorTestError{what: "greenhouse"},
			},
		},
		{
			name: "glaze command",
			command: &cobraErrorGlazeCommand{
				CommandDescription: cmds.NewCommandDescription("fail"),
				err:                &cobraErrorTestError{what: "greenhouse"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			built, err := BuildCobraCommandFromCommand(tt.command)
			require.NoError(t, err)
			assert.Nil(t, built.Run, "builders must not terminate errors through a non-returning Run callback")
			require.NotNil(t, built.RunE)

			root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
			root.AddCommand(built)
			root.SetArgs([]string{"fail"})

			err = root.Execute()
			require.Error(t, err)
			var target *cobraErrorTestError
			require.True(t, errors.As(err, &target))
			assert.Equal(t, "greenhouse", target.what)
		})
	}
}

func TestBuiltCobraCommandAliasPropagatesErrors(t *testing.T) {
	expected := &cobraErrorTestError{what: "greenhouse"}
	command := &cobraErrorBareCommand{
		CommandDescription: cmds.NewCommandDescription("fail"),
		err:                expected,
	}
	built, err := BuildCobraCommandAlias(&cmdalias.CommandAlias{
		Name:           "shortcut",
		AliasedCommand: command,
	})
	require.NoError(t, err)

	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(built)
	root.SetArgs([]string{"shortcut"})

	err = root.Execute()
	require.ErrorIs(t, err, expected)
}

func TestCreateStructuredOutputProcessorFromCobraReturnsSetupErrors(t *testing.T) {
	command := &cobra.Command{Use: "raw"}
	require.NoError(t, AddStructuredOutputFlagsToCobraCommand(command))
	require.NoError(t, command.Flags().Set("max-output-rows", "-1"))

	processor, formatter, err := CreateStructuredOutputProcessorFromCobra(command)
	require.EqualError(t, err, "max-output-rows must be greater than or equal to zero")
	assert.Nil(t, processor)
	assert.Nil(t, formatter)
}

func TestBuildCobraCommandFromCommandAndFuncPropagatesErrors(t *testing.T) {
	expected := &cobraErrorTestError{what: "greenhouse"}
	command := &cobraErrorBareCommand{
		CommandDescription: cmds.NewCommandDescription("fail"),
	}
	built, err := BuildCobraCommandFromCommandAndFunc(
		command,
		func(context.Context, *values.Values) error { return expected },
	)
	require.NoError(t, err)

	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	root.AddCommand(built)
	root.SetArgs([]string{"fail"})

	err = root.Execute()
	require.ErrorIs(t, err, expected)
}
