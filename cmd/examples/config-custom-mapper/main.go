package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/pkg/errors"
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
	demoLayer, err := schema.NewSection(
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
		cmds.WithShort("Custom config file mapper example"),
		cmds.WithLayersList(demoLayer),
	)

	return &DemoBareCommand{CommandDescription: desc}, nil
}

var _ cmds.BareCommand = &DemoBareCommand{}

func (c *DemoBareCommand) Run(ctx context.Context, vals *values.Values) error {
	s := &DemoSettings{}
	if err := vals.InitializeStruct("demo", s); err != nil {
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

// flatConfigMapper transforms a flat config structure to the layer map format.
// Example input: {"api_key": "secret", "threshold": 5}
// Example output: {"demo": {"api-key": "secret", "threshold": 5}}
//
// Also handles triple-nested structures like:
// {"app": {"settings": {"api": {"key": "secret"}}}}
func flatConfigMapper(rawConfig interface{}) (map[string]map[string]interface{}, error) {
	configMap, ok := rawConfig.(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("expected map[string]interface{}, got %T", rawConfig)
	}

	result := map[string]map[string]interface{}{
		"demo": make(map[string]interface{}),
	}

	// Map flat keys to layer parameters
	for key, value := range configMap {
		switch key {
		case "api_key":
			result["demo"]["api-key"] = value
		case "threshold":
			result["demo"]["threshold"] = value
		case "app":
			// Handle triple-nested structure: app.settings.api.key
			if appMap, ok := value.(map[string]interface{}); ok {
				if settingsMap, ok := appMap["settings"].(map[string]interface{}); ok {
					if apiMap, ok := settingsMap["api"].(map[string]interface{}); ok {
						if apiKey, ok := apiMap["key"]; ok {
							result["demo"]["api-key"] = apiKey
						}
					}
				}
			}
		// Ignore unknown keys
		default:
			continue
		}
	}

	return result, nil
}

func main() {
	root := &cobra.Command{Use: "config-custom-mapper"}

	demo, err := NewDemoBareCommand()
	if err != nil {
		panic(err)
	}

	// Build command with custom middleware that uses a custom config mapper
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		demo,
		cli.WithParserConfig(cli.CobraParserConfig{
			SkipCommandSettingsLayer: true,
			MiddlewaresFunc: func(parsedCommandLayers *values.Values, cmd *cobra.Command, args []string) ([]sources.Middleware, error) {
				return []sources.Middleware{
					// Highest priority: command-line flags
					sources.FromCobra(cmd, fields.WithSource("flags")),
					// Medium priority: custom config file with mapper
					sources.FromFile(
						"cmd/examples/config-custom-mapper/config.yaml",
						sources.WithConfigFileMapper(flatConfigMapper),
						sources.WithParseOptions(fields.WithSource("config")),
					),
					// Lowest priority: defaults
					sources.FromDefaults(fields.WithSource("defaults")),
				}, nil
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	root.AddCommand(cobraCmd)

	// validate command: validate config.yaml using the custom mapper and layer definitions
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate the custom-mapped config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Silence usage and cobra error prefix on failure for clean output
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true
			demo, err := NewDemoBareCommand()
			if err != nil {
				return err
			}
			path := "cmd/examples/config-custom-mapper/config.yaml"
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var raw interface{}
			if err := yaml.Unmarshal(b, &raw); err != nil {
				return err
			}
			mapped, err := flatConfigMapper(raw)
			if err != nil {
				return err
			}
			issues := []string{}
			// Validate mapped structure against known layers and params
			for layerSlug, kv := range mapped {
				layer, ok := demo.Description().Layers.Get(layerSlug)
				if !ok {
					issues = append(issues, fmt.Sprintf("unknown layer: %s", layerSlug))
					continue
				}
				pmap := kv
				pds := layer.GetDefinitions()
				known := map[string]bool{}
				pds.ForEach(func(pd *fields.Definition) { known[pd.Name] = true })
				for key, val := range pmap {
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
