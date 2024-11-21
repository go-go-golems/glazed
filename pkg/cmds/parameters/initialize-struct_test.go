package parameters_test

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters" // Replace with the actual package path
	"github.com/stretchr/testify/require"
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

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.NoError(t, err)
	assert.Equal(t, "John Doe", testStruct.Name)
	assert.Equal(t, 30, testStruct.Age)
}

// TestInitializeStructWithNilInput tests initializing a struct with a nil input
func TestInitializeStructWithNilInput(t *testing.T) {
	parsedParams := &parameters.ParsedParameters{}

	err := parsedParams.InitializeStruct(nil)

	assert.Error(t, err)
	assert.Equal(t, "Can't initialize nil struct", err.Error())
}

// TestInitializeStructWithNonPointerInput tests initializing a struct with a non-pointer input
func TestInitializeStructWithNonPointerInput(t *testing.T) {
	parsedParams := &parameters.ParsedParameters{}

	testStruct := TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Equal(t, "s is not a pointer", err.Error())
}

// TestInitializeStructWithNonStructPointer tests initializing a struct with a pointer to a non-struct type
func TestInitializeStructWithNonStructPointer(t *testing.T) {
	parsedParams := &parameters.ParsedParameters{}

	nonStruct := "I am not a struct"
	err := parsedParams.InitializeStruct(&nonStruct)

	assert.Error(t, err)
	assert.Equal(t, "s is not a pointer to a struct", err.Error())
}

// TestInitializeStructWithMissingParameters tests initializing a struct with missing parameters
func TestInitializeStructWithMissingParameters(t *testing.T) {
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString),
			"name",
			"John Doe"),
	)

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.NoError(t, err)
	assert.Equal(t, "John Doe", testStruct.Name)
	assert.Equal(t, 0, testStruct.Age) // Age should be the zero value since it's missing from parsedParams
}

// TestInitializeStructWithJSONTagOnNonPointerField tests initializing a struct with a `from_json` tag on a non-pointer field
func TestInitializeStructWithJSONTagOnNonPointerField(t *testing.T) {
	type TestStructWithJSONTag struct {
		Config string `glazed.parameter:"config,from_json"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			`{"key": "value"}`),
	)

	testStruct := &TestStructWithJSONTag{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from_json tag can only be used on pointer fields")
}

// TestParseStringListSuccessfully tests parsing a string list successfully
func TestParseStringListSuccessfully(t *testing.T) {
	type TestStruct struct {
		Tags []string `glazed.parameter:"tags"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList),
			"tags",
			[]string{"go", "testing", "glazed"}),
	)

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.NoError(t, err)
	assert.Equal(t, []string{"go", "testing", "glazed"}, testStruct.Tags)
}

// TestParseIntListSuccessfully tests parsing an integer list successfully
func TestParseIntListSuccessfully(t *testing.T) {
	type TestStruct struct {
		Numbers []int `glazed.parameter:"numbers"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"numbers",
				parameters.ParameterTypeIntegerList),
			"numbers",
			[]int{1, 2, 3, 4, 5}),
	)

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, testStruct.Numbers)
}

// TestParseObjectFromFileSuccessfully tests parsing an object from file successfully
func TestParseObjectFromFileSuccessfully(t *testing.T) {
	type TestStruct struct {
		Config map[string]interface{} `glazed.parameter:"config"`
	}

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

	testStruct := &TestStruct{}

	err := parsedParams.InitializeStruct(testStruct)

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
	assert.Contains(t, err.Error(), "failed to unmarshal JSON")
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

// TestStruct for wildcard functionality
type TestStructWithWildcard struct {
	ApiKeys map[string]string `glazed.parameter:"*_api_key"`
}

// TestInitializeStructWithWildcardMultipleMatches tests wildcard pattern matching multiple parameters
func TestInitializeStructWithWildcardMultipleMatches(t *testing.T) {
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"openai_api_key",
				parameters.ParameterTypeString),
			"openai_api_key",
			"openai-secret"),
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"google_api_key",
				parameters.ParameterTypeString),
			"google_api_key",
			"google-secret"),
	)

	testStruct := &TestStructWithWildcard{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"openai_api_key": "openai-secret",
		"google_api_key": "google-secret",
	}, testStruct.ApiKeys)
}

// TestInitializeStructWithWildcardNoMatches tests wildcard pattern matching no parameters
func TestInitializeStructWithWildcardNoMatches(t *testing.T) {
	parsedParams := parameters.NewParsedParameters()

	testStruct := &TestStructWithWildcard{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Empty(t, testStruct.ApiKeys)
}

// TestInitializeStructWithWildcardOnNonMapField tests wildcard pattern used on a non-map field
func TestInitializeStructWithWildcardOnNonMapField(t *testing.T) {
	type TestStructNonMapWildcard struct {
		ApiKeys string `glazed.parameter:"*_api_key"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"openai_api_key",
				parameters.ParameterTypeString),
			"openai_api_key",
			"openai-secret"),
	)

	testStruct := &TestStructNonMapWildcard{}
	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wildcard parameters require a map field")
}

