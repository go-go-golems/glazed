package patternmapper_test

import (
	"os"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiMatchPolicy tests proposal 2: Multi-match policy for wildcards
func TestMultiMatchPolicy(t *testing.T) {
	// Create test layers
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("api-key", fields.TypeString),
		),
	)
	require.NoError(t, err)
	testLayers := schema.NewSchema(schema.WithSections(layer))

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

	rules := []pm.MappingRule{
		{
			Source:          "app.*.api_key",
			TargetLayer:     "demo",
			TargetParameter: "api-key",
		},
	}

	t.Run("default policy - should error on multiple distinct values", func(t *testing.T) {
		mapper, err := pm.NewConfigMapper(testLayers, rules...)
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

		mapper, err := pm.NewConfigMapper(testLayers, rules...)
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
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("api-key", fields.TypeString),
		),
	)
	require.NoError(t, err)
	testLayers := schema.NewSchema(schema.WithSections(layer))

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

	rules := []pm.MappingRule{
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
		mapper, err := pm.NewConfigMapper(testLayers, rules...)
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
		layerMulti, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithFields(
				fields.New("api-key", fields.TypeString),
				fields.New("threshold", fields.TypeInteger),
			),
		)
		require.NoError(t, err)
		testLayersMulti := schema.NewSchema(schema.WithSections(layerMulti))

		rulesMulti := []pm.MappingRule{
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

		mapper, err := pm.NewConfigMapper(testLayersMulti, rulesMulti...)
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
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithPrefix("demo-"),
			schema.WithFields(
				fields.New("demo-api-key", fields.TypeString),
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

		rules := []pm.MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key", // Without prefix
			},
		}

		// Build config and map; prefix is added automatically
		config := map[string]interface{}{
			"app": map[string]interface{}{
				"settings": map[string]interface{}{
					"api_key": "secret",
				},
			},
		}
		mapper, err := pm.NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)
		result, err := mapper.Map(config)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "secret", result["demo"]["demo-api-key"])
	})

	t.Run("error message shows both unprefixed and prefixed names (compile-time)", func(t *testing.T) {
		// Create a layer with a prefix
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithPrefix("demo-"),
			schema.WithFields(
				fields.New("demo-threshold", fields.TypeInteger),
				// Note: demo-api-key does NOT exist, so we can test error
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

		rules := []pm.MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "api-key", // This will resolve to demo-api-key which doesn't exist
			},
		}

		// No need to build a config; compile-time validation triggers

		mapper, err := pm.NewConfigMapper(testLayers, rules...)
		assert.Error(t, err)
		assert.Nil(t, mapper)
		// Error should mention both the user-provided name and the checked name
		assert.Contains(t, err.Error(), "api-key")
		assert.Contains(t, err.Error(), "demo-api-key")
		assert.Contains(t, err.Error(), "checked as")
	})

	t.Run("error message for parameter with prefix already included (compile-time)", func(t *testing.T) {
		// Create a layer with a prefix
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithPrefix("demo-"),
			schema.WithFields(
				fields.New("demo-threshold", fields.TypeInteger),
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

		rules := []pm.MappingRule{
			{
				Source:          "app.settings.api_key",
				TargetLayer:     "demo",
				TargetParameter: "demo-api-key", // With prefix already
			},
		}

		mapper, err := pm.NewConfigMapper(testLayers, rules...)
		assert.Error(t, err)
		assert.Nil(t, mapper)
		// Error should only mention demo-api-key once (not duplicated)
		assert.Contains(t, err.Error(), "demo-api-key")
		// Should not have "checked as" since prefix already included
		assert.NotContains(t, err.Error(), "checked as")
	})
}

// TestCombinedScenarios tests combinations of the proposals
func TestCombinedScenarios(t *testing.T) {
	t.Run("capture shadowing warning on nested rules", func(t *testing.T) {
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithFields(
				fields.New("dev-api-key", fields.TypeString),
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

		// Parent captures {env}, child also captures {env} -> shadowing warning
		rules := []pm.MappingRule{
			{
				Source:      "app.{env}",
				TargetLayer: "demo",
				Rules: []pm.MappingRule{
					{Source: "{env}.api_key", TargetParameter: "{env}-api-key"},
				},
			},
		}

		// Capture stderr
		old := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		_, _ = pm.NewConfigMapper(testLayers, rules...)

		_ = w.Close()
		os.Stderr = old
		buf := make([]byte, 2048)
		n, _ := r.Read(buf)
		out := string(buf[:n])
		assert.Contains(t, out, "capture shadowing", "expected shadowing warning")
		assert.Contains(t, out, "{env}")
	})
	t.Run("multi-match with collision detection", func(t *testing.T) {
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithFields(
				fields.New("api-key", fields.TypeString),
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

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

		rules := []pm.MappingRule{
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

		mapper, err := pm.NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		// Should error on multi-match before collision
		result, err := mapper.Map(config)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "matched multiple distinct values")
	})

	t.Run("prefix-aware error with collision", func(t *testing.T) {
		layer, err := schema.NewSection(
			"demo",
			"Demo Layer",
			schema.WithPrefix("demo-"),
			schema.WithFields(
				fields.New("demo-api-key", fields.TypeString),
			),
		)
		require.NoError(t, err)
		testLayers := schema.NewSchema(schema.WithSections(layer))

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

		rules := []pm.MappingRule{
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

		mapper, err := pm.NewConfigMapper(testLayers, rules...)
		require.NoError(t, err)

		result, err := mapper.Map(config)

		assert.Error(t, err)
		assert.Nil(t, result)
		// Error should mention the resolved parameter name (with prefix)
		assert.Contains(t, err.Error(), "demo-api-key")
		assert.Contains(t, err.Error(), "collision")
	})
}
