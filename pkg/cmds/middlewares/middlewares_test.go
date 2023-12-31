package middlewares_test

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds/helpers"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/helpers/yaml"
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
	test `yaml:",inline"`
}

func TestSetFromDefaults(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]setFromDefaultsTest](setFromDefaultsTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			middleware := middlewares.SetFromDefaults()
			err := middleware(func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error {
				return nil
			})(layers_, parsedLayers)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				helpers.TestExpectedOutputs(t, tt.ExpectedLayers, parsedLayers)
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
	tests, err := yaml.LoadTestFromYAML[[]updateFromMapTest](updateFromMapTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			err = middlewares.ExecuteMiddlewares(
				layers_, parsedLayers,
				middlewares.UpdateFromMap(tt.UpdateMaps),
			)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				helpers.TestExpectedOutputs(t, tt.ExpectedLayers, parsedLayers)
			}
		})
	}
}

//go:embed tests/update-from-map-as-default.yaml
var updateFromMapAsDefaultsTestYAML string

func TestUpdateFromMapAsDefault(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]updateFromMapTest](updateFromMapAsDefaultsTestYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := helpers.NewTestParsedLayers(layers_, tt.ParsedLayers)

			err = middlewares.ExecuteMiddlewares(
				layers_, parsedLayers,
				middlewares.UpdateFromMapAsDefault(tt.UpdateMaps),
			)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				helpers.TestExpectedOutputs(t, tt.ExpectedLayers, parsedLayers)
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
	tests, err := yaml.LoadTestFromYAML[[]multiUpdateFromMapTest](multiUpdateFromMapTestsYAML)
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

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				helpers.TestExpectedOutputs(t, tt.ExpectedLayers, parsedLayers)
			}
		})
	}
}

//go:embed tests/wrap-with-restricted-layers.yaml
var wrapWithRestrictedLayersTestsYAML string

type wrapWithRestrictedLayersTest struct {
	test                  `yaml:",inline"`
	BlacklistedUpdateMaps map[string]map[string]interface{} `yaml:"blacklistedUpdateMaps"`
	WhitelistedUpdateMaps map[string]map[string]interface{} `yaml:"whitelistedUpdateMaps"`
	BeforeUpdateMaps      map[string]map[string]interface{} `yaml:"beforeUpdateMaps"`
	AfterUpdateMaps       map[string]map[string]interface{} `yaml:"afterUpdateMaps"`
	BlacklistedSlugs      []string                          `yaml:"blacklistedSlugs"`
	WhitelistedSlugs      []string                          `yaml:"whitelistedSlugs"`
}

func TestWrapWithRestrictedLayers(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]wrapWithRestrictedLayersTest](wrapWithRestrictedLayersTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			layers_ := helpers.NewTestParameterLayers(tt.ParameterLayers)
			parsedLayers := layers.NewParsedLayers()

			ms_ := []middlewares.Middleware{}

			if tt.BeforeUpdateMaps != nil {
				ms_ = append(ms_, middlewares.UpdateFromMap(tt.BeforeUpdateMaps))
			}
			if tt.BlacklistedUpdateMaps != nil {
				ms_ = append(ms_, middlewares.WrapWithBlacklistedLayers(tt.BlacklistedSlugs,
					middlewares.UpdateFromMap(tt.BlacklistedUpdateMaps)))
			}
			if tt.WhitelistedUpdateMaps != nil {
				ms_ = append(ms_, middlewares.WrapWithWhitelistedLayers(tt.WhitelistedSlugs,
					middlewares.UpdateFromMap(tt.WhitelistedUpdateMaps)))
			}
			if tt.AfterUpdateMaps != nil {
				ms_ = append(ms_, middlewares.UpdateFromMap(tt.AfterUpdateMaps))
			}

			err = middlewares.ExecuteMiddlewares(
				layers_, parsedLayers,
				ms_...)
			require.NoError(t, err)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				helpers.TestExpectedOutputs(t, tt.ExpectedLayers, parsedLayers)
			}
		})
	}
}
