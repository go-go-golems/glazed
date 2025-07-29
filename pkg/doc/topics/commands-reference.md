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

The Glazed command system provides a structured approach to building CLI applications that handle multiple output formats, complex parameter validation, and reusable components. This reference covers the complete command system architecture, interfaces, and implementation patterns.

Building CLI tools typically involves handling parameter parsing, validation, output formatting, and configuration management. Glazed addresses these concerns through a layered architecture that separates command logic from presentation and parameter management.

The core principle is separation of concerns: commands focus on business logic while Glazed handles parameter parsing, validation, and output formatting. This approach enables automatic support for multiple output formats, consistent parameter handling across commands, and reusable parameter groups.

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

1. **Command Interfaces**: Define how commands produce output - direct output (BareCommand), text streams (WriterCommand), or structured data (GlazeCommand)
2. **CommandDescription**: Contains metadata about a command (name, description, parameters, etc.)
3. **Parameter Layers**: Organize parameters into logical groups (database, logging, output, etc.)
4. **Parameters**: Define command inputs with type information, validation, and help text
5. **ParsedLayers**: Runtime values after collecting from CLI flags, environment, config files, and defaults
6. **Execution Methods**: Different approaches for running commands - CLI integration, programmatic execution, etc.

## Core Packages

The Glazed framework is organized into distinct packages to separate concerns like command definition, parameter handling, and output processing. This modular design makes the system extensible and easier to maintain. Key packages handle command interfaces (`cmds`), parameter definitions (`parameters`), and integration with CLI libraries like Cobra (`cli`).

- `github.com/go-go-golems/glazed/pkg/cmds`: Core command interfaces and descriptions
- `github.com/go-go-golems/glazed/pkg/cmds/parameters`: Parameter types and definitions
- `github.com/go-go-golems/glazed/pkg/cmds/layers`: Parameter layering system
- `github.com/go-go-golems/glazed/pkg/cmds/runner`: Programmatic command execution
- `github.com/go-go-golems/glazed/pkg/middlewares`: Output processing
- `github.com/go-go-golems/glazed/pkg/types`: Structured data types
- `github.com/go-go-golems/glazed/pkg/cli`: Helpers for integrating with Cobra
- `github.com/go-go-golems/glazed/pkg/settings`: Standard Glazed parameter layers

## Command Interfaces

A command's output contract is defined by the interface it implements. Glazed offers three primary interfaces to support different use cases: `BareCommand` for direct `stdout` control, `WriterCommand` for sending text to any `io.Writer`, and `GlazeCommand` for producing structured data that can be automatically formatted. This design allows a command's business logic to be decoupled from its final output format.

Glazed provides three main command interfaces, each designed for different output scenarios:

1. **BareCommand**: For commands that handle their own output directly
2. **WriterCommand**: For commands that write to a provided writer (like a file or console)  
3. **GlazeCommand**: For commands that produce structured data using Glazed's processing pipeline

Additionally, **Dual Commands** can implement multiple interfaces and switch between output modes at runtime based on flags.

### BareCommand

BareCommand provides direct control over output. Commands implementing this interface handle their own output formatting and presentation.

```go
// BareCommand interface definition
type BareCommand interface {
    Command
    Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error
}
```

**Use cases:**
- Commands that print progress updates or status messages
- Utilities that perform actions (file operations, service management) rather than data processing
- Commands requiring custom output formatting
- Quick prototypes and simple utilities

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
    
    fmt.Printf("Starting cleanup in %s...\n", s.Directory)
    
    files, err := findOldFiles(s.Directory, s.OlderThan)
    if err != nil {
        return fmt.Errorf("failed to scan directory: %w", err)
    }
    
    if len(files) == 0 {
        fmt.Printf("Directory is clean - no files older than %s found.\n", s.OlderThan)
        return nil
    }
    
    fmt.Printf("Found %d files to clean up:\n", len(files))
    for i, file := range files {
        fmt.Printf("  %d. %s\n", i+1, file)
        if !s.DryRun {
            if err := os.Remove(file); err != nil {
                fmt.Printf("     Failed to remove: %s\n", err)
            } else {
                fmt.Printf("     Removed\n")
            }
        }
    }
    
    if s.DryRun {
        fmt.Printf("Dry run completed. Use --execute to actually remove files.\n")
    } else {
        fmt.Printf("Cleanup completed successfully.\n")
    }
    
    return nil
}
```

### WriterCommand

WriterCommand allows commands to write text output to any destination (files, stdout, network connections) without knowing the specific target.

```go
// WriterCommand interface definition
type WriterCommand interface {
    Command
    RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error
}
```

This interface separates content generation from output destination, improving testability and reusability.

**Use cases:**
- Report generators that output to files or stdout
- Log processors that transform and forward data
- Commands generating substantial text output (documentation, configuration files)
- Commands where output destination should be configurable

**Example implementation:**
```go
type HealthReportCommand struct {
    *cmds.CommandDescription
}

