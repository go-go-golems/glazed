---
Title: Glazed Commands Reference
Slug: commands-reference
Short: Complete reference for creating, configuring, and running commands in Glazed
Topics:
- commands
- interfaces
- flags
- arguments
- layers
- dual-commands
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Commands Reference

## Overview

The `glazed` framework provides a powerful system for creating command-line interfaces with rich features like structured data output, parameter validation, and multiple output formats. This reference guide covers all aspects of the command system, from basic concepts to advanced usage patterns.

## Architecture of Glazed Commands

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

### BareCommand

This is the simplest interface, suitable when you just need to perform an action without structured output:

```go
// BareCommand interface definition
type BareCommand interface {
    Command
    Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error
}
```

**Use cases:**
- Simple utilities that print basic text output
- Commands that perform actions without returning data
- Quick prototyping and testing

**Example implementation:**
```go
type CleanupCommand struct {
    *cmds.CommandDescription
}

func (c *CleanupCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &CleanupSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    fmt.Printf("Cleaning up %s...\n", s.Directory)
    // Perform cleanup operations
    fmt.Println("Cleanup completed successfully")
    return nil
}
```

### WriterCommand

When you need to write text output to a specific destination:

```go
// WriterCommand interface definition
type WriterCommand interface {
    Command
    RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error
}
```

**Use cases:**
- Commands that generate reports or logs
- File processing utilities
- Commands that stream output

**Example implementation:**
```go
type ReportCommand struct {
    *cmds.CommandDescription
}

func (c *ReportCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
    s := &ReportSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    fmt.Fprintf(w, "Report for %s\n", s.Date)
    fmt.Fprintf(w, "Generated at: %s\n", time.Now().Format(time.RFC3339))
    
    // Write report content to writer
    for i, item := range s.Items {
        fmt.Fprintf(w, "%d. %s\n", i+1, item)
    }
    
    return nil
}
```

### GlazeCommand

For producing structured data that can be formatted in multiple ways (JSON, YAML, CSV, etc.):

```go
// GlazeCommand interface definition
type GlazeCommand interface {
    Command
    RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error
}
```

**Use cases:**
- Commands that return structured data
- APIs and data processing tools
- Commands that need multiple output formats

**Example implementation:**
```go
type ListUsersCommand struct {
    *cmds.CommandDescription
}

func (c *ListUsersCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &ListUsersSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    users := getUsersFromDatabase(s.Filter)
    
    for _, user := range users {
        row := types.NewRow(
            types.MRP("id", user.ID),
            types.MRP("name", user.Name),
            types.MRP("email", user.Email),
            types.MRP("created_at", user.CreatedAt),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Dual Commands

Dual commands implement multiple interfaces, allowing them to run in different output modes based on runtime flags. This is particularly useful when you want to offer both simple text output for human consumption and structured data for programmatic use.

```go
// Dual command example
type StatusCommand struct {
    *cmds.CommandDescription
}

// Implement BareCommand for classic mode
func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Human-readable output
    fmt.Printf("System Status:\n")
    fmt.Printf("  CPU: %.1f%%\n", s.CPUUsage)
    fmt.Printf("  Memory: %s\n", s.MemoryUsage)
    return nil
}

// Implement GlazeCommand for structured output mode
func (c *StatusCommand) RunIntoGlazeProcessor(
    ctx context.Context, 
    parsedLayers *layers.ParsedLayers, 
    gp middlewares.Processor,
) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Structured data output
    row := types.NewRow(
        types.MRP("cpu_usage", s.CPUUsage),
        types.MRP("memory_usage", s.MemoryUsage),
        types.MRP("timestamp", time.Now()),
    )
    return gp.AddRow(ctx, row)
}

// Ensure both interfaces are implemented
var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}
```

## Command Creation Patterns

### Basic Command Structure

All Glazed commands follow this basic pattern:

1. **Define the command struct** that embeds `*cmds.CommandDescription`
2. **Create a settings struct** (optional but recommended)
3. **Implement the appropriate run method(s)**
4. **Create a constructor function**

```go
// Step 1: Define command struct
type MyCommand struct {
    *cmds.CommandDescription
}

// Step 2: Define settings struct
type MyCommandSettings struct {
    Count  int    `glazed.parameter:"count"`
    Format string `glazed.parameter:"format"`
    Verbose bool  `glazed.parameter:"verbose"`
}

