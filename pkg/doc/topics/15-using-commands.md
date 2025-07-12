---
Title: Using Commands in Glazed
Slug: using-commands
Short: Comprehensive guide on creating, configuring, running, and loading commands in Glazed.
Topics:
- commands
- flags
- arguments
- layers
- loaders
- runner
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Using Commands in Glazed

## Overview

The `glazed` framework provides a powerful system for creating command-line interfaces with rich features like structured data output, parameter validation, and multiple output formats. This guide will walk you through the process of creating, configuring, and running commands with Glazed, starting from basic concepts and progressing to more advanced usage.

## Architecture of Glazed Commands

Before diving into implementation details, let's understand the overall architecture of Glazed's command system:

```
┌─────────────────────────────────────────────┐
│                Command Interface             │
├────────────┬─────────────────┬──────────────┤
│BareCommand │ WriterCommand   │ GlazeCommand │
└────────────┴─────────────────┴──────────────┘
                      │
            ┌─────────┼─────────┐
            ▼         ▼         ▼
┌─────────────────────────────────────────────┐
│          CommandDescription                  │
│  (name, flags, arguments, layers, etc.)     │
└─────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────┐
│               Parameter Layers               │
│     ┌───────────┬────────────┬────────┐     │
│     │Default    │Glazed      │Custom  │     │
│     │(flags/args)│(output fmt)│(any)   │     │
│     └───────────┴────────────┴────────┘     │
└─────────────────────────────────────────────┘
               │        │
     ┌─────────┘        └────────┐
     ▼                           ▼
┌──────────────┐          ┌─────────────────┐
│  Parameters  │          │ ParsedLayers    │
│ (definitions)│─────────▶│(runtime values) │
└──────────────┘          └─────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────┐
│     Execution (via Run methods or runner)    │
└─────────────────────────────────────────────┘
                │
    ┌───────────┼────────────┐
    ▼           ▼            ▼
┌────────┐ ┌──────────┐ ┌──────────────┐
│Direct  │ │Writer    │ │GlazeProcessor│
│Output  │ │Output    │ │(structured)  │
└────────┘ └──────────┘ └──────────────┘
                 │            │
                 └─────┬──────┘
                       ▼
            ┌─────────────────────┐
            │   Dual Commands     │
            │ (can run in both    │
            │ classic & glaze)    │
            └─────────────────────┘
```

Key components:

1. **Command Interfaces**: The core abstractions that define how commands behave and produce output
2. **CommandDescription**: Contains metadata about a command (name, description, parameters, etc.)
3. **Parameter Layers**: Organize parameters into logical groups
4. **Parameters**: Define the inputs a command accepts (flags, arguments)
5. **ParsedLayers**: Runtime values of parameters after parsing from various sources
6. **Execution Methods**: Different ways to run commands and handle their output

Now that we understand the big picture, let's explore each component in detail.

## Core Packages

Glazed's command functionality is distributed across several packages:

- `github.com/go-go-golems/glazed/pkg/cmds`: Core command interfaces and descriptions
- `github.com/go-go-golems/glazed/pkg/cmds/parameters`: Parameter types and definitions
- `github.com/go-go-golems/glazed/pkg/cmds/layers`: Parameter layering system
- `github.com/go-go-golems/glazed/pkg/cmds/runner`: Programmatic command execution
- `github.com/go-go-golems/glazed/pkg/middlewares`: Output processing
- `github.com/go-go-golems/glazed/pkg/types`: Structured data types
- `github.com/go-go-golems/glazed/pkg/cli`: Helpers for integrating with Cobra
- `github.com/go-go-golems/glazed/pkg/settings`: Standard Glazed parameter layers

## Command Interfaces

At the heart of Glazed are three main command interfaces, each designed for different output scenarios:

1. **BareCommand**: For commands that handle their own output directly
2. **WriterCommand**: For commands that write to a provided writer (like a file or console)
3. **GlazeCommand**: For commands that produce structured data using Glazed's processing pipeline

Additionally, Glazed supports **Dual Commands** that can implement multiple interfaces and switch between output modes at runtime.

Let's understand the differences:

### BareCommand

This is the simplest interface, suitable when you just need to perform an action without structured output:

```go
// BareCommand interface definition (simplified)
type BareCommand interface {
    Command
    Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error
}
```

### WriterCommand

When you need to write text output to a specific destination:

```go
// WriterCommand interface definition (simplified)
type WriterCommand interface {
    Command
    RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error
}
```

### GlazeCommand

For producing structured data that can be formatted in multiple ways (JSON, YAML, CSV, etc.):

```go
// GlazeCommand interface definition (simplified)
type GlazeCommand interface {
    Command
    RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error
}
```

### Dual Commands

Dual commands implement multiple interfaces, allowing them to run in different output modes based on runtime flags. This is particularly useful when you want to offer both simple text output for human consumption and structured data for programmatic use.

A dual command typically implements both `BareCommand` (or `WriterCommand`) and `GlazeCommand`:

```go
// Dual command example
type MyDualCommand struct {
    *cmds.CommandDescription
}

// Implement BareCommand for classic mode
func (c *MyDualCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Handle simple text output
    fmt.Println("Classic mode output")
    return nil
}

// Implement GlazeCommand for structured output mode
func (c *MyDualCommand) RunIntoGlazeProcessor(
    ctx context.Context, 
    parsedLayers *layers.ParsedLayers, 
    gp middlewares.Processor,
) error {
    // Handle structured data output
    row := types.NewRow(types.MRP("message", "Structured output"))
    return gp.AddRow(ctx, row)
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &MyDualCommand{}
var _ cmds.GlazeCommand = &MyDualCommand{}
```

#### Using the Dual Command Builder

Glazed provides a special builder for dual commands that automatically handles mode switching:

```go
// Create the dual command
dualCmd, err := NewMyDualCommand()
if err != nil {
    log.Fatalf("Error creating command: %v", err)
}

// Build using the dual mode builder
cobraCmd, err := cli.BuildCobraCommandDualMode(
    dualCmd,
    cli.WithGlazeToggleFlag("with-glaze-output"), // Optional: customize flag name
)
if err != nil {
    log.Fatalf("Error building command: %v", err)
}

// Add to your root command
rootCmd.AddCommand(cobraCmd)
```

#### Running Dual Commands

Users can switch between modes using the toggle flag:

```bash
# Classic mode (default)
myapp command --arg value

# Glaze mode with table output
myapp command --arg value --with-glaze-output

# Glaze mode with JSON output
myapp command --arg value --with-glaze-output --output json
```

#### Dual Command Options

The dual command builder supports several customization options:

```go
cobraCmd, err := cli.BuildCobraCommandDualMode(
    dualCmd,
    // Customize the toggle flag name
    cli.WithGlazeToggleFlag("structured-output"),
    
    // Hide specific glaze flags even when in glaze mode
    cli.WithHiddenGlazeFlags("template", "select"),
    
    // Make glaze mode the default (adds --no-glaze-output flag instead)
    cli.WithDefaultToGlaze(),
)
```

## Creating a Command Step-by-Step

Now that we understand the interfaces, let's walk through creating a command. We'll start with the basic structure and then build it up incrementally.

### Step 1: Define the Command Structure

First, create a struct that embeds `*cmds.CommandDescription` and implements one of the command interfaces:

```go
// Example for a GlazeCommand
type MyCommand struct {
    *cmds.CommandDescription
}

// Ensure the interface is implemented
var _ cmds.GlazeCommand = &MyCommand{}
```

### Step 2: Create a Settings Struct (Optional but Recommended)

For cleaner code, create a struct to hold your command's parameter values:

```go
// Settings for MyCommand
type MyCommandSettings struct {
    Count int  `glazed.parameter:"count"` 
    Test  bool `glazed.parameter:"test"`
}
```

The `glazed.parameter` struct tag maps to parameters by name, making it easy to access values.

### Step 3: Implement the Run Method

Depending on which interface you're implementing, add the appropriate Run method:

For a GlazeCommand:

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // 1. Parse settings from layers
    s := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // 2. Command logic goes here
    
    // 3. For a GlazeCommand, output rows to the processor
    row := types.NewRow(
        types.MRP("id", 1),
        types.MRP("name", "example"),
    )
    return gp.AddRow(ctx, row)
}
```

### Step 4: Create a Constructor Function

Finally, create a function to instantiate your command with all its parameters:

```go
func NewMyCommand() (*MyCommand, error) {
    // 1. Create the standard Glazed output layer (for GlazeCommand)
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    // 2. Create the command description with its parameters
    cmdDesc := cmds.NewCommandDescription(
        "mycommand", // Command name
        cmds.WithShort("My example command"),
        
        // 3. Define flags (optional parameters)
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "count",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("Number of items"),
                parameters.WithDefault(10),
            ),
            parameters.NewParameterDefinition(
                "test",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Enable test mode"),
                parameters.WithDefault(false),
            ),
        ),
        
        // 4. Define arguments (required positional parameters)
        cmds.WithArguments(
            parameters.NewParameterDefinition(
                "input",
                parameters.ParameterTypeString,
                parameters.WithHelp("Input file path"),
                parameters.WithRequired(true),
            ),
        ),
        
        // 5. Add parameter layers
        cmds.WithLayersList(
            glazedLayer, // Add standard Glazed output layer
        ),
    )
    
    // 6. Return the command instance
    return &MyCommand{
        CommandDescription: cmdDesc,
    }, nil
}
```

Let's break down the key components in more detail:

## Understanding Parameters

Parameters are the inputs to your command and are defined using `parameters.ParameterDefinition`. There are two types:

1. **Flags**: Optional named parameters (like `--count=10`)
2. **Arguments**: Required positional parameters (like `input.txt`)

### Parameter Types

Glazed supports various parameter types, each affecting how values are parsed and validated:

- Basic types: `ParameterTypeString`, `ParameterTypeSecret`, `ParameterTypeInteger`, `ParameterTypeBool`, `ParameterTypeFloat`, `ParameterTypeDate`
- Lists: `ParameterTypeStringList`, `ParameterTypeIntegerList`, `ParameterTypeFloatList`
- Choice-based: `ParameterTypeChoice`, `ParameterTypeChoiceList` (with predefined options)
- File-based: `ParameterTypeFile`, `ParameterTypeFileList`, `ParameterTypeStringFromFile`, etc.
- Key-value: `ParameterTypeKeyValue` (for map-like inputs)

**Note**: `ParameterTypeSecret` behaves like a string but masks sensitive values when displayed (e.g., `my***rd` for `"my-secret-password"`), protecting passwords, API keys, and other confidential data.

### Creating Parameter Definitions

Parameter definitions use a builder pattern to configure options:

```go
// Example of creating a parameter definition
paramDef := parameters.NewParameterDefinition(
    "name",                          // Parameter name
    parameters.ParameterTypeString,  // Parameter type
    parameters.WithHelp("User name"), // Help text
    parameters.WithDefault("guest"),  // Default value
    parameters.WithRequired(false),   // Is it required?
    parameters.WithShortFlag("n"),    // Short flag (-n)
)

// Example of a secret parameter for sensitive data
apiKeyParam := parameters.NewParameterDefinition(
    "api-key",
    parameters.ParameterTypeSecret,  // Masks value when displayed
    parameters.WithHelp("API key for authentication"),
    parameters.WithRequired(true),
    parameters.WithShortFlag("k"),
)
```

## Parameter Layers

Layers group related parameters together. This helps organize parameters and control parsing precedence.

### Default Layer

The "default" layer (`layers.DefaultSlug`) is automatically created when you use `cmds.WithFlags` or `cmds.WithArguments`. It contains your command-specific parameters.

### Glazed Layer

For GlazeCommands, the Glazed layer (`settings.GlazedSlug`) provides standard output formatting options like `--output=json` or `--fields=id,name`.

Create it with:

```go
glazedLayer, err := settings.NewGlazedParameterLayers()
if err != nil {
    return nil, err
}
```

Then add it to your command with `cmds.WithLayersList(glazedLayer)`.

### Custom Layers

You can create custom layers for configuration groups:

```go
// Example: Create a database configuration layer
dbLayer, err := layers.NewParameterLayer(
    "database",           // Layer slug/name
    "Database Settings",  // Display name
)
if err != nil {
    return nil, err
}

