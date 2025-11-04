package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/spf13/cobra"
)

type Settings struct {
	ApiKey    string `glazed.parameter:"api-key"`
	Threshold int    `glazed.parameter:"threshold"`
}

type Command struct{ *cmds.CommandDescription }

func NewCommand() (*Command, error) {
	demo, err := layers.NewParameterLayer(
		"demo",
		"Overlay override demo",
		layers.WithPrefix("demo-"),
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString, parameters.WithHelp("API key")),
			parameters.NewParameterDefinition("threshold", parameters.ParameterTypeInteger, parameters.WithDefault(10), parameters.WithHelp("Threshold")),
		),
	)
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription("overlay-override", cmds.WithShort("--config-file + <base>.override.yaml pattern"), cmds.WithLayersList(demo))
	return &Command{desc}, nil
}

var _ cmds.BareCommand = &Command{}

func (c *Command) Run(ctx context.Context, pl *layers.ParsedLayers) error {
	s := &Settings{}
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
	root := &cobra.Command{Use: "overlay-override"}
	cmd, err := NewCommand()
	if err != nil {
		panic(err)
	}

	// Config files resolver: start from --config-file if provided, then add sibling <base>.override.yaml
	resolver := func(parsed *layers.ParsedLayers, _ *cobra.Command, _ []string) ([]string, error) {
		cs := &cli.CommandSettings{}
		_ = parsed.InitializeStruct(cli.CommandSettingsSlug, cs)
		files := []string{}
		if cs.ConfigFile != "" {
			files = append(files, cs.ConfigFile)
			dir := filepath.Dir(cs.ConfigFile)
			base := filepath.Base(cs.ConfigFile)
			ext := filepath.Ext(base)
			stem := strings.TrimSuffix(base, ext)
			override := filepath.Join(dir, fmt.Sprintf("%s.override.yaml", stem))
			if _, err := os.Stat(override); err == nil {
				files = append(files, override)
			}
		}
		return files, nil
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			// Keep command-settings to parse --config-file
			ConfigFilesFunc: resolver,
		}),
	)
	if err != nil {
		panic(err)
	}

	root.AddCommand(cobraCmd)
	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
