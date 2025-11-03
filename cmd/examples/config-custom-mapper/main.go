package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
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
	demoLayer, err := layers.NewParameterLayer(
		"demo",
		"Demo settings",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"api-key",
				parameters.ParameterTypeString,
				parameters.WithHelp("API key from config/env/flags"),
			),
			parameters.NewParameterDefinition(
				"threshold",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(10),
				parameters.WithHelp("Numeric threshold"),
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

func (c *DemoBareCommand) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &DemoSettings{}
	if err := pl.InitializeStruct("demo", s); err != nil {
		return err
	}
	fmt.Printf("api_key=%s threshold=%d\n", s.ApiKey, s.Threshold)
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
			MiddlewaresFunc: func(parsedCommandLayers *layers.ParsedLayers, cmd *cobra.Command, args []string) ([]middlewares.Middleware, error) {
				return []middlewares.Middleware{
					// Highest priority: command-line flags
					middlewares.ParseFromCobraCommand(cmd,
						parameters.WithParseStepSource("flags"),
					),
					// Medium priority: custom config file with mapper
					middlewares.LoadParametersFromFile(
						"cmd/examples/config-custom-mapper/config.yaml",
						middlewares.WithConfigFileMapper(flatConfigMapper),
						middlewares.WithParseOptions(
							parameters.WithParseStepSource("config"),
						),
					),
					// Lowest priority: defaults
					middlewares.SetFromDefaults(
						parameters.WithParseStepSource(parameters.SourceDefaults),
					),
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
				pds := layer.GetParameterDefinitions()
				known := map[string]bool{}
				pds.ForEach(func(pd *parameters.ParameterDefinition) { known[pd.Name] = true })
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
