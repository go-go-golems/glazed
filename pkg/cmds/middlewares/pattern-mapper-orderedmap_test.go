package middlewares

import (
    "testing"

    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// TestDeterministicWildcardOrder_SortedKeys verifies that wildcard matching
// over map keys is deterministic due to ordered map traversal.
func TestDeterministicWildcardOrder_SortedKeys(t *testing.T) {
    // Create a simple layer
    layer, err := layers.NewParameterLayer(
        "demo",
        "Demo Layer",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
        ),
    )
    require.NoError(t, err)

    pls := layers.NewParameterLayers(layers.WithLayers(layer))

    // Rule uses wildcard: multiple environment keys under app
    rules := []MappingRule{
        {
            Source:          "app.*.api_key",
            TargetLayer:     "demo",
            TargetParameter: "api-key",
        },
    }

    mapper, err := NewConfigMapper(pls, rules...)
    require.NoError(t, err)

    // Config contains two environments; key order in Go map is nondeterministic
    // but mapper converts to ordered map sorted by key, so lexicographic order applies.
    // With keys "dev" and "prod", iteration is ["dev", "prod"], so last wins -> prod.
    config := map[string]interface{}{
        "app": map[string]interface{}{
            "prod": map[string]interface{}{
                "api_key": "prod-secret",
            },
            "dev": map[string]interface{}{
                "api_key": "dev-secret",
            },
        },
    }

    got, err := mapper.Map(config)
    require.NoError(t, err)

    assert.Equal(t, "prod-secret", got["demo"]["api-key"]) // deterministic last-wins with sorted keys
}


