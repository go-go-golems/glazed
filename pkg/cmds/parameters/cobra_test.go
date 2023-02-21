package parameters

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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
