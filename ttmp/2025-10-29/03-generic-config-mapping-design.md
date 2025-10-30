# Generic Config Mapping Design

## Problem Statement

The current `ConfigFileMapper` requires users to write custom Go functions to transform config file structures into layer maps. This is flexible but verbose and requires code changes for each new config format. We need a declarative, pattern-based system that allows users to specify mappings using string patterns, globs, and captured matches.

## Key Design Decisions

### Capture Semantics

**Important**: Wildcards (`*`) match but don't capture by name. To capture values use **named captures** only:

1. **Named captures**: Use `{name}` in source pattern → `app.{env}.api_key` captures `env` as `"dev"` or `"prod"`

Strict semantics (Phase 1+2):
- Wildcards that match multiple keys with different values are considered ambiguous and will error by default. Prefer named captures (e.g., `app.{env}.api_key`) to collect separate values.
- If a wildcard matches multiple keys with identical values and maps to a single parameter, mapping succeeds.

**Example**:
- ❌ `Map("app.*.api_key", "demo", "{env}-api-key")` - `{env}` won't work (wildcard doesn't capture)
- ✅ `Map("app.{env}.api_key", "demo", "{env}-api-key")` - Named capture works

### Nested Rules

Instead of repeating prefixes, map entire objects with child rules:

**Before** (repetitive):
```go
Map("app.settings.api_key", "demo", "api-key")
Map("app.settings.threshold", "demo", "threshold")
Map("app.settings.timeout", "demo", "timeout")
```

**After** (clean):
```go
MapObject("app.settings", "demo", []MappingRule{
    {Source: "api_key", TargetParameter: "api-key"},
    {Source: "threshold", TargetParameter: "threshold"},
    {Source: "timeout", TargetParameter: "timeout"},
})
```

Child rules inherit captures from parent patterns, making complex mappings much cleaner.

## Goals

1. **Declarative**: Users should be able to specify mappings without writing Go code
2. **Pattern-based**: Support path patterns, globs, and captured groups
3. **Flexible**: Handle flat, nested, and arbitrarily structured config files
4. **Type-safe**: Allow validation and type checking where possible
5. **Easy to use**: Simple API for common cases, powerful for complex ones

## Design Overview

### Core Concept

A mapping rule specifies:
- **Source path**: Where to find the value in the config file (supports globs/captures)
- **Target layer**: Which layer to place the value in
- **Target parameter**: Which parameter name to use

### Mapping Rule Format

```go
type MappingRule struct {
    // Source path pattern (e.g., "app.settings.api_key", "app.*.key", "app.{env}.api-key")
    Source string
    
    // Target layer slug (e.g., "demo")
    // If not set in child rules, inherits from parent rule
    // Ignored if TransformFunc is provided
    TargetLayer string
    
    // Target parameter name (supports captures like "{env}-api-key" or "{0}-api-key")
    // Ignored if TransformFunc is provided
    TargetParameter string
    
    // Optional: function to dynamically compute target layer and parameter
    // Takes captured values (named and positional) and returns target layer and parameter name
    // If provided, overrides TargetLayer and TargetParameter
    TransformFunc func(captures map[string]string, positional []string, value interface{}) (targetLayer string, targetParameter string)
    
    // Optional: nested rules for mapping child objects
    // If provided, Source should point to an object, and Rules maps its children
    Rules []MappingRule
    
    // Optional: default value if source not found
    Default interface{}
    
    // Optional: whether to skip if source doesn't exist (default: false, means use default)
    Required bool
}
```

### Path Pattern Syntax

1. **Exact match**: `app.settings.api_key` → matches exactly that path
2. **Wildcard**: `app.*.api_key` → matches `app.dev.api_key`, `app.prod.api_key`, etc. (does NOT capture)
3. **Capture groups**: `app.{env}.api_key` → matches and captures the segment (e.g., `env = "dev"`)
4. **Anonymous captures**: `app.*.api_key` with `{0}` in target → positional capture (first wildcard = `{0}`, second = `{1}`, etc.)
5. **Multiple wildcards**: `app.*.settings.*.key` → matches nested structures
6. **Array indexing**: `app.settings[0].key` → matches first element of array
7. **Deep wildcard**: `app.**.api_key` → matches at any depth (e.g., `app.settings.api_key`, `app.dev.settings.api_key`)

