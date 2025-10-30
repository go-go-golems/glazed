# Generic Config Mapping Design: Review and Recommendations

**Date**: 2025-10-29  
**Reviewer**: Design Review Committee  
**Status**: Design Analysis and Recommendations

---

## Executive Summary

The Generic Config Mapping Design proposes a declarative, pattern-based system for mapping arbitrary config file structures to Glazed's layer-based parameter system. This review synthesizes the design document, debate discussions, and architectural concerns to provide concrete recommendations for implementation.

**Key Recommendation**: Proceed with a phased implementation starting with Rules Array API, avoiding YAML support initially, and maintaining ConfigFileMapper as a first-class option for complex cases.

---

## Problem Statement Analysis

### Current State
- `ConfigFileMapper` requires custom Go functions for each config format
- Each implementation is slightly different (error handling, nested structures, etc.)
- No standardization or consistency across implementations
- Verbose boilerplate for simple mappings

### Proposed Solution
- Declarative pattern-based mapping system
- Standardized syntax for common cases
- Supports both programmatic and declarative (YAML) configuration

### Validation
The problem is **real and well-defined**. The debate showed consensus that standardization is needed, even if the solution approach differs.

---

## Core Design Decisions

### 1. Capture Semantics

**Design Decision**: Wildcards (`*`) match but don't capture. Named captures (`{name}`) or positional captures (`{0}`) are required.

**Analysis**:
- **Pros**: Consistent with regex patterns, performance (no unnecessary allocations), explicit
- **Cons**: Non-discoverable, requires explanation, cognitive overhead

**Recommendation**: **SIMPLIFY**
- **Primary option**: Use named captures only (`{name}`). Remove positional captures (`{0}`) from initial design.
- **Rationale**: 
  - Reduces cognitive load
  - More discoverable (`{env}` is clearer than `{0}`)
  - Covers 90% of use cases
  - Can add positional captures later if needed
- **Wildcards**: Keep `*` for matching without capture, but add clear documentation that to use the matched value, you must use named captures

**Example**:
```go
// ✅ Good: Named capture
Map("app.{env}.api_key", "demo", "{env}-api-key")

// ❌ Remove: Positional capture (too confusing)
Map("app.*.api_key", "demo", "{0}-api-key")

// ✅ Keep: Wildcard for matching only
Map("app.*.api_key", "demo", "api-key")  // All environments map to same param
```

### 2. Nested Rules

**Design Decision**: Parent rules can have child rules that inherit captures and layer targets.

**Analysis**:
- **Pros**: DRY, groups related mappings, captures inheritance
- **Cons**: Complexity, performance concerns (O(n²) potential), tight coupling to config structure

**Recommendation**: **INCLUDE, but with constraints**
- **Include**: Nested rules provide significant value for grouped mappings
- **Constraints**:
  - Limit nesting depth to 2 levels (parent → child, no deeper)
  - Require explicit documentation of capture inheritance
  - Optimize by caching parent matches
- **Performance**: Profile early - if performance is an issue, add a "flatten" option that converts nested rules to flat rules

**Example**:
```go
// ✅ Good: One level of nesting
MapObject("app.settings", "demo", []MappingRule{
    {Source: "api_key", TargetParameter: "api-key"},
})

// ⚠️ Consider: Two levels (require justification)
MapObject("app.{env}.settings", "demo", []MappingRule{
    {Source: "api_key", TargetParameter: "{env}-api-key"},
})

// ❌ Avoid: Three+ levels (too complex)
```

### 3. API Design

**Design Decision**: Three options - Builder Pattern, Rules Array, YAML Config.

**Analysis**:
- **Builder Pattern**: Fluent API, good developer experience, but harder to serialize/debug
- **Rules Array**: Type-safe, serializable, debuggable, supports dynamic config
- **YAML Config**: Most declarative, but loses type safety, runtime errors

**Recommendation**: **PHASED APPROACH**
1. **Phase 1**: Implement Rules Array API only
   - Type-safe
   - Compile-time validation
   - Serializable (can add YAML loader later)
   - Clear and explicit
2. **Phase 2**: Add Builder Pattern convenience API (optional wrapper)
   - Can be implemented as a thin wrapper around Rules Array
   - Provides fluent API for simple cases
   - Converts to Rules Array internally
3. **Phase 3**: YAML support (if needed)
   - Validate pattern syntax at load time
   - Provide schema validation
   - Consider using Rules Array internally, loading from YAML

