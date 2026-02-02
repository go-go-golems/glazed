---
Title: 'Brainstorm: renaming layers/parameters API for clarity'
Ticket: GLAZED-LAYER-RENAMING
Status: active
Topics:
    - glazed
    - api-design
    - naming
DocType: brainstorm
Intent: reviewed
Owners:
    - manuel
RelatedFiles:
    - Path: glazed/pkg/cmds/layers/layer.go
      Note: Defines ParameterLayer (schema-ish grouping)
    - Path: glazed/pkg/cmds/layers/parsed-layer.go
      Note: Defines ParsedLayer/ParsedLayers (resolved values) and InitializeStruct
    - Path: glazed/pkg/cmds/parameters/parameters.go
      Note: Defines ParameterDefinition + NewParameterDefinition
    - Path: glazed/pkg/doc/topics/layers-guide.md
      Note: User-facing docs using current names
    - Path: glazed/pkg/doc/tutorials/custom-layer.md
      Note: Tutorial using ParameterLayer/ParameterDefinition/ParsedLayers
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/analysis/01-option-a-implementation-plan-schema-field-values-renaming-transitional-api.md
      Note: Option A implementation deep dive derived from this brainstorm
    - Path: glazed/ttmp/2025/12/16/GLAZED-LAYER-RENAMING--rename-parameterlayer-parsedlayers-parameterdefinition-for-clarity/reference/01-debate-prep-candidates-and-questions-for-renaming-layers-parameters.md
      Note: Debate agenda/candidates built from brainstorm questions
ExternalSources: []
Summary: Explore naming schemes for ParameterLayer/ParameterDefinition/ParsedLayers and related packages/APIs
LastUpdated: 2025-12-16T18:04:17.646312962-05:00
---


# Brainstorm: renaming layers/parameters API for clarity

## Context / problem statement

Glazed’s “parameters” system is powerful but the current naming is easy to misunderstand on first contact:

- `fields.Definition` is really a **field spec** (name/type/default/help; sometimes a flag, sometimes an arg).
- `schema.Section` is really a **grouped schema/spec** (a “section” with a slug/prefix + a list of field specs).
- `values.SectionValues` and `values.Values` are really **resolved values** (values that came from some source(s) and have provenance).

The request: brainstorm alternative naming schemes—not only for the types, but also for **package names**, **functions**, and the **surface API vocabulary**, then evaluate tradeoffs.

This doc is intentionally exploratory (no decision yet).

## Naming goals (what “good” looks like)

- **Matches mental model**: names should align with “schema vs values vs sources”.
- **Minimize overload**: avoid “layer” meaning both “feature layer” and “schema section”.
- **Compositional**: names should read well in fluent Go call sites.
- **Migration-friendly**: feasible to roll out incrementally (aliases + deprecations).
- **Avoid ecosystem collisions**: “Config”, “Settings”, “Schema” are common; pick the least confusing within Glazed context.

## Current glossary (anchor to reality)

These are the *actual* meanings based on code:

- `fields.Definition`: declarative parameter description (flag or arg); includes `Type`, `Default`, `Required`, help text.
- `schema.Section`: groups many `ParameterDefinition`s + has `Slug`, `Name`, `Description`, `Prefix`, cloning, and default-init helpers.
- `values.SectionValues`: holds `Layer ParameterLayer` + `Parameters *parameters.ParsedParameters` (i.e., resolved values).
- `values.Values`: ordered map of layer slug → `*ParsedLayer`; has helpers like `InitializeStruct(layerKey, dst)`.

Notable “peripheral” vocabulary that also impacts readability:

- Package names: `cmds/layers`, `cmds/parameters`, `cmds/middlewares`, `cmds/runner`
- APIs: `NewParameterLayer`, `NewParameterDefinition`, `ExecuteMiddlewares`, `UpdateFromEnv`, `LoadParametersFromFiles`, `ParseFromCobraCommand`, `GatherArguments`
- Hydration: `InitializeStruct`, struct tags `glazed.parameter:"..."`

## Brainstorm naming bundles

