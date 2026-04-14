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
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
)

type Settings struct {
	ApiKey    string `glazed:"api-key"`
	Threshold int    `glazed:"threshold"`
}

type Command struct{ *cmds.CommandDescription }

func overrideSiblingSource(path string) glazedconfig.SourceSpec {
	return glazedconfig.SourceSpec{
		Name:       "explicit-override-config",
		Layer:      glazedconfig.LayerExplicit,
		SourceKind: "computed-override-file",
		Optional:   true,
		Discover: func(context.Context) ([]string, error) {
			if strings.TrimSpace(path) == "" {
				return nil, nil
			}
			dir := filepath.Dir(path)
			base := filepath.Base(path)
			ext := filepath.Ext(base)
			stem := strings.TrimSuffix(base, ext)
			override := filepath.Join(dir, fmt.Sprintf("%s.override.yaml", stem))
			if _, err := os.Stat(override); err != nil {
				if os.IsNotExist(err) {
					return nil, nil
				}
				return nil, err
			}
			return []string{override}, nil
		},
	}
}

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

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			// Keep command-settings to parse --config-file.
			ConfigPlanBuilder: func(parsed *values.Values, _ *cobra.Command, _ []string) (*glazedconfig.Plan, error) {
				cs := &cli.CommandSettings{}
				_ = parsed.DecodeSectionInto(cli.CommandSettingsSlug, cs)
				return glazedconfig.NewPlan(
					glazedconfig.WithLayerOrder(glazedconfig.LayerExplicit),
				).Add(
					glazedconfig.ExplicitFile(cs.ConfigFile).Named("explicit-config"),
					overrideSiblingSource(cs.ConfigFile),
				), nil
			},
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