func (c *HealthReportCommand) RunIntoWriter(ctx context.Context, parsedLayers *layers.ParsedLayers, w io.Writer) error {
    s := &HealthReportSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Generate a comprehensive system health report
    fmt.Fprintf(w, "System Health Report\n")
    fmt.Fprintf(w, "Generated: %s\n", time.Now().Format(time.RFC3339))
    fmt.Fprintf(w, "Host: %s\n\n", s.Hostname)
    
    // Check various system components
    components := []string{"CPU", "Memory", "Disk", "Network"}
    for _, component := range components {
        status, details := checkComponentHealth(component)
        fmt.Fprintf(w, "%s Status: %s\n", component, status)
        if s.Verbose {
            fmt.Fprintf(w, "  Details: %s\n", details)
        }
    }
    
    // Add recommendations if any issues found
    if recommendations := generateRecommendations(); len(recommendations) > 0 {
        fmt.Fprintf(w, "\nRecommendations:\n")
        for i, rec := range recommendations {
            fmt.Fprintf(w, "%d. %s\n", i+1, rec)
        }
    }
    
    return nil
}
```

This command can write its report to a file for archival, to stdout for immediate viewing, or even to a network connection for monitoring systems. The command logic doesn't change - only the destination.

### GlazeCommand

GlazeCommand produces structured data that Glazed processes into various output formats. Commands generate data rows instead of formatted text, enabling automatic format conversion and data processing.

```go
// GlazeCommand interface definition
type GlazeCommand interface {
    Command
    RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error
}
```

GlazeCommand generates structured data events that can be transformed, filtered, sorted, and formatted automatically.

**Key capabilities:**
- **Automatic formatting**: JSON, YAML, CSV, HTML tables, custom templates without additional code
- **Data processing**: Built-in filtering, sorting, column selection, and transformation
- **Composability**: Output can be piped to other tools or processed programmatically
- **Format independence**: New output formats can be added without changing command implementation

**Real-world example - A server monitoring command:**
```go
type MonitorServersCommand struct {
    *cmds.CommandDescription
}

