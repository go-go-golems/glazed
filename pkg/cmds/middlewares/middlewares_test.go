package middlewares_test

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

// Simplified structures for testing purposes
type parameterDefinition struct {
	Name    string
	Type    parameters.ParameterType
	Default interface{}
}

type parameterLayer struct {
	Name        string
	Definitions []parameterDefinition
}

type parsedParameter struct {
	Name  string
	Value interface{}
}

type parsedLayer struct {
	Name       string
	Parameters []parsedParameter
}

// Helper function to create a ParameterLayer from parameterDefinition
func newTestParameterLayer(l parameterLayer) layers.ParameterLayer {
	definitions_ := []*parameters.ParameterDefinition{}
	for _, d := range l.Definitions {
		pd := parameters.NewParameterDefinition(d.Name, d.Type, parameters.WithDefault(d.Default))
		definitions_ = append(definitions_, pd)
	}
	ret, err := layers.NewParameterLayer(l.Name, l.Name,
		layers.WithParameterDefinitions(definitions_...))
	if err != nil {
		panic(err)
	}

	return ret
}

func newParameterLayers(ls []parameterLayer) *layers.ParameterLayers {
	ret := layers.NewParameterLayers()
	for _, l := range ls {
		ret.Set(l.Name, newTestParameterLayer(l))
	}
	return ret
}

// Helper function to create a ParsedLayers from parsedParameter
func newTestParsedLayer(pl layers.ParameterLayer, l parsedLayer) *layers.ParsedLayer {
	params_ := parameters.NewParsedParameters()
	pds := pl.GetParameterDefinitions()
	for _, p := range l.Parameters {
		pd, ok := pds.Get(p.Name)
		if !ok {
			panic("parameter definition not found")
		}
		params_.UpdateValue(p.Name, pd, p.Value)
	}

	ret, err := layers.NewParsedLayer(pl, layers.WithParsedParameters(params_))
	if err != nil {
		panic(err)
	}

	return ret
}

func newParsedLayers(pls *layers.ParameterLayers, ls []parsedLayer) *layers.ParsedLayers {
	ret := layers.NewParsedLayers()
	for _, l := range ls {
		pl, ok := pls.Get(l.Name)
		if !ok {
			panic("parameter layer not found")
		}
		ret.Set(l.Name, newTestParsedLayer(pl, l))
	}
	return ret
}

type expectedLayer struct {
	Name   string
	Values map[string]interface{}
}

func TestFillFromDefaults(t *testing.T) {
	tests := []struct {
		name            string
		parameterLayers []parameterLayer
		parsedLayers    []parsedLayer
		expectedLayers  []expectedLayer
		expectedError   bool
	}{
		{
			name:            "Empty layers and parsedLayers",
			parameterLayers: []parameterLayer{},
			parsedLayers:    []parsedLayer{},
			expectedLayers:  []expectedLayer{},
		},
		{
			name: "Single layer with default",
			parameterLayers: []parameterLayer{
				{
					Name: "layer1",
					Definitions: []parameterDefinition{
						{Name: "param1", Type: "string", Default: "defaultVal"},
					},
				},
			},
			parsedLayers: []parsedLayer{
				{
					Name:       "layer1",
					Parameters: []parsedParameter{},
				},
			},
			expectedLayers: []expectedLayer{
				{
					Name: "layer1",
					Values: map[string]interface{}{
						"param1": "defaultVal",
					},
				},
			},
		},
		{
			name: "Multiple layers with defaults",
			parameterLayers: []parameterLayer{
				{
					Name: "layer1",
					Definitions: []parameterDefinition{
						{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
						{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
					},
				},
				{
					Name: "layer2",
					Definitions: []parameterDefinition{
						{Name: "param3", Type: parameters.ParameterTypeBool, Default: true},
					},
				},
			},
			parsedLayers: []parsedLayer{
				{Name: "layer1", Parameters: []parsedParameter{}},
				{Name: "layer2", Parameters: []parsedParameter{}},
			},
			expectedLayers: []expectedLayer{
				{
					Name: "layer1",
					Values: map[string]interface{}{
						"param1": "default1",
						"param2": 42,
					},
				},
				{
					Name: "layer2",
					Values: map[string]interface{}{
						"param3": true,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "Multiple layers with defaults (no defined target layers, should be created)",
			parameterLayers: []parameterLayer{
				{
					Name: "layer1",
					Definitions: []parameterDefinition{
						{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
						{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
					},
				},
				{
					Name: "layer2",
					Definitions: []parameterDefinition{
						{Name: "param3", Type: parameters.ParameterTypeBool, Default: true},
					},
				},
			},
			parsedLayers: []parsedLayer{},
			expectedLayers: []expectedLayer{
				{
					Name: "layer1",
					Values: map[string]interface{}{
						"param1": "default1",
						"param2": 42,
					},
				},
				{
					Name: "layer2",
					Values: map[string]interface{}{
						"param3": true,
					},
				},
			},
			expectedError: false,
		},
		// Test 4: Layer with No Default Values
		{
			name: "Layer with no default values",
			parameterLayers: []parameterLayer{
				{
					Name: "layer1",
					Definitions: []parameterDefinition{
						{Name: "param1", Type: parameters.ParameterTypeString},
						{Name: "param2", Type: parameters.ParameterTypeInteger},
					},
				},
			},
			parsedLayers: []parsedLayer{
				{Name: "layer1", Parameters: []parsedParameter{}},
			},
			expectedLayers: []expectedLayer{
				{
					Name:   "layer1",
					Values: map[string]interface{}{},
				},
			},
			expectedError: false,
		},
		// Test 5: Layer with Existing Values
		{
			name: "Layer with existing values",
			parameterLayers: []parameterLayer{
				{
					Name: "layer1",
					Definitions: []parameterDefinition{
						{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
						{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
					},
				},
			},
			parsedLayers: []parsedLayer{
				{
					Name: "layer1",
					Parameters: []parsedParameter{
						{Name: "param1", Value: "existingValue"},
					},
				},
			},
			expectedLayers: []expectedLayer{
				{
					Name: "layer1",
					Values: map[string]interface{}{
						"param1": "existingValue", // Should not be overwritten by default
						"param2": 42,              // Should be set to default
					},
				},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layers_ := newParameterLayers(tt.parameterLayers)
			parsedLayers := newParsedLayers(layers_, tt.parsedLayers)

			middleware := middlewares.FillFromDefaults()
			err := middleware(func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
				return nil
			})(layers_, parsedLayers)

			if err != nil {
				t.Errorf("FillFromDefaults() error = %v", err)
				return
			}

			for _, l_ := range tt.expectedLayers {
				l, ok := parsedLayers.Get(l_.Name)
				require.True(t, ok)

				actual := l.Parameters.ToMap()
				assert.Equal(t, l_.Values, actual)
			}
		})
	}
}
