---
Title: research-brainstorm-new-config-api-diary
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
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: CobraParserConfig supports AppName+ConfigFilesFunc; likely a base for FromFlags/FromEnv/FromConfigFile
    - Path: glazed/pkg/cmds/middlewares/config-mapper-interface.go
      Note: ConfigMapper interface used by LoadParametersFromFiles
    - Path: glazed/pkg/cmds/middlewares/load-parameters-from-json.go
      Note: Config file loading + ConfigMapper hook
    - Path: glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper.go
      Note: Pattern mapper implementation for nested config shapes
    - Path: glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper_builder.go
      Note: Builder API for pattern mapping rules
    - Path: glazed/pkg/cmds/parameters/cobra.go
      Note: How ParameterDefinitions become Cobra flags (constraints for naming scheme)
    - Path: glazed/pkg/cmds/parameters/initialize-struct.go
      Note: InitializeStruct tag mechanics we’ll reuse or wrap
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: ParameterDefinition/ParameterType core primitives
    - Path: glazed/pkg/cmds/runner/run.go
      Note: Programmatic parsing (ParseCommandParameters) we can reuse for struct-first API
    - Path: glazed/pkg/config/resolve.go
      Note: ResolveAppConfigPath for default config discovery
    - Path: glazed/pkg/doc/topics/24-config-files.md
      Note: Docs on config precedence + mapping; informs new API behavior
    - Path: moments/backend/pkg/appconfig/derive.go
      Note: DeriveSchema + naming strategies (kebab/snake) and Go-type->ParamType mapping
    - Path: moments/backend/pkg/appconfig/gather_viper_nested.go
      Note: 'Demonstrates correct semantics: config keys unprefixed'
    - Path: moments/backend/pkg/appconfig/initialize.go
      Note: InitializeFromConfigFiles shows end-to-end bootstrap in Moments
    - Path: moments/backend/pkg/appconfig/layers.go
      Note: BuildLayers schema->Glazed layers mapping
    - Path: moments/backend/pkg/appconfig/pathmap.go
      Note: effectivePathFor implements per-slug nested config path resolution
    - Path: moments/backend/pkg/appconfig/store.go
      Note: Parse(parsed) hydrates typed structs from ParsedLayers via schema
    - Path: moments/backend/pkg/appconfig/types.go
      Note: Schema/Field types (ConfigPath + Prefix) closely match Slack proposal
    - Path: moments/backend/pkg/appconfig/viper_merge.go
      Note: Multi-file precedence + env override semantics used by Moments
    - Path: moments/docs/backend/appconfig-quickstart.md
      Note: Moments prior art for schema-first typed config on top of Glazed
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-15T08:54:53.970085543-05:00
---


# Research / Lab Book: Brainstorm new config API on top of Glazed

## Goal

Produce a deeply technical analysis of how the Slack-proposed API could be implemented on top of Glazed, whether we should **change existing Glazed APIs** vs **layer on top**, and what (if anything) could be renamed/repackaged for approachability.

This document is intentionally written like a **lab notebook**: every search, file read, and design fork is recorded.

## Slack-proposed API (problem statement / hypothesis under test)

This is my understanding of the Slack dump you provided, expressed as a “working spec” for this ticket.

### What the API is trying to achieve (human intent)

- Only the **top-most app bootstrap** should talk about config parsing mechanics.
- Subsystems should receive **typed settings structs**, not a config registry, not Glazed layers, and not `ParsedLayers`.
- The API should be “LLM friendly” and easy to adopt: minimal boilerplate, obvious defaults, and sensible precedence.

### The proposed shape (as stated in Slack)

1) A root settings struct representing the whole app:

- `type AppSettings struct { Redis RedisSettings ...; DB DBSettings ... }`
- Struct tags would express config location and naming, but without mentioning “glazed”:
  - `appconfig.path:"tools.redis"` (or similar) for nested config paths
  - optional `appconfig.name:"url_foobar"` overrides, otherwise default name derived from field name (snake/kebab conversion)

2) A generic parser constructed once at top-level:

- `configParser := appconfig.NewConfigParser[AppSettings]( ...sources... )`

Sources are composable and explicit, e.g.:

- `appconfig.FromFlags()`
- `appconfig.FromEnv(appconfig.WithPrefix("MENTO_"))`
- `appconfig.FromConfigFile([]string{"local.yaml","base.yaml"})` (low→high precedence)
- `appconfig.FromVault(appconfig.WithOnlyCredentialFields())` (future)

3) A one-liner to produce Cobra command wiring:

- `return configParser.ToCobraCommand(...)`

4) Command run:

- `settings := configParser.Parse()`
- `app := NewApp(settings)`
- `app.Start()`

### Important “simplification” constraints explicitly desired

- No manual middleware chains in app code.
- No explicit layer slugs in app code.
- No need for a config registry.
- The only reference to config parsing concepts should be in the top-level bootstrap (“really top top top level”).
- Losing some “statically enforced help” is acceptable, but an optional schema hook could restore it (e.g. `GetAppConfigSchema()` on structs).

### My interpretation of the hypothesis (what must be true for this to work well)

To make this ergonomic **and** faithful to Glazed fundamentals, the implementation likely needs:

- A struct-schema builder that can:
  - derive stable parameter names for flags/env (kebab/SCREAMING_SNAKE),
  - derive nested config file paths (dot-paths like `tools.redis.host`),
  - map nested config structures into Glazed’s internal layer/parameter updates (probably via `ConfigMapper` + `patternmapper`).
