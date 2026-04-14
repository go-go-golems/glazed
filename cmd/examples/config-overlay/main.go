package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type OverlaySettings struct {
	ApiKey    string `glazed:"api-key"`
	Threshold int    `glazed:"threshold"`
}

type OverlayCommand struct{ *cmds.CommandDescription }

func exampleFileSource(name string, layer glazedconfig.ConfigLayer, path string) glazedconfig.SourceSpec {
	return glazedconfig.SourceSpec{
		Name:       name,
		Layer:      layer,
		SourceKind: "example-config",
		Optional:   true,
		Discover: func(context.Context) ([]string, error) {
			if _, err := os.Stat(path); err != nil {
				if os.IsNotExist(err) {
					return nil, nil
				}
				return nil, err
			}
			return []string{path}, nil
		},
	}
}

func NewOverlayCommand() (*OverlayCommand, error) {
	demo, err := schema.NewSection(
		"demo",
		"Overlay demo",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New("api-key", fields.TypeString, fields.WithHelp("API key")),
			fields.New("threshold", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Threshold")),
		),
	)
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription("overlay", cmds.WithShort("Multiple config overlays"), cmds.WithSections(demo))
	return &OverlayCommand{desc}, nil
}

var _ cmds.BareCommand = &OverlayCommand{}

func (c *OverlayCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &OverlaySettings{}
	if err := vals.DecodeSectionInto("demo", s); err != nil {
		return err
	}
	// Censor API key for security
	apiKeyMasked := "***"
	if len(s.ApiKey) > 0 {
		if len(s.ApiKey) > 4 {
			apiKeyMasked = s.ApiKey[:4] + "***"
		} else {
			apiKeyMasked = "***"
		}
	}
	fmt.Printf("api_key=%s threshold=%d\n", apiKeyMasked, s.Threshold)
	return nil
}

func main() {
	root := &cobra.Command{Use: "config-overlay"}
	overlay, err := NewOverlayCommand()
	if err != nil {
		panic(err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		overlay,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsSection: true,
			ConfigPlanBuilder: func(_ *values.Values, _ *cobra.Command, _ []string) (*glazedconfig.Plan, error) {
				return glazedconfig.NewPlan(
					glazedconfig.WithLayerOrder(glazedconfig.LayerSystem, glazedconfig.LayerUser, glazedconfig.LayerCWD),
				).Add(
					exampleFileSource("base-config", glazedconfig.LayerSystem, "cmd/examples/config-overlay/base.yaml"),
					exampleFileSource("env-config", glazedconfig.LayerUser, "cmd/examples/config-overlay/env.yaml"),
					exampleFileSource("local-config", glazedconfig.LayerCWD, "cmd/examples/config-overlay/local.yaml"),
				), nil
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	// validate command: validates each overlay file individually
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the overlay config files (per-file validation)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Silence Cobra usage and error prefix for cleaner validator output
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			overlayCmd, err := NewOverlayCommand()
			if err != nil {
				return err
			}
			files := []string{
				"cmd/examples/config-overlay/base.yaml",
				"cmd/examples/config-overlay/env.yaml",
				"cmd/examples/config-overlay/local.yaml",
			}
			issues := []string{}
			for _, f := range files {
				b, err := os.ReadFile(f)
				if err != nil {
					issues = append(issues, fmt.Sprintf("%s: %v", f, err))
					continue
				}
				var raw map[string]interface{}
				if err := yaml.Unmarshal(b, &raw); err != nil {
					issues = append(issues, fmt.Sprintf("%s: %v", f, err))
					continue
				}
				for sectionSlug, v := range raw {
					section, ok := overlayCmd.Description().Schema.Get(sectionSlug)
					if !ok {
						issues = append(issues, fmt.Sprintf("%s: unknown section %s", f, sectionSlug))
						continue
					}
					m, ok := v.(map[string]interface{})
					if !ok {
						issues = append(issues, fmt.Sprintf("%s: section %s must be an object", f, sectionSlug))
						continue
					}
					pds := section.GetDefinitions()
					known := map[string]bool{}
					pds.ForEach(func(pd *fields.Definition) { known[pd.Name] = true })
					for key, val := range m {
						if !known[key] {
							issues = append(issues, fmt.Sprintf("%s: unknown field %s.%s", f, sectionSlug, key))
							continue
						}
						pd, _ := pds.Get(key)
						if _, err := pd.CheckValueValidity(val); err != nil {
							issues = append(issues, fmt.Sprintf("%s: invalid value for %s.%s: %v", f, sectionSlug, key, err))
						}
					}
				}
			}
			if len(issues) > 0 {
				for _, i := range issues {
					fmt.Println(i)
				}
				return fmt.Errorf("validation failed")
			}
			fmt.Println("OK")
			return nil
		},
	}
	root.AddCommand(validateCmd)

	root.AddCommand(cobraCmd)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
