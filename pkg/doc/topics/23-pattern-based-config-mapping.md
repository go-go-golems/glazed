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

The pattern-based config mapping system provides a declarative way to map arbitrary config file structures to Glazed's layer-based parameter system without writing custom Go functions. Instead of implementing `ConfigFileMapper` functions with manual config traversal, you define mapping rules that specify patterns to match in config files and how to map matched values to parameters. This keeps configuration logic concise, testable, and consistent across commands.

## How it works (mental model)
-## Sections at a glance

- Quick Start: Minimal code to get started
- Mapping rules from YAML/JSON files: Define and load rules from files
- Pattern Syntax: Exact, captures, wildcards, nested rules, inheritance
- Using the Mapper: Wire into `LoadParametersFromFile`
- Validation: Build-time vs runtime checks, validate-only pass
- Matching, ambiguity, and errors: Deterministic traversal and fail-fast rules
- When to Use: Choose pattern mappers vs custom mappers
- Complete Example: End-to-end demo
- Reference: `MappingRule` and the Builder API
- Current Limitations and See Also


Pattern mapping runs in two stages:

1) Build-time (rule compilation)
- Parse and validate rule syntax (segments, wildcards, named captures)
- Verify target layers exist and static target parameters are valid (prefix-aware)
- Check that any `{name}` referenced in `TargetParameter` is captured in `Source`

2) Runtime (matching and writes)
- Traverse the config in deterministic (lexicographic) order
- For each pattern, collect matches; resolve `{captures}` into parameter names
- Write values to the target layer/parameter; error on ambiguity or collisions
- Respect `Required: true` by failing if no match is found (with path hints)

## Quick Start

A minimal example shows how to map a simple config structure:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
)

// Create a pattern mapper
mapper, err := pm.NewConfigMapper(layers_,
    pm.MappingRule{
        Source:          "app.settings.api_key",
        TargetLayer:     "demo",
        TargetParameter: "api-key",
    },
)

// Use with LoadParametersFromFile
mw := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

### Quick Start (Builder API)

Prefer a fluent API? Use the builder to assemble rules, then build a mapper with the same strict validation and semantics:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
)

b := pm.NewConfigMapperBuilder(layers_).
    Map("app.settings.api_key", "demo", "api-key")

mapper, err := b.Build()
if err != nil {
    panic(err)
}

// Use with LoadParametersFromFile
mw := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

## Mapping rules from YAML/JSON files

For larger rule sets or configuration-driven apps, define rules in YAML/JSON and load them at startup. The loader accepts either a top-level `mappings` array or a bare array of rules.

### File format

Top-level `mappings`:

```yaml
mappings:
  - source: "app.settings.api_key"
    target_layer: "demo"
    target_parameter: "api-key"
  - source: "app.{env}.api_key"
    target_layer: "demo"
    target_parameter: "{env}-api-key"
```

Bare array:

```yaml
- source: "app.settings.threshold"
  target_layer: "demo"
  target_parameter: "threshold"
```

### Loading rules (and mapper)

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
)

rules, err := pm.LoadRulesFromFile("mappings.yaml")
if err != nil { panic(err) }

// Option 1: build explicitly
mapper, err := pm.NewConfigMapper(layers_, rules...)
if err != nil { panic(err) }

// Option 2: convenience
mapper2, err := pm.LoadMapperFromFile(layers_, "mappings.yaml")
if err != nil { panic(err) }

mw := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

## Pattern Syntax

Pattern matching enables flexible config file mapping through several mechanisms, each designed for specific config structures. The goal is to make rules readable and intentional while keeping runtime behavior deterministic and debuggable.

### Exact Match

Exact match patterns map specific config paths to parameters with no variation:

```go
patternmapper.MappingRule{
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
patternmapper.MappingRule{
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
patternmapper.MappingRule{
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

Note: Wildcards (`*`) match but don't capture. To use the matched value, use named captures `{name}` instead.

Important: When a wildcard pattern matches multiple keys with different values, the mapper treats this as an ambiguity and returns an error by default. Use named captures (e.g., `app.{env}.api_key`) if you need to collect multiple values, or ensure matched values are identical if a single target parameter is intended. This prevents accidental aggregation of unrelated values.

### Nested Rules

Nested rules group related mappings together for cleaner syntax and avoid repeating common prefixes:

```go
patternmapper.MappingRule{
    Source:      "app.settings",
    TargetLayer: "demo",
    Rules: []patternmapper.MappingRule{
        {Source: "api_key", TargetParameter: "api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
        {Source: "timeout", TargetParameter: "timeout"},
    },
}
```

Builder equivalent:

```go
b := patternmapper.NewConfigMapperBuilder(layers).
    MapObject("app.settings", "demo", []patternmapper.MappingRule{
        patternmapper.Child("api_key", "api-key"),
        patternmapper.Child("threshold", "threshold"),
        patternmapper.Child("timeout", "timeout"),
    })
mapper, err := b.Build()
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
patternmapper.MappingRule{
    Source:      "app.{env}.settings",
    TargetLayer: "demo",
    Rules: []patternmapper.MappingRule{
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

## Builder API

Use the builder for a fluent way to assemble rules while keeping the same strict behavior and validation. Builders are useful when you want to co-locate mapping intent with code or construct rules conditionally based on application state.

```go
b := patternmapper.NewConfigMapperBuilder(layers).
    Map("app.settings.api_key", "demo", "api-key", true).
    MapObject("app.{env}.settings", "demo", []patternmapper.MappingRule{
        patternmapper.Child("api_key", "{env}-api-key"),
        patternmapper.Child("threshold", "threshold"),
    })

mapper, err := b.Build() // Validates via NewConfigMapper
```

Notes:
- Same strict semantics: multi-match ambiguity and cross-rule collisions error by default.
- Prefix-aware parameter resolution and compile-time validation of static targets apply at Build().
- One level of nested rules is supported.
- Positional captures are not supported (use named captures).

## Validation

Pattern mappers validate at creation time to catch errors early and provide clear feedback. This happens when you call `NewConfigMapper`, not at runtime when processing config files. At runtime, `Map` enforces dynamic constraints and reports precise errors with path hints.

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
patternmapper.MappingRule{
    Source:          "app.settings.api_key",
    TargetLayer:     "demo",
    TargetParameter: "api-key",
    Required:        true,  // Error if not found
}
```

Without `Required: true`, patterns that don't match are silently skipped, allowing optional config values.

## Ambiguity Handling

The mapper fails fast on ambiguous situations to prevent unpredictable writes and hard-to-debug behavior:

- Multi-match: If a single pattern matches multiple paths that would resolve to the same target parameter with different values (e.g., `app.*.api_key` for `dev` and `prod`), an error is returned.
- Collisions: If different patterns resolve to the same target parameter (e.g., `app.settings.api_key` and `config.api_key` both mapping to `demo.api-key`), an error is returned.

## Error Handling

Runtime errors occur when config files don't match expectations or reference nonexistent parameters. The system provides detailed error messages to aid debugging. Error messages include both the user-provided target name and the canonical (prefix-aware) parameter name where relevant.

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

**Example error for missing parameter with prefix**:
```
target parameter "api-key" (checked as "demo-api-key") does not exist in layer "demo" (pattern: "app.settings.api_key")
```

When a layer has a prefix and the target parameter name doesn't include it, the error message shows both the provided name and the resolved canonical name (with prefix). This helps debug parameter name mismatches.

## Matching Order and Overwrites

- Deterministic traversal: The mapper traverses config objects in lexicographic key order to ensure stable behavior across runs.
- Prefer captures over wildcards: When collecting multiple values (e.g., per-environment), named captures make intent explicit and avoid ambiguity.
- Overwrites: Overwrites across different rules to the same parameter are considered collisions and will error.

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

A real-world example showing pattern mapper integration. This example highlights capture inheritance, prefix-aware parameters, and minimal application wiring.

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
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
    mapper, err := patternmapper.NewConfigMapper(paramLayers,
        patternmapper.MappingRule{
            Source:      "app.{env}.settings",
            TargetLayer: "demo",
            Rules: []patternmapper.MappingRule{
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
patternMapper, _ := patternmapper.NewConfigMapper(layers, rules...)
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

<!-- Validation content consolidated in the main Validation section above -->

## Best practices

- Keep rule sets small and focused (3–10 rules). Split by feature if needed.
- Prefer named captures with descriptive names (`{env}`, `{tenant}`) over wildcards.
- Mark critical paths `Required: true` to fail fast when config drifts.
- Start with a validate-only pass in CI to surface actionable errors early.
- Use mapping files when rules are shared across commands or must be updated without recompiling.

