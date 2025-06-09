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

func TestGatherFlagsFromCustomProfiles(t *testing.T) {
	// Create a temporary profile file
	tempDir := t.TempDir()
	profileFile := filepath.Join(tempDir, "test-profiles.yaml")
	
	profileContent := `
development:
  config:
    host: "dev.example.com"
    port: 3000
    debug: true
production:
  config:
    host: "prod.example.com"
    port: 8080
    debug: false
`
	
	err := os.WriteFile(profileFile, []byte(profileContent), 0644)
	require.NoError(t, err)

	// Create parameter definitions
	hostParam := &parameters.ParameterDefinition{
		Name: "host",
		Type: parameters.ParameterTypeString,
	}
	portParam := &parameters.ParameterDefinition{
		Name: "port",
		Type: parameters.ParameterTypeInteger,
	}
	debugParam := &parameters.ParameterDefinition{
		Name: "debug",
		Type: parameters.ParameterTypeBool,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("config", "Config layer", layers.WithParameterDefinitions(
		hostParam, portParam, debugParam,
	))
	require.NoError(t, err)
	
	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("config", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with custom profile file - development profile
	middleware := GatherFlagsFromCustomProfiles(
		"development",
		WithProfileFile(profileFile),
		WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	require.NoError(t, err)

	// Check that development values were loaded
	parsedLayer, ok := parsedLayers.Get("config")
	require.True(t, ok)
	require.NotNil(t, parsedLayer)

	hostParamParsed, ok := parsedLayer.Parameters.Get("host")
	require.True(t, ok)
	require.NotNil(t, hostParamParsed)
	assert.Equal(t, "dev.example.com", hostParamParsed.Value)

	portParamParsed, ok := parsedLayer.Parameters.Get("port")
	require.True(t, ok)
	require.NotNil(t, portParamParsed)
	assert.Equal(t, 3000, portParamParsed.Value)

	debugParamParsed, ok := parsedLayer.Parameters.Get("debug")
	require.True(t, ok)
	require.NotNil(t, debugParamParsed)
	assert.Equal(t, true, debugParamParsed.Value)
}

func TestGatherFlagsFromCustomProfilesProduction(t *testing.T) {
	// Create a temporary profile file
	tempDir := t.TempDir()
	profileFile := filepath.Join(tempDir, "test-profiles.yaml")
	
	profileContent := `
development:
  config:
    host: "dev.example.com"
    port: 3000
    debug: true
production:
  config:
    host: "prod.example.com"
    port: 8080
    debug: false
`
	
	err := os.WriteFile(profileFile, []byte(profileContent), 0644)
	require.NoError(t, err)

	// Create parameter definitions
	hostParam := &parameters.ParameterDefinition{
		Name: "host",
		Type: parameters.ParameterTypeString,
	}
	portParam := &parameters.ParameterDefinition{
		Name: "port",
		Type: parameters.ParameterTypeInteger,
	}
	debugParam := &parameters.ParameterDefinition{
		Name: "debug",
		Type: parameters.ParameterTypeBool,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("config", "Config layer", layers.WithParameterDefinitions(
		hostParam, portParam, debugParam,
	))
	require.NoError(t, err)
	
	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("config", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with custom profile file - production profile
	middleware := GatherFlagsFromCustomProfiles(
		"production",
		WithProfileFile(profileFile),
		WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	require.NoError(t, err)

	// Check that production values were loaded
	parsedLayer, ok := parsedLayers.Get("config")
	require.True(t, ok)
	require.NotNil(t, parsedLayer)

	hostParamParsed, ok := parsedLayer.Parameters.Get("host")
	require.True(t, ok)
	require.NotNil(t, hostParamParsed)
	assert.Equal(t, "prod.example.com", hostParamParsed.Value)

	portParamParsed, ok := parsedLayer.Parameters.Get("port")
	require.True(t, ok)
	require.NotNil(t, portParamParsed)
	assert.Equal(t, 8080, portParamParsed.Value)

	debugParamParsed, ok := parsedLayer.Parameters.Get("debug")
	require.True(t, ok)
	require.NotNil(t, debugParamParsed)
	assert.Equal(t, false, debugParamParsed.Value)
}

func TestGatherFlagsFromCustomProfilesWithAppName(t *testing.T) {
	// Create parameter definitions
	testParam := &parameters.ParameterDefinition{
		Name: "test-param",
		Type: parameters.ParameterTypeString,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("config", "Config layer", layers.WithParameterDefinitions(
		testParam,
	))
	require.NoError(t, err)
	
	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("config", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with app name (should not fail even if config doesn't exist)
	middleware := GatherFlagsFromCustomProfiles(
		"non-existent-profile",
		WithProfileAppName("non-existent-app"),
		WithProfileParseOptions(parameters.WithParseStepSource("app-profiles")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	// Should not error even if profile file doesn't exist
	require.NoError(t, err)
}

func TestGatherFlagsFromCustomProfilesProfileNotFound(t *testing.T) {
	// Create a temporary profile file without the requested profile
	tempDir := t.TempDir()
	profileFile := filepath.Join(tempDir, "test-profiles.yaml")
	
	profileContent := `
development:
  config:
    host: "dev.example.com"
`
	
	err := os.WriteFile(profileFile, []byte(profileContent), 0644)
	require.NoError(t, err)

	// Create parameter definitions
	testParam := &parameters.ParameterDefinition{
		Name: "test-param",
		Type: parameters.ParameterTypeString,
	}

	// Create layers
	layer, err := layers.NewParameterLayer("config", "Config layer", layers.WithParameterDefinitions(
		testParam,
	))
	require.NoError(t, err)
	
	parameterLayers := layers.NewParameterLayers()
	parameterLayers.Set("config", layer)

	parsedLayers := layers.NewParsedLayers()

	// Test middleware with non-existent profile but required
	middleware := GatherFlagsFromCustomProfiles(
		"staging",
		WithProfileFile(profileFile),
		WithProfileRequired(true),
		WithProfileParseOptions(parameters.WithParseStepSource("custom-profiles")),
	)

	handler := middleware(Identity)
	err = handler(parameterLayers, parsedLayers)
	// Should error because profile is required but not found
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile staging not found")
}

func TestProfileOptions(t *testing.T) {
	config := &ProfileConfig{}
	
	// Test WithProfileFile
	WithProfileFile("/test/path")(config)
	assert.Equal(t, "/test/path", config.ProfileFile)
	
	// Test WithProfileAppName
	WithProfileAppName("test-app")(config)
	assert.Equal(t, "test-app", config.AppName)
	
	// Test WithProfileRequired
	WithProfileRequired(true)(config)
	assert.True(t, config.Required)
	
	// Test WithProfileParseOptions
	option := parameters.WithParseStepSource("test")
	WithProfileParseOptions(option)(config)
	assert.Len(t, config.ParseOptions, 1)
}
