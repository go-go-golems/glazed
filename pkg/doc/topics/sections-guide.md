---
Title: Glazed Command Sections Guide
Slug: sections-guide  
Short: Complete guide to understanding and working with command field sections in Glazed
Topics:
- sections
- fields
- configuration
- organization
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Command Sections: Complete Guide

*Building maintainable CLI applications through modular field organization*

## Table of Contents
1. [Overview](#overview)
2. [Field Organization Challenges](#field-organization-challenges)
3. [Section System Architecture](#section-system-architecture)
4. [Core Section Concepts](#core-section-concepts)
5. [Section Types & Components](#section-types--components)
6. [Creating and Working with Sections](#creating-and-working-with-sections)
7. [Practical Examples](#practical-examples)
8. [Advanced Patterns](#advanced-patterns)
9. [Best Practices](#best-practices)
10. [Testing Sections](#testing-sections)

## Overview

Field sections organize related command fields into reusable groups. This modular approach addresses common CLI development challenges including field proliferation, code duplication, naming conflicts, and maintenance complexity in growing applications.

The section system enables developers to:

- **Organize fields** logically by functionality (database, logging, output)
- **Reuse field definitions** across multiple commands without duplication
- **Compose command interfaces** by combining only required functionality
- **Maintain consistency** through centralized field definitions
- **Scale applications** without accumulating technical debt

### Traditional Field Management Problems

CLI applications often start simple but accumulate complexity as they grow:

```bash
myapp server --host localhost --port 8080 --db-host postgres --db-port 5432 --log-level info
myapp client --endpoint http://localhost:8080 --timeout 30s --log-level debug  
myapp backup --db-host postgres --db-port 5432 --output backup.sql --log-level warn
```

This traditional approach creates several maintenance issues:

- **Field pollution**: Commands inherit irrelevant flags, cluttering help screens
- **Naming conflicts**: Ambiguous flag names require awkward prefixes (`--web-host` vs `--db-host`)
- **Code duplication**: Field definitions copy across commands, violating DRY principles
- **Poor organization**: Fields scatter without logical grouping
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

### Glazed Section Solution

The section system eliminates these problems by treating fields as modular components that can be shared, composed, reused, maintained centrally, and extended without breaking existing commands.

```go
// Define reusable sections once
var (
    DatabaseSection = CreateDatabaseSection()  // --db-host, --db-port, --db-name
    LoggingSection  = CreateLoggingSection()   // --log-level, --log-file, --verbose
    ServerSection   = CreateServerSection()    // --host, --port, --timeout
    GlazedSection   = CreateGlazedSection()    // --output, --fields, --format
)

// Compose commands by combining relevant sections
ServerCommand := cmds.NewCommandDescription("server",
    cmds.WithSections(ServerSection, DatabaseSection, LoggingSection))

ClientCommand := cmds.NewCommandDescription("client", 
    cmds.WithSections(ServerSection, LoggingSection))  // No database needed

BackupCommand := cmds.NewCommandDescription("backup",
    cmds.WithSections(DatabaseSection, LoggingSection, GlazedSection))  // No server needed
```

## Field Organization Challenges

Complex CLI applications face predictable field management challenges that sections directly address.

### Field Explosion

As applications add features, commands accumulate fields that may not be relevant to their specific functionality. This creates cognitive overhead for users and increases implementation complexity.

### Naming Conflicts and Namespace Issues

Multiple subsystems often require similar fields (host, port, timeout). Without organization, developers resort to verbose prefixes that reduce usability and create inconsistent interfaces.

### Code Duplication and Maintenance

Field definitions scattered across command implementations require manual synchronization. Adding SSL configuration to database connections means updating every command that uses databases.

### Inconsistent User Interfaces

Without centralized field definitions, similar functionality develops different flag names across commands, creating inconsistent user experiences and requiring additional documentation.

## Section System Architecture

The section system separates field definition from runtime value resolution, enabling flexible composition while maintaining type safety.

### System Components

```
┌─────────────────────────────────────────────┐
│                Command                       │ ← CLI command implementation
│  ┌─────────────────────────────────────────┤
│  │            Command Description           │ ← Metadata + section references
│  │                                         │
│  │  ┌─────────────────────────────────────┤
│  │  │          Field Sections           │ ← Section definitions (design time)
│  │  │                                     │
│  │  │  ┌─────────────┬─────────────────┤   │
│  │  │  │Default Section│  Custom Sections  │   │ ← Different types of sections
│  │  │  │(flags/args) │  (database,    │   │
│  │  │  │             │   logging, etc.)│   │
│  │  │  └─────────────┴─────────────────┘   │
│  │  └─────────────────────────────────────┘ │
│  └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
                       │
                       ▼ (at runtime)
┌─────────────────────────────────────────────┐
│              Runtime Parsing                 │ ← Field resolution
│                                             │
│  Field Sources (in priority order):     │ ← Multi-source configuration
│  1. Command line arguments                  │
│  2. Environment variables                   │
│  3. Configuration files                     │
│  4. Default values from section definitions   │
└─────────────────────────────────────────────┘
                       │
                       ▼ (final result)
┌─────────────────────────────────────────────┐
│               Parsed Sections                  │ ← Runtime values
│        (type-safe field values)         │
└─────────────────────────────────────────────┘
```

This architecture provides clear separation between:

1. **Section Definitions** (design time): Field specifications and constraints
2. **Runtime Parsing** (execution time): Value resolution from multiple sources
3. **Parsed Sections** (application time): Type-safe access to resolved values

### Section Lifecycle

Sections progress through distinct phases from definition to runtime use:

1. **Definition Phase**: Fields defined with types, defaults, and validation rules
2. **Composition Phase**: Sections combined into command descriptions
3. **Parsing Phase**: User input resolved against definitions using source priority
4. **Execution Phase**: Commands access type-safe field values

## Core Section Concepts

### Section Identity and Organization

Every section requires unique identification and logical organization within the application field namespace.

#### Section Hierarchy

```
CommandDescription
├── Default Section (always present)
│   ├── Command-specific flags      ← Unique to this command
│   └── Command-specific arguments  ← Command's core functionality
├── Standard Sections (optional, commonly used)
│   ├── Glazed Section (for structured output)  ← Output formatting
│   ├── Logging Section (for logging configuration)  ← Debug & monitoring
│   └── Database Section (for database connections)  ← Data persistence
└── Custom Sections (application-specific)
    ├── Authentication Section        ← Security & access control
    ├── API Configuration Section     ← External service integration
    └── Feature Toggle Section        ← Experimental or optional features
```

This hierarchy enables logical grouping, reusability across commands, extensibility without breaking changes, and automatic propagation of section updates.

### Field Definitions vs. Parsed Values

The system distinguishes between field specifications (what's possible) and runtime values (what's actual).

**Field Definitions** (specifications stored in sections):
```go
// This defines what's POSSIBLE
fields.New(
    "log-level",                    // Field name
    fields.TypeChoice, // Data type constraint
    fields.WithDefault("info"), // Default value
    fields.WithChoices("debug", "info", "warn", "error"), // Valid options
    fields.WithHelp("Set the logging level"), // User guidance
)
```

**Parsed Values** (actual runtime values):
```go
// After user runs: myapp --log-level debug
parsedValue := "debug"  // Actual value used by application

// Application accesses the final value:
loggingSection, ok := parsedSections.Get("logging")
if ok {
    logLevel, ok := loggingSection.GetField("log-level")
    // logLevel is now "debug" (not "info" default)
}
```

This separation provides:
- **Reusable definitions** across commands
- **Context-specific values** reflecting user choices
- **Automatic validation** based on definition constraints
- **Type safety** from definition through parsed values

### Section Composition Patterns

Section composition determines command interfaces by selecting appropriate field groups.

#### Basic Composition

Simple addition of required functionality:

```go
// Combine only needed sections
commandDesc := cmds.NewCommandDescription("mycommand",
    cmds.WithSections(databaseSection, loggingSection))
```

#### Conditional Composition

Dynamic section assembly based on runtime conditions:

```go
// Build section list based on features
sections := []schema.Section{baseSection}

if needsDatabase {
    sections = append(sections, databaseSection)
}

if needsOutput {
    sections = append(sections, glazedSection)
}

commandDesc := cmds.NewCommandDescription("mycommand",
    cmds.WithSectionsList(sections...))
```

#### Section Extension

Building specialized variants from existing sections:

```go
// Extend existing section without modification
extendedDbSection := databaseSection.Clone()
extendedDbSection.AddFields(
    fields.New("pool-size", fields.TypeInteger),
    fields.New("connection-timeout", fields.TypeString),
)
```

## Section Types & Components

### Built-in Section Types

Glazed provides standard section types for common CLI application requirements.

#### Default Section

The Default Section contains command-specific fields unique to individual commands.

- **Purpose**: Command-specific flags and arguments defining core functionality
- **Slug**: `"default"` (constant: `schema.DefaultSlug`) 
- **Creation**: Automatically created with `cmds.WithFlags()` or `cmds.WithArguments()`
- **Use Case**: Fields fundamental to command operation, unlikely to be shared

```go
// Default section created automatically
commandDesc := cmds.NewCommandDescription("serve",
    cmds.WithFlags(
        fields.New("port", fields.TypeInteger),
        fields.New("host", fields.TypeString),
    ),
    cmds.WithArguments(
        fields.New("config-file", fields.TypeString),
    ),
)
// Fields live in default section, unique to "serve" command
```

#### Glazed Section

The Glazed Section provides comprehensive output formatting capabilities for commands producing structured data.

- **Purpose**: Output formatting, filtering, and transformation options
- **Slug**: `"glazed"` (constant: `settings.GlazedSlug`)
- **Creation**: `settings.NewGlazedSchema()`
- **Use Case**: Commands outputting structured data requiring flexible formatting

```go
glazedSection, err := settings.NewGlazedSchema()
if err != nil {
    return nil, fmt.Errorf("failed to create glazed section: %w", err)
}
// Provides: --output, --fields, --sort-columns, --filter, etc.
// Users can run: myapp data --output json --fields name,age --filter "age > 25"
```

### Custom Section Components

Application-specific sections address domain requirements through custom field definitions.

#### Field Definitions

Individual field specifications define acceptable input and behavior:

```go
paramDef := fields.New(
    "connection-timeout",              // Field name
    fields.TypeString,    // Type constraint (use string for duration parsing)
    fields.WithDefault("30s"),     // Default value
    fields.WithHelp("Connection timeout for database operations"), // User guidance
    fields.WithRequired(false),    // Whether required
    fields.WithShortFlag("t"),     // Short flag convenience
)
```

Each definition specifies:
- **Name**: Field identifier for CLI flags
- **Type**: Data type with automatic validation
- **Default**: Value used when not specified
- **Help**: User guidance for help text
- **Constraints**: Validation rules and requirements

#### Validation and Constraints

Field definitions include built-in validation for common patterns:

```go
// Choice fields with automatic validation
fields.New(
    "log-level",
    fields.TypeChoice,
    fields.WithChoices("debug", "info", "warn", "error", "fatal"),
    fields.WithDefault("info"),
)
// Automatically rejects invalid choices and shows valid options

// Numeric fields with defaults
fields.New(
    "retry-count",
    fields.TypeInteger,
    fields.WithDefault(3),
)
```

This validation approach catches user errors early and provides helpful feedback, improving CLI usability and reducing support requirements.

## Creating and Working with Sections

### Method 1: Simple Section Creation

Direct section creation for straightforward field grouping:

```go
func NewDatabaseSection() (schema.Section, error) {
    return schema.NewSection(
        "database",                    // Section identifier
        "Database Configuration",      // Human-readable name
        schema.WithFields(
            fields.New(
                "db-host",
                fields.TypeString,
                fields.WithDefault("localhost"),
                fields.WithHelp("Database host to connect to"),
            ),
            fields.New(
                "db-port",
                fields.TypeInteger,
                fields.WithDefault(5432),
                fields.WithHelp("Database port (PostgreSQL default: 5432)"),
            ),
            fields.New(
                "db-name",
                fields.TypeString,
                fields.WithHelp("Database name (required for connection)"),
                fields.WithRequired(true),
            ),
        ),
    )
}
```

### Method 2: Type-Safe Section with Settings Struct

For complex sections requiring type safety and structured access:

```go
// 1. Define settings struct
type DatabaseSettings struct {
    Host     string `glazed:"db-host"`
    Port     int    `glazed:"db-port"`
    Name     string `glazed:"db-name"`
    Username string `glazed:"db-username"`
    Password string `glazed:"db-password"`
    SSLMode  string `glazed:"db-ssl-mode"`
}

// 2. Create section with field definitions
func NewDatabaseSection() (schema.Section, error) {
    return schema.NewSection(
        "database",
        "Database Configuration",
        schema.WithFields(
            fields.New(
                "db-host",
                fields.TypeString,
                fields.WithDefault("localhost"),
                fields.WithHelp("Database host"),
            ),
            fields.New(
                "db-port", 
                fields.TypeInteger,
                fields.WithDefault(5432),
                fields.WithHelp("Database port"),
            ),
            fields.New(
                "db-name",
                fields.TypeString,
                fields.WithHelp("Database name"),
                fields.WithRequired(true),
            ),
            fields.New(
                "db-username",
                fields.TypeString,
                fields.WithHelp("Database username"),
            ),
            fields.New(
                "db-password",
                fields.TypeSecret,  // Masked in output
                fields.WithHelp("Database password"),
            ),
            fields.New(
                "db-ssl-mode",
                fields.TypeChoice,
                fields.WithChoices("disable", "require", "verify-ca", "verify-full"),
                fields.WithDefault("require"),
                fields.WithHelp("SSL mode for database connection"),
            ),
        ),
    )
}

// 3. Helper function for settings extraction
func GetDatabaseSettings(parsedSections *values.Values) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedSections.DecodeSectionInto("database", settings)
    return settings, err
}

// 4. Usage in command implementation
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedSections *values.Values,
    gp middlewares.Processor,
) error {
    // Extract database settings from the "database" section
    dbSettings, err := GetDatabaseSettings(parsedSections)
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

### Method 3: Section Builder Pattern

For complex scenarios requiring conditional fields:

```go
type DatabaseSectionBuilder struct {
    section      schema.Section
    includeSSL bool
    includePool bool
}

func NewDatabaseSectionBuilder() *DatabaseSectionBuilder {
    section, _ := schema.NewSection("database", "Database Configuration")
    return &DatabaseSectionBuilder{section: section}
}

func (b *DatabaseSectionBuilder) WithSSL() *DatabaseSectionBuilder {
    b.includeSSL = true
    return b
}

func (b *DatabaseSectionBuilder) WithConnectionPool() *DatabaseSectionBuilder {
    b.includePool = true
    return b
}

func (b *DatabaseSectionBuilder) Build() (schema.Section, error) {
    // Add basic fields
    b.section.AddFields(
        fields.New("db-host", fields.TypeString, 
            fields.WithDefault("localhost")),
        fields.New("db-port", fields.TypeInteger,
            fields.WithDefault(5432)),
    )
    
    // Conditionally add SSL fields
    if b.includeSSL {
        b.section.AddFields(
            fields.New("db-ssl-mode", fields.TypeChoice,
                fields.WithChoices("disable", "require", "verify-ca")),
            fields.New("db-ssl-cert", fields.TypeFile),
        )
    }
    
    // Conditionally add connection pool fields
    if b.includePool {
        b.section.AddFields(
            fields.New("db-max-connections", fields.TypeInteger,
                fields.WithDefault(10)),
            fields.New("db-idle-timeout", fields.TypeString,
                fields.WithDefault("5m")),
        )
    }
    
    return b.section, nil
}

// Usage:
dbSection, _ := NewDatabaseSectionBuilder().
    WithSSL().
    WithConnectionPool().
    Build()
```

### Method 4: Registering Sections with Explicit Slugs on Commands

In some cases you may want to register sections on a command under explicit slugs that differ from the section's internal slug. Use `cmds.WithSectionsMap` to provide a map of slug-to-section entries when creating a command description.

```go
dbSection, _ := NewDatabaseSection()        // internal slug: "database"
loggingSection, _ := NewLoggingSection()    // internal slug: "logging"

cmd := cmds.NewCommandDescription(
    "process",
    cmds.WithSectionsMap(map[string]schema.Section{
        "db":  dbSection,    // registered under explicit slug "db"
        "log": loggingSection,
    }),
)

// Later at runtime, access by the registration slugs (e.g., "db", "log")
```

Note:
- If a section's internal slug differs from the map key and the section is a `*schema.SectionImpl`, Glazed will clone the section and align its slug to the provided key for consistent runtime behavior.
- For custom section implementations, prefer using matching internal and registration slugs when possible.

## Practical Examples

### Example 1: Web Server Application

Complete section implementation for a web server with database, logging, and server configuration:

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/schema"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
)

// Settings structs for type safety
type ServerSettings struct {
    Host         string `glazed:"host"`
    Port         int    `glazed:"port"`
    ReadTimeout  string `glazed:"read-timeout"`
    WriteTimeout string `glazed:"write-timeout"`
}

type LoggingSettings struct {
    Level  string `glazed:"log-level"`
    Format string `glazed:"log-format"`
    File   string `glazed:"log-file"`
}

type DatabaseSettings struct {
    Host     string `glazed:"db-host"`
    Port     int    `glazed:"db-port"`
    Name     string `glazed:"db-name"`
    Username string `glazed:"db-username"`
    Password string `glazed:"db-password"`
}

// Section creation functions
func NewServerSection() (schema.Section, error) {
    return schema.NewSection(
        "server",
        "Web Server Configuration",
        schema.WithFields(
            fields.New(
                "host",
                fields.TypeString,
                fields.WithDefault("localhost"),
                fields.WithHelp("Server host to bind to"),
                fields.WithShortFlag("H"),
            ),
            fields.New(
                "port",
                fields.TypeInteger,
                fields.WithDefault(8080),
                fields.WithHelp("Server port to listen on"),
                fields.WithShortFlag("p"),
            ),
            fields.New(
                "read-timeout",
                fields.TypeString,
                fields.WithDefault("30s"),
                fields.WithHelp("HTTP read timeout"),
            ),
            fields.New(
                "write-timeout",
                fields.TypeString,
                fields.WithDefault("30s"),
                fields.WithHelp("HTTP write timeout"),
            ),
        ),
    )
}

func NewLoggingSection() (schema.Section, error) {
    return schema.NewSection(
        "logging",
        "Logging Configuration",
        schema.WithFields(
            fields.New(
                "log-level",
                fields.TypeChoice,
                fields.WithChoices("debug", "info", "warn", "error", "fatal"),
                fields.WithDefault("info"),
                fields.WithHelp("Logging level"),
            ),
            fields.New(
                "log-format",
                fields.TypeChoice,
                fields.WithChoices("text", "json"),
                fields.WithDefault("text"),
                fields.WithHelp("Log output format"),
            ),
            fields.New(
                "log-file",
                fields.TypeString,
                fields.WithHelp("Log file path (default: stderr)"),
            ),
        ),
    )
}

func NewDatabaseSection() (schema.Section, error) {
    return schema.NewSection(
        "database",
        "Database Configuration",
        schema.WithFields(
            fields.New(
                "db-host",
                fields.TypeString,
                fields.WithDefault("localhost"),
                fields.WithHelp("Database host"),
            ),
            fields.New(
                "db-port",
                fields.TypeInteger,
                fields.WithDefault(5432),
                fields.WithHelp("Database port"),
            ),
            fields.New(
                "db-name",
                fields.TypeString,
                fields.WithHelp("Database name"),
                fields.WithRequired(true),
            ),
            fields.New(
                "db-username",
                fields.TypeString,
                fields.WithHelp("Database username"),
            ),
            fields.New(
                "db-password",
                fields.TypeSecret,
                fields.WithHelp("Database password"),
            ),
        ),
    )
}

// Command creation with section composition
func NewServerCommand() (*cmds.CommandDescription, error) {
    // Create sections
    serverSection, err := NewServerSection()
    if err != nil {
        return nil, err
    }
    
    loggingSection, err := NewLoggingSection()
    if err != nil {
        return nil, err
    }
    
    databaseSection, err := NewDatabaseSection()
    if err != nil {
        return nil, err
    }
    
    // Compose command with relevant sections
    return cmds.NewCommandDescription(
        "serve",
        cmds.WithShort("Start the web server"),
        cmds.WithLong("Start the web server with the specified configuration"),
        cmds.WithSectionsList(serverSection, databaseSection, loggingSection),
    ), nil
}

func NewHealthCheckCommand() (*cmds.CommandDescription, error) {
    // Health check only needs server configuration, not database
    serverSection, err := NewServerSection()
    if err != nil {
        return nil, err
    }
    
    loggingSection, err := NewLoggingSection()
    if err != nil {
        return nil, err
    }
    
    return cmds.NewCommandDescription(
        "health",
        cmds.WithShort("Check server health"),
        cmds.WithFlags(
            fields.New(
                "endpoint",
                fields.TypeString,
                fields.WithDefault("/health"),
                fields.WithHelp("Health check endpoint"),
            ),
        ),
        cmds.WithSectionsList(serverSection, loggingSection), // No database section
    ), nil
}

// Settings extraction helpers demonstrate how to use DecodeSectionInto with section-specific settings
func GetServerSettings(parsedSections *values.Values) (*ServerSettings, error) {
    settings := &ServerSettings{}
    err := parsedSections.DecodeSectionInto("server", settings)
    return settings, err
}

func GetLoggingSettings(parsedSections *values.Values) (*LoggingSettings, error) {
    settings := &LoggingSettings{}
    err := parsedSections.DecodeSectionInto("logging", settings)
    return settings, err
}

func GetDatabaseSettings(parsedSections *values.Values) (*DatabaseSettings, error) {
    settings := &DatabaseSettings{}
    err := parsedSections.DecodeSectionInto("database", settings)
    return settings, err
}

// Example command implementation using multiple section settings
type ServerCommand struct {
    *cmds.CommandDescription
}

func (c *ServerCommand) Run(ctx context.Context, parsedSections *values.Values) error {
    // Extract settings from each section
    serverSettings, err := GetServerSettings(parsedSections)
    if err != nil {
        return fmt.Errorf("failed to get server settings: %w", err)
    }
    
    dbSettings, err := GetDatabaseSettings(parsedSections)
    if err != nil {
        return fmt.Errorf("failed to get database settings: %w", err)
    }
    
    logSettings, err := GetLoggingSettings(parsedSections)
    if err != nil {
        return fmt.Errorf("failed to get logging settings: %w", err)
    }
    
    // Use settings from multiple sections
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

Section composition for applications with conditional functionality. This example shows how to extract settings from optional sections and use them together:

```go
// Feature sections for optional inclusion
func NewCacheSection() (schema.Section, error) {
    return schema.NewSection(
        "cache",
        "Caching Configuration",
        schema.WithFields(
            fields.New(
                "cache-enabled",
                fields.TypeBool,
                fields.WithDefault(true),
                fields.WithHelp("Enable caching"),
            ),
            fields.New(
                "cache-ttl",
                fields.TypeString,
                fields.WithDefault("1h"),
                fields.WithHelp("Cache time-to-live"),
            ),
            fields.New(
                "cache-size",
                fields.TypeInteger,
                fields.WithDefault(1000),
                fields.WithHelp("Maximum cache entries"),
            ),
        ),
    )
}

func NewMetricsSection() (schema.Section, error) {
    return schema.NewSection(
        "metrics",
        "Metrics and Monitoring",
        schema.WithFields(
            fields.New(
                "metrics-enabled",
                fields.TypeBool,
                fields.WithDefault(false),
                fields.WithHelp("Enable metrics collection"),
            ),
            fields.New(
                "metrics-port",
                fields.TypeInteger,
                fields.WithDefault(9090),
                fields.WithHelp("Metrics server port"),
            ),
            fields.New(
                "metrics-path",
                fields.TypeString,
                fields.WithDefault("/metrics"),
                fields.WithHelp("Metrics endpoint path"),
            ),
        ),
    )
}

// Command builder with optional features
type AppCommandBuilder struct {
    baseSections    []schema.Section
    enableCache   bool
    enableMetrics bool
    enableAuth    bool
}

func NewAppCommandBuilder() *AppCommandBuilder {
    loggingSection, _ := NewLoggingSection()
    
    return &AppCommandBuilder{
        baseSections: []schema.Section{loggingSection},
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
    commandSections := append([]schema.Section{}, b.baseSections...)
    
    // Add optional sections based on enabled features
    if b.enableCache {
        cacheSection, err := NewCacheSection()
        if err != nil {
            return nil, err
        }
        commandSections = append(commandSections, cacheSection)
    }
    
    if b.enableMetrics {
        metricsSection, err := NewMetricsSection()
        if err != nil {
            return nil, err
        }
        commandSections = append(commandSections, metricsSection)
    }
    
    if b.enableAuth {
        authSection, err := NewAuthSection() // Assume this exists
        if err != nil {
            return nil, err
        }
        commandSections = append(commandSections, authSection)
    }
    
    return cmds.NewCommandDescription(
        "process",
        cmds.WithShort("Process data with optional features"),
        cmds.WithFlags(
            fields.New(
                "input-file",
                fields.TypeFile,
                fields.WithHelp("Input file to process"),
                fields.WithRequired(true),
            ),
            fields.New(
                "output-file", 
                fields.TypeString,
                fields.WithHelp("Output file path"),
            ),
        ),
        cmds.WithSectionsList(commandSections...),
    ), nil
}

// Settings structs for optional features
type CacheSettings struct {
    Enabled bool   `glazed:"cache-enabled"`
    TTL     string `glazed:"cache-ttl"`
    Size    int    `glazed:"cache-size"`
}

type MetricsSettings struct {
    Enabled bool   `glazed:"metrics-enabled"`
    Port    int    `glazed:"metrics-port"`
    Path    string `glazed:"metrics-path"`
}

// Helper functions for optional section settings
func GetCacheSettings(parsedSections *values.Values) (*CacheSettings, error) {
    settings := &CacheSettings{}
    err := parsedSections.DecodeSectionInto("cache", settings)
    return settings, err
}

func GetMetricsSettings(parsedSections *values.Values) (*MetricsSettings, error) {
    settings := &MetricsSettings{}
    err := parsedSections.DecodeSectionInto("metrics", settings)
    return settings, err
}

// Command implementation that handles optional sections
type ProcessCommand struct {
    *cmds.CommandDescription
}

func (c *ProcessCommand) Run(ctx context.Context, parsedSections *values.Values) error {
    // Always extract logging settings
    logSettings, err := GetLoggingSettings(parsedSections)
    if err != nil {
        return fmt.Errorf("failed to get logging settings: %w", err)
    }
    
    fmt.Printf("Starting processing with log level: %s\n", logSettings.Level)
    
    // Try to extract cache settings (may not exist)
    if parsedSections.Has("cache") {
        cacheSettings, err := GetCacheSettings(parsedSections)
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
    if parsedSections.Has("metrics") {
        metricsSettings, err := GetMetricsSettings(parsedSections)
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

### 1. Section Inheritance and Composition

Extending existing sections without modification:

```go
// Base database section
func NewBaseDatabaseSection() (schema.Section, error) {
    return schema.NewSection(
        "database",
        "Database Configuration",
        schema.WithFields(
            fields.New("db-host", fields.TypeString,
                fields.WithDefault("localhost")),
            fields.New("db-port", fields.TypeInteger,
                fields.WithDefault(5432)),
        ),
    )
}

// Extended database section with additional features
func NewAdvancedDatabaseSection() (schema.Section, error) {
    // Start with base section
    baseSection, err := NewBaseDatabaseSection()
    if err != nil {
        return nil, err
    }
    
    // Clone to avoid modifying the original
    advancedSection := baseSection.Clone()
    
    // Add additional fields
    advancedSection.AddFields(
        fields.New("db-pool-size", fields.TypeInteger,
            fields.WithDefault(10)),
        fields.New("db-ssl-mode", fields.TypeChoice,
            fields.WithChoices("disable", "require", "verify-full")),
        fields.New("db-connection-timeout", fields.TypeDuration,
            fields.WithDefault("30s")),
    )
    
    return advancedSection, nil
}
```

### 2. Environment-Specific Section Configuration

Adapting sections for different deployment environments:

```go
type EnvironmentConfig struct {
    Environment string // "development", "staging", "production"
    Features    []string
}

func NewEnvironmentAwareDatabaseSection(config EnvironmentConfig) (schema.Section, error) {
    section, err := NewBaseDatabaseSection()
    if err != nil {
        return nil, err
    }
    
    // Add environment-specific fields
    switch config.Environment {
    case "development":
        section.AddFields(
            fields.New("db-debug-queries", fields.TypeBool,
                fields.WithDefault(true)),
            fields.New("db-auto-migrate", fields.TypeBool,
                fields.WithDefault(true)),
        )
    case "production":
        section.AddFields(
            fields.New("db-ssl-mode", fields.TypeChoice,
                fields.WithChoices("require", "verify-full"),
                fields.WithDefault("verify-full")),
            fields.New("db-connection-pool-size", fields.TypeInteger,
                fields.WithDefault(50)),
        )
    }
    
    // Add feature-specific fields
    for _, feature := range config.Features {
        switch feature {
        case "monitoring":
            section.AddFields(
                fields.New("db-monitor-slow-queries", fields.TypeBool),
                fields.New("db-slow-query-threshold", fields.TypeDuration,
                    fields.WithDefault("1s")),
            )
        case "backup":
            section.AddFields(
                fields.New("db-backup-enabled", fields.TypeBool),
                fields.New("db-backup-schedule", fields.TypeString),
            )
        }
    }
    
    return section, nil
}
```

### 3. Dynamic Section Registration

Runtime section registration for plugin systems:

```go
type SectionRegistry struct {
    sections map[string]schema.Section
    mutex  sync.RWMutex
}

func NewSectionRegistry() *SectionRegistry {
    return &SectionRegistry{
        sections: make(map[string]schema.Section),
    }
}

func (r *SectionRegistry) RegisterSection(slug string, section schema.Section) error {
    r.mutex.Lock()
    defer r.mutex.Unlock()
    
    if _, exists := r.sections[slug]; exists {
        return fmt.Errorf("section %s already registered", slug)
    }
    
    r.sections[slug] = section
    return nil
}

func (r *SectionRegistry) GetSection(slug string) (schema.Section, error) {
    r.mutex.RLock()
    defer r.mutex.RUnlock()
    
    section, exists := r.sections[slug]
    if !exists {
        return nil, fmt.Errorf("section %s not found", slug)
    }
    
    return section, nil
}

func (r *SectionRegistry) BuildCommand(name string, sectionSlugs []string) (*cmds.CommandDescription, error) {
    var commandSections []schema.Section
    
    for _, slug := range sectionSlugs {
        section, err := r.GetSection(slug)
        if err != nil {
            return nil, err
        }
        commandSections = append(commandSections, section)
    }
    
    return cmds.NewCommandDescription(name,
        cmds.WithSectionsList(commandSections...)), nil
}

// Plugin system usage
func InitializePluginSystem() *SectionRegistry {
    registry := NewSectionRegistry()
    
    // Register core sections
    databaseSection, _ := NewDatabaseSection()
    registry.RegisterSection("database", databaseSection)
    
    loggingSection, _ := NewLoggingSection()
    registry.RegisterSection("logging", loggingSection)
    
    // Plugins can register additional sections
    // registry.RegisterSection("custom-feature", customSection)
    
    return registry
}
```

### 4. Section Validation and Constraints

Complex validation rules across section fields:

```go
type SectionValidator struct {
    rules []ValidationRule
}

type ValidationRule func(*values.Values) error

func NewSectionValidator() *SectionValidator {
    return &SectionValidator{}
}

func (v *SectionValidator) AddRule(rule ValidationRule) {
    v.rules = append(v.rules, rule)
}

func (v *SectionValidator) Validate(parsedSections *values.Values) error {
    for _, rule := range v.rules {
        if err := rule(parsedSections); err != nil {
            return err
        }
    }
    return nil
}

// Cross-section validation rules
func DatabaseConnectionRule(parsedSections *values.Values) error {
    dbSection, ok := parsedSections.Get("database")
    if !ok {
        return nil // Skip if database section not present
    }
    
    host, ok := dbSection.GetField("db-host")
    if !ok {
        return nil
    }
    
    port, ok := dbSection.GetField("db-port")
    if !ok {
        return nil
    }
    
    // Validate connection fields make sense together
    if host == "localhost" && port.(int) < 1024 {
        return fmt.Errorf("localhost connections should use ports >= 1024")
    }
    
    return nil
}

func SSLConfigurationRule(parsedSections *values.Values) error {
    dbSection, ok := parsedSections.Get("database")
    if !ok {
        return nil // Skip if database section not present
    }
    
    sslMode, ok := dbSection.GetField("db-ssl-mode")
    if !ok {
        return nil // Skip if SSL not configured
    }
    
    if sslMode == "verify-full" {
        // Ensure SSL certificate is provided when required
        cert, ok := dbSection.GetField("db-ssl-cert")
        if !ok || cert == "" {
            return fmt.Errorf("SSL certificate required when ssl-mode is verify-full")
        }
    }
    
    return nil
}

// Usage in command implementation
func (c *MyCommand) Run(ctx context.Context, parsedSections *values.Values) error {
    // Validate section configuration
    validator := NewSectionValidator()
    validator.AddRule(DatabaseConnectionRule)
    validator.AddRule(SSLConfigurationRule)
    
    if err := validator.Validate(parsedSections); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }
    
    // Continue with command execution
    return nil
}
```

## Best Practices

### Section Design Principles

**Single Responsibility**: Each section should handle one logical area of configuration. Database sections handle database fields, logging sections handle logging configuration.

**Clear Naming**: Use descriptive section slugs and field names. Prefer `database-connection-timeout` over `timeout`.

**Sensible Defaults**: Provide reasonable default values that work in common scenarios. Users should be able to run commands without extensive configuration.

**Consistent Interfaces**: Use similar field names across sections. If one section uses `host`, avoid `hostname` in another section for the same concept.

### Field Organization

Group related fields logically within sections. Database sections should include host, port, credentials, and connection options together.

Use consistent naming patterns across your application. Establish conventions for common concepts like timeouts, ports, and file paths.

Consider field relationships when designing sections. Fields that are frequently used together belong in the same section.

### Command Composition

Only include sections that provide fields relevant to the command's functionality. Avoid section pollution by being selective.

Use builder patterns for commands with many optional features. This provides flexibility while maintaining clean interfaces.

Consider creating specialized section variants for different command types. A read-only database section might exclude authentication fields.

### Error Handling and Validation

Validate section configuration early in command execution. Fail fast with clear error messages about field issues.

Provide helpful validation messages that guide users toward correct configuration. Include examples of valid values when rejecting input.

Use type-safe field extraction where possible. Struct-based settings reduce runtime errors and improve code clarity.

### Testing and Maintenance

Write unit tests for section definitions to ensure field validation works correctly. Test edge cases and error conditions.

Test section composition to verify that combined sections work correctly together. Check for field conflicts and validation interactions.

Use integration tests to verify that commands work correctly with different section combinations and field sources.

Document section dependencies and relationships. Explain when sections should be used together and any constraints.

Keep section definitions close to their usage when possible. This improves maintainability and reduces the chance of configuration drift.

Version section definitions carefully in evolving applications. Consider backward compatibility when modifying existing sections.

## Testing Sections

### Unit Testing Section Definitions

Test individual section creation and field validation:

```go
func TestDatabaseSection(t *testing.T) {
    section, err := NewDatabaseSection()
    assert.NoError(t, err)
    assert.Equal(t, "database", section.GetSlug())
    
    // Test field definitions
    params := section.GetDefinitions()
    assert.Contains(t, params, "db-host")
    assert.Contains(t, params, "db-port")
    assert.Contains(t, params, "db-name")
    
    // Test default values
    hostParam := params["db-host"]
    assert.Equal(t, "localhost", hostParam.Default)
    
    portParam := params["db-port"]
    assert.Equal(t, 5432, portParam.Default)
}

func TestFieldValidation(t *testing.T) {
    section, _ := NewDatabaseSection()
    
    // Test valid choices
    logLevelParam := section.GetDefinitions()["log-level"]
    validChoices := []string{"debug", "info", "warn", "error", "fatal"}
    assert.Equal(t, validChoices, logLevelParam.Choices)
    
    // Test required fields
    dbNameParam := section.GetDefinitions()["db-name"]
    assert.True(t, dbNameParam.Required)
}
```

### Integration Testing Section Composition

Test command creation with multiple sections:

```go
func TestCommandComposition(t *testing.T) {
    serverSection, _ := NewServerSection()
    databaseSection, _ := NewDatabaseSection()
    loggingSection, _ := NewLoggingSection()
    
    command, err := cmds.NewCommandDescription("test-command",
        cmds.WithSectionsList(serverSection, databaseSection, loggingSection))
    
    assert.NoError(t, err)
    assert.NotNil(t, command)
    
    // Verify all sections are present
    sections := command.GetSections()
    assert.Len(t, sections, 3)
    
    // Verify no field conflicts
    allParams := make(map[string]bool)
    for _, section := range sections {
        for paramName := range section.GetDefinitions() {
            assert.False(t, allParams[paramName], 
                "Field %s defined in multiple sections", paramName)
            allParams[paramName] = true
        }
    }
}
```

### Testing Field Resolution

Test field value resolution from different sources:

```go
func TestFieldResolution(t *testing.T) {
    // Create test command with sections
    command, _ := createTestCommand()
    
    // Test CLI argument parsing
    args := []string{"--db-host", "testhost", "--db-port", "3306", "--log-level", "debug"}
    parsedSections, err := command.ParseSections(args)
    assert.NoError(t, err)
    
    // Verify parsed values
    dbSection, ok := parsedSections.Get("database")
    assert.True(t, ok)
    
    dbHost, ok := dbSection.GetField("db-host")
    assert.True(t, ok)
    assert.Equal(t, "testhost", dbHost)
    
    dbPort, ok := dbSection.GetField("db-port")
    assert.True(t, ok)
    assert.Equal(t, 3306, dbPort)
    
    // Test struct initialization
    dbSettings := &DatabaseSettings{}
    err = parsedSections.DecodeSectionInto("database", dbSettings)
    assert.NoError(t, err)
    assert.Equal(t, "testhost", dbSettings.Host)
    assert.Equal(t, 3306, dbSettings.Port)
}

func TestDefaultValues(t *testing.T) {
    command, _ := createTestCommand()
    
    // Parse with no arguments - should use defaults
    parsedSections, err := command.ParseSections([]string{})
    assert.NoError(t, err)
    
    // Verify default values are used
    dbSection, _ := parsedSections.Get("database")
    dbHost, _ := dbSection.GetField("db-host")
    assert.Equal(t, "localhost", dbHost)
    
    dbPort, _ := dbSection.GetField("db-port")
    assert.Equal(t, 5432, dbPort)
    
    loggingSection, _ := parsedSections.Get("logging")
    logLevel, _ := loggingSection.GetField("log-level")
    assert.Equal(t, "info", logLevel)
}
```

### Testing Section Builders and Dynamic Composition

Test builder patterns and conditional section inclusion:

```go
func TestDatabaseSectionBuilder(t *testing.T) {
    // Test basic section
    basicSection, err := NewDatabaseSectionBuilder().Build()
    assert.NoError(t, err)
    
    basicParams := basicSection.GetDefinitions()
    assert.Contains(t, basicParams, "db-host")
    assert.Contains(t, basicParams, "db-port")
    assert.NotContains(t, basicParams, "db-ssl-mode")
    
    // Test section with SSL
    sslSection, err := NewDatabaseSectionBuilder().WithSSL().Build()
    assert.NoError(t, err)
    
    sslParams := sslSection.GetDefinitions()
    assert.Contains(t, sslParams, "db-host")
    assert.Contains(t, sslParams, "db-ssl-mode")
    assert.Contains(t, sslParams, "db-ssl-cert")
    
    // Test section with connection pool
    poolSection, err := NewDatabaseSectionBuilder().WithConnectionPool().Build()
    assert.NoError(t, err)
    
    poolParams := poolSection.GetDefinitions()
    assert.Contains(t, poolParams, "db-max-connections")
    assert.Contains(t, poolParams, "db-idle-timeout")
}

func TestConditionalSectionComposition(t *testing.T) {
    builder := NewAppCommandBuilder()
    
    // Test basic command
    basicCmd, err := builder.BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, basicCmd.GetSections(), 1) // Only logging section
    
    // Test command with cache
    cacheCmd, err := builder.WithCache().BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, cacheCmd.GetSections(), 2) // Logging + cache sections
    
    // Test command with all features
    fullCmd, err := builder.WithCache().WithMetrics().WithAuth().BuildProcessCommand()
    assert.NoError(t, err)
    assert.Len(t, fullCmd.GetSections(), 4) // All sections
}
```

This comprehensive testing approach ensures sections work correctly individually and in composition, field resolution functions properly across different sources, and dynamic section construction produces expected results.
