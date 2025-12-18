---
Title: 'Implementation plan: wrapper packages + example program'
Ticket: 001-REFACTOR-NEW-PACKAGES
Status: active
Topics:
    - glazed
    - api-design
    - refactor
    - backwards-compatibility
    - migration
    - schema
    - examples
DocType: planning
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Step-by-step implementation plan for adding schema/fields/values/sources wrapper packages and an example program exercising env+cobra parsing and struct decoding."
LastUpdated: 2025-12-17T09:01:05.565205899-05:00
---

## 0. Ticket scaffolding (already done)

- [x] Create ticket `001-REFACTOR-NEW-PACKAGES`
- [x] Create design doc
- [x] Create implementation plan doc
- [x] Create diary doc

## 1. Add new wrapper packages (type aliases + helper functions)

### 1.1 `glazed/pkg/cmds/schema`

- [x] Add directory `glazed/pkg/cmds/schema/`
- [x] Add `glazed/pkg/cmds/schema/schema.go` with:
  - [x] `type Section = layers.ParameterLayer`
  - [x] `type Schema = layers.ParameterLayers`
  - [x] `type SectionImpl = layers.ParameterLayerImpl`
  - [x] `type SectionOption = layers.ParameterLayerOptions`
  - [x] `type SchemaOption = layers.ParameterLayersOption`
  - [x] `const DefaultSlug = layers.DefaultSlug`
  - [x] `func NewSection(slug, name string, opts ...SectionOption) (*SectionImpl, error)` (wrap `layers.NewParameterLayer`)
  - [x] `func NewSchema(opts ...SchemaOption) *Schema` (wrap `layers.NewParameterLayers`)
  - [x] `func WithSections(sections ...Section) SchemaOption` (wrap `layers.WithLayers`)
  - [x] `func NewGlazedSchema(opts ...settings.GlazeParameterLayerOption) (Section, error)` (wrap `settings.NewGlazedParameterLayers`)

### 1.2 `glazed/pkg/cmds/fields`

- [x] Add directory `glazed/pkg/cmds/fields/`
- [x] Add `glazed/pkg/cmds/fields/fields.go` with:
  - [x] `type Definition = parameters.ParameterDefinition`
  - [x] `type Definitions = parameters.ParameterDefinitions`
  - [x] `type Type = parameters.ParameterType`
  - [x] `type Option = parameters.ParameterDefinitionOption`
  - [x] `func New(name string, t Type, opts ...Option) *Definition` (wrap `parameters.NewParameterDefinition`)
  - [x] Re-export common options:
    - [x] `WithHelp`, `WithShortFlag`, `WithDefault`, `WithChoices`, `WithRequired`, `WithIsArgument`
  - [x] Re-export parameter type constants (at least the ones used by the example program; optionally all):
    - [x] `TypeString`, `TypeBool`, `TypeInteger`, `TypeChoice`, `TypeStringList`, … (map to `parameters.ParameterType*`)

### 1.3 `glazed/pkg/cmds/values`

- [x] Add directory `glazed/pkg/cmds/values/`
- [x] Add `glazed/pkg/cmds/values/values.go` with:
  - [x] `type SectionValues = layers.ParsedLayer`
  - [x] `type Values = layers.ParsedLayers`
  - [x] `type ValuesOption = layers.ParsedLayersOption`
  - [x] `func New(opts ...ValuesOption) *Values` (wrap `layers.NewParsedLayers`)
  - [x] `func DecodeInto(v *SectionValues, dst any) error` (wrap `v.InitializeStruct(dst)`)
  - [x] `func DecodeSectionInto(vs *Values, sectionSlug string, dst any) error` (wrap `vs.InitializeStruct(sectionSlug, dst)`)
  - [x] (Optional) `func AsMap(vs *Values) map[string]any` (wrap `vs.GetDataMap()`)

### 1.4 `glazed/pkg/cmds/sources`

