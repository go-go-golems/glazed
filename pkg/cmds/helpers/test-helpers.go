package helpers

import (
	"fmt"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/errgo.v2/fmt/errors"
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

type TestExpectedLayer struct {
	Name   string                            `yaml:"name"`
	Values map[string]interface{}            `yaml:"values"`
	Logs   map[string][]parameters.ParseStep `yaml:"logs"`
}

type TestMiddlewareName string

const TestMiddlewareSetFromDefaults = "setFromDefaults"
const TestMiddlewareUpdateFromMap = "updateFromMap"
const TestMiddlewareUpdateFromMapAsDefault = "updateFromMapAsDefault"
const TestMiddlewareUpdateFromEnv = "updateFromEnv"
const TestWhitelistLayers = "whitelistLayers"
const TestWhitelistLayersFirst = "whitelistLayersFirst"
const TestWhitelistLayerParameters = "whitelistLayerParameters"
const TestWhitelistLayerParametersFirst = "whitelistLayerParametersFirst"
const TestBlacklistLayers = "blacklistLayers"
const TestBlacklistLayersFirst = "blacklistLayersFirst"
const TestBlacklistLayerParameters = "blacklistLayerParameters"
const TestBlacklistLayerParametersFirst = "blacklistLayerParametersFirst"

type TestParseStepOptionName string

const TestParseStepOptionSource = "source"
const TestParseStepOptionValue = "value"
const TestParseStepOptionMetadata = "metadata"

type TestParseStepOption struct {
	Name  TestParseStepOptionName `yaml:"name"`
	Value interface{}             `yaml:"value"`
}

type TestMiddleware struct {
	Name       TestMiddlewareName                 `yaml:"name"`
	Options    []TestParseStepOption              `yaml:"options"`
	Map        *map[string]map[string]interface{} `yaml:"map"`
	Prefix     *string                            `yaml:"prefix"`
	Layers     *[]string                          `yaml:"layers"`
	Parameters *map[string][]string               `yaml:"parameters"`
}

type TestMiddlewares []TestMiddleware

func (t TestMiddlewares) ToMiddlewares() ([]middlewares.Middleware, error) {
	ret := []middlewares.Middleware{}
	for _, m := range t {
		options := []parameters.ParseStepOption{}
		for _, o := range m.Options {
			switch o.Name {
			case TestParseStepOptionSource:
				options = append(options, parameters.WithParseStepSource(o.Value.(string)))
			case TestParseStepOptionValue:
				options = append(options, parameters.WithParseStepValue(o.Value))
			case TestParseStepOptionMetadata:
				options = append(options, parameters.WithParseStepMetadata(o.Value.(map[string]interface{})))
			default:
				return nil, errors.Newf("unknown option name %s", o.Name)
			}
		}

		switch m.Name {
		case TestMiddlewareSetFromDefaults:
			ret = append(ret, middlewares.SetFromDefaults(options...))
		case TestMiddlewareUpdateFromMap:
			ret = append(ret, middlewares.UpdateFromMap(*m.Map, options...))
		case TestMiddlewareUpdateFromMapAsDefault:
			ret = append(ret, middlewares.UpdateFromMapAsDefault(*m.Map, options...))
		case TestMiddlewareUpdateFromEnv:
			ret = append(ret, middlewares.UpdateFromEnv(*m.Prefix, options...))
		case TestWhitelistLayers:
			ret = append(ret, middlewares.WhitelistLayers(*m.Layers))
		case TestWhitelistLayersFirst:
			ret = append(ret, middlewares.WhitelistLayersFirst(*m.Layers))
		case TestWhitelistLayerParameters:
			ret = append(ret, middlewares.WhitelistLayerParameters(*m.Parameters))
		case TestWhitelistLayerParametersFirst:
			ret = append(ret, middlewares.WhitelistLayerParametersFirst(*m.Parameters))
		case TestBlacklistLayers:
			ret = append(ret, middlewares.BlacklistLayers(*m.Layers))
		case TestBlacklistLayersFirst:
			ret = append(ret, middlewares.BlacklistLayersFirst(*m.Layers))
		case TestBlacklistLayerParameters:
			ret = append(ret, middlewares.BlacklistLayerParameters(*m.Parameters))
		case TestBlacklistLayerParametersFirst:
			ret = append(ret, middlewares.BlacklistLayerParametersFirst(*m.Parameters))
		default:
			return nil, errors.Newf("unknown middleware name %s", m.Name)
		}
	}

	return ret, nil
}

// NewTestParameterLayer is a helper function to create a ParameterLayer from parameterDefinition
func NewTestParameterLayer(l TestParameterLayer) layers.ParameterLayer {
	definitions_ := []*parameters.ParameterDefinition{}
	definitions_ = append(definitions_, l.Definitions...)
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
		err := params_.UpdateValue(p.Name, pd, p.Value)
		if err != nil {
			panic(err)
		}
	}

	ret, err := layers.NewParsedLayer(pl, layers.WithParsedParameters(params_))
	if err != nil {
		panic(err)
	}

	return ret
}

func NewTestParsedLayers(pls *layers.ParameterLayers, ls ...TestParsedLayer) *layers.ParsedLayers {
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

// compareValues handles comparison of values with special cases for slice types
func compareValues(t *testing.T, expected, actual interface{}, key string) {
	normalizedExpected, err := cast.NormalizeValue(expected)
	require.NoError(t, err, "failed to normalize expected value for key %s", key)

	normalizedActual, err := cast.NormalizeValue(actual)
	require.NoError(t, err, "failed to normalize actual value for key %s", key)

	assert.Equal(t, normalizedExpected, normalizedActual, "mismatch for key %s", key)
}

func TestExpectedOutputs(t *testing.T, expectedLayers []TestExpectedLayer, parsedLayers *layers.ParsedLayers) {
	expectedLayers_ := map[string]TestExpectedLayer{}
	for _, l_ := range expectedLayers {
		expectedLayers_[l_.Name] = l_
		l, ok := parsedLayers.Get(l_.Name)
		require.True(t, ok)

		actual, err := l.Parameters.ToInterfaceMap()
		require.NoError(t, err)

		// Compare each value using the helper
		for k, expectedValue := range l_.Values {
			compareValues(t, expectedValue, actual[k], k)
		}

		for k, expectedLog := range l_.Logs {
			actual, ok := l.Parameters.Get(k)
			require.True(t, ok)

			// Compare each log entry using the helper
			for i, expectedEntry := range expectedLog {
				if i < len(actual.Log) {
					compareValues(t, expectedEntry.Value, actual.Log[i].Value, fmt.Sprintf("%s.Log[%d].Value", k, i))
					assert.Equal(t, expectedEntry.Source, actual.Log[i].Source, fmt.Sprintf("%s.Log[%d].Source", k, i))
				}
			}
		}
	}

	parsedLayers.ForEach(func(key string, l *layers.ParsedLayer) {
		if _, ok := expectedLayers_[key]; !ok {
			t.Errorf("did not expect layer %s to be present", key)
		}
	})
}
