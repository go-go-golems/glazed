## Validating Glazed Config Files (layers and pattern mappers)

Date: 2025-11-03

### Purpose & scope

Goal: add validation for configuration files used with Glazed, covering:
- Layer-based validation: ensure config keys align with the command’s parameter layers and definitions, including type checks and requireds.
- Pattern-based mapper validation: ensure mapping rules and their application to config files are valid (captures, targets, collisions).

This note summarizes existing behavior, feasibility, and concrete, low-effort ways to add validation (including wiring in `main.go`).

### What exists today

- Pattern mapper compile-time checks (on `NewConfigMapper` / builder `Build()`):
  - Pattern syntax (segments, wildcards, named captures)
  - Target layer existence
  - Capture references in target parameters must exist in the source pattern
  - Static target parameter names are checked early with layer prefix-awareness (errors include canonical name)

- Pattern mapper runtime checks (when mapping a raw config):
  - Required patterns: fail with a helpful “nearest existing path / missing segment / available keys” message
  - Parameter existence (prefix-aware) for dynamic target names
  - Multi-match ambiguity: same rule matches different values for one target parameter → error
  - Cross-rule collisions: two rules writing the same parameter → error
  - Deterministic traversal (lexicographic) for stable behavior

- Default file loader (no mapper):
  - Loads layer maps (layerSlug → { paramName: value }) and merges into parsed layers
  - Unknown layers/keys are silently ignored: only defined parameters are read; there is no error for extra keys
  - Type validation happens when values are applied (parameter definition `CheckValueValidity`)
  - Required parameters are enforced only where parsing functions consider them; the config loader itself passes `onlyProvided=true`, so it won’t error on missing requireds by itself

### Layer-based validation strategies

Add a “strict validation” pass for the default structure (no mapper):
- Unknown layer detection: error if a layer key in the config is not present in `ParameterLayers`.
- Unknown parameter detection: within known layers, error on any key that does not match a parameter (considering layer prefix-canonicalization if desired).
- Type checking: call `ParameterDefinition.CheckValueValidity` for each provided value to catch type/enum issues without mutating state.

Required handling:
- Required parameters are enforced by the parameter system during parsing/merging. Config-only validation does not need to re-enforce required. Optionally, you can add a post-parse assertion that requireds are satisfied across all sources (see helper below).

Effort: small. These are straightforward loops over `layers.ParameterLayers`, `ParameterDefinitions`, and `ParsedLayers`.

### Proposed API: Config validation helpers (layer-structured files)

Constraints: keep it easy for apps to validate their config files; allow soft/hard handling of extraneous layers and extraneous settings within a layer. No generic schema validation is attempted beyond what layers already define.

```go
// Severity for validation findings
type ValidationMode int

const (
    ValidationIgnore ValidationMode = iota
    ValidationWarn
    ValidationError
)

// Policy controlling validation behavior for default-structured config files
type ConfigValidationPolicy struct {
    // Unknown layer keys at top-level (not present in ParameterLayers)
    UnknownLayers ValidationMode // default: ValidationIgnore (back-compat)

    // Unknown parameter keys inside a known layer
    UnknownParameters ValidationMode // default: ValidationWarn

    // Validate provided values against parameter types/enums
    TypeCheck bool // default: true
}

// Single-file/raw-object validation for default structure
// raw is a YAML/JSON-unmarshaled object (map[string]interface{} at top-level)
func ValidateRawConfigAgainstLayers(
    layers *layers.ParameterLayers,
    raw interface{},
    policy ConfigValidationPolicy,
) ([]ValidationIssue, error)

// Multi-file convenience helper; aggregates issues and returns error if any
// issue with mode ValidationError occurred under the provided policy.
func ValidateConfigFilesAgainstLayers(
    layers *layers.ParameterLayers,
    files []string,
    policy ConfigValidationPolicy,
) ([]ValidationIssue, error)

// ValidationIssue describes a single finding; caller may log or surface it.
type ValidationIssueKind string

const (
    IssueUnknownLayer     ValidationIssueKind = "unknown-layer"
    IssueUnknownParameter ValidationIssueKind = "unknown-parameter"
    IssueTypeMismatch     ValidationIssueKind = "type-mismatch"
)

type ValidationIssue struct {
    Kind      ValidationIssueKind
    File      string            // optional
    Layer     string            // slug, for layer-related findings
    Parameter string            // parameter name, for param-related findings
    Message   string            // human-readable detail
}
```

Middleware integration (optional, for ease-of-use):

```go
// Extend config file options to support validation for default-structured files
func WithConfigValidation(policy ConfigValidationPolicy) middlewares.ConfigFileOption

// Behavior:
// - If no mapper is set, loader will unmarshal raw -> run ValidateRawConfigAgainstLayers
//   -> log warnings / return error per policy -> then apply values.
// - If mapper is set, loader will not run layer-structure validation; apps should validate
//   by calling mapper.Map(raw) in a dedicated validation step (see pattern-mapper section).
```

### Pattern-mapper validation strategies

At startup (compile-time):
- Construct the mapper (or load it via `patternmapper.LoadMapperFromFile`); this already validates syntax, layers, capture references, static targets. If it fails here, it’s a fast fail in `main.go`.

