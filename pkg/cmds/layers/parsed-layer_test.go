package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewParsedLayers(t *testing.T) {
	parsedLayers := NewParsedLayers()
	assert.NotNil(t, parsedLayers)
	assert.Equal(t, 0, parsedLayers.Len())
}

func TestParsedLayersWithParsedLayer(t *testing.T) {
	layer := createParameterLayer(t, "test", "Test Layer")
	parsedLayer := createParsedLayer(t, layer, nil)

	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	assert.Equal(t, 1, parsedLayers.Len())
	val, present := parsedLayers.Get("test")
	assert.True(t, present)
	assert.Equal(t, parsedLayer, val)
}

func TestParsedLayersClone(t *testing.T) {
	layer := createParameterLayer(t, "test", "Test Layer")
	parsedLayer := createParsedLayer(t, layer, nil)
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
	layer := createParameterLayer(t, "test", "Test Layer")

	// Get non-existent layer (should create)
	parsedLayer := parsedLayers.GetOrCreate(layer)
	assert.NotNil(t, parsedLayer)
	assert.Equal(t, layer, parsedLayer.Layer)

	// Get existing layer
	sameLayer := parsedLayers.GetOrCreate(layer)
	assert.Equal(t, parsedLayer, sameLayer)
}

func TestParsedLayersGetDataMap(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1",
		parameters.NewParameterDefinition("param1", parameters.ParameterTypeString),
	)
	parsedLayer1 := createParsedLayer(t, layer1, map[string]interface{}{"param1": "value1"})

	layer2 := createParameterLayer(t, "layer2", "Layer 2",
		parameters.NewParameterDefinition("param2", parameters.ParameterTypeInteger),
	)
	parsedLayer2 := createParsedLayer(t, layer2, map[string]interface{}{"param2": 42})

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

	layer := createParameterLayer(t, "test", "Test Layer",
		parameters.NewParameterDefinition("param1", parameters.ParameterTypeString),
		parameters.NewParameterDefinition("param2", parameters.ParameterTypeInteger),
	)
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	})
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	var result TestStruct
	err := parsedLayers.InitializeStruct("test", &result)
	assert.NoError(t, err)
	assert.Equal(t, "value1", result.Param1)
	assert.Equal(t, 42, result.Param2)
}

func TestParsedLayersGetAllParsedParameters(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1",
		parameters.NewParameterDefinition("param1", parameters.ParameterTypeString),
	)
	parsedLayer1 := createParsedLayer(t, layer1, map[string]interface{}{"param1": "value1"})

	layer2 := createParameterLayer(t, "layer2", "Layer 2",
		parameters.NewParameterDefinition("param2", parameters.ParameterTypeInteger),
	)
	parsedLayer2 := createParsedLayer(t, layer2, map[string]interface{}{"param2": 42})

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
	layer := createParameterLayer(t, "test", "Test Layer",
		parameters.NewParameterDefinition("param", parameters.ParameterTypeString),
	)
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{"param": "value"})
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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	parsedLayer1 := createParsedLayer(t, layer1, nil)

	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	parsedLayer2 := createParsedLayer(t, layer2, nil)

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
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	parsedLayer1 := createParsedLayer(t, layer1, nil)

	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	parsedLayer2 := createParsedLayer(t, layer2, nil)

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

func TestParsedLayerInitializeStructWithUnexportedFields(t *testing.T) {
	layer := createParameterLayer(t, "test", "Test Layer",
		parameters.NewParameterDefinition("exported", parameters.ParameterTypeString),
	)
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{"exported": "value"})

	type TestStruct struct {
		Exported   string `glazed.parameter:"exported"`
		unexported string
	}

	result := TestStruct{
		unexported: "unexported",
	}
	err := parsedLayer.InitializeStruct(&result)
	assert.NoError(t, err)
	assert.Equal(t, "value", result.Exported)
	assert.Equal(t, "unexported", result.unexported)

}

func TestParsedLayersInitializeStructWithNonPointer(t *testing.T) {
	parsedLayers := NewParsedLayers()
	layer := createParameterLayer(t, "test", "Test Layer",
		parameters.NewParameterDefinition("param", parameters.ParameterTypeString),
	)
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{"param": "value"})
	parsedLayers.Set("test", parsedLayer)

	type TestStruct struct {
		Param string `glazed.parameter:"param"`
	}

	var result TestStruct
	err := parsedLayers.InitializeStruct("test", result) // Note: passing result, not &result
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer")
}

func TestParsedLayersGetParameterNonExistentLayer(t *testing.T) {
	parsedLayers := NewParsedLayers()

	_, present := parsedLayers.GetParameter("non_existent", "param")
	assert.False(t, present)
}

func TestParsedLayersGetOrCreateNilLayer(t *testing.T) {
	parsedLayers := NewParsedLayers()

	// Depending on how you want to handle this case, you might expect an error or a new empty layer
	// Alternatively, if you want to panic:
	assert.Panics(t, func() { parsedLayers.GetOrCreate(nil) })
}

func TestParsedLayersInitializeStructUnsupportedTypes(t *testing.T) {
	parsedLayers := NewParsedLayers()
	layer := createParameterLayer(t, "test", "Test Layer",
		parameters.NewParameterDefinition("supported", parameters.ParameterTypeString),
		parameters.NewParameterDefinition("unsupported", parameters.ParameterTypeString),
	)
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{
		"supported":   "value",
		"unsupported": make(chan int), // channels are not supported
	})
	parsedLayers.Set("test", parsedLayer)

	type TestStruct struct {
		Supported   string   `glazed.parameter:"supported"`
		Unsupported chan int `glazed.parameter:"unsupported"`
	}

	// as long as the given values are compatible, we let it slide...
	var result TestStruct
	err := parsedLayers.InitializeStruct("test", &result)
	assert.NoError(t, err)
}

func TestParsedLayersForEachEWithError(t *testing.T) {
	layer1 := createParameterLayer(t, "layer1", "Layer 1")
	parsedLayer1 := createParsedLayer(t, layer1, nil)
	layer2 := createParameterLayer(t, "layer2", "Layer 2")
	parsedLayer2 := createParsedLayer(t, layer2, nil)

	parsedLayers := NewParsedLayers(
		WithParsedLayer("layer1", parsedLayer1),
		WithParsedLayer("layer2", parsedLayer2),
	)

	count := 0
	err := parsedLayers.ForEachE(func(k string, v *ParsedLayer) error {
		count++
		if k == "layer2" {
			return errors.New("intentional error")
		}
		return nil
	})

	assert.Error(t, err)
	assert.Equal(t, "intentional error", err.Error())
	assert.Equal(t, 2, count) // The loop should have proceeded to the second layer before stopping
}
