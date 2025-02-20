package parameters

import (
	"testing"
	"time"

	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsedParametersFromDefaults_BasicTypes(t *testing.T) {
	tests := []struct {
		name                 string
		parameterDefinitions *ParameterDefinitions
		expectedValues       map[string]interface{}
		expectedError        string
	}{
		{
			name: "basic types with defaults",
			parameterDefinitions: NewParameterDefinitions(
				WithParameterDefinitionList([]*ParameterDefinition{
					{
						Name:    "string-param",
						Type:    ParameterTypeString,
						Default: cast.InterfaceAddr("default-string"),
					},
					{
						Name:    "int-param",
						Type:    ParameterTypeInteger,
						Default: cast.InterfaceAddr(42),
					},
					{
						Name:    "bool-param",
						Type:    ParameterTypeBool,
						Default: cast.InterfaceAddr(true),
					},
					{
						Name:    "choice-param",
						Type:    ParameterTypeChoice,
						Default: cast.InterfaceAddr("choice1"),
						Choices: []string{"choice1", "choice2"},
					},
					{
						Name:    "date-param",
						Type:    ParameterTypeDate,
						Default: cast.InterfaceAddr("2024-01-01"),
					},
				}),
			),
			expectedValues: map[string]interface{}{
				"string-param": "default-string",
				"int-param":    42,
				"bool-param":   true,
				"choice-param": "choice1",
				// date will be checked separately due to time.Time comparison
			},
		},
		{
			name: "mixed defaults and no defaults",
			parameterDefinitions: NewParameterDefinitions(
				WithParameterDefinitionList([]*ParameterDefinition{
					{
						Name:    "with-default",
						Type:    ParameterTypeString,
						Default: cast.InterfaceAddr("has-default"),
					},
					{
						Name:    "no-default",
						Type:    ParameterTypeString,
						Default: nil,
					},
				}),
			),
			expectedValues: map[string]interface{}{
				"with-default": "has-default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.parameterDefinitions.ParsedParametersFromDefaults()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check each expected value
			for paramName, expectedValue := range tt.expectedValues {
				param, ok := result.Get(paramName)
				assert.True(t, ok, "parameter %s should exist", paramName)
				assert.Equal(t, expectedValue, param.Value, "parameter %s should have correct value", paramName)
			}

			// Special check for date parameter in the first test case
			if tt.name == "basic types with defaults" {
				dateParam, ok := result.Get("date-param")
				assert.True(t, ok, "date parameter should exist")
				dateTime, ok := dateParam.Value.(time.Time)
				assert.True(t, ok, "date parameter should be time.Time")
				expectedDate, _ := time.ParseInLocation("2006-01-02", "2024-01-01", time.Local)
				assert.Equal(t, expectedDate, dateTime)
			}
		})
	}
}

func TestParsedParametersFromDefaults_EdgeCases(t *testing.T) {
	tests := []struct {
		name                 string
		parameterDefinitions *ParameterDefinitions
		expectedValues       map[string]interface{}
		expectedError        string
	}{
		{
			name:                 "empty parameter definitions",
			parameterDefinitions: NewParameterDefinitions(),
			expectedValues:       map[string]interface{}{},
		},
		{
			name: "all nil defaults",
			parameterDefinitions: NewParameterDefinitions(
				WithParameterDefinitionList([]*ParameterDefinition{
					{
						Name:    "param1",
						Type:    ParameterTypeString,
						Default: nil,
					},
					{
						Name:    "param2",
						Type:    ParameterTypeInteger,
						Default: nil,
					},
				}),
			),
			expectedValues: map[string]interface{}{},
		},
		{
			name: "zero values as defaults",
			parameterDefinitions: NewParameterDefinitions(
				WithParameterDefinitionList([]*ParameterDefinition{
					{
						Name:    "empty-string",
						Type:    ParameterTypeString,
						Default: cast.InterfaceAddr(""),
					},
					{
						Name:    "zero-int",
						Type:    ParameterTypeInteger,
						Default: cast.InterfaceAddr(0),
					},
					{
						Name:    "false-bool",
						Type:    ParameterTypeBool,
						Default: cast.InterfaceAddr(false),
					},
				}),
			),
			expectedValues: map[string]interface{}{
				"empty-string": "",
				"zero-int":     0,
				"false-bool":   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.parameterDefinitions.ParsedParametersFromDefaults()
			if tt.expectedError != "" {
				assert.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, result)

			// Check that we have exactly the expected number of parameters
			assert.Equal(t, len(tt.expectedValues), result.Len(),
				"number of parameters should match expected")

			// Check each expected value
			for paramName, expectedValue := range tt.expectedValues {
				param, ok := result.Get(paramName)
				assert.True(t, ok, "parameter %s should exist", paramName)
				assert.Equal(t, expectedValue, param.Value, "parameter %s should have correct value", paramName)
			}
		})
	}
}

func TestParsedParametersFromDefaults_ListTypes(t *testing.T) {
	pd := NewParameterDefinitions(
		WithParameterDefinitionList([]*ParameterDefinition{
			{
				Name:    "string-list",
				Type:    ParameterTypeStringList,
				Default: cast.InterfaceAddr([]string{"one", "two", "three"}),
			},
			{
				Name:    "integer-list",
				Type:    ParameterTypeIntegerList,
				Default: cast.InterfaceAddr([]int{1, 2, 3}),
			},
			{
				Name:    "choice-list",
				Type:    ParameterTypeChoiceList,
				Default: cast.InterfaceAddr([]string{"choice1", "choice2"}),
				Choices: []string{"choice1", "choice2", "choice3"},
			},
			{
				Name:    "float-list",
				Type:    ParameterTypeFloatList,
				Default: cast.InterfaceAddr([]float64{1.1, 2.2, 3.3}),
			},
		}),
	)

	result, err := pd.ParsedParametersFromDefaults()
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check string list
	param, ok := result.Get("string-list")
	assert.True(t, ok, "string-list parameter should exist")
	assert.Equal(t, []string{"one", "two", "three"}, param.Value)

	// Check integer list
	param, ok = result.Get("integer-list")
	assert.True(t, ok, "integer-list parameter should exist")
	assert.Equal(t, []int{1, 2, 3}, param.Value)

	// Check choice list
	param, ok = result.Get("choice-list")
	assert.True(t, ok, "choice-list parameter should exist")
	assert.Equal(t, []string{"choice1", "choice2"}, param.Value)

	// Check float list
	param, ok = result.Get("float-list")
	assert.True(t, ok, "float-list parameter should exist")
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, param.Value)
}

