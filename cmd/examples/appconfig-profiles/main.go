package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/appconfig"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// This example exists to make debugging profile bootstrap behavior easy and local:
// - it creates a temp profiles.yaml with `default` + `prod` profiles
// - it wires appconfig.WithProfile into a small cobra command
// - it prints the final resolved settings (and where the generated files live)
//
// Try:
//
//	MYAPP_PROFILE=prod go run ./cmd/examples/appconfig-profiles
//	go run ./cmd/examples/appconfig-profiles --profile prod
//	MYAPP_HOST=from-env go run ./cmd/examples/appconfig-profiles
//	go run ./cmd/examples/appconfig-profiles --host from-flag
func main() {
	tmpDir, err := os.MkdirTemp("", "glazed-appconfig-profiles-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}

	profilesPath := filepath.Join(tmpDir, "profiles.yaml")
	profilesYAML := `default:
  redis:
    host: from-profile-default
prod:
  redis:
    host: from-profile-prod
`
	if err := os.WriteFile(profilesPath, []byte(profilesYAML), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write profiles.yaml: %v\n", err)
		os.Exit(1)
	}

	redisLayer := mustSection(schema.NewSection(
		"redis",
		"Redis",
		schema.WithFields(
			fields.New("host", fields.TypeString, fields.WithDefault("from-default")),
		),
	))

	root := &cobra.Command{
		Use:   "appconfig-profiles",
		Short: "Minimal appconfig.WithProfile example",
		RunE: func(cmd *cobra.Command, args []string) error {
			type AppSettings struct {
				Redis struct {
					Host string `glazed.parameter:"host"`
				}
			}

			useConfig, err := cmd.Flags().GetBool("use-config")
			if err != nil {
				return err
			}

			// Optional config file that demonstrates config overriding profiles.
			// If enabled, it sets redis.host=from-config (which should beat profiles).
			var cfgPath string
			if useConfig {
				cfgPath = filepath.Join(tmpDir, "config.yaml")
				cfgYAML := `profile-settings:
  profile: prod
redis:
  host: from-config
`
				if err := os.WriteFile(cfgPath, []byte(cfgYAML), 0o644); err != nil {
					return errors.Wrap(err, "failed to write config.yaml")
				}
			}

			opts := []appconfig.ParserOption{
				appconfig.WithDefaults(),
				appconfig.WithProfile("myapp",
					// Avoid touching ~/.config during example runs; keep everything self-contained.
					appconfig.WithProfileDefaultFile(profilesPath),
					appconfig.WithProfileFile(profilesPath),
				),
			}
			if useConfig {
				opts = append(opts, appconfig.WithConfigFiles(cfgPath))
			}
			opts = append(opts,
				// General settings overrides.
				appconfig.WithEnv("MYAPP"),
				appconfig.WithCobra(cmd, args),
			)

			parser, err := appconfig.NewParser[AppSettings](opts...)
			if err != nil {
				return err
			}

			if err := parser.Register("redis", redisLayer, func(t *AppSettings) any { return &t.Redis }); err != nil {
				return err
			}

			cfg, err := parser.Parse()
			if err != nil {
				return err
			}

			fmt.Printf("tmp_dir=%s\n", tmpDir)
			fmt.Printf("profiles_file=%s\n", profilesPath)
			if useConfig {
				fmt.Printf("config_file=%s\n", cfgPath)
			}
			fmt.Printf("redis.host=%s\n", cfg.Redis.Host)
			return nil
		},
	}

	root.Flags().Bool("use-config", false, "Enable a generated config.yaml that selects prod and overrides redis.host")

	// Add flags up front so Cobra accepts them before RunE executes.
	_ = addLayer(root, redisLayer)
	if psLayer, err := cli.NewProfileSettingsLayer(); err == nil {
		_ = addLayer(root, psLayer)
	}

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func mustSection(section *schema.SectionImpl, err error) schema.Section {
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create section: %v\n", err)
		os.Exit(1)
	}
	return section
}

func addLayer(cmd *cobra.Command, layer schema.Section) error {
	cobraLayer, ok := layer.(schema.CobraSection)
	if !ok {
		return errors.Errorf("layer %s is not a CobraSection", layer.GetSlug())
	}
	return cobraLayer.AddLayerToCobraCommand(cmd)
}
