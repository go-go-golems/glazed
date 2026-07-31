---
Title: 'Intern Guide: Extending glazed-migrate to All Removed Glazed Settings APIs'
Ticket: GLZ-MIGRATE-ALL-APIS
Status: proposed
Topics:
    - glazed
    - cli
    - migration
    - settings
    - api-design
    - go-analysis
    - intern-guide
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: repo://cmd/tools/glazed-migrate/main.go
      Note: Tool entrypoint; singlechecker.Main(glazedmigration.Analyzer), the single analyzer to extend
    - Path: repo://pkg/analysis/glazedmigration/analyzer.go
      Note: The analyzer under extension; hardcoded to NewGlazedSchema only, with shadow-safe fix logic
    - Path: repo://pkg/cli/cobra.go
      Note: BuildCobraCommand auto-injects the structured-output section for GlazeCommand (line 225); basis for R1 deletion
    - Path: repo://pkg/cmds/fields/definitions.go
      Note: InitializeDefaultsFromMap (line 393) errors on unknown keys; why output->format rename is required
    - Path: repo://pkg/settings/structured_output.go
      Note: 'Replacement API: NewStructuredOutputSection, SetupStructuredOutput, StructuredOutputSlug, StructuredOutputFlag=format'
ExternalSources: []
Summary: Explains why glazed-migrate covers only one of many removed Glazed settings APIs, maps the full removal surface, and gives an intern-ready plan to extend the analyzer with deterministic and report-only rules.
LastUpdated: 2026-07-28T00:00:00-04:00
WhatFor: Understand the glazed-migrate tool and the full set of migrations it must perform after the GLZ-OUTPUT-FLAGS-CLEANUP removals.
WhenToUse: Start here when extending the migration analyzer or migrating a consuming repository.
---






# Intern Guide: Extending glazed-migrate to All Removed Glazed Settings APIs

## 1. Executive summary

The `glazed-migrate` tool exists to rewrite source code that calls removed Glazed
public APIs. As of the version under review, the tool implements exactly **one**
rewrite rule: `settings.NewGlazedSchema()` → `settings.NewStructuredOutputSection()`.
Every other API removed by the `GLZ-OUTPUT-FLAGS-CLEANUP` cleanup is invisible to
the tool, so running it against a consuming repository (for example `remarquee`)
produces only Go type-checker errors and applies no fixes.

This document maps the complete removal surface and specifies the rules the tool
must learn. The work falls into three difficulty tiers:

1. **Redundant-section deletions** (224 call sites of `NewGlazedSchema`/`NewGlazedSection`
   with no options, or with only an output-format default). Because `cli.BuildCobraCommand`
   now auto-injects the structured-output section for every `cmds.GlazeCommand`,
   the explicit constructor call is usually redundant and can be deleted along
   with its `cmds.WithSections` argument.
2. **Semantic rewrites** (the `output` → `format` flag rename and the
   `GlazedSlug` → `StructuredOutputSlug` constant rename, 37 sites). These are
   deterministic string replacements but touch call sites the analyzer cannot
   always prove safe, so they are best handled as report-only diagnostics with
   machine-checkable suggested fixes.
3. **Runtime-helper rewrites** (`SetupTableProcessor`, `SetupProcessorOutput`,
   `SetupTableOutputFormatter`, `NewOutputFormatterSettings`, 11 sites). The old
   two-call pattern collapses into a single `SetupStructuredOutput` call, and the
   return-value shape changes, so these require structural edits the analyzer
   should report but a human should apply.

The remainder of this guide explains the system from first principles, then gives
file-anchored evidence, pseudocode for each new analyzer rule, decision records,
a phased implementation plan, and a test strategy.

## 2. Problem statement and scope

### 2.1 What the tool is supposed to do

`glazed-migrate` is a Go `analysis.Analyzer` packaged behind the
`singlechecker` driver. Consuming repositories run it with:

```bash
go run github.com/go-go-golems/glazed/cmd/tools/glazed-migrate@latest -fix ./...
```

With `-fix`, the analyzer edits source in place; without `-fix` it prints
diagnostics only. The analyzer sets `RunDespiteErrors: true` so that it still
runs on packages that fail to compile — which is exactly the situation a
migration tool faces, because the removed APIs make the consuming code fail to
type-check.

### 2.2 The gap

The cleanup commit `cdb5537` ("Cleanup flags and middlewares") removed an entire
layer of the Glazed settings package: the aggregate `GlazedSection`, all of its
sub-section constructors, all of the per-feature settings types, and the runtime
`Setup*` helpers. The migration analyzer, however, was written against only the
single `NewGlazedSchema` constructor. As a result:

- Running `glazed-migrate` on `remarquee` (which uses `NewGlazedSection` and
  `WithOutputSectionOptions`, never `NewGlazedSchema`) reports compile errors but
  applies zero fixes.
- Across all `go-go-golems` repositories there are **314 non-test call sites**
  of removed APIs that the current tool cannot touch.

### 2.3 Scope of this work

In scope:

- Enumerate every removed public symbol that still has call sites in consuming
  repositories.
- Specify a deterministic or report-only migration rule for each.
- Extend the analyzer (or add sibling analyzers) to cover them.
- Cover the cross-cutting constant and flag renames (`GlazedSlug`,
  `StructuredOutputSlug`, the `output` → `format` field name).
- Provide a test matrix and a validation procedure.

Out of scope:

- Re-introducing the removed feature sections (select, rename, replace, template,
  jq, sort, skip-limit, fields-filters). Those features are intentionally gone;
  consuming code that relied on their middlewares must be redesigned, not
  migrated mechanically. The analyzer should report such sites clearly.
