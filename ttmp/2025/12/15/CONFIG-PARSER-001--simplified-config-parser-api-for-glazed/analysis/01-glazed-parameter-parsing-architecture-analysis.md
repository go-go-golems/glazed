---
Title: Glazed Parameter Parsing Architecture Analysis
Status: active
Intent: long-term
Topics:
  - glazed
  - config
  - api-design
  - parsing
Owners:
  - manuel
Created: 2025-12-15
---

# Glazed Parameter Parsing Architecture Analysis

## Purpose

This document provides a comprehensive analysis of how Glazed handles parameter definitions, layers, middleware registration, and command-line integration. The goal is to understand all the locations involved in the parameter parsing system to enable the design of a simplified API that maps struct tags directly to configuration sources without requiring manual layer and middleware setup.

## Executive Summary

Glazed's parameter system is built on four core concepts:
1. **ParameterDefinitions**: Typed parameter specifications with validation
2. **ParameterLayers**: Logical groupings of parameters (e.g., "default", "glazed", "logging")
3. **ParsedLayers**: Runtime values after parsing from multiple sources
4. **Middlewares**: Chainable functions that populate ParsedLayers from various sources (CLI flags, env vars, config files, defaults)

The current API requires explicit creation of layers, parameter definitions, and middleware chains. A simplified API would automatically derive these from struct tags.

## Core Components

### 1. Parameter Definitions

**Location**: `glazed/pkg/cmds/parameters/parameters.go`

**Key Types**:
- `ParameterDefinition`: Core struct defining a parameter (name, type, default, help, etc.)
- `ParameterDefinitions`: Ordered map of ParameterDefinition objects
- `ParameterType`: Enum of supported types (String, Integer, Bool, File, Choice, etc.)

**Key Functions**:
- `NewParameterDefinition()`: Creates a parameter definition with options
- `InitializeDefaultsFromStruct()`: Sets parameter defaults from struct fields with `glazed.parameter` tags
- `InitializeDefaultsFromMap()`: Sets parameter defaults from a map

**Key Symbols**:
```go
type ParameterDefinition struct {
    Name       string
    ShortFlag  string
    Type       ParameterType
    Help       string
    Default    *interface{}
    Choices    []string
    Required   bool
    IsArgument bool
}

func NewParameterDefinition(
    name string,
    parameterType ParameterType,
    options ...ParameterDefinitionOption,
) *ParameterDefinition
```

**How It Works**:
- Parameters are defined declaratively with type information
- Types include validation rules (e.g., ParameterTypeInteger validates ranges)
- Defaults can be set from struct fields via `InitializeDefaultsFromStruct()`
- The system supports both flags and positional arguments

### 2. Parameter Layers

**Location**: `glazed/pkg/cmds/layers/layer.go`, `glazed/pkg/cmds/layers/layer-impl.go`

**Key Types**:
- `ParameterLayer`: Interface for grouping parameters
- `ParameterLayerImpl`: Standard implementation
- `ParameterLayers`: Ordered map of ParameterLayer objects
- `CobraParameterLayer`: Interface for layers that integrate with Cobra

**Key Functions**:
- `NewParameterLayer()`: Creates a new layer with options
- `AddLayerToCobraCommand()`: Adds layer flags to a Cobra command
- `ParseLayerFromCobraCommand()`: Parses layer values from Cobra flags
- `InitializeParameterDefaultsFromStruct()`: Sets layer parameter defaults from struct

**Key Symbols**:
```go
type ParameterLayer interface {
    AddFlags(flag ...*parameters.ParameterDefinition)
    GetParameterDefinitions() *parameters.ParameterDefinitions
    InitializeParameterDefaultsFromStruct(s interface{}) error
    GetName() string
    GetSlug() string
    GetDescription() string
    GetPrefix() string
    Clone() ParameterLayer
}

type ParameterLayerImpl struct {
    Name                 string
    Slug                 string
    Description          string
    Prefix               string
    ParameterDefinitions *parameters.ParameterDefinitions
    ChildLayers          []ParameterLayer
}

const DefaultSlug = "default"
```

**How It Works**:
- Layers group related parameters (e.g., "default" for command-specific params, "glazed" for output formatting)
- Each layer has a slug (unique identifier), name, description, and optional prefix
- Layers can be added to Cobra commands, which creates flags automatically
- Layers support nested/child layers for hierarchical organization

### 3. Parsed Layers

**Location**: `glazed/pkg/cmds/layers/parsed-layer.go`

**Key Types**:
- `ParsedLayer`: Runtime values for a single layer
- `ParsedLayers`: Ordered map of ParsedLayer objects