func (c *MonitorServersCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    s := &MonitorSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    
    // Get server data from various sources
    servers := getServersFromInventory(s.Environment)
    
    for _, server := range servers {
        // Check server health
        health := checkServerHealth(server.Hostname)
        
        // Produce a rich data row with nested information
        row := types.NewRow(
            types.MRP("hostname", server.Hostname),
            types.MRP("environment", server.Environment),
            types.MRP("cpu_percent", health.CPUPercent),
            types.MRP("memory_used_gb", health.MemoryUsedGB),
            types.MRP("memory_total_gb", health.MemoryTotalGB),
            types.MRP("disk_used_percent", health.DiskUsedPercent),
            types.MRP("status", health.Status),
            types.MRP("last_seen", health.LastSeen),
            types.MRP("alerts", health.ActiveAlerts), // Can be an array
            types.MRP("metadata", map[string]interface{}{ // Nested objects work too
                "os_version": server.OSVersion,
                "kernel": server.KernelVersion,
                "uptime_days": health.UptimeDays,
            }),
        )
        
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

Now your users can run:
- `monitor --output table` for a human-readable overview
- `monitor --output json | jq '.[] | select(.status != "healthy")'` to find problem servers  
- `monitor --output csv > servers.csv` to import into spreadsheets
- `monitor --filter 'cpu_percent > 80' --sort cpu_percent` to find CPU hotspots
- `monitor --template custom.tmpl` to generate custom reports

All from the same command implementation.

### Dual Commands

Dual commands implement multiple interfaces and switch between output modes based on runtime flags. This approach provides both human-readable text output and structured data from a single command.

Dual commands address the need for different output formats: interactive use typically requires readable text output, while scripts need structured data. Rather than maintaining separate commands, dual commands adapt their behavior based on context.

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

## Command Implementation

A well-structured Glazed command separates its identity and logic. The recommended pattern involves a `Command` struct embedding a `CommandDescription` for metadata, a separate `Settings` struct for type-safe parameter access via `glazed.parameter` tags, and a `Run` method containing the business logic. This separation is bridged at runtime by `InitializeStruct`, which populates the `Settings` struct from parsed command-line values.

Glazed commands follow a consistent structure with four key components:

### Command Structure

**Command Struct**: Contains the command's identity and embeds `CommandDescription` which holds metadata (name, flags, help text) separately from business logic.

**Settings Struct**: Provides type safety by defining a struct that mirrors command inputs. Glazed automatically maps parameters to struct fields through `glazed.parameter` tags.

**Run Method**: Contains business logic. The method signature depends on the implemented interface, but the pattern is consistent: extract settings using `InitializeStruct`, execute logic, return results.

**Constructor Function**: Creates the command description with its parameters and layers.

### Settings Structs and InitializeStruct Pattern

Settings structs provide type-safe access to parsed command parameters. Each field uses a `glazed.parameter` tag that must match the parameter name defined in the command description:

```go
// Settings struct with glazed.parameter tags
type MyCommandSettings struct {
    Count   int    `glazed.parameter:"count"`     // Maps to "count" parameter
    Format  string `glazed.parameter:"format"`   // Maps to "format" parameter  
    Verbose bool   `glazed.parameter:"verbose"`  // Maps to "verbose" parameter
    DryRun  bool   `glazed.parameter:"dry-run"`  // Maps to "dry-run" parameter
}
```

The `InitializeStruct` method populates the settings struct from parsed layers. Always specify the correct layer slug (use `layers.DefaultSlug` for command-specific parameters):

```go
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Create settings struct instance
    s := &MyCommandSettings{}
    
    // Extract values from the "default" layer into the struct
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return fmt.Errorf("failed to initialize settings: %w", err)
    }
    
    // Now use the populated struct fields
    for i := 0; i < s.Count; i++ {
        if s.Verbose {
            log.Printf("Processing item %d with format %s", i, s.Format)
        }
        
        if s.DryRun {
            fmt.Printf("Would process item %d\n", i)
            continue
        }
        
        row := types.NewRow(
            types.MRP("id", i),
            types.MRP("format", s.Format),
        )
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Working with Multiple Layers

Commands often use multiple parameter layers. Extract settings from each layer separately:

```go
type DatabaseSettings struct {
    Host     string `glazed.parameter:"db-host"`
    Port     int    `glazed.parameter:"db-port"`
    Database string `glazed.parameter:"db-name"`
}

type LoggingSettings struct {
    Level string `glazed.parameter:"log-level"`
    File  string `glazed.parameter:"log-file"`
}

func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Extract command-specific settings
    cmdSettings := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, cmdSettings); err != nil {
        return err
    }
    
    // Extract database layer settings
    dbSettings := &DatabaseSettings{}
    if err := parsedLayers.InitializeStruct("database", dbSettings); err != nil {
        return err
    }
    
    // Extract logging layer settings
    logSettings := &LoggingSettings{}
    if err := parsedLayers.InitializeStruct("logging", logSettings); err != nil {
        return err
    }
    
    // Use all settings together
    fmt.Printf("Connecting to %s:%d/%s with log level %s\n", 
        dbSettings.Host, dbSettings.Port, dbSettings.Database, logSettings.Level)
    
    // ... rest of command logic
    return nil
}
```

### Common InitializeStruct Patterns

**Pattern 1: Inline struct definition (for simple cases)**
```go
func (c *ExampleCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := struct {
        Message string `glazed.parameter:"message"`
        Count   int    `glazed.parameter:"count"`
    }{}

    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, &s); err != nil {
        return err
    }

    // Use s.Message and s.Count
    return nil
}
```

**Pattern 2: Named settings struct (recommended for complex commands)**
```go
type ExampleSettings struct {
    Message string `glazed.parameter:"message"`
    Count   int    `glazed.parameter:"count"`
}

