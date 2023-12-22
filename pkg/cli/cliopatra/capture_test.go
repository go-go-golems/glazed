package cliopatra

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func makeParsedDefaultLayer(desc *cmds.CommandDescription, ps *parameters.ParsedParameters) map[string]*layers.ParsedParameterLayer {
	defaultLayer, ok := desc.GetLayer("default")
	if !ok {
		return nil
	}

	return map[string]*layers.ParsedParameterLayer{
		"default": {
			Layer:      defaultLayer,
			Parameters: ps,
		},
	}
}

func TestSingleFlag(t *testing.T) {
	testPd := parameters.NewParameterDefinition("test", parameters.ParameterTypeString)
	desc := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			testPd,
		),
	)
	p := NewProgramFromCapture(
		desc,
		makeParsedDefaultLayer(desc, parameters.NewParsedParameters(parameters.WithParsedParameter(testPd, "test", "foobar"))),
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
	pdTest := parameters.NewParameterDefinition("test",
		parameters.ParameterTypeString,
		parameters.WithDefault("foobar"),
		parameters.WithHelp("testing help"),
	)
	d := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			pdTest,
		),
	)
	p := NewProgramFromCapture(d, makeParsedDefaultLayer(d, parameters.NewParsedParameters(parameters.WithParsedParameter(pdTest, "test", "foobar"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 0)

	p = NewProgramFromCapture(d, makeParsedDefaultLayer(d, parameters.NewParsedParameters(parameters.WithParsedParameter(pdTest, "test", "foobar2"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "testing help", p.Flags[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Flags[0].Type)
	assert.Equal(t, "foobar2", p.Flags[0].Value)
}

func TestTwoFlags(t *testing.T) {
	pd1 := parameters.NewParameterDefinition("test", parameters.ParameterTypeString)
	pd2 := parameters.NewParameterDefinition("test2", parameters.ParameterTypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithFlags(
			pd1,
			pd2,
		),
	)

	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d, parameters.NewParsedParameters(
			parameters.WithParsedParameter(pd1, "test", "foobar"),
			parameters.WithParsedParameter(pd2, "test2", "foobar2"),
		)))

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
	pd := parameters.NewParameterDefinition("test", parameters.ParameterTypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithArguments(
			pd,
		),
	)
	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d,
			parameters.NewParsedParameters(
				parameters.WithParsedParameter(pd, "test", "foobar"))))

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Args, 1)
	assert.Equal(t, "test", p.Args[0].Name)
	assert.Equal(t, "", p.Args[0].Short)
	assert.Equal(t, parameters.ParameterTypeString, p.Args[0].Type)
	assert.Equal(t, "foobar", p.Args[0].Value)
}

func TestTwoArgsTwoFlags(t *testing.T) {
	pd1 := parameters.NewParameterDefinition("test", parameters.ParameterTypeString)
	pd2 := parameters.NewParameterDefinition("test2", parameters.ParameterTypeString)
	pd3 := parameters.NewParameterDefinition("test3", parameters.ParameterTypeString)
	pd4 := parameters.NewParameterDefinition("test4", parameters.ParameterTypeString)
	d := cmds.NewCommandDescription("test",
		cmds.WithArguments(
			pd1,
			pd2,
		),
		cmds.WithFlags(
			pd3,
			pd4,
		),
	)
	p := NewProgramFromCapture(
		d,
		makeParsedDefaultLayer(d, parameters.NewParsedParameters(
			parameters.WithParsedParameter(pd1, "test", "foobar"),
			parameters.WithParsedParameter(pd2, "test2", "foobar2"),
			parameters.WithParsedParameter(pd3, "test3", "foobar3"),
			parameters.WithParsedParameter(pd4, "test4", "foobar4"),
		)),
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
	pd := parameters.NewParameterDefinition("test", parameters.ParameterTypeString)
	layer, err2 := layers.NewParameterLayer("test-layer", "test-layer",
		layers.WithParameters(
			pd,
		),
	)
	require.NoError(t, err2)

	d := cmds.NewCommandDescription("test",
		cmds.WithLayers(
			layer,
		),
	)
	p := NewProgramFromCapture(
		d,
		map[string]*layers.ParsedParameterLayer{
			"test-layer": {
				Layer: layer,
				Parameters: parameters.NewParsedParameters(
					parameters.WithParsedParameter(pd, "test", "foobar"))}})

	assert.Equal(t, "test", p.Name)
	assert.Equal(t, "", p.Description)
	assert.Len(t, p.Flags, 1)
	assert.Equal(t, "test", p.Flags[0].Name)
	assert.Equal(t, "", p.Flags[0].Short)
}
