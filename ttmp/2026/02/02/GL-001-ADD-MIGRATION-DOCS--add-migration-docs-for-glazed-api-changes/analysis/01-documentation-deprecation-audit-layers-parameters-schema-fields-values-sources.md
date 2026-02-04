---
Title: Documentation Deprecation Audit — Schema/Fields/Values/Sources Renames
Ticket: GL-001-ADD-MIGRATION-DOCS
Status: active
Topics:
    - glazed
    - migration
    - docs
    - audit
DocType: analysis
Intent: short-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/doc
      Note: Documentation corpus scanned
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/analysis/02-doc-deprecation-scan.json
      Note: Raw scan output (per-file matches)
    - Path: glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/doc_deprecation_scan.py
      Note: Scan script used to generate this audit
    - Path: pkg/doc/topics/13-layers-and-parsed-layers.md
      Note: High-signal doc with legacy terminology
    - Path: ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/analysis/02-doc-deprecation-scan.json
      Note: Raw scan output (per-file matches)
    - Path: ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/doc_deprecation_scan.py
      Note: Scanner script
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Documentation Deprecation Audit — Schema/Fields/Values/Sources Renames

This report enumerates every scanned documentation file and flags legacy API names that were removed in the no-return refactor. It is generated from a deterministic scan (see `02-doc-deprecation-scan.json`) and annotated with per-file remediation guidance.

## Summary
- Repository root: `/home/manuel/workspaces/2026-02-02/refactor-glazed-names`
- Files scanned: 67
- Patterns scanned: 26
- Files with legacy hits: 19

### Global rename map (no-return)
- `layers` → `schema` (sections + schema collections)
- `parameters` → `fields` (definitions + parsed parameters)
- `middlewares` → `sources` (parse middleware chain)
- `ParsedLayer(s)` → `values.SectionValues` / `values.Values`
- `ParameterDefinition(s)` → `Definition(s)`
- `AddFlags` → `AddFields`
- `ExecuteMiddlewares` → `sources.Execute`
- `LoadParametersFromFile(s)` → `sources.FromFile(s)`
- `UpdateFromEnv` → `sources.FromEnv`
- `WithParseStepSource` → `fields.WithSource`

### Pattern frequency (all files)
- ParameterType: 43
- pkg/cmds/layers import: 24
- pkg/cmds/middlewares import: 24
- ParameterDefinition: 22
- LoadParametersFromFile: 19
- ParsedLayers: 19
- pkg/cmds/parameters import: 19
- AddFlags: 12
- ParameterLayers: 11
- ParameterLayer: 10
- SetFromDefaults: 9
- ParsedLayer: 8
- UpdateFromEnv: 8
- ExecuteMiddlewares: 7
- LoadParametersFromFiles: 7
- ParameterDefinitions: 6
- ParseFromCobraCommand: 4
- layers.ParameterLayer: 4
- WithParseStepSource: 3
- layers.ParameterLayers: 2
- parameters.ParameterDefinition: 2
- CobraParameterLayer: 1
- CommandDefinition: 1
- GatherArguments: 1
- parameters.ParameterDefinitions: 1

## Deprecated / Remove
- `glazed/pkg/doc/tutorials/migrating-to-facade-packages.md` — fully deprecated (facade packages removed). Replace with a no-return migration doc that targets schema/fields/values/sources directly.

## Per-file index (exhaustive)

