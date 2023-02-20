package cmds

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
	"gopkg.in/yaml.v3"
	"testing"
	"time"
)

func TestAddZeroArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	// assert that err is nil
	require.Nil(t, err)
}

func TestAddSingleRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeString,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))

	values, err := GatherArguments([]string{"bar"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "bar", values["foo"])

	_, err = GatherArguments([]string{}, desc.Arguments, false)
	assert.Error(t, err)

	_, err = GatherArguments([]string{"foo", "bla"}, desc.Arguments, false)
	assert.Error(t, err)
}

func TestAddTwoRequiredArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeString,
			},
			{
				Name:     "bar",
				Required: true,
				Type:     ParameterTypeString,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar"}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	values, err := GatherArguments([]string{"bar", "foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 2, len(values))
	assert.Equal(t, "bar", values["foo"])
	assert.Equal(t, "foo", values["bar"])

	_, err = GatherArguments([]string{}, desc.Arguments, false)
	assert.Error(t, err)

	_, err = GatherArguments([]string{"bar"}, desc.Arguments, false)
	assert.Error(t, err)

	_, err = GatherArguments([]string{"bar", "foo", "baz"}, desc.Arguments, false)
	assert.Error(t, err)
}

func TestOneRequiredOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeString,
			},
			{
				Name:    "bar",
				Type:    ParameterTypeString,
				Default: "baz",
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))

	values, err := GatherArguments([]string{"bar", "foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 2, len(values))
	assert.Equal(t, "bar", values["foo"])
	assert.Equal(t, "foo", values["bar"])

	values, err = GatherArguments([]string{"foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 2, len(values))
	assert.Equal(t, "foo", values["foo"])
	assert.Equal(t, "baz", values["bar"])

	_, err = GatherArguments([]string{}, desc.Arguments, false)
	assert.Error(t, err)

	_, err = GatherArguments([]string{"bar", "foo", "baz"}, desc.Arguments, false)
	assert.Error(t, err)
}

func TestOneOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:    "foo",
				Default: "123",
				Type:    ParameterTypeString,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	values, err := GatherArguments([]string{"foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "foo", values["foo"])

	values, err = GatherArguments([]string{}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, "123", values["foo"])
}

func TestDefaultIntValue(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:    "foo",
				Default: 123,
				Type:    ParameterTypeInteger,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	values, err := GatherArguments([]string{}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, 123, values["foo"])

	values, err = GatherArguments([]string{"234"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, 1, len(values))
	assert.Equal(t, 234, values["foo"])

	_, err = GatherArguments([]string{"foo"}, desc.Arguments, false)
	assert.Error(t, err)
}

func TestParseDate(t *testing.T) {
	// set default time for unit tests
	refTime_ := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	refTime = &refTime_

	testCases := []struct {
		Value  string
		Result time.Time
	}{
		{Value: "2018-01-01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "2018/01/01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		//{Value: "January First 2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "January 1st 2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "2018-01-01T00:00:00+00:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "2018-01-01T00:00:00+01:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.FixedZone("", 3600))},
		{Value: "2018-01-01T00:00:00-01:00", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.FixedZone("", -3600))},
		{Value: "2018", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "2018-01", Result: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "last year", Result: time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "last hour", Result: time.Date(2017, 12, 31, 23, 0, 0, 0, time.UTC)},
		{Value: "last month", Result: time.Date(2017, 12, 1, 0, 0, 0, 0, time.UTC)},
		{Value: "last week", Result: time.Date(2017, 12, 25, 0, 0, 0, 0, time.UTC)},
		{Value: "last monday", Result: time.Date(2017, 12, 25, 0, 0, 0, 0, time.UTC)},
		{Value: "10 days ago", Result: time.Date(2017, 12, 22, 0, 0, 0, 0, time.UTC)},
	}

	for _, testCase := range testCases {
		result, err := ParseDate(testCase.Value)
		require.Nil(t, err)
		if !result.Equal(testCase.Result) {
			t.Errorf("Expected %s to parse to %s, got %s", testCase.Value, testCase.Result, result)
		}
	}
}

type DefaultTypeTestCase struct {
	Type  ParameterType
	Value interface{}
	Args  []string
}

func TestValidDefaultValue(t *testing.T) {
	testCases := []DefaultTypeTestCase{
		{Type: ParameterTypeString, Value: "foo"},
		{Type: ParameterTypeInteger, Value: 123},
		{Type: ParameterTypeBool, Value: false},
		{Type: ParameterTypeDate, Value: "2018-01-01"},
		{Type: ParameterTypeStringList, Value: []string{"foo", "bar"}},
		{Type: ParameterTypeIntegerList, Value: []int{1, 2, 3}},
		{Type: ParameterTypeStringList, Value: []string{}},
		{Type: ParameterTypeIntegerList, Value: []int{}},
	}
	for _, testCase := range testCases {
		param := &ParameterDefinition{
			Name:    "foo",
			Default: testCase.Value,
			Type:    testCase.Type,
		}
		err := param.CheckParameterDefaultValueValidity()
		assert.Nil(t, err)
	}
}

func TestValidChoiceDefaultValue(t *testing.T) {
	param := &ParameterDefinition{
		Name:    "foo",
		Default: "bar",
		Type:    ParameterTypeChoice,
		Choices: []string{"foo", "bar"},
	}
	err := param.CheckParameterDefaultValueValidity()
	assert.Nil(t, err)
}

func TestInvalidChoiceDefaultValue(t *testing.T) {
	testCases := []interface{}{
		"baz",
		123,
		"flop",
	}
	for _, testCase := range testCases {
		param := &ParameterDefinition{
			Name:    "foo",
			Default: testCase,
			Type:    ParameterTypeChoice,
			Choices: []string{"foo", "bar"},
		}
		err := param.CheckParameterDefaultValueValidity()
		assert.Error(t, err)
	}
}

func TestInvalidDefaultValue(t *testing.T) {
	cmd := &cobra.Command{}
	failingTypes := []DefaultTypeTestCase{
		{Type: ParameterTypeString, Value: 123},
		{Type: ParameterTypeString, Value: []string{"foo"}},
		{Type: ParameterTypeInteger, Value: "foo"},
		{Type: ParameterTypeInteger, Value: []int{1}},
		// so oddly enough this is a valid word
		{Type: ParameterTypeDate, Value: "22#@!"},
		{Type: ParameterTypeStringList, Value: "foo"},
		{Type: ParameterTypeIntegerList, Value: "foo"},
		{Type: ParameterTypeStringList, Value: []int{1, 2, 3}},
		{Type: ParameterTypeStringList, Value: []int{}},
		{Type: ParameterTypeIntegerList, Value: []string{"1", "2", "3"}},
		{Type: ParameterTypeIntegerList, Value: []string{}},
	}
	for _, failingType := range failingTypes {
		desc := CommandDescription{
			Arguments: []*ParameterDefinition{
				{
					Name:    "foo",
					Default: failingType.Value,
					Type:    failingType.Type,
				},
			},
		}
		err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
		if err == nil {
			t.Errorf("Expected error for type %s and value %v\n", failingType.Type, failingType.Value)
		}
		assert.Error(t, err)
	}
}

func TestTwoOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "blop"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestFailAddingRequiredAfterOptional(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name:     "bar",
				Required: true,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	assert.Error(t, err)
}

func TestAddStringListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeStringList,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))
}

func TestAddStringListOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:    "foo",
				Type:    ParameterTypeStringList,
				Default: []string{"baz"},
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))

	values, err := GatherArguments([]string{"bar", "foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, []string{"bar", "foo"}, values["foo"])

	values, err = GatherArguments([]string{"foo"}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, []string{"foo"}, values["foo"])

	values, err = GatherArguments([]string{}, desc.Arguments, false)
	require.Nil(t, err)
	assert.Equal(t, []string{"baz"}, values["foo"])
}

func TestFailAddingArgumentAfterStringList(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name: "foo",
				Type: ParameterTypeStringList,
			},
			{
				Name: "bar",
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	assert.Error(t, err)
}

func TestAddIntegerListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeIntegerList,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"1", "2"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListRequiredAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
			},
			{
				Name:     "bar",
				Type:     ParameterTypeStringList,
				Required: true,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListOptionalAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:     "foo",
				Required: true,
			},
			{
				Name:    "bar",
				Type:    ParameterTypeStringList,
				Default: []string{"blop"},
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListOptionalAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name:    "foo",
				Type:    ParameterTypeString,
				Default: "blop",
			},
			{
				Name:    "bar",
				Type:    ParameterTypeStringList,
				Default: []string{"bloppp"},
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListRequiredAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*ParameterDefinition{
			{
				Name: "foo",
			},
			{
				Name:     "bar",
				Type:     ParameterTypeStringList,
				Required: true,
			},
		},
	}
	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
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
	var argumentParameters map[string]interface{}

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			flagParameters, flagsError = GatherFlagsFromCobraCommand(cmd, desc.Flags, false)
			if flagsError != nil {
				return
			}
			argumentParameters, argsError = GatherArguments(args, desc.Arguments, false)
			if argsError != nil {
				return
			}
		},
	}

	err := AddArgumentsToCobraCommand(cmd, desc.Arguments)
	require.Nil(t, err)
	err = AddFlagsToCobraCommand(cmd.Flags(), desc.Flags)
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