**Important**: Wildcards (`*`) match but don't capture by name. To capture, use named captures `{name}` or positional references `{0}`, `{1}`, etc.

### Target Parameter Syntax

- **Literal**: `api-key` → uses literal string
- **Named capture reference**: `{env}-api-key` → uses captured value from source pattern (e.g., `app.{env}.api_key`)
- **Positional capture reference**: `{0}-api-key` → uses first anonymous wildcard match (e.g., `app.*.api_key`)
- **Multiple captures**: `{env}-{region}-api-key` → combines multiple captures
- **Capture indices**: `{0}-{1}-key` → uses positional captures in order
- **Lambda function**: `TransformFunc` → programmatically compute target layer and parameter from captures (programmatic API only)

## API Design

### Option 1: Builder Pattern

```go
mapper := patternmapper.NewConfigMapper().
    // Simple mappings
    Map("app.settings.api_key", "demo", "api-key").
    Map("app.settings.threshold", "demo", "threshold").
    
    // Named capture (correct way)
    Map("app.{env}.api_key", "demo", "{env}-api-key").
    
    // Anonymous wildcard with positional capture
    Map("app.*.api_key", "demo", "{0}-api-key"). // {0} = first wildcard match
    
    // Lambda transformation: dynamically compute layer and parameter
    MapWithTransform("app.{env}.{service}.api_key", func(captures map[string]string, positional []string, value interface{}) (string, string) {
        // Custom logic to determine layer and parameter based on captures
        env := captures["env"]
        service := captures["service"]
        layer := fmt.Sprintf("%s-%s", env, service)
        param := "api-key"
        return layer, param
    }).
    
    // Nested rules: map entire object with sub-rules
    MapObject("app.settings", "demo", []MappingRule{
        {Source: "api_key", TargetParameter: "api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
        {Source: "timeout", TargetParameter: "timeout", Default: 30},
    }).
    
    Build()
```

### Option 2: Rules Array

```go
rules := []patternmapper.MappingRule{
    {
        Source: "app.settings.api_key",
        TargetLayer: "demo",
        TargetParameter: "api-key",
    },
    {
        // Named capture - correct way
        Source: "app.{env}.api_key",
        TargetLayer: "demo",
        TargetParameter: "{env}-api-key",
    },
    // (Removed positional capture example. Prefer named captures instead.)
    {
        // Lambda transformation: dynamically compute layer and parameter
        Source: "app.{env}.{service}.api_key",
        TransformFunc: func(captures map[string]string, positional []string, value interface{}) (string, string) {
            env := captures["env"]
            service := captures["service"]
            // Custom logic: use environment as layer, service as part of parameter name
            layer := env
            param := fmt.Sprintf("%s-api-key", service)
            return layer, param
        },
    },
    {
        // Nested rules: map entire object
        Source: "app.settings",
        TargetLayer: "demo",
        Rules: []patternmapper.MappingRule{
            {Source: "api_key", TargetParameter: "api-key"},
            {Source: "threshold", TargetParameter: "threshold"},
        },
    },
}

mapper := patternmapper.NewConfigMapper(rules...)
```

### Option 3: YAML/JSON Config (Most Declarative)

```yaml
mappings:
  # Simple mapping
  - source: "app.settings.api_key"
    target_layer: "demo"
    target_parameter: "api-key"
  
  # Named capture
  - source: "app.{env}.api_key"
    target_layer: "demo"
    target_parameter: "{env}-api-key"
  
  # Anonymous wildcard with positional capture
  - source: "app.*.api_key"
    target_layer: "demo"
    target_parameter: "{0}-api-key"  # {0} = first wildcard match
  
  # Nested rules: map entire object
  - source: "app.settings"
    target_layer: "demo"
    rules:
      - source: "api_key"
        target_parameter: "api-key"
      - source: "threshold"
        target_parameter: "threshold"
        default: 10
  
  # Nested rules with captures
  - source: "app.{env}.settings"
    target_layer: "demo"
    rules:
      - source: "api_key"
        target_parameter: "{env}-api-key"  # Can use parent captures
      - source: "threshold"
        target_parameter: "threshold"
```

## Implementation Details

### Path Matching Algorithm

