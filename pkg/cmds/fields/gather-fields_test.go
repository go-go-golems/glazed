package fields

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/helpers/yaml"
	"github.com/stretchr/testify/require"
	"testing"
)

// TestCase represents a single test case for the GatherArguments function.
type TestCase struct {
	Title       string        `yaml:"title"`
	Description string        `yaml:"description"`
	FieldDefs   []*Definition `yaml:"fieldDefs"`
	// The actual map fromwhich the fields are gathered
	Data map[string]interface{} `yaml:"data"`
	// Only gather fields that are provided in the map
	OnlyProvided bool `yaml:"onlyProvided"`
	// The expected result of the test
	ExpectedResult map[string]interface{} `yaml:"expectedResult"`
	ExpectedError  string                 `yaml:"expectedError"`
}

//go:embed test-data/gather-fields.yaml
var gatherFieldsYAML string

func TestGatherFieldsFromMap(t *testing.T) {
	tests, err := yaml.LoadTestFromYAML[[]TestCase](gatherFieldsYAML)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.Title, func(t *testing.T) {
			pds := NewDefinitions(WithDefinitionList(tt.FieldDefs))

			fieldValues, err := pds.GatherFieldsFromMap(tt.Data, tt.OnlyProvided)

			if tt.ExpectedError != "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.ExpectedResult, fieldValues.ToMap())
			}
		})
	}
}