// Add parameters to the layer
dbLayer.AddFlags(
    parameters.NewParameterDefinition(
        "host",
        parameters.ParameterTypeString,
        parameters.WithDefault("localhost"),
    ),
    parameters.NewParameterDefinition(
        "port",
        parameters.ParameterTypeInteger,
        parameters.WithDefault(5432),
    ),
)

// Add the layer to your command
cmds.WithLayersList(dbLayer)
```

## Working with Command Outputs

Each command type handles output differently:

### BareCommand Output

BareCommands handle their own output directly:

```go
func (c *MyBareCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Access parameters as needed
    verbose, _ := parsedLayers.GetParameterValue(layers.DefaultSlug, "verbose")
    
    // Handle output directly
    fmt.Println("Executing command...")
    
    // ...command logic...
    
    return nil
}
```

### WriterCommand Output

WriterCommands write to a provided `io.Writer`:

```go
func (c *MyWriterCommand) RunIntoWriter(
    ctx context.Context, 
    parsedLayers *layers.ParsedLayers, 
    w io.Writer,
) error {
    // Write output to the provided writer
    _, err := fmt.Fprintf(w, "Command result: %s\n", "success")
    return err
}
```

### GlazeCommand Output

GlazeCommands emit structured data via rows:

```go
func (c *MyGlazeCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Create and add structured data rows
    for i := 0; i < 5; i++ {
        row := types.NewRow(
            types.MRP("id", i),
            types.MRP("name", fmt.Sprintf("Item %d", i)),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Understanding Row Creation

For GlazeCommands, rows form the structured output. Some key ways to create rows:

1. Using `types.NewRow` with `MRP` (MapRowPair):
   ```go
   row := types.NewRow(
       types.MRP("field1", value1),
       types.MRP("field2", value2),
   )
   ```

2. From a map:
   ```go
   data := map[string]interface{}{
       "field1": value1,
       "field2": value2,
   }
   row := types.NewRowFromMap(data)
   ```

3. From a struct:
   ```go
   type Person struct {
       ID   int
       Name string
   }
   person := Person{ID: 1, Name: "John"}
   row := types.NewRowFromStruct(&person, true) // true = lowercase field names
   ```

## Running Commands

There are two main ways to run Glazed commands:

1. **Within a CLI application using Cobra**
2. **Programmatically using the runner package**

Let's explore both approaches:

### Integration with Cobra

To use your Glazed command in a CLI application built with Cobra:

#### Step 1: Create your command instance

```go
myCmd, err := NewMyCommand()
if err != nil {
    log.Fatalf("Error creating command: %v", err)
}
```

#### Step 2: Convert to a Cobra command

There are several ways to convert a Glazed command to a Cobra command:

##### Option A: Automatic Builder (Recommended)
```go
// Automatically selects the appropriate builder based on command type
cobraCmd, err := cli.BuildCobraCommandFromCommand(myCmd)
if err != nil {
    log.Fatalf("Error building Cobra command: %v", err)
}
```

##### Option B: Specific Builder
```go
// Use specific builders for more control
if glazeCmd, ok := myCmd.(cmds.GlazeCommand); ok {
    cobraCmd, err = cli.BuildCobraCommandFromGlazeCommand(glazeCmd)
} else if writerCmd, ok := myCmd.(cmds.WriterCommand); ok {
    cobraCmd, err = cli.BuildCobraCommandFromWriterCommand(writerCmd)
} else if bareCmd, ok := myCmd.(cmds.BareCommand); ok {
    cobraCmd, err = cli.BuildCobraCommandFromBareCommand(bareCmd)
}
```

##### Option C: Dual Command Builder (For Dual Commands)
```go
// For commands that implement multiple interfaces
cobraCmd, err := cli.BuildCobraCommandDualMode(
    myCmd,
    cli.WithGlazeToggleFlag("with-glaze-output"),
)
if err != nil {
    log.Fatalf("Error building Cobra command: %v", err)
}
```

#### Step 3: Add to your root command

```go
// Add to your root command
rootCmd.AddCommand(cobraCmd)
```

#### Step 4: Execute the root command

```go
// Execute the CLI application
if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
}
```

### Running Programmatically

To run a command programmatically (without Cobra):

#### Step 1: Create the command instance

```go
cmd, err := NewMyCommand()
if err != nil {
    log.Fatalf("Error creating command: %v", err)
}
```

#### Step 2: Set up the execution context

```go
ctx := context.Background()
```

#### Step 3: Define parameter values (optional)

```go
// Provide parameter values
parseOptions := []runner.ParseOption{
    // Set specific values
    runner.WithValuesForLayers(map[string]map[string]interface{}{
        "default": {
            "count": 20,
            "input": "file.txt",
        },
    }),
    
    // Load from environment variables
    runner.WithEnvMiddleware("MYAPP_"),
}
```

#### Step 4: Define output options (optional)

```go
// Configure output
runOptions := []runner.RunOption{
    // Set output writer for WriterCommand
    runner.WithWriter(os.Stdout),
}
```

#### Step 5: Run the command

```go
// Run the command with the defined options
err = runner.ParseAndRun(ctx, cmd, parseOptions, runOptions)
if err != nil {
    log.Fatalf("Error running command: %v", err)
}
```

## Parameter Loading Sources

When using the runner package, parameters can be loaded from multiple sources:

1. **Default values** (from parameter definitions)
2. **Explicit values** (via `runner.WithValuesForLayers`)
3. **Environment variables** (via `runner.WithEnvMiddleware("PREFIX_")`)
4. **Viper configuration** (via `runner.WithViper()`)
5. **Custom middleware** (via `runner.WithAdditionalMiddlewares`)

Sources later in the list override earlier ones, allowing flexible configuration.

## Defining Commands in YAML

For simpler use cases or dynamic command loading, Glazed allows defining commands in YAML files:

### Example Command YAML

```yaml
name: list-items
short: List items with optional filtering
# Define flags (optional parameters)
flags:
  - name: tag
    type: string
    help: Filter by tag
    default: ""
  - name: limit
    type: integer
    help: Maximum number of items
    default: 50
# Define arguments (required parameters)
arguments:
  - name: source
    type: string
    help: Data source location
    required: true
```

### Loading Commands from YAML

To load commands from YAML files:

#### Step 1: Create a command loader

```go
// Create a YAML command loader
yamlLoader := loaders.NewYAMLCommandLoader()
```

#### Step 2: Load commands from a file or directory

```go
// Load from a specific file
commands, err := yamlLoader.LoadCommands(
    os.DirFS("."),    // Working directory as filesystem
    "commands.yaml",  // File to load
    nil,              // Optional command description modifiers
    nil,              // Optional alias options
)
if err != nil {
    log.Fatalf("Error loading commands: %v", err)
}
```

#### Step 3: Integrate with your application

```go
// Integrate loaded commands with Cobra
for _, cmd := range commands {
    cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
    if err != nil {
        log.Fatalf("Error building Cobra command: %v", err)
    }
    rootCmd.AddCommand(cobraCmd)
}
```

## Advanced Topic: JSON Schema Generation

Glazed commands can generate JSON Schema representations of their parameters, useful for documentation, validation, and UI generation:

```go
// Get command instance
cmd, err := NewMyCommand()
if err != nil {
    log.Fatalf("Error creating command: %v", err)
}

// Generate JSON schema
schema, err := cmd.Description().ToJsonSchema()
if err != nil {
    log.Fatalf("Error generating schema: %v", err)
}

// Output the schema
encoder := json.NewEncoder(os.Stdout)
encoder.SetIndent("", "  ")
encoder.Encode(schema)
```

The generated schema includes:
- Parameter types and descriptions
- Required vs. optional parameters
- Default values
- Enum choices for choice parameters
- Schema structure for complex parameters

## Complete Examples

Let's walk through complete examples of creating and using different command types:

### Example 1: GlazeCommand

### Step 1: Define the command and settings

```go
// UserListCommand lists users with filtering options
type UserListCommand struct {
    *cmds.CommandDescription
}

// Settings for the command
type UserListSettings struct {
    Limit   int    `glazed.parameter:"limit"`
    Filter  string `glazed.parameter:"filter"`
    Verbose bool   `glazed.parameter:"verbose"`
}

// Ensure interface implementation
var _ cmds.GlazeCommand = &UserListCommand{}
```

### Step 2: Implement the processor method

```go
func (c *UserListCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Parse settings
    s := &UserListSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Log if verbose
    if s.Verbose {
        fmt.Println("Fetching users with limit:", s.Limit)
    }
    
    // Simulate fetching users (in a real app, this would be a database call)
    users := []struct {
        ID       int
        Username string
        Email    string
        Role     string
    }{
        {1, "admin", "admin@example.com", "admin"},
        {2, "user1", "user1@example.com", "user"},
        {3, "user2", "user2@example.com", "user"},
        {4, "editor", "editor@example.com", "editor"},
        {5, "guest", "guest@example.com", "guest"},
    }
    
    // Apply filter if provided
    var filteredUsers []struct {
        ID       int
        Username string
        Email    string
        Role     string
    }
    
    if s.Filter != "" {
        for _, user := range users {
            if strings.Contains(user.Username, s.Filter) || 
               strings.Contains(user.Email, s.Filter) ||
               strings.Contains(user.Role, s.Filter) {
                filteredUsers = append(filteredUsers, user)
            }
        }
    } else {
        filteredUsers = users
    }
    
    // Apply limit
    if s.Limit > 0 && s.Limit < len(filteredUsers) {
        filteredUsers = filteredUsers[:s.Limit]
    }
    
    // Output as rows
    for _, user := range filteredUsers {
        row := types.NewRowFromStruct(&user, true)
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Step 3: Create the constructor

```go
func NewUserListCommand() (*UserListCommand, error) {
    // Create the Glazed layer for output formatting
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    // Create command description
    cmdDesc := cmds.NewCommandDescription(
        "list-users",
        cmds.WithShort("List users with optional filtering"),
        cmds.WithLong(`
List all users in the system with optional filtering by name, email or role.
Supports various output formats through standard Glazed flags.

Examples:
  list-users --limit=10
  list-users --filter=admin --output=json
  list-users --filter=user --fields=username,email
        `),
        // Define command flags
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "limit",
                parameters.ParameterTypeInteger,
                parameters.WithHelp("Maximum number of users to return"),
                parameters.WithDefault(0),
            ),
            parameters.NewParameterDefinition(
                "filter",
                parameters.ParameterTypeString,
                parameters.WithHelp("Filter users by username, email or role"),
                parameters.WithDefault(""),
            ),
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Enable verbose output"),
                parameters.WithDefault(false),
            ),
        ),
        // Add parameter layers
        cmds.WithLayersList(
            glazedLayer,
        ),
    )
    
    return &UserListCommand{
        CommandDescription: cmdDesc,
    }, nil
}
```

### Step 4: Use in a CLI application

```go
package main

