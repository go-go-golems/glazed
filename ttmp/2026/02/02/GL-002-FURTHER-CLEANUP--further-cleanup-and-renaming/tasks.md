# Tasks

## TODO (active)

- [x] Rename command description API: `Layers` → `Schema`, `WithLayers*` → `WithSections*`/`WithSchema`, `GetLayer*` → `GetSection*`, `SetLayers` → `SetSections`, remove `NewCommandDefinition`, update YAML tags.
- [x] Update callers in cmd/pkg (cli, help, lua, runner, template/json-schema, logging, alias) to new schema/section API and remove `layer`/`parameter` identifiers in Go code.
- [x] Rename logging layer surface (`LoggingLayer*`, `AddLoggingLayer*`) to section naming and update docs/tests.
- [x] Update examples (cmd/examples + cmd/glaze) to section/field naming and rename files/dirs that mention parameter/layer.
- [x] Update docs/tutorials/prompto/pinocchio/README/changelog to remove `layer`/`parameter` vocabulary and adjust filenames/links.
- [x] Run gofmt + `go test ./...` to confirm build.
- [ ] Update diary + changelog + index, then commit (no push).

## DONE

- [x] Define canonical struct tag (`glazed:`) and identify all existing `glazed` tag usage (code + docs).
- [x] Update tag parsing/decoding code paths to accept only `glazed:` (no backwards compatibility).
- [x] Update AST migration tool to rewrite struct tags + key renames (Layer/Parameter → Section/Field) in Go examples.
- [x] Run AST tool on `cmd/examples` + `pkg/doc` code snippets as needed; review & fix any misses.
- [x] Replace remaining `glazed` tag mentions in docs/tutorials.
- [x] Re-run `rg -n -i "parameter|layer"` to validate removal scope (exclude non-API historical docs).
- [x] Generate symbol inventory for all non-ttmp files mentioning parameter/layer.
- [x] Rename `AddLayerToCobraCommand` -> `AddSectionToCobraCommand` (definitions + call sites).
- [x] Store temporary scripts in ticket `scripts/` (inventory + audit tooling).
- [x] Expand migration/refactor tooling to rename ParsedParameter(s) → FieldValue(s) and helper APIs.
- [x] Rename `fields` parsed value types + helpers (FieldValue(s), Decode helpers) and update call sites.
- [x] Rename `values` API surface to Section/Field naming (fields, options, accessors, errors).
- [x] Publish inventory report from scripts (analysis/03-layer-parameter-inventory.md).
- [x] Pattern mapper rename (TargetParameter → TargetField) including config tags, error strings, and tests.
- [x] Appconfig layer vocabulary cleanup (LayerSlug → SectionSlug, WithValuesForLayers → WithValuesForSections, error strings, tests, examples).
- [x] Lua conversion naming cleanup (ParseLuaTableToSection, ParseFieldFromLua, error strings).
- [x] CLI flag + help text cleanup (parsed-parameters → parsed-fields; dropped legacy load-from-file flag).
- [x] Remove legacy cobra builder wrappers + update parser config fields in examples/ttmp.
- [x] Rename settings/helpers/tests using Parameter/Layer names in identifiers + strings.
- [x] Rename sources layer helpers to section naming and update YAML fixtures.
- [x] Rename field type naming (ParameterType → FieldType) across code/examples/docs.
