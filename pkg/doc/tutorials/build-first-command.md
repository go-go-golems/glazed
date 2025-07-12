---
Title: Build Your First Glazed Command
Slug: build-first-command
Short: Quick hands-on tutorial to build, run, and use a Glazed command with structured output
Topics:
- tutorial
- commands
- quick-start
- glazed
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Build Your First Glazed Command: Quick Start

This tutorial will get you building and running your first Glazed command in under 10 minutes. We'll create a simple user management command that demonstrates the core concepts.

## Prerequisites

- Go 1.19+ installed
- Basic familiarity with Go and command-line tools

## Step 1: Set Up Your Project

```bash
mkdir glazed-quickstart
cd glazed-quickstart
go mod init glazed-quickstart
go get github.com/go-go-golems/glazed
go get github.com/spf13/cobra
```

## Step 2: Create Your First Command

Create `main.go`:

```go
package main

import (
    "context"
    "fmt"
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

// Step 2.1: Define your command struct
type ListUsersCommand struct {
    *cmds.CommandDescription
}

// Step 2.2: Define settings for type-safe parameter access
type ListUsersSettings struct {
    Limit  int    `glazed.parameter:"limit"`
    Filter string `glazed.parameter:"filter"`
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
                "filter",
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
            if !contains(user.Name, filter) && !contains(user.Email, filter) && !contains(user.Department, filter) {
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

func contains(s, substr string) bool {
    return len(s) >= len(substr) && 
           (s == substr || 
            len(s) > len(substr) && 
            (s[:len(substr)] == substr || 
             s[len(s)-len(substr):] == substr || 
             indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
    for i := 0; i <= len(s)-len(substr); i++ {
        if s[i:i+len(substr)] == substr {
            return i
        }
    }
    return -1
}

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
```

## Step 3: Build and Test

```bash
# Build the application
go build -o glazed-quickstart

# Test basic functionality
./glazed-quickstart list-users --help

# Try different output modes
./glazed-quickstart list-users
./glazed-quickstart list-users --limit 3
./glazed-quickstart list-users --filter Engineering
./glazed-quickstart list-users --active-only
```

## Step 4: Explore Output Formats

The beauty of Glazed is automatic support for multiple output formats:

```bash
# Table output (default)
./glazed-quickstart list-users --limit 3

# JSON output
./glazed-quickstart list-users --limit 3 --output json

# YAML output
./glazed-quickstart list-users --limit 3 --output yaml

# CSV output
./glazed-quickstart list-users --limit 3 --output csv

# Select specific fields
./glazed-quickstart list-users --fields id,name,email

# Sort by field
./glazed-quickstart list-users --sort-columns name

# Combine options
./glazed-quickstart list-users --filter Engineering --output json --fields name,department
```

## Step 5: Add a Dual Command (Optional)

Now let's enhance our example with a dual command that can run in both simple text mode and structured data mode:

Add this to your `main.go`:

```go
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
```

Then in your `main()` function, add:

```go
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
```

Test the dual command:

```bash
# Rebuild
go build -o glazed-quickstart

# Classic mode (default)
./glazed-quickstart status
./glazed-quickstart status --verbose

# Glaze mode
./glazed-quickstart status --with-glaze-output
./glazed-quickstart status --with-glaze-output --output json
./glazed-quickstart status --with-glaze-output --verbose --output yaml
```

## What You've Learned

Congratulations! You've just built your first Glazed application. Here's what you've accomplished:

### Core Concepts
- **Command Structure**: How to structure Glazed commands with embedded `CommandDescription`
- **Settings Structs**: Using struct tags for type-safe parameter access
- **Interface Implementation**: Implementing `GlazeCommand` for structured output
- **Parameter Definition**: Creating flags with types, defaults, and help text

### Output Capabilities
- **Multiple Formats**: Automatic support for JSON, YAML, CSV, and table output
- **Field Selection**: Using `--fields` to select specific columns
- **Sorting**: Using `--sort-columns` to order results
- **Filtering**: Implementing custom filtering logic

### Advanced Features
- **Dual Commands**: Commands that support both simple and structured output modes
- **Mode Switching**: Using `--with-glaze-output` to toggle between modes
- **Layer System**: Understanding how Glazed layers organize parameters

## Next Steps

Now that you've built your first command, explore these advanced topics:

1. **[Layers Guide](../topics/layers-guide.md)**: Learn about parameter organization and reusable layers
2. **[Commands Reference](../topics/commands-reference.md)**: Comprehensive command system documentation
3. **[Custom Layer Tutorial](./custom-layer.md)**: Build your own parameter layers
4. **[Middlewares Guide](../topics/middlewares-guide.md)**: Load parameters from multiple sources

## Common Patterns

### Adding Validation
```go
func (c *ListUsersCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    settings := &ListUsersSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    // Validate settings
    if settings.Limit < 1 {
        return fmt.Errorf("limit must be at least 1")
    }
    if settings.Limit > 1000 {
        return fmt.Errorf("limit cannot exceed 1000")
    }
    
    // Continue with command logic...
}
```

### Adding More Parameter Types
```go
cmds.WithFlags(
    // File parameter
    parameters.NewParameterDefinition(
        "config-file",
        parameters.ParameterTypeFile,
        parameters.WithHelp("Configuration file path"),
    ),
    
    // Choice parameter
    parameters.NewParameterDefinition(
        "format",
        parameters.ParameterTypeChoice,
        parameters.WithChoices("json", "yaml", "xml"),
        parameters.WithDefault("json"),
        parameters.WithHelp("Output format"),
    ),
    
    // Duration parameter
    parameters.NewParameterDefinition(
        "timeout",
        parameters.ParameterTypeDuration,
        parameters.WithDefault("30s"),
        parameters.WithHelp("Request timeout"),
    ),
)
```

### Error Handling
```go
func (c *ListUsersCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    settings := &ListUsersSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return fmt.Errorf("failed to parse settings: %w", err)
    }
    
    users, err := fetchUsersFromDatabase(settings)
    if err != nil {
        return fmt.Errorf("failed to fetch users: %w", err)
    }
    
    for _, user := range users {
        row := types.NewRowFromStruct(&user, true)
        if err := gp.AddRow(ctx, row); err != nil {
            return fmt.Errorf("failed to add user row: %w", err)
        }
    }
    
    return nil
}
```

You're now ready to build powerful command-line applications with Glazed!
