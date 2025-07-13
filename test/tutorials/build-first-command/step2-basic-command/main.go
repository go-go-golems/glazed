package main

import (
	"context"
	"fmt"
	"os"
	"strings"
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

// Step 2.1: Define your command struct
type ListUsersCommand struct {
	*cmds.CommandDescription
}

// Step 2.2: Define settings for type-safe parameter access
type ListUsersSettings struct {
	Limit  int    `glazed.parameter:"limit"`
	Filter string `glazed.parameter:"name-filter"`
	Active bool   `glazed.parameter:"active-only"`
}

// Step 2.3: Implement the GlazeCommand interface
func (c *ListUsersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	// Parse settings from command line
	settings := &ListUsersSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	// Simulate getting users (in real app, this would be a database call)
	users := generateMockUsers(settings.Limit, settings.Filter, settings.Active)

	// Output structured data as rows
	for _, user := range users {
		row := types.NewRow(
			types.MRP("id", user.ID),
			types.MRP("name", user.Name),
			types.MRP("email", user.Email),
			types.MRP("department", user.Department),
			types.MRP("active", user.Active),
			types.MRP("created_at", user.CreatedAt.Format("2006-01-02")),
		)

		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

// Step 2.4: Create constructor function
func NewListUsersCommand() (*ListUsersCommand, error) {
	// Create glazed layer for output formatting options
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	// Define command with parameters
	cmdDesc := cmds.NewCommandDescription(
		"list-users",
		cmds.WithShort("List users in the system"),
		cmds.WithLong(`
List all users with optional filtering and limiting.
Supports multiple output formats including JSON, YAML, CSV, and tables.

Examples:
  list-users                           # List all users as table
  list-users --limit 5                 # Show only first 5 users
  list-users --filter admin            # Filter users containing "admin"
  list-users --active-only             # Show only active users
  list-users --output json             # Output as JSON
  list-users --output csv              # Output as CSV
        `),

		// Define command flags
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(10),
				parameters.WithHelp("Maximum number of users to show"),
				parameters.WithShortFlag("l"),
			),
			parameters.NewParameterDefinition(
				"name-filter",
				parameters.ParameterTypeString,
				parameters.WithDefault(""),
				parameters.WithHelp("Filter users by name or email"),
				parameters.WithShortFlag("f"),
			),
			parameters.NewParameterDefinition(
				"active-only",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Show only active users"),
				parameters.WithShortFlag("a"),
			),
		),

		// Add glazed layer for output formatting
		cmds.WithLayersList(glazedLayer),
	)

	return &ListUsersCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &ListUsersCommand{}

// Mock data structures and generation
type User struct {
	ID         int
	Name       string
	Email      string
	Department string
	Active     bool
	CreatedAt  time.Time
}

func generateMockUsers(limit int, filter string, activeOnly bool) []User {
	allUsers := []User{
		{1, "Alice Johnson", "alice@company.com", "Engineering", true, time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
		{2, "Bob Smith", "bob@company.com", "Marketing", true, time.Date(2023, 2, 20, 0, 0, 0, 0, time.UTC)},
		{3, "Charlie Brown", "charlie@company.com", "Engineering", false, time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC)},
		{4, "Diana Prince", "diana@company.com", "HR", true, time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC)},
		{5, "Eve Adams", "eve@company.com", "Sales", true, time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC)},
		{6, "Frank Miller", "frank@company.com", "Engineering", false, time.Date(2023, 6, 8, 0, 0, 0, 0, time.UTC)},
		{7, "Grace Hopper", "grace@company.com", "Engineering", true, time.Date(2023, 7, 22, 0, 0, 0, 0, time.UTC)},
		{8, "Henry Ford", "henry@company.com", "Operations", true, time.Date(2023, 8, 14, 0, 0, 0, 0, time.UTC)},
	}

	var filtered []User
	for _, user := range allUsers {
		// Apply active filter
		if activeOnly && !user.Active {
			continue
		}

		// Apply text filter
		if filter != "" {
			if !strings.Contains(user.Name, filter) && !strings.Contains(user.Email, filter) && !strings.Contains(user.Department, filter) {
				continue
			}
		}

		filtered = append(filtered, user)

		// Apply limit
		if len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// Custom string functions removed - now using strings.Contains() from standard library

// Step 3: Set up CLI application
func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "glazed-quickstart",
		Short: "A quick start example of Glazed commands",
		Long:  "Demonstrates how to build commands with Glazed framework",
	}

	// Create and register our command
	listUsersCmd, err := NewListUsersCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Convert to Cobra command
	cobraListUsersCmd, err := cli.BuildCobraCommandFromGlazeCommand(listUsersCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
		os.Exit(1)
	}

	// Add to root command
	rootCmd.AddCommand(cobraListUsersCmd)

	// Execute
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