- Migrating documentation, tutorials, or YAML flag files. Those are tracked
  separately under `GLZ-OUTPUT-FLAGS-CLEANUP`.

## 3. Background: how the migration analyzer works

An intern should be able to read this section and understand the moving parts
without prior context.

### 3.1 The Go analysis framework

`golang.org/x/tools/go/analysis` is the same machinery that backs `go vet` and
`gopls` refactorings. The core abstractions are:

- **`analysis.Analyzer`**: a struct with a `Run` function that receives a
  `*analysis.Pass`. The pass exposes parsed files (`pass.Files`), type
  information (`pass.TypesInfo`), and a `Report` callback for emitting
  diagnostics.
- **`analysis.Diagnostic`**: a positioned message with optional
  `SuggestedFixes`. Each `SuggestedFix` carries `TextEdits` that describe exact
  byte replacements. When the driver runs with `-fix`, it applies those edits.
- **`singlechecker.Main`**: a driver that runs one analyzer as a command-line
  tool, wiring up `-fix`, package loading, and output formatting.

`RunDespiteErrors: true` is essential for a migration tool: the consuming code
does not compile because the target symbols are deleted, so the type checker
produces "undefined" errors. Without that flag the analyzer would bail out
before running.

### 3.2 The current analyzer

File: `glazed/pkg/analysis/glazedmigration/analyzer.go`

The analyzer has one rule, encoded as two constants:

```go
const (
    settingsImportPath = "github.com/go-go-golems/glazed/pkg/settings"
    oldConstructor     = "NewGlazedSchema"
    newConstructor     = "NewStructuredOutputSection"
)
```

Its algorithm, in prose:

1. For each file in the pass, collect the names under which
   `github.com/go-go-golems/glazed/pkg/settings` is imported (qualified alias,
   default `settings`, or dot-import).
2. Skip the file if the settings package is not imported at all.
3. Walk the AST. For every `*ast.CallExpr`, ask `legacyConstructorName` whether
   the called function is `settings.NewGlazedSchema` (or the dot-imported
   `NewGlazedSchema`).
4. If it is, emit a diagnostic.
5. Decide whether a fix is safe:
   - **No arguments**: the call is `NewGlazedSchema()`. The fix is a pure
     identifier rename. It is applied unless a dot-imported replacement would be
     shadowed by a local `NewStructuredOutputSection` declaration.
   - **One or more arguments**: legacy `GlazeSectionOption` arguments do not
     map cleanly onto `schema.SectionOption`. The analyzer reports only, with
     the message "legacy GlazeSectionOption arguments require manual migration".

The `legacyConstructorName` helper is deliberately defensive: it uses
`pass.TypesInfo.Uses` to confirm that the qualifier resolves to the real Glazed
settings package, so an unrelated `settings` alias from another module is never
migrated. When type information is unavailable (tests, heavily broken
packages), it falls back to AST-only matching.

### 3.3 Why the analyzer is structured this way

The design optimizes for **safety over coverage**. A migration tool that
silently rewrites the wrong call is worse than one that reports nothing,
because the failure is invisible. Every fix is gated on either a type-checker
proof (the qualifier resolves to the Glazed package) or a conservative AST
heuristic (no local declaration shadows the replacement).

The shadow-checking path (`dotImportReplacementIsSafe`) is the subtlest part.
For a dot import, `NewStructuredOutputSection` could resolve to either the
package symbol or a local declaration of the same name. The analyzer prefers
the type checker's scope (`LookupParent`); if scopes are unavailable it scans
the file for any top-level or local declaration named `NewStructuredOutputSection`
and withholds the fix. This is the correct conservative behavior.

## 4. Current-state analysis: the full removal surface

### 4.1 What commit `cdb5537` removed

The commit deleted ten files from `pkg/settings/` and added `structured_output.go`
as the replacement. The removed files and their public surface:

| Removed file | Removed public symbols |
|---|---|
| `glazed_section.go` | `GlazedSection`, `GlazeSectionOption`, `NewGlazedSchema`, `NewGlazedSection`, eight `With*SectionOptions` wrappers, `SetupRowOutputFormatter`, `SetupTableOutputFormatter`, `SetupSimpleTableProcessor`, `SetupTableProcessor`, `SetupProcessorOutput`, `GlazedSlug` |
| `settings_fields-filters.go` | `FieldsFilterFlagsDefaults`, `FieldsFiltersSection`, `FieldsFilterSettings`, `NewFieldsFiltersSection`, `NewFieldsFilterSettings` |
| `settings_select.go` | `SelectSettings`, `SelectSection`, `NewSelectSection`, `NewSelectSettingsFromValues` |
| `settings_rename.go` | `RenameSettings`, `RenameFlagsDefaults`, `RenameSection`, `NewRenameSection`, `NewRenameSettingsFromValues` |
| `settings_replace.go` | `ReplaceSettings`, `ReplaceSection`, `NewReplaceSection`, `NewReplaceSettingsFromValues` |
| `settings_template.go` | `TemplateSettings`, `TemplateFlagsDefaults`, `TemplateSection`, `GlazedTemplateSectionSlug`, `NewTemplateSection`, `NewTemplateSettings` |
| `settings_jq.go` | `JqSettings`, `JqSection`, `NewJqSection`, `NewJqSettingsFromValues`, `NewJqMiddlewaresFromSettings` |
| `settings_sort.go` | `SortFlagsSettings`, `SortSection`, `NewSortSection`, `NewSortSettingsFromValues` |
| `settings_skip_limit.go` | `SkipLimitSettings`, `SkipLimitSection`, `NewSkipLimitSection`, `NewSkipLimitSettingsFromValues` |
| `settings_output.go` | `NewOutputFormatterSettings`, the legacy output-flag definitions |

