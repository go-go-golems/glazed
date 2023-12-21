package cliopatra

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleFlag(t *testing.T) {
	p := NewProgramFromCapture(
		cmds.NewCommandDescription("test",
			cmds.WithFlags(
				parameters.NewParameterDefinition("test", parameters.ParameterTypeString),
			),
		),
		map[string]interface{}{
			"test": "foobar",
		},
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar", p.Flags[0].Value)
}

func TestSingleFlagDefaultValue(t *testing.T) {
	d := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			parameters.NewParameterDefinition("test",
				parameters.ParameterTypeString,
				parameters.WithDefault("foobar"),
				parameters.WithHelp("testing help"),
			),
		),
	)
	p := NewProgramFromCapture(d, map[string]interface{}{})

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 0)

	p = NewProgramFromCapture(d, map[string]interface{}{
		"test": "foobar2",
	})

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "testing help", p.Flags[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar2", p.Flags[0].Value)
}

func TestTwoFlags(t *testing.T) {
	p := NewProgramFromCapture(
		cmds.NewCommandDescription("test",
			cmds.WithFlags(
				parameters.NewParameterDefinition("test", parameters.ParameterTypeString),
				parameters.NewParameterDefinition("test2", parameters.ParameterTypeString),
			),
		),
		map[string]interface{}{
			"test":  "foobar",
			"test2": "foobar2",
		},
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 2)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar", p.Flags[0].Value)
	assert.Equal(t, "test2", p.Flags[1].Name)
	assert.Equal(t, "", p.Flags[1].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[1].Type)
	assert.Equal(t, "foobar2", p.Flags[1].Value)
}

func TestSingleArg(t *testing.T) {
	p := NewProgramFromCapture(
		cmds.NewCommandDescription("test",
			cmds.WithArguments(
				parameters.NewParameterDefinition("test", parameters.ParameterTypeString),
			),
		),
		map[string]interface{}{
			"test": "foobar",
		},
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Args, 1)
	assert.Equal(t, "test", p.Args[0].Name)
	assert.Equal(t, "", p.Args[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Args[0].Type)
	assert.Equal(t, "foobar", p.Args[0].Value)
}

func TestTwoArgsTwoFlags(t *testing.T) {
	p := NewProgramFromCapture(
		cmds.NewCommandDescription("test",
			cmds.WithArguments(
				parameters.NewParameterDefinition("test", parameters.ParameterTypeString),
				parameters.NewParameterDefinition("test2", parameters.ParameterTypeString),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition("test3", parameters.ParameterTypeString),
				parameters.NewParameterDefinition("test4", parameters.ParameterTypeString),
			),
		),
		map[string]interface{}{
			"test":  "foobar",
			"test2": "foobar2",
			"test3": "foobar3",
			"test4": "foobar4",
		},
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Args, 2)
	assert.Equal(t, "test", p.Args[0].Name)
	assert.Equal(t, "", p.Args[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Args[0].Type)
	assert.Equal(t, "foobar", p.Args[0].Value)
	assert.Equal(t, "test2", p.Args[1].Name)
	assert.Equal(t, "", p.Args[1].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Args[1].Type)
	assert.Equal(t, "foobar2", p.Args[1].Value)
	assert.Len(t, p.Flags, 2)
	assert.Equal(t, "test3", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar3", p.Flags[0].Value)
	assert.Equal(t, "test4", p.Flags[1].Name)
	assert.Equal(t, "", p.Flags[1].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[1].Type)
	assert.Equal(t, "foobar4", p.Flags[1].Value)
}

func TestSingleLayer(t *testing.T) {
	layer, err2 := layers.NewParameterLayer("test-layer", "test-layer",
		layers.WithParameters(
			parameters.NewParameterDefinition("test", parameters.ParameterTypeString),
		),
	)
	require.NoError(t, err2)

	p := NewProgramFromCapture(
		cmds.NewCommandDescription("test",
			cmds.WithLayers(
				layer,
			),
		),
		map[string]interface{}{
			"test":  "foobar",
			"test2": "foobar2",
		},
	)

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
}
