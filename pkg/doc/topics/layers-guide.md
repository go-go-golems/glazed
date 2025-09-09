---
Title: Glazed Command Layers Guide
Slug: layers-guide  
Short: Complete guide to understanding and working with command parameter layers in Glazed
Topics:
- layers
- parameters
- configuration
- organization
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Command Layers: Complete Guide

*Building maintainable CLI applications through modular parameter organization*

## Table of Contents
1. [Overview](#overview)
2. [Parameter Organization Challenges](#parameter-organization-challenges)
3. [Layer System Architecture](#layer-system-architecture)
4. [Core Layer Concepts](#core-layer-concepts)
5. [Layer Types & Components](#layer-types--components)
6. [Creating and Working with Layers](#creating-and-working-with-layers)
7. [Practical Examples](#practical-examples)
8. [Advanced Patterns](#advanced-patterns)
9. [Best Practices](#best-practices)
10. [Testing Layers](#testing-layers)

## Overview

Parameter layers organize related command parameters into reusable groups. This modular approach addresses common CLI development challenges including parameter proliferation, code duplication, naming conflicts, and maintenance complexity in growing applications.

The layer system enables developers to:

- **Organize parameters** logically by functionality (database, logging, output)
- **Reuse parameter definitions** across multiple commands without duplication
- **Compose command interfaces** by combining only required functionality
- **Maintain consistency** through centralized parameter definitions
- **Scale applications** without accumulating technical debt

### Traditional Parameter Management Problems

CLI applications often start simple but accumulate complexity as they grow:

```bash
myapp server --host localhost --port 8080 --db-host postgres --db-port 5432 --log-level info
myapp client --endpoint http://localhost:8080 --timeout 30s --log-level debug  
myapp backup --db-host postgres --db-port 5432 --output backup.sql --log-level warn
```

This traditional approach creates several maintenance issues:

- **Parameter pollution**: Commands inherit irrelevant flags, cluttering help screens
- **Naming conflicts**: Ambiguous flag names require awkward prefixes (`--web-host` vs `--db-host`)
- **Code duplication**: Parameter definitions copy across commands, violating DRY principles
- **Poor organization**: Parameters scatter without logical grouping
- **Maintenance burden**: Changes require updates across multiple command definitions
- **Interface inconsistency**: Similar functionality uses different flag names across commands

```go
// Traditional approach - scattered and repetitive
type ServerCommand struct {
    Host     string // Web server host
    Port     int    // Web server port
    DbHost   string // Database host
    DbPort   int    // Database port
    LogLevel string // Logging level
    Timeout  string // Request timeout
}

type ClientCommand struct {
    Endpoint string // API endpoint
    Timeout  string // Request timeout (duplicated with different meaning)
    LogLevel string // Logging level (duplicated)
    APIKey   string // API authentication
}

type BackupCommand struct {
    DbHost   string // Database host (duplicated again)
    DbPort   int    // Database port (duplicated again)
    Output   string // Output file
    LogLevel string // Logging level (duplicated yet again)
    Format   string // Backup format
}
```

### Glazed Layer Solution

The layer system eliminates these problems by treating parameters as modular components that can be shared, composed, reused, maintained centrally, and extended without breaking existing commands.

```go
// Define reusable layers once
var (
    DatabaseLayer = CreateDatabaseLayer()  // --db-host, --db-port, --db-name
    LoggingLayer  = CreateLoggingLayer()   // --log-level, --log-file, --verbose
    ServerLayer   = CreateServerLayer()    // --host, --port, --timeout
    GlazedLayer   = CreateGlazedLayer()    // --output, --fields, --format
)

// Compose commands by combining relevant layers
ServerCommand := cmds.NewCommandDescription("server",
    cmds.WithLayers(ServerLayer, DatabaseLayer, LoggingLayer))

ClientCommand := cmds.NewCommandDescription("client", 
    cmds.WithLayers(ServerLayer, LoggingLayer))  // No database needed

BackupCommand := cmds.NewCommandDescription("backup",
    cmds.WithLayers(DatabaseLayer, LoggingLayer, GlazedLayer))  // No server needed
```

## Parameter Organization Challenges

Complex CLI applications face predictable parameter management challenges that layers directly address.

### Parameter Explosion

As applications add features, commands accumulate parameters that may not be relevant to their specific functionality. This creates cognitive overhead for users and increases implementation complexity.

### Naming Conflicts and Namespace Issues

Multiple subsystems often require similar parameters (host, port, timeout). Without organization, developers resort to verbose prefixes that reduce usability and create inconsistent interfaces.

### Code Duplication and Maintenance

Parameter definitions scattered across command implementations require manual synchronization. Adding SSL configuration to database connections means updating every command that uses databases.

### Inconsistent User Interfaces

Without centralized parameter definitions, similar functionality develops different flag names across commands, creating inconsistent user experiences and requiring additional documentation.

## Layer System Architecture

The layer system separates parameter definition from runtime value resolution, enabling flexible composition while maintaining type safety.

### System Components

```
┌─────────────────────────────────────────────┐
│                Command                       │ ← CLI command implementation
│  ┌─────────────────────────────────────────┤
│  │            Command Description           │ ← Metadata + layer references
│  │                                         │
│  │  ┌─────────────────────────────────────┤
│  │  │          Parameter Layers           │ ← Layer definitions (design time)
│  │  │                                     │
│  │  │  ┌─────────────┬─────────────────┤   │
│  │  │  │Default Layer│  Custom Layers  │   │ ← Different types of layers
│  │  │  │(flags/args) │  (database,    │   │
│  │  │  │             │   logging, etc.)│   │
│  │  │  └─────────────┴─────────────────┘   │
│  │  └─────────────────────────────────────┘ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
                       │
                       ▼ (at runtime)
┌─────────────────────────────────────────────┐
│              Runtime Parsing                 │ ← Parameter resolution
│                                             │
│  Parameter Sources (in priority order):     │ ← Multi-source configuration
│  1. Command line arguments                  │
│  2. Environment variables                   │
│  3. Configuration files                     │
│  4. Default values from layer definitions   │
└─────────────────────────────────────────────┘
                       │
                       ▼ (final result)
┌─────────────────────────────────────────────┐
│               Parsed Layers                  │ ← Runtime values
│        (type-safe parameter values)         │
└─────────────────────────────────────────────┘
```

This architecture provides clear separation between:

1. **Layer Definitions** (design time): Parameter specifications and constraints
2. **Runtime Parsing** (execution time): Value resolution from multiple sources
3. **Parsed Layers** (application time): Type-safe access to resolved values

### Layer Lifecycle

Layers progress through distinct phases from definition to runtime use:

1. **Definition Phase**: Parameters defined with types, defaults, and validation rules
2. **Composition Phase**: Layers combined into command descriptions
3. **Parsing Phase**: User input resolved against definitions using source priority
4. **Execution Phase**: Commands access type-safe parameter values

## Core Layer Concepts

### Layer Identity and Organization

Every layer requires unique identification and logical organization within the application parameter namespace.

#### Layer Hierarchy

```
CommandDescription
├── Default Layer (always present)
│   ├── Command-specific flags      ← Unique to this command
│   └── Command-specific arguments  ← Command's core functionality
├── Standard Layers (optional, commonly used)
│   ├── Glazed Layer (for structured output)  ← Output formatting
│   ├── Logging Layer (for logging configuration)  ← Debug & monitoring
│   └── Database Layer (for database connections)  ← Data persistence
└── Custom Layers (application-specific)
    ├── Authentication Layer        ← Security & access control
    ├── API Configuration Layer     ← External service integration
    └── Feature Toggle Layer        ← Experimental or optional features
```

This hierarchy enables logical grouping, reusability across commands, extensibility without breaking changes, and automatic propagation of layer updates.

### Parameter Definitions vs. Parsed Values

The system distinguishes between parameter specifications (what's possible) and runtime values (what's actual).

**Parameter Definitions** (specifications stored in layers):
```go
// This defines what's POSSIBLE
parameters.NewParameterDefinition(
    "log-level",                    // Parameter name
    parameters.ParameterTypeChoice, // Data type constraint
    parameters.WithDefault("info"), // Default value
    parameters.WithChoices("debug", "info", "warn", "error"), // Valid options
    parameters.WithHelp("Set the logging level"), // User guidance
)
```

**Parsed Values** (actual runtime values):
```go
// After user runs: myapp --log-level debug
parsedValue := "debug"  // Actual value used by application

// Application accesses the final value:
loggingLayer, ok := parsedLayers.Get("logging")
if ok {
    logLevel, ok := loggingLayer.GetParameter("log-level")
    // logLevel is now "debug" (not "info" default)
}
```

This separation provides:
- **Reusable definitions** across commands
- **Context-specific values** reflecting user choices
- **Automatic validation** based on definition constraints
- **Type safety** from definition through parsed values

### Layer Composition Patterns

Layer composition determines command interfaces by selecting appropriate parameter groups.

#### Basic Composition

Simple addition of required functionality:

```go
// Combine only needed layers
commandDesc := cmds.NewCommandDescription("mycommand",
    cmds.WithLayers(databaseLayer, loggingLayer))
```

#### Conditional Composition

Dynamic layer assembly based on runtime conditions:

```go
// Build layer list based on features
layers := []layers.ParameterLayer{baseLayer}

if needsDatabase {
    layers = append(layers, databaseLayer)
}

if needsOutput {
    layers = append(layers, glazedLayer)
}

commandDesc := cmds.NewCommandDescription("mycommand",
    cmds.WithLayersList(layers...))
```

#### Layer Extension

Building specialized variants from existing layers:

```go
// Extend existing layer without modification
extendedDbLayer := databaseLayer.Clone()
extendedDbLayer.AddFlags(
    parameters.NewParameterDefinition("pool-size", parameters.ParameterTypeInteger),
    parameters.NewParameterDefinition("connection-timeout", parameters.ParameterTypeString),
)
```

## Layer Types & Components

### Built-in Layer Types

Glazed provides standard layer types for common CLI application requirements.

#### Default Layer

The Default Layer contains command-specific parameters unique to individual commands.

- **Purpose**: Command-specific flags and arguments defining core functionality
- **Slug**: `"default"` (constant: `layers.DefaultSlug`) 
- **Creation**: Automatically created with `cmds.WithFlags()` or `cmds.WithArguments()`
- **Use Case**: Parameters fundamental to command operation, unlikely to be shared

```go
// Default layer created automatically
commandDesc := cmds.NewCommandDescription("serve",
    cmds.WithFlags(
        parameters.NewParameterDefinition("port", parameters.ParameterTypeInteger),
        parameters.NewParameterDefinition("host", parameters.ParameterTypeString),
    ),
    cmds.WithArguments(
        parameters.NewParameterDefinition("config-file", parameters.ParameterTypeString),
    ),
)
// Parameters live in default layer, unique to "serve" command
```

#### Glazed Layer

The Glazed Layer provides comprehensive output formatting capabilities for commands producing structured data.

- **Purpose**: Output formatting, filtering, and transformation options
- **Slug**: `"glazed"` (constant: `settings.GlazedSlug`)
- **Creation**: `settings.NewGlazedParameterLayers()`
- **Use Case**: Commands outputting structured data requiring flexible formatting

```go
glazedLayer, err := settings.NewGlazedParameterLayers()
if err != nil {
    return nil, fmt.Errorf("failed to create glazed layer: %w", err)
}
// Provides: --output, --fields, --sort-columns, --filter, etc.
// Users can run: myapp data --output json --fields name,age --filter "age > 25"
```

### Custom Layer Components

Application-specific layers address domain requirements through custom parameter definitions.

#### Parameter Definitions

Individual parameter specifications define acceptable input and behavior:

```go
paramDef := parameters.NewParameterDefinition(
    "connection-timeout",              // Parameter name
    parameters.ParameterTypeString,    // Type constraint (use string for duration parsing)
    parameters.WithDefault("30s"),     // Default value
    parameters.WithHelp("Connection timeout for database operations"), // User guidance
    parameters.WithRequired(false),    // Whether required
    parameters.WithShortFlag("t"),     // Short flag convenience
)
```

Each definition specifies:
- **Name**: Parameter identifier for CLI flags
- **Type**: Data type with automatic validation
- **Default**: Value used when not specified
- **Help**: User guidance for help text
- **Constraints**: Validation rules and requirements

#### Validation and Constraints

Parameter definitions include built-in validation for common patterns:

```go
// Choice parameters with automatic validation
parameters.NewParameterDefinition(
    "log-level",
    parameters.ParameterTypeChoice,
    parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
    parameters.WithDefault("info"),
)
// Automatically rejects invalid choices and shows valid options

// Numeric parameters with defaults
parameters.NewParameterDefinition(
    "retry-count",
    parameters.ParameterTypeInteger,
    parameters.WithDefault(3),
)
```

This validation approach catches user errors early and provides helpful feedback, improving CLI usability and reducing support requirements.

## Creating and Working with Layers

### Method 1: Simple Layer Creation

Direct layer creation for straightforward parameter grouping:

```go
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",                    // Layer identifier
        "Database Configuration",      // Human-readable name
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "db-host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Database host to connect to"),
            ),
            parameters.NewParameterDefinition(
                "db-port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(5432),
                parameters.WithHelp("Database port (PostgreSQL default: 5432)"),
            ),
            parameters.NewParameterDefinition(
                "db-name",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database name (required for connection)"),
                parameters.WithRequired(true),
            ),
        ),
    )
}
```

### Method 2: Type-Safe Layer with Settings Struct

For complex layers requiring type safety and structured access:

```go
// 1. Define settings struct
type DatabaseSettings struct {
    Host     string `glazed.parameter:"db-host"`
    Port     int    `glazed.parameter:"db-port"`
    Name     string `glazed.parameter:"db-name"`
    Username string `glazed.parameter:"db-username"`
    Password string `glazed.parameter:"db-password"`
    SSLMode  string `glazed.parameter:"db-ssl-mode"`
}

// 2. Create layer with parameter definitions
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "db-host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Database host"),
            ),
            parameters.NewParameterDefinition(
                "db-port", 
                parameters.ParameterTypeInteger,
                parameters.WithDefault(5432),
                parameters.WithHelp("Database port"),
            ),
            parameters.NewParameterDefinition(
                "db-name",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database name"),
                parameters.WithRequired(true),
            ),
            parameters.NewParameterDefinition(
                "db-username",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database username"),
            ),
            parameters.NewParameterDefinition(
                "db-password",
                parameters.ParameterTypeSecret,  // Masked in output
                parameters.WithHelp("Database password"),
            ),
            parameters.NewParameterDefinition(
                "db-ssl-mode",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("disable", "require", "verify-ca", "verify-full"),
                parameters.WithDefault("require"),
                parameters.WithHelp("SSL mode for database connection"),
            ),
        ),
    )
}

// 3. Helper function for settings extraction
func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedLayers.InitializeStruct("database", settings)
    return settings, err
}

// 4. Usage in command implementation
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Extract database settings from the "database" layer
    dbSettings, err := GetDatabaseSettings(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to get database settings: %w", err)
    }
    
    // Connect to database using the settings
    dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=%s",
        dbSettings.Host, dbSettings.Port, dbSettings.Name, 
        dbSettings.Username, dbSettings.SSLMode)
    
    if dbSettings.Password != "" {
        dsn += fmt.Sprintf(" password=%s", dbSettings.Password)
    }
    
    // ... rest of command logic using database connection
    return nil
}
```

### Method 3: Layer Builder Pattern

For complex scenarios requiring conditional parameters:

```go
type DatabaseLayerBuilder struct {
    layer      layers.ParameterLayer
    includeSSL bool
    includePool bool
}

func NewDatabaseLayerBuilder() *DatabaseLayerBuilder {
    layer, _ := layers.NewParameterLayer("database", "Database Configuration")
    return &DatabaseLayerBuilder{layer: layer}
}

func (b *DatabaseLayerBuilder) WithSSL() *DatabaseLayerBuilder {
    b.includeSSL = true
    return b
}

func (b *DatabaseLayerBuilder) WithConnectionPool() *DatabaseLayerBuilder {
    b.includePool = true
    return b
}

func (b *DatabaseLayerBuilder) Build() (layers.ParameterLayer, error) {
    // Add basic parameters
    b.layer.AddFlags(
        parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString, 
            parameters.WithDefault("localhost")),
        parameters.NewParameterDefinition("db-port", parameters.ParameterTypeInteger,
            parameters.WithDefault(5432)),
    )
    
    // Conditionally add SSL parameters
    if b.includeSSL {
        b.layer.AddFlags(
            parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
                parameters.WithChoices("disable", "require", "verify-ca")),
            parameters.NewParameterDefinition("db-ssl-cert", parameters.ParameterTypeFile),
        )
    }
    
    // Conditionally add connection pool parameters
    if b.includePool {
        b.layer.AddFlags(
            parameters.NewParameterDefinition("db-max-connections", parameters.ParameterTypeInteger,
                parameters.WithDefault(10)),
            parameters.NewParameterDefinition("db-idle-timeout", parameters.ParameterTypeString,
                parameters.WithDefault("5m")),
        )
    }
    
    return b.layer, nil
}

// Usage:
dbLayer, _ := NewDatabaseLayerBuilder().
    WithSSL().
    WithConnectionPool().
    Build()
```

### Method 4: Registering Layers with Explicit Slugs on Commands

In some cases you may want to register layers on a command under explicit slugs that differ from the layer's internal slug. Use `cmds.WithLayersMap` to provide a map of slug-to-layer entries when creating a command description.

```go
dbLayer, _ := NewDatabaseLayer()        // internal slug: "database"
loggingLayer, _ := NewLoggingLayer()    // internal slug: "logging"

cmd := cmds.NewCommandDescription(
    "process",
    cmds.WithLayersMap(map[string]layers.ParameterLayer{
        "db":  dbLayer,    // registered under explicit slug "db"
        "log": loggingLayer,
    }),
)

// Later at runtime, access by the registration slugs (e.g., "db", "log")
```

Note:
- If a layer's internal slug differs from the map key and the layer is a `*layers.ParameterLayerImpl`, Glazed will clone the layer and align its slug to the provided key for consistent runtime behavior.
- For custom layer implementations, prefer using matching internal and registration slugs when possible.

## Practical Examples

### Example 1: Web Server Application

Complete layer implementation for a web server with database, logging, and server configuration:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Settings structs for type safety
type ServerSettings struct {
    Host         string `glazed.parameter:"host"`
    Port         int    `glazed.parameter:"port"`
    ReadTimeout  string `glazed.parameter:"read-timeout"`
    WriteTimeout string `glazed.parameter:"write-timeout"`
}

type LoggingSettings struct {
    Level  string `glazed.parameter:"log-level"`
    Format string `glazed.parameter:"log-format"`
    File   string `glazed.parameter:"log-file"`
}

type DatabaseSettings struct {
    Host     string `glazed.parameter:"db-host"`
    Port     int    `glazed.parameter:"db-port"`
    Name     string `glazed.parameter:"db-name"`
    Username string `glazed.parameter:"db-username"`
    Password string `glazed.parameter:"db-password"`
}

// Layer creation functions
func NewServerLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "server",
        "Web Server Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Server host to bind to"),
                parameters.WithShortFlag("H"),
            ),
            parameters.NewParameterDefinition(
                "port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(8080),
                parameters.WithHelp("Server port to listen on"),
                parameters.WithShortFlag("p"),
            ),
            parameters.NewParameterDefinition(
                "read-timeout",
                parameters.ParameterTypeString,
                parameters.WithDefault("30s"),
                parameters.WithHelp("HTTP read timeout"),
            ),
            parameters.NewParameterDefinition(
                "write-timeout",
                parameters.ParameterTypeString,
                parameters.WithDefault("30s"),
                parameters.WithHelp("HTTP write timeout"),
            ),
        ),
    )
}

func NewLoggingLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "logging",
        "Logging Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "log-level",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
                parameters.WithDefault("info"),
                parameters.WithHelp("Logging level"),
            ),
            parameters.NewParameterDefinition(
                "log-format",
                parameters.ParameterTypeChoice,
                parameters.WithChoices("text", "json"),
                parameters.WithDefault("text"),
                parameters.WithHelp("Log output format"),
            ),
            parameters.NewParameterDefinition(
                "log-file",
                parameters.ParameterTypeString,
                parameters.WithHelp("Log file path (default: stderr)"),
            ),
        ),
    )
}

func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "db-host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Database host"),
            ),
            parameters.NewParameterDefinition(
                "db-port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(5432),
                parameters.WithHelp("Database port"),
            ),
            parameters.NewParameterDefinition(
                "db-name",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database name"),
                parameters.WithRequired(true),
            ),
            parameters.NewParameterDefinition(
                "db-username",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database username"),
            ),
            parameters.NewParameterDefinition(
                "db-password",
                parameters.ParameterTypeSecret,
                parameters.WithHelp("Database password"),
            ),
        ),
    )
}

