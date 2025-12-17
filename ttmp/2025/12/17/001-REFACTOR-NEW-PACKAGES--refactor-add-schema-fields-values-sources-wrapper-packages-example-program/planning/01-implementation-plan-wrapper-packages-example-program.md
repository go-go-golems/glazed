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

- [ ] Add directory `glazed/pkg/cmds/schema/`
- [ ] Add `glazed/pkg/cmds/schema/schema.go` with:
  - [ ] `type Section = layers.ParameterLayer`
  - [ ] `type Sections = layers.ParameterLayers`
  - [ ] `type SectionImpl = layers.ParameterLayerImpl`
  - [ ] `type SectionOption = layers.ParameterLayerOptions`
  - [ ] `type SectionsOption = layers.ParameterLayersOption`
  - [ ] `const DefaultSlug = layers.DefaultSlug`
  - [ ] `func NewSection(slug, name string, opts ...SectionOption) (*SectionImpl, error)` (wrap `layers.NewParameterLayer`)
  - [ ] `func NewSections(opts ...SectionsOption) *Sections` (wrap `layers.NewParameterLayers`)
  - [ ] `func WithSections(sections ...Section) SectionsOption` (wrap `layers.WithLayers`)

### 1.2 `glazed/pkg/cmds/fields`

- [ ] Add directory `glazed/pkg/cmds/fields/`
- [ ] Add `glazed/pkg/cmds/fields/fields.go` with:
  - [ ] `type Definition = parameters.ParameterDefinition`
  - [ ] `type Definitions = parameters.ParameterDefinitions`
  - [ ] `type Type = parameters.ParameterType`
  - [ ] `type Option = parameters.ParameterDefinitionOption`
  - [ ] `func New(name string, t Type, opts ...Option) *Definition` (wrap `parameters.NewParameterDefinition`)
  - [ ] Re-export common options:
    - [ ] `WithHelp`, `WithShortFlag`, `WithDefault`, `WithChoices`, `WithRequired`, `WithIsArgument`
  - [ ] Re-export parameter type constants (at least the ones used by the example program; optionally all):
    - [ ] `TypeString`, `TypeBool`, `TypeInteger`, `TypeChoice`, `TypeStringList`, … (map to `parameters.ParameterType*`)

### 1.3 `glazed/pkg/cmds/values`

- [ ] Add directory `glazed/pkg/cmds/values/`
- [ ] Add `glazed/pkg/cmds/values/values.go` with:
  - [ ] `type SectionValues = layers.ParsedLayer`
  - [ ] `type Values = layers.ParsedLayers`
  - [ ] `type ValuesOption = layers.ParsedLayersOption`
  - [ ] `func New(opts ...ValuesOption) *Values` (wrap `layers.NewParsedLayers`)
  - [ ] `func DecodeInto(v *SectionValues, dst any) error` (wrap `v.InitializeStruct(dst)`)
  - [ ] `func DecodeSectionInto(vs *Values, sectionSlug string, dst any) error` (wrap `vs.InitializeStruct(sectionSlug, dst)`)
  - [ ] (Optional) `func AsMap(vs *Values) map[string]any` (wrap `vs.GetDataMap()`)

### 1.4 `glazed/pkg/cmds/sources`

- [ ] Add directory `glazed/pkg/cmds/sources/`
- [ ] Add `glazed/pkg/cmds/sources/sources.go` with:
  - [ ] `type Middleware = middlewares.Middleware`
  - [ ] Wrapper functions around common sources:
    - [ ] `FromCobra(cmd *cobra.Command, opts ...parameters.ParseStepOption) Middleware`
    - [ ] `FromArgs(args []string, opts ...parameters.ParseStepOption) Middleware`
    - [ ] `FromEnv(prefix string, opts ...parameters.ParseStepOption) Middleware`
    - [ ] `FromDefaults(opts ...parameters.ParseStepOption) Middleware`
    - [ ] (Optional) `FromConfigFilesForCobra(...) Middleware`
  - [ ] `func Execute(sections *schema.Sections, vals *values.Values, ms ...Middleware) error` (wrap `middlewares.ExecuteMiddlewares`)

