package cmds

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"
	"testing"
	"time"
)

func TestAddZeroArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{},
	}
	err := AddArguments(cmd, &desc)
	// assert that err is nil
	require.Nil(t, err)
}

func TestAddSingleRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeString,
			},
		},
	}
	err := AddArguments(cmd, &desc)
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
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
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
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
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
		Arguments: []*Parameter{
			{
				Name:    "foo",
				Default: "123",
				Type:    ParameterTypeString,
			},
		},
	}
	err := AddArguments(cmd, &desc)
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
		Arguments: []*Parameter{
			{
				Name:    "foo",
				Default: 123,
				Type:    ParameterTypeInteger,
			},
		},
	}
	err := AddArguments(cmd, &desc)
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
		result, err := parseDate(testCase.Value)
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
		param := &Parameter{
			Name:    "foo",
			Default: testCase.Value,
			Type:    testCase.Type,
		}
		err := param.CheckParameterDefaultValueValidity()
		assert.Nil(t, err)
	}
}

func TestValidChoiceDefaultValue(t *testing.T) {
	param := &Parameter{
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
		param := &Parameter{
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
			Arguments: []*Parameter{
				{
					Name:    "foo",
					Default: failingType.Value,
					Type:    failingType.Type,
				},
			},
		}
		err := AddArguments(cmd, &desc)
		if err == nil {
			t.Errorf("Expected error for type %s and value %v\n", failingType.Type, failingType.Value)
		}
		assert.Error(t, err)
	}
}

func TestTwoOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
		},
	}
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Error(t, cmd.Args(cmd, []string{"bar", "foo", "blop"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestFailAddingRequiredAfterOptional(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name: "foo",
			},
			{
				Name:     "bar",
				Required: true,
			},
		},
	}
	err := AddArguments(cmd, &desc)
	assert.Error(t, err)
}

func TestAddStringListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeStringList,
			},
		},
	}
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"bar", "foo", "baz"}))
}

func TestAddStringListOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name:    "foo",
				Type:    ParameterTypeStringList,
				Default: []string{"baz"},
			},
		},
	}
	err := AddArguments(cmd, &desc)
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
		Arguments: []*Parameter{
			{
				Name: "foo",
				Type: ParameterTypeStringList,
			},
			{
				Name: "bar",
			},
		},
	}
	err := AddArguments(cmd, &desc)
	assert.Error(t, err)
}

func TestAddIntegerListRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name:     "foo",
				Required: true,
				Type:     ParameterTypeIntegerList,
			},
		},
	}
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"1", "2"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Nil(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListRequiredAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
	assert.Error(t, cmd.Args(cmd, []string{"1"}))
	assert.Nil(t, cmd.Args(cmd, []string{"1", "4", "2"}))
}

func TestAddStringListOptionalAfterRequiredArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Error(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListOptionalAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar", "baz"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo", "bar"}))
	assert.Nil(t, cmd.Args(cmd, []string{"foo"}))
	assert.Nil(t, cmd.Args(cmd, []string{}))
}

func TestAddStringListRequiredAfterOptionalArgument(t *testing.T) {
	cmd := &cobra.Command{}
	desc := CommandDescription{
		Arguments: []*Parameter{
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
	err := AddArguments(cmd, &desc)
	assert.Error(t, err)
}

type expectedCommandResults struct {
	ExpectedArgumentParameters map[string]interface{}
	ExpectedFlagParameters     map[string]interface{}
	ExpectedFlagError          bool
	ExpectedArgumentError      bool
	Args                       []string
}

func TestGatherCommand(t *testing.T) {
	desc := CommandDescription{
		Arguments: []*Parameter{
			{
				Name:     "foo",
				Type:     ParameterTypeString,
				Required: true,
			},
			{
				Name:     "bar",
				Type:     ParameterTypeStringList,
				Required: true,
			},
		},
		Flags: []*Parameter{
			{
				Name:    "baz",
				Type:    ParameterTypeString,
				Default: "blop",
			},
		},
	}

	expectedResults := []expectedCommandResults{
		{
			Args: []string{"--baz", "blip", "foo", "bar", "baz"},
			ExpectedArgumentParameters: map[string]interface{}{
				"foo": "foo",
				"bar": []string{"bar", "baz"},
			},
			ExpectedFlagParameters: map[string]interface{}{
				"baz": "blip",
			},
		},
		{
			Args: []string{"foo", "bar"},
			ExpectedArgumentParameters: map[string]interface{}{
				"foo": "foo",
				"bar": []string{"bar"},
			},
			ExpectedFlagParameters: map[string]interface{}{
				"baz": "blop",
			},
		},
		{
			Args:                  []string{"foo"},
			ExpectedArgumentError: true,
		},
	}

	for _, expected := range expectedResults {
		testCommandParseHelper(t, desc, &expected)
	}

}

func testCommandParseHelper(t *testing.T, desc CommandDescription, expected *expectedCommandResults) {
	var flagsError error
	var argsError error
	var flagParameters map[string]interface{}
	var argumentParameters map[string]interface{}

	cmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			flagParameters, flagsError = GatherFlags(cmd, desc.Flags, false)
			if flagsError != nil {
				return
			}
			argumentParameters, argsError = GatherArguments(args, desc.Arguments, false)
			if argsError != nil {
				return
			}
		},
	}

	err := AddArguments(cmd, &desc)
	require.Nil(t, err)
	err = AddFlags(cmd, &desc)
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

	assert.Equal(t, expected.ExpectedArgumentParameters, argumentParameters)
	assert.Equal(t, expected.ExpectedFlagParameters, flagParameters)
}