**Code Structure**:
```go
// Phase 1: Rules Array (core)
type MappingRule struct { ... }
func NewConfigMapper(rules ...MappingRule) *ConfigMapper

// Phase 2: Builder Pattern (convenience)
type ConfigMapperBuilder struct { ... }
func NewConfigMapper() *ConfigMapperBuilder

// Phase 3: YAML (declarative)
func LoadMapperFromYAML(data []byte) (*ConfigMapper, error)
```

### 4. Default Values

**Design Decision**: Mapping rules can specify defaults.

**Analysis**:
- **Concern**: Parameter definitions already have defaults
- **Conflict**: Two sources of defaults = confusion

**Recommendation**: **CLARIFY SEMANTICS**
- **Removed**: `Default` field from `MappingRule`
- **Rationale**: 
  - Defaults belong in parameter definitions (single source of truth)
  - Mapping rules map VALUES, not defaults
  - If a config value is missing, the parameter definition default applies
- **Exception**: If environment-specific defaults are needed, use `TransformFunc` or `ConfigFileMapper` (complex cases)

**Precedence** (no mapping rule defaults):
1. Mapped value from config file
2. Parameter definition default
3. Error if required

### 5. Error Handling

**Design Decision**: `Required` flag determines behavior when pattern doesn't match.

**Analysis**:
- Need clear error messages
- Need observability (which patterns matched/failed)
- Need flexibility (dev vs prod behavior)

**Recommendation**: **STRICT but OBSERVABLE**
- **Required: true**: Return error if pattern doesn't match
- **Required: false**: Skip silently (no error), but log at DEBUG level
- **Error messages**: Must include:
  - Pattern that failed
  - Config file paths checked
  - Expected structure
  - Actual structure found (if available)
- **Observability**: Log all pattern match attempts (success/failure) at DEBUG level
- **Production**: Strict errors (no silent failures), but detailed logging for debugging

**Example**:
```go
// Required: true - must match
{
    Source: "app.settings.api_key",
    Required: true,  // Error if not found
}

// Required: false - optional
{
    Source: "app.optional.feature",
    Required: false,  // Skip silently if not found, log at DEBUG
}
```

### 6. Array Handling

**Design Decision**: Mentioned but not fully specified.

**Analysis**:
- Arrays are common in config files
- Multiple ways to handle: map to array param, map to multiple params, flatten
- Complex to encode in pattern syntax

**Recommendation**: **DEFER or LIMIT**
- **Phase 1**: Support simple array indexing (`app.settings[0].key`)
- **Defer**: Wildcard array matching (`app.services[*].name`)
- **Reason**: Complex semantics, unclear use cases
- **Alternative**: Use `ConfigFileMapper` for array handling (complex cases)

**If needed later**:
- Explicit syntax: `app.services[*].name` → `service-names` (array parameter)
- Or: `app.services[{index}].name` → `service-{index}-name` (multiple parameters)
- Must validate parameter type matches array type

### 7. TransformFunc

**Design Decision**: Programmatic API only, not in YAML.

**Analysis**:
- Provides flexibility for complex cases
- Requires Go code (type safety)
- Boundary between simple and complex cases

**Recommendation**: **KEEP as PROGRAMMATIC ONLY**
- **Keep**: TransformFunc is valuable for edge cases
- **Keep**: Programmatic only (no YAML support)
- **Documentation**: Make it clear TransformFunc is for advanced use cases
- **Guideline**: If you need TransformFunc, consider if ConfigFileMapper is simpler

**Clear Boundary**:
- Simple mappings: Pattern matching
- Complex mappings: ConfigFileMapper function
- Very complex mappings: TransformFunc (rare)

### 8. Middleware Integration

**Design Decision**: Unclear - replace or enhance `LoadParametersFromFile`?

**Analysis**:
- Two code paths = complexity
- Need backward compatibility
- Same output format (layer maps)

**Recommendation**: **ENHANCE EXISTING MIDDLEWARE**
- **Approach**: Enhance `LoadParametersFromFile` to accept either:
  - `ConfigFileMapper` function (existing)
  - `ConfigMapper` (pattern matching) (new)
- **Interface**: Create a `ConfigMapper` interface:
  ```go
  type ConfigMapper interface {
      Map(rawConfig interface{}) (map[string]map[string]interface{}, error)
  }
  ```