import (
    "fmt"
    "os"
    
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/help"
    help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
    "github.com/spf13/cobra"
)

func main() {
    // Create root command
    rootCmd := &cobra.Command{
        Use:   "userapp",
        Short: "User management application",
    }
    
    // Initialize help system
    helpSystem := help.NewHelpSystem()
    help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
    
    // Create the user list command
    userListCmd, err := NewUserListCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
        os.Exit(1)
    }
    
    // Convert to Cobra command
    userListCobraCmd, err := cli.BuildCobraCommandFromCommand(userListCmd)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
        os.Exit(1)
    }
    
    // Add to root command
    rootCmd.AddCommand(userListCobraCmd)
    
    // Execute
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### Example 2: Dual Command

Now let's create a dual command that can output both human-readable text and structured data:

#### Step 1: Define the dual command structure

```go
// StatusCommand shows system status information
type StatusCommand struct {
    *cmds.CommandDescription
}

// Settings for the command
type StatusSettings struct {
    Verbose bool `glazed.parameter:"verbose"`
    Format  bool `glazed.parameter:"format-bytes"`
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}
```

#### Step 2: Implement both run methods

```go
// Classic mode - human-readable output
func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Simulate getting system status
    status := getSystemStatus()
    
    fmt.Printf("System Status:\n")
    fmt.Printf("  CPU Usage: %.1f%%\n", status.CPUUsage)
    if s.FormatBytes {
        fmt.Printf("  Memory: %s / %s\n", formatBytes(status.MemoryUsed), formatBytes(status.MemoryTotal))
        fmt.Printf("  Disk: %s / %s\n", formatBytes(status.DiskUsed), formatBytes(status.DiskTotal))
    } else {
        fmt.Printf("  Memory: %d / %d MB\n", status.MemoryUsed/1024/1024, status.MemoryTotal/1024/1024)
        fmt.Printf("  Disk: %d / %d GB\n", status.DiskUsed/1024/1024/1024, status.DiskTotal/1024/1024/1024)
    }
    fmt.Printf("  Uptime: %s\n", status.Uptime)
    
    if s.Verbose {
        fmt.Printf("  Processes: %d\n", status.ProcessCount)
        fmt.Printf("  Load Average: %.2f\n", status.LoadAverage)
    }
    
    return nil
}

// Glaze mode - structured data output
func (c *StatusCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Get system status
    status := getSystemStatus()
    
    // Create a row with all status information
    row := types.NewRow(
        types.MRP("cpu_usage", status.CPUUsage),
        types.MRP("memory_used", status.MemoryUsed),
        types.MRP("memory_total", status.MemoryTotal),
        types.MRP("disk_used", status.DiskUsed),
        types.MRP("disk_total", status.DiskTotal),
        types.MRP("uptime", status.Uptime.String()),
        types.MRP("process_count", status.ProcessCount),
        types.MRP("load_average", status.LoadAverage),
    )
    
    return gp.AddRow(ctx, row)
}

// Helper struct for system status
type SystemStatus struct {
    CPUUsage     float64
    MemoryUsed   int64
    MemoryTotal  int64
    DiskUsed     int64
    DiskTotal    int64
    Uptime       time.Duration
    ProcessCount int
    LoadAverage  float64
}

// Simulate getting system status (in real app, this would query the system)
func getSystemStatus() SystemStatus {
    return SystemStatus{
        CPUUsage:     45.2,
        MemoryUsed:   8 * 1024 * 1024 * 1024,  // 8GB
        MemoryTotal:  16 * 1024 * 1024 * 1024, // 16GB
        DiskUsed:     250 * 1024 * 1024 * 1024, // 250GB
        DiskTotal:    500 * 1024 * 1024 * 1024, // 500GB
        Uptime:       time.Hour * 24 * 7,        // 7 days
        ProcessCount: 156,
        LoadAverage:  1.23,
    }
}

func formatBytes(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

#### Step 3: Create the constructor

```go
func NewStatusCommand() (*StatusCommand, error) {
    cmdDesc := cmds.NewCommandDescription(
        "status",
        cmds.WithShort("Show system status information"),
        cmds.WithLong(`
Show system status including CPU, memory, disk usage and uptime.
Can output in human-readable format (default) or structured data format.

Examples:
  status                                    # Human-readable output
  status --verbose                          # Include additional details
  status --with-glaze-output                # Structured table output
  status --with-glaze-output --output=json  # JSON output
        `),
        
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Show additional status details"),
                parameters.WithDefault(false),
            ),
            parameters.NewParameterDefinition(
                "format-bytes",
                parameters.ParameterTypeBool,
                parameters.WithHelp("Format byte values as human-readable (KB, MB, GB)"),
                parameters.WithDefault(true),
            ),
        ),
    )
    
    return &StatusCommand{
        CommandDescription: cmdDesc,
    }, nil
}
```

#### Step 4: Use in a CLI application with dual mode

```go
func main() {
    rootCmd := &cobra.Command{
        Use:   "sysmon",
        Short: "System monitoring tool",
    }
    
    // Create the status command
    statusCmd, err := NewStatusCommand()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
        os.Exit(1)
    }
    
    // Build using dual mode - this allows both classic and glaze output
    statusCobraCmd, err := cli.BuildCobraCommandDualMode(
        statusCmd,
        cli.WithGlazeToggleFlag("with-glaze-output"),
    )
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error building command: %v\n", err)
        os.Exit(1)
    }
    
    rootCmd.AddCommand(statusCobraCmd)
    
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}
```

#### Step 5: Usage examples

```bash
# Classic mode - human-readable output
$ sysmon status
System Status:
  CPU Usage: 45.2%
  Memory: 8.0 GB / 16.0 GB
  Disk: 250.0 GB / 500.0 GB
  Uptime: 168h0m0s

