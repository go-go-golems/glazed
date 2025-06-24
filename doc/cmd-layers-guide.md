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
    DbHost   string // Database host (duplicated!)
    DbPort   int    // Database port (duplicated!)
    Output   string // Backup file path
    LogLevel string // Logging level (duplicated again!)
    // ... more fields
}
```

This approach suffers from:
- **No reusability**: Can't share database configuration between commands
- **No flexibility**: Can't conditionally include parameter groups
- **Poor discoverability**: Help text becomes overwhelming
- **Testing complexity**: Hard to test parameter combinations in isolation

## Mental Model & Why Use Layers

### The Layer Mental Model

Think of parameter layers like **modular configuration building blocks**. Each layer represents a cohesive group of related parameters that can be:

1. **Defined once, used everywhere**: A "database" layer with connection parameters
2. **Mixed and matched**: Combine different layers for different commands
3. **Processed independently**: Each layer can have its own logic and validation
4. **Conditionally included**: Only add layers that make sense for each command

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│   Database      │  │   Logging       │  │   Server        │
│   Layer         │  │   Layer         │  │   Layer         │
├─────────────────┤  ├─────────────────┤  ├─────────────────┤
│ --db-host       │  │ --log-level     │  │ --host          │
│ --db-port       │  │ --log-format    │  │ --port          │
│ --db-user       │  │ --verbose       │  │ --workers       │
│ --db-password   │  │ --log-file      │  │ --timeout       │
└─────────────────┘  └─────────────────┘  └─────────────────┘

Command Assembly:
server-cmd = Database + Logging + Server layers
client-cmd = Logging layer only
backup-cmd = Database + Logging layers
```

### Key Benefits of the Layer Approach

#### 1. **Reusability & DRY Principle**
Define parameter groups once, use them everywhere:

```go
// Define once
databaseLayer := NewDatabaseLayer()
loggingLayer := NewLoggingLayer()

// Use everywhere
serverCmd := NewCommand("server", databaseLayer, loggingLayer, serverLayer)
backupCmd := NewCommand("backup", databaseLayer, loggingLayer)
clientCmd := NewCommand("client", loggingLayer)
```

#### 2. **Logical Organization & Discoverability**
Parameters are grouped by functionality, making help text cleaner:

```bash
$ myapp server --help
Server Options:
  --host string     Server host (default "localhost")
  --port int        Server port (default 8080)

Database Options:
  --db-host string  Database host (default "localhost")
  --db-port int     Database port (default 5432)

Logging Options:
  --log-level string  Log level (default "info")
  --verbose          Enable verbose logging
```

#### 3. **Namespace Management**
Each layer can have its own prefix to avoid naming conflicts:

```go
databaseLayer := NewParameterLayer("database", "Database configuration",
    WithPrefix("db-"),
    WithParameterDefinitions(
        parameters.NewParameterDefinition("host", ...),
        parameters.NewParameterDefinition("port", ...),
    ),
)

serverLayer := NewParameterLayer("server", "Server configuration", 
    WithParameterDefinitions(
        parameters.NewParameterDefinition("host", ...),  // No conflict!
        parameters.NewParameterDefinition("port", ...),  // No conflict!
    ),
)
```

This creates flags like `--db-host` and `--host` that don't conflict.

#### 4. **Independent Processing**
Each layer can be processed, validated, and transformed independently:

```go
// Extract just the database settings
dbSettings := &DatabaseSettings{}
err := parsedLayers.InitializeStruct("database", dbSettings)

// Extract just the logging settings
logSettings := &LoggingSettings{}
err := parsedLayers.InitializeStruct("logging", logSettings)
```

#### 5. **Conditional Logic**
Only include layers that make sense for each command:

```go
func NewServerCommand() *Command {
    layers := []ParameterLayer{serverLayer, loggingLayer}
    
    if needsDatabase {
        layers = append(layers, databaseLayer)
    }
    
    if isProduction {
        layers = append(layers, monitoringLayer)
    }
    
    return NewCommand("server", layers...)
}
```

#### 6. **Testing Isolation**
Test parameter groups in isolation:

```go
func TestDatabaseLayer(t *testing.T) {
    layer := NewDatabaseLayer()
    
    // Test just database parameter validation
    parsedLayer, err := layer.ParseFromMap(map[string]interface{}{
        "host": "invalid-host-name",
        "port": -1,
    })
    
    assert.Error(t, err) // Should fail validation
}
```

### Real-World Layer Architecture

Consider a microservice CLI tool with these layers:

```go
// Core layers used by all commands
authLayer := NewAuthLayer()       // --api-key, --token, --auth-url
loggingLayer := NewLoggingLayer() // --log-level, --verbose, --log-file

// Service-specific layers
databaseLayer := NewDatabaseLayer()     // --db-host, --db-port, --db-name
kafkaLayer := NewKafkaLayer()          // --kafka-brokers, --kafka-topic
redisLayer := NewRedisLayer()          // --redis-host, --redis-port

// Deployment layers
kubernetesLayer := NewKubernetesLayer() // --namespace, --context
awsLayer := NewAWSLayer()               // --region, --profile

// Command compositions
deployCmd := NewCommand("deploy", 
    authLayer, loggingLayer, kubernetesLayer, awsLayer)

migrateCmd := NewCommand("migrate", 
    authLayer, loggingLayer, databaseLayer)

streamCmd := NewCommand("stream", 
    authLayer, loggingLayer, kafkaLayer, redisLayer)
```

Each command only gets the parameters it needs, but layers are reused across commands where appropriate.

## Understanding Layers

### The Two-Phase Design

Glazed's layer system operates in two distinct phases:

#### Phase 1: Definition (ParameterLayers)
- **What**: Define the structure and metadata of parameters
- **When**: At compile time or application startup
- **Purpose**: Describe what parameters exist, their types, defaults, validation rules

#### Phase 2: Runtime (ParsedLayers)  
- **What**: Store actual parameter values after parsing from various sources
- **When**: At runtime, after processing CLI args, config files, environment variables, etc.
- **Purpose**: Hold the actual values that your application logic will use