// Step 3: Implement run method
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Command logic here
    for i := 0; i < s.Count; i++ {
        row := types.NewRow(types.MRP("id", i))
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}

// Step 4: Constructor function
func NewMyCommand() (*MyCommand, error) {
    glazedLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }
    
    cmdDesc := cmds.NewCommandDescription(
        "mycommand",
        cmds.WithShort("My example command"),
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "count",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(10),
                parameters.WithHelp("Number of items to generate"),
            ),
            parameters.NewParameterDefinition(
                "format",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("json", "yaml", "table"),
                parameters.WithDefault("table"),
                parameters.WithHelp("Output format"),
            ),
            parameters.NewParameterDefinition(
                "verbose",
                parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Enable verbose output"),
            ),
        ),
        cmds.WithLayersList(glazedLayer),
    )
    
    return &MyCommand{
        CommandDescription: cmdDesc,
    }, nil
}

// Ensure interface compliance
var _ cmds.GlazeCommand = &MyCommand{}
```

## Parameter System Reference

### Parameter Types

Glazed supports various parameter types, each affecting how values are parsed and validated:

#### Basic Types
- `ParameterTypeString`: Plain text values
- `ParameterTypeSecret`: String values that are masked when displayed (e.g., passwords)
- `ParameterTypeInteger`: Whole numbers
- `ParameterTypeFloat`: Decimal numbers
- `ParameterTypeBool`: Boolean flags (true/false)
- `ParameterTypeDate`: Date and time values

#### Collection Types
- `ParameterTypeStringList`: Lists of strings
- `ParameterTypeIntegerList`: Lists of integers
- `ParameterTypeFloatList`: Lists of floating point numbers

#### Choice Types
- `ParameterTypeChoice`: Single selection from predefined options
- `ParameterTypeChoiceList`: Multiple selections from predefined options

#### File Types
- `ParameterTypeFile`: File path (validates existence)
- `ParameterTypeFileList`: List of file paths
- `ParameterTypeStringFromFile`: Read string content from file
- `ParameterTypeStringListFromFile`: Read string list from file

#### Special Types
- `ParameterTypeKeyValue`: Key-value pairs (map-like inputs)
- `ParameterTypeDuration`: Time duration values (e.g., "30s", "1h30m")

### Parameter Definition Options

```go
parameters.NewParameterDefinition(
    "parameter-name",                    // Required: parameter name
    parameters.ParameterTypeString,      // Required: parameter type
    
    // Optional configuration
    parameters.WithDefault("default"),   // Default value
    parameters.WithHelp("Description"),  // Help text
    parameters.WithRequired(true),       // Mark as required
    parameters.WithShortFlag("n"),       // Short flag (-n)
    
    // For choice types
    parameters.WithChoices("opt1", "opt2", "opt3"),
    
    // For file types
    parameters.WithFileExtensions(".txt", ".md"),
)
```

### Working with Arguments

Arguments are positional parameters that don't use flags:

```go
cmds.WithArguments(
    parameters.NewParameterDefinition(
        "input-file",
        parameters.ParameterTypeFile,
        parameters.WithHelp("Input file to process"),
        parameters.WithRequired(true),
    ),
    parameters.NewParameterDefinition(
        "output-file",
        parameters.ParameterTypeString,
        parameters.WithHelp("Output file path"),
        parameters.WithRequired(false),
    ),
)
```

## Command Building and Registration

### Integration with Cobra

Glazed provides several ways to convert commands to Cobra commands:

#### Option A: Automatic Builder (Recommended)
```go
// Automatically selects the appropriate builder based on command type
cobraCmd, err := cli.BuildCobraCommandFromCommand(myCmd)
if err != nil {
    log.Fatalf("Error building Cobra command: %v", err)
}
```

#### Option B: Specific Builders
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

#### Option C: Dual Command Builder (For Dual Commands)
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

### Dual Command Builder Options

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

## Running Commands

### Programmatic Execution

To run a command programmatically without Cobra:

```go
// Create command instance
cmd, err := NewMyCommand()
if err != nil {
    log.Fatalf("Error creating command: %v", err)
}

// Set up execution context
ctx := context.Background()

