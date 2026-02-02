package appconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cli"
	schema "github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestWithProfile_AppliesDefaultProfile(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	profilesPath := filepath.Join(dir, "profiles.yaml")
	require.NoError(t, os.WriteFile(profilesPath, []byte(`default:
  redis:
    host: from-profile-default
prod:
  redis:
    host: from-profile-prod
`), 0o644))

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(profilesPath),
			WithProfileFile(profilesPath),
		),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-profile-default", cfg.Redis.Host)
}

func TestWithProfile_ProfileSelection_FromEnv(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	profilesPath := filepath.Join(dir, "profiles.yaml")
	require.NoError(t, os.WriteFile(profilesPath, []byte(`default:
  redis:
    host: from-profile-default
prod:
  redis:
    host: from-profile-prod
`), 0o644))

	t.Setenv("MYAPP_PROFILE", "prod")

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(profilesPath),
			WithProfileFile(profilesPath),
		),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-profile-prod", cfg.Redis.Host)
}

func TestWithProfile_ProfileSelection_FromConfig(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	profilesPath := filepath.Join(dir, "profiles.yaml")
	require.NoError(t, os.WriteFile(profilesPath, []byte(`default:
  redis:
    host: from-profile-default
prod:
  redis:
    host: from-profile-prod
`), 0o644))

	cfgPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`profile-settings:
  profile: prod
`), 0o644))

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(profilesPath),
			WithProfileFile(profilesPath),
		),
		WithConfigFiles(cfgPath),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-profile-prod", cfg.Redis.Host)
}

func TestWithProfile_ProfileSelection_FromCobraFlag(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	profilesPath := filepath.Join(dir, "profiles.yaml")
	require.NoError(t, os.WriteFile(profilesPath, []byte(`default:
  redis:
    host: from-profile-default
prod:
  redis:
    host: from-profile-prod
`), 0o644))

	rootCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			type app struct {
				Redis testRedisSettings
			}

			p, err := NewParser[app](
				WithDefaults(),
				WithProfile("myapp",
					WithProfileDefaultFile(profilesPath),
					WithProfileFile(profilesPath),
				),
				WithCobra(cmd, args),
			)
			require.NoError(t, err)
			require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

			cfg, err := p.Parse()
			require.NoError(t, err)
			require.Equal(t, "from-profile-prod", cfg.Redis.Host)
			return nil
		},
	}

	// Ensure flags exist so Cobra accepts them.
	psLayer, err := cli.NewProfileSettingsLayer()
	require.NoError(t, err)
	require.NoError(t, psLayer.(schema.CobraSection).AddLayerToCobraCommand(rootCmd))

	// Also add the redis layer flags (not strictly needed for this test, but keeps the pattern consistent).
	require.NoError(t, layer.(schema.CobraSection).AddLayerToCobraCommand(rootCmd))

	rootCmd.SetArgs([]string{"--profile", "prod"})
	require.NoError(t, rootCmd.Execute())
}

func TestWithProfile_Precedence_FlagsOverrideEnvConfigProfilesDefaults(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	profilesPath := filepath.Join(dir, "profiles.yaml")
	require.NoError(t, os.WriteFile(profilesPath, []byte(`default:
  redis:
    host: from-profile
`), 0o644))

	cfgPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`redis:
  host: from-config
`), 0o644))

	t.Setenv("MYAPP_HOST", "from-env")

	rootCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			type app struct {
				Redis testRedisSettings
			}

			p, err := NewParser[app](
				WithDefaults(),
				WithProfile("myapp",
					WithProfileDefaultFile(profilesPath),
					WithProfileFile(profilesPath),
				),
				WithConfigFiles(cfgPath),
				WithEnv("MYAPP"),
				WithCobra(cmd, args),
			)
			require.NoError(t, err)
			require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

			cfg, err := p.Parse()
			require.NoError(t, err)
			require.Equal(t, "from-flag", cfg.Redis.Host)
			return nil
		},
	}

	// Ensure flags exist so Cobra accepts them.
	require.NoError(t, layer.(schema.CobraSection).AddLayerToCobraCommand(rootCmd))
	psLayer, err := cli.NewProfileSettingsLayer()
	require.NoError(t, err)
	require.NoError(t, psLayer.(schema.CobraSection).AddLayerToCobraCommand(rootCmd))

	rootCmd.SetArgs([]string{"--host", "from-flag"})
	require.NoError(t, rootCmd.Execute())
}

func TestWithProfile_MissingDefaultFile_DefaultProfile_Skips(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	// Ensure the default file path does not exist.
	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.yaml")

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(missing),
			WithProfileFile(missing),
		),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	cfg, err := p.Parse()
	require.NoError(t, err)
	require.Equal(t, "from-default", cfg.Redis.Host)
}

func TestWithProfile_MissingDefaultFile_NonDefaultProfile_Errors(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	missing := filepath.Join(dir, "does-not-exist.yaml")

	t.Setenv("MYAPP_PROFILE", "prod")

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(missing),
			WithProfileFile(missing),
		),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	_, err = p.Parse()
	require.Error(t, err)
}

func TestWithProfile_MissingExplicitProfileFile_Errors(t *testing.T) {
	const redisSlug LayerSlug = "redis"
	layer := newTestRedisLayer("from-default")

	dir := t.TempDir()
	defaultMissing := filepath.Join(dir, "default-profiles.yaml")
	explicitMissing := filepath.Join(dir, "explicit-profiles.yaml")

	// Simulate user pointing to a non-default profile file.
	t.Setenv("MYAPP_PROFILE_FILE", explicitMissing)

	type app struct {
		Redis testRedisSettings
	}

	p, err := NewParser[app](
		WithDefaults(),
		WithProfile("myapp",
			WithProfileDefaultFile(defaultMissing),
		),
	)
	require.NoError(t, err)
	require.NoError(t, p.Register(redisSlug, layer, func(t *app) any { return &t.Redis }))

	_, err = p.Parse()
	require.Error(t, err)
}