- A default precedence model consistent with existing Glazed docs:
  - Defaults < Config files (low→high) < Env < Positional args < Flags
- Support for “dynamic overlays” (profiles) which require at least a two-phase parse (selection first, overlay second, then re-apply env/flags).

## Working hypothesis (initial)

The Slack API can be implemented **mostly by layering**:

- Use Glazed’s existing parsing primitives:
  - `glazed/pkg/cmds/runner.ParseCommandParameters` for programmatic parsing.
  - `glazed/pkg/cmds/middlewares` for sources (flags/env/config/defaults/profiles).
  - `glazed/pkg/cmds/parameters` for type parsing and validation.
- Add a struct-driven “schema” builder that:
  - generates `ParameterDefinitions` and `ParameterLayers` from `T`,
  - generates config-file mapping rules (likely using `patternmapper`),
  - produces a **clean, single-entry-point API** (e.g. `config.NewConfigParser[T](...)`).

Two big open constraints surfaced immediately:

1. **Env var naming**: the current env middleware cannot work with “dot” keys like `TOOLS.REDIS.HOST` (shell-unfriendly). We likely need a **dash/underscore** naming scheme for env/flags.
2. **Dynamic sources** (profiles): profile selection is a known trap (values must be read *after* env/config/flags are applied). A good new API should support **multi-phase parsing** or **dynamic middleware**.

## Step 1: Create the research diary + establish traceability

I created this separate “research-brainstorm-new-config-api-diary” document because the main `01-diary.md` is more of a ticket progress diary, while this file is an “as-I-think” lab book.

### What I did

- Created doc: `reference/02-research-brainstorm-new-config-api-diary.md`.
- Related the key implementation files I expected to touch/read heavily:
  - `glazed/pkg/cmds/runner/run.go`
  - `glazed/pkg/cmds/middlewares/load-parameters-from-json.go`
  - `glazed/pkg/cmds/middlewares/config-mapper-interface.go`
  - `glazed/pkg/cmds/middlewares/patternmapper/*`
  - `glazed/pkg/config/resolve.go`
  - `glazed/pkg/cli/cobra-parser.go`
  - `glazed/pkg/cmds/parameters/*`
  - `glazed/pkg/doc/topics/24-config-files.md`

### Why

If we implement a new config parser API, reviewers need a crisp map of “which files mattered” and “why”.

### What I learned

- The repo already contains most of the *engine* for the Slack API (config/env middleware, config discovery, config mappers); the missing part is the **struct-first schema builder + user-facing API**.

## Step 2: Inventory of “already existing” primitives that map directly to the Slack API

This step is the “health inspector tour”: identify the pieces that already exist and how closely they match the Slack pseudo-code.

### What I did

#### 2.1 Search: programmatic parsing entrypoints

- Search (grep): look for `ParseOptions`, `WithEnvPrefix`, `ParseAndRun`.
- Files hit:
  - `glazed/pkg/cmds/runner/run.go`

**Result brought forward**:

- `ParseCommandParameters(cmd, ...)` already exists and is essentially “`NewConfigParser().Parse()` for a Glazed `cmds.Command`”.
- It already composes a middleware chain (env, config files, provided values, defaults).

#### 2.2 Read: `glazed/pkg/cmds/runner/run.go`

Key symbols found:

- `ParseCommandParameters(cmd cmds.Command, options ...ParseOption) (*layers.ParsedLayers, error)`
- `WithEnvMiddleware(prefix string)`
- `WithConfigFiles(files ...string)`
- `WithValuesForLayers(values map[string]map[string]interface{})`

Key observation:

- The parsing order inside `ParseCommandParameters` is assembled as:
  - additional middlewares
  - env (if enabled)
  - config files (if any)
  - values-for-layers (if any)
  - defaults

Because Glazed middlewares execute in reverse, the **effective precedence** is:

Defaults < provided-values < config < env < (plus any later-injected middlewares depending on how you add them)

This is almost exactly the precedence described in `glazed/pkg/doc/topics/24-config-files.md`, except the Cobra-flag middleware is not included (runner is library-only).

#### 2.3 Read: config-file middleware + mapping hook

Read file: `glazed/pkg/cmds/middlewares/load-parameters-from-json.go`

Key symbols found:

- `LoadParametersFromFiles(files []string, options ...ConfigFileOption) Middleware`
- `WithConfigFileMapper(mapper ConfigFileMapper) ConfigFileOption`
- `WithConfigMapper(mapper ConfigMapper) ConfigFileOption`

Key observation:

- The config file loader expects default “layer map” structure:

```yaml
layer-slug:
  parameter-name: value
```

- BUT it can accept a custom mapper (`ConfigMapper`) which transforms arbitrary config structure → `map[layerSlug]map[paramName]value`.

This is **the crucial hook** we need to support Slack’s `tools.redis.*` nested style while keeping Glazed’s internal shape stable.

#### 2.4 Read: `ConfigMapper` interface

Read file: `glazed/pkg/cmds/middlewares/config-mapper-interface.go`

Key symbol:

- `type ConfigMapper interface { Map(rawConfig interface{}) (map[string]map[string]interface{}, error) }`

This means we can implement:

- A **custom mapper** generated from our struct schema, OR
- Reuse the existing pattern mapper (below).

#### 2.5 Read: pattern mapper implementation + builder

Read:

- `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper.go`
- `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper_builder.go`
- `glazed/pkg/cmds/middlewares/patternmapper/loader.go`
- `glazed/pkg/cmds/middlewares/patternmapper/exports.go`

Key symbols:

