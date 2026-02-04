package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
)

type Settings struct {
	ApiKey    string `glazed:"api-key"`
	Threshold int    `glazed:"threshold"`
}

type Command struct{ *cmds.CommandDescription }

func NewCommand() (*Command, error) {
	demo, err := schema.NewSection(
		"demo",
		"Overlay override demo",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New("api-key", fields.TypeString, fields.WithHelp("API key")),
			fields.New("threshold", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Threshold")),
		),
	)
	if err != nil {
		return nil, err
	}
	desc := cmds.NewCommandDescription("overlay-override", cmds.WithShort("--config-file + <base>.override.yaml pattern"), cmds.WithSections(demo))
	return &Command{desc}, nil
}

var _ cmds.BareCommand = &Command{}

func (c *Command) Run(ctx context.Context, vals *values.Values) error {
	s := &Settings{}
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
	root := &cobra.Command{Use: "overlay-override"}
	cmd, err := NewCommand()
	if err != nil {
		panic(err)
	}

	// Config files resolver: start from --config-file if provided, then add sibling <base>.override.yaml
	resolver := func(parsed *values.Values, _ *cobra.Command, _ []string) ([]string, error) {
		cs := &cli.CommandSettings{}
		_ = parsed.DecodeSectionInto(cli.CommandSettingsSlug, cs)
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
