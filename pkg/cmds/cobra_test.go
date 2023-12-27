package cmds

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestAddZeroArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	definitions := parameters.NewParameterDefinitions()
	err := definitions.AddParametersToCobraCommand(cmd, "")
	// assert that err is nil
	require.Nil(t, err)
}

func TestAddSingleRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}

	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
		))
	require.Nil(t, err)
	desc := CommandDescription{
		Layers: layers.NewParameterLayers(layers.WithLayers(defaultLayer)),
	}
	err = defaultLayer.AddLayerToCobraCommand(cmd)

	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"foo", "bla"}, false, false)
	assert.Error(t, err)
}

func TestAddTwoRequiredArguments(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
			&parameters.ParameterDefinition{
				Name:     "bar",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithLayers(defaultLayer))
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar"}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar", "foo", "baz"}, false, false)
	assert.Error(t, err)
}

func TestOneRequiredOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
			&parameters.ParameterDefinition{
				Name:    "bar",
				Type:    parameters.ParameterTypeString,
				Default: cast.InterfaceAddr("baz"),
			},
		),
	)
	require.Nil(t, err)
	desc := CommandDescription{
		Layers: layers.NewParameterLayers(layers.WithLayers(defaultLayer)),
	}
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1.Value)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2.Value)

	values, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1.Value)
	v2, ok = values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "baz", v2.Value)

	_, err = defaultArguments.GatherArguments([]string{}, false, false)
	assert.Error(t, err)

	_, err = defaultArguments.GatherArguments([]string{"bar", "foo", "baz"}, false, false)
	assert.Error(t, err)
}

func TestOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:    "foo",
				Default: cast.InterfaceAddr("123"),
				Type:    parameters.ParameterTypeString,
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithLayers(defaultLayer))
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1.Value)

	values, err = defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "123", v1.Value)
}

func TestDefaultIntValue(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:    "foo",
				Default: cast.InterfaceAddr(123),
				Type:    parameters.ParameterTypeInteger,
			},
		),
	)
	require.Nil(t, err)
	desc := CommandDescription{
		Layers: layers.NewParameterLayers(layers.WithLayers(defaultLayer)),
	}
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 123, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{"234"}, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 234, v1.Value)

	_, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	assert.Error(t, err)
}

type DefaultTypeTestCase struct {
	Type  parameters.ParameterType
	Value interface{}
	Args  []string
}

func TestInvalidDefaultValue(t *testing.T) {
	cmd := &cobra.Command{}
	failingTypes := []DefaultTypeTestCase{
		{Type: parameters.ParameterTypeString, Value: 123},
		{Type: parameters.ParameterTypeString, Value: []string{"foo"}},
		{Type: parameters.ParameterTypeInteger, Value: "foo"},
		{Type: parameters.ParameterTypeInteger, Value: []int{1}},
		// so oddly enough this is a valid word
		{Type: parameters.ParameterTypeDate, Value: "22#@!"},
		{Type: parameters.ParameterTypeStringList, Value: "foo"},
		{Type: parameters.ParameterTypeIntegerList, Value: "foo"},
		{Type: parameters.ParameterTypeStringList, Value: []int{1, 2, 3}},
		{Type: parameters.ParameterTypeStringList, Value: []int{}},
		{Type: parameters.ParameterTypeIntegerList, Value: []string{"1", "2", "3"}},
		{Type: parameters.ParameterTypeIntegerList, Value: []string{}},
	}
	for _, failingType := range failingTypes {
		defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
			layers.WithArguments(
				&parameters.ParameterDefinition{
					Name:    "foo",
					Default: cast.InterfaceAddr(failingType.Value),
					Type:    failingType.Type,
				},
			),
		)
		require.Nil(t, err)
		err = defaultLayer.AddLayerToCobraCommand(cmd)
		if err == nil {
			t.Errorf("Expected error for type %s and value %v\n", failingType.Type, failingType.Value)
		}
		assert.Error(t, err)
	}
}

func TestTwoOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name: "foo",
			},
			&parameters.ParameterDefinition{
				Name: "bar",
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "blop"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestFailAddingRequiredAfterOptional(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name: "foo",
			},
			&parameters.ParameterDefinition{
				Name:     "bar",
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	assert.Error(t, err)
}

func TestAddStringListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeStringList,
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))
}

func TestAddStringListOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:    "foo",
				Type:    parameters.ParameterTypeStringList,
				Default: cast.InterfaceAddr([]string{"baz"}),
			},
		),
	)
	require.Nil(t, err)
	desc := NewCommandDescription("test", WithLayers(defaultLayer))
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	defaultArguments := desc.GetDefaultArguments()
	values, err := defaultArguments.GatherArguments([]string{"bar", "foo"}, false, false)
	require.Nil(t, err)
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"bar", "foo"}, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{"foo"}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"foo"}, v1.Value)

	values, err = defaultArguments.GatherArguments([]string{}, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"baz"}, v1.Value)
}

func TestFailAddingArgumentAfterStringList(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name: "foo",
				Type: parameters.ParameterTypeStringList,
			},
			&parameters.ParameterDefinition{
				Name: "bar",
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	assert.Error(t, err)
}

func TestAddIntegerListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeIntegerList,
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"1", "2"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListRequiredAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
			},
			&parameters.ParameterDefinition{
				Name:     "bar",
				Type:     parameters.ParameterTypeStringList,
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListOptionalAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:     "foo",
				Required: true,
			},
			&parameters.ParameterDefinition{
				Name:    "bar",
				Type:    parameters.ParameterTypeStringList,
				Default: cast.InterfaceAddr([]string{"blop"}),
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
}

func strAddr(v string) *interface{} {
	v_ := interface{}(v)
	return &v_
}

func TestAddStringListOptionalAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name:    "foo",
				Type:    parameters.ParameterTypeString,
				Default: cast.InterfaceAddr("blop"),
			},
			&parameters.ParameterDefinition{
				Name:    "bar",
				Type:    parameters.ParameterTypeStringList,
				Default: cast.InterfaceAddr([]string{"bloppp"}),
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListRequiredAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	defaultLayer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
		layers.WithArguments(
			&parameters.ParameterDefinition{
				Name: "foo",
			},
			&parameters.ParameterDefinition{
				Name:     "bar",
				Type:     parameters.ParameterTypeStringList,
				Required: true,
			},
		),
	)
	require.Nil(t, err)
	err = defaultLayer.AddLayerToCobraCommand(cmd)
	assert.Error(t, err)
}

// expectedCommandResults is a struct that contains the expected results of a command,
// which is a list of parsed flag parameters, a list of parsed arguments parameters,
// as well as potential errors.
//
// This is used to quickly check the result of passing a set of arguments to a command.
type expectedCommandResults struct {
	Name                       string                 `yaml:"name"`
	ExpectedArgumentParameters map[string]interface{} `yaml:"argumentParameters"`
	ExpectedFlagParameters     map[string]interface{} `yaml:"flagParameters"`
	ExpectedFlagError          bool                   `yaml:"flagError"`
	ExpectedArgumentError      bool                   `yaml:"argumentError"`
	Args                       []string               `yaml:"args"`
}

type commandDescription struct {
	CommandDescription
	Flags     []*parameters.ParameterDefinition `yaml:"flags"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments"`
}

type commandTest struct {
	Description *commandDescription       `yaml:"description"`
	Tests       []*expectedCommandResults `yaml:"tests"`
}

//go:embed "test-data/cobra/*.yaml"
var cobraData embed.FS

func TestCommandArgumentsParsing(t *testing.T) {
	// enumerate all the test files in cobraData
	files, err := cobraData.ReadDir("test-data/cobra")
	require.NoError(t, err)

	for _, file := range files {
		// load yaml from file
		testSuite := &commandTest{}
		fileData, err := cobraData.ReadFile("test-data/cobra/" + file.Name())
		require.NoError(t, err)

		err = yaml.Unmarshal(fileData, testSuite)
		require.NoError(t, err)

		layer, err := layers.NewParameterLayer(layers.DefaultSlug, "Default",
			layers.WithArguments(testSuite.Description.Arguments...),
			layers.WithParameterDefinitions(testSuite.Description.Flags...),
		)
		require.NoError(t, err)
		testSuite.Description.Layers = layers.NewParameterLayers(layers.WithLayers(layer))

		if testSuite.Description.Name != "string-from-file" {
			// XXX hack to debug
			continue
		}

		for _, test := range testSuite.Tests {
			test2 := test
			if test2.ExpectedArgumentParameters == nil {
				test2.ExpectedArgumentParameters = map[string]interface{}{}
			}
			if test2.ExpectedFlagParameters == nil {
				test2.ExpectedFlagParameters = map[string]interface{}{}
			}

			t.Run(
				fmt.Sprintf("%s/%s", testSuite.Description.Name, test2.Name),
				func(t *testing.T) {
					testCommandParseHelper(t, &testSuite.Description.CommandDescription, test2)
				})
		}
	}
}

func testCommandParseHelper(
	t *testing.T,
	desc *CommandDescription,
	expected *expectedCommandResults,
) {
	var flagsError error
	var argsError error
	var flagParameters *parameters.ParsedParameters
	var argumentParameters *parameters.ParsedParameters

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			flagParameters, flagsError = desc.GetDefaultFlags().GatherFlagsFromCobraCommand(cmd, false, false, "")
			if flagsError != nil {
				return
			}
			argumentParameters, argsError = desc.GetDefaultArguments().GatherArguments(args, false, false)
			if argsError != nil {
				return
			}
		},
	}

	defaultLayer, ok := desc.GetDefaultLayer()
	require.True(t, ok)
	defaultLayer_, ok := defaultLayer.(layers.CobraParameterLayer)
	require.True(t, ok)

	err := defaultLayer_.AddLayerToCobraCommand(cmd)
	require.Nil(t, err)

	cmd.SetArgs(expected.Args)

	_ = capturer.CaptureStderr(func() {
		err = cmd.Execute()
	})

	if expected.ExpectedFlagError || expected.ExpectedArgumentError {
		assert.Errorf(t, err, "expected error for %v", expected.Args)
	} else {
		assert.NoErrorf(t, err, "unexpected error for %v", expected.Args)
	}

	if err != nil {
		return
	}

	if expected.ExpectedFlagError {
		assert.Errorf(t, flagsError, "expected flag error for %v", expected.Args)
		return
	} else {
		assert.NoErrorf(t, flagsError, "Unexpected error parsing flags: %v", expected.Args)
	}
	if expected.ExpectedArgumentError {
		assert.Errorf(t, argsError, "expected error for %v", expected.Args)
		return
	} else {
		assert.NoErrorf(t, argsError, "expected no error for %v", expected.Args)
	}

	assertJsonEquivalent(t, expected.ExpectedArgumentParameters, argumentParameters)
	assertJsonEquivalent(t, expected.ExpectedFlagParameters, flagParameters)
}

func assertJsonEquivalent(t *testing.T, expected interface{}, actual interface{}) {
	expectedBytes, err := json.MarshalIndent(expected, "", "  ")
	require.NoError(t, err)
	actualBytes, err := json.MarshalIndent(actual, "", "  ")
	require.NoError(t, err)
	assert.JSONEq(t, string(expectedBytes), string(actualBytes))
}
