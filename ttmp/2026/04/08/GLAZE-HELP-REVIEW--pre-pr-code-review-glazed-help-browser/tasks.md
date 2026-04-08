# Tasks

## TODO

### Phase 1: Delete dead code (zero risk, pure deletion)

- [x] **T1: Delete `pkg/help/store/compat.go`**
  - The `store.HelpSystem` struct is never imported or used anywhere in the codebase. `cmd/glaze/main.go` uses `help.HelpSystem` directly.
  - Delete the entire file.
  - Run `go build ./...` to confirm nothing breaks.
  - Files: `pkg/help/store/compat.go`

- [x] **T2: Delete `pkg/help/example_store_usage.go`**
  - `ExampleStoreUsage()` is a dead example function never called from anywhere.
  - Delete the entire file.
  - Files: `pkg/help/example_store_usage.go`

- [x] **T3: Remove dead `HelpSystem interface{}` field from `model.Section`**
  - The field `HelpSystem interface{} \`json:"-" yaml:"-"\`` at `pkg/help/model/section.go:47` is never set and never read.
  - Remove the field and its comment.
  - Remove the `IsForCommand`, `IsForFlag`, `IsForTopic` methods from `model.Section` — they are duplicated on `help.Section` which is the actual wrapper used by callers. (Verify no external caller uses the model-level methods first.)
  - Files: `pkg/help/model/section.go`

- [x] **T4: Remove backward-compat re-exports from `help.go`**
  - The type aliases and const re-exports at the top of `pkg/help/help.go` are legacy:
    ```go
    type SectionType = model.SectionType
    const (
        SectionGeneralTopic = model.SectionGeneralTopic
        SectionExample      = model.SectionExample
        SectionApplication  = model.SectionApplication
        SectionTutorial     = model.SectionTutorial
    )
    var SectionTypeFromString = model.SectionTypeFromString
    ```
  - Replace all usages across the codebase with direct `model.SectionGeneralTopic`, `model.SectionType`, `model.SectionTypeFromString` references.
  - Key callers: `pkg/help/help.go`, `pkg/help/query.go`, `pkg/help/dsl_bridge.go`, `pkg/help/render.go`, `pkg/help/cmd/cobra.go`, `pkg/help/ui/model.go`.
  - Files: `pkg/help/help.go`, and every file that uses the re-exported names.

### Phase 2: Fix the DSL bridge performance bug

- [x] **T5: Rewrite `QuerySections()` to use `store.Find()` directly**
  - Current code in `pkg/help/dsl_bridge.go` calls `evaluatePredicate()` which creates a new in-memory SQLite database per section — O(N) databases for N sections.
  - Replace the entire `QuerySections → evaluatePredicate` path with a direct `store.Find(ctx, predicate)` call, matching the pattern already used by `queryLegacy()`.
  - The `evaluatePredicate()` method and the temp-store creation should be deleted entirely.
  - Update `dsl_bridge_test.go` if it tests the old path.
  - Files: `pkg/help/dsl_bridge.go`, `pkg/help/dsl_bridge_test.go`

### Phase 3: Consolidate markdown parsing (eliminate duplication)

- [x] **T6: Extract `ParseSectionFromMarkdown()` into `model/` package**
  - Create a new function `model.ParseSectionFromMarkdown(data []byte) (*Section, error)` in `pkg/help/model/parse.go`.
  - This function should be the superset of both existing parsers:
    - Initialize empty slices for `Topics`, `Flags`, `Commands` when absent (from `loader.go`)
    - Handle both `int` and `float64` for `Order` field (from `loader.go`)
    - Default `SectionType` to `SectionGeneralTopic` when absent
    - Validate slug and title are present
  - Files: `pkg/help/model/parse.go` (new)

- [x] **T7: Replace `help.LoadSectionFromMarkdown()` with delegation to `model.ParseSectionFromMarkdown()`**
  - The current `LoadSectionFromMarkdown()` in `pkg/help/help.go` should become a thin wrapper:
    ```go
    func LoadSectionFromMarkdown(data []byte) (*Section, error) {
        modelSection, err := model.ParseSectionFromMarkdown(data)
        if err != nil { return nil, err }
        return &Section{Section: modelSection}, nil
    }
    ```
  - Or delete it entirely and have callers use the model function + wrap.
  - Remove the `frontmatter` import from `help.go` if no longer needed.
  - Files: `pkg/help/help.go`

