package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func createSimpleParameterLayer(t *testing.T, options ...ParameterLayerOptions) *ParameterLayerImpl {
	options_ := append([]ParameterLayerOptions{
		WithFlags(
			parameters.NewParameterDefinition("flag1", parameters.ParameterTypeString),
		),
	}, options...)
	layer, err := NewParameterLayer("simple", "Simple", options_...)

	require.NoError(t, err)
	return layer
}

func TestAddFlagsToCobraCommandSimple(t *testing.T) {
	layer := createSimpleParameterLayer(t)

	cmd := &cobra.Command{
		Use: "test",
	}

	err := layer.AddFlagsToCobraCommand(cmd)
	require.NoError(t, err)

	flagGroupUsage := ComputeCommandFlagGroupUsage(cmd)
	localGroupUsage := flagGroupUsage.LocalGroupUsages
	require.Len(t, localGroupUsage, 2)

	usage := localGroupUsage[0]
	require.Equal(t, "Flags", usage.Name)

	usage = localGroupUsage[1]
	require.Equal(t, "Simple", usage.Name)
	flagUsages := usage.FlagUsages
	require.Len(t, flagUsages, 1)
	flagUsage := flagUsages[0]
	require.Equal(t, "flag1", flagUsage.Long)
}

func TestAddFlagsToCobraCommandPrefix(t *testing.T) {
	layer := createSimpleParameterLayer(t, WithPrefix("test-"))

	cmd := &cobra.Command{
		Use: "test",
	}

	err := layer.AddFlagsToCobraCommand(cmd)
	require.NoError(t, err)

	flagGroupUsage := ComputeCommandFlagGroupUsage(cmd)
	localGroupUsage := flagGroupUsage.LocalGroupUsages
	require.Len(t, localGroupUsage, 2)

	usage := localGroupUsage[0]
	require.Equal(t, "Flags", usage.Name)

	usage = localGroupUsage[1]
	require.Equal(t, "Simple", usage.Name)
	flagUsages := usage.FlagUsages
	require.Len(t, flagUsages, 1)
	flagUsage := flagUsages[0]
	require.Equal(t, "test-flag1", flagUsage.Long)
}