Evidence command:

```bash
git show cdb5537~1:pkg/settings/glazed_section.go | grep -E '^func |^type |^const '
```

### 4.2 What survives in `pkg/settings`

After the cleanup, `pkg/settings/` contains only:

- `structured_output.go` — the replacement API.
- `logcopter.go` — generated logging.

The replacement public API (file: `glazed/pkg/settings/structured_output.go`):

```go
const StructuredOutputSlug = "structured-output"
const StructuredOutputFlag = "format"

type OutputFormat string  // table | json | jsonl | csv | tsv | yaml

type StructuredOutputSettings struct {
    Format        OutputFormat `glazed:"format"`
    OutputFields  []string     `glazed:"output-fields"`
    MaxOutputRows int          `glazed:"max-output-rows"`
}

func NewStructuredOutputSection(options ...schema.SectionOption) (*schema.SectionImpl, error)
func DecodeStructuredOutputSettings(sectionValues *values.SectionValues) (*StructuredOutputSettings, error)
func SetupStructuredProcessor(sectionValues *values.SectionValues, options ...middlewares.TableProcessorOption) (*middlewares.TableProcessor, *StructuredOutputSettings, error)
func SetupStructuredOutput(sectionValues *values.SectionValues, writer io.Writer, options ...middlewares.TableProcessorOption) (*middlewares.TableProcessor, formatters.OutputFormatter, error)
```

### 4.3 The flag rename: `output` → `format`

This is the most consequential semantic change and the one a mechanical tool is
most likely to get wrong. The old aggregate section exposed a flag named
`output` (see removed `pkg/settings/flags/output.yaml`):

```yaml
flags:
  - name: output
    type: choice
    default: table
    choices: [table, csv, tsv, json, yaml, sql, template, markdown, excel]
```

The new section exposes a flag named `format` with a narrower choice set
(`table|json|jsonl|csv|tsv|yaml`). The struct tag confirms the field name:

```go
Format OutputFormat `glazed:"format"`
```

Default-value maps are matched by field name. The relevant function,
`(*Definitions).InitializeDefaultsFromMap` in
`glazed/pkg/cmds/fields/definitions.go:393`, looks up each map key against the
field definitions and **returns an error** for an unknown key:

```go
field, ok := pds.Get(k)
if !ok {
    return errors.Errorf("unknown field when initializing defaults from map: %s", k)
}
```

Consequence: a default map written as `{"output": "json"}` against the new
section fails at runtime with "unknown field when initializing defaults from
map: output". Every such map must be rewritten to `{"format": "json"}`.

The value set also shrank. The old `output` accepted `sql`, `template`,
`markdown`, and `excel`; the new `format` does not. A default value of
`"sql"` or `"excel"` has no replacement and must be redesigned.

### 4.4 The constant rename: `GlazedSlug` → `StructuredOutputSlug`

Code that looked up parsed values by slug used `settings.GlazedSlug` (value
`"glazed"`). The new slug is `settings.StructuredOutputSlug` (value
`"structured-output"`). This affects two idioms:

```go
// old
glazedValues, _ := parsedValues.Get(settings.GlazedSlug)
desc.Schema.Set(settings.GlazedSlug, glazedSection)
outputParameter, _ := parsedLayers.GetParameter(settings.GlazedSlug, "output")

// new
structuredValues, _ := parsedValues.Get(settings.StructuredOutputSlug)
desc.Schema.Set(settings.StructuredOutputSlug, structuredOutputSection)
formatParameter, _ := parsedLayers.GetParameter(settings.StructuredOutputSlug, "format")
```

### 4.5 Auto-injection: most explicit sections are now redundant

The single most important fact for migration planning is in
`glazed/pkg/cli/cobra.go:225`. `BuildCobraCommand` now auto-injects a
structured-output section for every `cmds.GlazeCommand`:

```go
if _, isGlazeCmd := s.(cmds.GlazeCommand); isGlazeCmd {
    structuredSchema := description.Schema.Clone()
    if _, ok := structuredSchema.Get(settings.StructuredOutputSlug); !ok {
        structuredOutputSection, err := settings.NewStructuredOutputSection()
        // ... sets it on the schema
    }
}
```

This means the dominant pattern in consuming code — explicitly constructing a
section and passing it to `cmds.WithSections(...)` — is now redundant whenever
the command implements `GlazeCommand`. Verified example: `dmeta`'s
`ListComponentsCommand` declares `var _ cmds.GlazeCommand =
(*ListComponentsCommand)(nil)` and implements `RunIntoGlazeProcessor`, then
calls `settings.NewGlazedSchema()` and passes the result to
`cmds.WithSections(glazedSection, ...)`. The entire `glazedSection` block is
deletable; `BuildCobraCommand` injects an equivalent default section.

This transforms the migration problem: for the common case, the correct rule is
**delete, not rename**.

## 5. Evidence: call-site census across consuming repositories

The following counts were produced by scanning every repository under
`~/code/wesen/go-go-golems`, excluding `*/ttmp/`, `glazed/pkg/settings/`,
`glazed/pkg/analysis/`, `glazed/cmd/tools/glazed-migrate`, and `*_test.go`.