- `type MappingRule struct { Source, TargetLayer, TargetParameter string; Rules []MappingRule; Required bool }`
- `NewConfigMapper(layers *layers.ParameterLayers, rules ...MappingRule) (middlewares.ConfigMapper, error)`
- `ConfigMapperBuilder` fluent builder
- `LoadMapperFromFile(layers, filename)` and mapping file schema (snake_case keys)

Key observation:

- Patterns are dot-path like: `tools.redis.host`, support:
  - exact segments
  - wildcard `*`
  - named capture `{env}`
- Target parameters can reference captures: `"{env}-api-key"`.

Implication:

- A struct schema builder can generate **MappingRule per leaf field**:
  - `Source: "<path>.<leaf>"` (e.g., `tools.redis.host`)
  - `TargetLayer: <internal-layer-slug>`
  - `TargetParameter: <flattened-flag-name>` (e.g., `tools-redis-host`)

This yields nested YAML support “for free” while staying compatible with env vars and CLI flags.

#### 2.6 Read: config discovery helper

Read: `glazed/pkg/config/resolve.go`

Key symbol:

- `ResolveAppConfigPath(appName, explicit string) (string, error)`

Search order (if explicit empty):

1. `$XDG_CONFIG_HOME/<appName>/config.yaml`
2. `$HOME/.<appName>/config.yaml`
3. `/etc/<appName>/config.yaml`

This is already documented in `glazed/pkg/doc/topics/24-config-files.md` and matches the Slack “FromConfigFile default discovery” spirit.

#### 2.7 Read: Cobra integration supports app name + config resolution

Read: `glazed/pkg/cli/cobra-parser.go` (already referenced in the main diary).

Key observation:

- `CobraParserConfig` already has:
  - `AppName`, `ConfigPath`, and `ConfigFilesFunc`
- The “default middlewares func” created in `NewCobraParserFromLayers` already wires:
  - flags, args, env (if AppName set), config (via resolver), defaults

This is extremely close to the Slack pseudo code:

- `FromFlags()` → `ParseFromCobraCommand` + `GatherArguments`
- `FromEnv(WithPrefix(...))` → `UpdateFromEnv`
- `FromConfigFile([...])` → `LoadParametersFromResolvedFilesForCobra`

### Why

Before designing a new API, we must know whether we’re inventing new machinery or just building a façade.

### What worked

- The combination of `LoadParametersFromFiles` + `ConfigMapper` + `patternmapper` is exactly the kind of extension point we need.

### What didn’t work / gaps found

- There is **no** existing “struct → ParameterDefinitions + bindings” builder.
  - There *is* `InitializeDefaultsFromStruct` (struct → defaults), but it requires ParameterDefinitions to already exist.
- We must decide a naming scheme that works across:
  - config file paths (nested tree)
  - CLI flags (kebab-case)
  - env vars (UPPER_SNAKE_CASE)

### What I learned (important)

The repo already contains an internal mini-history of a similar migration (Viper → middlewares):

- `glazed/ttmp/2025-10-29/01-how-to-remove-viper-from-glazed-applications.md`
- `glazed/ttmp/2025-11-04/02-config-file-feature-code-review.md`
- `pinocchio/ttmp/2025-11-18/PIN-20251118/analysis/01-config-and-profile-migration-analysis.md`

These docs highlight two key risks that the new API must avoid:

1. **Config file format mismatch**: most users historically have flat config; Glazed prefers layer-based config unless you provide a mapper.
2. **Profile selection timing bug**: building profile middleware using *default* profile values causes env-selected profiles to not apply (must read profile after env/config).

## Step 3: “Where should this live?” package-level brainstorm (early)

This is an early brainstorm, not a final decision.

### Observation: `glazed/pkg/config` already exists and is already imported as `appconfig` in docs

- `ResolveAppConfigPath` is in `package config` and is commonly imported as:
  - `appconfig "github.com/go-go-golems/glazed/pkg/config"`

Slack pseudo-code uses `appconfig.NewConfigParser[T]`, which strongly suggests:

- Implement `ConfigParser[T]` in `glazed/pkg/config` (so callers can alias import as `appconfig`).

### Potential complication: `glazed/pkg/config` currently contains Viper-based config editing commands

Read: `glazed/pkg/config/cobra-config-command.go`

This file uses `github.com/spf13/viper` for “config editing” commands, which is orthogonal to parsing middlewares.

Implication:

- Putting the new parser in `glazed/pkg/config` is still viable, but:
  - we must keep parser code viper-free (it can be),
  - the package’s conceptual scope becomes “config discovery + editing + parsing façade”.

## Open questions (to resolve before writing the design doc)

1. **Naming & ergonomics**:
   - Is the package name `config` acceptable as “appconfig” alias? (Seems aligned with existing docs.)
   - Do we want `ConfigParser[T]` or `Parser[T]`?

2. **Struct schema representation**:
   - One internal layer (“config”) with flattened param names (simplest)?
   - Multiple layers per subsystem struct (better help grouping)?

3. **Tag strategy**:
   - No tags by default (derive names from field names), optional overrides via `appconfig.name`, `appconfig.path`, etc?
   - Or require explicit tags for everything?

4. **Dynamic overlays**:
   - How do we support profiles (parse base config → decide profile → apply profile overlay)?
   - Do we bake it into the new API (recommended), or leave it to advanced callers?

## Next action (immediately)

Continue the “health inspector tour” for the remaining missing piece:

- Find whether any other packages (Clay/Pinocchio/Geppetto) already have “struct-to-parameters” helpers we can reuse.
- Inspect patternmapper tests to understand edge cases we must support (captures, collisions, required paths).

