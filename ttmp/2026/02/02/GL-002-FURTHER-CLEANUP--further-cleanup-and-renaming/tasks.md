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
