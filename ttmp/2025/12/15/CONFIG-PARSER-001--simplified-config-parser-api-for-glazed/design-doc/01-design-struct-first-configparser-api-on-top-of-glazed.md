---
Title: 'Design: Struct-first ConfigParser API on top of Glazed'
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
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParserConfig default middleware chain (likely needs mapper hook)
    - Path: glazed/pkg/cmds/middlewares/load-parameters-from-json.go
      Note: Config file loading + ConfigMapper hook (foundation for nested YAML support)
    - Path: glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper.go
      Note: Current prefix semantics mismatch with struct-path config (design discussion)
    - Path: glazed/pkg/cmds/middlewares/update.go
      Note: Env var precedence and env key naming (layer prefix + param name)
    - Path: glazed/pkg/cmds/runner/run.go
      Note: runner.ParseCommandParameters currently lacks config-mapper option (design C2)
    - Path: moments/backend/pkg/appconfig/gather_viper_nested.go
      Note: Prior art for ConfigPath + prefix stripping (config keys unprefixed
    - Path: moments/backend/pkg/appconfig/store.go
      Note: Prior art for schema-driven hydration without glazed.parameter tags
    - Path: pinocchio/cmd/pinocchio/main.go
      Note: Primary motivating example of complex manual Glazed setup
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/analysis/01-glazed-parameter-parsing-architecture-analysis.md
      Note: Architecture analysis of Glazed parameter parsing system
    - Path: glazed/ttmp/2025/12/15/CONFIG-PARSER-001--simplified-config-parser-api-for-glazed/reference/02-research-brainstorm-new-config-api-diary.md
      Note: Lab book with full searches and findings leading to this design
ExternalSources: []
Summary: Propose a struct-first ConfigParser API that hides Glazed layers/middlewares while still using them internally (flags/env/config files → ParsedLayers → typed structs).
LastUpdated: 2025-12-15T08:51:42.421259975-05:00
---


## Status note (superseded)

This design doc reflects an earlier direction for CONFIG-PARSER-001 (“struct-first schema + mapping”). The ticket has since been reset toward an incremental **`appconfig.Parser`** design (explicit layer + settings struct registration, configurable Parse middlewares, `Parse() -> T`).

Current design: `design-doc/02-design-appconfig-module-register-layers-and-parse.md`.

## Context

The motivating pain is that “plain” Glazed usage requires the caller to manually:

- define `ParameterDefinition`s,
- create `ParameterLayer`s,
- assemble a middleware chain in the correct precedence order (defaults < config < env < args < flags),
- and then hydrate typed settings structs from `ParsedLayers`.

In practice (see `pinocchio/cmd/pinocchio/main.go`), this becomes a large amount of glue code, especially once:

- configuration is stored in nested YAML (“real app config”),
- flags must be namespaced to avoid collisions,
- multiple config files are overlaid,
- environment variables override config files,
- and “profiles”/dynamic overlays change which config sources are active.

CONFIG-PARSER-001 aims to introduce a **struct-first** API that keeps Glazed as the underlying engine but hides the boilerplate from application code.

For the full “lab book” of searches and findings that led here, see:

- `reference/02-research-brainstorm-new-config-api-diary.md`

## Goals

- **G1: Struct-first**: user defines settings in Go structs (nested structs allowed).
- **G2: Hide Glazed plumbing**: app code should not have to touch `ParameterDefinition`, `ParameterLayer`, `Middleware`, or `ParsedLayers`.
- **G3: Multiple sources**: parse settings from:
  - defaults,
  - config files (including nested YAML),
  - environment variables,
  - flags/args (Cobra).
- **G4: Precedence**: preserve Glazed precedence: defaults < config files < env < args < flags.
- **G5: Cobra-friendly**: easy CLI integration that produces flags + parsing, without bespoke middleware glue.
- **G6: Extensible**: allow custom sources (e.g. vault/credentials provider) and dynamic config file resolution (profiles).

## Non-goals (initial iteration)

- **NG1**: Replace Cobra or rewrite Glazed parsing core.
- **NG2**: A perfect “universal” YAML mapping system; focus on deterministic struct-path mapping.
- **NG3**: Support every possible Go type; start with the common set and grow incrementally.

## Existing primitives we can (and should) reuse

From the Glazed analysis:

- **Definitions / layers / parsing**:
  - `glazed/pkg/cmds/parameters` (`ParameterDefinition`, `ParameterDefinitions`)
  - `glazed/pkg/cmds/layers` (`ParameterLayer`, `ParameterLayers`, `ParsedLayers`)
  - `glazed/pkg/cmds/middlewares` (`SetFromDefaults`, `UpdateFromEnv`, `LoadParametersFromFiles`, `ParseFromCobraCommand`, …)
- **Config mapping**:
  - `glazed/pkg/cmds/middlewares` `ConfigMapper` interface (hook used by `LoadParametersFromFiles`)
  - `glazed/pkg/cmds/middlewares/patternmapper` (declarative mapping rules)
- **CLI integration**:
  - `glazed/pkg/cli` `CobraParser` and `BuildCobraCommandFromCommand`
- **Type hydration**:
  - `glazed/pkg/helpers/reflect.SetReflectValue` (conversion glue used by `InitializeStruct`)

## Prior art: Moments `appconfig`

Moments contains a working “Glazed boundary” (`moments/backend/pkg/appconfig`) that:

- registers schema for settings structs,
- builds Glazed layers from schemas,
- gathers nested config values via a custom step,
- calls `Parse(parsedLayers)` to hydrate typed structs,
- and provides global typed getters.

This is useful evidence and suggests specific mechanics to borrow (schema derivation, nested path support, “prefix is external” semantics).

However, it also illustrates what we likely want to avoid for CONFIG-PARSER-001:

- global registry + side-effect imports,
- global typed store (`Must[T]()` everywhere),
- viper coupling (and global viper swapping).

## Proposed API (Slack hypothesis, adapted)

### API sketch

Target ergonomics (roughly):

```go
type AppSettings struct {
  Redis RedisSettings `appconfig.path:"tools.redis"`
  DB    DBSettings    `appconfig.path:"tools.db"`
}

type RedisSettings struct {
  Host string `appconfig.name:"host"`
  Port int    `appconfig.name:"port"`
}

parser, err := appconfig.NewConfigParser[AppSettings](
  appconfig.FromDefaults(),                    // optional
  appconfig.FromConfigFiles("base.yaml", "..."),
  appconfig.FromEnv("MENTO"),
  appconfig.FromFlags(),                       // cobra flags
)

cmd := parser.ToCobraCommand("serve", func(ctx context.Context, cfg *AppSettings) error {
  // no glazed imports required
  return Serve(ctx, cfg)
})
```

### Interpretation

- “No layers” is an ergonomic goal. Internally we will still build layers because Glazed needs them to:
  - generate Cobra flags,
  - keep parameter namespaces separated,
  - and record parse logs with layer/parameter metadata.

## Key design decisions

### D1: Layering model (internal)

Two plausible internal approaches:

1) **One layer for everything** (flat parameter namespace)
   - parameter names become “flattened paths” like `tools-redis-host`
   - simplest mapping story, but weak help grouping and higher chance of naming conflicts.

