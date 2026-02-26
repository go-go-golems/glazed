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

func TestGatherFieldsFromMap_MapValueMetadataScopedPerField(t *testing.T) {
	pds := NewDefinitions(WithDefinitionList([]*Definition{
		New("ai-engine", TypeString),
		New("openai-api-key", TypeString),
	}))

	fieldValues, err := pds.GatherFieldsFromMap(
		map[string]interface{}{
			"ai-engine":      "gpt-4o-mini",
			"openai-api-key": "sk-test-secret",
		},
		true,
		WithSource("config"),
		WithMetadata(map[string]interface{}{
			"config_file": "config.yaml",
			"index":       0,
		}),
	)
	require.NoError(t, err)

	aiEngine, ok := fieldValues.Get("ai-engine")
	require.True(t, ok)
	require.Len(t, aiEngine.Log, 1)
	require.Equal(t, "gpt-4o-mini", aiEngine.Log[0].Metadata["map-value"])
	require.Equal(t, "config.yaml", aiEngine.Log[0].Metadata["config_file"])
	require.Equal(t, 0, aiEngine.Log[0].Metadata["index"])

	openAIKey, ok := fieldValues.Get("openai-api-key")
	require.True(t, ok)
	require.Len(t, openAIKey.Log, 1)
	require.Equal(t, "sk-test-secret", openAIKey.Log[0].Metadata["map-value"])
	require.Equal(t, "config.yaml", openAIKey.Log[0].Metadata["config_file"])
	require.Equal(t, 0, openAIKey.Log[0].Metadata["index"])
}