## 2. Create example program (acceptance test)

### 2.1 Add example program skeleton

- [ ] Add directory `glazed/cmd/examples/refactor-new-packages/`
- [ ] Add `glazed/cmd/examples/refactor-new-packages/main.go`

### 2.2 Program requirements

- [ ] Define a command with **multiple schema sections** using new packages:
  - [ ] `schema.NewSection(schema.DefaultSlug, "Default", ...)` for positional args (optional but recommended)
  - [ ] `schema.NewSection("app", "App", schema.WithPrefix("app-"), ...)`
  - [ ] `schema.NewSection("output", "Output", schema.WithPrefix("output-"), ...)`
- [ ] Define field definitions using `fields.New(...)`
- [ ] Ensure cobra flags are registered via the existing mechanism:
  - Use `cli.BuildCobraCommand(...)` (or `cli.NewCobraParserFromLayers` + `AddToCobraCommand`) so we exercise the actual production codepaths.
- [ ] Parse from **env + cobra**:
  - [ ] Set `cli.CobraParserConfig.AppName = "demo"` (or similar) so env parsing is enabled (prefix becomes `DEMO_`).
  - [ ] Demonstrate env key format (from `middlewares.UpdateFromEnv`):
    - Global prefix: `DEMO_`
    - Per-section prefix: `layerPrefix` from `schema.WithPrefix(...)` (hyphen becomes underscore for env keys)
    - Field name: `p.Name`
    - Example: `DEMO_APP_VERBOSE=true` sets field `verbose` in the `app` section when `WithPrefix("app-")`.
- [ ] Decode resolved values into structs:
  - [ ] `values.DecodeSectionInto(parsed, "app", &AppSettings{})`
  - [ ] `values.DecodeSectionInto(parsed, "output", &OutputSettings{})`
- [ ] Print the resulting struct(s) and optionally the parsed values map for easy manual verification.

### 2.3 Precedence scenario to demonstrate

- [ ] Provide an example run sequence in comments/README:
  - [ ] Defaults set lowest precedence
  - [ ] Config file (optional) overrides defaults
  - [ ] Env overrides config/defaults
  - [ ] Cobra flags override env/config/defaults (highest precedence)

## 3. Tests / validation

### 3.1 Compile-time validation

- [ ] Add lightweight tests that:
  - [ ] import `schema/fields/values/sources`
  - [ ] compile basic usage (e.g., `schema.NewSection`, `fields.New`, `values.DecodeSectionInto`, `sources.Execute`)

### 3.2 Example program build/run validation

- [ ] Confirm it builds:
  - [ ] `go run ./glazed/cmd/examples/refactor-new-packages --help`
- [ ] Confirm env parsing:
  - [ ] `DEMO_APP_VERBOSE=true go run ./glazed/cmd/examples/refactor-new-packages ...`
- [ ] Confirm cobra override:
  - [ ] `DEMO_APP_VERBOSE=true go run ./glazed/cmd/examples/refactor-new-packages --app-verbose=false ...`

## 4. Acceptance criteria

- [ ] New packages exist and are importable:
  - [ ] `glazed/pkg/cmds/schema`
  - [ ] `glazed/pkg/cmds/fields`
  - [ ] `glazed/pkg/cmds/values`
  - [ ] `glazed/pkg/cmds/sources`
- [ ] Example program demonstrates:
  - [ ] multiple schema sections
  - [ ] env + cobra resolution
  - [ ] decoding into structs (per section)
- [ ] `go test ./...` passes

## 5. Follow-ups (optional / out of scope for this ticket unless requested)

- [ ] Update Glazed tutorials to use the new packages (keeping old examples valid)
- [ ] Add “migration notes” doc for downstream repos (old imports still supported)
