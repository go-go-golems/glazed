package layers

import (
	"fmt"
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

func TestParameterLayersSubset(t *testing.T) {
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)
	layer3, err := NewParameterLayer("layer3", "Layer 3")
	assert.NoError(t, err)
	assert.NotNil(t, layer3)

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
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	count := 0
	layers.ForEach(func(key string, p ParameterLayer) {
		count++
		assert.Contains(t, []string{"layer1", "layer2"}, key)
	})

	assert.Equal(t, 2, count)
}

func TestParameterLayersForEachE(t *testing.T) {
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	count := 0
	err = layers.ForEachE(func(key string, p ParameterLayer) error {
		count++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestParameterLayersAppendLayers(t *testing.T) {
	layers := NewParameterLayers()
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)

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
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)

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
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)
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
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
	layer2, err := NewParameterLayer("layer2", "Layer 2")
	assert.NoError(t, err)
	assert.NotNil(t, layer2)
	layers := NewParameterLayers(WithLayers(layer1, layer2))

	list := layers.AsList()

	assert.Equal(t, 2, len(list))
	assert.Contains(t, list, layer1)
	assert.Contains(t, list, layer2)
}

func TestParameterLayersClone(t *testing.T) {
	layer1, err := NewParameterLayer("layer1", "Layer 1")
	assert.NoError(t, err)
	assert.NotNil(t, layer1)
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
	layer1, err := NewParameterLayer("layer1", "Layer 1",
		WithPrefix("l1_"),
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param1", parameters.ParameterTypeString,
			),
		),
	)
	require.NoError(t, err)
	layer2, err := NewParameterLayer("layer2", "Layer 2",
		WithPrefix("l2_"),
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param2", parameters.ParameterTypeInteger),
		),
	)
	require.NoError(t, err)

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

func TestParameterLayersWithDuplicateSlugs(t *testing.T) {
	layer1, _ := NewParameterLayer("duplicate", "Layer 1")
	layer2, _ := NewParameterLayer("duplicate", "Layer 2")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 1, layers.Len())
	val, present := layers.Get("duplicate")
	assert.True(t, present)
	assert.Equal(t, "Layer 2", val.GetName())
}

func TestParameterLayersSubsetWithNonExistentLayers(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	layers := NewParameterLayers(WithLayers(layer1))

	subset := layers.Subset("layer1", "non_existent")

	assert.Equal(t, 1, subset.Len())
	_, present := subset.Get("layer1")
	assert.True(t, present)
	_, present = subset.Get("non_existent")
	assert.False(t, present)
}

func TestParameterLayersMergeWithOverlappingLayers(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1 - Original")
	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	layers1 := NewParameterLayers(WithLayers(layer1, layer2))

	layer1Duplicate, _ := NewParameterLayer("layer1", "Layer 1 - Duplicate")
	layer3, _ := NewParameterLayer("layer3", "Layer 3")
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
		layer, _ := NewParameterLayer(fmt.Sprintf("layer%d", i), fmt.Sprintf("Layer %d", i))
		layers.AppendLayers(layer)
	}

	assert.Equal(t, numLayers, layers.Len())
	_, present := layers.Get("layer0")
	assert.True(t, present)
	_, present = layers.Get(fmt.Sprintf("layer%d", numLayers-1))
	assert.True(t, present)
}

func TestParameterLayersWithUnicodeLayerNames(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1 - 你好")
	layer2, _ := NewParameterLayer("layer2", "Layer 2 - こんにちは")

	layers := NewParameterLayers(WithLayers(layer1, layer2))

	assert.Equal(t, 2, layers.Len())
	val, present := layers.Get("layer1")
	assert.True(t, present)
	assert.Equal(t, "Layer 1 - 你好", val.GetName())
	val, present = layers.Get("layer2")
	assert.True(t, present)
	assert.Equal(t, "Layer 2 - こんにちは", val.GetName())
}

func TestNewParsedLayers(t *testing.T) {
	parsedLayers := NewParsedLayers()
	assert.NotNil(t, parsedLayers)
	assert.Equal(t, 0, parsedLayers.Len())
}

func TestParsedLayersWithParsedLayer(t *testing.T) {
	layer, err := NewParameterLayer("test", "Test Layer")
	require.NoError(t, err)
	parsedLayer, err := NewParsedLayer(layer)
	require.NoError(t, err)

	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	assert.Equal(t, 1, parsedLayers.Len())
	val, present := parsedLayers.Get("test")
	assert.True(t, present)
	assert.Equal(t, parsedLayer, val)
}

func TestParsedLayersClone(t *testing.T) {
	layer, _ := NewParameterLayer("test", "Test Layer")
	parsedLayer, _ := NewParsedLayer(layer)
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	cloned := parsedLayers.Clone()

	assert.Equal(t, parsedLayers.Len(), cloned.Len())
	originalVal, _ := parsedLayers.Get("test")
	clonedVal, present := cloned.Get("test")
	assert.True(t, present)
	assert.NotSame(t, originalVal, clonedVal)
	assert.Equal(t, originalVal.Layer, clonedVal.Layer)
}

func TestParsedLayersGetOrCreate(t *testing.T) {
	parsedLayers := NewParsedLayers()
	layer, _ := NewParameterLayer("test", "Test Layer")

	// Get non-existent layer (should create)
	parsedLayer := parsedLayers.GetOrCreate(layer)
	assert.NotNil(t, parsedLayer)
	assert.Equal(t, layer, parsedLayer.Layer)

	// Get existing layer
	sameLayer := parsedLayers.GetOrCreate(layer)
	assert.Equal(t, parsedLayer, sameLayer)
}

