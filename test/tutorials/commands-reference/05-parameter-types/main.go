package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// ParameterTypesCommand demonstrates all parameter types from the documentation
type ParameterTypesCommand struct {
	*cmds.CommandDescription
}

// ParameterTypesSettings demonstrates various parameter types
type ParameterTypesSettings struct {
	// Basic types
	Name        string    `glazed.parameter:"name"`
	Secret      string    `glazed.parameter:"secret"`
	Count       int       `glazed.parameter:"count"`
	Ratio       float64   `glazed.parameter:"ratio"`
	Enabled     bool      `glazed.parameter:"enabled"`
	CreatedDate time.Time `glazed.parameter:"created-date"`

	// Collection types
	Tags      []string  `glazed.parameter:"tags"`
	Ports     []int     `glazed.parameter:"ports"`
	Weights   []float64 `glazed.parameter:"weights"`

	// Choice types
	Environment string   `glazed.parameter:"environment"`
	Features    []string `glazed.parameter:"features"`

	// File types (simplified)
	ConfigFile string `glazed.parameter:"config-file"`

	// Special types
	EnvVars  map[string]string `glazed.parameter:"env-vars"`
	MaxRetry int               `glazed.parameter:"max-retry"`
}

func (c *ParameterTypesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ParameterTypesSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Create a row showing all the parsed parameter values
	row := types.NewRow(
		// Basic types
		types.MRP("name", s.Name),
		types.MRP("secret_length", len(s.Secret)), // Don't expose secret value
		types.MRP("count", s.Count),
		types.MRP("ratio", s.Ratio),
		types.MRP("enabled", s.Enabled),
		types.MRP("created_date", s.CreatedDate),

		// Collection types
		types.MRP("tags", s.Tags),
		types.MRP("ports", s.Ports),
		types.MRP("weights", s.Weights),

		// Choice types
		types.MRP("environment", s.Environment),
		types.MRP("features", s.Features),

		// File types (simplified)
		types.MRP("config_file", s.ConfigFile),

		// Special types
		types.MRP("env_vars", s.EnvVars),
		types.MRP("max_retry", s.MaxRetry),
	)

	return gp.AddRow(ctx, row)
}