// Command creation with layer composition
func NewServerCommand() (*cmds.CommandDescription, error) {
    // Create layers
    serverLayer, err := NewServerLayer()
    if err != nil {
        return nil, err
    }
    
    loggingLayer, err := NewLoggingLayer()
    if err != nil {
        return nil, err
    }
    
    databaseLayer, err := NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Compose command with relevant layers
    return cmds.NewCommandDescription(
        "serve",
        cmds.WithShort("Start the web server"),
        cmds.WithLong("Start the web server with the specified configuration"),
        cmds.WithLayersList(serverLayer, databaseLayer, loggingLayer),
    ), nil
}

func NewHealthCheckCommand() (*cmds.CommandDescription, error) {
    // Health check only needs server configuration, not database
    serverLayer, err := NewServerLayer()
    if err != nil {
        return nil, err
    }
    
    loggingLayer, err := NewLoggingLayer()
    if err != nil {
        return nil, err
    }
    
    return cmds.NewCommandDescription(
        "health",
        cmds.WithShort("Check server health"),
        cmds.WithFlags(
            parameters.NewParameterDefinition(
                "endpoint",
                parameters.ParameterTypeString,
                parameters.WithDefault("/health"),
                parameters.WithHelp("Health check endpoint"),
            ),
        ),
        cmds.WithLayersList(serverLayer, loggingLayer), // No database layer
    ), nil
}

