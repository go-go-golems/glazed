package middlewares_test

import (
	"bufio"
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/helpers"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

type test struct {
	Name            string                       `yaml:"name"`
	ParameterLayers []helpers.TestParameterLayer `yaml:"parameterLayers"`
	ParsedLayers    []helpers.TestParsedLayer    `yaml:"parsedLayers"`
	ExpectedLayers  []helpers.TestExpectedLayer  `yaml:"expectedLayers"`
	ExpectedError   bool                         `yaml:"expectedError"`
}

//go:embed tests/fill-from-defaults.yaml
var fillFromDefaultsTestsYAML string

func TestFillFromDefaults(t *testing.T) {
	tests, err := helpers.LoadTestFromYAML[[]test](fillFromDefaultsTestsYAML)
	require.NoError(t, err)

	//tests := []test{
	//	{
	//		"Empty layers and parsedLayers",
	//		[]parameters.TestParameterLayer{},
	//		[]parameters.TestParsedLayer{},
	//		[]parameters.TestExpectedLayer{},
	//		false,
	//	},
	//	{
	//		name: "Single layer with default",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: "string", Default: "defaultVal"},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{
	//				Name:       "layer1",
	//				Parameters: []parameters.TestParsedParameter{},
	//			},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name: "layer1",
	//				Values: map[string]interface{}{
	//					"param1": "defaultVal",
	//				},
	//			},
	//		},
	//	},
	//	{
	//		name: "Multiple layers with defaults",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
	//					{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
	//				},
	//			},
	//			{
	//				Name: "layer2",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param3", Type: parameters.ParameterTypeBool, Default: true},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//			{Name: "layer2", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name: "layer1",
	//				Values: map[string]interface{}{
	//					"param1": "default1",
	//					"param2": 42,
	//				},
	//			},
	//			{
	//				Name: "layer2",
	//				Values: map[string]interface{}{
	//					"param3": true,
	//				},
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	{
	//		name: "Multiple layers with defaults (no defined target layers, should be created)",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
	//					{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
	//				},
	//			},
	//			{
	//				Name: "layer2",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param3", Type: parameters.ParameterTypeBool, Default: true},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name: "layer1",
	//				Values: map[string]interface{}{
	//					"param1": "default1",
	//					"param2": 42,
	//				},
	//			},
	//			{
	//				Name: "layer2",
	//				Values: map[string]interface{}{
	//					"param3": true,
	//				},
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	// Test 4: Layer with No Default Values
	//	{
	//		name: "Layer with no default values",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString},
	//					{Name: "param2", Type: parameters.ParameterTypeInteger},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name:   "layer1",
	//				Values: map[string]interface{}{},
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	// Test 5: Layer with Existing Values
	//	{
	//		name: "Layer with existing values",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
	//					{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{
	//				Name: "layer1",
	//				Parameters: []parameters.TestParsedParameter{
	//					{Name: "param1", Value: "existingValue"},
	//				},
	//			},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name: "layer1",
	//				Values: map[string]interface{}{
	//					"param1": "existingValue", // Should not be overwritten by default
	//					"param2": 42,              // Should be set to default
	//				},
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	{
	//		// Test 6: Layer with Partially Set Values
	//		name: "Layer with partially set values",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
	//					{Name: "param2", Type: parameters.ParameterTypeInteger, Default: 42},
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{
	//				Name: "layer1",
	//				Parameters: []parameters.TestParsedParameter{
	//					{Name: "param1", Value: "existingValue"},
	//					// param2 is not set and should be filled with the default
	//				},
	//			},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name: "layer1",
	//				Values: map[string]interface{}{
	//					"param1": "existingValue", // Existing value should remain
	//					"param2": 42,              // Default value should be set
	//				},
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	// Test 7: Layer with Invalid Default Values
	//	{
	//		name: "Layer with invalid default values",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "notAnInt"}, // Invalid default value
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{},
	//		expectedError:  true, // Expect an error due to invalid default value
	//	},
	//	// Test 8: Layer with Required Parameters Without Defaults
	//	{
	//		name: "Layer with required parameters without defaults",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Required: true}, // Required but no default
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name:   "layer1",
	//				Values: map[string]interface{}{}, // No default should be set
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	// Test 9: Layer with Optional Parameters Without Defaults
	//	{
	//		name: "Layer with optional parameters without defaults",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString}, // Optional and no default
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{
	//			{
	//				Name:   "layer1",
	//				Values: map[string]interface{}{}, // No default should be set
	//			},
	//		},
	//		expectedError: false,
	//	},
	//	// Test 10: Error During Parameter Definition Iteration
	//	{
	//		name: "Error during parameter definition iteration",
	//		parameterLayers: []parameters.TestParameterLayer{
	//			{
	//				Name: "layer1",
	//				Definitions: []*parameters.ParameterDefinition{
	//					{Name: "param1", Type: parameters.ParameterTypeString, Default: "default1"},
	//					// Assume the next parameter causes an error during iteration
	//				},
	//			},
	//		},
	//		parsedLayers: []parameters.TestParsedLayer{
	//			{Name: "layer1", Parameters: []parameters.TestParsedParameter{}},
	//		},
	//		expectedLayers: []parameters.TestExpectedLayer{}, // No layers should be set due to the error
	//		expectedError:  true,                             // Expect an error during iteration
	//	},
	//}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			middleware := middlewares.FillFromDefaults()
			err := middleware(func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
				return nil
			})(layers_, parsedLayers)

			if err != nil {
				t.Errorf("FillFromDefaults() error = %v", err)
				return
			}

			for _, l_ := range tt.ExpectedLayers {
				l, ok := parsedLayers.Get(l_.Name)
				require.True(t, ok)

				actual := l.Parameters.ToMap()
				assert.Equal(t, l_.Values, actual)
			}
		})
	}
}

func CountLines(r io.Reader) (int, error) {
	sc := bufio.NewScanner(r)
	lines := 0

	for sc.Scan() {
		lines++
	}

	return lines, nil
}
