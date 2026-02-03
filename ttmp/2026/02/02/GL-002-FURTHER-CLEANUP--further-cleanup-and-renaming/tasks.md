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
- [ ] Expand migration/refactor tool with Layer→Section + Parameter→Field identifier renames (no compat).
- [ ] Run refactor tool across Go sources and fix compile errors.
- [ ] Rename files and directories containing `layer`/`parameter` in their names (code + docs + examples).
- [ ] Update docs/tutorials/examples/prompto/pinocchio content to remove `layer`/`parameter` vocabulary.
- [ ] Update tests and fixtures that embed `layer`/`parameter`.
- [ ] Re-run `go test ./...` (and linters if required by hooks).
- [ ] Update diary + changelog + index, then commit and push.