func TestParsedLayersGetDataMap(t *testing.T) {
	layer1, err := NewParameterLayer("layer1", "Layer 1",
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param1", parameters.ParameterTypeString)))
	assert.NoError(t, err)
	require.NotNil(t, layer1)
	parsedLayer1, err := NewParsedLayer(layer1, WithParsedParameterValue("param1", "value1"))
	assert.NoError(t, err)
	require.NotNil(t, parsedLayer1)

	layer2, err := NewParameterLayer("layer2", "Layer 2",
		WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"param2", parameters.ParameterTypeInteger)))
	assert.NoError(t, err)
	require.NotNil(t, layer2)
	parsedLayer2, err := NewParsedLayer(layer2, WithParsedParameterValue("param2", 42))
	assert.NoError(t, err)
	require.NotNil(t, parsedLayer2)

	parsedLayers := NewParsedLayers(
		WithParsedLayer("layer1", parsedLayer1),
		WithParsedLayer("layer2", parsedLayer2),
	)

	dataMap := parsedLayers.GetDataMap()
	assert.Equal(t, 2, len(dataMap))
	assert.Equal(t, "value1", dataMap["param1"])
	assert.Equal(t, 42, dataMap["param2"])
}

func TestParsedLayersInitializeStruct(t *testing.T) {
	type TestStruct struct {
		Param1 string `glazed.parameter:"param1"`
		Param2 int    `glazed.parameter:"param2"`
	}

	layer, _ := NewParameterLayer("test", "Test Layer")
	parsedLayer, _ := NewParsedLayer(layer,
		WithParsedParameterValue("param1", "value1"),
		WithParsedParameterValue("param2", 42),
	)
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	var result TestStruct
	err := parsedLayers.InitializeStruct("test", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result.Param1)
	assert.Equal(t, 42, result.Param2)
}

func TestParsedLayersGetAllParsedParameters(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	parsedLayer1, _ := NewParsedLayer(layer1, WithParsedParameterValue("param1", "value1"))

	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	parsedLayer2, _ := NewParsedLayer(layer2, WithParsedParameterValue("param2", 42))

	parsedLayers := NewParsedLayers(
		WithParsedLayer("layer1", parsedLayer1),
		WithParsedLayer("layer2", parsedLayer2),
	)

	allParams := parsedLayers.GetAllParsedParameters()
	assert.Equal(t, 2, allParams.Len())
	param1, present := allParams.Get("param1")
	assert.True(t, present)
	assert.Equal(t, "value1", param1.Value)
	param2, present := allParams.Get("param2")
	assert.True(t, present)
	assert.Equal(t, 42, param2.Value)
}

func TestParsedLayersGetParameter(t *testing.T) {
	layer, _ := NewParameterLayer("test", "Test Layer")
	parsedLayer, _ := NewParsedLayer(layer, WithParsedParameterValue("param", "value"))
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	param, present := parsedLayers.GetParameter("test", "param")
	assert.True(t, present)
	assert.Equal(t, "value", param.Value)

	_, present = parsedLayers.GetParameter("non_existent", "param")
	assert.False(t, present)

	_, present = parsedLayers.GetParameter("test", "non_existent")
	assert.False(t, present)
}

func TestParsedLayersGetDefaultParameterLayer(t *testing.T) {
	parsedLayers := NewParsedLayers()

	defaultLayer := parsedLayers.GetDefaultParameterLayer()
	assert.NotNil(t, defaultLayer)
	assert.Equal(t, DefaultSlug, defaultLayer.Layer.GetSlug())

	// Calling it again should return the same layer
	sameDefaultLayer := parsedLayers.GetDefaultParameterLayer()
	assert.Equal(t, defaultLayer, sameDefaultLayer)
}

func TestParsedLayersForEach(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	parsedLayer1, _ := NewParsedLayer(layer1)

	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	parsedLayer2, _ := NewParsedLayer(layer2)

	parsedLayers := NewParsedLayers(
		WithParsedLayer("layer1", parsedLayer1),
		WithParsedLayer("layer2", parsedLayer2),
	)

	count := 0
	parsedLayers.ForEach(func(k string, v *ParsedLayer) {
		count++
		assert.Contains(t, []string{"layer1", "layer2"}, k)
	})
	assert.Equal(t, 2, count)
}

func TestParsedLayersForEachE(t *testing.T) {
	layer1, _ := NewParameterLayer("layer1", "Layer 1")
	parsedLayer1, _ := NewParsedLayer(layer1)

	layer2, _ := NewParameterLayer("layer2", "Layer 2")
	parsedLayer2, _ := NewParsedLayer(layer2)

	parsedLayers := NewParsedLayers(
		WithParsedLayer("layer1", parsedLayer1),
		WithParsedLayer("layer2", parsedLayer2),
	)

	count := 0
	err := parsedLayers.ForEachE(func(k string, v *ParsedLayer) error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 2, count)

	// Test with error
	errorOnSecond := parsedLayers.ForEachE(func(k string, v *ParsedLayer) error {
		if k == "layer2" {
			return assert.AnError
		}
		return nil
	})
	assert.Error(t, errorOnSecond)
}