## Step 4: Investigate “dynamic overlays” (profiles) and middleware tooling for multi-phase parsing

The Slack proposal includes “sources” like config files, env, flags, and potentially Vault. Profiles are a canonical example of a **selection-based overlay**:

- you must first determine *which profile is active* (from env/flags/config),
- then load profile values,
- but still allow env/flags to override those loaded profile values.

This creates a “dependency” between sources that cannot be expressed as a single linear pass without either:

- multiple passes, or
- a source that can run “late” but does not clobber higher precedence.

### What I did

#### 4.1 Read prior investigation: Pinocchio profile failure analysis

Read: `pinocchio/ttmp/2025-11-18/PIN-20251118/analysis/01-config-and-profile-migration-analysis.md`

Key finding brought forward:

- The profile middleware in Geppetto captures profile name **too early** (before env/flags/config parsing), so env-selected profile never takes effect.

This is directly relevant because the new “struct-first config API” must not repeat this pitfall:

- avoid “reading config selection values only from defaults” during middleware construction.

#### 4.2 Inspect existing Glazed middleware utilities for whitelisting / sub-parsing

Read: `glazed/pkg/cmds/middlewares/whitelist.go`

Key symbols:

- `WhitelistLayersHandler`, `WhitelistLayerParametersHandler`
- `WrapWithWhitelistedLayers(...)`
- `WrapWithWhitelistedParameterLayers(...)`

Key capability:

- We can execute a subset of middlewares only on a subset of layers/parameters by cloning layers and running a nested chain.

### Why

Profiles are the most concrete “hard case” we must support if we want this API to be widely usable.

### What I learned / design impact

We likely need a first-class concept in the new API for “selection-based overlays”, e.g.:

- `FromProfiles(...)` or `FromOverlay(SelectBy(...), Load(...))`

Implementation idea (high-level):

1. Apply defaults + base config to all params.
2. Parse only the selection params (e.g. `profile`, `profile-file`) from env/flags.
3. Load profile overlay based on those selection values.
4. Re-apply env/flags for all params (so they override profile).

This ordering is not optional; it is required to prevent “profile overrides env” bugs.

## Step 5: Search for existing “struct → ParameterDefinitions” generators and name-casing helpers

The Slack API wants struct-first. That means we must:

1. infer `ParameterType` from Go field types,
2. generate stable, human-friendly parameter names (kebab/snake),
3. generate help/required/choices metadata via tags or optional schema hooks.

### What I did

#### 5.1 Search: “generate ParameterDefinitions from struct”

Search question (semantic): “Is there any code that generates glazed parameters or ParameterDefinitions from a Go struct type using reflection?”

Result:

- No existing dedicated “struct → ParameterDefinitions” builder was found.
- Closest existing pieces are *tag-driven value mappers*, not schema generators:
  - `glazed/pkg/cmds/parameters/initialize-struct.go` (`InitializeStruct`, `StructToDataMap`)
  - `glazed/pkg/helpers/maps/maps.go` (`GlazedStructToMap` for tags → map)

#### 5.2 Search: case conversion helpers

Search question (semantic): “Where do we convert Go field names to kebab-case or snake_case in this repo?”

Result brought forward:

- Dependency `github.com/iancoleman/strcase` is present in `go.work.sum`, but I did not find current code using it directly.
- Docs generally recommend kebab-case for flags/commands (`glazed/pkg/doc/topics/19-writing-yaml-commands.md` mentions kebab-case).

#### 5.3 Read: parameter type system (needed for type inference)

Read:

- `glazed/pkg/cmds/parameters/parameter-type.go`
- `glazed/pkg/cmds/parameters/parse.go` (beginning section)

Key mapping brought forward (docs in code):

- `ParameterTypeString` → `string`
- `ParameterTypeInteger` → `int`
- `ParameterTypeFloat` → `float64`
- `ParameterTypeBool` → `bool`
- `ParameterTypeDate` → `time.Time`
- `ParameterTypeStringList` → `[]string`
- `ParameterTypeIntegerList` → `[]int`
- `ParameterTypeFloatList` → `[]float64`
- plus file-backed and object-backed types.

### Why

Type inference and naming conventions are the backbone of a “LLM-friendly, boilerplate-free” API. If we can’t infer these reliably, the API will force users back into manual schema declarations.

### What I learned / initial decisions

- We will likely implement our own field-name conversion (or vendor a small helper) to produce kebab-case from `FieldName`.
- Avoid dots in **parameter names** (env vars); use nested YAML + mapper to bridge to dashed keys.
- A minimal V1 can focus on core scalar/list types and postpone advanced file/object types unless demanded.

## Step 6: Prefix semantics mismatch (pattern mapper vs “prefix for CLI/env”) — investigate + decide mitigation

I hit a potential conceptual inconsistency that matters a lot for the new API:

- In *most* of Glazed, `layer.Prefix` is used to prefix **flag names** and **env keys**, while parameter names stay unprefixed inside the layer.
- In the pattern mapper, the implementation and tests strongly imply that a layer prefix is part of the **canonical parameter name** and should be prepended to the target parameter.

If we get this wrong in the new API, config-file loading will silently fail (values mapped to wrong parameter names), or flags/env will get “double prefixes”.

### What I did

#### 6.1 Search: find patternmapper prefix-related tests

- Search (grep): `TestLayerPrefix|PrefixAware|WithPrefix\\("`
- Results:
  - `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper_edge_cases_test.go` (`TestLayerPrefix`)
  - `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper_proposals_test.go` (`TestPrefixAwareErrorMessages`)