func TestParsedParametersFromDefaults_MapTypes(t *testing.T) {
	pd := NewParameterDefinitions(
		WithParameterDefinitionList([]*ParameterDefinition{
			{
				Name:    "key-value",
				Type:    ParameterTypeKeyValue,
				Default: cast.InterfaceAddr(map[string]string{"key1": "value1", "key2": "value2"}),
			},
			{
				Name:    "object-from-file",
				Type:    ParameterTypeObjectFromFile,
				Default: cast.InterfaceAddr(map[string]interface{}{"name": "test", "value": 42}),
			},
		}),
	)

	result, err := pd.ParsedParametersFromDefaults()
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check key-value map
	param, ok := result.Get("key-value")
	assert.True(t, ok, "key-value parameter should exist")
	assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, param.Value)

	// Check object map
	param, ok = result.Get("object-from-file")
	assert.True(t, ok, "object-from-file parameter should exist")
	assert.Equal(t, map[string]interface{}{"name": "test", "value": 42}, param.Value)
}

func TestParsedParametersFromDefaults_FileLoadingTypes(t *testing.T) {
	pd := NewParameterDefinitions(
		WithParameterDefinitionList([]*ParameterDefinition{
			{
				Name:    "string-list-from-file",
				Type:    ParameterTypeStringListFromFile,
				Default: cast.InterfaceAddr([]string{"file1", "file2"}),
			},
			{
				Name: "object-list-from-file",
				Type: ParameterTypeObjectListFromFile,
				Default: cast.InterfaceAddr([]map[string]interface{}{
					{"name": "obj1", "value": 1},
					{"name": "obj2", "value": 2},
				}),
			},
		}),
	)

	result, err := pd.ParsedParametersFromDefaults()
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check string list from file
	param, ok := result.Get("string-list-from-file")
	assert.True(t, ok, "string-list-from-file parameter should exist")
	assert.Equal(t, []string{"file1", "file2"}, param.Value)

	// Check object list from file
	param, ok = result.Get("object-list-from-file")
	assert.True(t, ok, "object-list-from-file parameter should exist")
	assert.Equal(t, []map[string]interface{}{
		{"name": "obj1", "value": 1},
		{"name": "obj2", "value": 2},
	}, param.Value)
}