```go
// Phase 1: Definition
databaseLayer := NewParameterLayer("database", "Database configuration",
    WithParameterDefinitions(
        parameters.NewParameterDefinition("host", 
            parameters.ParameterTypeString,
            parameters.WithDefault("localhost"),
            parameters.WithHelp("Database host"),
        ),
        parameters.NewParameterDefinition("port",
            parameters.ParameterTypeInt, 
            parameters.WithDefault(5432),
            parameters.WithHelp("Database port"),
        ),
    ),
)

// Phase 2: Runtime parsing
parsedLayers := layers.NewParsedLayers()
err := middlewares.ExecuteMiddlewares(
    parameterLayers,  // Definitions from Phase 1
    parsedLayers,     // Values filled during Phase 2
    middlewares...,
)

// Phase 2: Extract values for use
dbSettings := &DatabaseSettings{}
err := parsedLayers.InitializeStruct("database", dbSettings)
```

This separation enables:
- **Static analysis**: Know what parameters exist before runtime
- **Dynamic configuration**: Parse values from multiple sources
- **Flexible processing**: Apply different parsing strategies for different scenarios

## Core Layer Concepts

### Parameter Layer Interface

Every layer implements the `ParameterLayer` interface:

```go
type ParameterLayer interface {
    // Core identification
    GetName() string        // Human-readable name: "Database Configuration"
    GetSlug() string        // Machine identifier: "database"
    GetDescription() string // Help text: "Database connection settings"
    GetPrefix() string      // Flag prefix: "db-" (creates --db-host, --db-port)
    
    // Parameter management
    AddFlags(...*parameters.ParameterDefinition)
    GetParameterDefinitions() *parameters.ParameterDefinitions
    InitializeParameterDefaultsFromStruct(interface{}) error
    
    // Lifecycle
    Clone() ParameterLayer
}
```

### Parameter Layers Collection

`ParameterLayers` is an ordered collection of individual `ParameterLayer` objects:

```go
// Create layers collection
layers := layers.NewParameterLayers(
    layers.WithLayers(
        databaseLayer,    // slug: "database"
        serverLayer,      // slug: "server"  
        loggingLayer,     // slug: "logging"
    ),
)

// Access individual layers
dbLayer, exists := layers.Get("database")

// Iterate through all layers
layers.ForEach(func(slug string, layer ParameterLayer) {
    fmt.Printf("Layer: %s (%s)\n", layer.GetName(), slug)
})
```

### Parsed Layer Structure

After processing, `ParsedLayer` contains the actual runtime values:

```go
type ParsedLayer struct {
    Layer      ParameterLayer              // Reference to original definition
    Parameters *parameters.ParsedParameters // Actual runtime values
}

// Access parsed values
value, exists := parsedLayer.GetParameter("host")
if exists {
    fmt.Printf("Database host: %v\n", value)
}
```

### Parsed Layers Collection

`ParsedLayers` is the runtime collection containing all parsed values:

```go
// Extract settings for a specific layer
dbSettings := &DatabaseSettings{}
err := parsedLayers.InitializeStruct("database", dbSettings)

// Get a specific parameter value
host, exists := parsedLayers.GetParameter("database", "host")

// Get all parameters as a flat map (useful for templates)
allParams := parsedLayers.GetDataMap()
```

## Layer Types & Components

### Standard Layer Implementation

The most common layer type is `ParameterLayerImpl`:

```go
// Create a basic layer
layer, err := layers.NewParameterLayer(
    "api",                              // slug
    "API Configuration",                // name
    layers.WithDescription("REST API connection settings"),
    layers.WithPrefix("api-"),          // Creates --api-host, --api-port, etc.
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition("host",
            parameters.ParameterTypeString,
            parameters.WithDefault("localhost"),
            parameters.WithHelp("API server host"),
        ),
        parameters.NewParameterDefinition("port",
            parameters.ParameterTypeInt,
            parameters.WithDefault(8080),
            parameters.WithHelp("API server port"),
        ),
        parameters.NewParameterDefinition("timeout",
            parameters.ParameterTypeDuration,
            parameters.WithDefault("30s"),
            parameters.WithHelp("Request timeout"),
        ),
    ),
)
```

### Cobra Integration Layer

For CLI integration, layers implement `CobraParameterLayer`:

```go
// Add layer flags to Cobra command
err := layer.AddLayerToCobraCommand(cobraCmd)

// Parse values from Cobra command
parsedLayer, err := layer.ParseLayerFromCobraCommand(cobraCmd)
```

### Specialized Layer Types

#### Glazed Settings Layer
A pre-built layer for output formatting:

```go
glazedLayer, err := settings.NewGlazedParameterLayers()
// Provides --output, --fields, --format, etc.
```

#### Custom Domain Layers
Create domain-specific layers:

```go
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer("database", "Database Configuration",
        layers.WithPrefix("db-"),
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("host", 
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
            ),
            parameters.NewParameterDefinition("port",
                parameters.ParameterTypeInt,
                parameters.WithDefault(5432),
                parameters.WithValidation(validatePort),
            ),
            parameters.NewParameterDefinition("ssl-mode",
                parameters.ParameterTypeChoice,
                parameters.WithDefault("prefer"),
                parameters.WithChoices("disable", "prefer", "require"),
            ),
        ),
    )
}

func validatePort(value interface{}) error {
    port := value.(int)
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535")
    }
    return nil
}
```

## Creating and Working with Layers

### Step 1: Define Parameter Layers

```go
// Define individual parameter definitions
hostParam := parameters.NewParameterDefinition("host",
    parameters.ParameterTypeString,
    parameters.WithDefault("localhost"),
    parameters.WithHelp("Server host address"),
    parameters.WithRequired(false),
)

portParam := parameters.NewParameterDefinition("port",
    parameters.ParameterTypeInt,
    parameters.WithDefault(8080),
    parameters.WithHelp("Server port number"),
    parameters.WithValidation(func(v interface{}) error {
        port := v.(int)
        if port < 1 || port > 65535 {
            return fmt.Errorf("port must be between 1 and 65535")
        }
        return nil
    }),
)

// Create the layer
serverLayer, err := layers.NewParameterLayer(
    "server",                     // slug
    "Server Configuration",       // name
    layers.WithDescription("HTTP server configuration options"),
    layers.WithParameterDefinitions(hostParam, portParam),
)
```