Validate config files (runtime):
- For each config file, unmarshal raw YAML/JSON and call `mapper.Map(raw)`. Discard the result if you’re doing a “validate-only” pass. Any ambiguity, collisions, missing required matches, or unknown target parameters will return an error.

Overlay semantics (multiple files):
- Option A (simple): Validate each file independently. This is strict for `Required` rules (they must match per file) and is easy to wire.
- Option B (overlay-aware): Validate “across the overlay” by concatenating raw configs into a single composite raw object (or run `Map` per file and aggregate matches) so `Required` is satisfied if any later file provides it. This is more involved if you want a true raw-object merge; a practical alternative is: run `Map` on each file and track if any file produced at least one match for every required pattern.

Effort: minimal for per-file validation; moderate for overlay-aware requireds if you want to merge raw config trees.

### Wiring in main.go (low-effort options)

- Validate mapper at startup (fast-fail):
  - Build or load a mapper as part of command construction; exit if invalid.

- Validate config files before executing command logic:
  - If you have a resolver producing files (low → high), unmarshal each and call `mapper.Map(raw)`. For default structure (no mapper), run the strict-keys/type pass instead.
  - Provide a flag like `--validate-config-only` that runs validation and exits 0/1.

Example (pseudocode):

```go
func validateConfigFilesStrict(layers *layers.ParameterLayers, files []string, mapper middlewares.ConfigMapper) error {
    for _, f := range files {
        raw, err := readYAMLOrJSON(f)
        if err != nil { return err }

        if mapper != nil {
            // Pattern-mapper validation path (does not mutate state)
            if _, err := mapper.Map(raw); err != nil { return fmt.Errorf("%s: %w", f, err) }
            continue
        }

        // Default structure: { layerSlug: { paramName: value } }
        m, ok := raw.(map[string]interface{})
        if !ok { return fmt.Errorf("%s: expected object at top-level", f) }
        for layerSlug, v := range m {
            layer, exists := layers.Get(layerSlug)
            if !exists { return fmt.Errorf("%s: unknown layer %q", f, layerSlug) }
            kv, ok := v.(map[string]interface{})
            if !ok { return fmt.Errorf("%s: layer %q must map to an object", f, layerSlug) }
            pds := layer.GetParameterDefinitions()
            for key, val := range kv {
                canonical := key
                if prefix := layer.GetPrefix(); prefix != "" && !strings.HasPrefix(key, prefix) {
                    canonical = prefix + key
                }
                pd, ok := pds.Get(canonical)
                if !ok || pd == nil { return fmt.Errorf("%s: unknown parameter %q in layer %q", f, key, layerSlug) }
                if _, err := pd.CheckValueValidity(val); err != nil {
                    return fmt.Errorf("%s: invalid value for %s.%s: %w", f, layerSlug, key, err)
                }
            }
        }
    }
    return nil
}
```

Post-parse requireds (overlay-aware, across all sources):

```go
func validateRequiredAfterParse(layers_ *layers.ParameterLayers, parsed *layers.ParsedLayers) error {
    missing := []string{}
    _ = layers_.ForEachE(func(_ string, l layers.ParameterLayer) error {
        parsedL := parsed.GetOrCreate(l)
        pds := l.GetParameterDefinitions()
        return pds.ForEachE(func(p *parameters.ParameterDefinition) error {
            if p.Required {
                if _, ok := parsedL.Parameters.Get(p.Name); !ok {
                    missing = append(missing, fmt.Sprintf("%s.%s", l.GetSlug(), p.Name))
                }
            }
            return nil
        })
    })
    if len(missing) > 0 {
        return fmt.Errorf("missing required parameters: %v", missing)
    }
    return nil
}
```

Hook into `main.go` (conceptually):

```go
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // Build or load mapper here (fast-fail validates rules)
    mapper, err := patternmapper.LoadMapperFromFile(layers, mappingRulesPath)
    if err != nil { return err }

    files, err := resolver(parsedCommandLayers, cmd, args)
    if err != nil { return err }
    if err := validateConfigFilesStrict(layers, files, mapper); err != nil { return err }
    return nil
}
```

### Complexity & effort estimate

- Pattern mapper validation: mostly done. Compile-time checks exist; runtime “validate-only” is trivial (call `Map(raw)` on each file). Low effort to wire into `main.go`.
- Layer-based strict validation (default structure): small function (~50–100 LOC) to check unknown layers/keys and type-check values; very low risk.
- Post-parse required validation: small function (~30–50 LOC). Low effort.
- Overlay-aware requireds for mappers (single-pass): Moderate if you need true raw-merge; simple if you accept “any file satisfies required”.
- JSON Schema path (optional, future): medium; would require generating a schema for layers and using a validator.

### Recommended next steps

1) Add a `validateConfigFilesStrict(...)` helper in your app (or as a middleware option like `WithStrictConfigValidation`).
2) Build/load the mapper in `main.go` so compile-time validation runs at startup.
3) Optionally add `--validate-config-only` to run validation and exit.
4) Add a final “requireds after parse” check to guarantee correctness across all sources.


