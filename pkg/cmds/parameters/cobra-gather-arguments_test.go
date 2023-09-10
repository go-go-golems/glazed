package parameters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test no arguments are passed to the function
func TestGatherArguments_NoArguments(t *testing.T) {
	_, err := GatherArguments([]string{}, []*ParameterDefinition{}, true)
	assert.NoError(t, err)
}

// Test missing required argument
func TestGatherArguments_RequiredMissing(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name:     "Test",
			Required: true,
		},
	}
	_, err := GatherArguments([]string{}, arg, true)
	assert.EqualError(t, err, "Argument Test not found")
}

// Test the parsing of every kind of parameter type for provided args
// This should be broken down into individual tests for each parameter types.
// However a generic example of such test might look like:
func TestGatherArguments_ParsingProvidedArguments(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name: "Test",
			Type: ParameterTypeString,
		},
	}
	res, err := GatherArguments([]string{"value"}, arg, true)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, v, "value")
}

// Test parsing of list-type parameter with multiple arguments
func TestGatherArguments_ListParameterParsing(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name: "Test",
			Type: ParameterTypeStringList,
		},
	}
	res, err := GatherArguments([]string{"value1", "value2"}, arg, true)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, v, []string{"value1", "value2"})
}

// Test handling of default values when onlyProvided is set to false
func TestGatherArguments_DefaultsWhenProvidedFalse(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name:    "Test",
			Type:    ParameterTypeString,
			Default: "default",
		},
	}
	res, err := GatherArguments([]string{}, arg, false)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, v, "default")
}

// Test handling of default values when onlyProvided is set to true
func TestGatherArguments_NoDefaultsWhenProvidedTrue(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name:    "Test",
			Type:    ParameterTypeString,
			Default: "default",
		},
	}
	v, err := GatherArguments([]string{}, arg, true)
	assert.NoError(t, err)
	// check that Test is not in v
	_, present := v.Get("Test")
	assert.False(t, present)
}

// Test the error condition of providing too many arguments
func TestGatherArguments_TooManyArguments(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name: "Test",
			Type: ParameterTypeString,
		},
	}
	v, err := GatherArguments([]string{"value1", "value2"}, arg, true)
	_ = v
	assert.EqualError(t, err, "Too many arguments")
}

// Test the correct sequencing of arguments
func TestGatherArguments_CorrectSequence(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name: "Test1",
			Type: ParameterTypeString,
		},
		{
			Name: "Test2",
			Type: ParameterTypeString,
		},
	}
	res, err := GatherArguments([]string{"value1", "value2"}, arg, true)
	assert.NoError(t, err)
	v1, present := res.Get("Test1")
	assert.True(t, present)
	assert.Equal(t, v1, "value1")
	v2, present := res.Get("Test2")
	assert.True(t, present)
	assert.Equal(t, v2, "value2")
}

// Test various combinations of list and non-list arguments
func TestGatherArguments_CombinationsListNonList(t *testing.T) {
	arg := []*ParameterDefinition{
		{
			Name: "Test1",
			Type: ParameterTypeString,
		},
		{
			Name: "Test2",
			Type: ParameterTypeStringList,
		},
	}
	res, err := GatherArguments([]string{"value1", "value2", "value3"}, arg, true)
	assert.NoError(t, err)
	v1, present := res.Get("Test1")
	assert.True(t, present)
	assert.Equal(t, v1, "value1")
	v2, present := res.Get("Test2")
	assert.True(t, present)
	assert.Equal(t, v2, []string{"value2", "value3"})
}

func TestListParsingWithDefaults(t *testing.T) {
	args := []string{"data1", "data2"}
	arguments := []*ParameterDefinition{
		{
			Name:    "arg1",
			Type:    ParameterTypeStringList,
			Default: []string{"default1", "default2"},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"data1", "data2"}, v1)
}

