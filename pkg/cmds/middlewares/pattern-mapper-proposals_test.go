package middlewares

import (
    "testing"

    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestMultiMatchPolicy tests proposal 2: Multi-match policy for wildcards
func TestMultiMatchPolicy(t *testing.T) {
	// Create test layers
	layer, err := layers.NewParameterLayer(
		"demo",
		"Demo Layer",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
		),
	)
	require.NoError(t, err)
	testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

	config := map[string]interface{}{
		"app": map[string]interface{}{
			"dev": map[string]interface{}{
				"api_key": "dev-secret",
			},
			"prod": map[string]interface{}{
				"api_key": "prod-secret",
			},
		},
	}

	rules := []MappingRule{
		{
			Source:          "app.*.api_key",
			TargetLayer:     "demo",
			TargetParameter: "api-key",
		},
	}

    t.Run("default policy - should error on multiple distinct values", func(t *testing.T) {
        mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "matched multiple distinct values")
		assert.Contains(t, err.Error(), "app.*.api_key")
		assert.Contains(t, err.Error(), "api-key")
	})

	t.Run("same values - should not trigger multi-match", func(t *testing.T) {
		configSame := map[string]interface{}{
			"app": map[string]interface{}{
				"dev": map[string]interface{}{
					"api_key": "same-secret",
				},
				"prod": map[string]interface{}{
					"api_key": "same-secret",
				},
			},
		}

        mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(configSame)

		// Should not error because values are the same
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "same-secret", result["demo"]["api-key"])
	})
}

// TestCollisionDetection tests proposal 3: Collision detection across rules
func TestCollisionDetection(t *testing.T) {
	// Create test layers
	layer, err := layers.NewParameterLayer(
		"demo",
		"Demo Layer",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
		),
	)
	require.NoError(t, err)
	testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

	config := map[string]interface{}{
		"app": map[string]interface{}{
			"settings": map[string]interface{}{
				"api_key": "secret1",
			},
		},
		"config": map[string]interface{}{
			"api_key": "secret2",
		},
	}

	rules := []MappingRule{
		{
			Source:          "app.settings.api_key",
			TargetLayer:     "demo",
			TargetParameter: "api-key",
		},
		{
			Source:          "config.api_key",
			TargetLayer:     "demo",
			TargetParameter: "api-key",
		},
	}

    t.Run("default policy - should error on collision", func(t *testing.T) {
        mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "collision")
		assert.Contains(t, err.Error(), "api-key")
		assert.Contains(t, err.Error(), "app.settings.api_key")
		assert.Contains(t, err.Error(), "config.api_key")
	})

	t.Run("no collision - different parameters", func(t *testing.T) {
		layerMulti, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
				parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger),
			),
		)
		require.NoError(t, err)
		testLayersMulti := layers.NewParameterLayers(layers.WithLayers(layerMulti))

		rulesMulti := []MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
			{
				Source:          "config.threshold",
				TargetLayer:     "demo",
				TargetParameter: "threshold",
			},
		}

		configMulti := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret",
				},
			},
			"config": map[string]interface{}{
				"threshold": 42,
			},
		}

        mapper, err := NewConfigMapper(testLayersMulti, rulesMulti...)
		require.NoError(t, err)

		result, err := mapper.Map(configMulti)

		// Should not error because different parameters
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "secret", result["demo"]["api-key"])
		assert.Equal(t, 42, result["demo"]["threshold"])
	})
}

// TestPrefixAwareErrorMessages tests proposal 4: Prefix-aware error messages
func TestPrefixAwareErrorMessages(t *testing.T) {
	t.Run("error message includes prefix-adjusted name", func(t *testing.T) {
		// Create a layer with a prefix
		layer, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithPrefix("demo-"),
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("demo-api-key", parameters.ParameterTypeString),
			),
		)
		require.NoError(t, err)
		testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

		rules := []MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key", // Without prefix
			},
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret",
				},
			},
		}

		mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		// This should succeed because prefix is added automatically
		result, err := mapper.Map(config)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "secret", result["demo"]["demo-api-key"])
	})

	t.Run("error message shows both unprefixed and prefixed names", func(t *testing.T) {
		// Create a layer with a prefix
		layer, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithPrefix("demo-"),
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("demo-threshold", parameters.ParameterTypeInteger),
				// Note: demo-api-key does NOT exist, so we can test error
			),
		)
		require.NoError(t, err)
		testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

		rules := []MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key", // This will resolve to demo-api-key which doesn't exist
			},
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret",
				},
			},
		}

		mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Error should mention both the user-provided name and the checked name
		assert.Contains(t, err.Error(), "api-key")
		assert.Contains(t, err.Error(), "demo-api-key")
		assert.Contains(t, err.Error(), "checked as")
	})

	t.Run("error message for parameter with prefix already included", func(t *testing.T) {
		// Create a layer with a prefix
		layer, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithPrefix("demo-"),
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("demo-threshold", parameters.ParameterTypeInteger),
			),
		)
		require.NoError(t, err)
		testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

		rules := []MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "demo-api-key", // With prefix already
			},
		}

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret",
				},
			},
		}

		mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Error should only mention demo-api-key once (not duplicated)
		assert.Contains(t, err.Error(), "demo-api-key")
		// Should not have "checked as" since prefix already included
		assert.NotContains(t, err.Error(), "checked as")
	})
}

// TestCombinedScenarios tests combinations of the proposals
func TestCombinedScenarios(t *testing.T) {
	t.Run("multi-match with collision detection", func(t *testing.T) {
		layer, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
			),
		)
		require.NoError(t, err)
		testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"dev": map[string]interface{}{
					"api_key": "dev-secret",
				},
				"prod": map[string]interface{}{
					"api_key": "prod-secret",
				},
			},
			"config": map[string]interface{}{
				"api_key": "config-secret",
			},
		}

		rules := []MappingRule{
			{
				Source:          "app.*.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
			{
				Source:          "config.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
		}

        mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		// Should error on multi-match before collision
		result, err := mapper.Map(config)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "matched multiple distinct values")
	})

	t.Run("prefix-aware error with collision", func(t *testing.T) {
		layer, err := layers.NewParameterLayer(
			"demo",
			"Demo Layer",
			layers.WithPrefix("demo-"),
			layers.WithParameterDefinitions(
				parameters.NewParameterDefinition("demo-api-key", parameters.ParameterTypeString),
			),
		)
		require.NoError(t, err)
		testLayers := layers.NewParameterLayers(layers.WithLayers(layer))

		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret1",
				},
			},
			"config": map[string]interface{}{
				"api_key": "secret2",
			},
		}

		rules := []MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
			{
				Source:          "config.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key",
			},
		}

        mapper, err := NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Error should mention the resolved parameter name (with prefix)
		assert.Contains(t, err.Error(), "demo-api-key")
		assert.Contains(t, err.Error(), "collision")
	})
}