// Define parameter values
parseOptions := []runner.ParseOption{
    runner.WithValuesForLayers(map[string]map[string]interface{}{
        "default": {
            "count": 20,
            "format": "json",
        },
    }),
    runner.WithEnvMiddleware("MYAPP_"),
}

// Configure output
runOptions := []runner.RunOption{
    runner.WithWriter(os.Stdout),
}

// Run the command
err = runner.ParseAndRun(ctx, cmd, parseOptions, runOptions)
if err != nil {
    log.Fatalf("Error running command: %v", err)
}
```

### Parameter Loading Sources

Parameters can be loaded from multiple sources (in priority order):

1. **Command line arguments** (highest priority)
2. **Environment variables** 
3. **Configuration files**
4. **Default values** (lowest priority)

```go
parseOptions := []runner.ParseOption{
    // Load from environment with prefix
    runner.WithEnvMiddleware("MYAPP_"),
    
    // Load from configuration file
    runner.WithViper(),
    
    // Set explicit values
    runner.WithValuesForLayers(map[string]map[string]interface{}{
        "default": {"count": 10},
    }),
    
    // Add custom middleware
    runner.WithAdditionalMiddlewares(customMiddleware),
}
```

## Structured Data Output (GlazeCommand)

### Creating Rows

For GlazeCommands, rows form the structured output:

#### Using NewRow with MRP (MapRowPair)
```go
row := types.NewRow(
    types.MRP("id", 1),
    types.MRP("name", "John Doe"),
    types.MRP("email", "john@example.com"),
    types.MRP("active", true),
)
```

#### From a map
```go
data := map[string]interface{}{
    "id":     1,
    "name":   "John Doe",
    "email":  "john@example.com",
    "active": true,
}
row := types.NewRowFromMap(data)
```

#### From a struct
```go
type User struct {
    ID     int    `json:"id"`
    Name   string `json:"name"`
    Email  string `json:"email"`
    Active bool   `json:"active"`
}

user := User{ID: 1, Name: "John Doe", Email: "john@example.com", Active: true}
row := types.NewRowFromStruct(&user, true) // true = lowercase field names
```

### Adding Rows to Processor

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Process data and create rows
    for _, item := range data {
        row := types.NewRow(
            types.MRP("field1", item.Value1),
            types.MRP("field2", item.Value2),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

## Advanced Patterns

### Conditional Interface Implementation

Commands can implement multiple interfaces and choose behavior at runtime:

```go
type FlexibleCommand struct {
    *cmds.CommandDescription
    mode string // "simple" or "structured"
}

func (c *FlexibleCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Simple text output
    fmt.Println("Simple output mode")
    return nil
}

func (c *FlexibleCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Structured output
    row := types.NewRow(types.MRP("mode", "structured"))
    return gp.AddRow(ctx, row)
}

// Interface assertions
var _ cmds.BareCommand = &FlexibleCommand{}
var _ cmds.GlazeCommand = &FlexibleCommand{}
```

### Command Composition

Create commands that delegate to other commands:

```go
type CompositeCommand struct {
    *cmds.CommandDescription
    subCommands []cmds.Command
}

func (c *CompositeCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    for _, subCmd := range c.subCommands {
        if glazeCmd, ok := subCmd.(cmds.GlazeCommand); ok {
            if err := glazeCmd.RunIntoGlazeProcessor(ctx, parsedLayers, gp); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### Dynamic Command Generation

Generate commands at runtime based on configuration:

```go
func CreateCommandsFromConfig(config *Config) ([]cmds.Command, error) {
    var commands []cmds.Command
    
    for _, cmdConfig := range config.Commands {
        cmd, err := createCommandFromConfig(cmdConfig)
        if err != nil {
            return nil, err
        }
        commands = append(commands, cmd)
    }
    
    return commands, nil
}

func createCommandFromConfig(config CommandConfig) (cmds.Command, error) {
    // Create layers based on config
    var layers []layers.ParameterLayer
    
    for _, layerConfig := range config.Layers {
        layer, err := createLayerFromConfig(layerConfig)
        if err != nil {
            return nil, err
        }
        layers = append(layers, layer)
    }
    
    // Create command description
    cmdDesc := cmds.NewCommandDescription(
        config.Name,
        cmds.WithShort(config.Short),
        cmds.WithLong(config.Long),
        cmds.WithLayersList(layers...),
    )
    
    // Return appropriate command type based on config
    switch config.Type {
    case "bare":
        return &DynamicBareCommand{CommandDescription: cmdDesc}, nil
    case "glaze":
        return &DynamicGlazeCommand{CommandDescription: cmdDesc}, nil
    default:
        return nil, fmt.Errorf("unknown command type: %s", config.Type)
    }
}
```

## Error Handling Patterns

### Graceful Error Handling

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return fmt.Errorf("failed to initialize settings: %w", err)
    }
    
    // Validate settings
    if err := c.validateSettings(s); err != nil {
        return fmt.Errorf("invalid settings: %w", err)
    }
    
    // Process with context cancellation support
    for i := 0; i < s.Count; i++ {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            row := types.NewRow(types.MRP("id", i))
            if err := gp.AddRow(ctx, row); err != nil {
                return fmt.Errorf("failed to add row %d: %w", i, err)
            }
        }
    }
    
    return nil
}