// TestInitializeStructWithWildcardOnMapWithIncorrectValueTypes tests wildcard pattern on a map with incorrect value types
func TestInitializeStructWithWildcardOnMapWithIncorrectValueTypes(t *testing.T) {
	type TestStructMapWildcardIncorrectTypes struct {
		ApiKeys map[string]int `glazed.parameter:"*_api_key"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"openai_api_key",
				parameters.ParameterTypeString),
			"openai_api_key",
			"openai-secret"),
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"google_api_key",
				parameters.ParameterTypeString),
			"google_api_key",
			"google-secret"),
	)

	testStruct := &TestStructMapWildcardIncorrectTypes{}
	err := parsedParams.InitializeStruct(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set wildcard values for ")
}

// TestInitializeStructWithWildcardOnMapWithCorrectValueTypes tests wildcard pattern on a map with correct value types
func TestInitializeStructWithWildcardOnMapWithCorrectValueTypes(t *testing.T) {
	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"openai_api_key",
				parameters.ParameterTypeString),
			"openai_api_key",
			"openai-secret"),
	)

	testStruct := &TestStructWithWildcard{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"openai_api_key": "openai-secret",
	}, testStruct.ApiKeys)
}

// TestInitializeStructWithWildcardComplexPatterns tests wildcard pattern with complex patterns
func TestInitializeStructWithWildcardComplexPatterns(t *testing.T) {
	type TestStructComplexWildcard struct {
		ApiKeys map[string]string `glazed.parameter:"*api_key"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"openai_api_key",
				parameters.ParameterTypeString),
			"openai_api_key",
			"openai-secret"),
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"google_api_key",
				parameters.ParameterTypeString),
			"google_api_key",
			"google-secret"),
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"api_key_unrelated",
				parameters.ParameterTypeString),
			"api_key_unrelated",
			"should-not-match"),
	)

	testStruct := &TestStructComplexWildcard{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"openai_api_key": "openai-secret",
		"google_api_key": "google-secret",
	}, testStruct.ApiKeys)
}

// TestInitializeStructNormalBehavior tests the normal behavior without wildcards
func TestInitializeStructNormalBehavior(t *testing.T) {
	type TestStructNormal struct {
		Name string `glazed.parameter:"name"`
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString),
			"name",
			"John Doe"),
	)

	testStruct := &TestStructNormal{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, "John Doe", testStruct.Name)
}

func TestStructToDataMapWithWildcardMultipleMatches(t *testing.T) {
	testStruct := &TestStructWithWildcard{
		ApiKeys: map[string]string{
			"openai_api_key": "openai-secret",
			"google_api_key": "google-secret",
		},
	}

	dataMap, err := parameters.StructToDataMap(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"openai_api_key": "openai-secret",
		"google_api_key": "google-secret",
	}, dataMap)
}

func TestStructToDataMapWithWildcardNoMatches(t *testing.T) {
	testStruct := &TestStructWithWildcard{
		ApiKeys: map[string]string{},
	}

	dataMap, err := parameters.StructToDataMap(testStruct)

	require.NoError(t, err)
	assert.Empty(t, dataMap)
}

func TestStructToDataMapWithWildcardOnNonMapField(t *testing.T) {
	type TestStructNonMapWildcard struct {
		ApiKeys string `glazed.parameter:"*_api_key"`
	}

	testStruct := &TestStructNonMapWildcard{
		ApiKeys: "invalid",
	}

	_, err := parameters.StructToDataMap(testStruct)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wildcard parameters require a map field")
}

type TestStructWithJSON struct {
	Data map[string]interface{} `glazed.parameter:"data,from_json"`
}

type TestStructWithTags struct {
	Name string `glazed.parameter:"name"`
	Age  int    `glazed.parameter:"age"`
}

func TestStructToDataMapWithJSON(t *testing.T) {
	testStruct := &TestStructWithJSON{
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	dataMap, err := parameters.StructToDataMap(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"data": map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}, dataMap)
}

func TestStructToDataMapWithTags(t *testing.T) {
	testStruct := &TestStructWithTags{
		Name: "John Doe",
		Age:  30,
	}

	dataMap, err := parameters.StructToDataMap(testStruct)

	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}, dataMap)
}

func TestStructToDataMapWithNilInput(t *testing.T) {
	_, err := parameters.StructToDataMap(nil)

	assert.Error(t, err)
	assert.Equal(t, "cannot convert nil struct to data map", err.Error())
}