- [x] Add directory `glazed/pkg/cmds/sources/`
- [x] Add `glazed/pkg/cmds/sources/sources.go` with:
  - [x] `type Middleware = middlewares.Middleware`
  - [x] Wrapper functions around common sources:
    - [x] `FromCobra(cmd *cobra.Command, opts ...parameters.ParseStepOption) Middleware`
    - [x] `FromArgs(args []string, opts ...parameters.ParseStepOption) Middleware`
    - [x] `FromEnv(prefix string, opts ...parameters.ParseStepOption) Middleware`
    - [x] `FromDefaults(opts ...parameters.ParseStepOption) Middleware`
    - [ ] (Optional) `FromConfigFilesForCobra(...) Middleware`
  - [x] `func Execute(schema *schema.Schema, vals *values.Values, ms ...Middleware) error` (wrap `middlewares.ExecuteMiddlewares`)

## 2. Create example program (acceptance test)

### 2.1 Add example program skeleton

- [x] Add directory `glazed/cmd/examples/refactor-new-packages/`
- [x] Add `glazed/cmd/examples/refactor-new-packages/main.go`

### 2.2 Program requirements

- [x] Define a command with **multiple schema sections** using new packages:
  - [x] `schema.NewSection(schema.DefaultSlug, "Default", ...)` for positional args (optional but recommended)
  - [x] `schema.NewSection("app", "App", schema.WithPrefix("app-"), ...)`
  - [x] `schema.NewSection("output", "Output", schema.WithPrefix("output-"), ...)`
- [x] Define field definitions using `fields.New(...)`
- [x] Ensure cobra flags are registered via the existing mechanism:
  - Use `cli.BuildCobraCommand(...)` (or `cli.NewCobraParserFromLayers` + `AddToCobraCommand`) so we exercise the actual production codepaths.
- [x] Parse from **env + cobra**:
  - [x] Set `cli.CobraParserConfig.AppName = "demo"` (or similar) so env parsing is enabled (prefix becomes `DEMO_`).
  - [x] Demonstrate env key format (from `middlewares.UpdateFromEnv`):
    - Global prefix: `DEMO_`
    - Per-section prefix: `layerPrefix` from `schema.WithPrefix(...)` (hyphen becomes underscore for env keys)
    - Field name: `p.Name`
    - Example: `DEMO_APP_VERBOSE=true` sets field `verbose` in the `app` section when `WithPrefix("app-")`.
- [x] Decode resolved values into structs:
  - [x] `values.DecodeSectionInto(parsed, "app", &AppSettings{})`
  - [x] `values.DecodeSectionInto(parsed, "output", &OutputSettings{})`
- [x] Print the resulting struct(s) and optionally the parsed values map for easy manual verification.

### 2.3 Precedence scenario to demonstrate

- [x] Provide an example run sequence in comments/README:
  - [x] Defaults set lowest precedence
  - [x] Config file (optional) overrides defaults
  - [x] Env overrides config/defaults
  - [x] Cobra flags override env/config/defaults (highest precedence)

## 3. Tests / validation

### 3.1 Compile-time validation

- [x] Add lightweight tests that:
  - [x] import `schema/fields/values/sources`
  - [x] compile basic usage (e.g., `schema.NewSection`, `fields.New`, `values.DecodeSectionInto`, `sources.Execute`)

### 3.2 Example program build/run validation

- [x] Confirm it builds:
  - [x] `go run ./glazed/cmd/examples/refactor-new-packages --help`
- [x] Confirm env parsing:
  - [x] `DEMO_APP_VERBOSE=true go run ./glazed/cmd/examples/refactor-new-packages refactor-demo input.txt`
- [x] Confirm cobra override:
  - [x] `DEMO_APP_VERBOSE=true go run ./glazed/cmd/examples/refactor-new-packages refactor-demo --app-verbose=false input.txt`

## 4. Acceptance criteria

- [x] New packages exist and are importable:
  - [x] `glazed/pkg/cmds/schema`
  - [x] `glazed/pkg/cmds/fields`
  - [x] `glazed/pkg/cmds/values`
  - [x] `glazed/pkg/cmds/sources`
- [x] Example program demonstrates:
  - [x] multiple schema sections
  - [x] env + cobra resolution
  - [x] decoding into structs (per section)
- [x] `go test ./...` passes

## 5. Follow-ups (optional / out of scope for this ticket unless requested)

- [ ] Update Glazed tutorials to use the new packages (keeping old examples valid)
- [ ] Add “migration notes” doc for downstream repos (old imports still supported)