func (c *MyCommand) validateSettings(s *MyCommandSettings) error {
    if s.Count < 0 {
        return errors.New("count must be non-negative")
    }
    if s.Count > 1000 {
        return errors.New("count cannot exceed 1000")
    }
    return nil
}
```

### Exit Control

Commands can control application exit behavior:

```go
func (c *MyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // For early exit without error
    if shouldExit {
        return &cmds.ExitWithoutGlazeError{}
    }
    
    // Normal processing
    return nil
}
```

## Performance Considerations

### Efficient Row Creation

For large datasets, optimize row creation:

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Pre-allocate row slice for better performance
    const batchSize = 1000
    
    for batch := 0; batch < totalBatches; batch++ {
        // Process in batches to avoid memory issues
        batchData := getDataBatch(batch, batchSize)
        
        for _, item := range batchData {
            // Reuse row objects when possible
            row := types.NewRow(
                types.MRP("id", item.ID),
                types.MRP("value", item.Value),
            )
            
            if err := gp.AddRow(ctx, row); err != nil {
                return err
            }
        }
        
        // Check for cancellation between batches
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }
    }
    
    return nil
}
```

### Memory Management

For commands processing large amounts of data:

```go
func (c *LargeDataCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Use streaming approach instead of loading all data at once
    reader, err := openDataReader()
    if err != nil {
        return err
    }
    defer reader.Close()
    
    scanner := bufio.NewScanner(reader)
    for scanner.Scan() {
        // Process one line at a time
        line := scanner.Text()
        data, err := parseLineToData(line)
        if err != nil {
            continue // Skip invalid lines
        }
        
        row := types.NewRowFromStruct(data, true)
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return scanner.Err()
}
```

## Best Practices

### Command Design
1. **Choose the Right Interface**: Use BareCommand for simple output, WriterCommand for text streams, GlazeCommand for structured data, and Dual Commands for maximum flexibility
2. **Use Settings Structs**: Always create settings structs with `glazed.parameter` tags for type safety
3. **Provide Good Defaults**: Set reasonable default values for all optional parameters
4. **Add Comprehensive Help**: Include detailed help text for the command and all parameters
5. **Handle Context Cancellation**: Always check `ctx.Done()` in long-running operations

### Error Handling
1. **Wrap Errors**: Use `fmt.Errorf` with `%w` verb to wrap errors with context
2. **Validate Early**: Validate settings immediately after parsing
3. **Fail Fast**: Return errors as soon as they're encountered
4. **Provide Clear Messages**: Include specific details about what went wrong

### Performance
1. **Stream Large Data**: Don't load large datasets entirely into memory
2. **Use Batching**: Process data in batches for better memory usage
3. **Check Cancellation**: Regularly check for context cancellation in loops
4. **Optimize Row Creation**: Reuse objects when possible for high-throughput scenarios

### Documentation
1. **Document Interfaces**: Clearly specify which interfaces your command implements
2. **Provide Examples**: Include usage examples in command descriptions
3. **Explain Complex Parameters**: Add detailed help for complex or non-obvious parameters
4. **Document Side Effects**: Mention any side effects or requirements in command documentation

This reference provides comprehensive coverage of Glazed's command system. For practical, hands-on learning, see the [Commands Tutorial](../tutorials/build-first-command.md) which walks through building your first command step by step.