// Settings extraction helpers demonstrate how to use InitializeStruct with layer-specific settings
func GetServerSettings(parsedLayers *layers.ParsedLayers) (*ServerSettings, error) {
    settings := &ServerSettings{}
    err := parsedLayers.InitializeStruct("server", settings)
    return settings, err
}

func GetLoggingSettings(parsedLayers *layers.ParsedLayers) (*LoggingSettings, error) {
    settings := &LoggingSettings{}
    err := parsedLayers.InitializeStruct("logging", settings)
    return settings, err
}

func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedLayers.InitializeStruct("database", settings)
    return settings, err
}

// Example command implementation using multiple layer settings
type ServerCommand struct {
    *cmds.CommandDescription
}

func (c *ServerCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Extract settings from each layer
    serverSettings, err := GetServerSettings(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to get server settings: %w", err)
    }
    
    dbSettings, err := GetDatabaseSettings(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to get database settings: %w", err)
    }
    
    logSettings, err := GetLoggingSettings(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to get logging settings: %w", err)
    }
    
    // Use settings from multiple layers
    fmt.Printf("Starting server on %s:%d\n", serverSettings.Host, serverSettings.Port)
    fmt.Printf("Database: %s:%d/%s\n", dbSettings.Host, dbSettings.Port, dbSettings.Name)
    fmt.Printf("Log level: %s\n", logSettings.Level)
    
    // Parse timeout values
    readTimeout, err := time.ParseDuration(serverSettings.ReadTimeout)
    if err != nil {
        return fmt.Errorf("invalid read timeout: %w", err)
    }
    
    writeTimeout, err := time.ParseDuration(serverSettings.WriteTimeout)
    if err != nil {
        return fmt.Errorf("invalid write timeout: %w", err)
    }
    
    // ... start server with these settings
    _ = readTimeout
    _ = writeTimeout
    
    return nil
}
```

### Example 2: CLI Tool with Optional Features

Layer composition for applications with conditional functionality. This example shows how to extract settings from optional layers and use them together:

```go
// Feature layers for optional inclusion
func NewCacheLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "cache",
        "Caching Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "cache-enabled",
                parameters.ParameterTypeBool,
                parameters.WithDefault(true),
                parameters.WithHelp("Enable caching"),
            ),
            parameters.NewParameterDefinition(
                "cache-ttl",
                parameters.ParameterTypeString,
                parameters.WithDefault("1h"),
                parameters.WithHelp("Cache time-to-live"),
            ),
            parameters.NewParameterDefinition(
                "cache-size",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(1000),
                parameters.WithHelp("Maximum cache entries"),
            ),
        ),
    )
}

func NewMetricsLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "metrics",
        "Metrics and Monitoring",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "metrics-enabled",
                parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Enable metrics collection"),
            ),
            parameters.NewParameterDefinition(
                "metrics-port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(9090),
                parameters.WithHelp("Metrics server port"),
            ),
            parameters.NewParameterDefinition(
                "metrics-path",
                parameters.ParameterTypeString,
                parameters.WithDefault("/metrics"),
                parameters.WithHelp("Metrics endpoint path"),
            ),
        ),
    )
}

// Command builder with optional features
type AppCommandBuilder struct {
    baseLayers    []layers.ParameterLayer
    enableCache   bool
    enableMetrics bool
    enableAuth    bool
}

func NewAppCommandBuilder() *AppCommandBuilder {
    loggingLayer, _ := NewLoggingLayer()
    
    return &AppCommandBuilder{
        baseLayers: []layers.ParameterLayer{loggingLayer},
    }
}

func (b *AppCommandBuilder) WithCache() *AppCommandBuilder {
    b.enableCache = true
    return b
}

func (b *AppCommandBuilder) WithMetrics() *AppCommandBuilder {
    b.enableMetrics = true
    return b
}

func (b *AppCommandBuilder) WithAuth() *AppCommandBuilder {
    b.enableAuth = true
    return b
}

