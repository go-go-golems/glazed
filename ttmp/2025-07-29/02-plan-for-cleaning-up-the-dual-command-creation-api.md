# Refined Plan: Polished Unification of Glazed Cobra-Command API

**Goal**: Consolidate all command-builder and parser configuration into a single public option type, **`CobraOption`**, and streamline the implementation. Retain only one minimal backward-compatibility wrapper, `BuildCobraCommandDualMode`.

---

## 1. Foundation Types

1.1 Define **`CobraParserConfig`** (in `cobra-parser.go`) to encapsulate parser customization:
```go
// Declarative parser settings
type CobraParserConfig struct {
    MiddlewaresFunc                CobraMiddlewaresFunc
    ShortHelpLayers               []string
    SkipCommandSettingsLayer      bool
    EnableProfileSettingsLayer    bool
    EnableCreateCommandSettingsLayer bool
}
```

1.2 Introduce **`CobraOption`** as the single builder option:
```go
// Public API: configures command and parser
type CobraOption func(cfg *commandBuildConfig)
```

1.3 Internal **`commandBuildConfig`** (unexported):
```go
// Internal aggregate of all builder settings
type commandBuildConfig struct {
    DualMode         bool
    GlazeToggleFlag  string
    DefaultToGlaze   bool
    HiddenGlazeFlags []string
    ParserCfg        CobraParserConfig
}
```

1.4 Provide **helper factories** (all return `CobraOption`):
- `WithParserConfig(cfg CobraParserConfig)`
- `WithDualMode(enabled bool)`
- `WithGlazeToggleFlag(name string)`
- `WithHiddenGlazeFlags(names ...string)`
- `WithDefaultToGlaze()`
- Delete the deprecated helper factories of the same name in cobra-parser.go
- *Plus any convenience functions mirroring legacy parser flags.*

---

## 2. Extract Common Run Flow

2.1 Create a **`runCobraCommand`** helper:
```go
func runCobraCommand(
    cmd *cobra.Command,
    runFunc func(context.Context, *values.Values) error,
    parser *CobraParser,
    cfg commandBuildConfig,
) {
    // 1. Parse layers
    layers, err := parser.Parse(cmd, cmd.Flags().Args())
    // 2. Handle debug flags (--print-yaml, --print-schema)
    // 3. Determine mode (dual vs single)
    // 4. Dispatch to GlazeCommand or Writer/Bare
    // ... common cleanup
}
```

2.2 Refactor **all** builder code paths to call this single helper for `cmd.Run`.

---

## 3. Cobra Parser Refactor

3.1 **Change** `NewCobraParserFromLayers` to:
```go
func NewCobraParserFromLayers(
    layers *schema.Schema,
    cfg *CobraParserConfig,
) (*CobraParser, error)
```
- If `cfg == nil`, apply current defaults.
- Drop all `CobraParserOption` and its helpers if still present.

---

## 4. Builder Overhaul

4.1 **Update** signature of primary builder:
```go
func BuildCobraCommandFromCommand(
    c cmds.Command,
    opts ...CobraOption,
) (*cobra.Command, error)
```

4.2 **Implement** inside:
1. Instantiate default `commandBuildConfig` (`GlazeToggleFlag="with-glaze-output"`).
2. Apply all `opts` to populate `cfg`.
3. Call `NewCobraParserFromLayers(description.Layers, &cfg.ParserCfg)`.
4. Create `cobra.Command` via `NewCobraCommandFromCommandDescription` (ensure `Annotations` map initialized).
5. Add glaze toggle flag and hide flags as per `cfg`.
6. Call `parser.AddToCobraCommand(cmd)`.
7. **Call** shared `runCobraCommand(cmd, c, parser, cfg)` to set `cmd.Run`.
8. Return `cmd`.

4.3 **Restore missing features** within `runCobraCommand`:
- Debug flags: `--print-yaml`, `--print-schema`, `--print-parsed-parameters`.
- `cliopatra` and `CreateAlias` outputs.
- `CreateCommand` YAML generation.

4.4 Ensure **prioritized dispatch**: `GlazeCommand` → `WriterCommand` → `BareCommand`.

---

## 5. Minimal Compatibility Wrapper

5.1 Provide one thin, deprecated wrapper for dual-mode:
```go
// Deprecated: use BuildCobraCommandFromCommand(c, WithDualMode(true)).
func BuildCobraCommandDualMode(
    c cmds.Command,
    _ ...interface{}, // ignore old options
) (*cobra.Command, error) {
    return BuildCobraCommandFromCommand(c, WithDualMode(true))
}
```

5.2 **Remove** all other legacy aliases and functional-option wrappers.

---

## 6. Call-site Migration

6.1 **Replace** `BuildCobraCommandDualMode(...)` calls with:
```go
BuildCobraCommandFromCommand(myCmd, WithDualMode(true))
```

6.2 **Remove** imports of old helper types (`CobraParserOption`, `DualModeOption`).

---

## 7. Tests & Documentation



7.3 **Update** tutorials and documentation by grepping for the old API and replacing it with the new one.

---

## 8. Cleanup & Release

8.1 **Delete** dead code and legacy stubs after migration.

---

### Rationale

- **Single option type** simplifies the public API.  
- **Centralized run helper** avoids duplication and ensures parity across modes.  
- **Restored functionality** prevents regressions for end users.  
- **Minimal compatibility surface** reduces maintenance overhead. 