### glazed/README.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/applications/01-exposing-a-simple-sql-table.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/applications/02-iterating-over-a-column-in-shell.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/applications/03-user-store-command.md
- Status: Update required
- Legacy match count: 6
- Pattern labels: pkg/cmds/layers import, pkg/cmds/parameters import
- Recommended updates:
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 134 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 213 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 265 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 135 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 214 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 266 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
```

### glazed/pkg/doc/examples/cliopatra/cliopatra.md
- Status: Update required
- Legacy match count: 1
- Pattern labels: ParameterDefinition
- Recommended updates:
  - Replace `ParameterDefinition` with `Definition`.
- Matches:

```text
L  31 [ParameterDefinition] map it to the ParameterDefinition the Command uses, and create a YAML file with the default values.
```

### glazed/pkg/doc/examples/filter/remove-duplicates.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/help/help-example-1.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/help/help-example-2.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/jq/01-jq-replace.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/jq/02-jq-filter.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/output/multiple-output-file.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/output/sql-output.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/regex-filters/regex-filters.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/rename/01-rename-column.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/rename/02-rename-regexps.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/rename/03-rename-yaml.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/replace/replace-add-fields.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/select/select-example-1.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/select/select-example-2.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/skip-limit/01-skip-limit.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/sort/01-sort-by.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/templates/templates-example-1.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/templates/templates-example-2.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/templates/templates-example-3.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/examples/yaml/yaml-sanitize.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/00-documentation-guidelines.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/01-help-system.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/02-markdown-style.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/03-templates.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/04-flag-groups.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/05-table-format.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/06-usage-string.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/07-load-parameters-from-json.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/08-file-parameter-type.md
- Status: Update required
- Legacy match count: 1
- Pattern labels: ParameterType
- Recommended updates:
  - Replace `ParameterType*` with `fields.Type*` names.
- Matches:

```text
L  86 [ParameterType] ## The `file` ParameterType
```

### glazed/pkg/doc/topics/09-gather-flags-from-string-list.md
- Status: Update required
- Legacy match count: 5
- Pattern labels: ParameterDefinition, ParameterType
- Recommended updates:
  - Replace `ParameterDefinition` with `Definition`.
  - Replace `ParameterType*` with `fields.Type*` names.
- Matches:

```text
L  25 [ParameterDefinition] params []*ParameterDefinition,
L  35 [ParameterDefinition] - `params`: a slice of `*ParameterDefinition` representing the parameter definitions.
L  49 [ParameterDefinition] params := []*ParameterDefinition{
L  50 [ParameterType] {Name: "verbose", ShortFlag: "v", Type: ParameterTypeBool},
L  51 [ParameterType] {Name: "output", ShortFlag: "o", Type: ParameterTypeString},
```

### glazed/pkg/doc/topics/10-template-command.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/11-markdown-code-blocks.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/12-profiles-use-code.md
- Status: Update required
- Legacy match count: 2
- Pattern labels: LoadParametersFromFile, LoadParametersFromFiles
- Recommended updates:
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `LoadParametersFromFiles` with `sources.FromFiles`.
- Matches:

```text
L  71 [LoadParametersFromFile] For advanced use cases, combine profile middleware with additional config files using `LoadParametersFromFile` or `LoadParametersFromFiles`:
L  71 [LoadParametersFromFiles] For advanced use cases, combine profile middleware with additional config files using `LoadParametersFromFile` or `LoadParametersFromFiles`:
```

### glazed/pkg/doc/topics/13-layers-and-parsed-layers.md
- Status: Update required
- Legacy match count: 37
- Pattern labels: AddFlags, LoadParametersFromFile, LoadParametersFromFiles, ParameterDefinition, ParameterLayer, ParameterLayers, ParseFromCobraCommand, ParsedLayer, ParsedLayers, SetFromDefaults, UpdateFromEnv, WithParseStepSource
- Recommended updates:
  - Replace `AddFlags` with `AddFields` (schema sections now add fields).
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `LoadParametersFromFiles` with `sources.FromFiles`.
  - Replace `ParameterDefinition` with `Definition`.
  - Replace `ParameterLayer` with `Section` (schema section).
  - Replace `ParameterLayers` with `Schema`.
  - Replace `ParseFromCobraCommand` with `sources.FromCobra` (middleware).
  - Replace `ParsedLayer` with `values.SectionValues`.
  - Replace `ParsedLayers` with `values.Values`.
  - Replace `SetFromDefaults` with `sources.FromDefaults`.
  - Replace `UpdateFromEnv` with `sources.FromEnv`.
  - Replace `WithParseStepSource` with `fields.WithSource`.
- Matches:

```text
L  32 [ParameterLayer] 1. **ParameterLayer**: An interface that groups parameter definitions and provides metadata.
L  33 [ParameterLayer] 2. **ParameterLayers**: A collection of ParameterLayer objects.
L  38 [ParameterLayer] `ParameterLayer` interface.
L 201 [ParameterLayer] parameterLayers.ForEach(func(key string, p ParameterLayer) {
L 205 [ParameterLayer] err := parameterLayers.ForEachE(func(key string, p ParameterLayer) error {
L 255 [ParameterLayer] arguments, configuration files, or environment variables) using a ParameterLayer
L 258 [ParameterLayer] 1. **Layer**: A reference to the original ParameterLayer used for parsing.
L  33 [ParameterLayers] 2. **ParameterLayers**: A collection of ParameterLayer objects.
L 156 [ParameterLayers] ### Creating ParameterLayers
L 213 [ParameterLayers] Create a new ParameterLayers containing only the specified layers:
L 230 [ParameterLayers] Merge two ParameterLayers collections:
L 238 [ParameterLayers] Create a deep copy of ParameterLayers:
L 380 [ParameterLayers] - **Middleware Structure**: Each middleware processes parameters before and/or after calling the next handler in the chain. They work with `ParameterLayers` and `ParsedLayers` to manage parameter definitions and values.
L 254 [ParsedLayer] A ParsedLayer is the result of parsing input data (such as command-line
L 264 [ParsedLayer] ParsedLayers is a collection of ParsedLayer objects, typically representing all the layers used in a command or application.
L 277 [ParsedLayer] ### Creating a ParsedLayer
L 339 [ParsedLayer] parsedLayers.ForEach(func(k string, v *ParsedLayer) {
L 343 [ParsedLayer] err := parsedLayers.ForEachE(func(k string, v *ParsedLayer) error {
L 359 [ParsedLayer] Get an existing ParsedLayer or create a new one if it doesn't exist:
L 264 [ParsedLayers] ParsedLayers is a collection of ParsedLayer objects, typically representing all the layers used in a command or application.
L 266 [ParsedLayers] ### Usage of ParsedLayers
L 268 [ParsedLayers] ParsedLayers are primarily used to:
L 288 [ParsedLayers] ### Creating ParsedLayers
L 306 [ParsedLayers] ### Initializing Structs from ParsedLayers
L 321 [ParsedLayers] ### Merging ParsedLayers
L 367 [ParsedLayers] Create a deep copy of ParsedLayers:
L 380 [ParsedLayers] - **Middleware Structure**: Each middleware processes parameters before and/or after calling the next handler in the chain. They work with `ParameterLayers` and `ParsedLayers` to manage parameter definitions and values.
L  30 [ParameterDefinition] A `ParameterDefinition` defines a parameter's properties, including name, type, default value, choices, and required status.
L 260 [ParameterDefinition] - The original ParameterDefinition
L  60 [AddFlags] You can add parameters to an existing layer using the `AddFlags` method:
L  63 [AddFlags] layer.AddFlags(
L 411 [ParseFromCobraCommand] 4. **ParseFromCobraCommand**: Parses parameter values from a Cobra command, typically used for CLI applications.
L 393 [UpdateFromEnv] 2. **UpdateFromEnv**: Loads values from environment variables.
L 386 [SetFromDefaults] 1. **SetFromDefaults**: Populates parameters with their default values if no value exists.
L 400 [LoadParametersFromFile] 3. **LoadParametersFromFile / LoadParametersFromFiles**: Load parameters from JSON or YAML files.
L 400 [LoadParametersFromFiles] 3. **LoadParametersFromFile / LoadParametersFromFiles**: Load parameters from JSON or YAML files.
L 436 [WithParseStepSource] - **Source Tracking**: Use `WithParseStepSource` to track where parameter values originate.
```

### glazed/pkg/doc/topics/14-writing-help-entries.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/15-profiles.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/16-adding-parameter-types.md
- Status: Update required
- Legacy match count: 27
- Pattern labels: ParameterDefinition, ParameterDefinitions, ParameterType
- Recommended updates:
  - Replace `ParameterDefinition` with `Definition`.
  - Replace `ParameterDefinitions` with `Definitions`.
  - Replace `ParameterType*` with `fields.Type*` names.
- Matches:

```text
L  88 [ParameterDefinition] func (p *ParameterDefinition) ParseParameter(v []string, options ...ParseStepOption) (*ParsedParameter, error) {
L 136 [ParameterDefinition] func (p *ParameterDefinition) ParseFromReader(f io.Reader, filename string, options ...ParseStepOption) (*ParsedParameter, error) {
L 163 [ParameterDefinition] func (p *ParameterDefinition) CheckValueValidity(v interface{}) (interface{}, error) {
L 203 [ParameterDefinition] func (p *ParameterDefinition) InitializeValueToEmptyValue(value reflect.Value) error {
L 216 [ParameterDefinition] func (p *ParameterDefinition) SetValueFromInterface(value reflect.Value, v interface{}) error {
L 445 [ParameterDefinition] pd := &ParameterDefinition{
L 237 [ParameterDefinitions] func (ps *ParameterDefinitions) AddToCobraCommand(cmd *cobra.Command) error {
L 261 [ParameterDefinitions] func (ps *ParameterDefinitions) SetupCobraCompletions(cmd *cobra.Command) error {
L  60 [ParameterType] ParameterTypeCredentials ParameterType = "credentials"
L  67 [ParameterType] func (p ParameterType) IsList() bool {
L  69 [ParameterType] case ParameterTypeCredentials:
L  94 [ParameterType] case ParameterTypeCredentials:
L 142 [ParameterType] case ParameterTypeCredentials:
L 169 [ParameterType] case ParameterTypeCredentials:
L 207 [ParameterType] case ParameterTypeCredentials:
L 222 [ParameterType] case ParameterTypeCredentials:
L 243 [ParameterType] case ParameterTypeCredentials:
L 265 [ParameterType] case ParameterTypeCredentials:
L 278 [ParameterType] func RenderValue(parameterType ParameterType, value interface{}) (string, error) {
L 282 [ParameterType] case ParameterTypeCredentials:
L 371 [ParameterType] Add a field to the `ParameterTypesSettings` struct:
L 374 [ParameterType] type ParameterTypesSettings struct {
L 447 [ParameterType] Type: ParameterTypeCredentials,
L 471 [ParameterType] ParameterTypeCredentials ParameterType = "credentials"
L 474 [ParameterType] func (p ParameterType) IsKeyValue() bool {
L 476 [ParameterType] case ParameterTypeKeyValue, ParameterTypeCredentials:
L 508 [ParameterType] 1. **Consistent naming**: Use the pattern `ParameterType<Name>` for constants
```

### glazed/pkg/doc/topics/16-parsing-parameters.md
- Status: Update required
- Legacy match count: 13
- Pattern labels: ParameterDefinition, ParameterDefinitions, ParameterType, pkg/cmds/parameters import
- Recommended updates:
  - Replace `ParameterDefinition` with `Definition`.
  - Replace `ParameterDefinitions` with `Definitions`.
  - Replace `ParameterType*` with `fields.Type*` names.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L  91 [pkg/cmds/parameters import] import "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L  21 [ParameterDefinition] ### ParameterDefinition
L  23 [ParameterDefinition] A `ParameterDefinition` defines a parameter's properties, including name, type, default value, choices, and required status.
L  26 [ParameterDefinition] type ParameterDefinition struct {
L  40 [ParameterDefinition] `ParameterDefinitions` is an ordered map of `ParameterDefinition` instances, indexed by name.
L  44 [ParameterDefinition] *orderedmap.OrderedMap[string, *ParameterDefinition]
L  50 [ParameterDefinition] A `ParsedParameter` contains the parsed value, its `ParameterDefinition`, and a log of parsing steps.
L  55 [ParameterDefinition] ParameterDefinition *ParameterDefinition
L  88 [ParameterDefinition] Define parameters using `ParameterDefinition`, specifying name, type, and options like default values or choices.
L  38 [ParameterDefinitions] ### ParameterDefinitions
L  40 [ParameterDefinitions] `ParameterDefinitions` is an ordered map of `ParameterDefinition` instances, indexed by name.
L  43 [ParameterDefinitions] type ParameterDefinitions struct {
L  29 [ParameterType] Type       ParameterType `yaml:"type"`
```

### glazed/pkg/doc/topics/17-processor.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/18-lua.md
- Status: Update required
- Legacy match count: 3
- Pattern labels: ParsedLayer, ParsedLayers
- Recommended updates:
  - Replace `ParsedLayer` with `values.SectionValues`.
  - Replace `ParsedLayers` with `values.Values`.
- Matches:

```text
L  95 [ParsedLayer] Parses a Lua table into a ParsedLayer.
L  65 [ParsedLayers] Middleware to parse nested Lua tables into ParsedLayers.
L 101 [ParsedLayers] Parses a nested Lua table into ParsedLayers.
```

### glazed/pkg/doc/topics/19-writing-yaml-commands.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/20-using-multi-loader.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/21-cmds-middlewares.md
- Status: Update required
- Legacy match count: 39
- Pattern labels: ExecuteMiddlewares, LoadParametersFromFile, ParameterLayer, ParameterLayers, ParseFromCobraCommand, ParsedLayers, SetFromDefaults, UpdateFromEnv, WithParseStepSource, pkg/cmds/layers import, pkg/cmds/middlewares import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `ExecuteMiddlewares` with `sources.Execute`.
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `ParameterLayer` with `Section` (schema section).
  - Replace `ParameterLayers` with `Schema`.
  - Replace `ParseFromCobraCommand` with `sources.FromCobra` (middleware).
  - Replace `ParsedLayers` with `values.Values`.
  - Replace `SetFromDefaults` with `sources.FromDefaults`.
  - Replace `UpdateFromEnv` with `sources.FromEnv`.
  - Replace `WithParseStepSource` with `fields.WithSource`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/sources`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 756 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 757 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 758 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 750 [ParameterLayer] The foundation of Glazed's parameter system is the `ParameterLayer`. Before we can use middlewares, we need to define our parameter structure. This example shows how to create a layer that matches the architectural concepts discussed earlier:
L  40 [ParameterLayers] ### Relationship between ParameterLayers and ParsedLayers
L  42 [ParameterLayers] - **ParameterLayers**: These are collections of parameter definitions. They define the structure and metadata of parameters, such as their names, types, and default values.
L  70 [ParameterLayers] - `ParameterLayers`: Contains parameter definitions
L  40 [ParsedLayers] ### Relationship between ParameterLayers and ParsedLayers
L  44 [ParsedLayers] - **ParsedLayers**: These are collections of parsed parameter values. They store the actual values obtained from various sources like command-line arguments, environment variables, or configuration files.
L  71 [ParsedLayers] - `ParsedLayers`: Stores the actual parameter values
L 833 [ParsedLayers] - Creation of empty ParsedLayers to store results
L  10 [ExecuteMiddlewares] - ExecuteMiddlewares
L  75 [ExecuteMiddlewares] Middlewares are executed in reverse order of how they're provided to `ExecuteMiddlewares`. For example:
L  78 [ExecuteMiddlewares] ExecuteMiddlewares(layers, parsedLayers,
L 443 [ExecuteMiddlewares] ExecuteMiddlewares(layers, parsedLayers,
L 451 [ExecuteMiddlewares] 3. **Error Handling**: Always check for errors returned by `ExecuteMiddlewares`:
L  14 [ParseFromCobraCommand] - ParseFromCobraCommand
L 447 [ParseFromCobraCommand] ParseFromCobraCommand(),     // Most specific
L  12 [UpdateFromEnv] - UpdateFromEnv
L  80 [UpdateFromEnv] UpdateFromEnv("APP"),
L  87 [UpdateFromEnv] 2. UpdateFromEnv
L 106 [UpdateFromEnv] Use `UpdateFromEnv` to load values from environment variables:
L 445 [UpdateFromEnv] UpdateFromEnv("APP"),        // More specific
L  11 [SetFromDefaults] - SetFromDefaults
L  79 [SetFromDefaults] SetFromDefaults(),
L  88 [SetFromDefaults] 3. SetFromDefaults
L  94 [SetFromDefaults] Use `SetFromDefaults` to populate parameters with their default values:
L 444 [SetFromDefaults] SetFromDefaults(),           // Most general
L 802 [SetFromDefaults] #### SetFromDefaults Middleware
L 804 [SetFromDefaults] The `SetFromDefaults` middleware demonstrates the basic middleware pattern of processing parameters after the next handler:
L  13 [LoadParametersFromFile] - LoadParametersFromFile
L  81 [LoadParametersFromFile] LoadParametersFromFile("config.yaml"),
L  86 [LoadParametersFromFile] 1. LoadParametersFromFile
L 120 [LoadParametersFromFile] Load parameters from JSON or YAML files using `LoadParametersFromFile`:
L 130 [LoadParametersFromFile] By default, `LoadParametersFromFile` expects the config file to have this structure:
L 446 [LoadParametersFromFile] LoadParametersFromFile(),    // More specific
L 688 [LoadParametersFromFile] if commandSettings.LoadParametersFromFile != "" {
L 690 [LoadParametersFromFile] sources.FromFile(commandSettings.LoadParametersFromFile))
L 439 [WithParseStepSource] 1. **Source Tracking**: Always specify the source using `WithParseStepSource` to track where values came from.
```

### glazed/pkg/doc/topics/22-command-loaders.md
- Status: Update required
- Legacy match count: 1
- Pattern labels: pkg/cmds/parameters import
- Recommended updates:
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 146 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
```

### glazed/pkg/doc/topics/22-templating-helpers.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/23-pattern-based-config-mapping.md
- Status: Update required
- Legacy match count: 16
- Pattern labels: LoadParametersFromFile, pkg/cmds/layers import, pkg/cmds/middlewares import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/sources`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L  56 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L  85 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 137 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 476 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 479 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L  57 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L  58 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L  86 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L  87 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L 138 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 139 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L 477 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 478 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L  26 [LoadParametersFromFile] - Using the Mapper: Wire into `LoadParametersFromFile`
L  70 [LoadParametersFromFile] // Use with LoadParametersFromFile
L  98 [LoadParametersFromFile] // Use with LoadParametersFromFile
```

### glazed/pkg/doc/topics/24-config-files.md
- Status: Update required
- Legacy match count: 22
- Pattern labels: LoadParametersFromFile, LoadParametersFromFiles, pkg/cmds/layers import, pkg/cmds/middlewares import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `LoadParametersFromFiles` with `sources.FromFiles`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/sources`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L  31 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L  73 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 110 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 129 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 171 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 216 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 308 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 352 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L  33 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L  74 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 219 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 309 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L  32 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 111 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 130 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 210 [pkg/cmds/middlewares import] Use `github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper` to declare mapping rules and pass the mapper to `LoadParametersFromFile`.
L 217 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 218 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L 256 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 353 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L 210 [LoadParametersFromFile] Use `github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper` to declare mapping rules and pass the mapper to `LoadParametersFromFile`.
L 390 [LoadParametersFromFiles] Legacy Viper-based middlewares like `GatherFlagsFromViper` and per-command `--load-parameters-from-file` are deprecated. Prefer config middlewares (`LoadParametersFromFiles`) with resolvers and `--config-file`.
```

### glazed/pkg/doc/topics/commands-reference.md
- Status: Update required
- Legacy match count: 21
- Pattern labels: ParameterType, ParsedLayers, pkg/cmds/layers import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `ParameterType*` with `fields.Type*` names.
  - Replace `ParsedLayers` with `values.Values`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L  96 [pkg/cmds/layers import] - `github.com/go-go-golems/glazed/pkg/cmds/layers`: Parameter layering system
L  95 [pkg/cmds/parameters import] - `github.com/go-go-golems/glazed/pkg/cmds/parameters`: Parameter types and definitions
L  56 [ParsedLayers] │  Parameters  │          │ ParsedLayers    │
L  87 [ParsedLayers] 5. **ParsedLayers**: Runtime values after collecting from CLI flags, environment, config files, and defaults
L 784 [ParameterType] Glazed treats command-line parameters as more than just strings. They are typed objects with built-in validation, default values, and help text. This approach shifts the burden of parsing and validation from the command's business logic to the framework itself. By defining a parameter's type (e.g., `ParameterTypeInteger`, `ParameterTypeDate`, `ParameterTypeFile`), you get automatic error handling and a more robust and user-friendly CLI.
L 793 [ParameterType] **`ParameterTypeString`**: The workhorse for text inputs - names, descriptions, URLs
L 794 [ParameterType] **`ParameterTypeSecret`**: Like strings, but values are masked in help and logs (perfect for passwords, API keys)
L 795 [ParameterType] **`ParameterTypeInteger`**: Whole numbers with automatic range validation
L 796 [ParameterType] **`ParameterTypeFloat`**: Decimal numbers for measurements, percentages, ratios
L 797 [ParameterType] **`ParameterTypeBool`**: True/false flags that work with `--flag` and `--no-flag` patterns
L 798 [ParameterType] **`ParameterTypeDate`**: Intelligent date parsing that handles multiple formats
L 801 [ParameterType] **`ParameterTypeStringList`**: Multiple values like `--tag web --tag api --tag production`
L 802 [ParameterType] **`ParameterTypeIntegerList`**: Lists of numbers for ports, IDs, quantities
L 803 [ParameterType] **`ParameterTypeFloatList`**: Multiple decimal values for coordinates, measurements
L 806 [ParameterType] **`ParameterTypeChoice`**: Single selection from predefined options (with tab completion!)
L 807 [ParameterType] **`ParameterTypeChoiceList`**: Multiple selections from predefined options
L 810 [ParameterType] **`ParameterTypeFile`**: File paths with existence validation and tab completion
L 811 [ParameterType] **`ParameterTypeFileList`**: Multiple file paths
L 812 [ParameterType] **`ParameterTypeStringFromFile`**: Read text content from a file (useful for large inputs)
L 813 [ParameterType] **`ParameterTypeStringListFromFile`**: Read line-separated lists from files
L 816 [ParameterType] **`ParameterTypeKeyValue`**: Map-like inputs: `--env DATABASE_URL=postgres://... --env DEBUG=true`
```

### glazed/pkg/doc/topics/how-to-write-good-documentation-pages.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/layers-guide.md
- Status: Update required
- Legacy match count: 11
- Pattern labels: AddFlags, pkg/cmds/layers import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `AddFlags` with `AddFields` (schema sections now add fields).
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 622 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 623 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 292 [AddFlags] extendedDbLayer.AddFlags(
L 551 [AddFlags] b.layer.AddFlags(
L 560 [AddFlags] b.layer.AddFlags(
L 569 [AddFlags] b.layer.AddFlags(
L1127 [AddFlags] advancedLayer.AddFlags(
L1159 [AddFlags] layer.AddFlags(
L1166 [AddFlags] layer.AddFlags(
L1179 [AddFlags] layer.AddFlags(
L1185 [AddFlags] layer.AddFlags(
```

### glazed/pkg/doc/topics/logging-layer.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/simple-query-dsl.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/user-query-dsl.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/topics/using-the-query-api.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/tutorials/01-a-simple-table-cli.md
- Status: Update required
- Legacy match count: 1
- Pattern labels: ParameterDefinition
- Recommended updates:
  - Replace `ParameterDefinition` with `Definition`.
- Matches:

```text
L  20 [ParameterDefinition] - `ParameterDefinition`: This struct is used to define the parameters (flags or arguments) that the command takes. It
```

### glazed/pkg/doc/tutorials/02-a-simple-help-system.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/tutorials/04-lua.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/tutorials/05-build-first-command.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/tutorials/config-files-quickstart.md
- Status: No legacy hits detected (spot-check for conceptual drift)
- Legacy match count: 0
- Pattern labels: none
- Recommended updates: none (scan found no deprecated identifiers).
- Matches:
  - None

### glazed/pkg/doc/tutorials/custom-layer.md
- Status: Update required
- Legacy match count: 5
- Pattern labels: pkg/cmds/layers import, pkg/cmds/parameters import
- Recommended updates:
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 271 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 421 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 493 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 272 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 494 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
```

### glazed/pkg/doc/tutorials/migrating-from-viper-to-config-files.md
- Status: Update required
- Legacy match count: 20
- Pattern labels: LoadParametersFromFile, LoadParametersFromFiles, UpdateFromEnv, pkg/cmds/layers import, pkg/cmds/middlewares import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `LoadParametersFromFiles` with `sources.FromFiles`.
  - Replace `UpdateFromEnv` with `sources.FromEnv`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/sources`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 360 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 420 [pkg/cmds/layers import] "github.com/go-go-golems/glazed/pkg/cmds/layers"
L 121 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 146 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 248 [pkg/cmds/parameters import] "github.com/go-go-golems/glazed/pkg/cmds/parameters"
L 120 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 145 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 213 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 247 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 481 [pkg/cmds/middlewares import] pm "github.com/go-go-golems/glazed/pkg/cmds/middlewares/patternmapper"
L 793 [pkg/cmds/middlewares import] "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
L 765 [UpdateFromEnv] 3. `UpdateFromEnv` middleware is included in your middleware chain
L  27 [LoadParametersFromFile] 1. **Config File Loading**: Replace `GatherFlagsFromViper()` and `GatherFlagsFromCustomViper()` with `LoadParametersFromFile()` or `LoadParametersFromFiles()`
L 141 [LoadParametersFromFile] ### After: Using LoadParametersFromFile
L 171 [LoadParametersFromFile] For applications with a single config file, use `LoadParametersFromFile`:
L 498 [LoadParametersFromFile] // Use with LoadParametersFromFile
L  27 [LoadParametersFromFiles] 1. **Config File Loading**: Replace `GatherFlagsFromViper()` and `GatherFlagsFromCustomViper()` with `LoadParametersFromFile()` or `LoadParametersFromFiles()`
L 191 [LoadParametersFromFiles] For applications that compose configuration from multiple files, use `LoadParametersFromFiles`:
L 243 [LoadParametersFromFiles] ### After: Using LoadParametersFromFiles
L 771 [LoadParametersFromFiles] 2. Check config file order in `LoadParametersFromFiles` (low → high)
```

### glazed/pkg/doc/tutorials/migrating-to-facade-packages.md
- Status: Deprecated — remove or replace
- Legacy match count: 36
- Pattern labels: AddFlags, CobraParameterLayer, CommandDefinition, ExecuteMiddlewares, GatherArguments, LoadParametersFromFile, ParameterDefinition, ParameterDefinitions, ParameterLayer, ParameterLayers, ParameterType, ParseFromCobraCommand, ParsedLayer, ParsedLayers, SetFromDefaults, UpdateFromEnv, WithParseStepSource, layers.ParameterLayer, layers.ParameterLayers, parameters.ParameterDefinition, parameters.ParameterDefinitions, pkg/cmds/layers import, pkg/cmds/middlewares import, pkg/cmds/parameters import
- Recommended updates:
  - Replace `AddFlags` with `AddFields` (schema sections now add fields).
  - Replace `CobraParameterLayer` with `schema.CobraSection`.
  - Replace `CommandDefinition` with `CommandDescription` if applicable.
  - Replace `ExecuteMiddlewares` with `sources.Execute`.
  - Ensure usage is `fields.Definitions.GatherArguments` (schema/fields naming).
  - Replace `LoadParametersFromFile` with `sources.FromFile`.
  - Replace `ParameterDefinition` with `Definition`.
  - Replace `ParameterDefinitions` with `Definitions`.
  - Replace `ParameterLayer` with `Section` (schema section).
  - Replace `ParameterLayers` with `Schema`.
  - Replace `ParameterType*` with `fields.Type*` names.
  - Replace `ParseFromCobraCommand` with `sources.FromCobra` (middleware).
  - Replace `ParsedLayer` with `values.SectionValues`.
  - Replace `ParsedLayers` with `values.Values`.
  - Replace `SetFromDefaults` with `sources.FromDefaults`.
  - Replace `UpdateFromEnv` with `sources.FromEnv`.
  - Replace `WithParseStepSource` with `fields.WithSource`.
  - Replace `layers.ParameterLayer` with `schema.Section`.
  - Replace `layers.ParameterLayers` with `schema.Schema`.
  - Replace `parameters.ParameterDefinition` with `fields.Definition`.
  - Replace `parameters.ParameterDefinitions` with `fields.Definitions`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/schema`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/sources`.
  - Update import path to `github.com/go-go-golems/glazed/pkg/cmds/fields`.
- Matches:

```text
L 107 [pkg/cmds/layers import] - `github.com/go-go-golems/glazed/pkg/cmds/layers`
L 108 [pkg/cmds/parameters import] - `github.com/go-go-golems/glazed/pkg/cmds/parameters`
L 109 [pkg/cmds/middlewares import] - `github.com/go-go-golems/glazed/pkg/cmds/middlewares`
L  42 [layers.ParameterLayer] - `pkg/cmds/layers.ParameterLayer` → `pkg/cmds/schema.Section`
L  43 [layers.ParameterLayer] - `pkg/cmds/layers.ParameterLayers` → `pkg/cmds/schema.Schema`
L  88 [layers.ParameterLayer] - `cmds.CommandDescription.Layers` is now `*schema.Schema` (was `*layers.ParameterLayers`).
L  89 [layers.ParameterLayer] - `layers.ParameterLayer` interface methods now use `*fields.Definition` / `*fields.Definitions`:
L  43 [layers.ParameterLayers] - `pkg/cmds/layers.ParameterLayers` → `pkg/cmds/schema.Schema`
L  88 [layers.ParameterLayers] - `cmds.CommandDescription.Layers` is now `*schema.Schema` (was `*layers.ParameterLayers`).
L  42 [ParameterLayer] - `pkg/cmds/layers.ParameterLayer` → `pkg/cmds/schema.Section`
L  89 [ParameterLayer] - `layers.ParameterLayer` interface methods now use `*fields.Definition` / `*fields.Definitions`:
L  43 [ParameterLayers] - `pkg/cmds/layers.ParameterLayers` → `pkg/cmds/schema.Schema`
L  88 [ParameterLayers] - `cmds.CommandDescription.Layers` is now `*schema.Schema` (was `*layers.ParameterLayers`).
L  51 [ParsedLayer] - `pkg/cmds/layers.ParsedLayer` → `pkg/cmds/values.SectionValues`
L  50 [ParsedLayers] - `pkg/cmds/layers.ParsedLayers` → `pkg/cmds/values.Values`
L  79 [ParsedLayers] …and it still satisfies interfaces that mention `*layers.ParsedLayers`, because `values.Values` is an alias for `layers.ParsedLayers`.
L 165 [ParsedLayers] func (c *MyCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
L  44 [parameters.ParameterDefinition] - `pkg/cmds/parameters.ParameterDefinition` → `pkg/cmds/fields.Definition`
L  45 [parameters.ParameterDefinition] - `pkg/cmds/parameters.ParameterDefinitions` → `pkg/cmds/fields.Definitions`
L  45 [parameters.ParameterDefinitions] - `pkg/cmds/parameters.ParameterDefinitions` → `pkg/cmds/fields.Definitions`
L  44 [ParameterDefinition] - `pkg/cmds/parameters.ParameterDefinition` → `pkg/cmds/fields.Definition`
L  45 [ParameterDefinitions] - `pkg/cmds/parameters.ParameterDefinitions` → `pkg/cmds/fields.Definitions`
L  46 [ParameterType] - `pkg/cmds/parameters.ParameterType*` → `pkg/cmds/fields.Type*`
L 125 [ParameterType] parameters.NewParameterDefinition("limit", parameters.ParameterTypeInteger, parameters.WithDefault(10))
L 144 [ParameterType] parameters.NewParameterDefinition("api-key", parameters.ParameterTypeString),
L  90 [AddFlags] - `AddFlags(...*fields.Definition)`
L 221 [CobraParameterLayer] - Cobra-only plumbing: attaching layers to Cobra uses `layers.CobraParameterLayer`.
L  66 [ExecuteMiddlewares] - `middlewares.ExecuteMiddlewares` → `sources.Execute`
L  95 [ExecuteMiddlewares] - `middlewares.HandlerFunc` / `middlewares.ExecuteMiddlewares` (now `sources.Execute`)
L  60 [ParseFromCobraCommand] - `middlewares.ParseFromCobraCommand` → `sources.FromCobra`
L  61 [GatherArguments] - `middlewares.GatherArguments` → `sources.FromArgs`
L  62 [UpdateFromEnv] - `middlewares.UpdateFromEnv` → `sources.FromEnv`
L  63 [SetFromDefaults] - `middlewares.SetFromDefaults` → `sources.FromDefaults`
L  64 [LoadParametersFromFile] - `middlewares.LoadParametersFromFile(s)` → `sources.FromFile` / `sources.FromFiles`
L  67 [WithParseStepSource] - `parameters.WithParseStepSource(...)` → `sources.WithSource(...)`
L  87 [CommandDefinition] - `cmds.CommandDefinition` and `cmds.CommandDefinitionOption` were removed. Use `cmds.CommandDescription` and `cmds.CommandDescriptionOption` instead.
```
