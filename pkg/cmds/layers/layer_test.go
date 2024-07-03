package layers

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Helper function to create a parameter layer
func createParameterLayer(t *testing.T, slug, name string, paramDefs ...*parameters.ParameterDefinition) ParameterLayer {
	layer, err := NewParameterLayer(slug, name, WithParameterDefinitions(paramDefs...))
	require.NoError(t, err)
	require.NotNil(t, layer)
	return layer
}

// Helper function to create a parsed layer
func createParsedLayer(t *testing.T, layer ParameterLayer, parsedValues map[string]interface{}) *ParsedLayer {
	options := make([]ParsedLayerOption, 0, len(parsedValues))
	for key, value := range parsedValues {
		options = append(options, WithParsedParameterValue(key, value))
	}
	parsedLayer, err := NewParsedLayer(layer, options...)
	require.NoError(t, err)
	require.NotNil(t, parsedLayer)
	return parsedLayer
}

func TestNewParameterLayers(t *testing.T) {
	layers := NewParameterLayers()
	assert.NotNil(t, layers)
	assert.Equal(t, 0, layers.Len())
}

func TestParameterLayersSubset(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	layer3 := createParameterLayer(t, "layer3", "Layer 3")

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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	count := 0
	layers.ForEach(func(key string, p ParameterLayer) {
		count++
		assert.Contains(t, []string{"layer1", "layer2"}, key)
	})

	assert.Equal(t, 2, count)
}

func TestParameterLayersForEachE(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")

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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")

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
	layer0 := createParameterLayer(t, "layer0", "Layer 0")

	layers := NewParameterLayers(
		WithLayers(layer0),
	)
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")

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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")
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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	layers := NewParameterLayers(WithLayers(layer1, layer2))

	list := layers.AsList()

	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, layer1)
	assert.Contains(t, list, layer2)
}

func TestParameterLayersClone(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
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
	layer1 := createParameterLayer(t, "layer1", "Layer 1",
		parameters.NewParameterDefinition("param1", parameters.ParameterTypeString),
	)
	layer2 := createParameterLayer(t, "layer2", "Layer 2",
		parameters.NewParameterDefinition("param2", parameters.ParameterTypeInteger),
	)

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	allDefs := layers.GetAllParameterDefinitions()

	assert.Equal(t, 2, allDefs.Len())
	val, present := allDefs.Get("param1")
	assert.True(t, present)
	assert.NotNil(t, val)
	val, present = allDefs.Get("param2")
	assert.True(t, present)
	assert.NotNil(t, val)
}

func TestParameterLayersWithLayers(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 2, layers.Len())
	val, present := layers.Get("layer1")
	assert.True(t, present)
	assert.Equal(t, layer1, val)
	val, present = layers.Get("layer2")
	assert.True(t, present)
	assert.Equal(t, layer2, val)
}

func TestParameterLayersWithDuplicateSlugs(t *testing.T) {
	layer1 := createParameterLayer(t, "duplicate", "Layer 1")
	layer2 := createParameterLayer(t, "duplicate", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 1, layers.Len())
	val, present := layers.Get("duplicate")
	assert.True(t, present)
	assert.Equal(t, "Layer 2", val.GetName())
}

func TestParameterLayersSubsetWithNonExistentLayers(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	layers := NewParameterLayers(WithLayers(layer1))

	subset := layers.Subset("layer1", "non_existent")

	assert.Equal(t, 1, subset.Len())
	_, present := subset.Get("layer1")
	assert.True(t, present)
	_, present = subset.Get("non_existent")
	assert.False(t, present)
}

func TestParameterLayersMergeWithOverlappingLayers(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1 - Original")
	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	layers1 := NewParameterLayers(WithLayers(layer1, layer2))

	layer1Duplicate := createParameterLayer(t, "layer1", "Layer 1 - Duplicate")
	layer3 := createParameterLayer(t, "layer3", "Layer 3")
	layers2 := NewParameterLayers(WithLayers(layer1Duplicate, layer3))

	merged := layers1.Merge(layers2)

	assert.Equal(t, 3, merged.Len())
	val, present := merged.Get("layer1")
	assert.True(t, present)
	assert.Equal(t, "Layer 1 - Duplicate", val.GetName())
	_, present = merged.Get("layer2")
	assert.True(t, present)
	_, present = merged.Get("layer3")
	assert.True(t, present)
}

func TestParameterLayersWithLargeNumberOfLayers(t *testing.T) {
	numLayers := 1000
	layers := NewParameterLayers()

	for i := 0; i < numLayers; i++ {
		layer := createParameterLayer(t, fmt.Sprintf("layer%d", i), fmt.Sprintf("Layer %d", i))
		layers.AppendLayers(layer)
	}

	assert.Equal(t, numLayers, layers.Len())
	_, present := layers.Get("layer0")
	assert.True(t, present)
	_, present = layers.Get(fmt.Sprintf("layer%d", numLayers-1))
	assert.True(t, present)
}

func TestParameterLayersWithUnicodeLayerNames(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1 - 你好")
	layer2 := createParameterLayer(t, "layer2", "Layer 2 - こんにちは")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 2, layers.Len())
	val, present := layers.Get("layer1")
	assert.True(t, present)
	assert.Equal(t, "Layer 1 - 你好", val.GetName())
	val, present = layers.Get("layer2")
	assert.True(t, present)
	assert.Equal(t, "Layer 2 - こんにちは", val.GetName())
}