2) **One layer per settings struct** (recommended)
   - each “sub-settings struct” becomes a layer (slug derived from field name or path)
   - parameters are leaf keys (`host`, `port`, …) within that layer
   - flags/env are namespaced using `layer.Prefix` (e.g. `tools-redis-`)
   - config files use nested paths (e.g. `tools.redis.host`) and do **not** include the flag prefix.

Option (2) matches both:

- Glazed’s own env logic (`layerPrefix + paramName`),
- and Moments’ explicit semantics (“prefix is for CLI flags, not config file keys”).

### D2: Naming conventions

Proposed defaults (configurable via options):

- **ParameterDefinition.Name**: kebab-case leaf key (derived from Go field name unless overridden)
- **Flag name**: `layerPrefix + paramName`
- **Env var name**: `ENV_PREFIX + "_" + upper(snake(layerPrefix+paramName))`
- **Config key name**: leaf key under the struct’s config path, without layer prefix (e.g. `tools.redis.host`)

### D3: Struct tags and schema providers

We likely need both:

- **Tags** for 80% cases:
  - `appconfig.path:"tools.redis"` (applies to struct fields of struct type)
  - `appconfig.name:"host"` (applies to scalar leaf fields)
  - optional: `appconfig.required:"true"`, `appconfig.help:"..."`, `appconfig.secret:"true"`, etc.

- **Schema provider interface** for 20% cases:
  - lets subsystems provide help/defaults/choices without relying on tags
  - similar to Moments’ explicit `Schema{Fields: ...}`

### D4: Config file mapping strategy

To support nested YAML cleanly, the ConfigParser should generate a **ConfigMapper** from the struct schema:

- For each settings leaf:
  - source path: `<configPath>.<leafName>` (dot-separated)
  - target: `(layerSlug, leafName)`

This mapper must implement the semantics:

- config paths are based on `appconfig.path` (or default),
- and **must ignore** `layer.Prefix` (flag/env namespace).

This is the key mismatch with today’s `patternmapper` prefix behavior (documented in the research diary).

### D5: Middleware chain and precedence