1. **Parse source pattern** into tokens:
   - Literal segments: `app`, `settings`
   - Wildcards: `*` (single level), `**` (deep)
   - Captures: `{name}` (capture and name)
   - Arrays: `[index]` or `[*]`

2. **Traverse config structure** recursively:
   - For each node, check if it matches the current pattern segment
   - Track captured values for use in target parameter
   - Continue matching child nodes

3. **Value extraction**:
   - When pattern fully matches, extract the value
   - Apply type conversion if specified
   - Use default if value missing and not required

### Example: Pattern Matching

Config file:
```yaml
app:
  dev:
    api_key: "dev-secret"
  prod:
    api_key: "prod-secret"
  settings:
    threshold: 42
```

Pattern: `app.{env}.api_key`

Matches:
- `app.dev.api_key` → captures `env="dev"`, maps to `demo.api-key` (or `demo.dev-api-key` if using capture)
- `app.prod.api_key` → captures `env="prod"`, maps to `demo.api-key`

### Capture Reference Resolution

When target parameter contains `{env}`:
- Use captured value from source pattern
- If multiple captures, use named references: `{env}`, `{region}`, etc.
- For positional captures: `{0}`, `{1}`, etc.
- **Nested rules inherit parent captures**: Child rules can reference captures from parent rules

### Nested Rules Processing

When a rule has `Rules` defined:

1. **Source resolution**: The `Source` path is resolved to find the target object in the config
2. **Context extraction**: Extract the matched object (e.g., `app.settings` → `{api_key: "...", threshold: 42}`)
3. **Relative path resolution**: Child rules' `Source` paths are relative to the parent's resolved object
4. **Capture inheritance**: Captures from parent pattern are available to child rules in their "capture environment"
5. **Layer inheritance**: Child rules inherit `TargetLayer` from parent unless explicitly overridden
6. **Recursive processing**: Child rules are processed in the context of the extracted object

**Capture Environment**: When processing nested rules, all captures from the parent rule are available in a "capture environment" that child rules can reference. This includes:
- Named captures: `{env}`, `{region}`, etc.
- Positional captures: `{0}`, `{1}`, etc.
- Multiple captures can be combined: `{env}-{region}-api-key`

**Layer Inheritance**: Child rules inherit the `TargetLayer` from their parent rule, but can override it if needed:
```yaml
- source: "app.settings"
  target_layer: "demo"      # Parent layer
  rules:
    - source: "api_key"
      target_parameter: "api-key"  # Uses parent's "demo" layer
    - source: "db_config"
      target_layer: "database"      # Overrides to "database" layer
      target_parameter: "host"
```

**TransformFunc in Nested Rules**: When using `TransformFunc` in nested rules, it receives captures from both the parent rule and any captures in its own source pattern:
```go
MapObject("app.{env}.settings", "demo", []patternmapper.MappingRule{
    {
        Source: "api_key",
        TransformFunc: func(captures map[string]string, positional []string, value interface{}) (string, string) {
            // captures["env"] available from parent rule
            env := captures["env"]
            return fmt.Sprintf("%s-settings", env), "api-key"
        },
    },
})
```

Example:
```yaml
# Config
app:
  settings:
    api_key: "secret"
    threshold: 42

# Mapping
- source: "app.settings"
  target_layer: "demo"
  rules:
    - source: "api_key"        # Relative to app.settings
      target_parameter: "api-key"
    - source: "threshold"
      target_parameter: "threshold"
```

With captures:
```yaml
# Config
app:
  dev:
    settings:
      api_key: "dev-secret"
  prod:
    settings:
      api_key: "prod-secret"

# Mapping
- source: "app.{env}.settings"    # Parent captures: {env: "dev"} or {env: "prod"}
  target_layer: "demo"
  rules:
    - source: "api_key"            # Child rule can reference {env} from parent
      target_parameter: "{env}-api-key"  # Uses parent capture
```

**Capture Environment Flow**:
```
1. Parent rule matches: "app.{env}.settings"
   - Matches "app.dev.settings" → Capture environment: {env: "dev"}
   - Matches "app.prod.settings" → Capture environment: {env: "prod"}

2. For each match, extract object: {api_key: "...", threshold: ...}

3. Process child rules with inherited capture environment:
   - Child rule: source="api_key", target="{env}-api-key"
   - env="dev" → creates "demo.dev-api-key"
   - env="prod" → creates "demo.prod-api-key"
```

