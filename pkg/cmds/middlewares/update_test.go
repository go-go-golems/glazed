package middlewares

import (
	"os"
	"testing"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/require"
)

func TestUpdateFromEnvParsesTypedValues(t *testing.T) {
	// Define a layer with a prefix so env keys are: PREFIX + "_" + UPPER(prefix+name)
	cfgLayer, err := layers.NewParameterLayer("cfg", "Config",
		layers.WithPrefix("cfg-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool),
			parameters.NewParameterDefinition("retries", parameters.ParameterTypeInteger),
			parameters.NewParameterDefinition("ratio", parameters.ParameterTypeFloat),
			parameters.NewParameterDefinition("start", parameters.ParameterTypeDate),
			parameters.NewParameterDefinition("mode", parameters.ParameterTypeChoice, parameters.WithChoices("a", "b")),
			parameters.NewParameterDefinition("names", parameters.ParameterTypeStringList),
			parameters.NewParameterDefinition("nums", parameters.ParameterTypeIntegerList),
			parameters.NewParameterDefinition("floats", parameters.ParameterTypeFloatList),
			parameters.NewParameterDefinition("labels", parameters.ParameterTypeKeyValue),
			parameters.NewParameterDefinition("user", parameters.ParameterTypeString),
		),
	)
	require.NoError(t, err)

	pl := layers.NewParameterLayers(layers.WithLayers(cfgLayer))
	parsed := layers.NewParsedLayers()

	// Set env vars
	env := map[string]string{
		"APP_CFG_VERBOSE": "true",
		"APP_CFG_RETRIES": "3",
		"APP_CFG_RATIO":   "0.5",
		"APP_CFG_START":   "2025-01-02",
		"APP_CFG_MODE":    "a",
		"APP_CFG_NAMES":   "alice,bob",
		"APP_CFG_NUMS":    "1,2,3",
		"APP_CFG_FLOATS":  "0.1,2.3",
		"APP_CFG_LABELS":  "k1:v1,k2:v2",
		"APP_CFG_USER":    "manuel",
	}
	// remember old env and restore
	prev := map[string]*string{}
	for k, v := range env {
		if old, ok := os.LookupEnv(k); ok {
			prev[k] = &old
		} else {
			prev[k] = nil
		}
		err := os.Setenv(k, v)
		require.NoError(t, err)
	}
	t.Cleanup(func() {
		for k, v := range prev {
			if v == nil {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, *v)
			}
		}
	})

	err = ExecuteMiddlewares(pl, parsed,
		UpdateFromEnv("APP", parameters.WithParseStepSource("env")),
	)
	require.NoError(t, err)

	layer, ok := parsed.Get("cfg")
	require.True(t, ok)

	get := func(name string) interface{} {
		v, ok := layer.Parameters.Get(name)
		require.True(t, ok, "parameter %s should be set", name)
		return v.Value
	}

	require.Equal(t, true, get("verbose"))
	require.Equal(t, 3, get("retries"))
	require.InDelta(t, 0.5, get("ratio").(float64), 1e-9)
	require.Equal(t, "a", get("mode"))
	require.Equal(t, "manuel", get("user"))

	// Date: compare YYYY-MM-DD
	d := get("start").(time.Time)
	require.Equal(t, "2025-01-02", d.Format("2006-01-02"))

	// Lists and key-value
	require.Equal(t, []string{"alice", "bob"}, get("names"))
	require.Equal(t, []int{1, 2, 3}, get("nums"))
	require.InDeltaSlice(t, []float64{0.1, 2.3}, get("floats").([]float64), 1e-9)
	require.Equal(t, map[string]string{"k1": "v1", "k2": "v2"}, get("labels"))

	// Check env_key metadata exists on one param (verbose)
	vp, ok := layer.Parameters.Get("verbose")
	require.True(t, ok)
	require.NotEmpty(t, vp.Log)
	found := false
	for _, step := range vp.Log {
		if step.Source == "env" && step.Metadata != nil {
			if ek, ok := step.Metadata["env_key"]; ok && ek == "APP_CFG_VERBOSE" {
				found = true
				break
			}
		}
	}
	require.True(t, found, "expected env_key metadata on verbose")
}

func TestUpdateFromEnvInvalidChoice(t *testing.T) {
	cfgLayer, err := layers.NewParameterLayer("cfg", "Config",
		layers.WithPrefix("cfg-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("mode", parameters.ParameterTypeChoice, parameters.WithChoices("a", "b")),
		),
	)
	require.NoError(t, err)

	pl := layers.NewParameterLayers(layers.WithLayers(cfgLayer))
	parsed := layers.NewParsedLayers()

	prev, had := os.LookupEnv("APP_CFG_MODE")
	_ = os.Setenv("APP_CFG_MODE", "c") // invalid
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("APP_CFG_MODE", prev)
		} else {
			_ = os.Unsetenv("APP_CFG_MODE")
		}
	})

	err = ExecuteMiddlewares(pl, parsed, UpdateFromEnv("APP"))
	require.Error(t, err, "should error on invalid choice")
}