| Removed symbol | Non-test call sites |
|---|---|
| `settings.NewGlazedSchema(` | 135 |
| `settings.NewGlazedSection(` | 89 |
| `settings.GlazedSlug` (constant) | 35 |
| `settings.WithOutputSectionOptions` | 25 |
| `settings.SetupTableProcessor` | 5 |
| `settings.SetupProcessorOutput` | 3 |
| `settings.NewOutputFormatterSettings` | 2 |
| `settings.GlazedSection` (type, incl. the 89 calls) | 90 |
| `settings.SetupTableOutputFormatter` | 1 |
| `settings.WithFieldsFiltersSectionOptions` | 1 |
| `settings.WithSelectSectionOptions` | 0 |
| `settings.WithTemplateSectionOptions` | 0 |
| `settings.WithRenameSectionOptions` | 0 |
| `settings.WithReplaceSectionOptions` | 0 |
| `settings.WithJqSectionOptions` | 0 |
| `settings.WithSortSectionOptions` | 0 |
| `settings.WithSkipLimitSectionOptions` | 0 |
| Per-feature `New*Section`/`New*Settings` constructors | 0 each |
| `GetParameter(GlazedSlug, "output")` / `GetField(GlazedSlug, "output")` | 2 |

Representative usage sites (for orientation, not exhaustive):

- **Simple redundant section** — `dmeta/pkg/dmeta/cmds/component_inspect.go:61`:
  `glazedSection, err := settings.NewGlazedSchema()` then
  `cmds.WithSections(glazedSection, ...)`.
- **Section with output default** — `remarquee/cmd/remarquee/cmds/cloud/ls.go:47`:
  `settings.NewGlazedSection(settings.WithOutputSectionOptions(schema.WithDefaults(map{"output":"json"})))`.
- **Slug lookup** — `devctl/cmd/devctl/cmds/lifecycle.go:228`:
  `glazedValues, exists := parsedValues.Get(glazedsettings.GlazedSlug)`.
- **Runtime helpers** — `devctl/cmd/devctl/cmds/lifecycle.go:243,249`:
  `SetupTableProcessor(glazedValues)` then
  `SetupProcessorOutput(tableProcessor, glazedValues, out)`.
- **Field-name lookup** — `sqleton/cmd/sqleton/cmds/mcp/mcp.go:104`:
  `parsedValues.GetField(settings.GlazedSlug, "output")`.

The census command, reproducible by an intern:

```bash
cd ~/code/wesen/go-go-golems
for pat in 'NewGlazedSchema\(' 'NewGlazedSection\(' 'WithOutputSectionOptions' \
           'GlazedSlug\b' 'SetupTableProcessor\b' 'SetupProcessorOutput\b' \
           'SetupTableOutputFormatter\b' 'NewOutputFormatterSettings\b'; do
  n=$(grep -rEn --include='*.go' "$pat" . \
      | grep -v '/ttmp/' | grep -v 'glazed/pkg/settings/' \
      | grep -v 'glazed/pkg/analysis/' | grep -v '_test.go' | wc -l)
  printf "%-32s %s\n" "$pat" "$n"
done
```

## 6. Gap analysis

The current analyzer satisfies exactly one of the required migrations
(`NewGlazedSchema` with no arguments). It does not satisfy:

1. **`NewGlazedSection(...)`** — same target as `NewGlazedSchema`, but the
   analyzer's `oldConstructor` constant is hardcoded to `"NewGlazedSchema"`, so
   the rule never fires for the 89 `NewGlazedSection` sites.
2. **`With*SectionOptions(...)` wrappers** — these accepted
   `schema.SectionOption` variadics and re-wrapped them as
   `GlazeSectionOption`. The new `NewStructuredOutputSection` accepts
   `schema.SectionOption` directly, so the wrapper is a no-op adapter that
   should be unwrapped (its arguments passed through). The analyzer has no rule
   for this.
3. **The `output` → `format` default-key rename** — 25 `WithOutputSectionOptions`
   sites and 2 direct field lookups pass `"output"` as a default key. After
   unwrapping, the key must change to `"format"` or the section construction
   fails at runtime. No rule exists.
4. **`GlazedSlug` → `StructuredOutputSlug`** — 35 sites. Pure identifier rename
   on a package-qualified selector. No rule exists.
5. **The `Setup*` runtime helpers** — 11 sites. The old two-call pattern
   (`SetupTableProcessor` + `SetupProcessorOutput`) collapses into one
   `SetupStructuredOutput` call with a different return tuple. Structural edit,
   no rule exists.
6. **Removed feature sections** (select, rename, replace, template, jq, sort,
   skip-limit, fields-filters) — the constructors have zero external call
   sites, but `WithFieldsFiltersSectionOptions` has one site (in `clay`). That
   site relies on a removed feature (`FieldsFilterFlagsDefaults`) and cannot be
   migrated mechanically; it must be flagged for manual redesign.

## 7. Proposed migration rules

This section defines one rule per removed symbol. Each rule states its
determinism tier, the precondition for an auto-fix, the edit, and the
fall-back diagnostic.

### Rule R1: redundant `NewGlazedSection()` / `NewGlazedSchema()` deletion

- **Tier:** deterministic, auto-fixable when proven safe.
- **Precondition:** the call has zero arguments, the result is assigned to a
  local variable, that variable is passed to `cmds.WithSections` as one of
  several arguments, and the enclosing command implements `cmds.GlazeCommand`.
- **Edit:** delete the assignment statement and remove the variable from the
  `WithSections` argument list.
- **Fall-back (when the GlazeCommand proof is unavailable):** rename the call to
  `NewStructuredOutputSection()` (the existing rule), which compiles and
  preserves behavior.
- **Rationale:** `BuildCobraCommand` auto-injects the default section, so the
  explicit call is dead weight. Renaming is always correct but leaves redundant
  code; deletion is cleaner.

### Rule R2: `NewGlazedSection(...)` → `NewStructuredOutputSection(...)`

- **Tier:** deterministic rename (the identifier half); structural for the
  argument list.
