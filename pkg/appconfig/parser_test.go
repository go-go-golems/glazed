package appconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/stretchr/testify/require"
)

type testAppSettings struct {
	Redis testRedisSettings
}

type testRedisSettings struct {
	Host string `glazed.parameter:"host"`
}

type testRedisSettingsNoTags struct {
	Host string
}

func newTestRedisLayer(defaultHost string) layers.ParameterLayer {
	layer, err := layers.NewParameterLayer(
		"redis",
		"Redis",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"host",
				parameters.ParameterTypeString,
				parameters.WithDefault(defaultHost),
			),
		),
	)
	if err != nil {
		// This is a test helper; fail fast.
		panic(err)
	}
	return layer
}

func TestParser_Register_Validation(t *testing.T) {
	layer := newTestRedisLayer("default")

	t.Run("empty slug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register("", layer, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("nil layer", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register("redis", nil, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("nil bind", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register("redis", layer, nil)
		require.Error(t, err)
	})

	t.Run("slug mismatch with layer.GetSlug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register("not-redis", layer, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("duplicate slug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		require.NoError(t, p.Register("redis", layer, func(t *testAppSettings) any { return &t.Redis }))
		err = p.Register("redis", layer, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})
}

func TestParser_Parse_BinderFailures(t *testing.T) {
	layer := newTestRedisLayer("default")

	t.Run("bind returns nil", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForLayers(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register("redis", layer, func(_ *testAppSettings) any { return nil }))

		_, err = p.Parse()
		require.Error(t, err)
	})

	t.Run("bind returns non-pointer", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForLayers(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register("redis", layer, func(t *testAppSettings) any { return t.Redis }))

		_, err = p.Parse()
		require.Error(t, err)
	})

	t.Run("bind returns nil pointer", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForLayers(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register("redis", layer, func(_ *testAppSettings) any {
			var rs *testRedisSettings
			return rs
		}))

		_, err = p.Parse()
		require.Error(t, err)
	})
}

func TestParser_Parse_Hydration_TagRequired(t *testing.T) {
	layer := newTestRedisLayer("default")

	type app struct {
		Redis testRedisSettingsNoTags
	}

	p, err := NewParser[app](WithValuesForLayers(map[string]map[string]interface{}{
		"redis": {"host": "from-map"},
	}))
	require.NoError(t, err)
	require.NoError(t, p.Register("redis", layer, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	// No glazed.parameter tags -> field is ignored by InitializeStruct -> zero value.
	require.Equal(t, "", cfg.Redis.Host)
}

func TestParser_Parse_Precedence_Defaults_Config_Env(t *testing.T) {
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "base.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("redis:\n  host: from-file\n"), 0o644))

	t.Setenv("MYAPP_HOST", "from-env")

	p, err := NewParser[testAppSettings](
		WithConfigFiles(cfgFile),
		WithEnv("MYAPP"),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register("redis", layer, func(t *testAppSettings) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-env", cfg.Redis.Host)
}