**Benefits of nested rules**:
- Avoid repeating common prefixes (`app.settings`)
- Group related mappings together
- Share captures across multiple child mappings
- Easier to maintain and understand

## Usage Examples

### Example 1: Simple Flat Config

Config:
```yaml
api_key: "secret"
threshold: 42
```

Mapping:
```go
mapper := patternmapper.NewConfigMapper().
    Map("api_key", "demo", "api-key").
    Map("threshold", "demo", "threshold").
    Build()
```

### Example 2: Nested Structure

Config:
```yaml
app:
  settings:
    api:
      key: "secret"
      threshold: 42
```

Mapping:
```go
mapper := patternmapper.NewConfigMapper().
    Map("app.settings.api.key", "demo", "api-key").
    Map("app.settings.api.threshold", "demo", "threshold").
    Build()
```

### Example 3: Environment-Specific with Captures

Config:
```yaml
environments:
  dev:
    api_key: "dev-secret"
  prod:
    api_key: "prod-secret"
```

Mapping:
```go
mapper := patternmapper.NewConfigMapper().
    Map("environments.{env}.api_key", "demo", "api-key").
    // Or with capture in target:
    Map("environments.{env}.api_key", "demo", "{env}-api-key").
    Build()
```

### Example 4: Wildcard Matching with Named Capture

Config:
```yaml
services:
  auth:
    api_key: "auth-key"
  payment:
    api_key: "payment-key"
```

Mapping:
```go
mapper := patternmapper.NewConfigMapper().
    Map("services.{service}.api_key", "demo", "{service}-api-key").
    Build()
```

### Example 6: Nested Rules (Best Practice)

Config:
```yaml
app:
  settings:
    api_key: "secret"
    threshold: 42
    timeout: 30
  database:
    host: "localhost"
    port: 5432
```

Mapping with nested rules and layer override:
```go
mapper := middlewares.NewConfigMapper().
    MapObject("app.settings", "demo", []patternmapper.MappingRule{
        {Source: "api_key", TargetParameter: "api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
        {Source: "timeout", TargetParameter: "timeout", Default: 30},
    }).
    MapObject("app.database", "demo", []patternmapper.MappingRule{
        // Child rules inherit "demo" layer from parent
        {Source: "host", TargetParameter: "db-host"},
        // Override target layer for this specific rule
        {Source: "port", TargetLayer: "database", TargetParameter: "port"},
    }).
    Build()
```

YAML equivalent:
```yaml
- source: "app.settings"
  target_layer: "demo"
  rules:
    - source: "api_key"
      target_parameter: "api-key"
    - source: "threshold"
      target_parameter: "threshold"
    - source: "timeout"
      target_parameter: "timeout"
      default: 30

- source: "app.database"
  target_layer: "demo"
  rules:
    - source: "host"
      target_parameter: "db-host"  # Uses parent's "demo" layer
    - source: "port"
      target_layer: "database"      # Overrides to "database" layer
      target_parameter: "port"
```

This creates:
- `demo.api-key = "secret"`
- `demo.threshold = 42`
- `demo.timeout = 30`
- `demo.db-host = "localhost"`
- `database.port = 5432` (in different layer due to override)

### Example 7: Nested Rules with Captures (Capture Environment)

Config:
```yaml
environments:
  dev:
    settings:
      api_key: "dev-secret"
      threshold: 10
  prod:
    settings:
      api_key: "prod-secret"
      threshold: 100
```

Mapping with MapObject:
```go
mapper := patternmapper.NewConfigMapper().
    MapObject("environments.{env}.settings", "demo", []patternmapper.MappingRule{
        // Child rules inherit {env} from parent in their capture environment
        {Source: "api_key", TargetParameter: "{env}-api-key"}, // Uses parent capture
        {Source: "threshold", TargetParameter: "threshold"},
    }).
    Build()
```

**Capture Environment Flow**:
1. Parent rule `"environments.{env}.settings"` matches:
   - `environments.dev.settings` → Capture environment: `{env: "dev"}`
   - `environments.prod.settings` → Capture environment: `{env: "prod"}`