### Step 2: Create Layers Collection

```go
// Combine multiple layers
parameterLayers := layers.NewParameterLayers(
    layers.WithLayers(
        serverLayer,
        databaseLayer,
        loggingLayer,
    ),
)

// Or build incrementally
parameterLayers := layers.NewParameterLayers()
parameterLayers.AppendLayers(serverLayer, databaseLayer)
parameterLayers.PrependLayers(loggingLayer) // Higher priority
```

### Step 3: Define Settings Structures

```go
// Create struct that maps to layer parameters
type ServerSettings struct {
    Host string `glazed.parameter:"host"`
    Port int    `glazed.parameter:"port"`
}

type DatabaseSettings struct {
    Host     string `glazed.parameter:"host"`
    Port     int    `glazed.parameter:"port"`
    Username string `glazed.parameter:"username"`
    Password string `glazed.parameter:"password"`
    SSL      bool   `glazed.parameter:"ssl"`
}

type LoggingSettings struct {
    Level   string `glazed.parameter:"level"`
    Verbose bool   `glazed.parameter:"verbose"`
    Format  string `glazed.parameter:"format"`
}
```

### Step 4: Process Parameters with Middlewares

```go
// Create empty parsed layers
parsedLayers := layers.NewParsedLayers()

// Configure middleware chain
middlewares := []middlewares.Middleware{
    middlewares.ParseFromCobraCommand(cmd),  // CLI args (highest priority)
    middlewares.LoadParametersFromFile("config.yaml"), // Config file
    middlewares.UpdateFromEnv("MYAPP"),      // Environment variables
    middlewares.SetFromDefaults(),           // Default values (lowest priority)
}

// Execute middleware chain
err := middlewares.ExecuteMiddlewares(parameterLayers, parsedLayers, middlewares...)
if err != nil {
    return fmt.Errorf("failed to parse parameters: %w", err)
}
```

### Step 5: Extract Settings

```go
// Extract layer-specific settings
serverSettings := &ServerSettings{}
err := parsedLayers.InitializeStruct("server", serverSettings)
if err != nil {
    return fmt.Errorf("failed to parse server settings: %w", err)
}

databaseSettings := &DatabaseSettings{}
err = parsedLayers.InitializeStruct("database", databaseSettings)
if err != nil {
    return fmt.Errorf("failed to parse database settings: %w", err)
}

loggingSettings := &LoggingSettings{}
err = parsedLayers.InitializeStruct("logging", loggingSettings)
if err != nil {
    return fmt.Errorf("failed to parse logging settings: %w", err)
}

// Now use the settings
server := NewServer(serverSettings)
db := ConnectDatabase(databaseSettings)
logger := SetupLogging(loggingSettings)
```

### Step 6: Command Integration

```go
type MyCommand struct {
    *cmds.CommandDescription
}

func (c *MyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Extract settings
    serverSettings := &ServerSettings{}
    if err := parsedLayers.InitializeStruct("server", serverSettings); err != nil {
        return err
    }
    
    // Use settings to run application logic
    server := NewServer(serverSettings.Host, serverSettings.Port)
    return server.Start(ctx)
}

// Create command with layers
func NewMyCommand() (*MyCommand, error) {
    serverLayer, err := NewServerLayer()
    if err != nil {
        return nil, err
    }
    
    return &MyCommand{
        CommandDescription: cmds.NewCommandDescription(
            "serve",
            cmds.WithShort("Start the HTTP server"),
            cmds.WithLayers(serverLayer),
        ),
    }, nil
}
```

## Practical Examples

### Example 1: Simple Database Command

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

// Settings struct
type DatabaseSettings struct {
    Host     string `glazed.parameter:"host"`
    Port     int    `glazed.parameter:"port"`
    Database string `glazed.parameter:"database"`
    Username string `glazed.parameter:"username"`
    Password string `glazed.parameter:"password"`
}

// Layer factory
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithPrefix("db-"),
        layers.WithDescription("Database connection parameters"),
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Database host"),
            ),
            parameters.NewParameterDefinition("port",
                parameters.ParameterTypeInt, 
                parameters.WithDefault(5432),
                parameters.WithHelp("Database port"),
            ),
            parameters.NewParameterDefinition("database",
                parameters.ParameterTypeString,
                parameters.WithDefault("myapp"),
                parameters.WithHelp("Database name"),
            ),
            parameters.NewParameterDefinition("username",
                parameters.ParameterTypeString,
                parameters.WithRequired(true),
                parameters.WithHelp("Database username"),
            ),
            parameters.NewParameterDefinition("password",
                parameters.ParameterTypeString,
                parameters.WithHelp("Database password"),
            ),
        ),
    )
}

// Command implementation
type ConnectCommand struct {
    *cmds.CommandDescription
}

func (c *ConnectCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    settings := &DatabaseSettings{}
    if err := parsedLayers.InitializeStruct("database", settings); err != nil {
        return fmt.Errorf("failed to parse database settings: %w", err)
    }
    
    fmt.Printf("Connecting to %s:%d/%s as %s\n", 
        settings.Host, settings.Port, settings.Database, settings.Username)
    
    // Actual database connection logic would go here
    return nil
}

