---
Title: "Design: appconfig.Parser (register layers + settings structs, configurable Parse middlewares)"
Ticket: CONFIG-PARSER-001
Status: active
Topics:
    - glazed
    - config
    - api-design
    - parsing
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/runner/run.go
      Note: ParseCommandParameters + ParseOption pattern (configures middleware chain for Parse)
    - Path: glazed/pkg/cmds/middlewares/middlewares.go
      Note: ExecuteMiddlewares ordering/precedence mechanics
    - Path: glazed/pkg/cmds/parameters/initialize-struct.go
      Note: ParsedLayers.InitializeStruct hydration path used by v1 (glazed.parameter tags)
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParserConfig provides a similar “configure middlewares” seam (future CLI integration)
ExternalSources: []
Summary: "Propose an appconfig.Parser API that is instantiated with a grouped AppSettings struct type T, lets callers register (settings struct ↔ ParameterLayer) pairs, configures parsing middlewares via options, and exposes Parse() that returns a filled T."
LastUpdated: 2025-12-16T00:00:00Z
---

## Context (why a new design doc)

The previous design doc (`01-design-struct-first-configparser-api-on-top-of-glazed.md`) drifted into a “struct-first schema + nested YAML path mapping” direction and got stuck on prefix/mapping semantics. That path may still be valuable later, but it is **not** the API shape we want to build next.

This document proposes a **simpler, incremental `appconfig.Parser`**:

- It does **not** attempt to derive layers from structs yet.
- Instead, it focuses on a clean façade around existing Glazed primitives:
  - register `ParameterLayer`s and their corresponding typed settings structs,
  - configure the **middleware chain** used for parsing via options,
  - return a fully-populated grouped `AppSettings` struct `T` from `Parse()`.

## Goals

- **G1: Grouped settings struct**: instantiate the module with a final grouped `AppSettings` struct type `T` (generic).
- **G2: Registration**: register `(settings struct ↔ ParameterLayer)` pairs.
  - “Struct-first layer generation” is **explicitly deferred**.
- **G3: Configurable parsing**: `New*` accepts options that configure the **middlewares** used in `Parse()`.
- **G4: Typed result**: `Parse()` returns `(*T, error)` (filled settings struct).
- **G5: Minimal Glazed surface in app code**: only bootstrap-level code should import Glazed/layers/middlewares.

## Non-goals (for this design)

- **NG1**: automatic struct → ParameterDefinition/Layer derivation (postponed).
- **NG2**: nested YAML path mapping / schema mappers / prefix semantics changes (postponed).
- **NG3**: Cobra-first UX (`ToCobraCommand`, etc.). We can add a CLI adapter later, but v1 centers on a library-style `Parse()`.

## Core API proposal

### High-level shape

We introduce a type:

- `appconfig.Parser[T]`

Assuming we want `import appconfig "github.com/go-go-golems/glazed/pkg/config"` (as used in docs), we can place it in `glazed/pkg/config` and expose it as `Parser[T]`.

### Constructor

```go
ac := appconfig.NewParser[AppSettings](
  appconfig.WithEnv("MYAPP"),                 // configures env middleware
  appconfig.WithConfigFiles("base.yaml"),     // configures config-files middleware
  appconfig.WithMiddlewares(custom...),       // injects extra middlewares (optional)
)
```

Design intent:

- The options are **declarative**, not “middleware order knowledge”.
- `Parser` owns the ordering to preserve a consistent precedence model.

### Registration

We register a `ParameterLayer` and how to hydrate into `T`.

Key design constraint: we want the grouped `T` to be the thing we return, but the registered structs are nested fields inside `T`.

So we propose a “binder” function:

```go
ac.Register(
  "redis", redisLayer,
  func(t *AppSettings) any { return &t.Redis },  // pointer to nested struct
)
```

Notes:

- `any` is used so we can support pointers to different struct types under a single generic `T`.
- In v1, the pointed-to struct is expected to use `glazed.parameter` tags (or whatever `ParsedLayers.InitializeStruct` requires today).

Alternative registration variants (optional, later):

- `RegisterLayer(slug string, layer layers.ParameterLayer, bind func(*T) any)`
- `Register(reg Registration[T])` where `Registration` is a small struct:
  - `Slug string`
  - `Layer layers.ParameterLayer`
  - `Bind func(*T) any`

### Parse

```go
cfg, err := ac.Parse()
if err != nil { ... }
// cfg is *AppSettings
```

Semantics:

1. Build a `layers.ParameterLayers` collection from all registered layers.
2. Execute a middleware chain (configured via options) to populate `ParsedLayers`.
3. For each registration:
   - call `parsedLayers.InitializeStruct(reg.Slug, reg.Bind(&t))`
4. Return the populated `t`.

## Precedence / middleware configuration model

We need a clear rule: options configure *which sources* are active, but AppConfig decides ordering.

### Baseline precedence (v1)

For the initial library-style parser, we mirror the runner precedence:

- defaults (lowest)
- config files (low→high)
- env (highest among these)
- (values-for-layers / programmatic overrides) — optional, if we include it

We can add “flags/args” later as part of a CLI adapter (either via CobraParserConfig or direct cobra middlewares).

### Reuse existing ParseOption patterns

We should leverage the existing `cmds/runner.ParseCommandParameters` ParseOption pattern:

- `runner.WithEnvMiddleware(prefix)`
- `runner.WithConfigFiles(files...)`
- `runner.WithAdditionalMiddlewares(...)`
- `runner.WithValuesForLayers(...)`

Two implementation strategies:

1) **Delegate to runner** (recommended for v1):
   - Build a tiny “command stub” whose `Description().Layers` are the registered layers.
   - Call `runner.ParseCommandParameters(stubCmd, runnerOptions...)`.

2) **Direct middleware execution**:
   - Call `cmd_middlewares.ExecuteMiddlewares(layers, parsedLayers, middlewares...)` directly.
   - This duplicates runner’s ordering logic unless we re-expose it.

Given we want option-driven middleware configuration, strategy (1) is the cleanest starting point.

## Hydration model (v1)

Hydration uses existing Glazed mechanics:

- `ParsedLayers.InitializeStruct(layerSlug, destStructPtr)`

This implies (based on how `ParsedParameters.InitializeStruct` is implemented today):

- the registered settings structs must be compatible with `InitializeStruct`,
- **fields are only populated if they have an explicit `glazed.parameter:"name"` tag**,
- and missing parameters are currently **skipped** (no error; field stays at zero value).

Future evolution:

- add a schema-driven hydration that does not require glazed tags (similar to Moments), but that is explicitly out-of-scope for this doc.

## Error handling and invariants

`Parser` should validate at registration time:

- `slug` uniqueness (no duplicate layer slugs).
- `layer` non-nil.
- `bind` non-nil and returns a non-nil pointer at parse time.

At parse time:

- If middleware parsing fails, return the wrapped error.
- If hydration fails for any registered struct, return the error with:
  - layer slug, and
  - a hint that tags/parameter names may not match.

## Open questions (to resolve before coding)

1. **Package placement**:
   - `glazed/pkg/config` (importable as `appconfig`) vs a new `glazed/pkg/appconfig`.
2. **Do we include a “values-for-layers” programmatic override** in v1?
3. **Do we accept “raw ParsedLayers escape hatch”** (probably no, but maybe for debugging)?
4. **Validation hooks**:
   - per-registration validator? (out-of-scope for v1, but likely soon)
5. **CLI integration**:
   - do we add a small adapter in v1 or keep library-only?

## Recommended next step

Implement a v1 prototype focused on:

- `NewParser[T](options...)`
- `Register(slug, layer, bind)`
- `Parse() (*T, error)` using runner ParseOptions

Then iterate based on real Pinocchio usage.