The ConfigParser should reuse existing Glazed middlewares, assembled in the correct precedence:

- defaults
- config files (low → high precedence)
- env
- cobra args/flags

Because Glazed middlewares often “call next first, then apply” (see `ExecuteMiddlewares` docs), the order you pass matters.
ConfigParser should hide this and just guarantee the precedence model.

### D6: Profiles / multi-phase parsing (advanced)

Profiles are a multi-phase parse problem:

- profile selection must be known before the profile-specific files/layers are decided.

ConfigParser should support this via either:

- a dedicated “early parse” for profile settings, followed by full parse, OR
- a “dynamic source” that can compute additional config files after env/flags are applied.

This can be added as an extension point without forcing it into the initial API.

## Modify existing Glazed vs layer on top (recommendations)

### Minimal, high-leverage upstream changes (recommended)

These are small changes that make a ConfigParser implementation much cleaner without breaking core semantics:

- **C1: Allow CobraParser’s config-file middleware to accept a `ConfigMapper`**
  - today, `cli.CobraParserConfig` ultimately routes to a helper that doesn’t accept mapper options
  - adding a mapper hook avoids needing to re-implement cobra parsing glue in ConfigParser

- **C2: Add a `WithConfigMapper`-style option in `cmds/runner.ParseCommandParameters`**
  - runner currently only supports “layer-map” config files; nested YAML requires additional middlewares

- **C3: Introduce a new mapper implementation for struct-path mapping**
  - don’t force `patternmapper` to change semantics (avoid breaking existing behavior)
  - instead add something like `glazed/pkg/cmds/middlewares/schemamapper` that is generated from schema/path metadata

### Larger changes (optional, but would improve accessibility)

- **R1: Clarify naming of `layer.Prefix`**
  - “prefix” is easy to misread as “config key prefix”
  - consider adding an alias API (`WithFlagPrefix`) and documenting it aggressively
  - avoid an immediate breaking rename

- **R2: Provide “batteries included” docs and examples**
  - a new doc topic for struct-first config parsing that includes nested config mapping + env + flags

## Implementation designs (with tradeoffs)

### Design A: Pure wrapper (no Glazed changes)

- **Approach**: implement `appconfig.ConfigParser[T]` that:
  - derives layers from structs
  - runs `ExecuteMiddlewares` directly (custom chain)
  - uses `LoadParametersFromFiles(..., WithConfigMapper(...))` for nested config
  - uses `ParseFromCobraCommand` for flags
- **Pros**: no upstream changes, fastest to prototype.
- **Cons**: duplicates CobraParser glue; likely re-implements config file discovery; harder to integrate with existing Glazed conventions (help output, config path discovery, profiles).

### Design B: Wrapper + minimal upstream changes (recommended)

- **Approach**:
  - add small hooks (C1/C2) upstream
  - implement ConfigParser on top of `cli.CobraParser` and/or `runner.ParseCommandParameters`
  - keep all new complexity in a single “configparser/appconfig” package
- **Pros**: best leverage/maintainability; reuses Glazed’s CLI integration; minimal duplication.
- **Cons**: requires touching Glazed core packages (but small, reviewable diffs).

### Design C: Fix `patternmapper` prefix semantics and generate pattern rules

- **Approach**:
  - modify `patternmapper` to treat `layer.Prefix` as external only
  - generate mapping rules for each struct-path leaf
- **Pros**: unifies around a single mapper mechanism; fewer mapper implementations.
- **Cons**: highest risk of unintended breakage; existing tests explicitly encode current behavior; unclear who relies on it.

## Recommendation

Proceed with **Design B**:

- implement a new struct-first ConfigParser package on top of Glazed,
- add small extension points to `cli.CobraParserConfig` and `runner.ParseCommandParameters`,
- and introduce a purpose-built struct-path mapper (separate from `patternmapper`).

This gives the best path to:

- supporting nested YAML + prefixed flags correctly,
- keeping backwards compatibility,
- and enabling profile-driven multi-phase parsing later.

## Migration strategy (Pinocchio as first adopter)

- [ ] Implement ConfigParser behind a new package (so old code keeps working).
- [ ] Add a small example command in `pinocchio/cmd/examples` showing:
  - nested config path mapping,
  - env override,
  - flag override,
  - typed struct hydration.
- [ ] Convert the most complex section of `pinocchio/cmd/pinocchio/main.go` to ConfigParser.
- [ ] Only once proven, consider upstream docs and refactors.

## Open questions

- How much of “schema” should be inferred vs specified?
- How should defaults be expressed (tags vs schema provider vs explicit default instance)?
- How should unknown config keys be handled (ignore vs strict mode)?
- How should profiles be expressed in the API surface (separate source vs integrated into config-file resolution)?