func NewConnectCommand() (*ConnectCommand, error) {
    databaseLayer, err := NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Add command settings layer for debugging capabilities
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    return &ConnectCommand{
        CommandDescription: cmds.NewCommandDescription(
            "connect",
            cmds.WithShort("Connect to the database"),
            cmds.WithLayers(databaseLayer, commandSettingsLayer),
        ),
    }, nil
}
```

### Example 2: Multi-Layer Web Service Command

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "time"
    
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Multiple setting structures
type ServerSettings struct {
    Host         string        `glazed.parameter:"host"`
    Port         int           `glazed.parameter:"port"`
    ReadTimeout  time.Duration `glazed.parameter:"read-timeout"`
    WriteTimeout time.Duration `glazed.parameter:"write-timeout"`
}

type DatabaseSettings struct {
    Host         string `glazed.parameter:"host"`
    Port         int    `glazed.parameter:"port"`
    Database     string `glazed.parameter:"database"`
    MaxConns     int    `glazed.parameter:"max-connections"`
    ConnTimeout  time.Duration `glazed.parameter:"connection-timeout"`
}

type LoggingSettings struct {
    Level  string `glazed.parameter:"level"`
    Format string `glazed.parameter:"format"`
    File   string `glazed.parameter:"file"`
}

// Layer factories
func NewServerLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "server",
        "HTTP Server Configuration", 
        layers.WithDescription("HTTP server settings"),
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
            ),
            parameters.NewParameterDefinition("port",
                parameters.ParameterTypeInt,
                parameters.WithDefault(8080),
            ),
            parameters.NewParameterDefinition("read-timeout",
                parameters.ParameterTypeDuration,
                parameters.WithDefault("30s"),
            ),
            parameters.NewParameterDefinition("write-timeout",
                parameters.ParameterTypeDuration,
                parameters.WithDefault("30s"),
            ),
        ),
    )
}

func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithPrefix("db-"),
        layers.WithDescription("Database connection settings"),
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
            ),
            parameters.NewParameterDefinition("port",
                parameters.ParameterTypeInt,
                parameters.WithDefault(5432),
            ),
            parameters.NewParameterDefinition("database",
                parameters.ParameterTypeString,
                parameters.WithDefault("myapp"),
            ),
            parameters.NewParameterDefinition("max-connections",
                parameters.ParameterTypeInt,
                parameters.WithDefault(10),
            ),
            parameters.NewParameterDefinition("connection-timeout",
                parameters.ParameterTypeDuration,
                parameters.WithDefault("10s"),
            ),
        ),
    )
}

func NewLoggingLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "logging",
        "Logging Configuration",
        layers.WithPrefix("log-"),
        layers.WithDescription("Logging settings"),
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("level",
                parameters.ParameterTypeChoice,
                parameters.WithDefault("info"),
                parameters.WithChoices("debug", "info", "warn", "error"),
            ),
            parameters.NewParameterDefinition("format",
                parameters.ParameterTypeChoice,
                parameters.WithDefault("text"),
                parameters.WithChoices("text", "json"),
            ),
            parameters.NewParameterDefinition("file",
                parameters.ParameterTypeString,
                parameters.WithDefault(""),
                parameters.WithHelp("Log file path (empty for stdout)"),
            ),
        ),
    )
}

// Command that uses all three layers
type ServeCommand struct {
    *cmds.CommandDescription
}

func (c *ServeCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Extract all settings
    serverSettings := &ServerSettings{}
    if err := parsedLayers.InitializeStruct("server", serverSettings); err != nil {
        return fmt.Errorf("failed to parse server settings: %w", err)
    }
    
    databaseSettings := &DatabaseSettings{}
    if err := parsedLayers.InitializeStruct("database", databaseSettings); err != nil {
        return fmt.Errorf("failed to parse database settings: %w", err)
    }
    
    loggingSettings := &LoggingSettings{}
    if err := parsedLayers.InitializeStruct("logging", loggingSettings); err != nil {
        return fmt.Errorf("failed to parse logging settings: %w", err)
    }
    
    // Setup logging
    log.Printf("Log level: %s, format: %s", loggingSettings.Level, loggingSettings.Format)
    if loggingSettings.File != "" {
        log.Printf("Logging to file: %s", loggingSettings.File)
    }
    
    // Setup database (placeholder)
    log.Printf("Database: %s:%d/%s (max %d connections, timeout %v)",
        databaseSettings.Host, databaseSettings.Port, databaseSettings.Database,
        databaseSettings.MaxConns, databaseSettings.ConnTimeout)
    
    // Setup and start server
    server := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", serverSettings.Host, serverSettings.Port),
        ReadTimeout:  serverSettings.ReadTimeout,
        WriteTimeout: serverSettings.WriteTimeout,
        Handler:      http.DefaultServeMux,
    }
    
    log.Printf("Starting server on %s:%d", serverSettings.Host, serverSettings.Port)
    return server.ListenAndServe()
}

func NewServeCommand() (*ServeCommand, error) {
    serverLayer, err := NewServerLayer()
    if err != nil {
        return nil, err
    }
    
    databaseLayer, err := NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    loggingLayer, err := NewLoggingLayer() 
    if err != nil {
        return nil, err
    }
    
    // Add command settings layer for debugging capabilities
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    return &ServeCommand{
        CommandDescription: cmds.NewCommandDescription(
            "serve",
            cmds.WithShort("Start the web server"),
            cmds.WithLong("Start the web server with database connectivity and logging"),
            cmds.WithLayers(serverLayer, databaseLayer, loggingLayer, commandSettingsLayer),
        ),
    }, nil
}
```

Usage would look like:
```bash
# Use defaults
myapp serve

# Override server settings
myapp serve --host 0.0.0.0 --port 9000

# Override database settings  
myapp serve --db-host postgres.example.com --db-port 5433 --db-database production

# Override logging
myapp serve --log-level debug --log-format json --log-file /var/log/myapp.log

# Mix and match
myapp serve --port 9000 --db-host postgres --log-level warn
```

### Example 3: Conditional Layer Inclusion

```go
type DeployCommand struct {
    *cmds.CommandDescription
}

func NewDeployCommand(environment string) (*DeployCommand, error) {
    // Always include these layers
    layers := []layers.ParameterLayer{}
    
    // Basic deployment layer
    deployLayer, err := NewDeploymentLayer()
    if err != nil {
        return nil, err
    }
    layers = append(layers, deployLayer)
    
    // Add environment-specific layers
    switch environment {
    case "kubernetes":
        k8sLayer, err := NewKubernetesLayer()
        if err != nil {
            return nil, err
        }
        layers = append(layers, k8sLayer)
        
    case "aws":
        awsLayer, err := NewAWSLayer()
        if err != nil {
            return nil, err
        }
        layers = append(layers, awsLayer)
        
    case "docker":
        dockerLayer, err := NewDockerLayer()
        if err != nil {
            return nil, err
        }
        layers = append(layers, dockerLayer)
    }
    
    // Optional monitoring layer
    if isProduction(environment) {
        monitoringLayer, err := NewMonitoringLayer()
        if err != nil {
            return nil, err
        }
        layers = append(layers, monitoringLayer)
    }
    
    return &DeployCommand{
        CommandDescription: cmds.NewCommandDescription(
            "deploy",
            cmds.WithShort(fmt.Sprintf("Deploy to %s", environment)),
            cmds.WithLayers(layers...),
        ),
    }, nil
}
```

