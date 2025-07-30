# Streamlined Plan: Unifying Glazed Cobra-Command Creation

Goal: Have a **single public option type – `CobraOption`** – that configures *both* the command-builder and the underlying parser.  The legacy `BuildCobraCommandDualMode`, `DualModeOption`, and `CobraParserOption` APIs become thin, deprecated wrappers.

---

## 1. Foundation Types

- [x] **1.1 `CobraParserConfig` (new)**  ✅ COMPLETED
      Located in `cobra-parser.go`; replaces the current flag booleans on `CobraParser`.
      ```go
      type CobraParserConfig struct {
          MiddlewaresFunc CobraMiddlewaresFunc
          ShortHelpLayers   []string
          SkipCommandSettingsLayer bool
          EnableProfileSettingsLayer bool
          EnableCreateCommandSettingsLayer bool
      }
      ```

- [x] **1.2 `CobraOption` (new single option type)**  ✅ COMPLETED
      ```go
      type CobraOption func(cfg *commandBuildConfig)
      ```
      Internal helper struct:
      ```go
      type commandBuildConfig struct {
          DualMode         bool
          GlazeToggleFlag  string
          DefaultToGlaze   bool
          HiddenGlazeFlags []string
          ParserCfg        CobraParserConfig
      }
      ```
      *`commandBuildConfig` is **internal**; callers only see `CobraOption`.*

- [x] **1.3 Helper Options**  ✅ COMPLETED (all return `CobraOption`)
      * WithParserConfig(cfg CobraParserConfig)
      * WithDualMode(enabled bool)
      * WithGlazeToggleFlag(name string)
      * WithHiddenGlazeFlags(names ...string)
      * WithDefaultToGlaze()
      * Add convenience functions mirroring current parser-option helpers (populate `ParserCfg`).

## 2. Cobra Parser Refactor

- [x] **2.1 Change `NewCobraParserFromLayers`**  ✅ COMPLETED
      New signature: `func NewCobraParserFromLayers(l *layers.ParameterLayers, cfg *CobraParserConfig) (*CobraParser, error)`.
      If `cfg == nil` use defaults equivalent to current behaviour.

- [ ] **2.2 Remove `CobraParserOption` and its helper funcs.**  ❌ TODO (still exported)

- [x] **2.3 Wrapper for compatibility**  ✅ COMPLETED
      ```go
      func NewCobraParserFromLayersWithOptions(l *layers.ParameterLayers, opts ...CobraOption)(*CobraParser,error){
          cfg := &commandBuildConfig{}
          for _, o := range opts { o(cfg) }
          return NewCobraParserFromLayers(l, &cfg.ParserCfg)
      }
      ```

## 3. Builder Overhaul

- [x] **3.1 `BuildCobraCommandFromCommand`**  ✅ COMPLETED
      Signature becomes `func BuildCobraCommandFromCommand(cmds.Command, ...CobraOption) (*cobra.Command,error)`.

- [x] **3.2 Inside**  ✅ COMPLETED
      1. Initialize default `commandBuildConfig` (toggle flag=`with-glaze-output`).
      2. Apply all `CobraOption`s.
      3. Construct `CobraParserConfig` from `cfg.ParserCfg` and call `NewCobraParserFromLayers`.
      4. If `cfg.DualMode==false` run old single-mode logic; else run merged dual-mode logic (copied from `BuildCobraCommandDualMode`).

## 4. Thin Compatibility Wrappers

- [x] **4.1 `BuildCobraCommandDualMode`**  ✅ COMPLETED
      ```go
      // Deprecated: Use BuildCobraCommandFromCommand(c, WithDualMode(true)) instead.
      func BuildCobraCommandDualMode(c cmds.Command, opts ...DualModeOption)(*cobra.Command,error){
          cfg := translateDualMode(opts) // builds []CobraOption
          return BuildCobraCommandFromCommand(c, cfg...)
      }
      ```
      Documented as *Deprecated*.

- [~] **4.2 `DualModeOption`**, `WithGlazeToggleFlag` (old), etc.  ⚠️ PARTIAL (needs deprecation comments)
      Move to `deprecated.go`, each translated to the corresponding `CobraOption`.

- [~] **4.3 Remove `CobraParserOption` from all public signatures**; ⚠️ PARTIAL (some remain)
      where still referenced (e.g. `BuildCobraCommandAlias`) wrap into `CobraOption` equivalents.

## 5. Helper Builders & Registries

- [~] Update `BuildCobraCommandAlias`, `AddCommandsToRootCommand`, etc., to accept `...CobraOption` and forward them. ⚠️ PARTIAL (done but inconsistencies remain)

## 6. Codebase Migration

- [ ] **6.1 Search & replace** direct uses of `BuildCobraCommandDualMode` → `BuildCobraCommandFromCommand` + `WithDualMode(true)`. ❌ TODO
- [ ] **6.2 Drop `CobraParserOption` imports in internal code; external users rely on wrappers. ❌ TODO

## 7. Tests & Docs

- [ ] **7.1 Unit/Integration tests** covering dual-mode flag behaviour and parser config pathways. ❌ TODO
- [ ] **7.2 Update tutorials, READMEs, generated docs to show `CobraOption` usage. ❌ TODO
- [ ] **7.3 MIGRATION.md** capturing deprecations and upgrade steps. ❌ TODO

## 8. Cleanup & Release

- [ ] Remove dead code confirmed unused. ❌ TODO
- [ ] Run `go vet`, `staticcheck`, `golangci-lint`. ❌ TODO
- [ ] Tag new minor release, announcing API simplification. ❌ TODO

---

## ✅ Completion Summary (as of 2025-07-29)

### Completed (8/16 items)
✅ Core unified API implementation  
✅ Foundation types and helper options  
✅ Parser refactor with compatibility wrapper  
✅ Builder overhaul with dual-mode support  
✅ Main compatibility wrapper for `BuildCobraCommandDualMode`  

### Partially Completed (3/16 items)  
⚠️ Deprecated functions moved but missing proper deprecation comments  
⚠️ Most public signatures updated but some inconsistencies remain  
⚠️ Helper builders updated but with remaining technical debt  

### Not Started (5/16 items)
❌ Legacy type removal  
❌ Codebase migration (search & replace)  
❌ Comprehensive testing  
❌ Documentation and migration guides  
❌ Final cleanup and release preparation  

### Key Issues Identified
1. **Feature regression**: Debug features (YAML/schema output) stubbed out in unified API
2. **Technical debt**: Duplicated execution logic across builders
3. **API boundary**: Old types still exported without deprecation warnings
4. **Missing tests**: No comprehensive test coverage for new functionality

**Status**: Core refactoring complete and functional, but needs polish and feature restoration before production use.

**Next Steps**: See detailed review in `ttmp/2025-07-29/04-review-of-the-dual-command-cobra-parser-options-refactor.md`

---

### Rationale

* One exported option type (`CobraOption`) keeps the public surface minimal.  
* Internal `commandBuildConfig` means we still gather state cleanly without exposing complexity.  
* `CobraParserConfig` provides a clear, declarative way to customise parsing, making functional options for the parser obsolete.  
* Thin wrappers + `deprecated.go` avoid immediate breaking changes for downstream users. 