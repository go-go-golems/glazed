# Config Pattern Mapper Example

This example demonstrates the new pattern-based config mapping system in Glazed.

## Overview

The pattern mapper allows you to declaratively map config file structures to layer parameters using pattern matching rules, without writing custom Go functions.

## Key Features

1. **Exact Match**: Simple one-to-one mappings
2. **Named Captures**: Extract values from config paths and use them in parameter names
3. **Wildcards**: Match multiple paths with a single pattern
4. **Nested Rules**: Group related mappings together with clean syntax
5. **Capture Inheritance**: Child rules inherit captures from parent rules
6. **Builder API**: Fluent way to assemble rules with the same strict validation
7. **YAML/JSON Loader**: Load mapping rules from a file (`mappings.yaml`)

## Running the Example

```bash
go run main.go
```

## Pattern Syntax

### Exact Match
```go
{
    Source: "app.settings.api_key",
    TargetLayer: "demo",
    TargetParameter: "api-key",
}
```

### Named Capture
```go
{
    Source: "app.{env}.api_key",  // Captures "env"
    TargetLayer: "demo",
    TargetParameter: "{env}-api-key",  // Uses captured value
}
```

### Wildcard
```go
{
    Source: "app.*.api_key",  // Matches any environment
    TargetLayer: "demo",
    TargetParameter: "api-key",
}
```

### Nested Rules
```go
{
    Source: "app.settings",
    TargetLayer: "demo",
    Rules: []MappingRule{
        {Source: "api_key", TargetParameter: "api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
    },
}
```

## Comparison with ConfigFileMapper

### Old Way (ConfigFileMapper function):
```go
mapper := func(rawConfig interface{}) (map[string]map[string]interface{}, error) {
    configMap := rawConfig.(map[string]interface{})
    result := map[string]map[string]interface{}{
        "demo": make(map[string]interface{}),
    }
    
    // Manual traversal and mapping
    if app, ok := configMap["app"].(map[string]interface{}); ok {
        if settings, ok := app["settings"].(map[string]interface{}); ok {
            if apiKey, ok := settings["api_key"]; ok {
                result["demo"]["api-key"] = apiKey
            }
        }
    }
    
    return result, nil
}
```

### New Way (Pattern Mapper - Rules Array):
```go
mapper, err := patternmapper.NewConfigMapper(layers,
    patternmapper.MappingRule{
        Source:          "app.settings.api_key",
        TargetLayer:     "demo",
        TargetParameter: "api-key",
    },
)
```

### New Way (Pattern Mapper - Builder API):
```go
b := patternmapper.NewConfigMapperBuilder(layers).
    Map("app.settings.api_key", "demo", "api-key").
    MapObject("environments.{env}.settings", "demo", []patternmapper.MappingRule{
        patternmapper.Child("api_key", "{env}-api-key"),
    })

mapper, err := b.Build() // Validates via NewConfigMapper
```

### New Way (Pattern Mapper - YAML/JSON Loader):
```go
// mappings.yaml
mappings:
  - source: "environments.{env}.settings"
    target_layer: "demo"
    rules:
      - source: "api_key"
        target_parameter: "{env}-api-key"

// Go
mapper, err := patternmapper.LoadMapperFromFile(layers, "mappings.yaml")
if err != nil { /* handle */ }
middleware := sources.FromFile(
    "config.yaml",
    middlewares.WithConfigMapper(mapper),
)
```

## When to Use

- **Pattern Mapper**: For simple to moderately complex mappings (5 rules or less)
- **ConfigFileMapper**: For complex logic, arrays, conditionals, or transformations

Both approaches are fully supported and can be used interchangeably.