**Key Functions**:
- `InitializeStruct()`: Populates a struct from parsed layer values
- `GetOrCreate()`: Gets or creates a parsed layer for a given ParameterLayer
- `Merge()`: Merges parsed layers together

**Key Symbols**:
```go
type ParsedLayer struct {
    Layer      ParameterLayer
    Parameters *parameters.ParsedParameters
}

type ParsedLayers struct {
    *orderedmap.OrderedMap[string, *ParsedLayer]
}

func (p *ParsedLayers) InitializeStruct(layerKey string, dst interface{}) error
```

**How It Works**:
- ParsedLayers contain the actual runtime values after parsing
- Each ParsedLayer is linked to its ParameterLayer definition
- Values are stored in ParsedParameters (map of name -> ParsedParameter)
- `InitializeStruct()` extracts values into struct fields using `glazed.parameter` tags

### 4. Struct Tag Parsing

**Location**: `glazed/pkg/cmds/parameters/initialize-struct.go`

**Key Functions**:
- `InitializeStruct()`: Main function that populates structs from ParsedParameters
- `parsedTagOptions()`: Parses `glazed.parameter` tag syntax

**Tag Syntax**:
```
glazed.parameter:"parameter-name"           // Basic parameter mapping
glazed.parameter:"parameter-name,from_json" // Parse JSON from string value
glazed.parameter:"pattern*"                // Wildcard matching for maps
```

**Key Symbols**:
```go
func (p *ParsedParameters) InitializeStruct(s interface{}) error
func parsedTagOptions(tag string) (*tagOptions, error)
```

**How It Works**:
- Struct fields tagged with `glazed.parameter:"name"` are populated from ParsedParameters
- The tag name must match a parameter name in the ParsedParameters
- Supports wildcards for map fields (e.g., `glazed.parameter:"env.*"` matches all `env.*` parameters)
- Supports `from_json` option to parse JSON strings into complex types

### 5. Middlewares

**Location**: `glazed/pkg/cmds/middlewares/`

**Key Types**:
- `Middleware`: Function type `func(HandlerFunc) HandlerFunc`
- `HandlerFunc`: Function type `func(*ParameterLayers, *ParsedLayers) error`

**Key Middleware Functions**:
- `ParseFromCobraCommand()`: Parses flags from Cobra command (highest priority)
- `GatherArguments()`: Parses positional arguments
- `UpdateFromEnv()`: Loads values from environment variables
- `LoadParametersFromFiles()`: Loads values from JSON/YAML config files
- `SetFromDefaults()`: Sets default values from ParameterDefinitions (lowest priority)
- `UpdateFromMap()`: Updates from a map (for programmatic use)

**Key Symbols**:
```go
type Middleware func(HandlerFunc) HandlerFunc
type HandlerFunc func(*layers.ParameterLayers, *layers.ParsedLayers) error

func ExecuteMiddlewares(
    layers_ *layers.ParameterLayers,
    parsedLayers *layers.ParsedLayers,
    middlewares ...Middleware,
) error
```

**How It Works**:
- Middlewares are executed in reverse order (first middleware wraps the last)
- Each middleware calls `next()` to continue the chain
- Middlewares modify `ParsedLayers` by merging new values
- Priority order: CLI flags > Arguments > Env vars > Config files > Defaults

**Middleware Files**:
- `middlewares.go`: Core middleware execution logic
- `cobra.go`: Cobra command parsing middleware
- `update.go`: Environment variable and map update middlewares
- `profiles.go`: Profile loading middleware
- `load-parameters-from-json.go`: Config file loading middleware
- `layers.go`: Layer manipulation middlewares
- `whitelist.go`: Filtering middlewares

### 6. Cobra Integration

**Location**: `glazed/pkg/cli/cobra-parser.go`, `glazed/pkg/cli/cobra.go`

**Key Types**:
- `CobraParser`: Parser that converts ParameterLayers to Cobra commands
- `CobraParserConfig`: Configuration for parser behavior
- `CobraMiddlewaresFunc`: Function that returns middlewares for a command

**Key Functions**:
- `NewCobraParserFromLayers()`: Creates parser from ParameterLayers
- `AddToCobraCommand()`: Adds layer flags to a Cobra command
- `Parse()`: Executes middlewares and returns ParsedLayers
- `BuildCobraCommand()`: Unified builder for converting commands to Cobra