func (b *AppCommandBuilder) BuildProcessCommand() (*cmds.CommandDescription, error) {
    commandLayers := append([]layers.ParameterLayer{}, b.baseLayers...)
    
    // Add optional layers based on enabled features
    if b.enableCache {
        cacheLayer, err := NewCacheLayer()
        if err != nil {
            return nil, err
        }
        commandLayers = append(commandLayers, cacheLayer)
    }
    
    if b.enableMetrics {
        metricsLayer, err := NewMetricsLayer()
        if err != nil {
            return nil, err
        }
        commandLayers = append(commandLayers, metricsLayer)
    }
    
    if b.enableAuth {
        authLayer, err := NewAuthLayer() // Assume this exists
        if err != nil {
            return nil, err
        }
        commandLayers = append(commandLayers, authLayer)
    }
    
    return cmds.NewCommandDescription(
        "process",
        cmds.WithShort("Process data with optional features"),
        cmds.WithFlags(
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
            ),
        ),
        cmds.WithLayersList(commandLayers...),
    ), nil
}

// Settings structs for optional features
type CacheSettings struct {
    Enabled bool   `glazed.parameter:"cache-enabled"`
    TTL     string `glazed.parameter:"cache-ttl"`
    Size    int    `glazed.parameter:"cache-size"`
}

type MetricsSettings struct {
    Enabled bool   `glazed.parameter:"metrics-enabled"`
    Port    int    `glazed.parameter:"metrics-port"`
    Path    string `glazed.parameter:"metrics-path"`
}

