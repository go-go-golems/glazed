---
Title: "Diary: appconfig.Parser design redo (register layers + structs, configurable Parse)"
Ticket: CONFIG-PARSER-001
Status: active
Topics:
    - glazed
    - config
    - api-design
    - parsing
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/design-doc/01-design-struct-first-configparser-api-on-top-of-glazed.md
      Note: Prior design that went off-track (kept for history)
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/design-doc/02-design-appconfig-module-register-layers-and-parse.md
      Note: New target design for AppConfig module
    - Path: glazed/pkg/cmds/runner/run.go
      Note: ParseCommandParameters provides ParseOption-driven middleware configuration (baseline for Parse)
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParserConfig shows another “middleware configuration seam” for future CLI adapter
    - Path: glazed/pkg/cmds/middlewares/load-parameters-from-json.go
      Note: Config file middleware accepts WithConfigMapper / WithConfigFileMapper (important option surface)
ExternalSources: []
Summary: "Implementation diary for redesigning CONFIG-PARSER-001 toward an incremental appconfig.Parser: generic grouped settings type, explicit layer+struct registration, and Parse() configured by middleware options."
LastUpdated: 2025-12-16T00:00:00Z
---

# Diary: appconfig.Parser design redo

## Goal

Capture the redesign thinking for CONFIG-PARSER-001: instead of a struct-first schema/mapping API (which got stuck in prefix/mapping semantics), we want an incremental **`appconfig.Parser`** that:

- is instantiated with a grouped settings struct type `T`,
- allows explicit registration of `(settings struct ↔ ParameterLayer)` pairs,
- is configured with options that influence the middlewares used in `Parse()`,
- and returns a filled `T` from `Parse()`.

## Step 1: Reframe the ticket and validate the `appconfig.Parser` seams in Glazed

This step reset the direction of the ticket and checked whether Glazed already has the right primitives to support a “register layers + configurable parse” façade. The goal was to avoid inventing new abstractions when existing, battle-tested seams already exist (`runner.ParseCommandParameters` and `cli.CobraParserConfig`).

**Commit (docs):** N/A — design + diary docs only

### What I did

- Re-read the existing lab book to identify where the previous design went off the intended path:
  - `reference/02-research-brainstorm-new-config-api-diary.md`
- Performed targeted searches to find “option-driven parsing entrypoints” and “middleware configuration seams”:
  - grep: `ParseCommandParameters`, `ParseOption`, `WithConfigFiles`, `WithEnvMiddleware`
  - grep: `CobraParserConfig`, `CobraCommandDefaultMiddlewares`, `NewCobraParserFromLayers`
  - grep: `WithConfigMapper`, `WithConfigFileMapper`, `LoadParametersFromFiles`, `type ConfigMapper`
- Read the key implementation files surfaced by those searches:
  - `glazed/pkg/cmds/runner/run.go`
  - `glazed/pkg/cli/cobra-parser.go`

### Why

The requested API is fundamentally a façade over:

- a set of registered `ParameterLayer`s,
- an option-configured middleware chain that produces `ParsedLayers`,
- and a hydration step that fills typed settings structs.

So we need to ensure those hooks exist cleanly and can be re-used without copying Cobra-specific glue.

### What worked

- **Found an existing ParseOption-driven parsing seam**:
  - `runner.ParseCommandParameters(cmd, options...)` and options like:
    - `WithEnvMiddleware(prefix)`
    - `WithConfigFiles(files...)`
    - `WithAdditionalMiddlewares(...)`
    - `WithValuesForLayers(...)`
- **Confirmed a second, CLI-oriented seam exists**:
  - `cli.CobraParserConfig` can build a default middleware chain (flags/args/env/config/defaults), and also allows overriding `MiddlewaresFunc`.

### What didn’t work

- N/A (this step was discovery + validation; no code changes attempted).

### What I learned

- `appconfig.Parser` is a better immediate fit for the repo because it composes existing pieces:
  - layers are already the unit of definition,
  - middlewares are already the unit of source/precedence wiring,
  - hydration already exists (`ParsedLayers.InitializeStruct`).
- This avoids the earlier trap: trying to solve nested config path mapping + prefix semantics while also designing the public API.

### What was tricky to build

- **Binder vs reflection**: If `T` is the grouped settings struct, registration needs a safe way to point at sub-structs inside `T` without making AppConfig “struct-first” prematurely. A binder function like `func(*T) any { return &t.Redis }` is a clean bridge.
- **Middleware ordering is easy to get wrong**: both runner and CobraParser rely on the “reverse execution” model of Glazed middlewares; the AppConfig options should be declarative (“enable env”, “use these files”), while the module owns ordering.

### What warrants a second pair of eyes