**Key Symbols**:
```go
type CobraParser struct {
    Layers              *layers.ParameterLayers
    middlewaresFunc     CobraMiddlewaresFunc
    shortHelpLayers     []string
    skipCommandSettingsLayer bool
    enableProfileSettingsLayer bool
    enableCreateCommandSettingsLayer bool
}

type CobraMiddlewaresFunc func(
    parsedCommandLayers *layers.ParsedLayers,
    cmd *cobra.Command,
    args []string,
) ([]cmd_middlewares.Middleware, error)

func CobraCommandDefaultMiddlewares(
    parsedCommandLayers *layers.ParsedLayers,
    cmd *cobra.Command,
    args []string,
) ([]cmd_middlewares.Middleware, error)
```

**How It Works**:
- `CobraParser` bridges ParameterLayers to Cobra commands
- `AddToCobraCommand()` iterates through layers and adds flags via `AddLayerToCobraCommand()`
- `Parse()` executes the middleware chain to populate ParsedLayers
- Default middleware chain: Cobra flags → Arguments → Defaults
- Can be extended with env vars, config files via `CobraParserConfig`

### 7. Command Building

**Location**: `glazed/pkg/cli/cobra.go`

**Key Functions**:
- `BuildCobraCommand()`: Unified builder (detects command type automatically)
- `BuildCobraCommandFromCommand()`: Determines run function based on interfaces
- `BuildCobraCommandFromCommandAndFunc()`: Core builder implementation

**Key Symbols**:
```go
func BuildCobraCommand(
    command cmds.Command,
    opts ...CobraOption,
) (*cobra.Command, error)

type CobraOption func(cfg *commandBuildConfig)

type commandBuildConfig struct {
    DualMode         bool
    GlazeToggleFlag  string
    DefaultToGlaze   bool
    HiddenGlazeFlags []string
    ParserCfg        CobraParserConfig
}
```

**How It Works**:
- Detects command interface (BareCommand, WriterCommand, GlazeCommand)
- Creates CobraParser from command's ParameterLayers
- Adds flags to Cobra command via parser
- Sets up Run function that parses layers and calls command's Run method
- Supports dual-mode commands (both classic and structured output)

## Data Flow

### Current Flow (Complex)

1. **Define Parameters**: Create `ParameterDefinition` objects manually
2. **Create Layers**: Group parameters into `ParameterLayer` objects
3. **Add to Command**: Attach layers to `CommandDescription`
4. **Build Cobra Command**: Convert to Cobra via `BuildCobraCommand()`
5. **Parse**: Execute middleware chain to populate `ParsedLayers`
6. **Extract**: Call `parsedLayers.InitializeStruct()` to populate settings struct

### Proposed Simplified Flow

1. **Define Struct**: Create settings struct with `glazed.parameter` tags
2. **Create Parser**: `appconfig.NewConfigParser[AppSettings](...)`
3. **Build Command**: `parser.ToCobraCommand(...)`
4. **Parse**: `settings := parser.Parse()` (handles all middlewares internally)
5. **Use**: Direct access to populated struct fields

## Key Files and Symbols Reference

### Parameter Definitions
- **File**: `glazed/pkg/cmds/parameters/parameters.go`
- **Types**: `ParameterDefinition`, `ParameterDefinitions`, `ParameterType`
- **Functions**: `NewParameterDefinition()`, `InitializeDefaultsFromStruct()`

### Struct Initialization
- **File**: `glazed/pkg/cmds/parameters/initialize-struct.go`
- **Functions**: `InitializeStruct()`, `parsedTagOptions()`
- **Tag**: `glazed.parameter:"name"`

### Layers
- **File**: `glazed/pkg/cmds/layers/layer.go`, `layer-impl.go`
- **Types**: `ParameterLayer`, `ParameterLayerImpl`, `ParameterLayers`
- **Constants**: `DefaultSlug = "default"`
- **Functions**: `NewParameterLayer()`, `AddLayerToCobraCommand()`

### Parsed Layers
- **File**: `glazed/pkg/cmds/layers/parsed-layer.go`
- **Types**: `ParsedLayer`, `ParsedLayers`
- **Functions**: `InitializeStruct()`, `GetOrCreate()`, `Merge()`

### Middlewares
- **File**: `glazed/pkg/cmds/middlewares/middlewares.go`
- **Types**: `Middleware`, `HandlerFunc`
- **Functions**: `ExecuteMiddlewares()`, `Chain()`
- **File**: `glazed/pkg/cmds/middlewares/cobra.go`
- **Functions**: `ParseFromCobraCommand()`, `GatherArguments()`
- **File**: `glazed/pkg/cmds/middlewares/update.go`
- **Functions**: `UpdateFromEnv()`, `UpdateFromMap()`, `SetFromDefaults()`
- **File**: `glazed/pkg/cmds/middlewares/load-parameters-from-json.go`
- **Functions**: `LoadParametersFromFiles()`