2. For each match, extract the settings object:
   - Dev: `{api_key: "dev-secret", threshold: 10}`
   - Prod: `{api_key: "prod-secret", threshold: 100}`

3. Process child rules with inherited capture environment:
   - Child rule `source="api_key", target="{env}-api-key"`:
     - With `env="dev"` → creates `demo.dev-api-key = "dev-secret"`
     - With `env="prod"` → creates `demo.prod-api-key = "prod-secret"`
   - Child rule `source="threshold", target="threshold"`:
     - Creates `demo.threshold = 10` (from dev) or `100` (from prod)

**Result**:
- `demo.dev-api-key = "dev-secret"` (from `environments.dev.settings.api_key`)
- `demo.prod-api-key = "prod-secret"` (from `environments.prod.settings.api_key`)
- `demo.threshold = 10` or `100` (depending on which environment matched)

YAML equivalent showing capture environment:
```yaml
- source: "environments.{env}.settings"  # Parent captures {env}
  target_layer: "demo"
  rules:
    # Child rules execute with parent's capture environment:
    # - First match: {env: "dev"}, processes {api_key: "dev-secret", threshold: 10}
    # - Second match: {env: "prod"}, processes {api_key: "prod-secret", threshold: 100}
    - source: "api_key"
      target_parameter: "{env}-api-key"  # {env} available from parent capture environment
    - source: "threshold"
      target_parameter: "threshold"
```

Config:
```yaml
app:
  dev:
    settings:
      api:
        key: "secret"
  prod:
    settings:
      api:
        key: "secret"
```

Mapping:
```go
mapper := patternmapper.NewConfigMapper().
    Map("app.**.api.key", "demo", "api-key").
    Build()
```

### Example 9: Multiple Captures in Nested Rules

Config:
```yaml
regions:
  us-east:
    dev:
      settings:
        api_key: "us-east-dev-key"
        threshold: 10
    prod:
      settings:
        api_key: "us-east-prod-key"
        threshold: 100
  us-west:
    dev:
      settings:
        api_key: "us-west-dev-key"
        threshold: 20
```

Mapping with MapObject and multiple captures:
```go
mapper := patternmapper.NewConfigMapper().
    MapObject("regions.{region}.{env}.settings", "demo", []patternmapper.MappingRule{
        // Child rules inherit BOTH {region} and {env} from parent capture environment
        {Source: "api_key", TargetParameter: "{region}-{env}-api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
    }).
    Build()
```

**Capture Environment Flow**:
1. Parent rule `"regions.{region}.{env}.settings"` matches multiple times:
   - `regions.us-east.dev.settings` → Capture environment: `{region: "us-east", env: "dev"}`
   - `regions.us-east.prod.settings` → Capture environment: `{region: "us-east", env: "prod"}`
   - `regions.us-west.dev.settings` → Capture environment: `{region: "us-west", env: "dev"}`

2. For each match, extract the settings object and process child rules with inherited captures:
   - Match 1: `{region: "us-east", env: "dev"}`, settings: `{api_key: "us-east-dev-key", threshold: 10}`
     - Creates: `demo.us-east-dev-api-key = "us-east-dev-key"`
   - Match 2: `{region: "us-east", env: "prod"}`, settings: `{api_key: "us-east-prod-key", threshold: 100}`
     - Creates: `demo.us-east-prod-api-key = "us-east-prod-key"`
   - Match 3: `{region: "us-west", env: "dev"}`, settings: `{api_key: "us-west-dev-key", threshold: 20}`
     - Creates: `demo.us-west-dev-api-key = "us-west-dev-key"`

**Key Point**: All captures from the parent rule (`{region}`, `{env}`) are available in the capture environment for all child rules, allowing them to combine multiple captures in their target parameters.

### Example 10: Lambda Transformation (Programmatic API)

Config:
```yaml
services:
  auth:
    api_key: "auth-key"
    port: 8080
  payment:
    api_key: "payment-key"
    port: 8081
```