- **Precondition:** call resolves to `settings.NewGlazedSection`.
- **Edit:** rename the selector. If arguments are present, proceed to R3/R4.
- **Note:** this is the missing twin of the existing `NewGlazedSchema` rule.
  The two constructors had identical semantics (`NewGlazedSchema` was a thin
  wrapper around `NewGlazedSection`).

### Rule R3: unwrap `With*SectionOptions(...)` wrappers

- **Tier:** deterministic, auto-fixable.
- **Precondition:** a call argument is `settings.WithOutputSectionOptions(args...)`
  (or any `With*SectionOptions` variant).
- **Edit:** replace the wrapper call with its arguments, spliced into the
  enclosing `NewStructuredOutputSection(args...)` call. The wrapper was a pure
  adapter from `[]schema.SectionOption` to `GlazeSectionOption`; the new
  constructor takes `schema.SectionOption` directly.
- **Pseudocode:**

  ```
  match CallExpr where Fun = SelectorExpr(settings, With*SectionOptions)
    parent = enclosing CallExpr  // NewGlazedSection/NewGlazedSchema
    replacement = parent.Args with this CallExpr replaced by its own Args
    emit TextEdit spanning parent.Args => replacement
  ```

### Rule R4: rename default-map key `"output"` → `"format"`

- **Tier:** deterministic within `WithOutputSectionOptions` sites; report-only
  elsewhere.
- **Precondition:** a `schema.WithDefaults(map[string]interface{}{...})` call
  whose map literal contains the key `"output"`.
- **Edit:** rewrite the key to `"format"`. If the value is one of
  `sql|template|markdown|excel` (formats removed in the new section), emit a
  report-only diagnostic instead: "value %q is not a supported structured-output
  format; choose table|json|jsonl|csv|tsv|yaml".
- **Rationale:** see §4.3. Unknown keys crash `InitializeDefaultsFromMap`.

### Rule R5: `GlazedSlug` → `StructuredOutputSlug`

- **Tier:** deterministic identifier rename.
- **Precondition:** a selector `settings.GlazedSlug` (or alias) resolved to the
  Glazed package.
- **Edit:** rename `GlazedSlug` → `StructuredOutputSlug`.
- **Fall-back:** if the selector cannot be type-proven, report-only.

### Rule R6: collapse `SetupTableProcessor` + `SetupProcessorOutput`

- **Tier:** structural, report-only with a suggested fix sketch.
- **Precondition:** a `settings.SetupTableProcessor(values)` call whose result
  is later passed to `settings.SetupProcessorOutput(proc, values, writer)`.
- **Edit:** replace the two calls with one
  `SetupStructuredOutput(values, writer)` call. Note the return-shape change:
  the old `SetupTableProcessor` returned `(*TableProcessor, error)`; the new
  `SetupStructuredOutput` returns `(*TableProcessor, formatters.OutputFormatter, error)`.
  Code that used the processor after the setup call keeps working; code that
  captured the formatter from `SetupProcessorOutput` now captures the second
  return value of `SetupStructuredOutput`.
- **Pseudocode:**

  ```
  // before
  proc, err := settings.SetupTableProcessor(vals)
  if err != nil { return err }
  _, err = settings.SetupProcessorOutput(proc, vals, w)
  if err != nil { return err }

  // after
  proc, _, err := settings.SetupStructuredOutput(vals, w)
  if err != nil { return err }
  ```

### Rule R7: `SetupTableOutputFormatter` → `SetupStructuredOutput`

- **Tier:** structural, report-only.
- **Precondition:** `settings.SetupTableOutputFormatter(values)` (1 site, in
  `parka`).
- **Edit:** the old function returned only a table formatter and did not attach
  it to a processor. The replacement `SetupStructuredOutput` both builds the
  processor and attaches the formatter. The call site must be restructured to
  pass a writer and use the returned processor.

### Rule R8: `NewOutputFormatterSettings` → `DecodeStructuredOutputSettings`

- **Tier:** deterministic rename with return-type change, report-only.
- **Precondition:** `settings.NewOutputFormatterSettings(values)` (2 sites, in
  `devctl`).
- **Edit:** rename to `DecodeStructuredOutputSettings`. The old type
  `OutputFormatterSettings` is gone; the new type is
  `StructuredOutputSettings`. Field access `.Output` becomes `.Format`.

### Rule R9: removed feature sections → manual redesign

- **Tier:** report-only, no fix.
- **Precondition:** any reference to `NewFieldsFiltersSection`,
  `NewSelectSection`, `NewRenameSection`, `NewReplaceSection`,
  `NewTemplateSection`, `NewJqSection`, `NewSortSection`, `NewSkipLimitSection`,
  their `*Settings` types, or `With*SectionOptions` for a removed feature.
- **Diagnostic:** "removed feature section %s has no mechanical migration;
  redesign using application fields or caller-side tools (jq)".
- **Rationale:** these sections implemented runtime middlewares (column
  projection, renaming, jq transforms, sorting) that were intentionally
  removed. Mechanical migration would resurrect deleted behavior.

## 8. Decision records

### Decision: deletion vs. rename for redundant sections

- **Context:** 224 sites construct a default section explicitly. `BuildCobraCommand`
  now auto-injects an equivalent section for `GlazeCommand`.
- **Options considered:**
  1. Always rename to `NewStructuredOutputSection()` (safe, leaves dead code).
  2. Delete the explicit block when the command is provably a `GlazeCommand`.
  3. Report-only and let humans delete.
- **Decision:** implement both. The analyzer attempts deletion (R1) when it can
  prove the `GlazeCommand` relationship via type information; otherwise it falls
  back to rename (R2). This maximizes safe automation while never leaving
  non-compiling code.