#### 6.2 Read: `TestLayerPrefix` evidence

Read: `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper_edge_cases_test.go`

Key excerpts (paraphrased, but check the file for exact code):

- Layer created with prefix:
  - `layers.WithPrefix("demo-")`
- Parameter definitions include the prefix in their names:
  - `"demo-api-key"`, `"demo-threshold"`
- Mapping rule targets unprefixed parameter name:
  - `TargetParameter: "api-key"`
- Expected mapping result uses prefixed parameter key:
  - `"demo-api-key": "secret"`

Conclusion from this test: **patternmapper assumes that if a layer has prefix `demo-`, then the canonical parameter name is `demo-<param>`**.

#### 6.3 Read: patternmapper implementation that enforces this

Read: `glazed/pkg/cmds/middlewares/patternmapper/pattern_mapper.go`

Key symbol:

- `resolveCanonicalParameterName(layer, targetParam)` currently does:
  - if `layer.GetPrefix() != ""` and `targetParam` doesn’t already start with it → return `layer.GetPrefix() + targetParam`

This is used both for:

- compile-time validation (“does target parameter exist in layer?”)
- runtime mapping output keys.

#### 6.4 Contrast: “prefix for flags/env” usage in other parts of Glazed

I needed a concrete comparison point, so I looked for code and docs where prefixes are used “normally” (prefix only affects external names).

Evidence sources:

1) **Docs**: `glazed/pkg/doc/topics/13-layers-and-parsed-layers.md`
   - Example creates a layer with `WithPrefix("config-")` and unprefixed parameters like `"verbose"` and `"output"`.
   - This strongly suggests `Prefix` is meant to affect external names (flags), not internal parameter keys.

2) **Example code**: `glazed/cmd/examples/config-custom-mapper/main.go`
   - Creates a layer with `layers.WithPrefix("demo-")` and parameter names `"api-key"` and `"threshold"` (unprefixed).
   - Its custom config mapper writes into the layer map using unprefixed keys:
     - `result["demo"]["api-key"] = value`
   - This is consistent with how `GatherParametersFromMap(...)` expects to find keys (`pd.Name`) and how Glazed stores parsed values (`p.Name`) regardless of prefix.

3) **Env middleware**: `glazed/pkg/cmds/middlewares/update.go`
   - Constructs env key from `layerPrefix + p.Name`, but writes the parsed value back under the logical parameter key `p.Name`.
   - This again implies prefix is an *external name concern*, not part of internal parameter key.

#### 6.5 Search: are prefixes + patternmapper used together in real code?

I wanted to know if changing patternmapper prefix semantics would be risky.

- Search (multiline grep): `patternmapper[\\s\\S]{0,200}WithPrefix|WithPrefix[\\s\\S]{0,200}patternmapper`
- Result: **no matches** in the repo.

Interpretation:

- Prefix-aware patternmapper behavior appears to be exercised mainly by its own tests and docs, not by production code.

#### 6.6 Search: how is config mapping wired in “real” stacks (Geppetto/Pinocchio)?

- Search (grep): `WithConfigMapper\\(|WithConfigFileMapper\\(`
- Key result brought forward:
  - `geppetto/pkg/layers/layers.go` uses `middlewares.WithConfigFileMapper(configMapper)` with a **custom** mapper (not patternmapper) to filter out non-layer keys like `repositories`.

This matters because it shows a real-world usage pattern:

- **custom mappers** are normal and acceptable for “structural” transformations,
- so a generated custom mapper from a struct schema would be aligned with existing practice.

### Why

I need to decide whether the new ConfigParser API can rely on the pattern mapper as a building block, or whether it should generate a custom mapper (or avoid prefixes).

### What I learned (design impact)

There is a **semantic mismatch** today:

- Pattern mapper treats `layer.Prefix` as part of the canonical parameter key.
- The rest of Glazed treats `layer.Prefix` as an external naming prefix for flags/env.

This directly impacts the new API, because the new API will almost certainly want to prefix flags/env with something like `tools-redis-` while keeping internal leaf names short.

### Mitigation options (not decided yet)

I see four viable paths:

1) **(Layering-only) Avoid layer prefixes** in the new API:
   - Encode full path into parameter names (e.g., `tools-redis-host`) and keep `layer.Prefix = ""`.
   - Pros: no changes to Glazed; patternmapper works; env/flags work; simplest integration.
   - Cons: long repetitive parameter names; less nice help grouping.

2) **(Layering-only) Generate a custom `ConfigFileMapper`** instead of pattern rules:
   - Keep “normal” Glazed prefix usage:
     - `layer.Prefix = "tools-redis-"`, parameter names = `"host"`, `"port"`, ...
   - Generate mapper that writes unprefixed keys into the layer map:
     - `result["tools-redis"]["host"] = <value>`
   - Pros: keeps nice external names + short internal names; no need to modify patternmapper.
   - Cons: requires implementing a generic reflection-based mapper generator (but that’s already required for the struct-first API anyway).

3) **(Change existing code) Fix patternmapper prefix semantics** to align with the rest of Glazed:
   - Stop prepending `layer.GetPrefix()` to target parameter names; treat prefix as external only.
   - Pros: removes inconsistency; patternmapper becomes usable with prefixed layers (common case).
   - Cons: requires code/test/doc changes; possibly breaking for anyone relying on current prefix-aware behavior (though search suggests low real usage).

4) **(Status quo workaround) Use patternmapper only with layers that “bake in” prefix**:
   - Keep `layer.Prefix` empty; define parameters as `"tools-redis-host"` etc.
   - This is basically option (1) but still allows multiple layers.

