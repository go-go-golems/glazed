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
	ApiKey       string `glazed:"api-key"`
	RequiredName string `glazed:"required-name"`
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
	cmd, getParsed := newExecutableParserCommandForTest(t, parser)
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())
	parsed := getParsed()
	require.NotNil(t, parsed)
	return parsed
}

func newExecutableParserCommandForTest(t *testing.T, parser *CobraParser) (*cobra.Command, func() *values.Values) {
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
	return cmd, func() *values.Values { return parsed }
}

func newRequiredDefaultSectionParserForTest(t *testing.T, fieldOptions ...fields.Option) *CobraParser {
	t.Helper()
	options := append([]fields.Option{fields.WithRequired(true)}, fieldOptions...)
	section, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithFields(fields.New("required-name", fields.TypeString, options...)),
	)
	require.NoError(t, err)

	parser, err := NewCobraParserFromSections(
		schema.NewSchema(schema.WithSections(section)),
		&CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			AppName:           "REQ_ENV_TEST",
		},
	)
	require.NoError(t, err)
	return parser
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

func TestCobraParserRequiredFieldCanBeSatisfiedFromEnv(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t)
	t.Setenv("REQ_ENV_TEST_REQUIRED_NAME", "from-env")

	parsed := executeParserForTest(t, parser, nil)
	fv, ok := parsed.GetField(schema.DefaultSlug, "required-name")
	require.True(t, ok)
	require.Equal(t, "from-env", fv.Value)
	require.NotEmpty(t, fv.Log)
	require.Equal(t, "env", fv.Log[len(fv.Log)-1].Source)
}

func TestCobraParserRequiredFieldCanBeSatisfiedFromConfig(t *testing.T) {
	section, err := schema.NewSection(
		"demo",
		"Demo",
		schema.WithFields(fields.New("required-name", fields.TypeString, fields.WithRequired(true))),
	)
	require.NoError(t, err)
	schema_ := schema.NewSchema(schema.WithSections(section))

	cfgFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("demo:\n  required-name: from-config\n"), 0o644))

	parser, err := NewCobraParserFromSections(schema_, &CobraParserConfig{
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
	require.Equal(t, "from-config", settings.RequiredName)
}

func TestCobraParserRequiredFieldMissingFailsNormally(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t)
	cmd, _ := newExecutableParserCommandForTest(t, parser)

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required-name")
}

func TestCobraParserRequiredFieldEmptyDefaultFailsNormally(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t, fields.WithDefault(""))
	cmd, _ := newExecutableParserCommandForTest(t, parser)

	err := cmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "required-name")
}

func TestCobraParserRequiredFieldFlagOverridesEnv(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t)
	t.Setenv("REQ_ENV_TEST_REQUIRED_NAME", "from-env")

	parsed := executeParserForTest(t, parser, []string{"--required-name", "from-flag"})
	fv, ok := parsed.GetField(schema.DefaultSlug, "required-name")
	require.True(t, ok)
	require.Equal(t, "from-flag", fv.Value)
	require.NotEmpty(t, fv.Log)
	require.Equal(t, "cobra", fv.Log[len(fv.Log)-1].Source)
}

func TestCobraParserRequiredFieldMissingDoesNotFailForPrintParsedFields(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t)

	parsed := executeParserForTest(t, parser, []string{"--print-parsed-fields"})
	require.NotNil(t, parsed)
	_, ok := parsed.GetField(schema.DefaultSlug, "required-name")
	require.False(t, ok)
}

func TestCobraParserRequiredFieldMissingDoesNotFailForHelp(t *testing.T) {
	parser := newRequiredDefaultSectionParserForTest(t)
	cmd, getParsed := newExecutableParserCommandForTest(t, parser)
	cmd.SetArgs([]string{"--help"})

	require.NoError(t, cmd.Execute())
	require.Nil(t, getParsed(), "Cobra help should not run the command parser")
}