- **Implementation**: 
  - `ConfigFileMapper` implements it (adapter)
  - Pattern matching mapper implements it
  - `LoadParametersFromFile` checks type and uses appropriate path
- **Backward Compatible**: Existing code continues to work

**Code**:
```go
func LoadParametersFromFile(filename string, options ...ConfigFileOption) Middleware {
    // ... existing code ...
    
    // Support both ConfigFileMapper and ConfigMapper
    if mapper, ok := opts.Mapper.(ConfigMapper); ok {
        // Pattern matching path
    } else if fn, ok := opts.Mapper.(ConfigFileMapper); ok {
        // Function path (existing)
    }
}
```

### 9. Migration Path

**Design Decision**: How to migrate existing ConfigFileMapper users?

**Analysis**:
- Can't break existing code
- Need clear guidelines
- Optional migration

**Recommendation**: **COEXISTENCE MODEL**
- **Not Migration**: Frame as "coexistence" not "migration"
- **Clear Guidelines**: 
  - Use pattern matching for: 5 rules or less, simple mappings
  - Use ConfigFileMapper for: Complex logic, arrays, transformations
- **Documentation**: Clear examples of when to use each
- **Tools**: Optional helper to convert simple ConfigFileMapper functions to pattern rules (for reference, not forced)

**Boundary Examples**:
```go
// ✅ Pattern matching: Simple, 5 rules or less
rules := []MappingRule{
    {Source: "app.settings.api_key", TargetLayer: "demo", TargetParameter: "api-key"},
    {Source: "app.settings.threshold", TargetLayer: "demo", TargetParameter: "threshold"},
}

// ✅ ConfigFileMapper: Complex logic
mapper := func(raw interface{}) (map[string]map[string]interface{}, error) {
    // Complex transformation logic
    // Arrays, conditionals, transformations
}
```

---

## Implementation Recommendations

### Phase 1: Core Pattern Matching (MVP)

**Scope**:
- Rules Array API only
- Named captures only (`{name}`)
- Simple patterns (exact match, wildcard, named capture)
- Nested rules (one level only)
- Required flag
- Basic error handling

**Exclude**:
- Positional captures (`{0}`)
- TransformFunc
- YAML support
- Array handling beyond indexing
- Default values in rules
- Deep wildcards (`**`)

**Deliverables**:
- `MappingRule` struct
- `ConfigMapper` type with pattern matching
- Integration with `LoadParametersFromFile`
- Comprehensive tests
- Documentation with examples

### Phase 2: Convenience APIs

**Scope**:
- Builder Pattern wrapper (optional convenience)
- Better error messages
- Observability (logging)

**Deliverables**:
- `ConfigMapperBuilder` (wrapper around Rules Array)
- Improved error messages
- Debug logging for pattern matches

### Phase 3: Advanced Features (if needed)

**Scope**:
- TransformFunc (programmatic only)
- YAML support (with validation)
- Array wildcards (if needed)
- Positional captures (if needed)

**Deliverables**:
- TransformFunc support
- YAML loader with schema validation
- Extended documentation

---

## Architectural Concerns

### 1. Type Safety

**Concern**: Pattern matching happens at runtime, but parameter definitions are at compile time.

**Recommendation**:
- Validate pattern syntax at mapper creation time (not runtime)
- Validate parameter names exist in target layer at mapper creation time
- Runtime: Validate mapped values match parameter types

**Code**:
```go
func NewConfigMapper(layers *layers.ParameterLayers, rules ...MappingRule) (*ConfigMapper, error) {
    // Validate patterns
    for _, rule := range rules {
        if err := validatePattern(rule.Source); err != nil {
            return nil, err
        }
        // Validate parameter exists
        if err := validateParameterExists(layers, rule.TargetLayer, rule.TargetParameter); err != nil {
            return nil, err
        }
    }
    // ...
}
```

### 2. Performance

**Concern**: Pattern matching adds overhead compared to direct function calls.

**Recommendation**:
- Compile patterns into efficient matchers (regex or custom tree)
- Cache compiled patterns
- Early exit on exact matches
- Profile early and optimize
- Benchmark against ConfigFileMapper

**Metrics**:
- Pattern compilation time
- Pattern matching time per config file
- Memory usage
- Comparison with ConfigFileMapper equivalent

### 3. Error Messages

**Concern**: Pattern matching errors can be cryptic.