My tentative lean for the Slack API is **(2)** (custom mapper generator) or **(1)** (no layer prefix) depending on how much we value shorter internal parameter names.

## Step 7: Real-world constraints from Geppetto/Pinocchio (config keys, “excluded keys”, and profile timing)

### What I did

#### 7.1 Read: Geppetto’s “config mapper to exclude non-layer keys”

Read: `geppetto/pkg/layers/layers.go` (config file loading section)

Key observations:

- A custom config mapper function is used to skip keys like `repositories` (handled separately) and only treat map-valued keys as layers.
- This is done by passing `middlewares.WithConfigFileMapper(configMapper)` into `LoadParametersFromFiles`.

This suggests the new API should support:

- excluding certain config subtrees (or ignoring unknown keys),
- and/or providing a “strict mode” that errors on unknown keys (optional).

#### 7.2 Read: Geppetto’s explicit comment on profile timing

Same file (`geppetto/pkg/layers/layers.go`) contains an explicit TODO-ish comment:

- profile name is still read from defaults at middleware construction time
- needs to be read after env/config are applied

This reinforces the earlier “dynamic overlay” conclusion: profile selection is a multi-phase parse problem.

### Why

These are the sorts of “sharp edges” that a top-level ConfigParser API must solve, otherwise it’s not actually simpler than hand-built middleware chains.

## Step 8: Existing design docs on config mapping + multi-file overlays (context for the new API)

### What I did

#### 8.1 Read: multi-config overlay design

Read: `glazed/ttmp/2025-10-29/02-multi-config-overlays-design.md`

Key takeaways:

- Explicitly documents the precedence model:
  - defaults < config[0..n] < env < args < flags
- Explicitly wants every config file tracked as its own parse step with metadata.

This design is effectively realized in current `LoadParametersFromFiles`:

- it iterates files low→high and annotates parse steps with `{config_file, index}` metadata.

#### 8.2 Read: generic config mapping design + review

Read:

- `glazed/ttmp/2025-10-29/03-generic-config-mapping-design.md`
- `glazed/ttmp/2025-10-29/05-config-mapping-review-and-recommendations.md`

Key takeaways that influence the new API:

- Prefer named captures over wildcards (avoid ambiguous mappings).
- Keep mapping logic declarative for simple cases, but keep `ConfigFileMapper` as first-class for complex transforms.
- Avoid “defaults in mapping rules” — defaults should come from parameter definitions, not mapping.

This supports the idea that the new struct-first API should:

- generate parameter-definition defaults (if it supports defaults),
- and generate mapping rules (or a mapper) purely as “value routing”.

## Step 9: Mapping parsed values into typed structs (reflection conversion helper)

### What I did

Read: `glazed/pkg/helpers/reflect/reflect.go`

Key symbol:

- `reflect.SetReflectValue(dst reflect.Value, src interface{}) error`

Key observations:

- This helper already handles a lot of “messy” numeric/string conversions and overflows.
- `parameters.ParsedParameters.InitializeStruct` already uses this helper for setting struct fields.

Design implication:

- If the new API must populate user structs without `glazed.parameter` tags, it can likely reuse this helper for conversion, but it still needs its own field-binding/tag interpretation.

## Step 10: Moments `appconfig` as prior art (and what it teaches us)

You asked me to look at the Moments `appconfig` implementation because it is “a bit messy”. This is extremely relevant: it’s a real-world attempt at exactly the sort of “Glazed boundary” you described in Slack — except it made different tradeoffs (global registry + global getters).

### What I did (searches + results for traceability)

#### 10.1 Search: locate the Moments `appconfig` package in this repo

- Search (file tree): list `/moments`
  - Result: `moments/backend/pkg/appconfig/` exists and contains ~18 `.go` files.

- Search (glob): `**/pkg/appconfig/**` under `moments/`
  - Result: files like:
    - `backend/pkg/appconfig/{types.go,registry.go,layers.go,store.go,derive.go,...}`

- Search (grep): `package appconfig`
  - Result: confirms those files all belong to `package appconfig`.

#### 10.2 Search: find the public API surface described in the Moments quickstart

- Search (grep): `RegisterSchema[`, `BuildLayers(`, `Parse(parsed`, `Must[`, `ResetForTests`, `SetForTests`
  - Result: many references, both in code and in Moments internal docs under `moments/ttmp/`.

### What I read (primary sources)

#### 10.3 Read: docs quickstart (Moments)

Read: `moments/docs/backend/appconfig-quickstart.md`

Key points (paraphrase):

- `RegisterSchema[T](Schema)` defines a Glazed-agnostic schema.
- `BuildLayers()` converts schemas into Glazed layers.
- `Parse(parsed)` hydrates all registered settings from parsed layers and stores them globally.
- Domain code uses `Must[T]()` and never touches `ParsedLayers`.
- Registration is done via init() side effects and blank imports.

This is conceptually adjacent to the Slack goal (“only bootstrap knows about config”), but it centralizes config access behind a global typed store.

#### 10.4 Read: core Moments implementation (code)

I read these files and recorded the key findings:

1) `moments/backend/pkg/appconfig/types.go`
   - Defines `Schema` and `Field` and a small param-type enum (`ParamString`, `ParamInt`, ...).
   - `Schema` includes:
     - `ConfigPath []string` (nested YAML path, defaulting to `[]string{Slug}`)
     - `Prefix string` (flag-prefix help, “avoid naming conflicts between layers”)