func (c *ExampleCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &ExampleSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }

    // Use s.Message and s.Count  
    return nil
}
```

**Pattern 3: Helper function for reusable settings**
```go
func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedLayers.InitializeStruct("database", settings)
    return settings, err
}

func (c *MyCommand) RunIntoGlazeProcessor(...) error {
    dbSettings, err := GetDatabaseSettings(parsedLayers)
    if err != nil {
        return err
    }
    // Use dbSettings
    return nil
}
```

## Advanced Patterns

While Glazed excels at building standard CLI tools, its architecture also supports more advanced use cases. Commands can be executed programmatically for testing or integration into other Go applications, and the parameter system can load values from multiple sources like environment variables and config files, not just CLI flags. These patterns allow you to build commands that are not just standalone tools, but reusable components in a larger software ecosystem.

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

### Structured Data Output (GlazeCommand)

#### Creating Rows

For GlazeCommands, rows form the structured output:

##### Using NewRow with MRP (MapRowPair)
```go
row := types.NewRow(
    types.MRP("id", 1),
    types.MRP("name", "John Doe"),
    types.MRP("email", "john@example.com"),
    types.MRP("active", true),
)
```

##### From a map
```go
data := map[string]interface{}{
    "id":     1,
    "name":   "John Doe",
    "email":  "john@example.com",
    "active": true,
}
row := types.NewRowFromMap(data)
```

##### From a struct
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

#### Adding Rows to Processor

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

### Error Handling Patterns

#### Graceful Error Handling

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

#### Exit Control

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

### Performance Considerations

#### Efficient Row Creation

For large datasets, optimize row creation:

```go
// Good: Processes data as it arrives
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    row := processLine(scanner.Text())
    if err := gp.AddRow(ctx, row); err != nil {
        return err
    }
}

// Problematic: Loads everything into memory first
allData := loadAllDataIntoMemory() // What if this is 10GB?
for _, item := range allData {
    // Process items...
}
```

#### Memory Management

For commands processing large amounts of data:

```go
// Good: Processes data as it arrives
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    row := processLine(scanner.Text())
    if err := gp.AddRow(ctx, row); err != nil {
        return err
    }
}

// Problematic: Loads everything into memory first
allData := loadAllDataIntoMemory() // What if this is 10GB?
for _, item := range allData {
    // Process items...
}
```

## Parameters

Glazed treats command-line parameters as more than just strings. They are typed objects with built-in validation, default values, and help text. This approach shifts the burden of parsing and validation from the command's business logic to the framework itself. By defining a parameter's type (e.g., `ParameterTypeInteger`, `ParameterTypeDate`, `ParameterTypeFile`), you get automatic error handling and a more robust and user-friendly CLI.

Glazed parameters are typed objects with validation rules and behavior, unlike traditional CLI libraries that treat parameters as simple strings requiring manual parsing and validation. This enables automatic validation, help generation, and multi-source value loading.

### Parameter Type System

Parameter types define data structure, parsing behavior, and validation rules. Each type handles string parsing, validation, and help text generation.

#### Basic Types
**`ParameterTypeString`**: The workhorse for text inputs - names, descriptions, URLs
**`ParameterTypeSecret`**: Like strings, but values are masked in help and logs (perfect for passwords, API keys)
**`ParameterTypeInteger`**: Whole numbers with automatic range validation
**`ParameterTypeFloat`**: Decimal numbers for measurements, percentages, ratios
**`ParameterTypeBool`**: True/false flags that work with `--flag` and `--no-flag` patterns
**`ParameterTypeDate`**: Intelligent date parsing that handles multiple formats

#### Collection Types
**`ParameterTypeStringList`**: Multiple values like `--tag web --tag api --tag production`
**`ParameterTypeIntegerList`**: Lists of numbers for ports, IDs, quantities
**`ParameterTypeFloatList`**: Multiple decimal values for coordinates, measurements

#### Choice Types  
**`ParameterTypeChoice`**: Single selection from predefined options (with tab completion!)
**`ParameterTypeChoiceList`**: Multiple selections from predefined options

#### File Types
**`ParameterTypeFile`**: File paths with existence validation and tab completion
**`ParameterTypeFileList`**: Multiple file paths
**`ParameterTypeStringFromFile`**: Read text content from a file (useful for large inputs)
**`ParameterTypeStringListFromFile`**: Read line-separated lists from files

#### Special Types
**`ParameterTypeKeyValue`**: Map-like inputs: `--env DATABASE_URL=postgres://... --env DEBUG=true`

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