- [x] **T8: Replace `store.Loader.LoadFromMarkdown()` with delegation to `model.ParseSectionFromMarkdown()`**
  - The current `LoadFromMarkdown()` in `pkg/help/store/loader.go` should delegate:
    ```go
    func (l *Loader) LoadFromMarkdown(data []byte) (*model.Section, error) {
        return model.ParseSectionFromMarkdown(data)
    }
    ```
  - Or inline the call at the two call sites (`loadFileFromFS`, `LoadSections`) and delete the method.
  - Remove the `frontmatter` and `strings2` imports from `loader.go`.
  - Files: `pkg/help/store/loader.go`

### Phase 4: Eliminate the `help.Section` wrapper

- [x] **T9: Flatten `HelpPage` to use `[]*model.Section` instead of `[]*help.Section`**
  - The `HelpPage` struct in `pkg/help/help.go` currently holds `[]*Section` fields. Change all fields to `[]*model.Section`.
  - `NewHelpPage(sections []*Section)` → `NewHelpPage(sections []*model.Section)`.
  - The struct fields (`DefaultGeneralTopics`, `AllExamples`, etc.) all become `[]*model.Section`.
  - The template data key `"Help"` currently receives `*HelpPage` — this still works because the fields are the same, just the element type changes.
  - Files: `pkg/help/help.go`

- [x] **T10: Remove the `help.Section` wrapper struct entirely**
  - After T9, `help.Section` is just `model.Section` with an unused `HelpSystem` back-reference.
  - Replace all `*help.Section` with `*model.Section` throughout the codebase:
    - `pkg/help/help.go` — delete the `Section` struct, update `HelpPage`, `LoadSectionFromMarkdown`, etc.
    - `pkg/help/query.go` — `FindSections` returns `[]*model.Section`
    - `pkg/help/dsl_bridge.go` — all return types
    - `pkg/help/render.go` — `RenderTopicHelp` takes `*model.Section`
    - `pkg/help/cmd/cobra.go` — all usage
    - `pkg/help/ui/model.go` — `Model.results`, `Model.CurrentSection`, `listItem.section`, etc.
    - `pkg/help/cmd/ui.go` — `RunUIWithOutput` return type
  - The `Section.IsForCommand()`, `IsForFlag()`, `IsForTopic()` methods move to `model.Section` (already exist there from T3).
  - Delete the `DefaultGeneralTopics()`, `OtherExamples()`, etc. methods — these are the back-reference query methods. Nobody calls them except through `HelpPage` construction, and `NewHelpPage` doesn't use them (it sorts directly).
  - Verify: `grep -rn "help\.Section" --include="*.go"` returns zero results.
  - Files: `pkg/help/help.go`, `pkg/help/query.go`, `pkg/help/dsl_bridge.go`, `pkg/help/render.go`, `pkg/help/cmd/cobra.go`, `pkg/help/ui/model.go`, `pkg/help/cmd/ui.go`

- [x] **T11: Update `LoadSectionsFromFS` and `AddSection` in `help.HelpSystem`**
  - `HelpSystem.LoadSectionsFromFS()` currently calls `LoadSectionFromMarkdown` then `AddSection` (which wraps in `help.Section`). After T10, it should call `model.ParseSectionFromMarkdown` and `Store.Upsert` directly.
  - `HelpSystem.AddSection()` currently takes `*Section` — change to `*model.Section`.
  - `HelpSystem.GetSectionWithSlug()` currently returns `(*Section, error)` — change to `(*model.Section, error)`.
  - Files: `pkg/help/help.go`

### Phase 5: Eliminate `SectionQuery` builder

- [x] **T12: Rewrite `cobra.go` to use `store.Predicate` directly**
  - The `NewCobraHelpCommand` and `renderCommandHelpPage` functions in `pkg/help/cmd/cobra.go` currently build `SectionQuery` objects and pass them via `RenderOptions.Query`.
  - Replace `RenderOptions.Query *SectionQuery` with a `store.Predicate` field (or pass predicates directly).
  - The cobra command builds predicates using the `store` package directly:
    - `ReturnAllTypes()` → no type filter
    - `ReturnOnlyShownByDefault()` → `store.ShownByDefault()`
    - `ReturnOnlyTopics(topic)` → `store.HasTopic(topic)`
    - `SearchForCommand(cmd)` → `store.HasCommand(cmd)`
    - `ReturnOnlyTopLevel()` → `store.IsTopLevel()`
  - Files: `pkg/help/cmd/cobra.go`, `pkg/help/render.go`