2) `moments/backend/pkg/appconfig/registry.go`
   - Implements a **global registry** of registrations:
     - either schema-based (`RegisterSchema[T]`) or legacy factory-based (`Register[T]` with layer factory)

3) `moments/backend/pkg/appconfig/layers.go`
   - `BuildLayers()` builds Glazed layers from registered schemas, then includes factory-built layers.
   - `buildLayerFromSchema()` maps appconfig `Field` to Glazed `ParameterDefinition`.
   - Uses `layers.WithPrefix(s.Prefix)` if schema prefix is set.

4) `moments/backend/pkg/appconfig/store.go`
   - This is the heart of the “Glazed boundary”:
     - `Parse(parsed *glazedLayers.ParsedLayers) error`
       - hydrates *every registered type*, validates via `Validator`,
       - stores in a global map keyed by reflect.Type.
   - It contains a **schema-driven hydration path** that does *not* require `glazed.parameter` tags:
     - `initializeStructFromSchema(...)` maps struct fields to schema param names using:
       1) `appcfg:"<param-name>"` tag
       2) fallback: kebab-case of the field name (via `toKebabCase`)
     - then reads values from the parsed layer map and assigns via `reflect.SetReflectValue`.
   - Then it does a second best-effort hydration:
     - `_ = parsed.InitializeStruct(r.slug, ptr)` (to allow glazed tags during migration).
   - It also stores a `globalParsedLayers` pointer (“escape hatch”).

5) `moments/backend/pkg/appconfig/derive.go`
   - Contains the “missing piece” we were looking for in Glazed:
     - reflection-based schema derivation:
       - `DeriveSchema[T](slug, ...)`
       - `DeriveSchemaFromTags[T](slug, ...)`
   - Includes a `NamingStrategy` (`kebab`, `snake`, `camel`) and conversion helpers `toKebabCase`, `toSnakeCase`.
   - Maps Go types to logical ParamTypes (`string`, `bool`, ints, slices).

6) `moments/backend/pkg/appconfig/pathmap.go`
   - Implements `effectivePathFor(slug, schema)`:
     1) explicit overrides
     2) schema `ConfigPath`
     3) default `[slug]`
   - This is how Moments binds “nested YAML path” to each schema slug.

7) Viper-specific glue (important because it highlights why the code is “messy”)
   - `moments/backend/pkg/appconfig/initialize.go`:
     - `InitializeFromViper(v *viper.Viper)` builds layers, gathers values from viper via a middleware, then calls `Parse(...)`.
     - `InitializeFromConfigFiles(envPrefix, configPaths)` uses `NewMergedViperYAML(...)` then calls `InitializeFromViper`.
   - `moments/backend/pkg/appconfig/viper_merge.go`:
     - `NewMergedViperYAML(filesLowToHigh, envPrefix)` merges YAML + env override (dash/dot -> underscore).
   - `moments/backend/pkg/appconfig/gather_viper_nested.go`:
     - Builds `slugToPrefix` where “prefix” is actually a nested YAML dot-path derived from `Schema.ConfigPath`.
     - Then gathers per-layer values from viper under that dot-prefix.
     - Critically: it implements **prefix-stripping** for config file keys:
       - “layer prefix is only for CLI flags, not config file keys”
       - It creates “unprefixed definitions” for Viper lookup and remaps keys back to the original names.
     - It swaps the *global* Viper instance temporarily and restores it afterwards (this is part of the “messy” feeling).

8) Config selection helpers
   - `moments/backend/pkg/appconfig/config_paths.go`: resolves base.yaml + env.yaml + local.yaml patterns for Moments’ own repo layout.
   - `moments/backend/pkg/appconfig/config_env.go`: reads `MOMENTS_CONFIG_ENV` (centralized env read).

9) Tests that confirm semantics (useful for us as “empirical validation”)
   - `moments/backend/pkg/appconfig/viper_merge_test.go`
     - Confirms merge order: later files override earlier ones.
     - Confirms env overrides win over merged files using env var naming:
       - example: `TESTAPP_SERVICES_MENTO_SERVER_PORT` overrides `services.mento.server.port`.
     - Confirms the viper instance stores `meta.env_prefix` for downstream gathering.
   - `moments/backend/pkg/appconfig/gather_viper_nested_test.go`
     - Confirms `Schema.ConfigPath` is used as a dot-prefix for config lookup (e.g. `alpha.beta.foo`).
     - Confirms default ConfigPath is `[]string{Slug}` when omitted (e.g. `unit-test-b.bar`).
     - Confirms env naming semantics for nested config paths:
       - `TESTAPP2_ALPHA_BETA_FOO` overrides `alpha.beta.foo`
       - `TESTAPP2_UNIT_TEST_B_BAR` overrides `unit-test-b.bar`

### Why this matters for CONFIG-PARSER-001

Moments `appconfig` provides evidence that the following ideas are not hypothetical — they work in a real system:

1) **Schema-first config**: a Glazed-agnostic schema can be converted into Glazed layers and hydrated back into typed structs.
2) **Typed hydration without glazed tags**: schema + reflection can populate user structs without requiring `glazed.parameter` tags everywhere.
3) **Nested config paths**: it’s feasible to support nested YAML paths per subsystem (`Schema.ConfigPath`), independent of CLI naming.
4) **Correct prefix semantics**: Moments explicitly encodes the desired semantics:
   - `layer.Prefix` is for flags (external), not for config key names (internal).

### What feels “messy” (and what we should avoid in the new API)

From a design purity perspective, the messy parts are:

- **Global registry + side-effect imports** (`RegisterSchema` in init() + blank imports).
- **Global store** (`Must[T]()` everywhere), which violates the Slack desire “no appconfig reference except top-level”.
- **Global parsed layers escape hatch** (`Parsed()`), which reintroduces Glazed coupling in places that reach for it.
- **Viper coupling**:
  - Merging YAML + env is done via Viper, then translated into parsed layers.
  - `GatherFromViperWithSchema` swaps global viper state (hard to reason about).

### What I think we should borrow (strong candidates)

Even if we don’t borrow code directly, we should borrow these patterns:

- `Schema{Slug, Fields, ConfigPath, Prefix}` concept (Slack’s `appconfig.path` is essentially `ConfigPath`).
- `DeriveSchema` + `NamingStrategy` + `toKebabCase`/`toSnakeCase`.
- Schema-driven hydration that does not require `glazed.parameter` tags.
- Explicit handling of “prefix only affects flags/env” (the config path uses unprefixed names).

### Immediate implication for our earlier “prefix mismatch” issue

Moments’ Viper gather logic is a strong argument that **patternmapper’s current prefix behavior is likely not the semantics we want** for the new API.

If we adopt Slack-style nested config paths + prefixed flags, then:

- config keys should be unprefixed leaf names under the nested path
- prefix should only affect flag/env names

This points to either:

- generating a custom mapper (struct-path → unprefixed param keys), OR
- changing patternmapper prefix semantics to align with Moments + general Glazed behavior.

## Step 11: Synthesis brainstorm (what the “struct-first ConfigParser” likely needs to be)

I’m switching from “inventory and inspection” into “design synthesis”. This is still part of the lab book: capturing hypotheses, design options, and where I think the real traps are.

### 11.1 The key friction points the new API must eliminate

From Glazed and real usage (Pinocchio + Moments), the pain comes from needing to:

- decide a parameter naming scheme that works for **flags**, **env vars**, and **config files**,
- create **layers** (or fake “no layers” but still implement layering internally),
- register and order the **middleware chain** correctly (defaults / config / env / flags),
- and (optionally) handle **multi-phase parsing** for profile-driven overlays.

Any ConfigParser API that doesn’t hide these details will still feel “complex”.

### 11.2 Strong hypothesis: internal parameter names should stay unprefixed; prefixes are external

Two independent lines of evidence suggest this:

- Glazed’s env middleware builds env keys from `layerPrefix + p.Name` and assumes `p.Name` is the canonical key.
- Moments explicitly treats “prefix is for CLI flags, not config file keys” and strips it for nested config gathering.

So for the new API, the likely correct model is:

- **ParameterDefinition.Name**: canonical internal leaf name (e.g. `host`, `database`, `slack-bot-token`).
- **Layer.Prefix**: external namespace for flags/env keys (e.g. `redis-`, `tools-redis-`).
- **ConfigPath** (new concept at ConfigParser/schema layer): nested YAML path prefix (e.g. `tools.redis`).

### 11.3 Patternmapper mismatch: we probably need a different config mapper for struct-path configs

Given the above, `patternmapper`’s current behavior (“expects parameter definitions to include the prefix in their names”) is a semantic mismatch for “ConfigPath + Prefix” systems.

This suggests three options:

- **Option A (no upstream changes)**: generate a custom `ConfigMapper` that walks nested YAML and maps:
  - `tools.redis.host` → layer `redis`, param `host`
  - (and ignores layer prefix entirely for config files)
- **Option B (upstream change)**: adjust `patternmapper` so it treats `layer.Prefix` as external-only, and never prepends it to canonical parameter names.
  - Risk: breaks existing tests and maybe existing users relying on the current behavior.
- **Option C (compatible migration)**: keep existing patternmapper but add a *new* mapper implementation (call it `pathmapper`/`schemamapper`) with the semantics we want.
  - This keeps old behavior for existing code and avoids breaking changes.

Right now, **Option C** looks like the safest path for a generic library.

### 11.4 API surface brainstorming (Slack goal vs Moments goal)

There are (at least) two distinct “top-level ergonomics” we can aim for:

1) **Slack-style injection (preferred for CONFIG-PARSER-001)**:
   - config is parsed at the top level and passed around as structs,
   - no global getters/registry.

2) **Moments-style global typed store**:
   - config is parsed once and stored globally,
   - domain code does `appconfig.Must[T]()` anywhere.

The new Glazed API can support (1) without supporting (2).
But it might still be worth providing an optional helper for (2), since Moments demonstrates it is pragmatically useful.

### 11.5 Cobra integration: likely needs a “builder” that hides layers + parser config

To be drop-in for CLI apps, ConfigParser probably needs either:

- `ToCobraCommand()` that directly returns a cobra command, OR
- a way to expose generated Glazed layers + a `CobraParserConfig` to re-use `cli.BuildCobraCommandFromCommand`.

I suspect we want:

- a high-level `ToCobraCommand(run func(ctx, cfg) error)` convenience,
- and a lower-level “export layers + middlewares” escape hatch for advanced apps (profiles).

### 11.6 Dynamic profiles are a “multi-phase parse” problem (still unsolved)

We saw earlier (Pinocchio + Geppetto) that profile settings must be known before you can resolve:

- which config files to load,
- which extra layers/flags become active.

This means the ConfigParser might need:

- a pre-parse phase for “early layers” (profile selection),
- then a second parse phase for full layers.

This is feasible with Glazed primitives (two `ExecuteMiddlewares` passes), but it must be designed carefully so precedence stays correct and UX doesn’t become surprising.