## Advanced Patterns

### Layer Composition and Inheritance

```go
// Base layer with common parameters
func NewBaseServiceLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "base",
        "Base Service Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("name",
                parameters.ParameterTypeString,
                parameters.WithRequired(true),
            ),
            parameters.NewParameterDefinition("version",
                parameters.ParameterTypeString,
                parameters.WithDefault("latest"),
            ),
        ),
    )
}

// Extended layer that builds on base
func NewWebServiceLayer() (layers.ParameterLayer, error) {
    baseLayer, err := NewBaseServiceLayer()
    if err != nil {
        return nil, err
    }
    
    // Start with base layer definitions
    webLayer, err := layers.NewParameterLayer(
        "web",
        "Web Service Configuration",
        layers.WithParameterDefinitions(baseLayer.GetParameterDefinitions().AsList()...),
    )
    if err != nil {
        return nil, err
    }
    
    // Add web-specific parameters
    webLayer.AddFlags(
        parameters.NewParameterDefinition("port",
            parameters.ParameterTypeInt,
            parameters.WithDefault(8080),
        ),
        parameters.NewParameterDefinition("tls",
            parameters.ParameterTypeBool,
            parameters.WithDefault(false),
        ),
    )
    
    return webLayer, nil
}
```

### Dynamic Layer Creation

```go
// Create layers from configuration
func NewLayersFromConfig(config *Config) (*layers.ParameterLayers, error) {
    result := layers.NewParameterLayers()
    
    for _, layerConfig := range config.Layers {
        layer, err := NewParameterLayer(
            layerConfig.Slug,
            layerConfig.Name,
            layers.WithDescription(layerConfig.Description),
            layers.WithPrefix(layerConfig.Prefix),
        )
        if err != nil {
            return nil, err
        }
        
        // Add parameters from config
        for _, paramConfig := range layerConfig.Parameters {
            param, err := createParameterFromConfig(paramConfig)
            if err != nil {
                return nil, err
            }
            layer.AddFlags(param)
        }
        
        result.AppendLayers(layer)
    }
    
    return result, nil
}
```

### Layer Validation and Transformation

```go
func NewValidatedDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer(
        "database",
        "Database Configuration",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithValidation(validateHostname),
            ),
            parameters.NewParameterDefinition("port",
                parameters.ParameterTypeInt,
                parameters.WithDefault(5432),
                parameters.WithValidation(validatePort),
                parameters.WithTransformation(normalizePort),
            ),
            parameters.NewParameterDefinition("connection-string",
                parameters.ParameterTypeString,
                parameters.WithValidation(validateConnectionString),
                parameters.WithTransformation(sanitizeConnectionString),
            ),
        ),
    )
}

func validateHostname(value interface{}) error {
    host := value.(string)
    if net.ParseIP(host) == nil {
        // Try to resolve hostname
        _, err := net.LookupHost(host)
        if err != nil {
            return fmt.Errorf("invalid hostname: %s", host)
        }
    }
    return nil
}

func validatePort(value interface{}) error {
    port := value.(int)
    if port < 1 || port > 65535 {
        return fmt.Errorf("port must be between 1 and 65535, got %d", port)
    }
    return nil
}

func normalizePort(value interface{}) (interface{}, error) {
    // Convert string port to int if needed
    switch v := value.(type) {
    case string:
        port, err := strconv.Atoi(v)
        if err != nil {
            return nil, fmt.Errorf("invalid port string: %s", v)
        }
        return port, nil
    case int:
        return v, nil
    default:
        return nil, fmt.Errorf("port must be string or int, got %T", v)
    }
}
```

### Layer Serialization and Persistence

```go
// Save layer configuration to file
func SaveLayersToFile(layers *layers.ParameterLayers, filename string) error {
    data, err := yaml.Marshal(layers)
    if err != nil {
        return err
    }
    
    return ioutil.WriteFile(filename, data, 0644)
}

// Load layer configuration from file
func LoadLayersFromFile(filename string) (*layers.ParameterLayers, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    
    var layers layers.ParameterLayers
    err = yaml.Unmarshal(data, &layers)
    if err != nil {
        return nil, err
    }
    
    return &layers, nil
}

// Save parsed layer values
func SaveParsedLayersToFile(parsedLayers *layers.ParsedLayers, filename string) error {
    data, err := yaml.Marshal(parsedLayers)
    if err != nil {
        return err
    }
    
    return ioutil.WriteFile(filename, data, 0644)
}
```

## Debugging and Introspection

### Built-in CLI Tools for Layer Discovery

Glazed provides several built-in flags to help you understand and debug your layer structure. These debugging flags are provided by the **Command Settings Layer** that you can add to your commands:

#### `--print-parsed-parameters`
Shows detailed information about all layers and their parsed parameters:

```bash
$ myapp serve --host 0.0.0.0 --db-host postgres --print-parsed-parameters
# server:
host: (string)
  value: '0.0.0.0'
        source: defaults: localhost
        source: cobra: 0.0.0.0
                flag: host
                layer: Server Configuration
port: (int)
  value: '8080'
        source: defaults: 8080

# database:
host: (string)
  value: 'postgres'
        source: defaults: localhost
        source: cobra: postgres
                flag: db-host
                layer: Database Configuration
```

This shows:
- **Layer organization**: Each `# layer-name:` section shows parameters grouped by layer
- **Parameter details**: Name, type, and current value
- **Source tracking**: Complete history of where each value came from (defaults → config files → CLI args)
- **Metadata**: Additional context like which CLI flag was used and which layer it belongs to

#### `--print-schema`
Outputs the JSON schema definition of all parameters:

```bash
$ myapp serve --print-schema
{
  "type": "object", 
  "properties": {
    "server": {
      "type": "object",
      "properties": {
        "host": {"type": "string", "default": "localhost"},
        "port": {"type": "integer", "default": 8080}
      }
    },
    "database": {
      "type": "object", 
      "properties": {
        "host": {"type": "string", "default": "localhost"},
        "port": {"type": "integer", "default": 5432}
      }
    }
  }
}
```

#### `--long-help`
Shows comprehensive help with all layers and parameters organized by group:

```bash
$ myapp serve --help --long-help

## Server Configuration:
  --host string     Server host address (default "localhost") 
  --port int        Server port number (default 8080)

## Database Configuration:
  --db-host string  Database host address (default "localhost")
  --db-port int     Database port number (default 5432)

## Logging Configuration:
  --log-level string  Log level (default "info")
  --log-format string Log format (default "text")
```

### Adding Debug Capabilities to Your Commands

To enable these debugging flags in your commands, you need to include the **Command Settings Layer**:

```go
import (
    "github.com/go-go-golems/glazed/pkg/cli"
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
)

func NewMyCommand() (*MyCommand, error) {
    // Create your application layers
    serverLayer, err := NewServerLayer()
    if err != nil {
        return nil, err
    }
    
    databaseLayer, err := NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Add the command settings layer for debugging flags
    commandSettingsLayer, err := cli.NewCommandSettingsLayer()
    if err != nil {
        return nil, err
    }
    
    return &MyCommand{
        CommandDescription: cmds.NewCommandDescription(
            "serve",
            cmds.WithShort("Start the server"),
            cmds.WithLayers(
                serverLayer,
                databaseLayer,
                commandSettingsLayer, // Enables --print-parsed-parameters, --print-schema, etc.
            ),
        ),
    }, nil
}
```

The Command Settings Layer provides these flags:
- `--print-parsed-parameters` - Shows detailed layer and parameter information
- `--print-schema` - Outputs JSON schema of all parameters
- `--print-yaml` - Prints the command's YAML representation
- `--load-parameters-from-file` - Loads parameters from a file

#### Other Useful Built-in Layers

Glazed also provides other built-in layers you can include:

```go
// Profile management layer
profileLayer, err := cli.NewProfileSettingsLayer()
// Provides: --profile, --profile-file

// Command creation layer (for development/testing)
createCommandLayer, err := cli.NewCreateCommandSettingsLayer() 
// Provides: --create-command, --create-alias, --create-cliopatra
```

### Debugging Common Issues

#### Layer Not Showing Up
If a layer isn't appearing in `--print-parsed-parameters`:

```go
// Check that layer is properly added to command
commandSettingsLayer, _ := cli.NewCommandSettingsLayer() // Don't forget this layer!

cmd := &MyCommand{
    CommandDescription: cmds.NewCommandDescription(
        "serve",
        cmds.WithLayers(
            serverLayer, 
            databaseLayer, 
            commandSettingsLayer, // Required for debugging flags
        ),
    ),
}
```

If the `--print-parsed-parameters` flag itself is missing, make sure you've included the Command Settings Layer.

#### Parameter Not Found
If a parameter isn't being parsed:

```go
// Check parameter name matches struct tag exactly
type Settings struct {
    Host string `glazed.parameter:"host"`     // Must match parameter definition name
    Port int    `glazed.parameter:"port"`
}

// Check parameter is defined in layer
layer, err := layers.NewParameterLayer("server", "Server Configuration",
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition("host", ...), // Must match struct tag
        parameters.NewParameterDefinition("port", ...), 
    ),
)
```

#### Wrong Values Being Used
Use `--print-parsed-parameters` to trace the parameter value lifecycle:

```bash
# Shows value progression: defaults → config → environment → CLI flags
$ myapp serve --host override --print-parsed-parameters
host: (string)
  value: 'override'
        source: defaults: localhost
        source: config: config-host  
        source: env: ENV_HOST
        source: cobra: override      # Final value wins
```

#### Prefix Conflicts
Check if prefixes are causing naming issues:

```bash
$ myapp serve --print-parsed-parameters | grep -A5 "host"
# Should show both:
# server layer: host (no prefix)
# database layer: db-host (with prefix)
```

## Best Practices

### 1. Layer Design Principles

#### Single Responsibility
Each layer should represent one cohesive concept:

```go
// Good: Focused on database concerns
func NewDatabaseLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer("database", "Database Configuration",
        layers.WithParameterDefinitions(
            // Only database-related parameters
            parameters.NewParameterDefinition("host", ...),
            parameters.NewParameterDefinition("port", ...),
            parameters.NewParameterDefinition("database", ...),
        ),
    )
}

// Bad: Mixed concerns
func NewMixedLayer() (layers.ParameterLayer, error) {
    return layers.NewParameterLayer("mixed", "Mixed Configuration",
        layers.WithParameterDefinitions(
            // Database parameters
            parameters.NewParameterDefinition("db-host", ...),
            // HTTP parameters
            parameters.NewParameterDefinition("http-port", ...),
            // Logging parameters  
            parameters.NewParameterDefinition("log-level", ...),
        ),
    )
}
```

#### Sensible Defaults
Always provide reasonable defaults:

```go
parameters.NewParameterDefinition("timeout",
    parameters.ParameterTypeDuration,
    parameters.WithDefault("30s"),          // Sensible default
    parameters.WithHelp("Connection timeout"),
)
```

#### Clear Naming
Use descriptive, unambiguous names:

```go
// Good: Clear purpose
parameters.NewParameterDefinition("max-connections", ...)
parameters.NewParameterDefinition("connection-timeout", ...)

// Bad: Ambiguous
parameters.NewParameterDefinition("max", ...)
parameters.NewParameterDefinition("timeout", ...)
```

### 2. Prefix Strategy

Use prefixes to avoid naming conflicts:

