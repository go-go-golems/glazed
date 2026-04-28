# Tasks

## TODO

- [x] Add implementation tasks to ticket
- [x] Task 1: Create `pkg/help/cmd/export.go` — ExportCommand with tabular output (`--format glazed`)
- [x] Define `ExportCommand` implementing `cmds.BareCommand`
- [x] Add glazed section to schema for `--output json/csv/table/yaml` flags
- [x] Implement `Run()` with `--format glazed` path: create processor, emit rows
- [x] Support `--with-content` flag (default `true`)
- [x] Support filter flags (`--type`, `--topic`, `--command`, `--flag`, `--slug`)
- [x] Implement `buildExportPredicate` helper
- [x] Add `AddExportCommand` helper for wiring into any binary
- [x] Task 2: Add disk-export modes (`--format files` and `--format sqlite`)
- [x] Implement `exportToFiles` with directory creation and markdown reconstruction
- [x] Implement `exportToSQLite` using `store.New` and `Upsert`
- [x] Support `--flatten` flag for files mode
- [x] Task 3: Add unit tests
- [x] Test tabular output with `--with-content=true` (default)
- [x] Test tabular output with `--with-content=false`
- [x] Test filtering by type, topic, command, flag, slug
- [x] Test `--format files` round-trip (export then re-parse)
- [x] Test `--format sqlite` (open DB and query)
- [x] Task 4: Wire into `cmd/glaze/main.go`
- [x] Find `help` subcommand and add `export` via `AddExportCommand`
- [x] Verify `glaze help export --help` shows correct flags
- [x] Task 5: Add documentation help section
- [x] Create `pkg/doc/help-export.md` with usage examples
- [x] Verify it loads into the help system
- [x] Task 6: Validate and commit
- [x] Run `go test ./pkg/help/cmd/...`
- [x] Run `go run ./cmd/glaze help export --help`
- [x] Manual verification checklist
- [x] Commit each task separately
- [x] Task 31: Update serve external sources design review: add --from-glazed-cmd and --with-embedded=false default
- [x] Task 32: Implement robust JSON section import for real help export rows (type/section_type compatibility)
- [x] Task 33: Implement ContentLoader source loaders for markdown paths, JSON, SQLite, arbitrary command, and glazed command shorthand
- [x] Task 34: Extend ServeCommand flags/settings/run flow with --from-json, --from-sqlite, --from-cmd, --from-glazed-cmd, --with-embedded=false
- [x] Task 35: Add unit/integration tests for loaders and ServeCommand source loading
- [x] Task 36: Update docs/diary/changelog, validate, commit, and upload updated bundle to reMarkable