func TestParsedParametersFromDefaults_EmptyCollections(t *testing.T) {
	pd := NewParameterDefinitions(
		WithParameterDefinitionList([]*ParameterDefinition{
			{
				Name:    "empty-string-list",
				Type:    ParameterTypeStringList,
				Default: cast.InterfaceAddr([]string{}),
			},
			{
				Name:    "empty-key-value",
				Type:    ParameterTypeKeyValue,
				Default: cast.InterfaceAddr(map[string]string{}),
			},
			{
				Name:    "empty-object",
				Type:    ParameterTypeObjectFromFile,
				Default: cast.InterfaceAddr(map[string]interface{}{}),
			},
		}),
	)

	result, err := pd.ParsedParametersFromDefaults()
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check empty string list
	param, ok := result.Get("empty-string-list")
	assert.True(t, ok, "empty-string-list parameter should exist")
	assert.Equal(t, []string{}, param.Value)

	// Check empty key-value map
	param, ok = result.Get("empty-key-value")
	assert.True(t, ok, "empty-key-value parameter should exist")
	assert.Equal(t, map[string]string{}, param.Value)

	// Check empty object
	param, ok = result.Get("empty-object")
	assert.True(t, ok, "empty-object parameter should exist")
	assert.Equal(t, map[string]interface{}{}, param.Value)
}

func TestParsedParametersFromDefaults_NilComplexTypes(t *testing.T) {
	pd := NewParameterDefinitions(
		WithParameterDefinitionList([]*ParameterDefinition{
			{
				Name:    "nil-string-list",
				Type:    ParameterTypeStringList,
				Default: nil,
			},
			{
				Name:    "nil-integer-list",
				Type:    ParameterTypeIntegerList,
				Default: nil,
			},
			{
				Name:    "nil-float-list",
				Type:    ParameterTypeFloatList,
				Default: nil,
			},
			{
				Name:    "nil-choice-list",
				Type:    ParameterTypeChoiceList,
				Default: nil,
				Choices: []string{"choice1", "choice2"},
			},
			{
				Name:    "nil-key-value",
				Type:    ParameterTypeKeyValue,
				Default: nil,
			},
			{
				Name:    "nil-object-from-file",
				Type:    ParameterTypeObjectFromFile,
				Default: nil,
			},
			{
				Name:    "nil-object-list-from-file",
				Type:    ParameterTypeObjectListFromFile,
				Default: nil,
			},
			{
				Name:    "nil-string-list-from-file",
				Type:    ParameterTypeStringListFromFile,
				Default: nil,
			},
			{
				Name:    "nil-string-from-file",
				Type:    ParameterTypeStringFromFile,
				Default: nil,
			},
			{
				Name:    "nil-file",
				Type:    ParameterTypeFile,
				Default: nil,
			},
			{
				Name:    "nil-file-list",
				Type:    ParameterTypeFileList,
				Default: nil,
			},
		}),
	)

	result, err := pd.ParsedParametersFromDefaults()
	require.NoError(t, err)
	require.NotNil(t, result)

	// All parameters should be excluded since they have nil defaults
	assert.Equal(t, 0, result.Len(), "no parameters should be included")

	// Verify each parameter is not present
	paramNames := []string{
		"nil-string-list", "nil-integer-list", "nil-float-list",
		"nil-choice-list", "nil-key-value", "nil-object-from-file",
		"nil-object-list-from-file", "nil-string-list-from-file",
		"nil-string-from-file", "nil-file", "nil-file-list",
	}

	for _, name := range paramNames {
		_, ok := result.Get(name)
		assert.False(t, ok, "parameter %s should not exist", name)
	}
}
