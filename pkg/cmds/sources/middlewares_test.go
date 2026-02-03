package sources_test

import (
	_ "embed"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/helpers"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/helpers/list"
	"github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type test struct {
	Name             string                        `yaml:"name"`
	Description      string                        `yaml:"description"`
	Sections         []helpers.TestSection         `yaml:"sections"`
	Values           []helpers.TestSectionValues   `yaml:"values"`
	ExpectedSections []helpers.TestExpectedSection `yaml:"expectedSections"`
	ExpectedError    bool                          `yaml:"expectedError"`
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
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := helpers.NewTestValues(schema_, tt.Values...)

			middleware := sources.FromDefaults(fields.WithSource(fields.SourceDefaults))
			err := middleware(func(schema_ *schema.Schema, parsedValues *values.Values) error {
				return nil
			})(schema_, parsedValues)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
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
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := helpers.NewTestValues(schema_, tt.Values...)

			err = sources.Execute(
				schema_, parsedValues,
				sources.FromMap(tt.UpdateMaps),
			)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
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
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := helpers.NewTestValues(schema_, tt.Values...)

			err = sources.Execute(
				schema_, parsedValues,
				sources.FromMapAsDefault(tt.UpdateMaps),
			)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
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
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := helpers.NewTestValues(schema_, tt.Values...)

			middlewares_ := []sources.Middleware{}
			// we want the first updates to be handled as the last middlewares, since middlewares
			// at the end are executed first
			for _, m := range tt.UpdateMaps {
				middlewares_ = list.Prepend(middlewares_, sources.FromMap(m))
			}
			err = sources.Execute(
				schema_, parsedValues,
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
				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
			}
		})
	}
}

//go:embed tests/wrap-with-restricted-sections.yaml
var wrapWithRestrictedSectionsTestsYAML string

type wrapWithRestrictedSectionsTest struct {
	test                  `yaml:",inline"`
	BlacklistedUpdateMaps map[string]map[string]interface{} `yaml:"blacklistedUpdateMaps"`
	WhitelistedUpdateMaps map[string]map[string]interface{} `yaml:"whitelistedUpdateMaps"`
	BeforeUpdateMaps      map[string]map[string]interface{} `yaml:"beforeUpdateMaps"`
	AfterUpdateMaps       map[string]map[string]interface{} `yaml:"afterUpdateMaps"`
	BlacklistedSlugs      []string                          `yaml:"blacklistedSlugs"`
	WhitelistedSlugs      []string                          `yaml:"whitelistedSlugs"`
}

func TestWrapWithRestrictedSections(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]wrapWithRestrictedSectionsTest](wrapWithRestrictedSectionsTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := values.New()

			ms_ := []sources.Middleware{}

			if tt.AfterUpdateMaps != nil {
				ms_ = append(ms_, sources.FromMap(tt.AfterUpdateMaps))
			}
			if tt.BlacklistedUpdateMaps != nil {
				ms_ = append(ms_, sources.WrapWithBlacklistedSections(tt.BlacklistedSlugs,
					sources.FromMap(tt.BlacklistedUpdateMaps)))
			}
			if tt.WhitelistedUpdateMaps != nil {
				ms_ = append(ms_, sources.WrapWithWhitelistedSections(tt.WhitelistedSlugs,
					sources.FromMap(tt.WhitelistedUpdateMaps)))
			}
			if tt.BeforeUpdateMaps != nil {
				ms_ = append(ms_, sources.FromMap(tt.BeforeUpdateMaps))
			}

			err = sources.Execute(
				schema_, parsedValues,
				ms_...)
			require.NoError(t, err)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
			}
		})
	}
}

//go:embed tests/middlewares.yaml
var middlewaresTestsYAML string

type middlewaresTest struct {
	test        `yaml:",inline"`
	Middlewares helpers.TestMiddlewares `yaml:"middlewares"`
}

func TestMiddlewares(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]middlewaresTest](middlewaresTestsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			schema_ := helpers.NewTestSchema(tt.Sections)
			parsedValues := helpers.NewTestValues(schema_, tt.Values...)

			middlewares_, err := helpers.TestMiddlewares(tt.Middlewares).ToMiddlewares()
			require.NoError(t, err)

			err = sources.Execute(schema_, parsedValues, middlewares_...)
			require.NoError(t, err)

			if tt.ExpectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				helpers.TestExpectedOutputs(t, tt.ExpectedSections, parsedValues)
			}
		})
	}
}
