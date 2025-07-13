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

// Command with validation examples from the tutorial
type ValidatedListUsersCommand struct {
	*cmds.CommandDescription
}

type ValidatedListUsersSettings struct {
	Limit  int    `glazed.parameter:"limit"`
	Filter string `glazed.parameter:"name-filter"`
	Active bool   `glazed.parameter:"active-only"`
}

func (c *ValidatedListUsersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	settings := &ValidatedListUsersSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Business rule validation from tutorial
	if settings.Limit < 1 {
		return fmt.Errorf("limit must be at least 1, got %d", settings.Limit)
	}
	if settings.Limit > 1000 {
		return fmt.Errorf("limit cannot exceed 1000 (got %d) - use filtering to narrow results", settings.Limit)
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
			return fmt.Errorf("failed to add user row: %w", err)
		}
	}

	return nil
}

func NewValidatedListUsersCommand() (*ValidatedListUsersCommand, error) {
	glazedLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"list-users",
		cmds.WithShort("List users with validation"),
		cmds.WithLong(`Test validation scenarios with business rules.`),
		cmds.WithFlags(
			parameters.NewParameterDefinition(
				"limit",
				parameters.ParameterTypeInteger,
				parameters.WithDefault(10),
				parameters.WithHelp("Maximum number of users to show (1-1000)"),
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

	return &ValidatedListUsersCommand{
		CommandDescription: cmdDesc,
	}, nil
}

var _ cmds.GlazeCommand = &ValidatedListUsersCommand{}

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
		Use:   "validation-test",
		Short: "Test validation scenarios",
	}

	listUsersCmd, err := NewValidatedListUsersCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraListUsersCmd, err := cli.BuildCobraCommandFromGlazeCommand(listUsersCmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraListUsersCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