- **Rationale:** rename is the floor of correctness (the code compiles and
  behaves identically). Deletion is the ceiling of cleanliness. Proving
  `GlazeCommand` membership is feasible because the type checker resolves the
  receiver type and the analyzer can check its method set.
- **Consequences:** consuming repositories get cleaner code where provable and
  correct-but-redundant code elsewhere. A follow-up cleanup pass can delete the
  survivors.
- **Status:** proposed.

### Decision: auto-fix the `output` → `format` key rename

- **Context:** 25 default maps use `"output"` as a key. Leaving it causes a
  runtime error in `InitializeDefaultsFromMap`.
- **Options considered:**
  1. Auto-rename the key to `"format"` (deterministic given the section).
  2. Report-only.
- **Decision:** auto-fix when the map is an argument to
  `WithOutputSectionOptions` (now unwrapped into `NewStructuredOutputSection`),
  because the target section is unambiguous and the key set is known. Report-only
  for the 2 direct `GetField`/`GetParameter` lookups, because those depend on
  whether the surrounding code still expects the old slug.
- **Rationale:** the key is provably wrong against the new section; not fixing it
  produces a runtime crash. The risk of auto-fixing is low because `"output"`
  has no other meaning in this context.
- **Consequences:** the 25 `WithOutputSectionOptions` sites become zero-touch.
  The value-set check (R4) catches the rare `sql`/`excel` defaults that need
  human decisions.
- **Status:** proposed.

### Decision: `Setup*` helpers are report-only

- **Context:** 11 sites use `SetupTableProcessor`/`SetupProcessorOutput`/etc.
  The replacement changes the return tuple and the call structure.
- **Options considered:**
  1. Auto-apply the collapse in R6.
  2. Report-only with a pseudocode sketch.
- **Decision:** report-only for R6/R7/R8. The return-tuple change and the
  processor-vs-formatter ownership shift are easy to get wrong mechanically;
  a human reviewing a one-line diagnostic with a sketch is faster and safer.
- **Rationale:** these sites are few and concentrated (`devctl`, `parka`); the
  cost of a wrong auto-fix (silent data-shape bug) exceeds the cost of manual
  review.
- **Consequences:** the analyzer covers 303 of 314 sites with fixes; the
  remaining 11 are reported with actionable guidance.
- **Status:** proposed.

### Decision: keep a single analyzer vs. a multichecker

- **Context:** `singlechecker.Main` runs one analyzer. Adding rules could mean
  either expanding the existing analyzer or adding sibling analyzers under a
  `unitchecker`/multichecker.
- **Options considered:**
  1. Expand the existing `glazedmigration.Analyzer` with all rules.
  2. Add sibling analyzers (`glazedslug`, `glazedsetup`) under a multichecker.
- **Decision:** expand the single analyzer. The rules share import detection,
  shadow checks, and the `settingsImportPath` constant; splitting them
  duplicates that machinery and complicates the `-fix` story.
- **Rationale:** the existing analyzer already centralizes the safe-call proof.
  One analyzer also produces one coherent diagnostic stream, which is easier to
  triage.
- **Consequences:** the analyzer file grows; keep it readable by extracting
  one helper function per rule.
- **Status:** proposed.

## 9. Architecture of the extended analyzer

### 9.1 Module layout

```
glazed/
  cmd/tools/glazed-migrate/
    main.go                  # singlechecker.Main(glazedmigration.Analyzer)
  pkg/analysis/glazedmigration/
    analyzer.go              # Analyzer + run() dispatch
    rules_constructor.go     # R1, R2 (constructor rename/delete)
    rules_wrapper.go         # R3, R4 (unwrap + key rename)
    rules_slug.go            # R5 (slug constant rename)
    rules_setup.go           # R6, R7, R8 (runtime helpers, report-only)
    rules_removed.go         # R9 (removed feature sections, report-only)
    shared.go                # settingsImports, resolveSelector, shadow checks
    analyzer_test.go         # table-driven tests per rule
    logcopter.go             # generated
```

The split keeps each rule testable in isolation and keeps `analyzer.go` as a
thin dispatcher.

### 9.2 The dispatch loop

```
func run(pass *analysis.Pass) (any, error) {
    for _, file := range pass.Files {
        imports := settingsImports(file)
        if !imports.usesSettings() { continue }

        applyConstructorRules(pass, file, imports)   // R1, R2
        applyWrapperRules(pass, file, imports)       // R3, R4
        applySlugRules(pass, file, imports)         // R5
        applySetupRules(pass, file, imports)         // R6, R7, R8 (report-only)
        applyRemovedFeatureRules(pass, file, imports) // R9 (report-only)
    }
    return nil, nil
}
```

Each `apply*` function walks the AST once and emits diagnostics with fixes.

### 9.3 Proving the `GlazeCommand` relationship for R1

To delete a redundant section safely, the analyzer must confirm that the command
whose `WithSections` receives the section implements `cmds.GlazeCommand`. The
proof proceeds in three steps:

1. From the `WithSections(...glazedSection...)` call, locate the enclosing
   function that builds the `cmds.CommandDescription` and returns the command
   struct.
2. Resolve the returned type via `pass.TypesInfo`. If it is a named type, look
   up its method set.
3. Confirm the method set contains `RunIntoGlazeProcessor` (the `GlazeCommand`
   interface method). If yes, deletion is safe.

When the proof fails (e.g., the command is returned through an interface or
constructed in a helper), fall back to rename.

