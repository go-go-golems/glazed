package cmds

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"github.com/zenizh/go-capturer"
	"gopkg.in/yaml.v3"
	"testing"
)

func TestAddZeroArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	// assert that err is nil
	require.Nil(t, err)
}

func TestAddSingleRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))

	values, err := parameters.GatherArguments([]string{"bar"}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1)

	_, err = parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	assert.Error(t, err)

	_, err = parameters.GatherArguments([]string{"foo", "bla"}, desc.Arguments, false, false)
	assert.Error(t, err)
}

func TestAddTwoRequiredArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
			{
				Name:     "bar",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	values, err := parameters.GatherArguments([]string{"bar", "foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2)

	_, err = parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	assert.Error(t, err)

	_, err = parameters.GatherArguments([]string{"bar"}, desc.Arguments, false, false)
	assert.Error(t, err)

	_, err = parameters.GatherArguments([]string{"bar", "foo", "baz"}, desc.Arguments, false, false)
	assert.Error(t, err)
}

func TestOneRequiredOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeString,
			},
			{
				Name:    "bar",
				Type:    parameters.ParameterTypeString,
				Default: "baz",
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	values, err := parameters.GatherArguments([]string{"bar", "foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 2, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "bar", v1)
	v2, ok := values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "foo", v2)

	values, err = parameters.GatherArguments([]string{"foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1)
	v2, ok = values.Get("bar")
	require.True(t, ok)
	assert.Equal(t, "baz", v2)

	_, err = parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	assert.Error(t, err)

	_, err = parameters.GatherArguments([]string{"bar", "foo", "baz"}, desc.Arguments, false, false)
	assert.Error(t, err)
}

func TestOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:    "foo",
				Default: "123",
				Type:    parameters.ParameterTypeString,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	values, err := parameters.GatherArguments([]string{"foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "foo", v1)

	values, err = parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, "123", v1)
}

func TestDefaultIntValue(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:    "foo",
				Default: 123,
				Type:    parameters.ParameterTypeInteger,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	values, err := parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 123, v1)

	values, err = parameters.GatherArguments([]string{"234"}, desc.Arguments, false, false)
	require.Nil(t, err)
	assert.Equal(t, 1, values.Len())
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, 234, v1)

	_, err = parameters.GatherArguments([]string{"foo"}, desc.Arguments, false, false)
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
		desc := CommandDescription{
			Arguments: []*parameters.ParameterDefinition{
				{
					Name:    "foo",
					Default: failingType.Value,
					Type:    failingType.Type,
				},
			},
		}
		err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
		if err == nil {
			t.Errorf("Expected error for type %s and value %v\n", failingType.Type, failingType.Value)
		}
		assert.Error(t, err)
	}
}

func TestTwoOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "blop"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestFailAddingRequiredAfterOptional(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name:     "bar",
				Required: true,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	assert.Error(t, err)
}

func TestAddStringListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeStringList,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))
}

func TestAddStringListOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:    "foo",
				Type:    parameters.ParameterTypeStringList,
				Default: []string{"baz"},
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	values, err := parameters.GatherArguments([]string{"bar", "foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	v1, ok := values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"bar", "foo"}, v1)

	values, err = parameters.GatherArguments([]string{"foo"}, desc.Arguments, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"foo"}, v1)

	values, err = parameters.GatherArguments([]string{}, desc.Arguments, false, false)
	require.Nil(t, err)
	v1, ok = values.Get("foo")
	require.True(t, ok)
	assert.Equal(t, []string{"baz"}, v1)
}

func TestFailAddingArgumentAfterStringList(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name: "foo",
				Type: parameters.ParameterTypeStringList,
			},
			{
				Name: "bar",
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	assert.Error(t, err)
}

func TestAddIntegerListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     parameters.ParameterTypeIntegerList,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"1", "2"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListRequiredAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
			},
			{
				Name:     "bar",
				Type:     parameters.ParameterTypeStringList,
				Required: true,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListOptionalAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
			},
			{
				Name:    "bar",
				Type:    parameters.ParameterTypeStringList,
				Default: []string{"blop"},
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListOptionalAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name:    "foo",
				Type:    parameters.ParameterTypeString,
				Default: "blop",
			},
			{
				Name:    "bar",
				Type:    parameters.ParameterTypeStringList,
				Default: []string{"bloppp"},
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListRequiredAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*parameters.ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name:     "bar",
				Type:     parameters.ParameterTypeStringList,
				Required: true,
			},
		},
	}
	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
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

type commandTest struct {
	Description *CommandDescription       `yaml:"description"`
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
					testCommandParseHelper(t, testSuite.Description, test2)
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
	var flagParameters map[string]interface{}
	var argumentParameters *orderedmap.OrderedMap[string, interface{}]

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			flagParameters, flagsError = parameters.GatherFlagsFromCobraCommand(cmd, desc.Flags, false, false, "")
			if flagsError != nil {
				return
			}
			argumentParameters, argsError = parameters.GatherArguments(args, desc.Arguments, false, false)
			if argsError != nil {
				return
			}
		},
	}

	err := parameters.addArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	err = parameters.AddParametersToCobraCommand(cmd.Flags(), desc.Flags, "")
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
