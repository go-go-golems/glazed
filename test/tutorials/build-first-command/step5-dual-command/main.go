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

// ListUsersCommand from Step 2
type ListUsersCommand struct {
	*cmds.CommandDescription
}

type ListUsersSettings struct {
	Limit  int    `glazed.parameter:"limit"`
	Filter string `glazed.parameter:"name-filter"`
	Active bool   `glazed.parameter:"active-only"`
}

func (c *ListUsersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ListUsersSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	users := generateMockUsers(settings.Limit, settings.Filter, settings.Active)

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

func NewListUsersCommand() (*ListUsersCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"list-users",
		cmds.WithShort("List users in the system"),
		cmds.WithLong(`List all users with optional filtering and limiting.`),
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
		cmds.WithLayersList(glazedLayer),
	)

	return &ListUsersCommand{
		CommandDescription: cmdDesc,
	}, nil
}

var _ cmds.GlazeCommand = &ListUsersCommand{}

// Dual command that implements both BareCommand and GlazeCommand
type StatusCommand struct {
	*cmds.CommandDescription
}

// Settings for status command
type StatusSettings struct {
	Verbose bool `glazed.parameter:"verbose"`
}

// Classic mode - simple text output
func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	fmt.Println("System Status:")
	fmt.Println("  Users: 8 total, 6 active")
	fmt.Println("  Departments: 5")
	fmt.Println("  Status: Healthy")

	if settings.Verbose {
		fmt.Println("  Last updated:", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("  Version: 1.0.0")
	}

	return nil
}

// Glaze mode - structured output
func (c *StatusCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &StatusSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return err
	}

	row := types.NewRow(
		types.MRP("total_users", 8),
		types.MRP("active_users", 6),
		types.MRP("departments", 5),
		types.MRP("status", "healthy"),
		types.MRP("timestamp", time.Now().Format(time.RFC3339)),
	)

	if settings.Verbose {
		row.Set("version", "1.0.0")
		row.Set("uptime", "24h30m")
	}

	return gp.AddRow(ctx, row)
}

// Constructor for status command
func NewStatusCommand() (*StatusCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"status",
		cmds.WithShort("Show system status"),
		cmds.WithLong("Show system status in either human-readable or structured format"),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"verbose",
				parameters.ParameterTypeBool,
				parameters.WithDefault(false),
				parameters.WithHelp("Show additional details"),
				parameters.WithShortFlag("v"),
			),
		),
	)

	return &StatusCommand{
		CommandDescription: cmdDesc,
	}, nil
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}

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
		if activeOnly && !user.Active {
			continue
		}

		if filter != "" {
			if !strings.Contains(user.Name, filter) && !strings.Contains(user.Email, filter) && !strings.Contains(user.Department, filter) {
				continue
			}
		}

		filtered = append(filtered, user)

		if len(filtered) >= limit {
			break
		}
	}

	return filtered
}

// Custom string functions removed - now using strings.Contains() from standard library

func main() {
	rootCmd := &cobra.Command{
		Use:   "glazed-quickstart",
		Short: "A quick start example of Glazed commands",
		Long:  "Demonstrates how to build commands with Glazed framework",
	}

	// Create list-users command
	listUsersCmd, err := NewListUsersCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating list-users command: %v\n", err)
		os.Exit(1)
	}

	cobraListUsersCmd, err := cli.BuildCobraCommandFromGlazeCommand(listUsersCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building list-users command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraListUsersCmd)

	// Create status command with dual mode
	statusCmd, err := NewStatusCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating status command: %v\n", err)
		os.Exit(1)
	}

	// Use dual mode builder
	cobraStatusCmd, err := cli.BuildCobraCommandDualMode(
		statusCmd,
		cli.WithGlazeToggleFlag("with-glaze-output"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building status command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraStatusCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
