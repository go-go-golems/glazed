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
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type DemoSettings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type DemoBareCommand struct {
	*cmds.CommandDescription
}

func NewDemoBareCommand() (*DemoBareCommand, error) {
	demoSection, err := schema.NewSection(
		"demo",
		"Demo settings",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New(
				"api-key",
				fields.TypeString,
				fields.WithHelp("API key from config/env/flags"),
			),
			fields.New(
				"threshold",
				fields.TypeInteger,
				fields.WithDefault(10),
				fields.WithHelp("Numeric threshold"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	desc := cmds.NewCommandDescription(
		"demo",
		cmds.WithShort("Minimal custom layer with single config file"),
		cmds.WithLayersList(demoSection),
	)

	return &DemoBareCommand{CommandDescription: desc}, nil
}

var _ cmds.BareCommand = &DemoBareCommand{}

func (c *DemoBareCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &DemoSettings{}
	if err := values.DecodeSectionInto(vals, "demo", s); err != nil {
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
	root := &cobra.Command{Use: "config-single"}

	demo, err := NewDemoBareCommand()
	if err != nil {
		panic(err)
	}

	// Use a single explicit config path for simplicity
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		demo,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsLayer: true,
			// Adjust path to your environment if needed
			ConfigPath: "cmd/examples/config-single/config.yaml",
		}),
	)
	if err != nil {
		panic(err)
	}
	root.AddCommand(cobraCmd)

	// validate command: checks the config file against layer definitions
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the config file for known layers, parameters, and types",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Do not print usage or cobra-managed error prefix on failure
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			// Recreate layers like the main command
			demoCmd, err := NewDemoBareCommand()
			if err != nil {
				return err
			}
			// Read config file
			path := "cmd/examples/config-single/config.yaml"
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var raw map[string]interface{}
			if err := yaml.Unmarshal(b, &raw); err != nil {
				return err
			}

			issues := []string{}
			// Validate top-level layers and parameters
			for layerSlug, v := range raw {
				layer, ok := demoCmd.Description().Layers.Get(layerSlug)
				if !ok {
					issues = append(issues, fmt.Sprintf("unknown layer: %s", layerSlug))
					continue
				}
				m, ok := v.(map[string]interface{})
				if !ok {
					issues = append(issues, fmt.Sprintf("layer %s must be an object", layerSlug))
					continue
				}
				pds := layer.GetParameterDefinitions()
				// Build set of known parameter names
				known := map[string]bool{}
				pds.ForEach(func(pd *fields.Definition) {
					known[pd.Name] = true
				})
				for key, val := range m {
					if !known[key] {
						issues = append(issues, fmt.Sprintf("unknown parameter in layer %s: %s", layerSlug, key))
						continue
					}
					pd, _ := pds.Get(key)
					if _, err := pd.CheckValueValidity(val); err != nil {
						issues = append(issues, fmt.Sprintf("invalid value for %s.%s: %v", layerSlug, key, err))
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

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
