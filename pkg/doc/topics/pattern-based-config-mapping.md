---
Title: Pattern-Based Config Mapping
Slug: pattern-based-config-mapping
Short: Declarative mapping of config files to parameter layers using pattern matching rules
Topics:
- configuration
- middlewares
- patterns
- mapping
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Pattern-Based Config Mapping

The pattern-based config mapping system provides a declarative way to map arbitrary config file structures to Glazed's layer-based parameter system without writing custom Go functions. Instead of implementing `ConfigFileMapper` functions with manual config traversal, you define mapping rules that specify patterns to match in config files and how to map matched values to parameters.

## Quick Start

A minimal example shows how to map a simple config structure:

```go
// Create a pattern mapper
mapper, err := middlewares.NewConfigMapper(layers,
    middlewares.MappingRule{
        Source:          "app.settings.api_key",
        TargetLayer:     "demo",
        TargetParameter: "api-key",
    },
)

// Use with LoadParametersFromFile
middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

## Pattern Syntax

Pattern matching enables flexible config file mapping through several mechanisms, each designed for specific config structures.

### Exact Match

Exact match patterns map specific config paths to parameters with no variation:

```go
middlewares.MappingRule{
    Source:          "app.settings.api_key",
    TargetLayer:     "demo",
    TargetParameter: "api-key",
}
```

**Config**:
```yaml
app:
  settings:
    api_key: "secret123"
```

**Result**: `demo.api-key = "secret123"`

### Named Captures

Named captures extract segments from config paths and use them in parameter names, enabling environment-specific or multi-tenant configurations:

```go
middlewares.MappingRule{
    Source:          "app.{env}.api_key",
    TargetLayer:     "demo",
    TargetParameter: "{env}-api-key",
}
```

**Config**:
```yaml
app:
  dev:
    api_key: "dev-secret"
  prod:
    api_key: "prod-secret"
```

**Result**:
- `demo.dev-api-key = "dev-secret"`
- `demo.prod-api-key = "prod-secret"`

The `{env}` capture extracts whatever value appears at that position in the config (here, "dev" or "prod") and makes it available for use in the target parameter name.

### Wildcards

Wildcards match any value at a specific level without capturing it, useful when you need to match patterns but don't need the matched value:

```go
middlewares.MappingRule{
    Source:          "app.*.api_key",
    TargetLayer:     "demo",
    TargetParameter: "api-key",
}
```

**Config**:
```yaml
app:
  dev:
    api_key: "dev-secret"
  prod:
    api_key: "prod-secret"
