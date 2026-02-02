---
Title: Naming Options and Rename Audit (Schema/Fields/Values/Sources)
Ticket: GL-001-ADD-MIGRATION-DOCS
Status: active
Topics:
    - glazed
    - api-design
    - naming
    - migration
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/cmds/cmds.go
      Note: CommandDescription includes schema field naming
    - Path: glazed/pkg/cmds/fields
      Note: Definition and parsed values types
    - Path: glazed/pkg/cmds/schema
      Note: Schema/Section definitions and helpers
    - Path: glazed/pkg/cmds/sources
      Note: Source middleware chain and helpers
    - Path: glazed/pkg/cmds/values
      Note: Values container + section values
    - Path: pkg/cmds/cmds.go
      Note: CommandDescription.Layers naming
    - Path: pkg/cmds/fields/parsed-parameter.go
      Note: ParsedParameters and ParsedParameter types
    - Path: pkg/cmds/schema/layer.go
      Note: Primary schema/section definitions
    - Path: pkg/cmds/sources/middlewares.go
      Note: Sources chain naming
    - Path: pkg/cmds/values/parsed-layer.go
      Note: Values and SectionValues types
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: Collect renaming options with usage context for external review
WhenToUse: When deciding final public API vocabulary
---


# Naming Options and Rename Audit (Schema/Fields/Values/Sources)

## Goal
Provide a complete, symbol-by-symbol rename audit of the new Glazed vocabulary and propose alternative names where the current naming feels unclear or misleading. This document is intended for external review ("big brother") and includes concrete usage snippets so reviewers can judge semantics in context.

## High-level observations
- We intentionally removed all backward compatibility; names here are final-facing and will be hard to change later.
- The API has a *three-layer model*:
  1) **schema**: describes what can be provided (sections + field definitions),
  2) **sources**: fills values from inputs (flags/env/files),
  3) **values**: holds resolved values and decodes them into structs.
- The biggest naming concerns are around **"ParsedParameters"** and **"Values"**, which may not convey the final resolved/validated nature or the section-aware container.

## Big questions (explicit prompts for review)
1) Is **Schema** the right name for the container of multiple sections? Or should it be **SectionSet** / **CommandSchema** / **Sections**?
2) Is **Section** the right name for a parameter grouping? Or should it be **Group** / **Scope** / **FieldGroup** / **Namespace**?
3) Is **Values** the right name for the container of section values? Or should it be **Resolved**, **ResolvedValues**, **SectionValuesSet**, **CommandValues**?
4) Is **ParsedParameters** the right name for a bag of resolved values? Or should it be **ResolvedFields**, **FieldValues**, **ParsedFields**, **DecodedFields**?
5) Should the decode method remain a *standalone* helper named **DecodeSectionInto** (per user preference), or is the current instance method **InitializeStruct** sufficient?

## Conventions used in this document
- Each symbol includes:
  - **Current meaning**: what the symbol represents in the API.
  - **Usage snippet**: minimal real usage.
  - **Rename options**: candidate names + rationale.
  - **Notes**: tradeoffs or ecosystem impact.

---

# Package `schema`

## `Schema`
**Current meaning:** Ordered collection of sections describing the command surface (previously ParameterLayers).

Usage snippet:
```go
cmdSchema := schema.NewSchema(
    schema.WithSections(defaultSection, loggingSection),
)
```

Rename options:
- `CommandSchema`: explicit scope; reduces ambiguity with data validation schemas.
- `SectionSet`: emphasizes container-of-sections semantics.
- `Sections`: simplest name, but conflicts with `Section` type and lacks container intent.
- `FieldSchema`: emphasizes fields rather than sections; may undersell grouping.

Notes:
- If "Schema" is retained, consider renaming `CommandDescription.Layers` to `Schema` for consistency.

## `Section`
**Current meaning:** Grouped schema definition (name/slug/description/prefix + definitions).