Mapping with lambda transformation:
```go
mapper := patternmapper.NewConfigMapper().
    MapWithTransform("services.{service}.api_key", func(captures map[string]string, positional []string, value interface{}) (string, string) {
        service := captures["service"]
        // Dynamic layer based on service name
        layer := fmt.Sprintf("%s-service", service)
        // Parameter name always "api-key"
        param := "api-key"
        return layer, param
    }).
    MapWithTransform("services.{service}.port", func(captures map[string]string, positional []string, value interface{}) (string, string) {
        service := captures["service"]
        // Use service name as layer
        layer := service
        // Parameter name includes service
        param := "port"
        return layer, param
    }).
    Build()
```

This creates:
- `auth-service.api-key = "auth-key"`
- `payment-service.api-key = "payment-key"`
- `auth.port = 8080`
- `payment.port = 8081`

**Lambda Function Signature**:
```go
type TransformFunc func(
    captures map[string]string,  // Named captures: {"env": "dev", "service": "auth"}
    positional []string,          // Positional captures: ["dev", "auth"] (for {0}, {1})
    value interface{},            // The actual config value being mapped
) (targetLayer string, targetParameter string)
```

**Use Cases for Lambda Transformation**:
- Complex naming logic based on captures
- Conditional layer selection
- Dynamic parameter name generation
- Custom validation or filtering
- Accessing the actual value for conditional logic

**Note**: Lambda transformations are only available in the programmatic API (Go code). YAML/JSON config files use string-based capture references.

### Usage in Middleware

```go
// Option 1: Direct usage
middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigFileMapper(
        patternmapper.NewConfigMapper().
            Map("app.settings.api_key", "demo", "api-key").
            Build(),
    ),
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)

// Option 2: Load from config file
middleware := middlewares.LoadParametersFromFile(
    "config.yaml",
    middlewares.WithConfigFileMapperFromFile("mappings.yaml"),
    middlewares.WithParseOptions(
        parameters.WithParseStepSource("config"),
    ),
)
```

### Helper Functions

```go
// Load mapper from YAML/JSON file
mapper, err := middlewares.LoadMapperFromFile("mappings.yaml")

// Create mapper from rules
mapper := patternmapper.NewConfigMapper(rules...)

// Add rules programmatically
mapper.AddRule(patternmapper.MappingRule{
    Source: "app.settings.api_key",
    TargetLayer: "demo",
    TargetParameter: "api-key",
})
```

## Advanced Features

### Default Values

```go
rules := []patternmapper.MappingRule{
    {
        Source: "app.settings.threshold",
        TargetLayer: "demo",
        TargetParameter: "threshold",
        Default: 10, // Use if source not found
    },
}
```

### Required Fields

```go
rules := []patternmapper.MappingRule{
    {
        Source: "app.settings.api_key",
        TargetLayer: "demo",
        TargetParameter: "api-key",
        Required: true, // Fail if not found
    },
}
```

### Conditional Mapping

```go
rules := []patternmapper.MappingRule{
    {
        Source: "app.settings.api_key",
        TargetLayer: "demo",
        TargetParameter: "api-key",
        Condition: func(config map[string]interface{}) bool {
            // Only apply if condition is met
            return config["environment"] == "production"
        },
    },
}
```

### Transform Functions

```go
rules := []patternmapper.MappingRule{
    {
        Source: "app.settings.api_key",
        TargetLayer: "demo",
        TargetParameter: "api-key",
        Transform: func(value interface{}) (interface{}, error) {
            // Transform value before mapping
            if str, ok := value.(string); ok {
                return strings.ToUpper(str), nil
            }
            return value, nil
        },
    },
}
```

## Error Handling

### Missing Source Path

- If `Required: false` (default): Use default value or skip
- If `Required: true`: Return error

### Pattern Matching Errors

- Return error if pattern syntax is invalid
- Log warnings for patterns that match nothing

**Note**: Type conversion and validation are handled by parameter definitions, not by the mapper. The mapper extracts values as-is from the config file, and parameter definitions handle type conversion and validation during parsing.

## Implementation Considerations

### Performance

- Compile patterns into efficient matchers (regex or custom tree)
- Cache compiled patterns
- Early exit on exact matches

### Path Traversal

- Use depth-first search for `**` patterns
- Limit depth to prevent infinite loops
- Support circular reference detection

### Memory

- Lazy evaluation of patterns
- Don't load entire config into memory if not needed
- Stream processing for large configs

## Alternatives Considered

### 1. JSONPath/JMESPath