Below are *coherent* bundles: if we adopt one, we should rename most of the related symbols for consistency.

### Bundle A: “Schema / Field / Values”

**Intent**: align with common vocabulary from config/validation libraries.

#### Type renames (core)

| Current | Proposed | Rationale |
|---|---|---|
| `fields.Definition` | `fields.FieldDefinition` or `schema.Field` | It describes one field’s spec. |
| `fields.Definitions` | `schema.Schema` or `fields.Definitions` | A set of field specs. |
| `schema.Section` | `schema.Section` or `schema.SchemaSection` | A named/slugged “section schema”. |
| `schema.Schema` | `schema.Sections` | Set of section schemas. |
| `values.SectionValues` | `values.SectionValues` | Values for one section schema. |
| `values.Values` | `values.Values` or `values.Sections` | Values across sections. |
| `parameters.ParsedParameter` | `values.FieldValue` | The resolved value + provenance. |
| `parameters.ParsedParameters` | `values.FieldValues` | Map of values for fields. |

#### Package renames (suggested)

- `pkg/cmds/layers` → `pkg/cmds/schema` (or `pkg/schema` if it’s not command-specific)
- `pkg/cmds/parameters` → `pkg/cmds/fields` (or `pkg/cmds/schema/fields`)
- `pkg/cmds/middlewares` → `pkg/cmds/sources` (see below)

#### API vocabulary (examples)

- `NewParameterLayer(...)` → `schema.NewSection(...)`
- `NewParameterDefinition(...)` → `fields.New(...)` / `fields.String("host")` / `fields.Int("port")`
- `ParsedLayers.InitializeStruct(...)` → `Values.DecodeSectionInto(sectionKey, dst)` or `Values.Bind(sectionKey, dst)`

#### Pros / cons

- **Pros**: widely understood; clearly distinguishes spec vs values; scales to non-CLI sources.
- **Cons**: “schema” can imply strict JSON-schema style; “section” may confuse with docs sections; lots of renames.

---

### Bundle B: “Options / OptionGroup / ResolvedOptions”

**Intent**: speak the language of CLI flags and “options” rather than “parameters”.

#### Type renames (core)

| Current | Proposed | Rationale |
|---|---|---|
| `ParameterDefinition` | `OptionDefinition` or `OptionSpec` | It defines a CLI option (flag/arg). |
| `ParameterLayer` | `OptionGroup` | Groups related options (logging/output/etc.). |
| `ParsedLayer` | `ResolvedOptionGroup` | Values after resolution. |
| `ParsedLayers` | `ResolvedOptions` | All resolved option groups. |

#### Package renames (suggested)

- `cmds/layers` → `cmds/options` or `cmds/optiongroups`
- `cmds/parameters` → `cmds/options`
- `cmds/middlewares` → `cmds/options/sources` (or `cmds/options/resolvers`)

#### API vocabulary (examples)

- `schema.NewSection("logging", ...)` → `options.NewGroup("logging", ...)`
- `fields.New("log-level", ...)` → `options.New("log-level", ...)`
- `parsedLayers.InitializeStruct("logging", &cfg.Logging)` → `resolved.Bind("logging", &cfg.Logging)`

#### Pros / cons

- **Pros**: very intuitive for Cobra/CLI users; aligns with “flags/options”.
- **Cons**: “options” can feel too CLI-specific for config files/env; may under-represent non-CLI sources like HTTP queries.

---

### Bundle C: “Inputs / InputGroup / ResolvedInputs”

**Intent**: explicitly represent that these values come from different “input sources”.

#### Type renames (core)

| Current | Proposed |
|---|---|
| `ParameterDefinition` | `InputDefinition` / `InputSpec` |
| `ParameterLayer` | `InputGroupSpec` |
| `ParsedLayer` | `ResolvedInputGroup` |
| `ParsedLayers` | `ResolvedInputs` |

#### Package renames (suggested)

- `cmds/parameters` → `cmds/inputs`
- `cmds/layers` → `cmds/inputgroups` (or keep `layers` but rename type names)
- `cmds/middlewares` → `cmds/sources` (very natural in this bundle)