// NewParameterTypesCommand creates a comprehensive parameter types demo command
func NewParameterTypesCommand() (*ParameterTypesCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Create some test files for file parameter demonstration
	createTestFiles()

	cmdDesc := cmds.NewCommandDescription(
		"param-types",
		cmds.WithShort("Demonstrate all Glazed parameter types"),
		cmds.WithLong("Comprehensive demonstration of all parameter types available in Glazed"),
		cmds.WithFlags(
			// Basic Types
			parameters.NewParameterDefinition(
				"name",
				parameters.ParameterTypeString,
				parameters.WithDefault("demo-app"),
				parameters.WithHelp("Application name (ParameterTypeString)"),
			),
			parameters.NewParameterDefinition(
				"secret",
				parameters.ParameterTypeSecret,
				parameters.WithDefault("my-secret-key"),
				parameters.WithHelp("Secret value - masked in help and logs (ParameterTypeSecret)"),
			),
			parameters.NewParameterDefinition(
				"count",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(42),
				parameters.WithHelp("Number of items (ParameterTypeInteger)"),
			),
			parameters.NewParameterDefinition(
				"ratio",
				parameters.ParameterTypeFloat,
				parameters.WithDefault(3.14),
				parameters.WithHelp("Ratio value (ParameterTypeFloat)"),
			),
			parameters.NewParameterDefinition(
				"enabled",
				parameters.ParameterTypeBool,
				parameters.WithDefault(true),
				parameters.WithHelp("Enable feature (ParameterTypeBool)"),
			),
			parameters.NewParameterDefinition(
				"created-date",
				parameters.ParameterTypeDate,
				parameters.WithDefault(time.Now()),
				parameters.WithHelp("Creation date (ParameterTypeDate)"),
			),

			// Collection Types
			parameters.NewParameterDefinition(
				"tags",
				parameters.ParameterTypeStringList,
				parameters.WithDefault([]string{"web", "api", "production"}),
				parameters.WithHelp("List of tags (ParameterTypeStringList)"),
			),
			parameters.NewParameterDefinition(
				"ports",
				parameters.ParameterTypeIntegerList,
				parameters.WithDefault([]int{80, 443, 8080}),
				parameters.WithHelp("List of ports (ParameterTypeIntegerList)"),
			),
			parameters.NewParameterDefinition(
				"weights",
				parameters.ParameterTypeFloatList,
				parameters.WithDefault([]float64{1.0, 2.5, 3.14}),
				parameters.WithHelp("List of weights (ParameterTypeFloatList)"),
			),

			// Choice Types
			parameters.NewParameterDefinition(
				"environment",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("development", "staging", "production"),
				parameters.WithDefault("development"),
				parameters.WithHelp("Deployment environment (ParameterTypeChoice)"),
			),
			parameters.NewParameterDefinition(
				"features",
				parameters.ParameterTypeChoiceList,
				parameters.WithChoices("auth", "logging", "metrics", "caching", "rate-limiting"),
				parameters.WithDefault([]string{"auth", "logging"}),
				parameters.WithHelp("Enabled features (ParameterTypeChoiceList)"),
			),

			// File Types (simplified for demo - complex file types need special setup)
			parameters.NewParameterDefinition(
				"config-file",
				parameters.ParameterTypeString,
				parameters.WithDefault("./test-config.yaml"),
				parameters.WithHelp("Configuration file path (simplified for demo)"),
			),

			// Special Types
			parameters.NewParameterDefinition(
				"env-vars",
				parameters.ParameterTypeKeyValue,
				parameters.WithDefault(map[string]string{
					"DATABASE_URL": "postgres://localhost:5432/mydb",
					"DEBUG":        "true",
				}),
				parameters.WithHelp("Environment variables (ParameterTypeKeyValue)"),
			),

			// With short flags and validation
			parameters.NewParameterDefinition(
				"max-retry",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(3),
				parameters.WithShortFlag("r"),
				parameters.WithHelp("Maximum retry attempts (with short flag -r)"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &ParameterTypesCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// createTestFiles creates test files for file parameter demonstration
func createTestFiles() {
	// Create test configuration file
	configContent := `# Test configuration
app:
  name: demo-app
  version: 1.0.0
database:
  host: localhost
  port: 5432
`
	os.WriteFile("./test-config.yaml", []byte(configContent), 0644)

	// Create test data files
	data1 := `{"id": 1, "name": "item1", "value": 100}`
	data2 := `{"id": 2, "name": "item2", "value": 200}`
	os.WriteFile("./test-data1.json", []byte(data1), 0644)
	os.WriteFile("./test-data2.json", []byte(data2), 0644)

	// Create test input file
	inputText := `This is test input content.
It has multiple lines.
Perfect for demonstrating ParameterTypeStringFromFile.
`
	os.WriteFile("./test-input.txt", []byte(inputText), 0644)
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ParameterTypesCommand{}

func main() {
	cmd, err := NewParameterTypesCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "param-types-demo",
		Short: "Demonstration of all Glazed parameter types",
		Long: `This demonstrates all parameter types available in Glazed:

Basic Types:
- String, Secret, Integer, Float, Bool, Date

Collection Types:
- StringList, IntegerList, FloatList

Choice Types:
- Choice (single), ChoiceList (multiple)

File Types:
- File, FileList, StringFromFile, StringListFromFile

Special Types:
- KeyValue (map-like)`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add comprehensive examples
	cobraCmd.Example = `  # Use defaults
  param-types-demo param-types

  # Override various parameters
  param-types-demo param-types --name "my-app" --count 100 --ratio 2.718

  # Work with lists
  param-types-demo param-types --tags web --tags api --tags staging --ports 80 --ports 8080

  # Choice parameters
  param-types-demo param-types --environment production --features auth --features metrics

  # Key-value parameters
  param-types-demo param-types --env-vars DATABASE_URL=postgres://... --env-vars DEBUG=false

  # Short flags
  param-types-demo param-types -r 5

  # JSON output to see all parsed values
  param-types-demo param-types --output json`

	// Cleanup test files on exit
	defer func() {
		os.Remove("./test-config.yaml")
		os.Remove("./test-data1.json")
		os.Remove("./test-data2.json")
		os.Remove("./test-input.txt")
	}()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