Usage snippet:
```go
section, _ := schema.NewSection("default", "Default",
    schema.WithPrefix("my-"),
    schema.WithFields(
        fields.New("name", fields.TypeString),
    ),
)
```

Rename options:
- `Group` / `FieldGroup`: expresses UI grouping; less "schema-y".
- `Namespace`: suggests prefixing of keys and grouping.
- `Scope`: emphasizes configuration scope; might be too generic.
- `Segment`: neutral grouping term; less standard.

Notes:
- Section is already used widely in docs; renaming has doc cost.

## `SectionImpl`
**Current meaning:** default concrete implementation of Section.

Usage snippet:
```go
section, _ := schema.NewSection("logging", "Logging")
_ = section.AddFields(fields.New("level", fields.TypeString))
```

Rename options:
- `SectionSpec`: emphasizes schema/spec rather than runtime.
- `SectionDef`: short, explicit.
- `SectionDescription`: mirrors CommandDescription naming.

Notes:
- `Impl` is Go-idiomatic but less expressive for API consumers.

## `SchemaOption` / `SectionOption`
**Current meaning:** functional options for constructors.

Usage snippet:
```go
schema.NewSchema(schema.WithSections(section))
```

Rename options:
- `SchemaOption` -> `SchemaOpt` (shorter) or `SchemaConfig` (verbose).
- `SectionOption` -> `SectionOption` is fine; only change if naming shift happens.

Notes:
- Low priority; rename only if you touch many symbols at once.

## `WithSections`
**Current meaning:** add sections into schema.

Usage snippet:
```go
schema.NewSchema(schema.WithSections(sectionA, sectionB))
```

Rename options:
- `WithSectionSet` (if Schema renamed)
- `WithGroups` / `WithFieldGroups` (if Section renamed)

Notes:
- Keep aligned with `Schema`/`Section` naming decision.

## `NewSchema`
**Current meaning:** constructor for Schema.

Usage snippet:
```go
cmdSchema := schema.NewSchema()
```

Rename options:
- `NewCommandSchema` if Schema renamed.
- `NewSections` if Schema -> Sections.

Notes:
- Straightforward; rename only if the type name changes.

## `NewSection`
**Current meaning:** constructor for SectionImpl.

Usage snippet:
```go
section, _ := schema.NewSection("default", "Default")
```

Rename options:
- `NewGroup` (if Section -> Group)
- `NewSectionDef` (if SectionImpl renamed)

Notes:
- Defer until Section naming is resolved.

## `WithFields` / `WithArguments`
**Current meaning:** register fields on a section; WithArguments marks definitions as positional.

Usage snippet:
```go
schema.NewSection("default", "Default",
    schema.WithFields(fields.New("name", fields.TypeString)),
    schema.WithArguments(fields.New("path", fields.TypeString)),
)
```

Rename options:
- `WithFieldDefs` / `WithDefinitions` for explicitness.
- `WithArgs` for brevity.

Notes:
- If `fields.Definition` renamed, follow that term.

## `AddFields`
**Current meaning:** imperative addition of fields to a section.

Usage snippet:
```go
section.AddFields(fields.New("name", fields.TypeString))
```

Rename options:
- `AddDefinitions` / `AddFieldDefs`
- `RegisterFields`

Notes:
- Align with decision on Definition naming.

## `AppendLayers` / `PrependLayers`
**Current meaning:** mutate schema ordering; still uses "layers" in names.

Usage snippet:
```go
schema.AppendLayers(sectionA, sectionB)
```

Rename options:
- `AppendSections` / `PrependSections`
- `AppendGroups` / `PrependGroups`

Notes:
- High-priority cleanup: this is a leftover of old naming.

## `DefaultSlug`
**Current meaning:** default section slug constant.

Usage snippet:
```go
if err := vals.InitializeStruct(schema.DefaultSlug, &settings); err != nil { ... }
```

Rename options:
- `DefaultSectionSlug` (more explicit)
- `DefaultSchemaSlug` (if Schema renamed)

