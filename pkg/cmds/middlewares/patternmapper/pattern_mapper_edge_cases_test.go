package patternmapper_test

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEdgeCases tests various edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupLayers func(t *testing.T) *schema.Schema
		rules       []pm.MappingRule
		config      map[string]interface{}
		expected    map[string]map[string]interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty config",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config:      map[string]interface{}{},
			expected:    map[string]map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "config with nil values",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": nil,
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key": nil,
				},
			},
			expectError: false,
		},
		{
			name:        "deeply nested config",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "a.b.c.d.e.f.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": map[string]interface{}{
							"d": map[string]interface{}{
								"e": map[string]interface{}{
									"f": map[string]interface{}{
										"api_key": "deep-secret",
									},
								},
							},
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key": "deep-secret",
				},
			},
			expectError: false,
		},
		{
			name:        "special characters in config keys",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key":             "secret",
						"key-with-dash":       "value",
						"key_with_underscore": "value",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key": "secret",
				},
			},
			expectError: false,
		},
		{
			name:        "numeric values",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.threshold",
					TargetLayer:     "demo",
					TargetParameter: "threshold",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"threshold": 12345,
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"threshold": 12345,
				},
			},
			expectError: false,
		},
		{
			name: "boolean values",
			setupLayers: func(t *testing.T) *schema.Schema {
				layer, err := schema.NewSection(
					"demo",
					"Demo Layer",
					schema.WithFields(
						fields.New("enabled", fields.TypeBool),
					),
				)
				require.NoError(t, err)
				return schema.NewSchema(schema.WithSections(layer))
			},
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.enabled",
					TargetLayer:     "demo",
					TargetParameter: "enabled",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"enabled": true,
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"enabled": true,
				},
			},
			expectError: false,
		},
		{
			name: "capture with special characters in matched value",
			setupLayers: func(t *testing.T) *schema.Schema {
				layer, err := schema.NewSection(
					"demo",
					"Demo Layer",
					schema.WithFields(
						fields.New("dev-us-east-1-api-key", fields.TypeString),
					),
				)
				require.NoError(t, err)
				return schema.NewSchema(schema.WithSections(layer))
			},
			rules: []pm.MappingRule{
				{
					Source:          "app.{env}.api_key",
					TargetLayer:     "demo",
					TargetParameter: "{env}-api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"dev-us-east-1": map[string]interface{}{
						"api_key": "secret",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"dev-us-east-1-api-key": "secret",
				},
			},
			expectError: false,
		},
		{
			name:        "multiple wildcards in path (same values avoids ambiguity)",
			setupLayers: createTestLayers,
			rules: []pm.MappingRule{
				{
					Source:          "app.*.settings.*.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"env1": map[string]interface{}{
						"settings": map[string]interface{}{
							"region1": map[string]interface{}{
								"api_key": "secret",
							},
							"region2": map[string]interface{}{
								"api_key": "secret",
							},
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"api-key": "secret",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testLayers := tt.setupLayers(t)
			mapper, err := pm.NewConfigMapper(testLayers, tt.rules...)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, mapper)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}
			require.NoError(t, err)
			result, err := mapper.Map(tt.config)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestErrorMessages verifies that error messages are clear and helpful
func TestErrorMessages(t *testing.T) {
	testLayers := createTestLayers(t)

	tests := []struct {
		name          string
		rules         []pm.MappingRule
		config        map[string]interface{}
		expectError   bool
		errorContains []string // Multiple strings that should be in the error
	}{
		{
			name: "required pattern not found - clear message",
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
					Required:        true,
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"other": "value",
				},
			},
			expectError: true,
			errorContains: []string{
				"required pattern",
				"app.settings.api_key",
				"did not match",
				"nearest existing path",
				"missing segment",
			},
		},
		{
			name: "parameter does not exist - shows pattern and layer",
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "nonexistent-param",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "secret",
					},
				},
			},
			expectError: true,
			errorContains: []string{
				"target parameter",
				"nonexistent-param",
				"does not exist",
				"demo",
				"app.settings.api_key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper, err := pm.NewConfigMapper(testLayers, tt.rules...)
			if err != nil {
				// Validation error
				for _, substr := range tt.errorContains {
					assert.Contains(t, err.Error(), substr, "Expected error to contain %q", substr)
				}
				return
			}

			_, err = mapper.Map(tt.config)

			if tt.expectError {
				require.Error(t, err)
				for _, substr := range tt.errorContains {
					assert.Contains(t, err.Error(), substr, "Expected error to contain %q", substr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLayerPrefix tests handling of layer prefixes
func TestLayerPrefix(t *testing.T) {
	// Create a layer with a prefix
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New("demo-api-key", fields.TypeString),
			fields.New("demo-threshold", fields.TypeInteger),
		),
	)
	require.NoError(t, err)

	testLayers := schema.NewSchema(schema.WithSections(layer))

	tests := []struct {
		name        string
		rules       []pm.MappingRule
		config      map[string]interface{}
		expected    map[string]map[string]interface{}
		expectError bool
	}{
		{
			name: "parameter name without prefix - should add prefix",
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "secret",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"demo-api-key": "secret",
				},
			},
			expectError: false,
		},
		{
			name: "parameter name with prefix - should not double prefix",
			rules: []pm.MappingRule{
				{
					Source:          "app.settings.api_key",
					TargetLayer:     "demo",
					TargetParameter: "demo-api-key",
				},
			},
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"settings": map[string]interface{}{
						"api_key": "secret",
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"demo-api-key": "secret",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper, err := pm.NewConfigMapper(testLayers, tt.rules...)
			require.NoError(t, err)

			result, err := mapper.Map(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestComplexCaptureScenarios tests complex capture patterns
func TestComplexCaptureScenarios(t *testing.T) {
	// Create test layers with environment-region parameters
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("us-east-dev-api-key", fields.TypeString),
			fields.New("us-west-prod-api-key", fields.TypeString),
			fields.New("eu-central-dev-api-key", fields.TypeString),
		),
	)
	require.NoError(t, err)

	testLayers := schema.NewSchema(schema.WithSections(layer))

	tests := []struct {
		name        string
		rules       []pm.MappingRule
		config      map[string]interface{}
		expected    map[string]map[string]interface{}
		expectError bool
	}{
		{
			name: "multiple captures in single pattern",
			rules: []pm.MappingRule{
				{
					Source:          "regions.{region}.{env}.api_key",
					TargetLayer:     "demo",
					TargetParameter: "{region}-{env}-api-key",
				},
			},
			config: map[string]interface{}{
				"regions": map[string]interface{}{
					"us-east": map[string]interface{}{
						"dev": map[string]interface{}{
							"api_key": "secret1",
						},
					},
					"us-west": map[string]interface{}{
						"prod": map[string]interface{}{
							"api_key": "secret2",
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"us-east-dev-api-key":  "secret1",
					"us-west-prod-api-key": "secret2",
				},
			},
			expectError: false,
		},
		{
			name: "nested rules with multiple captures from parent",
			rules: []pm.MappingRule{
				{
					Source:      "regions.{region}.environments.{env}.settings",
					TargetLayer: "demo",
					Rules: []pm.MappingRule{
						{Source: "api_key", TargetParameter: "{region}-{env}-api-key"},
					},
				},
			},
			config: map[string]interface{}{
				"regions": map[string]interface{}{
					"eu-central": map[string]interface{}{
						"environments": map[string]interface{}{
							"dev": map[string]interface{}{
								"settings": map[string]interface{}{
									"api_key": "eu-secret",
								},
							},
						},
					},
				},
			},
			expected: map[string]map[string]interface{}{
				"demo": {
					"eu-central-dev-api-key": "eu-secret",
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper, err := pm.NewConfigMapper(testLayers, tt.rules...)
			require.NoError(t, err)

			result, err := mapper.Map(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestConfigTypes tests handling of different value types
func TestConfigTypes(t *testing.T) {
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("string-param", fields.TypeString),
			fields.New("int-param", fields.TypeInteger),
			fields.New("float-param", fields.TypeFloat),
			fields.New("bool-param", fields.TypeBool),
			fields.New("list-param", fields.TypeStringList),
		),
	)
	require.NoError(t, err)

	testLayers := schema.NewSchema(schema.WithSections(layer))

	mapper, err := pm.NewConfigMapper(testLayers,
		pm.MappingRule{
			Source:          "config.string_val",
			TargetLayer:     "demo",
			TargetParameter: "string-param",
		},
		pm.MappingRule{
			Source:          "config.int_val",
			TargetLayer:     "demo",
			TargetParameter: "int-param",
		},
		pm.MappingRule{
			Source:          "config.float_val",
			TargetLayer:     "demo",
			TargetParameter: "float-param",
		},
		pm.MappingRule{
			Source:          "config.bool_val",
			TargetLayer:     "demo",
			TargetParameter: "bool-param",
		},
		pm.MappingRule{
			Source:          "config.list_val",
			TargetLayer:     "demo",
			TargetParameter: "list-param",
		},
	)
	require.NoError(t, err)

	config := map[string]interface{}{
		"config": map[string]interface{}{
			"string_val": "hello",
			"int_val":    42,
			"float_val":  3.14,
			"bool_val":   true,
			"list_val":   []interface{}{"a", "b", "c"},
		},
	}

	result, err := mapper.Map(config)
	require.NoError(t, err)

	assert.Equal(t, "hello", result["demo"]["string-param"])
	assert.Equal(t, 42, result["demo"]["int-param"])
	assert.Equal(t, 3.14, result["demo"]["float-param"])
	assert.Equal(t, true, result["demo"]["bool-param"])
	assert.Equal(t, []interface{}{"a", "b", "c"}, result["demo"]["list-param"])
}
