package fields

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithMetadataCopiesIncomingMap(t *testing.T) {
	shared := map[string]interface{}{
		"config_file": "config.yaml",
		"index":       0,
	}

	step := NewParseStep(WithMetadata(shared))
	shared["map-value"] = "sk-test-secret"

	require.Equal(t, "config.yaml", step.Metadata["config_file"])
	require.Equal(t, 0, step.Metadata["index"])
	_, hasMapValue := step.Metadata["map-value"]
	require.False(t, hasMapValue)
}

func TestWithMetadataNoCrossStepAliasing(t *testing.T) {
	shared := map[string]interface{}{
		"config_file": "config.yaml",
		"index":       0,
	}

	step1 := NewParseStep(WithMetadata(shared))
	step2 := NewParseStep(
		WithMetadata(shared),
		WithMetadata(map[string]interface{}{"map-value": "alpha"}),
	)
	step3 := NewParseStep(
		WithMetadata(shared),
		WithMetadata(map[string]interface{}{"map-value": "beta"}),
	)

	_, step1HasMapValue := step1.Metadata["map-value"]
	require.False(t, step1HasMapValue)
	require.Equal(t, "alpha", step2.Metadata["map-value"])
	require.Equal(t, "beta", step3.Metadata["map-value"])
}
