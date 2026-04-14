package cli

import (
	"os"
	"path/filepath"
	"testing"

	fields "github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

type cobraParserDemoSettings struct {
	ApiKey string `glazed:"api-key"`
}

func newCobraParserDemoSection(t *testing.T, defaultValue string) schema.Section {
	t.Helper()
	section, err := schema.NewSection(
		"demo",
		"Demo",
		schema.WithPrefix("demo-"),
		schema.WithFields(
			fields.New("api-key", fields.TypeString, fields.WithDefault(defaultValue)),
		),
	)
	require.NoError(t, err)
	return section
}

func executeParserForTest(t *testing.T, parser *CobraParser, args []string) *values.Values {
	t.Helper()
	var parsed *values.Values
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			parsed, err = parser.Parse(cmd, args)
			return err
		},
	}
	require.NoError(t, parser.AddToCobraCommand(cmd))
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())
	require.NotNil(t, parsed)
	return parsed
}

func TestCobraParserConfigPlanBuilderLoadsConfigFiles(t *testing.T) {
	section := newCobraParserDemoSection(t, "from-default")
	schema_ := schema.NewSchema(schema.WithSections(section))

	cfgFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("demo:\n  api-key: from-file\n"), 0o644))

	parser, err := NewCobraParserFromSections(schema_, &CobraParserConfig{
		SkipCommandSettingsSection: true,
		ConfigPlanBuilder: func(_ *values.Values, _ *cobra.Command, _ []string) (*glazedconfig.Plan, error) {
			return glazedconfig.NewPlan(
				glazedconfig.WithLayerOrder(glazedconfig.LayerExplicit),
			).Add(
				glazedconfig.ExplicitFile(cfgFile).Named("test-config"),
			), nil
		},
	})
	require.NoError(t, err)

	parsed := executeParserForTest(t, parser, nil)
	settings := &cobraParserDemoSettings{}
	require.NoError(t, parsed.DecodeSectionInto("demo", settings))
	require.Equal(t, "from-file", settings.ApiKey)
}

func TestCobraParserDoesNotImplicitlyLoadConfigFileWithoutPlanBuilder(t *testing.T) {
	section := newCobraParserDemoSection(t, "from-default")
	schema_ := schema.NewSchema(schema.WithSections(section))

	cfgFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("demo:\n  api-key: from-file\n"), 0o644))

	parser, err := NewCobraParserFromSections(schema_, &CobraParserConfig{
		AppName: "myapp",
	})
	require.NoError(t, err)

	parsed := executeParserForTest(t, parser, []string{"--config-file", cfgFile})
	settings := &cobraParserDemoSettings{}
	require.NoError(t, parsed.DecodeSectionInto("demo", settings))
	require.Equal(t, "from-default", settings.ApiKey)
}
