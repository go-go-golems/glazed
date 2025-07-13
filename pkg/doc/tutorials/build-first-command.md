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

# Build Your First Glazed Command

This tutorial demonstrates building a command-line tool that outputs data in multiple formats automatically. You'll implement a user management command that supports JSON, YAML, CSV, and table output formats without writing format-specific code.

**Learning objectives:**
- Create a functional CLI command with filtering and limiting options
- Implement automatic support for multiple output formats
- Learn fundamental patterns for structured data processing in Glazed
- Understand command configuration and parameter handling

## Prerequisites

- Go 1.19+ installed
- Basic familiarity with Go and command-line tools

## Step 1: Set Up Your Project

Create a workspace for the command:

```bash
mkdir glazed-quickstart
cd glazed-quickstart
go mod init glazed-quickstart
go get github.com/go-go-golems/glazed
go get github.com/spf13/cobra
```

**Project structure:**
- `glazed-quickstart` serves as the project directory
- `go mod init` creates a Go module for dependency tracking
- Two key dependencies:
  - `glazed` provides structured data processing capabilities
  - `cobra` handles command-line parsing (Glazed builds on this framework)

## Step 2: Create Your First Command

Create `main.go` with the complete command implementation:

```go
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
```

### Command Structure

Define the command structure using Glazed patterns:

```go
// Step 2.1: Define your command struct
type ListUsersCommand struct {
    *cmds.CommandDescription
}

// Step 2.2: Define settings for type-safe parameter access
type ListUsersSettings struct {
    Limit      int    `glazed.parameter:"limit"`
    NameFilter string `glazed.parameter:"name-filter"`
    Active     bool   `glazed.parameter:"active-only"`
}
```

**Key components:**

1. **Command Struct**: `ListUsersCommand` embeds `*cmds.CommandDescription`, which contains command metadata (name, help text, parameters)

2. **Settings Struct**: `ListUsersSettings` maps command-line flags to Go fields using struct tags. The `glazed.parameter` tags provide automatic type conversion and validation.

### Core Command Logic

Implement the data processing functionality:

```go

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
    users := generateMockUsers(settings.Limit, settings.NameFilter, settings.Active)

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
```

**Implementation details:**

1. **Settings Extraction**: `parsedLayers.InitializeStruct()` populates the settings struct from command-line flags with automatic parsing and validation
2. **Business Logic**: `generateMockUsers()` simulates data retrieval with the parsed settings
3. **Structured Output**: Creates `types.Row` objects instead of using direct output functions
4. **Row Structure**: `types.MRP("key", value)` creates key-value pairs for each data field

The `GlazeProcessor` collects these rows and can output them in multiple formats without additional format-specific code.

### Command Configuration and Parameters

Configure command metadata and parameter definitions:

```go
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
  list-users --name-filter admin       # Filter users containing "admin"
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
```

**Configuration components:**

1. **Glazed Layer**: `settings.NewGlazedParameterLayers()` adds built-in parameters like `--output`, `--fields`, `--sort-columns`
2. **Command Metadata**: Defines command name, short description, and comprehensive help text with usage examples
3. **Parameter Definitions**: Each flag specifies:
   - **Type**: Integer, String, Bool with automatic validation
   - **Default Value**: Behavior when the flag is not specified
   - **Help Text**: Displayed in `--help` output
   - **Short Flag**: Single-letter abbreviations for convenience
4. **Layer Composition**: Combines custom parameters with Glazed's built-in output parameters

### Interface Compliance and Mock Data

```go
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
            if !strings.Contains(strings.ToLower(user.Name), strings.ToLower(filter)) && 
               !strings.Contains(strings.ToLower(user.Email), strings.ToLower(filter)) && 
               !strings.Contains(strings.ToLower(user.Department), strings.ToLower(filter)) {
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


```

**Implementation notes:**

1. **Interface Compliance Check**: The `var _ cmds.GlazeCommand = &ListUsersCommand{}` line ensures the struct implements the required interface at compile time
2. **Mock Data**: Provides realistic sample data for development and testing. Replace `generateMockUsers()` with actual data sources in production
3. **Filtering Logic**: Demonstrates how command parameters control data processing

### CLI Application Integration

Wire the command into a CLI application:

```go
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

**Integration steps:**

1. **Root Command**: Creates a standard Cobra root command as the application entry point
2. **Command Creation**: `NewListUsersCommand()` creates the Glazed command with configuration
3. **Glazed-to-Cobra Bridge**: `cli.BuildCobraCommandFromGlazeCommand()` converts the Glazed command to a Cobra command, handling parameter setup and processing
4. **Registration**: Adds the converted command as a subcommand
5. **Execution**: Starts the CLI application and processes command-line arguments

## Step 3: Build and Test Your Command

Build and test the command functionality:

```bash
# Build the application
go build -o glazed-quickstart

# Test basic functionality
./glazed-quickstart list-users --help