Notes:
- Low priority, but improves readability in code.

## `CobraSection`
**Current meaning:** Section that can add flags to Cobra and parse values.

Usage snippet:
```go
if cobraSection, ok := section.(schema.CobraSection); ok {
    _ = cobraSection.AddLayerToCobraCommand(cmd)
}
```

Rename options:
- `CobraGroup` / `CobraSection` (keep), `CobraFieldGroup`
- Method rename: `AddLayerToCobraCommand` -> `AddSectionToCobraCommand`

Notes:
- Methods still use "Layer" in name; rename for consistency.

## `FlagGroup` / `FlagGroupUsage` / `CommandFlagGroupUsage`
**Current meaning:** help/usage grouping of Cobra flags.

Usage snippet:
```go
usage := schema.ComputeCommandFlagGroupUsage(cmd)
```

Rename options:
- If "Field" becomes primary term: `FieldGroup` / `FieldGroupUsage`.
- Otherwise keep: these are Cobra-centric and may remain "flag".

Notes:
- These names might be fine because they're explicitly about CLI flags, not schema groups.

---

# Package `fields`

## `Definition`
**Current meaning:** declarative field definition (name, type, default, etc.).

Usage snippet:
```go
def := fields.New("name", fields.TypeString, fields.WithRequired(true))
```

Rename options:
- `FieldDef` / `FieldDefinition`
- `Spec` / `FieldSpec`
- `ParamDef` (if "parameter" term retained)

Notes:
- `Definition` is generic and may collide with other packages in imports.

## `Definitions`
**Current meaning:** ordered map of `Definition` keyed by name.

Usage snippet:
```go
defs := fields.NewDefinitions()
defs.Set("name", def)
```

Rename options:
- `FieldDefs` / `FieldDefinitions`
- `FieldSet` (implies set semantics, but order matters)
- `DefinitionMap` (explicit about map)

Notes:
- If `Definition` changes, align naming with it.

## `ParsedParameter`
**Current meaning:** a single resolved field value + definition + parse log.

Usage snippet:
```go
pp := &fields.ParsedParameter{Definition: def}
_ = pp.Update("value", fields.WithSource("flags"))
```

Rename options:
- `ResolvedField` / `ResolvedValue`
- `FieldValue` (clear but hides parse log)
- `ParsedField` (if "parsed" retained)

Notes:
- "Parsed" reads like raw input; this is validated/resolved.

## `ParsedParameters`
**Current meaning:** ordered collection of parsed parameters.

Usage snippet:
```go
parsed := fields.NewParsedParameters()
_ = parsed.SetAsDefault("name", def, "default")
```

Rename options:
- `ResolvedFields`
- `FieldValues`
- `ParsedFields`
- `ResolvedFieldSet`

Notes:
- This is the most debated name; it represents resolved values, not raw parse.

## `Type`
**Current meaning:** field data type (string, bool, file, object, etc.).

Usage snippet:
```go
fields.New("count", fields.TypeInteger)
```

Rename options:
- `FieldType` (disambiguates)
- `ValueType` (emphasizes runtime type)

Notes:
- Current `Type` is short but easily collides with other packages.

## `ParseStep`, `ParseOption`, `WithSource`
**Current meaning:** provenance log for how a value was produced.

Usage snippet:
```go
pp.Update("value", fields.WithSource("env"))
```

Rename options:
- `ResolutionStep` / `ResolutionOption` (if moving away from "parse")
- `WithOrigin` instead of `WithSource`

Notes:
- Only consider if changing Parsed* naming.

## `InitializeStruct` (on `ParsedParameters`)
**Current meaning:** decode resolved values into a struct.

Usage snippet:
```go
if err := parsed.InitializeStruct(&settings); err != nil { ... }
```

Rename options:
- `DecodeInto` / `DecodeStruct`
- `HydrateStruct`