// Helper functions for optional layer settings
func GetCacheSettings(parsedLayers *layers.ParsedLayers) (*CacheSettings, error) {
    settings := &CacheSettings{}
    err := parsedLayers.InitializeStruct("cache", settings)
    return settings, err
}

func GetMetricsSettings(parsedLayers *layers.ParsedLayers) (*MetricsSettings, error) {
    settings := &MetricsSettings{}
    err := parsedLayers.InitializeStruct("metrics", settings)
    return settings, err
}

// Command implementation that handles optional layers
type ProcessCommand struct {
    *cmds.CommandDescription
}

func (c *ProcessCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Always extract logging settings
    logSettings, err := GetLoggingSettings(parsedLayers)
    if err != nil {
        return fmt.Errorf("failed to get logging settings: %w", err)
    }
    
    fmt.Printf("Starting processing with log level: %s\n", logSettings.Level)
    
    // Try to extract cache settings (may not exist)
    if parsedLayers.Has("cache") {
        cacheSettings, err := GetCacheSettings(parsedLayers)
        if err != nil {
            return fmt.Errorf("failed to get cache settings: %w", err)
        }
        
        if cacheSettings.Enabled {
            fmt.Printf("Cache enabled: TTL=%s, Size=%d\n", 
                cacheSettings.TTL, cacheSettings.Size)
            // Initialize cache with these settings
        }
    }
    
    // Try to extract metrics settings (may not exist)
    if parsedLayers.Has("metrics") {
        metricsSettings, err := GetMetricsSettings(parsedLayers)
        if err != nil {
            return fmt.Errorf("failed to get metrics settings: %w", err)
        }
        
        if metricsSettings.Enabled {
            fmt.Printf("Metrics enabled on port %d at %s\n", 
                metricsSettings.Port, metricsSettings.Path)
            // Start metrics server
        }
    }
    
    // ... rest of processing logic
    return nil
}

// Usage:
func CreateCommands() {
    // Basic command (only logging)
    basicCmd, _ := NewAppCommandBuilder().
        BuildProcessCommand()
        
    // Feature-rich command (cache + metrics + auth)
    advancedCmd, _ := NewAppCommandBuilder().
        WithCache().
        WithMetrics().
        WithAuth().
        BuildProcessCommand()
}
```

## Advanced Patterns

### 1. Layer Inheritance and Composition

Extending existing layers without modification:

```go
// Base database layer
func NewBaseDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString,
                parameters.WithDefault("localhost")),
            parameters.NewParameterDefinition("db-port", parameters.ParameterTypeInteger,
                parameters.WithDefault(5432)),
        ),
    )
}

// Extended database layer with additional features
func NewAdvancedDatabaseLayer() (layers.ParameterLayer, error) {
    // Start with base layer
    baseLayer, err := NewBaseDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Clone to avoid modifying the original
    advancedLayer := baseLayer.Clone()
    
    // Add additional parameters
    advancedLayer.AddFlags(
        parameters.NewParameterDefinition("db-pool-size", parameters.ParameterTypeInteger,
            parameters.WithDefault(10)),
        parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
            parameters.WithChoices("disable", "require", "verify-full")),
        parameters.NewParameterDefinition("db-connection-timeout", parameters.ParameterTypeDuration,
            parameters.WithDefault("30s")),
    )
    
    return advancedLayer, nil
}
```

### 2. Environment-Specific Layer Configuration

Adapting layers for different deployment environments:

```go
type EnvironmentConfig struct {
    Environment string // "development", "staging", "production"
    Features    []string
}

func NewEnvironmentAwareDatabaseLayer(config EnvironmentConfig) (layers.ParameterLayer, error) {
    layer, err := NewBaseDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Add environment-specific parameters
    switch config.Environment {
    case "development":
        layer.AddFlags(
            parameters.NewParameterDefinition("db-debug-queries", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
            parameters.NewParameterDefinition("db-auto-migrate", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
        )
    case "production":
        layer.AddFlags(
            parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
                parameters.WithChoices("require", "verify-full"),
                parameters.WithDefault("verify-full")),
            parameters.NewParameterDefinition("db-connection-pool-size", parameters.ParameterTypeInteger,
                parameters.WithDefault(50)),
        )
    }
    
    // Add feature-specific parameters
    for _, feature := range config.Features {
        switch feature {
        case "monitoring":
            layer.AddFlags(
                parameters.NewParameterDefinition("db-monitor-slow-queries", parameters.ParameterTypeBool),
                parameters.NewParameterDefinition("db-slow-query-threshold", parameters.ParameterTypeDuration,
                    parameters.WithDefault("1s")),
            )
        case "backup":
            layer.AddFlags(
                parameters.NewParameterDefinition("db-backup-enabled", parameters.ParameterTypeBool),
                parameters.NewParameterDefinition("db-backup-schedule", parameters.ParameterTypeString),
            )
        }
    }
    
    return layer, nil
}
```

### 3. Dynamic Layer Registration

Runtime layer registration for plugin systems:

```go
type LayerRegistry struct {
    layers map[string]layers.ParameterLayer
    mutex  sync.RWMutex
}

func NewLayerRegistry() *LayerRegistry {
    return &LayerRegistry{
        layers: make(map[string]layers.ParameterLayer),
    }
}

func (r *LayerRegistry) RegisterLayer(slug string, layer layers.ParameterLayer) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if _, exists := r.layers[slug]; exists {
        return fmt.Errorf("layer %s already registered", slug)
    }
    
    r.layers[slug] = layer
    return nil
}

func (r *LayerRegistry) GetLayer(slug string) (layers.ParameterLayer, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    layer, exists := r.layers[slug]
    if !exists {
        return nil, fmt.Errorf("layer %s not found", slug)
    }
    
    return layer, nil
}