- [x] **T13: Rewrite `render.go` to use `store.Predicate` directly**
  - `ComputeRenderData(userQuery *SectionQuery)` currently calls `userQuery.FindSections(ctx, hs.Store)`.
  - Replace with `ComputeRenderData(pred store.Predicate)` that calls `hs.Store.Find(ctx, pred)`.
  - The relaxation logic (widening query when no results) should work directly with predicates: clone the predicate builder and add fewer conditions.
  - `RenderTopicHelp` signature changes to take `*model.Section` instead of `*Section`.
  - `RenderOptions.Query` becomes `RenderOptions.Predicate store.Predicate`.
  - Files: `pkg/help/render.go`

- [x] **T14: Delete `pkg/help/query.go`**
  - After T12 and T13, no caller uses `SectionQuery`.
  - Delete the entire file.
  - Run `go build ./...` to confirm.
  - Files: `pkg/help/query.go`

- [x] **T15: Update `pkg/help/ui/model.go` to use `store.Predicate`**
  - The TUI model's search currently builds a `SectionQuery`. Replace with `store.Predicate` construction.
  - `Model.results` changes from `[]*help.Section` to `[]*model.Section`.
  - `listItem.section` changes type.
  - Files: `pkg/help/ui/model.go`, `pkg/help/ui/model_test.go`

### Phase 6: Fix server search to use FTS5 abstraction

- [x] **T16: Replace inline LIKE in `buildPredicate()` with `store.TextSearch()`**
  - In `pkg/help/server/handlers.go`, the `buildPredicate()` function has an inline `LIKE` for search:
    ```go
    qc.AddWhere("LOWER(s.title) LIKE ? OR LOWER(s.short) LIKE ? OR LOWER(s.content) LIKE ?", ...)
    ```
  - Replace with `store.TextSearch(params.Search)` which correctly uses FTS5 when the build tag is enabled, or LIKE as fallback.
  - Note: `store.TextSearch` does case-insensitive matching. The current inline code lowercases manually — verify the replacement handles case correctly.
  - Files: `pkg/help/server/handlers.go`

### Phase 7: Frontend cleanup

- [x] **T17: Fix `SectionDetail` to not re-declare parent fields**
  - In `web/src/types/index.ts`, `SectionDetail extends SectionSummary` re-declares `short` and `topics`:
    ```typescript
    export interface SectionDetail extends SectionSummary {
      short: string;     // already in SectionSummary
      topics: string[];  // already in SectionSummary
    ```
  - Remove the duplicated fields. Add `flags`, `commands`, `content` (the fields that are actually new).
  - Files: `web/src/types/index.ts`

- [x] **T18: Remove unused `searchSections` RTK Query endpoint**
  - In `web/src/services/api.ts`, `searchSections` is defined but never imported or used in the UI.
  - Remove the endpoint definition and the `useSearchSectionsQuery` export.
  - Files: `web/src/services/api.ts`

### Phase 8: Final verification

- [x] **T19: Run full test suite and verify compilation**
  - `GOWORK=off go build ./...`
  - `GOWORK=off go test ./pkg/help/... ./pkg/web/... -count=1`
  - `cd web && pnpm build`
  - `GOWORK=off go generate ./pkg/web && GOWORK=off go build ./cmd/glaze`
  - Files: all changed files

- [x] **T20: Verify no legacy references remain**
  - `grep -rn "help\.Section\b" --include="*.go" | grep -v "_test.go"` → empty
  - `grep -rn "SectionQuery" --include="*.go" | grep -v "_test.go"` → empty
  - `grep -rn "store\.HelpSystem" --include="*.go"` → empty
  - `grep -rn "compat\.go" --include="*.go"` → empty
  - `grep -rn "backward\|compat\|legacy" --include="*.go" | grep -v "_test.go"` → empty or only comments explaining history

## DONE

- [x] Analyze minitrace session data (21h, 1132 turns, 114 go-builds)
- [x] Code review: duplicated markdown parsing (help.go vs loader.go)
- [x] Code review: two HelpSystem types (help.HelpSystem vs store.HelpSystem compat)
- [x] Code review: Section wrapper with back-reference vs model.Section
- [x] Code review: SectionQuery builder vs store.Predicate pattern
- [x] Code review: DSL bridge evaluatePredicate creates temp store per section
- [x] Code review: build-web 30 edits, serve.go 25 rewrites
- [x] Code review: embed/frontend symlink confusion
- [x] Write final code review findings document