Glazed commands are defined independently of any specific CLI library, but they are most commonly used with Cobra. The `pkg/cli` package provides a bridge to convert a `cmds.Command` interface into a `cobra.Command`. This bridge automatically sets up flags, argument handling, and the execution flow, allowing you to benefit from Glazed's features within a standard Cobra application structure.

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

## Best Practices

Building effective command-line tools involves more than just making them work. A great CLI is maintainable, performant, and user-friendly. The following guidelines represent key design principles and patterns for building high-quality Glazed applications, from choosing the right command interface to writing clear documentation and handling errors gracefully.

### Interface Selection

Choose interfaces based on user requirements:

- **BareCommand** when users need rich feedback, progress updates, or interactive elements
- **WriterCommand** when output might go to files, logs, or other destinations
- **GlazeCommand** when data will be processed, filtered, or integrated with other tools
- **Dual Commands** when usage patterns vary

Example: A backup command might start as BareCommand for user feedback (`Backing up 1,247 files...`), but users eventually want structured output for monitoring scripts. A dual command serves both needs.

### Type Safety

Use settings structs with `glazed.parameter` tags to prevent type conversion errors:

```go
// Good: Type-safe and clear
type BackupSettings struct {
    Source      string        `glazed.parameter:"source"`
    Destination string        `glazed.parameter:"destination"`
    MaxAge      time.Duration `glazed.parameter:"max-age"`
    DryRun      bool          `glazed.parameter:"dry-run"`
}

// Avoid: Manual parameter extraction
source, _ := parsedLayers.GetString("source")
maxAge, _ := parsedLayers.GetString("max-age") // Bug waiting to happen!
```

### Defaults and Help

Provide sensible defaults so commands work with minimal flags. If your command requires multiple flags to be useful, reconsider the design.

Write clear help text with examples for complex parameters:

```go
parameters.NewParameterDefinition(
    "filter",
    parameters.ParameterTypeString,
    parameters.WithHelp("Filter results using SQL-like syntax. Examples: 'status = \"active\"', 'created_at > \"2023-01-01\"'"),
)
```

### Error Handling

Provide specific, actionable error messages:

```go
// Good: Specific and actionable
if s.Port < 1 || s.Port > 65535 {
    return fmt.Errorf("port %d is invalid; must be between 1 and 65535", s.Port)
}

// Poor: Vague and frustrating
if !isValidPort(s.Port) {
    return errors.New("invalid port")
}
```

Validate parameters before expensive operations. Always check context cancellation in loops and long operations.

### Performance

Design for streaming to handle large datasets:

```go
// Good: Processes data as it arrives
scanner := bufio.NewScanner(reader)
for scanner.Scan() {
    row := processLine(scanner.Text())
    if err := gp.AddRow(ctx, row); err != nil {
        return err
    }
}

// Problematic: Loads everything into memory first
allData := loadAllDataIntoMemory() // What if this is 10GB?
for _, item := range allData {
    // Process items...
}
```

### Documentation

Command help text is often the primary documentation users read:

- Include realistic examples in the command description
- Explain the purpose, not just the syntax
- Document any side effects or special requirements
- Cross-reference related commands

For GlazeCommands, document the output schema. Users building scripts need to know what fields are available and what they contain.

## Next Steps

1. Start with a hands-on tutorial for a practical introduction:
   ```
   glaze help build-first-command
   ```
2. Study the layer guide to understand parameter organization:
   ```
   glaze help layers-guide
   ```
3. Explore real-world examples in the applications section.
4. Build iteratively—start with something that works, then improve based on actual usage.
