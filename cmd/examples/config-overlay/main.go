package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	cmdmw "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type OverlaySettings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type OverlayCommand struct{ *cmds.CommandDescription }

func NewOverlayCommand() (*OverlayCommand, error) {
	demo, err := layers.NewParameterLayer(
		"demo",
		"Overlay demo",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString, parameters.WithHelp("API key")),
			parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger, parameters.WithDefault(10), parameters.WithHelp("Threshold")),
		),
	)
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription("overlay", cmds.WithShort("Multiple config overlays"), cmds.WithLayersList(demo))
	return &OverlayCommand{desc}, nil
}

var _ cmds.BareCommand = &OverlayCommand{}

func (c *OverlayCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &OverlaySettings{}
	if err := pl.InitializeStruct("demo", s); err != nil {
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

	// Provide a custom files resolver returning low->high precedence files
	filesResolver := func(_ *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
		return []string{
			"cmd/examples/config-overlay/base.yaml",
			"cmd/examples/config-overlay/env.yaml",
			"cmd/examples/config-overlay/local.yaml",
		}, nil
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		overlay,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsLayer: true,
			ConfigFilesFunc:          filesResolver,
		}),
	)
	if err != nil {
		panic(err)
	}

	// For demonstration, add an extra command to show parsed steps by printing the map
	// Users can also run with env overrides: e.g. DEMO_API_KEY and demo threshold flags
	_ = cmdmw.UpdateFromEnv // ensure middleware is referenced (no-op here)

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
				for layerSlug, v := range raw {
					layer, ok := overlayCmd.Description().Layers.Get(layerSlug)
					if !ok {
						issues = append(issues, fmt.Sprintf("%s: unknown layer %s", f, layerSlug))
						continue
					}
					m, ok := v.(map[string]interface{})
					if !ok {
						issues = append(issues, fmt.Sprintf("%s: layer %s must be an object", f, layerSlug))
						continue
					}
					pds := layer.GetParameterDefinitions()
					known := map[string]bool{}
					pds.ForEach(func(pd *parameters.ParameterDefinition) { known[pd.Name] = true })
					for key, val := range m {
						if !known[key] {
							issues = append(issues, fmt.Sprintf("%s: unknown parameter %s.%s", f, layerSlug, key))
							continue
						}
						pd, _ := pds.Get(key)
						if _, err := pd.CheckValueValidity(val); err != nil {
							issues = append(issues, fmt.Sprintf("%s: invalid value for %s.%s: %v", f, layerSlug, key, err))
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