# Classic mode with verbose details
$ sysmon status --verbose
System Status:
  CPU Usage: 45.2%
  Memory: 8.0 GB / 16.0 GB
  Disk: 250.0 GB / 500.0 GB
  Uptime: 168h0m0s
  Processes: 156
  Load Average: 1.23

# Glaze mode - structured table output
$ sysmon status --with-glaze-output
+-----------+-------------+--------------+-----------+------------+----------+---------------+--------------+
| cpu_usage | memory_used | memory_total | disk_used | disk_total | uptime   | process_count | load_average |
+-----------+-------------+--------------+-----------+------------+----------+---------------+--------------+
| 45.2      | 8589934592  | 17179869184  | 268435456000 | 536870912000 | 168h0m0s | 156           | 1.23         |
+-----------+-------------+--------------+-----------+------------+----------+---------------+--------------+

# Glaze mode - JSON output
$ sysmon status --with-glaze-output --output=json
[
{
  "cpu_usage": 45.2,
  "disk_total": 536870912000,
  "disk_used": 268435456000,
  "load_average": 1.23,
  "memory_total": 17179869184,
  "memory_used": 8589934592,
  "process_count": 156,
  "uptime": "168h0m0s"
}
]

# Glaze mode - specific fields only
$ sysmon status --with-glaze-output --fields=cpu_usage,memory_used,memory_total
+-----------+-------------+--------------+
| cpu_usage | memory_used | memory_total |
+-----------+-------------+--------------+
| 45.2      | 8589934592  | 17179869184  |
+-----------+-------------+--------------+
```

## Best Practices

To get the most out of Glazed, follow these best practices:

1. **Choose the Right Command Type**: 
   - Use **BareCommand** for simple actions without structured output
   - Use **WriterCommand** when you need to write text to a specific destination
   - Use **GlazeCommand** for structured data that benefits from multiple output formats
   - Use **Dual Commands** when you want to offer both human-readable and structured output

2. **Consider Dual Commands for Maximum Flexibility**: If your command could benefit from both human-readable output (for interactive use) and structured output (for scripting/automation), implement both `BareCommand` and `GlazeCommand` interfaces and use `BuildCobraCommandDualMode`.

3. **Organize Parameters into Layers**: Group related parameters together in logical layers for better organization and reusability.

4. **Create Settings Structs**: Use struct tags (`glazed.parameter`) to map parameters to fields for cleaner, more maintainable code.

5. **Provide Helpful Documentation**: Use `WithShort` and `WithLong` to describe your command, and `WithHelp` for each parameter.

6. **Set Sensible Defaults**: Use `WithDefault` to provide reasonable default values when possible.

7. **For GlazeCommands, Use Row Creation Helpers**: Leverage `types.NewRow`, `types.MRP`, and `types.NewRowFromStruct` for clean code.

8. **Use Proper Error Handling**: Wrap errors with context using `errors.Wrap` from the `github.com/pkg/errors` package.

9. **Add Validation**: Check parameter values and return meaningful error messages.

10. **Design Consistent Dual Command Interfaces**: When implementing dual commands, ensure that both output modes provide equivalent information, just formatted differently.

## Summary

Glazed provides a powerful framework for building command-line applications with rich features:

- Multiple command types for different output needs (BareCommand, WriterCommand, GlazeCommand)
- **Dual commands** that can switch between classic and structured output modes at runtime
- Flexible parameter system with type validation
- Layered parameter organization
- Structured data output with multiple formats (JSON, YAML, CSV, table, etc.)
- Integration with Cobra for CLI applications
- Programmatic execution
- Declarative YAML configuration
- JSON Schema generation

The new dual command functionality (`BuildCobraCommandDualMode`) is particularly useful for creating commands that serve both interactive users (who prefer human-readable output) and automation scripts (which need structured data), all while maintaining a single codebase and registration point.

By following this guide, you should now have a solid understanding of how to create, configure, and run commands using Glazed. Whether you're building simple utilities or complex applications, Glazed provides the tools to streamline your command-line interface development.