func (r *LayerRegistry) BuildCommand(name string, layerSlugs []string) (*cmds.CommandDescription, error) {
    var commandLayers []layers.ParameterLayer
    
    for _, slug := range layerSlugs {
        layer, err := r.GetLayer(slug)
        if err != nil {
            return nil, err
        }
        commandLayers = append(commandLayers, layer)
    }
    
    return cmds.NewCommandDescription(name,
        cmds.WithLayersList(commandLayers...)), nil
}

// Plugin system usage
func InitializePluginSystem() *LayerRegistry {
    registry := NewLayerRegistry()
    
    // Register core layers
    databaseLayer, _ := NewDatabaseLayer()
    registry.RegisterLayer("database", databaseLayer)
    
    loggingLayer, _ := NewLoggingLayer()
    registry.RegisterLayer("logging", loggingLayer)
    
    // Plugins can register additional layers
    // registry.RegisterLayer("custom-feature", customLayer)
    
    return registry
}
```

### 4. Layer Validation and Constraints

Complex validation rules across layer parameters:

```go
type LayerValidator struct {
    rules []ValidationRule
}

type ValidationRule func(*layers.ParsedLayers) error

func NewLayerValidator() *LayerValidator {
    return &LayerValidator{}
}

func (v *LayerValidator) AddRule(rule ValidationRule) {
    v.rules = append(v.rules, rule)
}

func (v *LayerValidator) Validate(parsedLayers *layers.ParsedLayers) error {
    for _, rule := range v.rules {
        if err := rule(parsedLayers); err != nil {
            return err
        }
    }
    return nil
}

// Cross-layer validation rules
func DatabaseConnectionRule(parsedLayers *layers.ParsedLayers) error {
    dbLayer, ok := parsedLayers.Get("database")
    if !ok {
        return nil // Skip if database layer not present
    }
    
    host, ok := dbLayer.GetParameter("db-host")
    if !ok {
        return nil
    }
    
    port, ok := dbLayer.GetParameter("db-port")
    if !ok {
        return nil
    }
    
    // Validate connection parameters make sense together
    if host == "localhost" && port.(int) < 1024 {
        return fmt.Errorf("localhost connections should use ports >= 1024")
    }
    
    return nil
}

func SSLConfigurationRule(parsedLayers *layers.ParsedLayers) error {
    dbLayer, ok := parsedLayers.Get("database")
    if !ok {
        return nil // Skip if database layer not present
    }
    
    sslMode, ok := dbLayer.GetParameter("db-ssl-mode")
    if !ok {
        return nil // Skip if SSL not configured
    }
    
    if sslMode == "verify-full" {
        // Ensure SSL certificate is provided when required
        cert, ok := dbLayer.GetParameter("db-ssl-cert")
        if !ok || cert == "" {
            return fmt.Errorf("SSL certificate required when ssl-mode is verify-full")
        }
    }
    
    return nil
}

// Usage in command implementation
func (c *MyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Validate layer configuration
    validator := NewLayerValidator()
    validator.AddRule(DatabaseConnectionRule)
    validator.AddRule(SSLConfigurationRule)
    
    if err := validator.Validate(parsedLayers); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }
    
    // Continue with command execution
    return nil
}
```

## Best Practices

### Layer Design Principles

**Single Responsibility**: Each layer should handle one logical area of configuration. Database layers handle database parameters, logging layers handle logging configuration.

**Clear Naming**: Use descriptive layer slugs and parameter names. Prefer `database-connection-timeout` over `timeout`.

**Sensible Defaults**: Provide reasonable default values that work in common scenarios. Users should be able to run commands without extensive configuration.

**Consistent Interfaces**: Use similar parameter names across layers. If one layer uses `host`, avoid `hostname` in another layer for the same concept.

### Parameter Organization

Group related parameters logically within layers. Database layers should include host, port, credentials, and connection options together.

Use consistent naming patterns across your application. Establish conventions for common concepts like timeouts, ports, and file paths.

Consider parameter relationships when designing layers. Parameters that are frequently used together belong in the same layer.

### Command Composition

Only include layers that provide parameters relevant to the command's functionality. Avoid layer pollution by being selective.

Use builder patterns for commands with many optional features. This provides flexibility while maintaining clean interfaces.

Consider creating specialized layer variants for different command types. A read-only database layer might exclude authentication parameters.

### Error Handling and Validation

Validate layer configuration early in command execution. Fail fast with clear error messages about parameter issues.

Provide helpful validation messages that guide users toward correct configuration. Include examples of valid values when rejecting input.

Use type-safe parameter extraction where possible. Struct-based settings reduce runtime errors and improve code clarity.

### Testing and Maintenance

Write unit tests for layer definitions to ensure parameter validation works correctly. Test edge cases and error conditions.

Test layer composition to verify that combined layers work correctly together. Check for parameter conflicts and validation interactions.

Use integration tests to verify that commands work correctly with different layer combinations and parameter sources.

Document layer dependencies and relationships. Explain when layers should be used together and any constraints.

Keep layer definitions close to their usage when possible. This improves maintainability and reduces the chance of configuration drift.

Version layer definitions carefully in evolving applications. Consider backward compatibility when modifying existing layers.

## Testing Layers

### Unit Testing Layer Definitions

Test individual layer creation and parameter validation:

```go
func TestDatabaseLayer(t *testing.T) {
    layer, err := NewDatabaseLayer()
    assert.NoError(t, err)
    assert.Equal(t, "database", layer.GetSlug())
    
    // Test parameter definitions
    params := layer.GetParameterDefinitions()
    assert.Contains(t, params, "db-host")
    assert.Contains(t, params, "db-port")
    assert.Contains(t, params, "db-name")
    
    // Test default values
    hostParam := params["db-host"]
    assert.Equal(t, "localhost", hostParam.Default)
    
    portParam := params["db-port"]
    assert.Equal(t, 5432, portParam.Default)
}