Notes:
- User preference: favor standalone `values.DecodeSectionInto(...)` for section-aware decoding.

---

# Package `values`

## `Values`
**Current meaning:** ordered map from section slug to `SectionValues`.

Usage snippet:
```go
vals := values.New()
err := vals.InitializeStruct(schema.DefaultSlug, &settings)
```

Rename options:
- `ResolvedValues`
- `SectionValuesSet`
- `CommandValues`
- `SectionValuesMap`

Notes:
- If most commands only use a single section, a singular name might fit better (`CommandValues`).
- If multiple sections are first-class, `SectionValuesSet` or `ResolvedValues` is clearer.

## `SectionValues`
**Current meaning:** resolved values for a single section (definition + parsed values).

Usage snippet:
```go
sectionValues, _ := values.NewSectionValues(section,
    values.WithParameterValue("name", "demo"),
)
```

Rename options:
- `ResolvedSection`
- `SectionData`
- `SectionResolvedValues`
- `FieldValues`

Notes:
- If `Values` is renamed, this likely should be renamed as well for symmetry.

## `SectionValuesOption`
**Current meaning:** functional options for `NewSectionValues`.

Usage snippet:
```go
values.NewSectionValues(section, values.WithParameterValue("name", "demo"))
```

Rename options:
- `SectionValueOption`
- `ResolvedSectionOption`

Notes:
- Low priority, but align with `SectionValues` rename.

## `InitializeStruct` (on `Values` and `SectionValues`)
**Current meaning:** decode values into a struct for a given section.

Usage snippet:
```go
_ = vals.InitializeStruct(schema.DefaultSlug, &settings)
```

Rename options:
- **Preferred per user note:** `values.DecodeSectionInto(vals, slug, &dst)` as a top-level helper.
- `DecodeSection` / `DecodeInto` for method names.
- `Hydrate` if you want a more narrative term.

Notes:
- If we keep `InitializeStruct`, rename the method or add a top-level helper for clarity.

## `DefaultSlug`
**Current meaning:** mirror of schema.DefaultSlug to avoid import cycle.

Usage snippet:
```go
_ = vals.InitializeStruct(values.DefaultSlug, &settings)
```

Rename options:
- `DefaultSectionSlug` (consistency with schema).

Notes:
- Low priority; only change if schema constant changes.

---

# Package `sources`

## `HandlerFunc`
**Current meaning:** function that mutates schema + values during resolution.

Usage snippet:
```go
type HandlerFunc func(layers *schema.Schema, parsedLayers *values.Values) error
```

Rename options:
- `ResolverFunc` / `ResolveFunc`
- `SourceFunc`
- `ApplyFunc`

Notes:
- Parameter names still use layers/parsedLayers; rename to `schema`/`vals`.

## `Middleware`
**Current meaning:** function transforming HandlerFunc (chain of sources).

Usage snippet:
```go
mw := sources.Chain(sources.FromDefaults(), sources.FromEnv("APP"))
```

Rename options:
- `Source` / `Resolver` (if you want to hide middleware pattern)
- `PipelineStep` (explicit pipeline semantics)

Notes:
- Current naming is technically accurate, but domain terminology is "sources".

## `Execute`
**Current meaning:** run a chain of middlewares to fill values.

Usage snippet:
```go
err := sources.Execute(schema_, vals,
    sources.FromDefaults(),
    sources.FromEnv("APP"),
)
```

Rename options:
- `Resolve` (reflects data resolution)
- `ApplySources`
- `Run` (short, generic)

Notes:
- `Execute` is fine but could be more domain-specific.

## `Chain`
**Current meaning:** compose middlewares.

Usage snippet:
```go
pipeline := sources.Chain(s1, s2, s3)
```

Rename options:
- `Pipeline` / `Compose`
- `CombineSources`

Notes:
- Keep if middleware pattern is retained.

## `FromDefaults`, `FromEnv`, `FromFile`, `FromCobra`, `FromArgs`
**Current meaning:** source constructors for different inputs.