# Try different parameter combinations
./glazed-quickstart list-users
./glazed-quickstart list-users --limit 3
./glazed-quickstart list-users --filter Engineering
./glazed-quickstart list-users --active-only
```

**Expected behavior:**

1. **Help Text**: `--help` displays auto-generated parameter descriptions and examples
2. **Parameter Validation**: Invalid values trigger automatic validation errors
3. **Default Behavior**: Without flags, shows the first 10 users in table format
4. **Filtering**: `--filter Engineering` displays only users matching the filter criteria

## Step 4: Multiple Output Formats

Test the automatic output format support:

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

**Key capabilities demonstrated:**

1. **Zero Additional Code**: All output formats work automatically through the `types.Row` and `GlazeProcessor` pattern
2. **Field Selection**: `--fields id,name,email` displays only specified columns
3. **Sorting**: `--sort-columns name` sorts alphabetically (use `--sort-columns -name` for reverse order)
4. **Composability**: All flags combine seamlessly for flexible data presentation

## Step 5: Dual Commands (Advanced)

Implement a command that supports both simple text output and structured data modes:

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

**Dual command pattern:**

The `StatusCommand` implements two interfaces:

1. **`BareCommand`** (via `Run` method): Produces human-readable text output
2. **`GlazeCommand`** (via `RunIntoGlazeProcessor` method): Produces structured data

**Interface differences:**
- **Human Mode**: Uses `fmt.Println()` for formatted text display
- **Structured Mode**: Uses `types.Row` for machine-parseable data
- **Shared Logic**: Both methods access the same parsed settings

### Integrating the Dual Command

Add the dual command to your application:

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

**Key differences from single-mode commands:**

1. **`BuildCobraCommandDualMode`**: Uses the dual-mode builder instead of the standard builder
2. **Toggle Flag**: `WithGlazeToggleFlag("with-glaze-output")` creates a flag that switches between interfaces
3. **Automatic Detection**: Glazed detects both interface implementations and configures the toggle mechanism

### Testing Dual Command

Test both output modes:

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

**Output comparison:**

- **Classic Mode**: Human-readable text with clear labels and formatting
- **Glaze Mode**: Structured data compatible with automation tools and scripts

## Implementation Summary

### Core Concepts Implemented

**Command Structure**
- Command structs embed `CommandDescription` for metadata
- Settings structs provide type-safe parameter access through struct tags
- Interface implementation defines command behavior

**Parameter Management**
- Declarative parameter definitions with types, defaults, and help text
- Automatic parsing, validation, and type conversion
- Layer composition for reusable parameter groups

**Output Processing**
- Structured data through `types.Row` objects enables multiple output formats
- `GlazeProcessor` handles format conversion automatically
- Dual command interfaces support both human and machine-readable output

### Key Patterns

**Input Validation**
```go
func (c *ListUsersCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    settings := &ListUsersSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
        return err
    }
    
    // Validate business rules
    if settings.Limit < 1 {
        return fmt.Errorf("limit must be at least 1, got %d", settings.Limit)
    }
    if settings.Limit > 1000 {
        return fmt.Errorf("limit cannot exceed 1000 (got %d) - use filtering to narrow results", settings.Limit)
    }
    
    // Continue with command logic...
}
```

**Extended Parameter Types**
```go
cmds.WithFlags(
    // File parameter - validates file exists
    parameters.NewParameterDefinition(
        "config-file",
        parameters.ParameterTypeFile,
        parameters.WithHelp("Configuration file path"),
    ),
    
    // Choice parameter - limits valid options
    parameters.NewParameterDefinition(
        "format",
        parameters.ParameterTypeChoice,
        parameters.WithChoices("json", "yaml", "xml"),
        parameters.WithDefault("json"),
        parameters.WithHelp("Output format"),
    ),
    
    // String parameter for timeout - in a real implementation you would parse this as a duration
    parameters.NewParameterDefinition(
        "timeout",
        parameters.ParameterTypeString,
        parameters.WithDefault("30s"),
        parameters.WithHelp("Request timeout (e.g., '30s', '5m')"),
    ),
)
```

**Error Handling**
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

## Next Steps

### For Data Processing Tools
- **[Middlewares Guide](../topics/middlewares-guide.md)**: Transform and filter data in processing pipelines
- **[Layers Guide](../topics/layers-guide.md)**: Create reusable parameter sets for common configurations

### For Application Suites
- **[Commands Reference](../topics/commands-reference.md)**: Organize commands and manage complex applications
- **[Custom Layer Tutorial](./custom-layer.md)**: Build domain-specific parameter layers

### For Advanced Patterns
- Study the implementation patterns in this tutorial for production-ready command structures
- Experiment with parameter types and output format combinations
- Implement validation and error handling appropriate for your use cases

This foundation provides the core patterns for building professional CLI applications with Glazed's structured data processing capabilities.
