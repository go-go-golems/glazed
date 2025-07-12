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

## Table of Contents
1. [The Big Picture](#the-big-picture)
2. [Mental Model & Why Use Layers](#mental-model--why-use-layers)
3. [Understanding Layers](#understanding-layers)
4. [Core Layer Concepts](#core-layer-concepts)
5. [Layer Types & Components](#layer-types--components)
6. [Creating and Working with Layers](#creating-and-working-with-layers)
7. [Practical Examples](#practical-examples)
8. [Advanced Patterns](#advanced-patterns)
9. [Best Practices](#best-practices)
10. [Testing Layers](#testing-layers)

## The Big Picture

Imagine you're building a CLI application with multiple subcommands, each requiring different sets of parameters:

```bash
myapp server --host localhost --port 8080 --db-host postgres --db-port 5432 --log-level info
myapp client --endpoint http://localhost:8080 --timeout 30s --log-level debug  
myapp backup --db-host postgres --db-port 5432 --output backup.sql --log-level warn
```

Traditional approaches often lead to parameter chaos:
- **Parameter pollution**: Every command gets every possible flag, even if irrelevant
- **Naming conflicts**: `--host` could mean web server host OR database host
- **Code duplication**: Same parameter definitions repeated across commands
- **Poor organization**: No logical grouping of related parameters
- **Maintenance nightmare**: Adding a new database parameter means updating every command

```go
// The traditional approach - scattered and repetitive
type ServerCommand struct {
    Host     string // Web server host
    Port     int    // Web server port
    DbHost   string // Database host
    DbPort   int    // Database port
    LogLevel string // Logging level
    // ... dozens more fields
}

type ClientCommand struct {
    Endpoint string // API endpoint
    Timeout  string // Request timeout
    LogLevel string // Logging level (duplicated!)
    // ... more fields
}

type BackupCommand struct {
    DbHost   string // Database host (duplicated again!)
    DbPort   int    // Database port (duplicated again!)
    Output   string // Output file
    LogLevel string // Logging level (duplicated yet again!)
    // ... more fields
}
```

**Glazed's layer system solves these problems** by organizing parameters into logical groups that can be:
- **Shared** across commands that need similar functionality
- **Composed** together to build complex command interfaces
- **Reused** without code duplication
- **Maintained** in one place

## Mental Model & Why Use Layers

Think of layers like **ingredient lists** for different types of dishes:

```
🥗 Salad Recipe:
├── Base Ingredients (lettuce, tomatoes, cucumbers)
├── Dressing Layer (oil, vinegar, seasoning)
└── Protein Layer (chicken, tofu, or cheese)

🍝 Pasta Recipe:
├── Base Ingredients (pasta, sauce, herbs)
├── Protein Layer (chicken, tofu, or cheese) ← REUSED!
└── Garnish Layer (parmesan, basil)

🥪 Sandwich Recipe:
├── Base Ingredients (bread, condiments)
├── Protein Layer (chicken, tofu, or cheese) ← REUSED AGAIN!
└── Vegetable Layer (lettuce, tomatoes) ← PARTIALLY REUSED!
```

In the same way, Glazed layers let you:

1. **Define reusable parameter groups** (like "protein layer" or "dressing layer")
2. **Compose commands** by mixing and matching the layers they need
3. **Avoid repetition** by reusing common layers across commands
4. **Maintain consistency** by keeping related parameters together

### The Glazed Way

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

**Result**: Clean, composable, maintainable parameter management.

## Understanding Layers

### What is a Layer?

A **layer** is a named collection of related parameter definitions. Each layer has:

- **A unique slug/identifier** (e.g., "database", "logging", "glazed")
- **A human-readable name** for documentation and help text
- **A collection of parameter definitions** (flags, arguments, options)
- **Optional metadata** (descriptions, validation rules, grouping information)

### Layer Architecture

```
┌─────────────────────────────────────────────┐
│                Command                       │
│  ┌─────────────────────────────────────────┤
│  │            Command Description           │
│  │                                         │
│  │  ┌─────────────────────────────────────┤
│  │  │          Parameter Layers           │
│  │  │                                     │
│  │  │  ┌─────────────┬─────────────────┤   │
│  │  │  │Default Layer│  Custom Layers  │   │
│  │  │  │(flags/args) │  (database,    │   │
│  │  │  │             │   logging, etc.)│   │
│  │  │  └─────────────┴─────────────────┘   │
│  │  └─────────────────────────────────────┘ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────┐
│              Runtime Execution               │
│                                             │
│  Parameter Sources (in priority order):     │
│  1. Command line arguments                  │
│  2. Environment variables                   │
│  3. Configuration files                     │
│  4. Default values from layer definitions   │
└─────────────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────┐
│               Parsed Layers                  │
│        (runtime parameter values)           │
└─────────────────────────────────────────────┘
```

## Core Layer Concepts

### 1. Layer Hierarchy and Organization

Layers are organized in a hierarchical fashion:

```
CommandDescription
├── Default Layer (always present)
│   ├── Command-specific flags
│   └── Command-specific arguments
├── Standard Layers (optional, commonly used)
│   ├── Glazed Layer (for structured output)
│   ├── Logging Layer (for logging configuration)
│   └── Database Layer (for database connections)
└── Custom Layers (application-specific)
    ├── Authentication Layer
    ├── API Configuration Layer
    └── Feature Toggle Layer
```

### 2. Parameter Definition vs. Parsed Values

Understanding the distinction between definitions and values is crucial:

**Parameter Definitions** (stored in layers):
```go
parameters.NewParameterDefinition(
    "log-level",                    // Name
    parameters.ParameterTypeChoice, // Type
    parameters.WithDefault("info"), // Default value
    parameters.WithChoices("debug", "info", "warn", "error"),
    parameters.WithHelp("Set the logging level"),
)
```

**Parsed Values** (runtime values from various sources):
```go
// After parsing command line: --log-level debug
parsedValue := "debug"

// Accessed through ParsedLayers:
logLevel, _ := parsedLayers.GetParameterValue("logging", "log-level")
```

### 3. Layer Composition Patterns

#### Basic Composition
```go
// Simple addition of layers
commandDesc := cmds.NewCommandDescription("mycommand",
    cmds.WithLayers(databaseLayer, loggingLayer))
```

#### Conditional Composition
```go
// Add layers based on conditions
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

#### Layer Inheritance and Extension
```go
// Extend an existing layer with additional parameters
extendedDbLayer := databaseLayer.Clone()
extendedDbLayer.AddFlags(
    parameters.NewParameterDefinition("pool-size", parameters.ParameterTypeInteger),
)
```

## Layer Types & Components

### 1. Built-in Layer Types

#### Default Layer
- **Purpose**: Contains command-specific flags and arguments
- **Slug**: `"default"` (constant: `layers.DefaultSlug`)
- **Creation**: Automatically created when using `cmds.WithFlags()` or `cmds.WithArguments()`
- **Use Case**: Parameters unique to a specific command

```go
// Default layer is created automatically
commandDesc := cmds.NewCommandDescription("serve",
    cmds.WithFlags(
        parameters.NewParameterDefinition("port", parameters.ParameterTypeInteger),
        parameters.NewParameterDefinition("host", parameters.ParameterTypeString),
    ),
    cmds.WithArguments(
        parameters.NewParameterDefinition("config-file", parameters.ParameterTypeString),
    ),
)
```

#### Glazed Layer
- **Purpose**: Provides structured output formatting options
- **Slug**: `"glazed"` (constant: `settings.GlazedSlug`)
- **Creation**: `settings.NewGlazedParameterLayers()`
- **Use Case**: Commands that output structured data (tables, JSON, YAML, etc.)

```go
glazedLayer, _ := settings.NewGlazedParameterLayers()
// Provides: --output, --fields, --sort-columns, --filter, etc.
```

### 2. Custom Layer Components

#### Parameter Definitions
Each layer contains parameter definitions that specify:

```go
paramDef := parameters.NewParameterDefinition(
    "connection-timeout",              // Parameter name
    parameters.ParameterTypeDuration,  // Type (string, int, bool, duration, etc.)
    parameters.WithDefault("30s"),     // Default value
    parameters.WithHelp("Connection timeout for database operations"),
    parameters.WithRequired(false),    // Whether parameter is required
    parameters.WithShortFlag("t"),     // Short flag (-t)
)
```

#### Validation and Constraints
```go
// Choice parameters with validation
parameters.NewParameterDefinition(
    "log-level",
    parameters.ParameterTypeChoice,
    parameters.WithChoices("debug", "info", "warn", "error", "fatal"),
    parameters.WithDefault("info"),
)

// Numeric parameters with ranges
parameters.NewParameterDefinition(
    "retry-count",
    parameters.ParameterTypeInteger,
    parameters.WithDefault(3),
    // Note: Range validation would be added through custom validation
)
```

## Creating and Working with Layers

### Method 1: Simple Layer Creation

For basic use cases, create layers directly:

```go
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",                    // Layer slug
        "Database Configuration",      // Human-readable name
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
        ),
    )
}
```

### Method 2: Advanced Layer with Settings Struct

For complex layers, use a settings struct for type safety:

```go
// 1. Define the settings struct
type DatabaseSettings struct {
    Host     string `glazed.parameter:"db-host"`
    Port     int    `glazed.parameter:"db-port"`
    Name     string `glazed.parameter:"db-name"`
    Username string `glazed.parameter:"db-username"`
    Password string `glazed.parameter:"db-password"`
    SSLMode  string `glazed.parameter:"db-ssl-mode"`
}

// 2. Create the layer with all parameter definitions
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

// 3. Helper function to extract settings from parsed layers
func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedLayers.InitializeStruct("database", settings)
    return settings, err
}
```

### Method 3: Layer Builder Pattern

For complex scenarios with conditional parameters:

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
    // Add basic database parameters
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
            parameters.NewParameterDefinition("db-idle-timeout", parameters.ParameterTypeDuration,
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

## Practical Examples

### Example 1: Web Server Application

Let's build layers for a web server application with database, logging, and server configuration:

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
                parameters.ParameterTypeDuration,
                parameters.WithDefault("30s"),
                parameters.WithHelp("HTTP read timeout"),
            ),
            parameters.NewParameterDefinition(
                "write-timeout",
                parameters.ParameterTypeDuration,
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

// Settings extraction helpers
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
```

### Example 2: CLI Tool with Optional Features

Here's how to create layers for a CLI tool where features are optional:

```go
// Feature layers that can be optionally included
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
                parameters.ParameterTypeDuration,
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
    // Create base layers that are always included
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

// Usage example:
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

Sometimes you need to extend existing layers or create hierarchical layer relationships:

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
    
    // Add advanced parameters
    advancedLayer.AddFlags(
        parameters.NewParameterDefinition("db-ssl-mode", parameters.ParameterTypeChoice,
            parameters.WithChoices("disable", "require", "verify-ca", "verify-full"),
            parameters.WithDefault("require")),
        parameters.NewParameterDefinition("db-max-connections", parameters.ParameterTypeInteger,
            parameters.WithDefault(25)),
        parameters.NewParameterDefinition("db-connection-timeout", parameters.ParameterTypeDuration,
            parameters.WithDefault("30s")),
    )
    
    return advancedLayer, nil
}
```

### 2. Dynamic Layer Configuration

Create layers that adapt based on runtime configuration or environment:

```go
type LayerConfig struct {
    Environment    string
    EnableDebug    bool
    EnableSecurity bool
}

func NewConfigurableLayer(config LayerConfig) (layers.ParameterLayer, error) {
    layer, err := layers.NewParameterLayer("configurable", "Configurable Settings")
    if err != nil {
        return nil, err
    }
    
    // Always include basic parameters
    layer.AddFlags(
        parameters.NewParameterDefinition("app-name", parameters.ParameterTypeString,
            parameters.WithDefault("myapp")),
    )
    
    // Environment-specific parameters
    switch config.Environment {
    case "development":
        layer.AddFlags(
            parameters.NewParameterDefinition("dev-mode", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
            parameters.NewParameterDefinition("hot-reload", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
        )
    case "production":
        layer.AddFlags(
            parameters.NewParameterDefinition("performance-mode", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
            parameters.NewParameterDefinition("compression", parameters.ParameterTypeBool,
                parameters.WithDefault(true)),
        )
    }
    
    // Debug parameters (only if debug is enabled)
    if config.EnableDebug {
        layer.AddFlags(
            parameters.NewParameterDefinition("debug-level", parameters.ParameterTypeInteger,
                parameters.WithDefault(1)),
            parameters.NewParameterDefinition("profile", parameters.ParameterTypeBool,
                parameters.WithDefault(false)),
        )
    }
    
    // Security parameters (only if security is enabled)
    if config.EnableSecurity {
        layer.AddFlags(
            parameters.NewParameterDefinition("api-key", parameters.ParameterTypeSecret,
                parameters.WithHelp("API key for authentication")),
            parameters.NewParameterDefinition("token-expiry", parameters.ParameterTypeDuration,
                parameters.WithDefault("24h")),
        )
    }
    
    return layer, nil
}
```

### 3. Layer Validation and Dependencies

Implement validation logic and layer dependencies:

```go
// Layer with validation
func NewValidatedDatabaseLayer() (layers.ParameterLayer, error) {
    layer, err := layers.NewParameterLayer("database", "Database Configuration")
    if err != nil {
        return nil, err
    }
    
    layer.AddFlags(
        parameters.NewParameterDefinition("db-host", parameters.ParameterTypeString,
            parameters.WithDefault("localhost")),
        parameters.NewParameterDefinition("db-port", parameters.ParameterTypeInteger,
            parameters.WithDefault(5432)),
        parameters.NewParameterDefinition("db-name", parameters.ParameterTypeString,
            parameters.WithRequired(true)),
    )
    
    return layer, nil
}

// Validation function for layer consistency
func ValidateDatabaseSettings(parsedLayers *layers.ParsedLayers) error {
    dbHost, _ := parsedLayers.GetParameterValue("database", "db-host")
    dbName, _ := parsedLayers.GetParameterValue("database", "db-name")
    
    // Custom validation logic
    if dbHost == "localhost" && dbName == "" {
        return fmt.Errorf("database name is required when using localhost")
    }
    
    return nil
}

// Layer dependency checking
func CheckLayerDependencies(commandLayers *layers.ParameterLayers) error {
    // Check if security layer is present when auth layer is used
    _, hasAuth := commandLayers.Get("auth")
    _, hasSecurity := commandLayers.Get("security")
    
    if hasAuth && !hasSecurity {
        return fmt.Errorf("auth layer requires security layer to be present")
    }
    
    return nil
}
```

### 4. Layer Middleware Integration

Create layers that work seamlessly with middleware for loading values:

```go
// Layer designed to work with environment variable middleware
func NewEnvAwareLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "env-config",
        "Environment Configuration",
        layers.WithParameterDefinitions(
            // These parameters will be automatically loaded from env vars
            // when using env middleware with prefix "MYAPP_"
            parameters.NewParameterDefinition("api-url", parameters.ParameterTypeString,
                parameters.WithDefault("http://localhost:8080"),
                parameters.WithHelp("API URL (env: MYAPP_API_URL)")),
            parameters.NewParameterDefinition("timeout", parameters.ParameterTypeDuration,
                parameters.WithDefault("30s"),
                parameters.WithHelp("Request timeout (env: MYAPP_TIMEOUT)")),
            parameters.NewParameterDefinition("debug", parameters.ParameterTypeBool,
                parameters.WithDefault(false),
                parameters.WithHelp("Enable debug mode (env: MYAPP_DEBUG)")),
        ),
    )
}

// Layer for configuration file integration
func NewConfigFileLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "config",
        "Configuration File Settings",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("config-file", parameters.ParameterTypeFile,
                parameters.WithHelp("Configuration file path")),
            parameters.NewParameterDefinition("config-format", parameters.ParameterTypeChoice,
                parameters.WithChoices("json", "yaml", "toml"),
                parameters.WithDefault("yaml"),
                parameters.WithHelp("Configuration file format")),
        ),
    )
}
```

## Best Practices

### 1. Layer Organization Principles

**Keep layers focused and cohesive**:
```go
// Good: Focused layer
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    // Only database-related parameters
}

// Bad: Mixed concerns
func NewDatabaseAndLoggingLayer() (layers.ParameterLayer, error) {
    // Mixes database and logging - should be separate layers
}
```

**Use descriptive names and help text**:
```go
parameters.NewParameterDefinition(
    "db-connection-timeout",              // Clear, descriptive name
    parameters.ParameterTypeDuration,
    parameters.WithHelp("Maximum time to wait for database connection establishment"),
    parameters.WithDefault("30s"),
)
```

**Provide sensible defaults**:
```go
// Good: Sensible defaults that work out of the box
parameters.NewParameterDefinition("log-level", parameters.ParameterTypeChoice,
    parameters.WithDefault("info"),    // Good default for most use cases
    parameters.WithChoices("debug", "info", "warn", "error", "fatal"))

// Consider environment-specific defaults
defaultLogLevel := "info"
if os.Getenv("NODE_ENV") == "development" {
    defaultLogLevel = "debug"
}
```

### 2. Layer Composition Guidelines

**Compose layers logically**:
```go
// Server command needs server, database, and logging configuration
serverCmd := cmds.NewCommandDescription("serve",
    cmds.WithLayersList(serverLayer, databaseLayer, loggingLayer))

// Client command only needs connection and logging
clientCmd := cmds.NewCommandDescription("client", 
    cmds.WithLayersList(connectionLayer, loggingLayer))

// Backup command needs database, output formatting, but no server
backupCmd := cmds.NewCommandDescription("backup",
    cmds.WithLayersList(databaseLayer, glazedLayer, loggingLayer))
```

**Order layers by precedence**:
```go
// Place more specific layers before general ones
// This helps with parameter resolution order
cmds.NewCommandDescription("command",
    cmds.WithLayersList(
        commandSpecificLayer,  // Most specific
        featureLayer,          // Feature-specific  
        loggingLayer,          // General infrastructure
    ))
```

### 3. Parameter Design Best Practices

**Use appropriate parameter types**:
```go
// Use specific types for better validation and help
parameters.NewParameterDefinition("timeout", parameters.ParameterTypeDuration,
    parameters.WithDefault("30s"))  // Better than string

parameters.NewParameterDefinition("log-level", parameters.ParameterTypeChoice,
    parameters.WithChoices("debug", "info", "warn"))  // Better than free-form string
    
parameters.NewParameterDefinition("api-key", parameters.ParameterTypeSecret)  // Masks value
```

**Group related parameters with prefixes**:
```go
// Database parameters
"db-host", "db-port", "db-name", "db-username", "db-password"

// Server parameters  
"server-host", "server-port", "server-timeout"

// Cache parameters
"cache-enabled", "cache-ttl", "cache-size"
```

**Use consistent naming conventions**:
```go
// Consistent patterns make CLI more intuitive
"db-*"      // Database parameters
"log-*"     // Logging parameters
"cache-*"   // Cache parameters
"*-timeout" // Timeout parameters
"*-enabled" // Boolean enable/disable flags
```

### 4. Error Handling and Validation

**Validate layer configuration early**:
```go
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    layer, err := layers.NewParameterLayer("database", "Database Configuration")
    if err != nil {
        return nil, fmt.Errorf("failed to create database layer: %w", err)
    }
    
    // Validate parameter definitions
    if err := validateParameterDefinitions(layer); err != nil {
        return nil, fmt.Errorf("invalid parameter definitions: %w", err)
    }
    
    return layer, nil
}
```

**Provide clear error messages**:
```go
func GetDatabaseSettings(parsedLayers *layers.ParsedLayers) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    if err := parsedLayers.InitializeStruct("database", settings); err != nil {
        return nil, fmt.Errorf("failed to initialize database settings: %w", err)
    }
    
    // Additional validation
    if settings.Host == "" {
        return nil, fmt.Errorf("database host cannot be empty")
    }
    
    if settings.Port <= 0 || settings.Port > 65535 {
        return nil, fmt.Errorf("database port must be between 1 and 65535, got %d", settings.Port)
    }
    
    return settings, nil
}
```

### 5. Documentation and Help Text

**Write helpful parameter descriptions**:
```go
parameters.NewParameterDefinition(
    "db-connection-timeout",
    parameters.ParameterTypeDuration,
    parameters.WithHelp("Maximum time to wait for database connection establishment. " +
                        "Use format like '30s', '1m', '5m30s'. Set to 0 to disable timeout."),
    parameters.WithDefault("30s"),
)
```

**Include examples in layer documentation**:
```go
// NewDatabaseLayer creates a database configuration layer.
// 
// Example usage:
//   --db-host localhost --db-port 5432 --db-name myapp
//   --db-host prod.example.com --db-port 5432 --db-ssl-mode require
//
// Environment variables (when using env middleware):
//   MYAPP_DB_HOST, MYAPP_DB_PORT, MYAPP_DB_NAME
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    // ...
}
```

## Testing Layers

### Unit Testing Layer Creation

```go
func TestNewDatabaseLayer(t *testing.T) {
    layer, err := NewDatabaseLayer()
    assert.NoError(t, err)
    assert.NotNil(t, layer)
    
    // Test layer metadata
    assert.Equal(t, "database", layer.GetSlug())
    assert.Equal(t, "Database Configuration", layer.GetName())
    
    // Test parameter definitions
    params := layer.GetParameterDefinitions()
    
    // Check that required parameters exist
    hostParam := params.Get("db-host")
    assert.NotNil(t, hostParam)
    assert.Equal(t, parameters.ParameterTypeString, hostParam.Type)
    assert.Equal(t, "localhost", hostParam.Default)
    
    portParam := params.Get("db-port")
    assert.NotNil(t, portParam)
    assert.Equal(t, parameters.ParameterTypeInteger, portParam.Type)
    assert.Equal(t, 5432, portParam.Default)
}
```

### Integration Testing with Commands

```go
func TestDatabaseLayerIntegration(t *testing.T) {
    // Create command with database layer
    dbLayer, err := NewDatabaseLayer()
    assert.NoError(t, err)
    
    cmd := cmds.NewCommandDescription("test-cmd",
        cmds.WithLayersList(dbLayer))
    
    // Test parameter parsing
    testCases := []struct {
        name     string
        args     []string
        expected DatabaseSettings
    }{
        {
            name: "default values",
            args: []string{},
            expected: DatabaseSettings{
                Host: "localhost",
                Port: 5432,
            },
        },
        {
            name: "custom values",
            args: []string{"--db-host", "prod.example.com", "--db-port", "3306"},
            expected: DatabaseSettings{
                Host: "prod.example.com", 
                Port: 3306,
            },
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Parse arguments using runner
            ctx := context.Background()
            parsedLayers, err := runner.ParseCommand(ctx, cmd, tc.args)
            assert.NoError(t, err)
            
            // Extract settings
            settings, err := GetDatabaseSettings(parsedLayers)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, *settings)
        })
    }
}
```

### Testing Layer Composition

```go
func TestLayerComposition(t *testing.T) {
    // Create multiple layers
    dbLayer, _ := NewDatabaseLayer()
    loggingLayer, _ := NewLoggingLayer()
    serverLayer, _ := NewServerLayer()
    
    // Test command with multiple layers
    cmd := cmds.NewCommandDescription("serve",
        cmds.WithLayersList(serverLayer, dbLayer, loggingLayer))
    
    // Verify all layers are present
    layers := cmd.Layers
    assert.True(t, layers.Has("server"))
    assert.True(t, layers.Has("database"))
    assert.True(t, layers.Has("logging"))
    
    // Test that parameters from all layers are available
    allParams := cmd.GetAllParameters()
    
    // Check server parameters
    assert.Contains(t, allParams, "host")
    assert.Contains(t, allParams, "port")
    
    // Check database parameters
    assert.Contains(t, allParams, "db-host")
    assert.Contains(t, allParams, "db-port")
    
    // Check logging parameters
    assert.Contains(t, allParams, "log-level")
    assert.Contains(t, allParams, "log-format")
}
```

---

This guide provides a comprehensive foundation for understanding and working with Glazed's layer system. The key takeaway is that layers enable you to build modular, reusable, and maintainable command-line interfaces by organizing related parameters into logical groups that can be composed as needed.

For hands-on practice, see the [Custom Layer Tutorial](../tutorials/custom-layer.md) which walks through creating a custom logging layer step by step.