```

**Result**: `demo.api-key = "prod-secret"`

Note on order: Wildcard iteration is deterministic. Keys are traversed in lexicographic order, so when multiple keys match (e.g., `dev`, `prod`), the last one in sorted order wins (here: `prod`).

**Note**: Wildcards (`*`) match but don't capture. To use the matched value, use named captures `{name}` instead.

### Nested Rules

Nested rules group related mappings together for cleaner syntax and avoid repeating common prefixes:

```go
middlewares.MappingRule{
    Source:      "app.settings",
    TargetLayer: "demo",
    Rules: []middlewares.MappingRule{
        {Source: "api_key", TargetParameter: "api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
        {Source: "timeout", TargetParameter: "timeout"},
    },
}
```

**Config**:
```yaml
app:
  settings:
    api_key: "secret123"
    threshold: 42
    timeout: 30
```

**Result**:
- `demo.api-key = "secret123"`
- `demo.threshold = 42`
- `demo.timeout = 30`

Child rules' source paths are relative to the parent's resolved object, eliminating the need to repeat `app.settings` for each mapping.

### Capture Inheritance

Nested rules inherit captures from parent patterns, enabling complex multi-level mappings:

```go
middlewares.MappingRule{
    Source:      "app.{env}.settings",
    TargetLayer: "demo",
    Rules: []middlewares.MappingRule{
        {Source: "api_key", TargetParameter: "{env}-api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
    },
}
```

**Config**:
```yaml
app:
  dev:
    settings:
      api_key: "dev-secret"
      threshold: 10
  prod:
    settings:
      api_key: "prod-secret"
      threshold: 100
```

**Result**:
- `demo.dev-api-key = "dev-secret"`
- `demo.prod-api-key = "prod-secret"`
- `demo.threshold = 10` *(from dev)*
- `demo.threshold = 100` *(from prod, overwrites)*

The `{env}` capture from the parent pattern is available in all child rules, allowing them to construct environment-specific parameter names.

## MappingRule Structure

The `MappingRule` struct defines a single mapping with these fields:

```go
type MappingRule struct {
    // Source path pattern (required)
    // Supports: exact match, wildcard (*), named capture ({name})
    Source string

    // Target layer slug (required for leaf rules)
    // Inherited by child rules if not set
    TargetLayer string

    // Target parameter name (required for leaf rules)
    // Supports capture references: "{env}-api-key"
    TargetParameter string

    // Optional: nested rules for mapping child objects
    Rules []MappingRule

    // Optional: whether pattern must match (default: false)
    Required bool
}
```

## Validation

Pattern mappers validate at creation time to catch errors early and provide clear feedback. This happens when you call `NewConfigMapper`, not at runtime when processing config files.

**Validation checks**:
1. **Pattern syntax**: Valid segments, capture groups, wildcards
2. **Capture references**: All `{name}` in target must exist in source
3. **Target layer**: Must exist in parameter layers
4. **Target parameter**: Must exist in target layer *(validated at runtime per match)*

**Common validation errors**:

```go
// Error: capture reference not in source
{
    Source: "app.settings.api_key",
    TargetParameter: "{env}-api-key",  // {env} not captured in source
}
// Error: "capture reference {env} in target parameter not found in source pattern"

// Error: target layer doesn't exist
{
    Source: "app.settings.api_key",
    TargetLayer: "nonexistent",
    TargetParameter: "api-key",
}
// Error: "target layer \"nonexistent\" does not exist"
```

## Required Patterns

Mark patterns as required to enforce that specific config values must be present:

```go
middlewares.MappingRule{
    Source:          "app.settings.api_key",
    TargetLayer:     "demo",
    TargetParameter: "api-key",
    Required:        true,  // Error if not found
}
```

Without `Required: true`, patterns that don't match are silently skipped, allowing optional config values.

## Error Handling

Runtime errors occur when config files don't match expectations or reference nonexistent parameters. The system provides detailed error messages to aid debugging.

**Error conditions**:
- Required pattern doesn't match
- Target parameter doesn't exist in layer
- Invalid pattern syntax *(caught at creation)*
- Invalid capture references *(caught at creation)*

**Example error message**:
```
required pattern "app.settings.api_key" did not match any paths in config
```

**Example error for missing parameter**:
```
target parameter "api-key" does not exist in layer "demo" (pattern: "app.settings.api_key")
```

## Matching Order and Overwrites

- Deterministic traversal: The mapper traverses config objects in lexicographic key order to ensure stable behavior across runs.
- Wildcards: If a wildcard pattern (e.g., `app.*.api_key`) matches multiple keys at the same level, the last match in sorted order wins. Prefer named captures (e.g., `app.{env}.api_key`) if you need all values.
- Overwrites: When multiple matches resolve to the same target parameter, later matches overwrite earlier ones. Structure rules to avoid unintentional collisions or use captures to disambiguate.

## When to Use

Choose between pattern mappers and `ConfigFileMapper` functions based on complexity:

**Use Pattern Mapper For:**
- Simple mappings (3-5 rules or less)
- Flat or nested config structures
- Environment-specific mappings with captures
- Grouped parameters (nested rules)
- Standard config transformations

**Use ConfigFileMapper For:**
- Complex transformations
- Array handling
- Conditional logic
- Value transformations
- More than 5 mappings
- Custom validation logic

Both approaches are fully supported and can be used interchangeably in the same codebase.

## Complete Example

A real-world example showing pattern mapper integration:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func main() {
    // Define parameter layers
    layer, _ := layers.NewParameterLayer("demo", "Demo",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
            parameters.NewParameterDefinition("dev-api-key", parameters.ParameterTypeString),
            parameters.NewParameterDefinition("prod-api-key", parameters.ParameterTypeString),
        ),
    )
    paramLayers := layers.NewParameterLayers(layers.WithLayers(layer))

    // Create pattern mapper with capture inheritance
    mapper, err := middlewares.NewConfigMapper(paramLayers,
        middlewares.MappingRule{
            Source:      "app.{env}.settings",
            TargetLayer: "demo",
            Rules: []middlewares.MappingRule{
                {Source: "api_key", TargetParameter: "{env}-api-key"},
            },
        },
    )
    if err != nil {
        panic(err)
    }

    // Use with middleware
    middleware := middlewares.LoadParametersFromFile(
        "config.yaml",
        middlewares.WithConfigMapper(mapper),
    )

    // Execute middleware chain
    parsedLayers := layers.NewParsedLayers()
    err = middlewares.ExecuteMiddlewares(paramLayers, parsedLayers, middleware)
    if err != nil {
        panic(err)
    }
}
```

## Backward Compatibility

Pattern mappers coexist with `ConfigFileMapper` functions through the `ConfigMapper` interface, ensuring existing code continues to work:

```go
// Old way (still fully supported)
funcMapper := func(raw interface{}) (map[string]map[string]interface{}, error) {
    // Custom logic
    return result, nil
}
middleware1 := middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithConfigFileMapper(funcMapper))

// New way (pattern-based)
patternMapper, _ := middlewares.NewConfigMapper(layers, rules...)
middleware2 := middlewares.LoadParametersFromFile("config.yaml",
    middlewares.WithConfigMapper(patternMapper))
```

Both approaches implement the `ConfigMapper` interface and work identically with the middleware system.

## Current Limitations

Phase 1 implementation focuses on core pattern matching:

1. **No deep wildcards**: `**` is not supported
2. **No array wildcards**: `[*]` is not supported
3. **No positional captures**: `{0}`, `{1}` are not supported (use named captures)
4. **One level nesting**: Nested rules support one level only
5. **No value transformations**: Use `ConfigFileMapper` for transforming values

These limitations may be addressed in future phases based on user feedback.

## See Also

For more information on related topics:

```
glaze help middlewares-guide
glaze help parameter-layers-and-parsed-layers
```

**Example code**: See `cmd/examples/config-pattern-mapper/` for working examples.

