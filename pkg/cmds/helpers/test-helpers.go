package helpers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"gopkg.in/yaml.v3"
)

// Package parameters provides structures and helper functions required for
// conducting table-driven tests in Golang using external YAML files.
//
// The YAML file typically includes test cases with input, expected output,
// and a descriptive name.
// All these will be loaded into structs that use the Test* structs defined here.
//
// This approach makes your tests more clear by extracting verbose the definitions and values.
//
// Here are the steps to use this package:
//
// 1. Define your test cases inside a YAML file.
//    The YAML can use the Test* structs defined here for specific fields.
// 2. Load the YAML file using the LoadTestFromYAML function into your table-driven test struct.
//
// 3. Use the New* functions to convert the data loaded from YAML into appropriate form for the glazed library (typically, ParsedLayers, ParsedDefinitions, ParameterLayers)
//
// 4. Write a standard table-driven test loop.
//
// ---
//
// test_cases.yaml
// ===============
// // fill with your test cases

// - name: "Empty layers and parsedLayers"
//  parameterLayers: []
//  parsedLayers: []
//  expectedLayers: []
//  expectedError: false
//
//- name: "Single layer with default"
//  parameterLayers:
//    - name: "layer1"
//      definitions:
//        - name: "param1"
//          type: "string"
//          default: "defaultVal"
//  parsedLayers:
//    - name: "layer1"
//  expectedLayers:
//    - name: "layer1"
//      values:
//        param1: "defaultVal"
//  ... # potentially additional fields
//
//- name: "Multiple layers with defaults"
//  parameterLayers:
//    - name: "layer1"
//      definitions:
//        - name: "param1"
//          type: "string"
//          default: "default1"
//        - name: "param2"
//          type: "integer"
//          default: 42
//    - name: "layer2"
//      definitions:
//        - name: "param3"
//          type: "bool"
//          default: true
//  parsedLayers:
//    - name: "layer1"
//		parameters:
//		  - name: "param1"
//          value: "existingValue"
//    - name: "layer2"
//  expectedLayers:
//    - name: "layer1"
//      values:
//        param1: "default1"
//        param2: 42
//    - name: "layer2"
//      values:
//        param3: true
//  expectedError: false
//  ... # potentially additional fields
//
// ---
//
// type test struct {
// 	name            string
// 	parameterLayers []parameters.TestParameterLayer
// 	parsedLayers    []parameters.TestParsedLayer
// 	expectedLayers  []parameters.TestExpectedLayer
// 	expectedError   bool
// }
//
// func TestYourFunc(t *testing.T) {
//  tests := LoadTestFromYAML(test_cases.yaml)
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			layers_ := parameters.NewTestParameterLayers(tt.parameterLayers)
// 			parsedLayers := parameters.NewTestParsedLayers(layers_, tt.parsedLayers)
//
//          YourTest()
//
// 			if err != nil {
// 				t.Errorf("YourTest() error = %v", err)
// 				return
// 			}
//
// 			for _, l_ := range tt.expectedLayers {
// 				l, ok := parsedLayers.Get(l_.Name)
// 				require.True(t, ok)
//
// 				actual := l.Parameters.ToMap()
// 				assert.Equal(t, l_.Values, actual)
// 			}
// 		})
// 	}
// }

type TestParameterLayer struct {
	Name        string                            `yaml:"name"`
	Slug        string                            `yaml:"slug"`
	Definitions []*parameters.ParameterDefinition `yaml:"definitions,omitempty"`
	Prefix      string                            `yaml:"prefix"`
}

type TestParsedParameter struct {
	Name  string      `yaml:"name"`
	Value interface{} `yaml:"value"`
}

type TestParsedLayer struct {
	Name       string                `yaml:"name"`
	Parameters []TestParsedParameter `yaml:"parameters"`
}

func LoadTestFromYAML[T any](yaml_ string) (T, error) {
	var ret T
	err := yaml.Unmarshal([]byte(yaml_), &ret)
	if err != nil {
		var zero T
		return zero, err
	}
	return ret, nil
}

type TestExpectedLayer struct {
	Name   string
	Values map[string]interface{}
}

// NewTestParameterLayer is a helper function to create a ParameterLayer from parameterDefinition
func NewTestParameterLayer(l TestParameterLayer) layers.ParameterLayer {
	definitions_ := []*parameters.ParameterDefinition{}
	for _, d := range l.Definitions {
		definitions_ = append(definitions_, d)
	}
	ret, err := layers.NewParameterLayer(l.Name, l.Name,
		layers.WithParameterDefinitions(definitions_...))
	if err != nil {
		panic(err)
	}

	return ret
}

func NewTestParameterLayers(ls []TestParameterLayer) *layers.ParameterLayers {
	ret := layers.NewParameterLayers()
	for _, l := range ls {
		ret.Set(l.Name, NewTestParameterLayer(l))
	}
	return ret
}

// NewTestParsedLayer helper function to create a ParsedLayers from TestParsedParameter
func NewTestParsedLayer(pl layers.ParameterLayer, l TestParsedLayer) *layers.ParsedLayer {
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

func NewTestParsedLayers(pls *layers.ParameterLayers, ls []TestParsedLayer) *layers.ParsedLayers {
	ret := layers.NewParsedLayers()
	for _, l := range ls {
		pl, ok := pls.Get(l.Name)
		if !ok {
			panic("parameter layer not found")
		}
		ret.Set(l.Name, NewTestParsedLayer(pl, l))
	}
	return ret
}