#### Pros / cons

- **Pros**: general purpose; fits “env/config/cobra/http” equally.
- **Cons**: “input” is generic; may lose the “CLI-parameter” clarity.

---

### Bundle D: “Config / Section / EffectiveConfig”

**Intent**: emphasize that the end-product is “effective configuration”, not “parsed params”.

#### Type renames (core)

| Current | Proposed |
|---|---|
| `ParameterDefinition` | `ConfigField` / `SettingField` |
| `ParameterLayer` | `ConfigSectionSchema` |
| `ParsedLayers` | `EffectiveConfig` / `ResolvedConfig` |

#### Pros / cons

- **Pros**: reads nicely for app devs (“effective config”).
- **Cons**: Glazed’s “layers” are used for more than config; “config” may be misleading inside the command framework.

## Peripheral naming: “middlewares” might really be “sources”

Many `cmds/middlewares` functions read like “sources/resolvers” rather than generic middleware:

- `UpdateFromEnv` → `FromEnv` / `EnvSource`
- `LoadParametersFromFiles` → `FromConfigFiles` / `ConfigFileSource`
- `ParseFromCobraCommand` → `FromCobraFlags`
- `GatherArguments` → `FromArgs`
- `SetFromDefaults` → `FromDefaults`
- `ExecuteMiddlewares` → `Resolve` / `ApplySources` / `RunResolvers`

Possible naming shift (conceptual):

- “middleware chain” → “source chain” / “resolution chain”
- “parse step source” → “value source” / “origin”

## API shape alternatives (beyond renaming)

If we *also* want to improve the surface API (future work), the naming bundles suggest different fluent APIs:

### API shape 1: explicit schema + sources

```go
schema := schema.NewSections(
  schema.Section("redis",
    schema.Prefix("redis-"),
    schema.Fields(
      fields.String("host").Default("127.0.0.1"),
      fields.Int("port").Default(6379),
    ),
  ),
)

values, err := sources.Resolve(
  schema,
  sources.Defaults(),
  sources.ConfigFiles("app.yaml"),
  sources.Env("MYAPP"),
  sources.Cobra(cmd, args),
)

var cfg RedisSettings
_ = values.DecodeSectionInto("redis", &cfg)
```

### API shape 2: preserve current but rename nouns

Keep the structure but rename the types:

- `schema.Section` → `layers.SchemaSection`
- `fields.Definition` → `parameters.FieldDefinition`
- `values.Values` → `layers.ResolvedValues`

This is less disruptive but still improves readability.

## Migration sketch (feasibility reality-check)

Renaming here is high-impact: it touches docs, public APIs, and downstream code.

An incremental strategy:

1. **Introduce new packages/types as aliases** (Go `type X = Y` where possible).
2. **Add wrappers for constructors** (`NewX` forwards to `NewY`) and keep old ones.
3. **Deprecate old names** using Go doc `// Deprecated:` comments.
4. **Update internal code + docs** to prefer new names.
5. **After 1–2 releases**, consider removing old names (or keep forever if compatibility matters).

Open question: do we want to rename packages (`cmds/layers`) or just types? Package renames are the most breaking because import paths change.

## Initial candidate recommendation (not final)

My current leaning (based on minimizing confusion for both CLI and “config sources” users):

- Prefer **Bundle A** (Schema/Field/Values) *or* **Bundle C** (Inputs/ResolvedInputs).
- Avoid Bundle D “Config” as a global rename inside `cmds/` because Glazed layers are broader than app config.
- Bundle B “Options” is great for Cobra, but may be too CLI-centric for config/env usage.

## Open questions to answer before deciding

- Should “layer” remain as the grouping concept (since it already exists everywhere), and we only rename “parameter layer” to “layer schema”?
- How much does `glazed/pkg/settings` (output formatting etc.) bias us toward “options”?
- Do we want to rename `InitializeStruct` → `DecodeInto` / `BindInto` globally for clarity?
- Do we want to keep `parameters` as the package name but rename the types (less breaking)?


