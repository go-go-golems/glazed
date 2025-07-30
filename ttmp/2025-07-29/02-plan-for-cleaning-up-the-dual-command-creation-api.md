# Streamlined Plan: Unifying Glazed Cobra-Command Creation

Goal: Have a **single public option type – `CobraOption`** – that configures *both* the command-builder and the underlying parser.  The legacy `BuildCobraCommandDualMode`, `DualModeOption`, and `CobraParserOption` APIs become thin, deprecated wrappers.

---

## 1. Foundation Types

- [ ] **1.1 `CobraParserConfig` (new)**  
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

- [ ] **1.2 `CobraOption` (new single option type)**  
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

- [ ] **1.3 Helper Options**  (all return `CobraOption`)
      * WithParserConfig(cfg CobraParserConfig)
      * WithDualMode(enabled bool)
      * WithGlazeToggleFlag(name string)
      * WithHiddenGlazeFlags(names ...string)
      * WithDefaultToGlaze()
      * Add convenience functions mirroring current parser-option helpers (populate `ParserCfg`).

## 2. Cobra Parser Refactor

- [ ] **2.1 Change `NewCobraParserFromLayers`**  
      New signature: `func NewCobraParserFromLayers(l *layers.ParameterLayers, cfg *CobraParserConfig) (*CobraParser, error)`.
      If `cfg == nil` use defaults equivalent to current behaviour.

- [ ] **2.2 Remove `CobraParserOption` and its helper funcs.**  

- [ ] **2.3 Wrapper for compatibility**  
      ```go
      func NewCobraParserFromLayersWithOptions(l *layers.ParameterLayers, opts ...CobraOption)(*CobraParser,error){
          cfg := &commandBuildConfig{}
          for _, o := range opts { o(cfg) }
          return NewCobraParserFromLayers(l, &cfg.ParserCfg)
      }
      ```

## 3. Builder Overhaul

- [ ] **3.1 `BuildCobraCommandFromCommand`**  
      Signature becomes `func BuildCobraCommandFromCommand(cmds.Command, ...CobraOption) (*cobra.Command,error)`.

- [ ] **3.2 Inside**
      1. Initialize default `commandBuildConfig` (toggle flag=`with-glaze-output`).
      2. Apply all `CobraOption`s.
      3. Construct `CobraParserConfig` from `cfg.ParserCfg` and call `NewCobraParserFromLayers`.
      4. If `cfg.DualMode==false` run old single-mode logic; else run merged dual-mode logic (copied from `BuildCobraCommandDualMode`).

## 4. Thin Compatibility Wrappers

- [ ] **4.1 `BuildCobraCommandDualMode`**  
      ```go
      // Deprecated: Use BuildCobraCommandFromCommand(c, WithDualMode(true)) instead.
      func BuildCobraCommandDualMode(c cmds.Command, opts ...DualModeOption)(*cobra.Command,error){
          cfg := translateDualMode(opts) // builds []CobraOption
          return BuildCobraCommandFromCommand(c, cfg...)
      }
      ```
      Documented as *Deprecated*.

- [ ] **4.2 `DualModeOption`**, `WithGlazeToggleFlag` (old), etc.  
      Move to `deprecated.go`, each translated to the corresponding `CobraOption`.

- [ ] **4.3 Remove `CobraParserOption` from all public signatures**; where still referenced (e.g. `BuildCobraCommandAlias`) wrap into `CobraOption` equivalents.

## 5. Helper Builders & Registries

- [ ] Update `BuildCobraCommandAlias`, `AddCommandsToRootCommand`, etc., to accept `...CobraOption` and forward them.

## 6. Codebase Migration

- [ ] **6.1 Search & replace** direct uses of `BuildCobraCommandDualMode` → `BuildCobraCommandFromCommand` + `BuildCobraCommand` + `WithDualMode(true)`.
- [ ] **6.2 Drop `CobraParserOption` imports in internal code; external users rely on wrappers.

## 7. Tests & Docs

- [ ] **7.1 Unit/Integration tests** covering dual-mode flag behaviour and parser config pathways.
- [ ] **7.2 Update tutorials, READMEs, generated docs to show `CobraOption` usage.
- [ ] **7.3 MIGRATION.md** capturing deprecations and upgrade steps.

## 8. Cleanup & Release

- [ ] Remove dead code confirmed unused.
- [ ] Run `go vet`, `staticcheck`, `golangci-lint`.
- [ ] Tag new minor release, announcing API simplification.

---

### Rationale

* One exported option type (`CobraOption`) keeps the public surface minimal.  
* Internal `commandBuildConfig` means we still gather state cleanly without exposing complexity.  
* `CobraParserConfig` provides a clear, declarative way to customise parsing, making functional options for the parser obsolete.  
* Thin wrappers + `deprecated.go` avoid immediate breaking changes for downstream users. 