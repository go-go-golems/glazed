package parameters_test

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters" // Replace with the actual package path
	"testing"

	"github.com/stretchr/testify/assert"
)

// Define a struct that matches the expected structure for InitializeStruct
type TestStruct struct {
	Name string `glazed.parameter:"name"`
	Age  int    `glazed.parameter:"age"`
}

// TestInitializeStructWithValidStruct tests initializing a struct with valid parameters
func TestInitializeStructWithValidStruct(t *testing.T) {
	// Create an instance of ParsedParameters with test data
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString),
			"name",
			"John Doe"),
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"age",
				parameters.ParameterTypeInteger),
			"age", 30),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStruct{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that there is no error and the struct is initialized correctly
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", testStruct.Name)
	assert.Equal(t, 30, testStruct.Age)
}

// TestInitializeStructWithNilInput tests initializing a struct with a nil input
func TestInitializeStructWithNilInput(t *testing.T) {
	// Create an instance of ParsedParameters (can be empty for this test)
	parsedParams := &parameters.ParsedParameters{}

	// Call InitializeStruct with a nil pointer
	err := parsedParams.InitializeStruct(nil)

	assert.Error(t, err)
	assert.Equal(t, "Can't initialize nil struct", err.Error())
}

// TestInitializeStructWithNonPointerInput tests initializing a struct with a non-pointer input
func TestInitializeStructWithNonPointerInput(t *testing.T) {
	// Create an instance of ParsedParameters (can be empty for this test)
	parsedParams := &parameters.ParsedParameters{}

	// Call InitializeStruct with a non-pointer value (struct value)
	testStruct := TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that an error is returned and the error message is correct
	assert.Error(t, err)
	assert.Equal(t, "s is not a pointer", err.Error())
}

// TestInitializeStructWithNonStructPointer tests initializing a struct with a pointer to a non-struct type
func TestInitializeStructWithNonStructPointer(t *testing.T) {
	// Create an instance of ParsedParameters (can be empty for this test)
	parsedParams := &parameters.ParsedParameters{}

	// Call InitializeStruct with a pointer to a non-struct type (e.g., string)
	nonStruct := "I am not a struct"
	err := parsedParams.InitializeStruct(&nonStruct)

	// Assert that an error is returned and the error message is correct
	assert.Error(t, err)
	assert.Equal(t, "s is not a pointer to a struct", err.Error())
}

// TestInitializeStructWithMissingParameters tests initializing a struct with missing parameters
func TestInitializeStructWithMissingParameters(t *testing.T) {
	// Create an instance of ParsedParameters with only one parameter defined
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString),
			"name",
			"John Doe"),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStruct{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that there is no error and the struct is partially initialized correctly
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", testStruct.Name)
	assert.Equal(t, 0, testStruct.Age) // Age should be the zero value since it's missing from parsedParams
}

// TestInitializeStructWithJSONTagOnNonPointerField tests initializing a struct with a `from_json` tag on a non-pointer field
func TestInitializeStructWithJSONTagOnNonPointerField(t *testing.T) {
	// Define a struct with a `from_json` tag on a non-pointer field
	type TestStructWithJSONTag struct {
		Config string `glazed.parameter:"config,from_json"`
	}

	// Create an instance of ParsedParameters with a JSON parameter
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			`{"key": "value"}`),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStructWithJSONTag{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that an error is returned and the error message is correct
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from_json tag can only be used on pointer fields")
}

// TestParseStringListSuccessfully tests parsing a string list successfully
func TestParseStringListSuccessfully(t *testing.T) {
	// Define a struct that matches the expected structure for InitializeStruct
	type TestStruct struct {
		Tags []string `glazed.parameter:"tags"`
	}

	// Create an instance of ParsedParameters with test data
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList),
			"tags",
			[]string{"go", "testing", "glazed"}),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStruct{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that there is no error and the struct is initialized correctly
	assert.NoError(t, err)
	assert.Equal(t, []string{"go", "testing", "glazed"}, testStruct.Tags)
}

// TestParseIntListSuccessfully tests parsing an integer list successfully
func TestParseIntListSuccessfully(t *testing.T) {
	// Define a struct that matches the expected structure for InitializeStruct
	type TestStruct struct {
		Numbers []int `glazed.parameter:"numbers"`
	}

	// Create an instance of ParsedParameters with test data
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"numbers",
				parameters.ParameterTypeIntegerList),
			"numbers",
			[]int{1, 2, 3, 4, 5}),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStruct{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that there is no error and the struct is initialized correctly
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, testStruct.Numbers)
}

// TestParseObjectFromFileSuccessfully tests parsing an object from file successfully
func TestParseObjectFromFileSuccessfully(t *testing.T) {
	// Define a struct that matches the expected structure for InitializeStruct
	type TestStruct struct {
		Config map[string]interface{} `glazed.parameter:"config"`
	}

	// Create an instance of ParsedParameters with test data
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeObjectFromFile),
			"config",
			map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			}),
	)

	// Create an instance of the struct to be initialized
	testStruct := &TestStruct{}

	// Call InitializeStruct with the test struct and parsed parameters
	err := parsedParams.InitializeStruct(testStruct)

	// Assert that there is no error and the struct is initialized correctly
	assert.NoError(t, err)
	expectedConfig := map[string]interface{}{
		"key1": "value1",
		"key2": 42, // JSON numbers are parsed as float64 by default
	}
	assert.Equal(t, expectedConfig, testStruct.Config)
}

// TestInitializeStructWithInvalidJSON tests initializing a struct with an invalid JSON string
func TestInitializeStructWithInvalidJSON(t *testing.T) {
	type TestStruct struct {
		Config *map[string]interface{} `glazed.parameter:"config,from_json"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			`{invalidJson: true}`), // Invalid JSON string
	)

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal json for config")
}

// TestInitializeStructWithUnsupportedTypeForJSON tests initializing a struct with an unsupported type for JSON
func TestInitializeStructWithUnsupportedTypeForJSON(t *testing.T) {
	type TestStruct struct {
		Active bool `glazed.parameter:"active,from_json"` // Unsupported type for JSON unmarshaling
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"active",
				parameters.ParameterTypeString),
			"active",
			`true`), // Valid JSON string
	)

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from_json tag can only be used on pointer fields")
}
