package appconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type testAppSettings struct {
	Redis testRedisSettings
}

type testRedisSettings struct {
	Host string `glazed:"host"`
}

type testRedisSettingsNoTags struct {
	Host string
}

func newTestRedisSection(defaultHost string) schema.Section {
	section, err := schema.NewSection(
		"redis",
		"Redis",
		schema.WithFields(
			fields.New(
				"host",
				fields.TypeString,
				fields.WithDefault(defaultHost),
			),
		),
	)
	if err != nil {
		// This is a test helper; fail fast.
		panic(err)
	}
	return section
}

func TestParser_Register_Validation(t *testing.T) {
	const redisSlug SectionSlug = "redis"
	section := newTestRedisSection("default")

	t.Run("empty slug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register("", section, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("nil section", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register(redisSlug, nil, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("nil bind", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register(redisSlug, section, nil)
		require.Error(t, err)
	})

	t.Run("slug mismatch with section.GetSlug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		err = p.Register(SectionSlug("not-redis"), section, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})

	t.Run("duplicate slug", func(t *testing.T) {
		p, err := NewParser[testAppSettings]()
		require.NoError(t, err)
		require.NoError(t, p.Register(redisSlug, section, func(t *testAppSettings) any { return &t.Redis }))
		err = p.Register(redisSlug, section, func(t *testAppSettings) any { return &t.Redis })
		require.Error(t, err)
	})
}

func TestParser_Parse_BinderFailures(t *testing.T) {
	const redisSlug SectionSlug = "redis"
	section := newTestRedisSection("default")

	t.Run("bind returns nil", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForSections(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register(redisSlug, section, func(_ *testAppSettings) any { return nil }))

		_, err = p.Parse()
		require.Error(t, err)
	})

	t.Run("bind returns non-pointer", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForSections(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register(redisSlug, section, func(t *testAppSettings) any { return t.Redis }))

		_, err = p.Parse()
		require.Error(t, err)
	})

	t.Run("bind returns nil pointer", func(t *testing.T) {
		p, err := NewParser[testAppSettings](WithValuesForSections(map[string]map[string]interface{}{
			"redis": {"host": "x"},
		}))
		require.NoError(t, err)
		require.NoError(t, p.Register(redisSlug, section, func(_ *testAppSettings) any {
			var rs *testRedisSettings
			return rs
		}))

		_, err = p.Parse()
		require.Error(t, err)
	})
}

func TestParser_Parse_Hydration_TagRequired(t *testing.T) {
	const redisSlug SectionSlug = "redis"
	section := newTestRedisSection("default")

	type app struct {
		Redis testRedisSettingsNoTags
	}

	p, err := NewParser[app](WithValuesForSections(map[string]map[string]interface{}{
		"redis": {"host": "from-map"},
	}))
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, section, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	// No glazed tags -> field is ignored by DecodeSectionInto -> zero value.
	require.Equal(t, "", cfg.Redis.Host)
}

func TestParser_Parse_Precedence_Defaults_Config_Env(t *testing.T) {
	const redisSlug SectionSlug = "redis"
	section := newTestRedisSection("from-default")

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "base.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("redis:\n  host: from-file\n"), 0o644))

	t.Setenv("MYAPP_HOST", "from-env")

	p, err := NewParser[testAppSettings](
		WithDefaults(),
		WithConfigFiles(cfgFile),
		WithEnv("MYAPP"),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, section, func(t *testAppSettings) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-env", cfg.Redis.Host)
}

func TestParser_Parse_Precedence_CobraFlagsOverrideEnv(t *testing.T) {
	const redisSlug SectionSlug = "redis"
	section := newTestRedisSection("from-default")

	// Env sets host to from-env.
	t.Setenv("MYAPP_HOST", "from-env")

	// Cobra flag should override env.
	rootCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			parser, err := NewParser[testAppSettings](
				WithEnv("MYAPP"),
				WithCobra(cmd, args),
			)
			require.NoError(t, err)
			require.NoError(t, parser.Register(redisSlug, section, func(t *testAppSettings) any { return &t.Redis }))

			cfg, err := parser.Parse()
			require.NoError(t, err)
			require.Equal(t, "from-flag", cfg.Redis.Host)
			return nil
		},
	}

	// IMPORTANT: appconfig.Parser expects the section flags to already be on the cobra command.
	cobraSection, ok := section.(schema.CobraSection)
	require.True(t, ok, "test section must implement schema.CobraSection")
	require.NoError(t, cobraSection.AddSectionToCobraCommand(rootCmd))

	rootCmd.SetArgs([]string{"--host", "from-flag"})
	require.NoError(t, rootCmd.Execute())
}
