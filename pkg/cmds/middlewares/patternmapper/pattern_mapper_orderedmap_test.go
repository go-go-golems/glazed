package patternmapper_test

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeterministicWildcardOrder_SortedKeys verifies that wildcard traversal
// over map keys is deterministic due to ordered map traversal, and when values
// are identical across matches, the resulting value is stable.
func TestDeterministicWildcardOrder_SortedKeys(t *testing.T) {
	// Create a simple layer
	layer, err := schema.NewSection(
		"demo",
		"Demo Layer",
		schema.WithFields(
			fields.New("api-key", fields.TypeString),
		),
	)
	require.NoError(t, err)

	pls := schema.NewSchema(schema.WithSections(layer))

	// Rule uses wildcard: multiple environment keys under app
	rules := []pm.MappingRule{
		{
			Source:          "app.*.api_key",
			TargetLayer:     "demo",
			TargetParameter: "api-key",
		},
	}

	mapper, err := pm.NewConfigMapper(pls, rules...)
	require.NoError(t, err)

	// Config contains two environments; key order in Go map is nondeterministic
	// but mapper converts to ordered map sorted by key, so lexicographic order applies.
	// With identical values across matches, mapping is unambiguous and stable.
	config := map[string]interface{}{
		"app": map[string]interface{}{
			"prod": map[string]interface{}{
				"api_key": "same-secret",
			},
			"dev": map[string]interface{}{
				"api_key": "same-secret",
			},
		},
	}

	got, err := mapper.Map(config)
	require.NoError(t, err)

	assert.Equal(t, "same-secret", got["demo"]["api-key"]) // stable value with identical matches
}
