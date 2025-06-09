package middlewares

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGatherFlagsFromCustomViper(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `
test-param: "config-value"
numeric-param: 42
bool-param: true
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create parameter definitions
	testParam := &parameters.ParameterDefinition{
		Name: "test-param",
		Type: parameters.ParameterTypeString,
	}
	numericParam := &parameters.ParameterDefinition{
		Name: "numeric-param",
		Type: parameters.ParameterTypeInteger,
	}
	boolParam := &parameters.ParameterDefinition{
		Name: "bool-param",
		Type: parameters.ParameterTypeBool,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("test", "Test layer", layers.WithParameterDefinitions(
		testParam, numericParam, boolParam,
	))
	require.NoError(t, err)

	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("test", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with custom config file
	middleware := GatherFlagsFromCustomViper(
		WithConfigFile(configFile),
		WithParseOptions(parameters.WithParseStepSource("test-config")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	require.NoError(t, err)

	// Check that values were loaded
	parsedLayer, ok := parsedLayers.Get("test")
	require.True(t, ok)
	require.NotNil(t, parsedLayer)

	testParamParsed, ok := parsedLayer.Parameters.Get("test-param")
	require.True(t, ok)
	require.NotNil(t, testParamParsed)
	assert.Equal(t, "config-value", testParamParsed.Value)

	numericParamParsed, ok := parsedLayer.Parameters.Get("numeric-param")
	require.True(t, ok)
	require.NotNil(t, numericParamParsed)
	assert.Equal(t, 42, numericParamParsed.Value)

	boolParamParsed, ok := parsedLayer.Parameters.Get("bool-param")
	require.True(t, ok)
	require.NotNil(t, boolParamParsed)
	assert.Equal(t, true, boolParamParsed.Value)
}

func TestGatherFlagsFromCustomViperWithAppName(t *testing.T) {
	// Create parameter definitions
	testParam := &parameters.ParameterDefinition{
		Name: "test-param",
		Type: parameters.ParameterTypeString,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("test", "Test layer", layers.WithParameterDefinitions(
		testParam,
	))
	require.NoError(t, err)

	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("test", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with app name (should not fail even if config doesn't exist)
	middleware := GatherFlagsFromCustomViper(
		WithAppName("non-existent-app"),
		WithParseOptions(parameters.WithParseStepSource("app-config")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	// Should not error even if config file doesn't exist
	require.NoError(t, err)
}

func TestCustomViperOptions(t *testing.T) {
	config := &CustomViperConfig{}

	// Test WithConfigFile
	WithConfigFile("/test/path")(config)
	assert.Equal(t, "/test/path", config.ConfigFile)

	// Test WithAppName
	WithAppName("test-app")(config)
	assert.Equal(t, "test-app", config.AppName)

	// Test WithParseOptions
	option := parameters.WithParseStepSource("test")
	WithParseOptions(option)(config)
	assert.Len(t, config.ParseOptions, 1)
}
