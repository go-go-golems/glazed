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