```go
// Database layer uses "db-" prefix
databaseLayer := layers.NewParameterLayer("database", "Database Configuration",
    layers.WithPrefix("db-"),
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition("host", ...),  // becomes --db-host
        parameters.NewParameterDefinition("port", ...),  // becomes --db-port
    ),
)

// Server layer has no prefix  
serverLayer := layers.NewParameterLayer("server", "Server Configuration",
    layers.WithParameterDefinitions(
        parameters.NewParameterDefinition("host", ...),  // becomes --host
        parameters.NewParameterDefinition("port", ...),  // becomes --port
    ),
)
```

### 3. Layer Organization

#### Group Related Commands
Use the same layers across related commands:

```go
// Both commands use the same database layer
migrateCmd := NewCommand("migrate", databaseLayer, loggingLayer)
seedCmd := NewCommand("seed", databaseLayer, loggingLayer)
```

#### Layer Hierarchy
Order layers by specificity:

```go
// Most specific first, most general last
layers := []layers.ParameterLayer{
    applicationSpecificLayer,  // App-specific parameters
    databaseLayer,            // Domain-specific parameters
    loggingLayer,             // Cross-cutting concerns
    debugLayer,               // Development/debugging
}
```

### 4. Settings Structure Design

#### Mirror Layer Structure
Create settings structs that mirror your layers:

```go
// Database layer parameters
type DatabaseSettings struct {
    Host     string `glazed.parameter:"host"`
    Port     int    `glazed.parameter:"port"`
    Database string `glazed.parameter:"database"`
    Username string `glazed.parameter:"username"`
    Password string `glazed.parameter:"password"`
}

// Separate struct for each layer
type ServerSettings struct {
    Host string `glazed.parameter:"host"`
    Port int    `glazed.parameter:"port"`
}
```

#### Use Validation Tags
Add validation to your settings structs:

```go
type DatabaseSettings struct {
    Host     string `glazed.parameter:"host" validate:"required,hostname"`
    Port     int    `glazed.parameter:"port" validate:"min=1,max=65535"`
    Database string `glazed.parameter:"database" validate:"required,min=1"`
    Username string `glazed.parameter:"username" validate:"required"`
    Password string `glazed.parameter:"password" validate:"min=8"`
}

// Validate after parsing
func validateSettings(settings *DatabaseSettings) error {
    validate := validator.New()
    return validate.Struct(settings)
}
```

### 5. Error Handling

#### Graceful Degradation
Handle missing optional layers gracefully:

```go
func (c *MyCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    // Required layer
    serverSettings := &ServerSettings{}
    if err := parsedLayers.InitializeStruct("server", serverSettings); err != nil {
        return fmt.Errorf("server configuration required: %w", err)
    }
    
    // Optional layer
    var monitoringSettings *MonitoringSettings
    if parsedLayer, exists := parsedLayers.Get("monitoring"); exists {
        monitoringSettings = &MonitoringSettings{}
        if err := parsedLayer.InitializeStruct(monitoringSettings); err != nil {
            log.Printf("Warning: invalid monitoring configuration: %v", err)
            monitoringSettings = nil
        }
    }
    
    // Use settings...
    return nil
}
```

#### Clear Error Messages
Provide context in error messages:

```go
if err := parsedLayers.InitializeStruct("database", dbSettings); err != nil {
    return fmt.Errorf("failed to parse database configuration: %w", err)
}
```

### 6. Documentation and Help

#### Layer Descriptions
Always provide helpful descriptions:

```go
layer := layers.NewParameterLayer(
    "database",
    "Database Configuration",
    layers.WithDescription("PostgreSQL database connection settings"),
    // ...
)
```

#### Parameter Help Text
Document each parameter:

```go
parameters.NewParameterDefinition("connection-timeout",
    parameters.ParameterTypeDuration,
    parameters.WithDefault("10s"),
    parameters.WithHelp("Maximum time to wait for database connection"),
)
```

## Testing Layers

### Unit Testing Individual Layers

```go
func TestDatabaseLayerDefaults(t *testing.T) {
    layer, err := NewDatabaseLayer()
    require.NoError(t, err)
    
    // Test parameter definitions
    definitions := layer.GetParameterDefinitions()
    
    hostParam, exists := definitions.Get("host")
    assert.True(t, exists)
    assert.Equal(t, "localhost", hostParam.Default)
    
    portParam, exists := definitions.Get("port")
    assert.True(t, exists)
    assert.Equal(t, 5432, portParam.Default)
}

func TestDatabaseLayerValidation(t *testing.T) {
    layer, err := NewDatabaseLayer()
    require.NoError(t, err)
    
    // Test validation with invalid values
    parsedParams, err := layer.GatherParametersFromMap(map[string]interface{}{
        "host": "",           // Invalid: empty host
        "port": -1,          // Invalid: negative port
        "database": "",      // Invalid: empty database name
    }, false)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "host")
}

func TestDatabaseLayerParsing(t *testing.T) {
    layer, err := NewDatabaseLayer()
    require.NoError(t, err)
    
    // Test parsing valid values
    parsedParams, err := layer.GatherParametersFromMap(map[string]interface{}{
        "host":     "postgres.example.com",
        "port":     5433,
        "database": "production",
        "username": "myuser",
        "password": "secret123",
    }, false)
    
    require.NoError(t, err)
    
    // Verify parsed values
    host, exists := parsedParams.Get("host")
    assert.True(t, exists)
    assert.Equal(t, "postgres.example.com", host.Value)
    
    port, exists := parsedParams.Get("port")
    assert.True(t, exists)
    assert.Equal(t, 5433, port.Value)
}
```

### Integration Testing with Middlewares