Usage snippet:
```go
sources.FromEnv("APP", sources.WithSource("env"))
```

Rename options:
- Prefix with `Source` for explicitness: `SourceFromEnv`, etc.
- Keep `FromX` for brevity.

Notes:
- These are already readable; rename only if you rename `sources` package.

## `SourceDefaults` (const)
**Current meaning:** default source label for default values.

Usage snippet:
```go
sources.FromDefaults(sources.WithSource(sources.SourceDefaults))
```

Rename options:
- `DefaultSource` / `DefaultsSource`

Notes:
- Low priority; aligns with `WithSource`/`WithOrigin` naming.

---

# Package `cmds`

## `CommandDescription.Layers`
**Current meaning:** schema for a command (now `*schema.Schema`).

Usage snippet:
```go
desc := &cmds.CommandDescription{
    Name: "demo",
    Layers: schema.NewSchema(schema.WithSections(section)),
}
```

Rename options:
- `Schema` (most consistent with new vocabulary)
- `Sections` (if Schema renamed to SectionSet)

Notes:
- High priority if we want to fully remove "layers" from public API.

## `WithLayers`, `WithLayersList`, `WithLayersMap`
**Current meaning:** option helpers for CommandDescription.

Usage snippet:
```go
cmds.WithLayers(schema.NewSchema(...))
```

Rename options:
- `WithSchema`
- `WithSections` (if Schema renamed)

Notes:
- `WithSchema` already exists as alias; consider removing Layers forms once migration is done.

---

# Additional naming leftovers (lower priority but visible)

These are visible or semi-visible symbols that still carry the old vocabulary:

- `SectionImpl.ChildLayers` (field): rename to `ChildSections` or `Children`.
- `AddLayerToCobraCommand` method on `CobraSection`: rename to `AddSectionToCobraCommand`.
- `ParseLayerFromCobraCommand` on `CobraSection`: rename to `ParseSectionFromCobraCommand`.
- `GetDefaultParameterLayer` on `values.Values`: rename to `GetDefaultSection` or `GetDefaultValues`.
- `AppendLayers` / `PrependLayers` in `schema.Schema`: rename to `AppendSections` / `PrependSections`.
- Variable names across public signatures: `layers`, `parsedLayers` -> `schema`, `vals` or `values`.

Example snippet (current):
```go
func (p *Values) GetDefaultParameterLayer() *SectionValues
```

Potential rename:
```go
func (p *Values) GetDefaultSectionValues() *SectionValues
```

---

# Recommendations (draft)

If we had to pick a consistent end-state:

1) **Schema/Section**: keep `Schema` and `Section`, but rename container field `Layers` to `Schema` in `CommandDescription` to remove cognitive dissonance.
2) **Values**: rename `Values` -> `ResolvedValues` and `SectionValues` -> `ResolvedSection` (or `SectionResolvedValues`).
3) **ParsedParameters**: rename to `ResolvedFields` (or `FieldValues`), and `ParsedParameter` -> `ResolvedField`.
4) **Decode helper**: add or keep a top-level `values.DecodeSectionInto(vals, slug, &dst)` for clarity, even if the method remains.
5) **Leftover layer terminology**: rename `AppendLayers`, `ChildLayers`, `AddLayerToCobraCommand` and related methods.

These would align the vocabulary with actual semantics: **Schema** defines, **Sources** resolve, **ResolvedValues** hold results, **DecodeSectionInto** hydrates structs.

---

# Open questions for external review

1) Does `ResolvedValues` feel too verbose? Is `Values` actually sufficient in practice?
2) Is `Section` the best grouping word for the CLI parameter model, or should it be `Group` or `Namespace`?
3) Should decoding be an explicit verb (`DecodeSectionInto`) instead of method `InitializeStruct`?
4) Are there any existing Go libraries or patterns we should align with (e.g., Cobra/Viper or config schema naming)?