func TestListDefault(t *testing.T) {
	args := []string{}
	arguments := []*ParameterDefinition{
		{
			Name:    "arg1",
			Type:    ParameterTypeStringList,
			Default: []string{"default1", "default2"},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"default1", "default2"}, v1)
}

func TestIntegerListParsing(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeIntegerList,
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []int{1, 2, 3}, v2)
}

func TestFloatListParsing(t *testing.T) {
	args := []string{"1.1", "2.2", "3.3"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeFloatList,
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, v2)
}

func TestChoiceListParsing(t *testing.T) {
	args := []string{"choice1", "choice2", "choice3"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeChoiceList,
			Choices: []string{
				"choice1",
				"choice2",
				"choice3",
			},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"choice1", "choice2", "choice3"}, v2)
}

func TestParsingErrorInvalidInt(t *testing.T) {
	args := []string{"1", "2", "3", "notanint"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeIntegerList,
		},
	}
	_, err := GatherArguments(args, arguments, false)
	assert.Error(t, err)
}

func TestSingleParametersFollowedByListDefaults(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeInteger,
		},
		{
			Name:    "arg2",
			Type:    ParameterTypeIntegerList,
			Default: []int{4, 5, 6},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, []int{2, 3}, v2)
}

func TestThreeSingleParametersFollowedByListDefaults(t *testing.T) {
	args := []string{"1", "2", "3", "4"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeInteger,
		},
		{
			Name: "arg2",
			Type: ParameterTypeInteger,
		},
		{
			Name: "arg3",
			Type: ParameterTypeInteger,
		},
		{
			Name:    "arg4",
			Type:    ParameterTypeIntegerList,
			Default: []int{5, 6, 7},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, 2, v2)
	v3, present := result.Get("arg3")
	assert.True(t, present)
	assert.Equal(t, 3, v3)
	v4, present := result.Get("arg4")
	assert.True(t, present)
	assert.Equal(t, []int{4}, v4)
}

func TestThreeSingleParametersFollowedByListDefaultsOnlyTwoValues(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeInteger,
		},
		{
			Name: "arg2",
			Type: ParameterTypeInteger,
		},
		{
			Name: "arg3",
			Type: ParameterTypeInteger,
		},
		{
			Name:    "arg4",
			Type:    ParameterTypeIntegerList,
			Default: []int{5, 6, 7},
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, 2, v2)
	v3, present := result.Get("arg3")
	assert.True(t, present)
	assert.Equal(t, 3, v3)
	v4, present := result.Get("arg4")
	assert.True(t, present)
	assert.Equal(t, []int{5, 6, 7}, v4)
}

// Test that an argument of type objectListFromFile from test-data/objectList.json correctly parses the argument
func TestObjectListFromFileParsing(t *testing.T) {
	args := []string{"test-data/objectList.json"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeObjectListFromFile,
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []interface{}{
		map[string]interface{}{
			"name": "objectList1",
			"type": "object",
		},
		map[string]interface{}{
			"name": "objectList2",
			"type": "object",
		},
	}, v1)
}

// Test that loading from multiple files with an argument of type objectListFromFiles correctly parses
// objectList.json objectList2.yaml and objectList3.csv
func TestObjectListFromFilesParsing(t *testing.T) {
	args := []string{"test-data/objectList.json", "test-data/objectList2.yaml", "test-data/objectList3.csv"}
	arguments := []*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeObjectListFromFiles,
		},
	}
	result, err := GatherArguments(args, arguments, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t,
		[]interface{}{map[string]interface{}{"name": "objectList1", "type": "object"},
			map[string]interface{}{"name": "objectList2", "type": "object"},
			map[string]interface{}{"name": "objectList3", "type": "object"},
			map[string]interface{}{"name": "objectList4", "type": "object"},
			map[string]interface{}{"name": "objectList5", "type": "object"},
			map[string]interface{}{"name": "objectList6", "type": "object"},
		}, v1)
}
