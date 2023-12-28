package middlewares_test

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/helpers"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type test struct {
	Name            string                       `yaml:"name"`
	Description     string                       `yaml:"description"`
	ParameterLayers []helpers.TestParameterLayer `yaml:"parameterLayers"`
	ParsedLayers    []helpers.TestParsedLayer    `yaml:"parsedLayers"`
	ExpectedLayers  []helpers.TestExpectedLayer  `yaml:"expectedLayers"`
	ExpectedError   bool                         `yaml:"expectedError"`
}

//go:embed tests/set-from-defaults.yaml
var setFromDefaultsTestsYAML string

type setFromDefaultsTest struct {
	test
}

func TestSetFromDefaults(t *testing.T) {
	tests, err := helpers.LoadTestFromYAML[[]setFromDefaultsTest](setFromDefaultsTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			middleware := middlewares.SetFromDefaults()
			err := middleware(func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
				return nil
			})(layers_, parsedLayers)

			if err != nil {
				t.Errorf("SetFromDefaults() error = %v", err)
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

type updateFromMapTest struct {
	test       `yaml:",inline"`
	UpdateMaps map[string]map[string]interface{} `yaml:"updateMaps"`
}

//go:embed tests/update-from-map.yaml
var updateFromMapTestsYAML string

func TestUpdateFromMap(t *testing.T) {
	tests, err := helpers.LoadTestFromYAML[[]updateFromMapTest](updateFromMapTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			err = middlewares.ExecuteMiddlewares(
				layers_, parsedLayers,
				middlewares.UpdateFromMap(tt.UpdateMaps),
			)

			if err != nil {
				t.Errorf("UpdateFromMap() error = %v", err)
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

type multiUpdateFromMapTest struct {
	test       `yaml:",inline"`
	UpdateMaps []map[string]map[string]interface{} `yaml:"updateMaps"`
}

//go:embed tests/multi-update-from-map.yaml
var multiUpdateFromMapTestsYAML string

func TestMultiUpdateFromMap(t *testing.T) {
	tests, err := helpers.LoadTestFromYAML[[]multiUpdateFromMapTest](multiUpdateFromMapTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			middlewares_ := []middlewares.Middleware{}
			for _, m := range tt.UpdateMaps {
				middlewares_ = append(middlewares_, middlewares.UpdateFromMap(m))
			}
			err = middlewares.ExecuteMiddlewares(
				layers_, parsedLayers,
				middlewares_...,
			)

			if err != nil {
				t.Errorf("MultiUpdateFromMap() error = %v", err)
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