func TestStructToDataMapWithNonStructInput(t *testing.T) {
	nonStruct := "I am not a struct"
	_, err := parameters.StructToDataMap(nonStruct)

	assert.Error(t, err)
	assert.Equal(t, "input must be a struct or a pointer to a struct", err.Error())
}

// TestInitializeStructWithJSONFromStruct tests initializing a struct field from a struct value
func TestInitializeStructWithJSONFromStruct(t *testing.T) {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	type TestStructWithJSONPtr struct {
		Config *Config `glazed.parameter:"config,from_json"`
	}

	inputConfig := Config{
		Host: "localhost",
		Port: 8080,
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			inputConfig),
	)

	testStruct := &TestStructWithJSONPtr{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	require.NotNil(t, testStruct.Config)
	assert.Equal(t, "localhost", testStruct.Config.Host)
	assert.Equal(t, 8080, testStruct.Config.Port)
}

// TestInitializeStructWithJSONFromMap tests initializing a struct field from a map value
func TestInitializeStructWithJSONFromMap(t *testing.T) {
	type Config struct {
		Settings map[string]interface{} `json:"settings"`
	}

	type TestStructWithJSONPtr struct {
		Config *Config `glazed.parameter:"config,from_json"`
	}

	inputMap := map[string]interface{}{
		"settings": map[string]interface{}{
			"debug": true,
			"rate":  123.45,
		},
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			inputMap),
	)

	testStruct := &TestStructWithJSONPtr{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	require.NotNil(t, testStruct.Config)
	require.NotNil(t, testStruct.Config.Settings)
	assert.Equal(t, true, testStruct.Config.Settings["debug"])
	assert.Equal(t, 123.45, testStruct.Config.Settings["rate"])
}

// TestInitializeStructWithJSONFromUnmarshallable tests initializing a struct field from an unmarshallable value
func TestInitializeStructWithJSONFromUnmarshallable(t *testing.T) {
	type TestStructWithJSONPtr struct {
		Config *struct{} `glazed.parameter:"config,from_json"`
	}

	// Create a channel which cannot be marshaled to JSON
	ch := make(chan int)

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			ch),
	)

	testStruct := &TestStructWithJSONPtr{}
	err := parsedParams.InitializeStruct(testStruct)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal value of type chan int to JSON")
}

// TestInitializeStructWithFileDataToString tests initializing a string field from FileData
func TestInitializeStructWithFileDataToString(t *testing.T) {
	type TestStruct struct {
		Content string `glazed.parameter:"content"`
	}

	fileData := &parameters.FileData{
		Content: "test content",
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"content",
				parameters.ParameterTypeString),
			"content",
			fileData),
	)

	testStruct := &TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, "test content", testStruct.Content)
}

// TestInitializeStructWithFileDataToBytes tests initializing a []byte field from FileData
func TestInitializeStructWithFileDataToBytes(t *testing.T) {
	type TestStruct struct {
		RawContent []byte `glazed.parameter:"raw_content"`
	}

	fileData := &parameters.FileData{
		RawContent: []byte("test content"),
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"raw_content",
				parameters.ParameterTypeString),
			"raw_content",
			fileData),
	)

	testStruct := &TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, []byte("test content"), testStruct.RawContent)
}

// TestInitializeStructWithFileDataToParsedContent tests initializing a struct field from FileData's ParsedContent
func TestInitializeStructWithFileDataToParsedContent(t *testing.T) {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	type TestStruct struct {
		Config Config `glazed.parameter:"config"`
	}

	parsedConfig := Config{
		Host: "localhost",
		Port: 8080,
	}

	fileData := &parameters.FileData{
		ParsedContent: parsedConfig,
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"config",
				parameters.ParameterTypeString),
			"config",
			fileData),
	)

	testStruct := &TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	require.NoError(t, err)
	assert.Equal(t, "localhost", testStruct.Config.Host)
	assert.Equal(t, 8080, testStruct.Config.Port)
}

// TestInitializeStructWithFileDataToIncompatibleType tests initializing an incompatible field type from FileData
func TestInitializeStructWithFileDataToIncompatibleType(t *testing.T) {
	type TestStruct struct {
		Content []int `glazed.parameter:"content"` // incompatible with FileData.RawContent
	}

	fileData := &parameters.FileData{
		RawContent: []byte("test content"),
	}

	parsedParams := parameters.NewParsedParameters(
		parameters.WithParsedParameter(
			parameters.NewParameterDefinition(
				"content",
				parameters.ParameterTypeString),
			"content",
			fileData),
	)

	testStruct := &TestStruct{}
	err := parsedParams.InitializeStruct(testStruct)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set FileData to slice of type int")
}
