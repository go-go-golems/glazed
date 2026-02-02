package layers

import (
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
		fields.New("param1", fields.TypeString),
	)
	parsedLayer1 := createParsedLayer(t, layer1, map[string]interface{}{"param1": "value1"})

	layer2 := createParameterLayer(t, "layer2", "Layer 2",
		fields.New("param2", fields.TypeInteger),
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
		fields.New("param1", fields.TypeString),
		fields.New("param2", fields.TypeInteger),
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
		fields.New("param1", fields.TypeString),
	)
	parsedLayer1 := createParsedLayer(t, layer1, map[string]interface{}{"param1": "value1"})

	layer2 := createParameterLayer(t, "layer2", "Layer 2",
		fields.New("param2", fields.TypeInteger),
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
		fields.New("param", fields.TypeString),
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
		fields.New("exported", fields.TypeString),
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
		fields.New("param", fields.TypeString),
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
	layer := createParameterLayer(t, "test", "Test Layer",
		fields.New("supported", fields.TypeString),
		fields.New("unsupported", fields.TypeString),
	)
	parsedValues := map[string]interface{}{
		"supported":   "value",
		"unsupported": make(chan int), // channels are not supported
	}
	options := make([]ParsedLayerOption, 0, len(parsedValues))
	for key, value := range parsedValues {
		options = append(options, WithParsedParameterValue(key, value))
	}
	_, err := NewParsedLayer(layer, options...)
	assert.Error(t, err)
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

func TestParsedLayersInitializeStructStringTypes(t *testing.T) {
	// Define custom types
	type StringAlias string
	type StringDeclaration = string

	// Define the test struct
	type TestStruct struct {
		StringField            string              `glazed.parameter:"string_field"`
		StringAliasField       StringAlias         `glazed.parameter:"string_alias_field"`
		StringDeclarationField StringDeclaration   `glazed.parameter:"string_declaration_field"`
		StringListField        []string            `glazed.parameter:"string_list_field"`
		StringAliasListField   []StringAlias       `glazed.parameter:"string_alias_list_field"`
		StringDeclListField    []StringDeclaration `glazed.parameter:"string_decl_list_field"`
	}

	// Create a parameter layer with all the necessary definitions
	layer := createParameterLayer(t, "test", "Test Layer",
		fields.New("string_field", fields.TypeString),
		fields.New("string_alias_field", fields.TypeString),
		fields.New("string_declaration_field", fields.TypeString),
		fields.New("string_list_field", fields.TypeStringList),
		fields.New("string_alias_list_field", fields.TypeStringList),
		fields.New("string_decl_list_field", fields.TypeStringList),
	)

	// Create a parsed layer with test values
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{
		"string_field":             "regular string",
		"string_alias_field":       "aliased string",
		"string_declaration_field": "declared string",
		"string_list_field":        []string{"a", "b", "c"},
		"string_alias_list_field":  []string{"x", "y", "z"},
		"string_decl_list_field":   []string{"1", "2", "3"},
	})

	// Create ParsedLayers and add the parsed layer
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	// Initialize the struct
	var result TestStruct
	err := parsedLayers.InitializeStruct("test", &result)

	// Assert no error occurred
	assert.NoError(t, err)

	// Verify each field was correctly initialized
	assert.Equal(t, "regular string", result.StringField)
	assert.Equal(t, StringAlias("aliased string"), result.StringAliasField)
	assert.Equal(t, StringDeclaration("declared string"), result.StringDeclarationField)
	assert.Equal(t, []string{"a", "b", "c"}, result.StringListField)
	assert.Equal(t, []StringAlias{"x", "y", "z"}, result.StringAliasListField)
	assert.Equal(t, []StringDeclaration{"1", "2", "3"}, result.StringDeclListField)

	// Additional type checks
	assert.IsType(t, "", result.StringField)
	assert.IsType(t, StringAlias(""), result.StringAliasField)
	assert.IsType(t, StringDeclaration(""), result.StringDeclarationField)
	assert.IsType(t, []string{}, result.StringListField)
	assert.IsType(t, []StringAlias{}, result.StringAliasListField)
	assert.IsType(t, []StringDeclaration{}, result.StringDeclListField)
}

func TestParsedLayersInitializeStructStringPointerTypes(t *testing.T) {
	// Define custom types
	type StringAlias string
	type StringDeclaration = string

	// Define the test struct with pointer fields
	type TestStruct struct {
		StringPtrField            *string            `glazed.parameter:"string_ptr_field"`
		StringAliasPtrField       *StringAlias       `glazed.parameter:"string_alias_ptr_field"`
		StringDeclarationPtrField *StringDeclaration `glazed.parameter:"string_declaration_ptr_field"`
	}

	// Create a parameter layer with all the necessary definitions
	layer := createParameterLayer(t, "test", "Test Layer",
		fields.New("string_ptr_field", fields.TypeString),
		fields.New("string_alias_ptr_field", fields.TypeString),
		fields.New("string_declaration_ptr_field", fields.TypeString),
	)

	// Create a parsed layer with test values
	parsedLayer := createParsedLayer(t, layer, map[string]interface{}{
		"string_ptr_field":             "regular string",
		"string_alias_ptr_field":       "aliased string",
		"string_declaration_ptr_field": "declared string",
	})

	// Create ParsedLayers and add the parsed layer
	parsedLayers := NewParsedLayers(WithParsedLayer("test", parsedLayer))

	// Initialize the struct
	var result TestStruct
	err := parsedLayers.InitializeStruct("test", &result)

	// Assert no error occurred
	assert.NoError(t, err)

	// Verify each field was correctly initialized
	assert.NotNil(t, result.StringPtrField)
	assert.NotNil(t, result.StringAliasPtrField)
	assert.NotNil(t, result.StringDeclarationPtrField)

	assert.Equal(t, "regular string", *result.StringPtrField)
	assert.Equal(t, StringAlias("aliased string"), *result.StringAliasPtrField)
	assert.Equal(t, StringDeclaration("declared string"), *result.StringDeclarationPtrField)

	// Additional type checks
	assert.IsType(t, (*string)(nil), result.StringPtrField)
	assert.IsType(t, (*StringAlias)(nil), result.StringAliasPtrField)
	assert.IsType(t, (*StringDeclaration)(nil), result.StringDeclarationPtrField)
}