```go
func TestDatabaseLayerWithMiddlewares(t *testing.T) {
    // Setup layer
    databaseLayer, err := NewDatabaseLayer()
    require.NoError(t, err)
    
    parameterLayers := layers.NewParameterLayers(
        layers.WithLayers(databaseLayer),
    )
    
    // Setup environment variables
    os.Setenv("MYAPP_DB_HOST", "env-host")
    os.Setenv("MYAPP_DB_PORT", "5433")
    defer func() {
        os.Unsetenv("MYAPP_DB_HOST")
        os.Unsetenv("MYAPP_DB_PORT")
    }()
    
    // Create middleware chain
    middlewares := []middlewares.Middleware{
        middlewares.UpdateFromEnv("MYAPP", 
            parameters.WithParseStepSource("env")),
        middlewares.SetFromDefaults(
            parameters.WithParseStepSource("defaults")),
    }
    
    // Execute middlewares
    parsedLayers := layers.NewParsedLayers()
    err = middlewares.ExecuteMiddlewares(parameterLayers, parsedLayers, middlewares...)
    require.NoError(t, err)
    
    // Extract and verify settings
    settings := &DatabaseSettings{}
    err = parsedLayers.InitializeStruct("database", settings)
    require.NoError(t, err)
    
    assert.Equal(t, "env-host", settings.Host)  // From environment
    assert.Equal(t, 5433, settings.Port)       // From environment
    assert.Equal(t, "myapp", settings.Database) // From defaults
}
```

### Testing Layer Combinations

```go
func TestMultipleLayersTogether(t *testing.T) {
    // Create multiple layers
    serverLayer, err := NewServerLayer()
    require.NoError(t, err)
    
    databaseLayer, err := NewDatabaseLayer()
    require.NoError(t, err)
    
    loggingLayer, err := NewLoggingLayer()
    require.NoError(t, err)
    
    parameterLayers := layers.NewParameterLayers(
        layers.WithLayers(serverLayer, databaseLayer, loggingLayer),
    )
    
    // Test that all layers are present
    assert.Equal(t, 3, parameterLayers.Len())
    
    // Test that we can access each layer
    server, exists := parameterLayers.Get("server")
    assert.True(t, exists)
    assert.Equal(t, "server", server.GetSlug())
    
    database, exists := parameterLayers.Get("database")
    assert.True(t, exists)
    assert.Equal(t, "database", database.GetSlug())
    
    logging, exists := parameterLayers.Get("logging")
    assert.True(t, exists)
    assert.Equal(t, "logging", logging.GetSlug())
    
    // Test that parameter definitions don't conflict
    allDefs := parameterLayers.GetAllParameterDefinitions()
    
    // Should have both server.host and database.host (with prefix)
    _, hasServerHost := allDefs.Get("host")         // server layer
    _, hasDatabaseHost := allDefs.Get("db-host")    // database layer with prefix
    
    assert.True(t, hasServerHost)
    assert.True(t, hasDatabaseHost)
}
```

### Mocking and Test Helpers

```go
// Test helper to create minimal layers for testing
func NewTestDatabaseLayer(overrides map[string]interface{}) (layers.ParameterLayer, error) {
    layer, err := NewDatabaseLayer()
    if err != nil {
        return nil, err
    }
    
    // Apply test overrides
    for param, value := range overrides {
        if paramDef, exists := layer.GetParameterDefinitions().Get(param); exists {
            paramDef.Default = value
        }
    }
    
    return layer, nil
}

// Mock layer for testing
type MockParameterLayer struct {
    slug        string
    definitions *parameters.ParameterDefinitions
}

func NewMockLayer(slug string, params map[string]interface{}) *MockParameterLayer {
    definitions := parameters.NewParameterDefinitions()
    
    for name, defaultValue := range params {
        param := parameters.NewParameterDefinition(name,
            parameters.ParameterTypeString, // Simplified for testing
            parameters.WithDefault(defaultValue),
        )
        definitions.Set(name, param)
    }
    
    return &MockParameterLayer{
        slug:        slug,
        definitions: definitions,
    }
}

func (m *MockParameterLayer) GetSlug() string { return m.slug }
func (m *MockParameterLayer) GetName() string { return m.slug }
func (m *MockParameterLayer) GetDescription() string { return "Mock layer" }
func (m *MockParameterLayer) GetPrefix() string { return "" }
func (m *MockParameterLayer) GetParameterDefinitions() *parameters.ParameterDefinitions {
    return m.definitions
}
// ... implement other ParameterLayer methods
```

### Integration Testing with CLI Flags

Test your layers using the built-in debugging flags:

```go
func TestLayerCLIIntegration(t *testing.T) {
    // Create a command with your layers
    cmd := NewMyCommand()
    
    // Build Cobra command
    cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
    require.NoError(t, err)
    
    // Test with --print-parsed-parameters to verify layer structure
    output := captureOutput(func() {
        cobraCmd.SetArgs([]string{"--host", "test-host", "--print-parsed-parameters"})
        err := cobraCmd.Execute()
        require.NoError(t, err)
    })
    
    // Verify layer appears in output
    assert.Contains(t, output, "# server:")
    assert.Contains(t, output, "host: (string)")
    assert.Contains(t, output, "value: 'test-host'")
    assert.Contains(t, output, "source: cobra: test-host")
}

func captureOutput(f func()) string {
    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    
    f()
    
    w.Close()
    os.Stdout = old
    
    var buf bytes.Buffer
    buf.ReadFrom(r)
    return buf.String()
}
```

## Conclusion

Glazed's layer system transforms CLI parameter management from a chaotic scattered approach into an organized, modular architecture. By understanding the two-phase design (definition vs. runtime), leveraging the layer abstraction, and following best practices, you can build CLI applications that are:

**Maintainable**: Parameters are organized into logical groups that are easy to understand and modify.

**Reusable**: Common parameter groups (database, logging, authentication) can be shared across multiple commands.

**Flexible**: Commands can mix and match only the parameter layers they need.

**Testable**: Each layer can be tested in isolation, and combinations can be verified systematically.

**User-friendly**: Help text is organized by functionality, making it easier for users to discover relevant options.

Key takeaways:
- **Design layers around cohesive concepts**: Database configuration, logging settings, server parameters
- **Use prefixes to avoid naming conflicts**: `--host` vs `--db-host`
- **Leverage middleware chains for flexible parameter processing**: CLI args → config files → environment → defaults
- **Create dedicated settings structs for type-safe parameter access**
- **Test layers individually and in combination**
- **Use built-in CLI tools for debugging**: `--print-parsed-parameters`, `--print-schema`, `--long-help`
- **Document layers and parameters thoroughly**

This system enables you to build sophisticated CLI applications that scale gracefully as your parameter requirements grow, while maintaining clean separation of concerns and excellent user experience.