```
function isGlazeCommand(pass, withSectionsCall) bool:
    retType = resolveReturnTypeOfEnclosingFunc(pass, withSectionsCall)
    if retType == nil { return false }
    ms = types.NewMethodSet(retType)
    for m in ms:
        if m.Name() == "RunIntoGlazeProcessor" { return true }
    return false
```

### 9.4 The unwrapping edit for R3

The wrapper `WithOutputSectionOptions(schema.WithDefaults(...))` appears as a
single `CallExpr` argument inside a `NewGlazedSection(...)` call. The edit
replaces the wrapper call with its own arguments, spliced into the parent's
argument list. Because the parent's argument list may have other arguments,
the edit must span the exact argument positions.

```
// AST shape:
// NewGlazedSection(
//   WithOutputSectionOptions(
//     schema.WithDefaults(map{...}),
//   ),
// )
//
// Target:
// NewStructuredOutputSection(
//   schema.WithDefaults(map{...}),
// )

edit:
  Pos = start of WithOutputSectionOptions( CallExpr
  End = end of that CallExpr
  NewText = serialize(WithOutputSectionOptions.Args...)  // the inner args, comma-joined
```

Combined with the R4 key rename, a single pass over the map literal rewrites
`"output"` to `"format"`.

## 10. Pseudocode and key flows

### 10.1 End-to-end migration of a typical command

Input (`dmeta/pkg/dmeta/cmds/component_inspect.go`, abridged):

```go
func NewListComponentsCommand() (*ListComponentsCommand, error) {
    glazedSection, err := settings.NewGlazedSchema()
    if err != nil {
        return nil, errors.Wrap(err, "create glazed section")
    }
    commandSettingsSection, err := cli.NewCommandSettingsSection()
    // ...
    desc := cmds.NewCommandDescription(
        "list-components",
        cmds.WithSections(glazedSection, commandSettingsSection),
    )
    return &ListComponentsCommand{CommandDescription: desc}, nil
}
```

Migration:

1. R1 fires: `NewGlazedSchema()` has no arguments; the enclosing command
   implements `GlazeCommand` (verified by method set). Deletion is safe.
2. The analyzer deletes the `glazedSection, err := ...` assignment and its
   error check, and removes `glazedSection` from `WithSections`.

Output:

```go
func NewListComponentsCommand() (*ListComponentsCommand, error) {
    commandSettingsSection, err := cli.NewCommandSettingsSection()
    // ...
    desc := cmds.NewCommandDescription(
        "list-components",
        cmds.WithSections(commandSettingsSection),
    )
    return &ListComponentsCommand{CommandDescription: desc}, nil
}
```

### 10.2 Migration of a command with an output default

Input (`remarquee/cmd/remarquee/cmds/cloud/ls.go`, abridged):

```go
glazedLayer, err := settings.NewGlazedSection(
    settings.WithOutputSectionOptions(
        schema.WithDefaults(map[string]interface{}{"output": "json"}),
    ),
)
```

Migration:

1. R2 renames `NewGlazedSection` → `NewStructuredOutputSection`.
2. R3 unwraps `WithOutputSectionOptions(...)`, splicing its argument
   (`schema.WithDefaults(...)`) into the parent call.
3. R4 rewrites the map key `"output"` → `"format"`. Value `"json"` is in the
   supported set, so the fix is applied.

Output:

```go
glazedLayer, err := settings.NewStructuredOutputSection(
    schema.WithDefaults(map[string]interface{}{"format": "json"}),
)
```

If R1 deletion were also applicable, the whole block could be removed and the
default supplied differently — but because this site sets a non-default format,
the explicit section is load-bearing and should be retained. R1's
precondition (zero arguments, default section) is not met here, so only R2–R4
fire.

### 10.3 Migration of the runtime helpers

Input (`devctl/cmd/devctl/cmds/lifecycle.go`, abridged):

```go
tableProcessor, setupErr := glazedsettings.SetupTableProcessor(glazedValues)
if setupErr != nil { return setupErr }
processor = tableProcessor
if _, err := glazedsettings.SetupProcessorOutput(tableProcessor, glazedValues, cmd.OutOrStdout()); err != nil {
    return err
}
```

Migration (report-only; the diagnostic includes the sketch):

```go
processor, _, err := glazedsettings.SetupStructuredOutput(glazedValues, cmd.OutOrStdout())
if err != nil {
    return err
}
```

The diagnostic message:

```
lifecycle.go:243: replace SetupTableProcessor + SetupProcessorOutput with a single
SetupStructuredOutput(values, writer) call; note the return tuple changes to
(*TableProcessor, OutputFormatter, error)
```

## 11. Implementation plan

### Phase 1: constructor coverage (R1, R2)

- Add `NewGlazedSection` to the rule set alongside `NewGlazedSchema` by
  generalizing `oldConstructor` into a set.
- Implement the `GlazeCommand` proof for R1 deletion.
- Add table-driven tests for: zero-arg rename, zero-arg deletion (proven
  GlazeCommand), zero-arg rename (unprovable), shadowed dot-import.
- Files: `rules_constructor.go`, `shared.go`, `analyzer_test.go`.

### Phase 2: wrapper unwrapping and key rename (R3, R4)

- Implement `applyWrapperRules`: detect `With*SectionOptions` arguments, splice
  their inner arguments into the parent constructor call.
- Implement the map-key rewrite for `"output"` → `"format"` within
  `schema.WithDefaults` literals, with value-set validation.
- Tests: single-wrapper unwrap, nested `WithDefaults`, unsupported value
  (`"excel"`) emits report-only.
- Files: `rules_wrapper.go`.

### Phase 3: slug constant rename (R5)

- Detect `settings.GlazedSlug` selectors; rename to `StructuredOutputSlug`.
- Tests: qualified, aliased, and dot-imported selectors.
- Files: `rules_slug.go`.