**Recommendation**:
- Detailed error messages with context
- Show expected vs actual structure
- Include config file path in errors
- Provide suggestions for common mistakes

**Example Error**:
```
Pattern "app.{env}.api_key" didn't match:
  Expected: app.<env>.api_key (e.g., app.dev.api_key, app.prod.api_key)
  Config file structure:
    app:
      dev:
        api_key: "found"  ✓
      prod:
        api_key: "found"  ✓
      staging:
        api_key: "found"  ✓
  Note: Pattern matched, but capture "env" cannot be used in target "{env}-api-key"
```

### 4. Observability

**Concern**: Need to debug which patterns matched/failed.

**Recommendation**:
- Log all pattern match attempts at DEBUG level
- Include: pattern, matched paths, capture values, result
- Provide debugging helper to visualize pattern matching
- Consider structured logging for pattern matching events

### 5. Backward Compatibility

**Concern**: Must not break existing ConfigFileMapper users.

**Recommendation**:
- ConfigFileMapper remains fully supported
- Pattern matching is additive, not replacement
- Clear documentation that both are valid options
- No deprecation warnings or pressure to migrate

---

## Questions to Resolve Before Implementation

### 1. Pattern Syntax Validation
- When should pattern syntax be validated? (Creation time vs runtime)
- Should invalid patterns fail fast or log warnings?

**Recommendation**: Validate at creation time, fail fast.

### 2. Missing Parameters
- What if a pattern matches a value but the target parameter doesn't exist?
- Create it dynamically or error?

**Recommendation**: Error - parameters must exist in layer definitions (type safety).

### 3. Multiple Matches
- What if a pattern matches multiple values? (e.g., `app.*.api_key` matches dev and prod)
- Create multiple parameters or error?

**Recommendation**: Create multiple parameters (one per match) - this is the intended behavior.

### 4. Precedence
- If multiple patterns match the same parameter, which wins?

**Recommendation**: Error - ambiguous mappings should be explicit. Require unique target parameters per layer.

### 5. Deep Wildcards
- Should `**` (deep wildcard) be supported in Phase 1?

**Recommendation**: Defer to Phase 2. Complex to implement correctly, unclear use cases.

---

## Success Criteria

### Phase 1 Success
- ✅ Can map simple config files (3-5 mappings) without writing Go code
- ✅ Type-safe (compile-time validation where possible)
- ✅ Clear error messages
- ✅ Backward compatible with ConfigFileMapper
- ✅ Performance acceptable (< 2x overhead vs ConfigFileMapper)

### Long-term Success
- ✅ Reduced boilerplate for common cases
- ✅ Standardized mapping approach
- ✅ Clear boundaries between simple and complex cases
- ✅ Developers choose the right tool for their use case

---

## Recommendations Summary

### Must Have (Phase 1)
1. ✅ Rules Array API
2. ✅ Named captures only (`{name}`)
3. ✅ Simple patterns (exact, wildcard, named capture)
4. ✅ Nested rules (one level)
5. ✅ Required flag
6. ✅ Integration with LoadParametersFromFile
7. ✅ Type safety (validate at creation time)

### Should Have (Phase 2)
1. ⚠️ Builder Pattern convenience API
2. ⚠️ Better error messages
3. ⚠️ Observability/logging

### Nice to Have (Phase 3)
1. ⚪ TransformFunc (programmatic only)
2. ⚪ YAML support (with validation)
3. ⚪ Array wildcards
4. ⚪ Positional captures

### Exclude
1. ❌ Default values in mapping rules (use parameter definitions)
2. ❌ Positional captures in Phase 1 (too confusing)
3. ❌ Deep wildcards in Phase 1 (too complex)
4. ❌ Array wildcards in Phase 1 (defer)

---

## Conclusion

The Generic Config Mapping Design addresses a real need for standardization in config file mapping. The phased approach recommended here balances:

- **Simplicity**: Start with the essential features
- **Type Safety**: Validate at creation time where possible
- **Flexibility**: Keep ConfigFileMapper as first-class option
- **Backward Compatibility**: No breaking changes
- **Clear Boundaries**: When to use pattern matching vs ConfigFileMapper

The key is to **avoid over-engineering** while providing value for common cases. Start simple, validate with real use cases, then extend based on feedback.

**Next Steps**:
1. Finalize Phase 1 design based on this review
2. Create implementation plan
3. Prototype core pattern matching
4. Validate with real config files
5. Iterate based on feedback

