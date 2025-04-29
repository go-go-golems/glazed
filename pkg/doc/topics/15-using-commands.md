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
                      ▼
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

- Basic types: `ParameterTypeString`, `ParameterTypeInteger`, `ParameterTypeBool`, `ParameterTypeFloat`, `ParameterTypeDate`
- Lists: `ParameterTypeStringList`, `ParameterTypeIntegerList`, `ParameterTypeFloatList`
- Choice-based: `ParameterTypeChoice`, `ParameterTypeChoiceList` (with predefined options)
- File-based: `ParameterTypeFile`, `ParameterTypeFileList`, `ParameterTypeStringFromFile`, etc.
- Key-value: `ParameterTypeKeyValue` (for map-like inputs)

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

```go
// Convert the Glazed command to a Cobra command
cobraCmd, err := cli.BuildCobraCommandFromCommand(myCmd)
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

## A Complete Example

Let's walk through a complete example of creating and using a GlazeCommand:

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
    helpSystem.SetupCobraRootCommand(rootCmd)
    
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

## Best Practices

To get the most out of Glazed, follow these best practices:

1. **Use the Right Command Type**: Choose BareCommand, WriterCommand, or GlazeCommand based on your output needs.

2. **Organize Parameters into Layers**: Group related parameters together in logical layers.

3. **Create Settings Structs**: Use struct tags (`glazed.parameter`) to map parameters to fields for cleaner code.

4. **Provide Helpful Documentation**: Use `WithShort` and `WithLong` to describe your command, and `WithHelp` for each parameter.

5. **Set Sensible Defaults**: Use `WithDefault` to provide reasonable default values when possible.

6. **For GlazeCommands, Use Row Creation Helpers**: Leverage `types.NewRow`, `types.MRP`, and `types.NewRowFromStruct` for clean code.

7. **Use Proper Error Handling**: Wrap errors with context using `errors.Wrap` from the `github.com/pkg/errors` package.

8. **Add Validation**: Check parameter values and return meaningful error messages.

## Summary

Glazed provides a powerful framework for building command-line applications with rich features:

- Multiple command types for different output needs
- Flexible parameter system with type validation
- Layered parameter organization
- Structured data output with multiple formats
- Integration with Cobra for CLI applications
- Programmatic execution
- Declarative YAML configuration
- JSON Schema generation

By following this guide, you should now have a solid understanding of how to create, configure, and run commands using Glazed. Whether you're building simple utilities or complex applications, Glazed provides the tools to streamline your command-line interface development.
