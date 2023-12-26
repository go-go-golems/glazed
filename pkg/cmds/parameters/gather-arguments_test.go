package parameters

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Test no arguments are passed to the function
func TestGatherArguments_NoArguments(t *testing.T) {
	_, err := NewParameterDefinitions().GatherArguments([]string{}, true, false)
	assert.NoError(t, err)
}

// Test missing required argument
func TestGatherArguments_RequiredMissing(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Required:   true,
			IsArgument: true,
		},
	}))
	_, err := arg.GatherArguments([]string{}, true, false)
	assert.EqualError(t, err, "Argument Test not found")
}

// Test the parsing of every kind of parameter type for provided args
// This should be broken down into individual tests for each parameter types.
// However a generic example of such test might look like:
func TestGatherArguments_ParsingProvidedArguments(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Type:       ParameterTypeString,
			IsArgument: true,
		},
	}))
	res, err := arg.GatherArguments([]string{"value"}, true, false)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, "value", v.Value)
}

// Test parsing of list-type parameter with multiple arguments
func TestGatherArguments_ListParameterParsing(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Type:       ParameterTypeStringList,
			IsArgument: true,
		},
	}))
	res, err := arg.GatherArguments([]string{"value1", "value2"}, true, false)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, []string{"value1", "value2"}, v.Value)
}

func interfaceAddr[T any](v T) *interface{} {
	v_ := interface{}(v)
	return &v_
}

// Test handling of default values when onlyProvided is set to false
func TestGatherArguments_DefaultsWhenProvidedFalse(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Type:       ParameterTypeString,
			Default:    interfaceAddr("default"),
			IsArgument: true,
		},
	}))
	res, err := arg.GatherArguments([]string{}, false, false)
	assert.NoError(t, err)
	v, present := res.Get("Test")
	assert.True(t, present)
	assert.Equal(t, "default", v.Value)
}

// Test handling of default values when onlyProvided is set to true
func TestGatherArguments_NoDefaultsWhenProvidedTrue(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Type:       ParameterTypeString,
			Default:    interfaceAddr("default"),
			IsArgument: true,
		},
	}))
	v, err := arg.GatherArguments([]string{}, true, false)
	assert.NoError(t, err)
	// check that Test is not in v
	_, present := v.Get("Test")
	assert.False(t, present)
}

// Test the error condition of providing too many arguments
func TestGatherArguments_TooManyArguments(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test",
			Type:       ParameterTypeString,
			IsArgument: true,
		},
	}))
	v, err := arg.GatherArguments([]string{"value1", "value2"}, true, false)
	_ = v
	assert.EqualError(t, err, "Too many arguments")
}

// Test the correct sequencing of arguments
func TestGatherArguments_CorrectSequence(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test1",
			Type:       ParameterTypeString,
			IsArgument: true,
		},
		{
			Name:       "Test2",
			Type:       ParameterTypeString,
			IsArgument: true,
		},
	}))
	res, err := arg.GatherArguments([]string{"value1", "value2"}, true, false)
	assert.NoError(t, err)
	v1, present := res.Get("Test1")
	assert.True(t, present)
	assert.Equal(t, "value1", v1.Value)
	v2, present := res.Get("Test2")
	assert.True(t, present)
	assert.Equal(t, "value2", v2.Value)
}

// Test various combinations of list and non-list arguments
func TestGatherArguments_CombinationsListNonList(t *testing.T) {
	arg := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "Test1",
			Type:       ParameterTypeString,
			IsArgument: true,
		},
		{
			Name:       "Test2",
			Type:       ParameterTypeStringList,
			IsArgument: true,
		},
	}))
	res, err := arg.GatherArguments([]string{"value1", "value2", "value3"}, true, false)
	assert.NoError(t, err)
	v1, present := res.Get("Test1")
	assert.True(t, present)
	assert.Equal(t, "value1", v1.Value)
	v2, present := res.Get("Test2")
	assert.True(t, present)
	assert.Equal(t, []string{"value2", "value3"}, v2.Value)
}

func TestListParsingWithDefaults(t *testing.T) {
	args := []string{"data1", "data2"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeStringList,
			Default:    interfaceAddr([]string{"default1", "default2"}),
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"data1", "data2"}, v1.Value)
}

func TestListDefault(t *testing.T) {
	args := []string{}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeStringList,
			Default:    interfaceAddr([]string{"default1", "default2"}),
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"default1", "default2"}, v1.Value)
}

func TestIntegerListParsing(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeIntegerList,
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []int{1, 2, 3}, v2.Value)
}

func TestFloatListParsing(t *testing.T) {
	args := []string{"1.1", "2.2", "3.3"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeFloatList,
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, v2.Value)
}

