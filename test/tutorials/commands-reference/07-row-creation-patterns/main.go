package main

import (
	"context"
	"fmt"
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

// RowCreationCommand demonstrates different row creation patterns from the documentation
type RowCreationCommand struct {
	*cmds.CommandDescription
}

// RowCreationSettings mirrors the command parameters
type RowCreationSettings struct {
	Method string `glazed.parameter:"method"`
}

// User struct for demonstrating struct-to-row conversion
type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
	Created time.Time `json:"created"`
}

func (c *RowCreationCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &RowCreationSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	switch s.Method {
	case "mrp":
		return c.demonstrateMRPMethod(ctx, gp)
	case "map":
		return c.demonstrateMapMethod(ctx, gp)
	case "struct":
		return c.demonstrateStructMethod(ctx, gp)
	case "all":
		if err := c.demonstrateMRPMethod(ctx, gp); err != nil {
			return err
		}
		if err := c.demonstrateMapMethod(ctx, gp); err != nil {
			return err
		}
		return c.demonstrateStructMethod(ctx, gp)
	default:
		return fmt.Errorf("unknown method: %s", s.Method)
	}
}

// demonstrateMRPMethod shows using NewRow with MRP (MapRowPair) as shown in documentation
func (c *RowCreationCommand) demonstrateMRPMethod(ctx context.Context, gp middlewares.Processor) error {
	fmt.Println("Creating rows using MRP (MapRowPair) method...")

	// Example from documentation
	row := types.NewRow(
		types.MRP("id", 1),
		types.MRP("name", "John Doe"),
		types.MRP("email", "john@example.com"),
		types.MRP("active", true),
		types.MRP("method", "MRP"),
		types.MRP("description", "Created using types.MRP()"),
		types.MRP("nested_data", map[string]interface{}{
			"department": "Engineering",
			"level":      "Senior",
			"skills":     []string{"Go", "Python", "SQL"},
		}),
	)

	return gp.AddRow(ctx, row)
}

// demonstrateMapMethod shows creating rows from maps as shown in documentation
func (c *RowCreationCommand) demonstrateMapMethod(ctx context.Context, gp middlewares.Processor) error {
	fmt.Println("Creating rows from map...")

	// Example from documentation
	data := map[string]interface{}{
		"id":          2,
		"name":        "Jane Smith",
		"email":       "jane@example.com",
		"active":      true,
		"method":      "Map",
		"description": "Created from map[string]interface{}",
		"permissions": []string{"read", "write", "admin"},
		"metadata":    map[string]string{
			"source":      "user_import",
			"import_date": time.Now().Format(time.RFC3339),
		},
	}
	row := types.NewRowFromMap(data)

	return gp.AddRow(ctx, row)
}

// demonstrateStructMethod shows creating rows from structs as shown in documentation
func (c *RowCreationCommand) demonstrateStructMethod(ctx context.Context, gp middlewares.Processor) error {
	fmt.Println("Creating rows from struct...")

	// Example from documentation
	user := User{
		ID:      3,
		Name:    "Bob Wilson",
		Email:   "bob@example.com",
		Active:  false,
		Created: time.Now().Add(-time.Hour * 24 * 30), // 30 days ago
	}

	// The second parameter (true) means lowercase field names
	row := types.NewRowFromStruct(&user, true)
	
	// Add extra fields to show this is from struct method
	row.Set("method", "Struct")
	row.Set("description", "Created from struct with lowercase field names")

	return gp.AddRow(ctx, row)
}

// NewRowCreationCommand creates a new row creation demo command
func NewRowCreationCommand() (*RowCreationCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"row-creation",
		cmds.WithShort("Demonstrate different row creation patterns"),
		cmds.WithLong("Demonstrate all three row creation methods: MRP, Map, and Struct"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"method",
				parameters.ParameterTypeChoice,
				parameters.WithChoices("mrp", "map", "struct", "all"),
				parameters.WithDefault("all"),
				parameters.WithHelp("Row creation method to demonstrate"),
			),
		),
		cmds.WithLayersList(glazedLayer),
	)

	return &RowCreationCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &RowCreationCommand{}

func main() {
	cmd, err := NewRowCreationCommand()
	if err != nil {
		log.Fatalf("Error creating command: %v", err)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
	if err != nil {
		log.Fatalf("Error building Cobra command: %v", err)
	}

	rootCmd := &cobra.Command{
		Use:   "row-creation-demo",
		Short: "Demonstration of row creation patterns from Glazed documentation",
		Long: `This demonstrates all three row creation patterns from the documentation:

1. MRP (MapRowPair): Using types.MRP() for explicit key-value pairs
2. Map: Using types.NewRowFromMap() for map[string]interface{}
3. Struct: Using types.NewRowFromStruct() for Go structs

Each method has different use cases and benefits.`,
	}
	rootCmd.AddCommand(cobraCmd)

	// Add comprehensive examples
	cobraCmd.Example = `  # Show all methods
  row-creation-demo row-creation

  # Show specific method
  row-creation-demo row-creation --method mrp
  row-creation-demo row-creation --method map
  row-creation-demo row-creation --method struct

  # Different output formats
  row-creation-demo row-creation --output json
  row-creation-demo row-creation --output yaml
  row-creation-demo row-creation --method struct --output table`

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
