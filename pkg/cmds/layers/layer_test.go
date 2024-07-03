package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewParameterLayers(t *testing.T) {
	layers := NewParameterLayers()
	assert.NotNil(t, layers)
	assert.Equal(t, 0, layers.Len())
}

func TestParameterLayersWithLayers(t *testing.T) {
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	require.NoError(t, err)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	require.NoError(t, err)

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 2, layers.Len())
	val, present := layers.Get("layer1")
	assert.True(t, present)
	assert.Equal(t, layer1, val)
	val, present = layers.Get("layer2")
	assert.True(t, present)
	assert.Equal(t, layer2, val)
}

func TestParameterLayersSubset(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	layer3, _ := NewParameterLayer("layer3", "Layer 3")

	layers := NewParameterLayers(WithLayers(layer1, layer2, layer3))

	subset := layers.Subset("layer1", "layer3")

	assert.Equal(t, 2, subset.Len())
	val, present := subset.Get("layer1")
	assert.NotNil(t, val)
	assert.True(t, present)
	val, present = subset.Get("layer2")
	assert.Nil(t, val)
	assert.False(t, present)
	val, present = subset.Get("layer3")
	assert.NotNil(t, val)
	assert.True(t, present)
}

func TestParameterLayersForEach(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	count := 0
	layers.ForEach(func(key string, p ParameterLayer) {
		count++
		assert.Contains(t, []string{"layer1", "layer2"}, key)
	})

	assert.Equal(t, 2, count)
}

func TestParameterLayersForEachE(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	count := 0
	err := layers.ForEachE(func(key string, p ParameterLayer) error {
		count++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestParameterLayersAppendLayers(t *testing.T) {
	layers := NewParameterLayers()
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")

	layers.AppendLayers(layer1, layer2)

	assert.Equal(t, 2, layers.Len())
	val, present := layers.Get("layer1")
	assert.Equal(t, layer1, val)
	assert.True(t, present)
	val, present = layers.Get("layer2")
	assert.Equal(t, layer2, val)
	assert.True(t, present)
}

func TestParameterLayersPrependLayers(t *testing.T) {
	layer, err := NewParameterLayer("layer0", "Layer 0")
	require.NoError(t, err)

	layers := NewParameterLayers(
		WithLayers(layer),
	)
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")

	layers.PrependLayers(layer1, layer2)

	assert.Equal(t, 3, layers.Len())
	first := layers.Oldest()
	assert.Equal(t, "layer1", first.Key)
	assert.Equal(t, layer1, first.Value)
	second := first.Next()
	assert.Equal(t, "layer2", second.Key)
	assert.Equal(t, layer2, second.Value)
	third := second.Next()
	assert.Equal(t, "layer0", third.Key)
	assert.Equal(t, "Layer 0", third.Value.GetName())
	assert.Nil(t, third.Next())
}

func TestParameterLayersMerge(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	layers1 := NewParameterLayers(WithLayers(layer1))
	layers2 := NewParameterLayers(WithLayers(layer2))

	merged := layers1.Merge(layers2)

	assert.Equal(t, 2, merged.Len())
	val, present := merged.Get("layer1")
	assert.NotNil(t, val)
	assert.True(t, present)
	val, present = merged.Get("layer2")
	assert.True(t, present)
	assert.NotNil(t, val)
}

func TestParameterLayersAsList(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	layers := NewParameterLayers(WithLayers(layer1, layer2))

	list := layers.AsList()

	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, layer1)
	assert.Contains(t, list, layer2)
}

func TestParameterLayersClone(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layers := NewParameterLayers(WithLayers(layer1))

	cloned := layers.Clone()

	assert.Equal(t, layers.Len(), cloned.Len())
	v1, p1 := layers.Get("layer1")
	assert.True(t, p1)
	assert.NotNil(t, v1)
	v2, p2 := cloned.Get("layer1")
	assert.True(t, p2)
	assert.NotNil(t, v2)
	assert.NotSame(t, v1, v2)
	assert.Equal(t, v1.GetSlug(), v2.GetSlug())
}

func TestParameterLayersGetAllParameterDefinitions(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1",
		WithPrefix("l1_"),
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param1", parameters.ParameterTypeString,
			),
		),
	)
	layer2, _ := NewParameterLayer("layer2", "Layer 2",
		WithPrefix("l2_"),
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param2", parameters.ParameterTypeInteger),
		),
	)

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	allDefs := layers.GetAllParameterDefinitions()

	assert.Equal(t, 2, allDefs.Len())
	val, present := allDefs.Get("l1_param1")
	assert.True(t, present)
	assert.NotNil(t, val)
	val, present = allDefs.Get("l2_param2")
	assert.True(t, present)
	assert.NotNil(t, val)
}

// Add more tests for edge cases and other methods as needed