**Pros**: Standard, well-known syntax
**Cons**: Less flexible for our use case, adds dependency

### 2. Template System

**Pros**: Very flexible
**Cons**: More complex, harder to validate

### 3. Regex Patterns

**Pros**: Powerful
**Cons**: Complex syntax, harder to read for nested structures

## Recommendation

Start with **Option 2 (Rules Array)** for programmatic API, then add **Option 3 (YAML/JSON Config)** support for declarative configuration. This provides:
- Flexibility for programmatic use
- Declarative configuration for end users
- Type safety and validation
- Easy to extend with advanced features

## Migration Path

1. Implement basic pattern matching (exact paths, wildcards)
2. Add capture groups
3. Add defaults and required fields
4. Add layer inheritance and override
5. Add advanced features (conditionals, transforms)

## Questions to Resolve

1. Should we support array mapping? (e.g., `items[*].key` → map to multiple parameters?)
2. Should we support multiple matches mapping to same parameter? (merge behavior?)
3. Should we support nested target layers? (e.g., `target.layer: "demo.nested"`?)
4. Performance: How deep can we go before it's too slow?
5. Should we support validation rules? (e.g., regex validation, min/max for numbers)

**Note**: Type conversion is handled by parameter definitions, not by mapping rules. The mapper extracts values as-is from the config file, and parameter definitions handle type conversion and validation.

## Builder API (Phase 2) — Intern Handoff

Goals:
- Provide a fluent, ergonomic way to assemble `[]MappingRule` without changing core semantics.
- Reuse existing strict behavior (ambiguous wildcards/collisions error by default), prefix-aware errors, and compile-time validation for static targets.

Scope (Phase 2 builder only):
- Support for `Source`, `TargetLayer`, `TargetParameter`, `Rules`, `Required`.
- One-level nested rules (parent → children).
- No positional captures, no TransformFunc, no YAML loader, no deep or array wildcards.

Proposed API:
```go
type ConfigMapperBuilder struct {
    layers *layers.ParameterLayers
    rules  []patternmapper.MappingRule
}

func NewConfigMapperBuilder(layers *layers.ParameterLayers) *ConfigMapperBuilder

// Map adds a simple leaf rule. If required is provided and true, sets Required.
func (b *ConfigMapperBuilder) Map(source string, targetLayer string, targetParameter string, required ...bool) *ConfigMapperBuilder

// MapObject adds a parent rule with children (one-level nesting)
func (b *ConfigMapperBuilder) MapObject(parentSource string, targetLayer string, childRules []patternmapper.MappingRule) *ConfigMapperBuilder

// Build validates via NewConfigMapper and returns a ConfigMapper
func (b *ConfigMapperBuilder) Build() (middlewares.ConfigMapper, error)
```

Behavioral Notes:
- `Build()` MUST call `NewConfigMapper(b.layers, b.rules...)` to reuse validation and semantics.
- Static targets (no captures) are validated at Build-time (compile), prefix-aware.
- Ambiguous wildcards (distinct values → same param) and cross-rule collisions MUST error.
- Capture shadowing warning is emitted during Build when applicable.

Helper for child rules (optional):
```go
func Child(source, target string) patternmapper.MappingRule { return patternmapper.MappingRule{Source: source, TargetParameter: target} }
```

Example usage:
```go
b := NewConfigMapperBuilder(paramLayers).
    Map("app.settings.api_key", "demo", "api-key").
    MapObject("app.{env}.settings", "demo", []patternmapper.MappingRule{
        {Source: "api_key", TargetParameter: "{env}-api-key"},
        {Source: "threshold", TargetParameter: "threshold"},
    })

mapper, err := b.Build()
if err != nil { /* handle */ }
```

Test Plan:
- Map(): adds expected rule; Required flag behavior.
- MapObject(): children inherit parent layer; one-level enforced.
- Build():
  - Invalid pattern syntax → error.
  - Invalid static target parameter (prefix-aware) → error.
  - Ambiguous wildcard (distinct values) → error; identical values OK.
  - Cross-rule collision → error.
  - Prefix-aware errors include "checked as" only when prefixed name differs.

Acceptance Criteria:
- API compiles; examples runnable.
- Unit tests cover all behaviors above.
- No changes to `ConfigMapper` interface or `NewConfigMapper` semantics.