func TestChoiceListParsing(t *testing.T) {
	args := []string{"choice1", "choice2", "choice3"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name: "arg1",
			Type: ParameterTypeChoiceList,
			Choices: []string{
				"choice1",
				"choice2",
				"choice3",
			},
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v2, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, []string{"choice1", "choice2", "choice3"}, v2.Value)
}

func TestParsingErrorInvalidInt(t *testing.T) {
	args := []string{"1", "2", "3", "notanint"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeIntegerList,
			IsArgument: true,
		},
	}))
	_, err := arguments.GatherArguments(args, false, false)
	assert.Error(t, err)
}

func TestSingleParametersFollowedByListDefaults(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg2",
			Type:       ParameterTypeIntegerList,
			Default:    interfaceAddr([]int{4, 5, 6}),
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1.Value)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, []int{2, 3}, v2.Value)
}

func TestThreeSingleParametersFollowedByListDefaults(t *testing.T) {
	args := []string{"1", "2", "3", "4"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg2",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg3",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg4",
			Type:       ParameterTypeIntegerList,
			Default:    interfaceAddr([]int{5, 6, 7}),
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1.Value)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, 2, v2.Value)
	v3, present := result.Get("arg3")
	assert.True(t, present)
	assert.Equal(t, 3, v3.Value)
	v4, present := result.Get("arg4")
	assert.True(t, present)
	assert.Equal(t, []int{4}, v4.Value)
}

func TestThreeSingleParametersFollowedByListDefaultsOnlyTwoValues(t *testing.T) {
	args := []string{"1", "2", "3"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg2",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg3",
			Type:       ParameterTypeInteger,
			IsArgument: true,
		},
		{
			Name:       "arg4",
			Type:       ParameterTypeIntegerList,
			Default:    interfaceAddr([]int{5, 6, 7}),
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
	assert.NoError(t, err)
	v1, present := result.Get("arg1")
	assert.True(t, present)
	assert.Equal(t, 1, v1.Value)
	v2, present := result.Get("arg2")
	assert.True(t, present)
	assert.Equal(t, 2, v2.Value)
	v3, present := result.Get("arg3")
	assert.True(t, present)
	assert.Equal(t, 3, v3.Value)
	v4, present := result.Get("arg4")
	assert.True(t, present)
	assert.Equal(t, []int{5, 6, 7}, v4.Value)
}

// Test that an argument of type objectListFromFile from test-data/objectList.json correctly parses the argument
func TestObjectListFromFileParsing(t *testing.T) {
	args := []string{"test-data/objectList.json"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeObjectListFromFile,
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
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
	}, v1.Value)
}

// Test that loading from multiple files with an argument of type objectListFromFiles correctly parses
// objectList.json objectList2.yaml and objectList3.csv
func TestObjectListFromFilesParsing(t *testing.T) {
	args := []string{"test-data/objectList.json", "test-data/objectList2.yaml", "test-data/objectList3.csv"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{
			Name:       "arg1",
			Type:       ParameterTypeObjectListFromFiles,
			IsArgument: true,
		},
	}))
	result, err := arguments.GatherArguments(args, false, false)
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
		}, v1.Value)
}

func TestGenerateUseString_NoArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions()
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test", result)
}

func TestGenerateUseString_RequiredArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList(
		[]*ParameterDefinition{{Name: "name", IsArgument: true, Required: true}}))
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test <name>", result)
}

func TestGenerateUseString_OptionalArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList(
		[]*ParameterDefinition{{Name: "name", IsArgument: true}}))
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test [name]", result)
}

func TestGenerateUseString_RequiredAndOptionalArguments(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{Name: "name", Required: true, IsArgument: true},
		{Name: "age", IsArgument: true},
	}))
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test <name> [age]", result)
}

func TestGenerateUseString_WithDefaultValue(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList(
		[]*ParameterDefinition{{Name: "name", Default: interfaceAddr("John"), IsArgument: true}}))
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test [name (default: John)]", result)
}

func TestGenerateUseString_WithMultipleValues(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList(
		[]*ParameterDefinition{{Name: "name", IsArgument: true, Type: ParameterTypeStringList}}))
	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test [name...]", result)
}

func TestGenerateUseString_RequiredWithMultipleValues(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	arguments := NewParameterDefinitions(WithParameterDefinitionList([]*ParameterDefinition{
		{Name: "name", Required: true, IsArgument: true, Type: ParameterTypeStringList},
	}))

	result := GenerateUseString(cmd.Use, arguments)
	require.Equal(t, "test <name...>", result)
}