- Validate that using runner parsing for v1 won’t surprise future CLI integration:
  - runner currently covers defaults/env/config/provided-values, but not cobra flags/args.
  - we may want a separate “CLI adapter” layer rather than merging concerns too early.

### What should be done in the future

- Lock down the precise v1 contract for hydration:
  - do we require `glazed.parameter` tags on settings structs, or do we add an alternate hydration path?
- Decide whether `appconfig.Parser` should live in `glazed/pkg/config` (importable as `appconfig`) or in a new `glazed/pkg/appconfig` package.

### Code review instructions

- Start with the new design doc:
  - `design-doc/02-design-appconfig-module-register-layers-and-parse.md`
- Validate the “seams” by reading:
  - `glazed/pkg/cmds/runner/run.go` (ParseOptions + ParseCommandParameters)
  - `glazed/pkg/cli/cobra-parser.go` (CobraParserConfig + middleware construction)

### Technical details

**Searches performed (grep) and key hits**:

- Query: `ParseCommandParameters(` / `type ParseOption` / `WithEnvMiddleware` / `WithConfigFiles`
  - Hit: `glazed/pkg/cmds/runner/run.go` (ParseOptions + ParseCommandParameters + env/config/defaults wiring)
- Query: `WithConfigMapper(` / `WithConfigFileMapper(` / `LoadParametersFromFiles(` / `type ConfigMapper`
  - Hit: `glazed/pkg/cmds/middlewares/load-parameters-from-json.go` (config loader options)
  - Hit: `glazed/pkg/cmds/middlewares/config-mapper-interface.go` (ConfigMapper interface)
- Query: `CobraParserConfig` / `CobraCommandDefaultMiddlewares` / `NewCobraParserFromLayers`
  - Hit: `glazed/pkg/cli/cobra-parser.go` (CLI parsing seam + default chain builder)

### What I’d do differently next time

- Capture the new requirements as a short “API contract” paragraph before reading anything, to avoid drifting back into deeper mapping/prefix debates too early.

## Step 2: Draft the new `appconfig.Parser` design and tighten the v1 hydration contract

This step turned the reframed requirements into a concrete design doc and validated one subtle assumption: whether Glazed’s existing struct hydration fills fields by convention or only via explicit tags. The result is a clearer v1 contract and less ambiguity for implementors.

**Commit (docs):** N/A — design + diary docs only

### What I did

- Drafted the new design doc:
  - `design-doc/02-design-appconfig-module-register-layers-and-parse.md`
- Verified the current semantics of struct hydration:
  - Read: `glazed/pkg/cmds/parameters/initialize-struct.go`
  - Confirmed that fields are only considered if they have a `glazed.parameter` tag, and missing parameters are skipped.

### Why

`appconfig.Parser` is intentionally “not struct-first yet”, which means:

- the caller must provide the `ParameterLayer` definitions, and
- the simplest v1 hydration path is to reuse `ParsedLayers.InitializeStruct(...)`.

We need to document the exact contract so adopters don’t expect “field-name-based magic” that doesn’t exist.

### What worked

- Confirmed the v1 requirement precisely:
  - without `glazed.parameter` tags, fields are ignored by `InitializeStruct`.
- The design doc now states this explicitly, so code + docs won’t drift.

### What didn’t work

- N/A.

### What I learned

- This design is naturally incremental:
  - v1 = “register layers + register tagged structs + parse”
  - later = “derive layers and/or binding info from structs”

### What was tricky to build

- The difference between “struct-first binding” and “struct hydration” is easy to conflate:
  - Glazed has hydration, but it’s tag-driven.
  - The future “struct-first” work is about schema derivation, naming, and automated layer creation.

### What warrants a second pair of eyes

- Ensure the design doc doesn’t accidentally reintroduce the earlier mapping/prefix rabbit hole:
  - keep v1 scope to runner-style sources (defaults/config/env/provided-values),
  - make CLI/flags an adapter story, not a core requirement yet.

### What should be done in the future

- If we want v1 settings structs to be tag-free, we’ll need:
  - a schema-driven hydration path (Moments-like), or
  - a convention-based field binding layer in `appconfig.Parser`.
  - This should be explicitly tracked as a follow-up rather than creeping into v1.

### Code review instructions

- Read `design-doc/02-design-appconfig-module-register-layers-and-parse.md`.
- Validate hydration semantics in:
  - `glazed/pkg/cmds/parameters/initialize-struct.go` (look for `field.Tag.Lookup("glazed.parameter")` and the “continue” behavior on missing params).

### Technical details

**Hydration semantics evidence**:

- In `ParsedParameters.InitializeStruct(...)`, the code checks for:
  - `tag, ok := field.Tag.Lookup("glazed.parameter")`
  - `if !ok { continue }` (no tag → ignored)
  - `parameter, ok := p.Get(options.Name)`
  - `if !ok { continue }` (missing param → ignored)