### Phase 4: runtime-helper reporting (R6, R7, R8)

- Implement `applySetupRules`: detect `SetupTableProcessor`,
  `SetupProcessorOutput`, `SetupTableOutputFormatter`,
  `NewOutputFormatterSettings`. Emit report-only diagnostics with the
  replacement signature and return-tuple note.
- Tests: each helper produces exactly one diagnostic with the expected message.
- Files: `rules_setup.go`.

### Phase 5: removed-feature reporting (R9)

- Implement `applyRemovedFeatureRules`: detect any `New*Section`/`New*Settings`
  for the removed features and `With*SectionOptions` for removed features.
  Emit report-only diagnostics directing the user to redesign.
- Tests: each removed symbol produces a diagnostic.
- Files: `rules_removed.go`.

### Phase 6: integration and validation

- Run the extended analyzer against `remarquee`, `devctl`, `dmeta`,
  `go-minitrace`, and `parka`. Confirm zero fixes regress and all 11
  report-only sites produce actionable diagnostics.
- Add a `make migrate-check` target to `glazed` that runs the analyzer without
  `-fix` on the consuming repos in the workspace.

## 12. Test strategy

### 12.1 Unit tests (table-driven, AST-only)

Each rule gets a table of `(source, wantDiagnostics, wantFixes)` cases, mirroring
the existing `TestAnalyzerSuggestedFix`. Use the same `parser.ParseFile` harness
so tests do not require a buildable module. Cover:

- Qualified, aliased, and dot imports.
- Shadowed replacements.
- Argument-bearing vs. zero-argument calls.
- `GlazeCommand`-proven vs. unprovable commands.
- Supported vs. unsupported format values.

### 12.2 Integration tests against real repos

- Create a `testdata/` tree with trimmed copies of representative call sites
  (one per rule).
- Run the analyzer with `-fix` and assert the resulting source equals a golden
  file.

### 12.3 Regression guard

- Before extending, capture the current analyzer output on `glazed` itself (the
  internal usages) as a baseline. After extension, confirm no new false
  positives on `glazed/pkg/cli` and `glazed/cmd/glaze`.

## 13. Risks, alternatives, and open questions

### Risks

- **R1 false deletion.** If the `GlazeCommand` proof is unsound, deleting a
  section could remove a load-bearing default. Mitigation: the proof requires a
  resolved named type with `RunIntoGlazeProcessor` in its method set; on any
  doubt, fall back to rename.
- **R4 value-set drift.** If a future change re-adds a format, the value
  allowlist becomes stale. Mitigation: derive the allowlist from
  `structuredOutputFormats` at analyzer init rather than hardcoding.
- **Report-only fatigue.** 11 report-only sites across repos could be ignored.
  Mitigation: the diagnostics include the exact replacement signature and a
  file:line, making them cheap to address.

### Alternatives considered

- **A codemod script instead of an analyzer.** Rejected: an analyzer reuses the
  type checker for safe-call proofs, which a regex codemod cannot do. The
  existing investment in the analyzer framework should be leveraged.
- **Re-adding the removed sections as shims.** Rejected by
  `GLZ-OUTPUT-FLAGS-CLEANUP`; the features are intentionally gone.

### Open questions

- Should R1 also delete the now-unused `import` of `settings` when no other
  reference remains? Probably yes, but `gofmt`/`goimports` already handles
  unused imports; the analyzer can leave that to the formatter.
- Should the analyzer emit a single consolidated diagnostic per file, or one
  per call site? One per call site is more actionable for `-fix`.

## 14. References

### Framework

- `golang.org/x/tools/go/analysis` — analyzer framework.
- `golang.org/x/tools/go/analysis/singlechecker` — CLI driver.

### Glazed source (current)

- `glazed/cmd/tools/glazed-migrate/main.go` — tool entrypoint.
- `glazed/pkg/analysis/glazedmigration/analyzer.go` — the analyzer under extension.
- `glazed/pkg/analysis/glazedmigration/analyzer_test.go` — existing test harness.
- `glazed/pkg/settings/structured_output.go` — replacement API.
- `glazed/pkg/cmds/fields/definitions.go:393` — `InitializeDefaultsFromMap` (key-matching semantics).
- `glazed/pkg/cmds/schema/section-impl.go:98` — `WithDefaults`.
- `glazed/pkg/cli/cobra.go:225` — `BuildCobraCommand` auto-injection.

### Glazed source (removed, via git)

- `git show cdb5537~1:pkg/settings/glazed_section.go` — removed aggregate section.
- `git show cdb5537~1:pkg/settings/flags/output.yaml` — old `output` flag.
- `git show cdb5537` — the cleanup commit diff.

### Consuming-repo evidence

- `remarquee/cmd/remarquee/cmds/cloud/ls.go:47` — `NewGlazedSection` + `WithOutputSectionOptions`.
- `dmeta/pkg/dmeta/cmds/component_inspect.go:61` — redundant `NewGlazedSchema`.
- `devctl/cmd/devctl/cmds/lifecycle.go:197,243,249` — slug lookup + `Setup*` helpers.
- `parka/pkg/glazed/handlers/text/text.go:93` — `SetupTableOutputFormatter`.
- `sqleton/cmd/sqleton/cmds/mcp/mcp.go:104` — `GetField(GlazedSlug, "output")`.
- `clay/pkg/cmds/commandmeta/list.go:43` — `WithFieldsFiltersSectionOptions` (removed feature).

### Related tickets

- `GLZ-OUTPUT-FLAGS-CLEANUP` — the cleanup that created this migration debt.
