package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/sources"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type DemoSettings struct {
	APIKey    string `glazed:"api-key" yaml:"api-key"`
	Threshold int    `glazed:"threshold" yaml:"threshold"`
	Profile   string `glazed:"profile" yaml:"profile"`
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "config-plan",
		Short: "Demonstrate declarative config plans and provenance-aware loading",
	}

	var explicit string
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Resolve the config plan, load the files, and print settings + parse history",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(explicit)
		},
	}
	showCmd.Flags().StringVar(&explicit, "explicit", "", "Optional explicit config file applied last")
	rootCmd.AddCommand(showCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runShow(explicit string) error {
	plan := glazedconfig.NewPlan(
		glazedconfig.WithLayerOrder(
			glazedconfig.LayerRepo,
			glazedconfig.LayerCWD,
			glazedconfig.LayerExplicit,
		),
		glazedconfig.WithDedupePaths(),
	).Add(
		glazedconfig.GitRootFile("cmd/examples/config-plan/repo.yaml").
			Named("repo-example-config").
			InLayer(glazedconfig.LayerRepo).
			Kind("example-config"),
		glazedconfig.WorkingDirFile("local.yaml").
			Named("cwd-example-config").
			InLayer(glazedconfig.LayerCWD).
			Kind("example-config"),
		glazedconfig.ExplicitFile(explicit).
			Named("explicit-example-config").
			InLayer(glazedconfig.LayerExplicit).
			Kind("explicit-file"),
	)

	files, report, err := plan.Resolve(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("== Resolved config plan ==")
	fmt.Println(report.String())
	fmt.Println()

	demoSection, err := schema.NewSection(
		"demo",
		"Demo settings",
		schema.WithFields(
			fields.New("api-key", fields.TypeString, fields.WithHelp("API key")),
			fields.New("threshold", fields.TypeInteger, fields.WithDefault(10), fields.WithHelp("Threshold")),
			fields.New("profile", fields.TypeString, fields.WithHelp("Selected profile label")),
		),
	)
	if err != nil {
		return err
	}

	parsed := values.New()
	if err := sources.Execute(
		schema.NewSchema(schema.WithSections(demoSection)),
		parsed,
		sources.FromResolvedFiles(files),
		sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
	); err != nil {
		return err
	}

	settings := &DemoSettings{}
	if err := parsed.DecodeSectionInto("demo", settings); err != nil {
		return err
	}

	fmt.Println("== Final settings ==")
	if err := writeYAML(settings); err != nil {
		return err
	}
	fmt.Println()

	fmt.Println("== Parsed fields with provenance ==")
	return writeYAML(parsed)
}

func writeYAML(v interface{}) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(v)
}
