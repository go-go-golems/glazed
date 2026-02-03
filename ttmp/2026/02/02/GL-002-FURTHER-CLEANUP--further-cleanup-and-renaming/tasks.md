# Tasks

## TODO

- [x] Define canonical struct tag (`glazed:`) and identify all existing `glazed` tag usage (code + docs).
- [x] Update tag parsing/decoding code paths to accept only `glazed:` (no backwards compatibility).
- [x] Update AST migration tool to rewrite struct tags + key renames (Layer/Parameter → Section/Field) in Go examples.
- [x] Run AST tool on `cmd/examples` + `pkg/doc` code snippets as needed; review & fix any misses.
- [x] Replace remaining `glazed` tag mentions in docs/tutorials.
- [x] Re-run `rg -n -i "parameter|layer"` to validate removal scope (exclude non-API historical docs).
- [x] Run `go test ./...` to confirm build.
- [x] Update diary + changelog, check off tasks, and commit.

## TODO (cleanup pass)

- [x] Generate symbol inventory for all non-ttmp files mentioning parameter/layer.
- [x] Rename `AddLayerToCobraCommand` -> `AddSectionToCobraCommand` (definitions + call sites).
- [x] Store temporary scripts in ticket `scripts/` (inventory + audit tooling).
- [x] Expand migration/refactor tooling to rename ParsedParameter(s) → FieldValue(s) and helper APIs.
- [x] Rename `fields` parsed value types + helpers (FieldValue(s), Decode helpers) and update call sites.
- [x] Rename `values` API surface to Section/Field naming (fields, options, accessors, errors).
- [ ] Rename pattern mapper config (TargetParameter → TargetField, YAML/JSON tags, files, tests).
- [ ] Rename settings/helpers/tests using Parameter/Layer names in identifiers + strings.
- [ ] Rename files/dirs containing `layer` or `parameter` in names (code + docs + examples).
- [ ] Update docs/tutorials/examples/prompto/pinocchio to remove `layer`/`parameter` vocabulary.
- [ ] Update fixtures/YAML/README/changelogs that embed old naming.
- [ ] Re-run `go test ./...` (and linters if required by hooks).
- [ ] Update diary + changelog + index, then commit and push.
