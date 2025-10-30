package middlewares

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestLayers creates a test parameter layer structure
func createTestLayers(t *testing.T) *layers.ParameterLayers {
	demoLayer, err := layers.NewParameterLayer(
		"demo",
		"Demo Layer",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
			parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger),
			parameters.NewParameterDefinition("timeout", parameters.ParameterTypeInteger),
			parameters.NewParameterDefinition("dev-api-key", parameters.ParameterTypeString),
			parameters.NewParameterDefinition("prod-api-key", parameters.ParameterTypeString),
			parameters.NewParameterDefinition("dev-threshold", parameters.ParameterTypeInteger),
			parameters.NewParameterDefinition("prod-threshold", parameters.ParameterTypeInteger),
			parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString),
		),
	)
	require.NoError(t, err)

	return layers.NewParameterLayers(
		layers.WithLayers(demoLayer),
	)
}

func TestNewConfigMapper_Validation(t *testing.T) {
	tests := []struct {
		name        string
		rules       []MappingRule
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid exact match pattern",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: false,
		},
		{
			name: "valid named capture pattern",
			rules: []MappingRule{
				{
					Source:          "app.{env}.api_key",
					TargetLayer:     "demo",
					TargetParameter: "{env}-api-key",
				},
			},
			expectError: false,
		},
		{
			name: "valid wildcard pattern",
			rules: []MappingRule{
				{
					Source:          "app.*.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: false,
		},
		{
			name: "invalid pattern - empty",
			rules: []MappingRule{
				{
					Source:          "",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: true,
			errorMsg:    "pattern cannot be empty",
		},
		{
			name: "invalid pattern - empty segment",
			rules: []MappingRule{
				{
					Source:          "app..api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: true,
			errorMsg:    "pattern cannot contain empty segments",
		},
		{
			name: "invalid capture - unclosed",
			rules: []MappingRule{
				{
					Source:          "app.{env.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: true,
			errorMsg:    "unclosed capture group",
		},
		{
			name: "invalid capture - empty name",
			rules: []MappingRule{
				{
					Source:          "app.{}.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			expectError: true,
			errorMsg:    "capture group name cannot be empty",
		},
		{
			name: "invalid capture reference - not in source",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "{env}-api-key",
				},
			},
			expectError: true,
			errorMsg:    "capture reference {env} in target parameter not found",
		},
		{
			name: "invalid target layer - does not exist",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "nonexistent",
					TargetParameter: "api-key",
				},
			},
			expectError: true,
			errorMsg:    "target layer \"nonexistent\" does not exist",
		},
		{
			name: "valid nested rules",
			rules: []MappingRule{
				{
					Source:      "app.settings",
					TargetLayer: "demo",
					Rules: []MappingRule{
						{Source: "api_key", TargetParameter: "api-key"},
						{Source: "threshold", TargetParameter: "threshold"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "valid nested rules with capture inheritance",
			rules: []MappingRule{
				{
					Source:      "app.{env}.settings",
					TargetLayer: "demo",
					Rules: []MappingRule{
						{Source: "api_key", TargetParameter: "{env}-api-key"},
						{Source: "threshold", TargetParameter: "{env}-threshold"},
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLayers := createTestLayers(t)
			mapper, err := NewConfigMapper(testLayers, tt.rules...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, mapper)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mapper)
			}
		})
	}
}

func TestPatternMapper_Map(t *testing.T) {
	tests := []struct {
		name         string
		rules        []MappingRule
		config       map[string]interface{}
		expected     map[string]map[string]interface{}
		expectError  bool
		errorMsg     string
	}{
		{
			name: "exact match - simple",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "secret123",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key": "secret123",
				},
			},
			expectError: false,
		},
		{
			name: "exact match - multiple rules",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
				{
					Source:          "app.settings.threshold",
					TargetLayer:     "demo",
					TargetParameter: "threshold",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key":   "secret123",
						"threshold": 42,
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key":   "secret123",
					"threshold": 42,
				},
			},
			expectError: false,
		},
		{
			name: "named capture - single",
			rules: []MappingRule{
				{
					Source:          "app.{env}.api_key",
					TargetLayer:     "demo",
					TargetParameter: "{env}-api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"dev": map[string]interface{}{
						"api_key": "dev-secret",
					},
					"prod": map[string]interface{}{
						"api_key": "prod-secret",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"dev-api-key":  "dev-secret",
					"prod-api-key": "prod-secret",
				},
			},
			expectError: false,
		},
        {
            name: "wildcard - matches all (same values avoids ambiguity)",
            rules: []MappingRule{
                {
                    Source:          "app.*.api_key",
                    TargetLayer:     "demo",
                    TargetParameter: "api-key",
                },
            },
            config: map[string]interface{}{
                "app": map[string]interface{}{
                    "dev": map[string]interface{}{
                        "api_key": "same-secret",
                    },
                    "prod": map[string]interface{}{
                        "api_key": "same-secret",
                    },
                },
            },
            expected: map[string]map[string]interface{}{
                "demo": {
                    "api-key": "same-secret",
                },
            },
            expectError: false,
        },
		{
			name: "nested rules - simple",
			rules: []MappingRule{
				{
					Source:      "app.settings",
					TargetLayer: "demo",
					Rules: []MappingRule{
						{Source: "api_key", TargetParameter: "api-key"},
						{Source: "threshold", TargetParameter: "threshold"},
					},
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key":   "secret123",
						"threshold": 42,
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key":   "secret123",
					"threshold": 42,
				},
			},
			expectError: false,
		},
		{
			name: "nested rules - with capture inheritance",
			rules: []MappingRule{
				{
					Source:      "app.{env}.settings",
					TargetLayer: "demo",
					Rules: []MappingRule{
						{Source: "api_key", TargetParameter: "{env}-api-key"},
						{Source: "threshold", TargetParameter: "{env}-threshold"},
					},
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"dev": map[string]interface{}{
						"settings": map[string]interface{}{
							"api_key":   "dev-secret",
							"threshold": 10,
						},
					},
					"prod": map[string]interface{}{
						"settings": map[string]interface{}{
							"api_key":   "prod-secret",
							"threshold": 100,
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"dev-api-key":      "dev-secret",
					"dev-threshold":    10,
					"prod-api-key":     "prod-secret",
					"prod-threshold":   100,
				},
			},
			expectError: false,
		},
		{
			name: "required pattern - missing",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
					Required:        true,
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"other": map[string]interface{}{
						"value": "something",
					},
				},
			},
			expectError: true,
			errorMsg:    "required pattern",
		},
		{
			name: "optional pattern - missing (no error)",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
					Required:        false,
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"other": map[string]interface{}{
						"value": "something",
					},
				},
			},
			expected:    map[string]map[string]interface{}{},
			expectError: false,
		},
		{
			name: "parameter does not exist in layer",
			rules: []MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "nonexistent-param",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "secret123",
					},
				},
			},
			expectError: true,
			errorMsg:    "target parameter \"nonexistent-param\" does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLayers := createTestLayers(t)
			mapper, err := NewConfigMapper(testLayers, tt.rules...)
			require.NoError(t, err)

			result, err := mapper.Map(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestValidatePatternSyntax(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid exact match",
			pattern:     "app.settings.api_key",
			expectError: false,
		},
		{
			name:        "valid wildcard",
			pattern:     "app.*.api_key",
			expectError: false,
		},
		{
			name:        "valid named capture",
			pattern:     "app.{env}.api_key",
			expectError: false,
		},
		{
			name:        "valid multiple captures",
			pattern:     "app.{region}.{env}.api_key",
			expectError: false,
		},
		{
			name:        "empty pattern",
			pattern:     "",
			expectError: true,
			errorMsg:    "pattern cannot be empty",
		},
		{
			name:        "empty segment",
			pattern:     "app..api_key",
			expectError: true,
			errorMsg:    "pattern cannot contain empty segments",
		},
		{
			name:        "unclosed capture",
			pattern:     "app.{env.api_key",
			expectError: true,
			errorMsg:    "unclosed capture group",
		},
		{
			name:        "empty capture name",
			pattern:     "app.{}.api_key",
			expectError: true,
			errorMsg:    "capture group name cannot be empty",
		},
		{
			name:        "invalid capture name - starts with number",
			pattern:     "app.{1env}.api_key",
			expectError: true,
			errorMsg:    "invalid capture group name",
		},
		{
			name:        "invalid capture name - special chars",
			pattern:     "app.{env-name}.api_key",
			expectError: true,
			errorMsg:    "invalid capture group name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePatternSyntax(tt.pattern)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExtractCaptureNames(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:     "no captures",
			pattern:  "app.settings.api_key",
			expected: nil,
		},
		{
			name:     "single capture",
			pattern:  "app.{env}.api_key",
			expected: []string{"env"},
		},
		{
			name:     "multiple captures",
			pattern:  "app.{region}.{env}.api_key",
			expected: []string{"region", "env"},
		},
		{
			name:     "captures with wildcards",
			pattern:  "app.*.{env}.*.api_key",
			expected: []string{"env"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCaptureNames(tt.pattern)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractCaptureReferences(t *testing.T) {
	tests := []struct {
		name     string
		target   string
		expected map[string]bool
	}{
		{
			name:     "no references",
			target:   "api-key",
			expected: map[string]bool{},
		},
		{
			name:   "single reference",
			target: "{env}-api-key",
			expected: map[string]bool{
				"env": true,
			},
		},
		{
			name:   "multiple references",
			target: "{region}-{env}-api-key",
			expected: map[string]bool{
				"region": true,
				"env":    true,
			},
		},
		{
			name:   "repeated reference",
			target: "{env}-{env}-api-key",
			expected: map[string]bool{
				"env": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractCaptureReferences(tt.target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResolveTargetParameter(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		captures    map[string]string
		expected    string
		expectError bool
	}{
		{
			name:     "no captures",
			target:   "api-key",
			captures: map[string]string{},
			expected: "api-key",
		},
		{
			name:   "single capture",
			target: "{env}-api-key",
			captures: map[string]string{
				"env": "dev",
			},
			expected: "dev-api-key",
		},
		{
			name:   "multiple captures",
			target: "{region}-{env}-api-key",
			captures: map[string]string{
				"region": "us-east",
				"env":    "dev",
			},
			expected: "us-east-dev-api-key",
		},
		{
			name:   "repeated capture",
			target: "{env}-{env}-api-key",
			captures: map[string]string{
				"env": "dev",
			},
			expected: "dev-dev-api-key",
		},
		{
			name:   "missing capture",
			target: "{env}-api-key",
			captures: map[string]string{
				"region": "us-east",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveTargetParameter(tt.target, tt.captures)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestConfigFileMapperAdapter(t *testing.T) {
	// Test that ConfigFileMapper works through the adapter
	testLayers := createTestLayers(t)

	mapperFunc := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
		// Simple flat mapper
		configMap := rawConfig.(map[string]interface{})
		result := map[string]map[string]interface{}{
			"demo": make(map[string]interface{}),
		}
		if apiKey, ok := configMap["api_key"]; ok {
			result["demo"]["api-key"] = apiKey
		}
		return result, nil
	}

	// Use adapter
	adapter := &configFileMapperAdapter{fn: mapperFunc}

	config := map[string]interface{}{
		"api_key": "secret123",
	}

	result, err := adapter.Map(config)
	assert.NoError(t, err)
	assert.Equal(t, map[string]map[string]interface{}{
		"demo": {
			"api-key": "secret123",
		},
	}, result)

	_ = testLayers // Keep compiler happy
}

func TestIntegrationWithLoadParametersFromFile(t *testing.T) {
	// This test validates the integration pattern but doesn't actually load files
	testLayers := createTestLayers(t)

	// Create a pattern mapper
	mapper, err := NewConfigMapper(testLayers, MappingRule{
		Source:          "app.settings.api_key",
		TargetLayer:     "demo",
		TargetParameter: "api-key",
	})
	require.NoError(t, err)

	// Verify it implements ConfigMapper
	var _ ConfigMapper = mapper

	// Test mapping
	config := map[string]interface{}{
		"app": map[string]interface{}{
			"settings": map[string]interface{}{
				"api_key": "test-secret",
			},
		},
	}

	result, err := mapper.Map(config)
	require.NoError(t, err)
	assert.Equal(t, "test-secret", result["demo"]["api-key"])
}

