package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ListUsersSettings struct {
	Limit      int    `glazed:"limit"`
	NameFilter string `glazed:"name-filter"`
	ActiveOnly bool   `glazed:"active-only"`
}

type ListUsersCommand struct {
	*cmds.CommandDescription
}

func NewListUsersCommand() (*ListUsersCommand, error) {
	desc := cmds.NewCommandDescription(
		"list-users",
		cmds.WithShort("List users (structured output)"),
		cmds.WithLong(`A tiny example showing the wrapper API (schema/fields/values) and Glazed output formats.

Examples:
  list-users --limit 3
  list-users --limit 3 --output json
  list-users --name-filter engineering --fields id,name,department
`),
		cmds.WithFlags(
			fields.New(
				"limit",
				fields.TypeInteger,
				fields.WithDefault(10),
				fields.WithHelp("Maximum number of users to emit"),
				fields.WithShortFlag("l"),
			),
			fields.New(
				"name-filter",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Filter by name/email/department substring (case-insensitive)"),
				fields.WithShortFlag("f"),
			),
			fields.New(
				"active-only",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Only emit active users"),
				fields.WithShortFlag("a"),
			),
		),
	)

	return &ListUsersCommand{CommandDescription: desc}, nil
}

var _ cmds.GlazeCommand = &ListUsersCommand{}

func (c *ListUsersCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	settings := &ListUsersSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return errors.Wrap(err, "failed to decode settings")
	}

	users := filterUsers(settings.Limit, settings.NameFilter, settings.ActiveOnly)
	for _, u := range users {
		row := types.NewRow(
			types.MRP("id", u.ID),
			types.MRP("name", u.Name),
			types.MRP("email", u.Email),
			types.MRP("department", u.Department),
			types.MRP("active", u.Active),
			types.MRP("created_at", u.CreatedAt.Format("2006-01-02")),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}

type User struct {
	ID         int
	Name       string
	Email      string
	Department string
	Active     bool
	CreatedAt  time.Time
}

func filterUsers(limit int, filter string, activeOnly bool) []User {
	all := []User{
		{1, "Alice Johnson", "alice@company.com", "Engineering", true, time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
		{2, "Bob Smith", "bob@company.com", "Marketing", true, time.Date(2023, 2, 20, 0, 0, 0, 0, time.UTC)},
		{3, "Charlie Brown", "charlie@company.com", "Engineering", false, time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC)},
		{4, "Diana Prince", "diana@company.com", "HR", true, time.Date(2023, 4, 5, 0, 0, 0, 0, time.UTC)},
		{5, "Eve Adams", "eve@company.com", "Sales", true, time.Date(2023, 5, 12, 0, 0, 0, 0, time.UTC)},
		{6, "Frank Miller", "frank@company.com", "Engineering", false, time.Date(2023, 6, 8, 0, 0, 0, 0, time.UTC)},
	}

	filterLower := strings.ToLower(filter)
	out := make([]User, 0, len(all))
	for _, u := range all {
		if activeOnly && !u.Active {
			continue
		}
		if filterLower != "" {
			if !strings.Contains(strings.ToLower(u.Name), filterLower) &&
				!strings.Contains(strings.ToLower(u.Email), filterLower) &&
				!strings.Contains(strings.ToLower(u.Department), filterLower) {
				continue
			}
		}
		out = append(out, u)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "new-api-build-first-command",
		Short: "New API example: build first command",
	}

	listUsers, err := NewListUsersCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating command: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommand(listUsers,
		cli.WithParserConfig(cli.CobraParserConfig{
			ShortHelpSections: []string{schema.DefaultSlug},
			MiddlewaresFunc:   cli.CobraCommandDefaultMiddlewares,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