func TestParameterValidation(t *testing.T) {
    layer, _ := NewDatabaseLayer()
    
    // Test valid choices
    logLevelParam := layer.GetParameterDefinitions()["log-level"]
    validChoices := []string{"debug", "info", "warn", "error", "fatal"}
    assert.Equal(t, validChoices, logLevelParam.Choices)
    
    // Test required parameters
    dbNameParam := layer.GetParameterDefinitions()["db-name"]
    assert.True(t, dbNameParam.Required)
}
```

### Integration Testing Layer Composition

Test command creation with multiple layers:

```go
func TestCommandComposition(t *testing.T) {
    serverLayer, _ := NewServerLayer()
    databaseLayer, _ := NewDatabaseLayer()
    loggingLayer, _ := NewLoggingLayer()
    
    command, err := cmds.NewCommandDescription("test-command",
        cmds.WithLayersList(serverLayer, databaseLayer, loggingLayer))
    
    assert.NoError(t, err)
    assert.NotNil(t, command)
    
    // Verify all layers are present
    layers := command.GetLayers()
    assert.Len(t, layers, 3)
    
    // Verify no parameter conflicts
    allParams := make(map[string]bool)
    for _, layer := range layers {
        for paramName := range layer.GetParameterDefinitions() {
            assert.False(t, allParams[paramName], 
                "Parameter %s defined in multiple layers", paramName)
            allParams[paramName] = true
        }
    }
}
```

### Testing Parameter Resolution

Test parameter value resolution from different sources:

```go
func TestParameterResolution(t *testing.T) {
    // Create test command with layers
    command, _ := createTestCommand()
    
    // Test CLI argument parsing
    args := []string{"--db-host", "testhost", "--db-port", "3306", "--log-level", "debug"}
    parsedLayers, err := command.ParseLayers(args)
    assert.NoError(t, err)
    
    // Verify parsed values
    dbLayer, ok := parsedLayers.Get("database")
    assert.True(t, ok)
    
    dbHost, ok := dbLayer.GetParameter("db-host")
    assert.True(t, ok)
    assert.Equal(t, "testhost", dbHost)
    
    dbPort, ok := dbLayer.GetParameter("db-port")
    assert.True(t, ok)
    assert.Equal(t, 3306, dbPort)
    
    // Test struct initialization
    dbSettings := &DatabaseSettings{}
    err = parsedLayers.InitializeStruct("database", dbSettings)
    assert.NoError(t, err)
    assert.Equal(t, "testhost", dbSettings.Host)
    assert.Equal(t, 3306, dbSettings.Port)
}

func TestDefaultValues(t *testing.T) {
    command, _ := createTestCommand()
    
    // Parse with no arguments - should use defaults
    parsedLayers, err := command.ParseLayers([]string{})
    assert.NoError(t, err)
    
    // Verify default values are used
    dbLayer, _ := parsedLayers.Get("database")
    dbHost, _ := dbLayer.GetParameter("db-host")
    assert.Equal(t, "localhost", dbHost)
    
    dbPort, _ := dbLayer.GetParameter("db-port")
    assert.Equal(t, 5432, dbPort)
    
    loggingLayer, _ := parsedLayers.Get("logging")
    logLevel, _ := loggingLayer.GetParameter("log-level")
    assert.Equal(t, "info", logLevel)
}
```

### Testing Layer Builders and Dynamic Composition

Test builder patterns and conditional layer inclusion:

```go
func TestDatabaseLayerBuilder(t *testing.T) {
    // Test basic layer
    basicLayer, err := NewDatabaseLayerBuilder().Build()
    assert.NoError(t, err)
    
    basicParams := basicLayer.GetParameterDefinitions()
    assert.Contains(t, basicParams, "db-host")
    assert.Contains(t, basicParams, "db-port")
    assert.NotContains(t, basicParams, "db-ssl-mode")
    
    // Test layer with SSL
    sslLayer, err := NewDatabaseLayerBuilder().WithSSL().Build()
    assert.NoError(t, err)
    
    sslParams := sslLayer.GetParameterDefinitions()
    assert.Contains(t, sslParams, "db-host")
    assert.Contains(t, sslParams, "db-ssl-mode")
    assert.Contains(t, sslParams, "db-ssl-cert")
    
    // Test layer with connection pool
    poolLayer, err := NewDatabaseLayerBuilder().WithConnectionPool().Build()
    assert.NoError(t, err)
    
    poolParams := poolLayer.GetParameterDefinitions()
    assert.Contains(t, poolParams, "db-max-connections")
    assert.Contains(t, poolParams, "db-idle-timeout")
}

func TestConditionalLayerComposition(t *testing.T) {
    builder := NewAppCommandBuilder()
    
    // Test basic command
    basicCmd, err := builder.BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, basicCmd.GetLayers(), 1) // Only logging layer
    
    // Test command with cache
    cacheCmd, err := builder.WithCache().BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, cacheCmd.GetLayers(), 2) // Logging + cache layers
    
    // Test command with all features
    fullCmd, err := builder.WithCache().WithMetrics().WithAuth().BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, fullCmd.GetLayers(), 4) // All layers
}
```

This comprehensive testing approach ensures layers work correctly individually and in composition, parameter resolution functions properly across different sources, and dynamic layer construction produces expected results.