### Cobra Integration
- **File**: `glazed/pkg/cli/cobra-parser.go`
- **Types**: `CobraParser`, `CobraParserConfig`, `CobraMiddlewaresFunc`
- **Functions**: `NewCobraParserFromLayers()`, `AddToCobraCommand()`, `Parse()`
- **File**: `glazed/pkg/cli/cobra.go`
- **Functions**: `BuildCobraCommand()`, `BuildCobraCommandFromCommand()`
- **File**: `glazed/pkg/cli/cli.go`
- **Functions**: `NewCommandSettingsLayer()`, `NewProfileSettingsLayer()`
- **Types**: `CommandSettings`, `ProfileSettings`

### Settings Layers
- **File**: `glazed/pkg/settings/glazed_layer.go`
- **Types**: `GlazedParameterLayers`
- **Functions**: `NewGlazedParameterLayers()`

## Implementation Insights

### How Structs Map to Parameters

1. **Struct → ParameterDefinitions**: 
   - Use reflection to scan struct fields with `glazed.parameter` tags
   - Infer parameter type from Go type (int → ParameterTypeInteger, string → ParameterTypeString)
   - Create `ParameterDefinition` for each tagged field
   - Use field name as parameter name (or tag value if specified)

2. **ParameterDefinitions → Layers**:
   - Group parameters by struct (one struct = one layer)
   - Or use nested structs to create multiple layers
   - Slug can be derived from struct name (snake_case)

3. **Layers → Cobra Flags**:
   - `ParameterLayerImpl.AddLayerToCobraCommand()` iterates through ParameterDefinitions
   - Calls `ParameterDefinitions.AddParametersToCobraCommand()`
   - Creates Cobra flags with proper types and help text

4. **Cobra Flags → ParsedLayers**:
   - `ParseFromCobraCommand()` middleware reads Cobra flag values
   - Creates `ParsedParameter` objects with values
   - Merges into appropriate `ParsedLayer`

5. **ParsedLayers → Struct**:
   - `InitializeStruct()` reads `glazed.parameter` tags
   - Looks up parameter value in ParsedParameters
   - Sets struct field value using reflection

### Middleware Chain Execution

The middleware chain executes in reverse order:

```go
// Given: [m1, m2, m3]
// Executes as: m1(m2(m3(handler)))(layers, parsedLayers)
```

Typical chain order (lowest to highest priority):
1. `SetFromDefaults()` - Set defaults from ParameterDefinitions
2. `LoadParametersFromFiles()` - Load from config files
3. `UpdateFromEnv()` - Load from environment variables
4. `GatherArguments()` - Parse positional arguments
5. `ParseFromCobraCommand()` - Parse CLI flags (highest priority)

### Environment Variable Naming

Environment variables are constructed as:
```
{PREFIX}_{LAYER_PREFIX}{PARAMETER_NAME}
```

Where:
- `PREFIX`: Application prefix (e.g., "PINOCCHIO")
- `LAYER_PREFIX`: Layer prefix (e.g., "db-")
- `PARAMETER_NAME`: Parameter name (e.g., "host")
- All converted to UPPER_CASE with hyphens replaced by underscores

Example: `PINOCCHIO_DB_HOST`

### Config File Resolution

Config files are resolved via `glazedConfig.ResolveAppConfigPath()`:
- Checks explicit `--config-file` flag
- Falls back to standard locations: `$HOME/.{appname}/config.yaml`, `$XDG_CONFIG_HOME/{appname}/config.yaml`
- Supports multiple files in precedence order (low → high)

## Current Complexity Points

1. **Manual Layer Creation**: Must create `ParameterLayer` objects and add `ParameterDefinition`s
2. **Middleware Setup**: Must explicitly create middleware chain with correct order
3. **Parser Configuration**: Must configure `CobraParserConfig` with app name, env prefix, etc.
4. **Multiple Steps**: Requires 5-6 distinct steps to go from struct to working command
5. **Layer Slugs**: Must manage layer slugs manually (though `DefaultSlug` helps)

## Opportunities for Simplification

1. **Auto-generate Layers**: Derive layers from struct types automatically
2. **Auto-generate ParameterDefinitions**: Infer from struct fields and tags
3. **Default Middleware Chain**: Provide sensible defaults (flags → env → config → defaults)
4. **Unified Parser**: Single `ConfigParser` type that handles all setup
5. **Type Inference**: Infer parameter types from Go types automatically
6. **Nested Struct Support**: Support nested structs as separate layers automatically

## Next Steps

1. Design simplified API (`appconfig` package)
2. Implement struct-to-layer conversion
3. Implement struct-to-parameter conversion
4. Create unified parser with default middleware chain
5. Add Cobra command builder integration
6. Test with real-world examples
