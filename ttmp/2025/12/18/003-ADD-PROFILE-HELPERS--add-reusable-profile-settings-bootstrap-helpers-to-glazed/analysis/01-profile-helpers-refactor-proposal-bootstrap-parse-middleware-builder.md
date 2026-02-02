---
Title: 'Profile helpers: refactor proposal (bootstrap parse + middleware builder)'
Ticket: 003-ADD-PROFILE-HELPERS
Status: active
Topics:
    - glazed
    - profiles
    - cobra
    - middleware
    - refactor
    - docs
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T15:17:08.460423379-05:00
---

# Profile helpers: refactor proposal (bootstrap parse + middleware builder)

## Problem statement

`geppetto/pkg/layers/layers.go` contains a correct but **verbose** “Option A bootstrap parse” to avoid
profile-selection circularity:

- resolve `command-settings` (for config-file selection)
- compute config file list (low → high)
- resolve `profile-settings` from defaults + config + env + flags
- instantiate `GatherFlagsFromProfiles(defaultFile, resolvedProfileFile, resolvedProfileName, ...)`

This pattern is broadly useful across Glazed-based CLIs. Keeping it copy/pasted in each app leads to:

- duplicated config-file resolution logic
- duplicated “which layers should env/config apply to” decisions
- inconsistent parse-step metadata (harder to debug)
- higher chance of subtle precedence regressions

## What already exists in Glazed (building blocks)

### 1) Config discovery and file list hooks

`cli.CobraParserConfig` supports:

- `AppName` → env prefix and default config discovery
- `ConfigFilesFunc` → custom “list of config files (low → high)”

The default discovery uses `config.ResolveAppConfigPath(appName, explicit)`:

- `$XDG_CONFIG_HOME/<app>/config.yaml` (via `os.UserConfigDir()`)
- `~/.<app>/config.yaml`
- `/etc/<app>/config.yaml`

### 2) Middleware primitives

- `sources.FromFiles(files, ...)`
- `sources.FromEnv(prefix, ...)`
- `sources.FromCobra(cmd, ...)`
- `middlewares.GatherFlagsFromProfiles(defaultProfileFile, profileFile, profileName, ...)`
- `middlewares.LoadParametersFromResolvedFilesForCobra(cmd, args, resolver, ...)` (Cobra-specific wrapper)

### 3) Profile path helpers (elsewhere)

In Clay there is `profiles.GetProfilesPathForApp(appName)` which implements the common default:

- `os.UserConfigDir() + "/<app>/profiles.yaml"`

Glazed profile middleware also has `GatherFlagsFromCustomProfiles(... WithProfileAppName(appName) ...)`
which resolves `~/.config/<appName>/profiles.yaml`.

## What’s missing (gap)

There is no single helper that resolves:

- the **selected** `cli.ProfileSettings` (`profile` + `profile-file`)
- using the **real precedence** sources (defaults/config/env/flags)
- and returns a reusable artifact:
  - either a `ProfileSettings` struct, or
  - a `middlewares.Middleware` (or middleware slice) ready to append

This is why apps like Geppetto currently implement a mini bootstrap chain manually.

## Proposed helper design (API)

### Goals

- **Reusable** across apps (not Pinocchio-specific)
- **Composable** with `CobraParserConfig.ConfigFilesFunc` and config mapping
- **No context refactor**: keep current middleware signature; accept `*cobra.Command` directly
- **Observable**: include parse-step metadata on profile load step (`{profile, profileFile}`)

### Option 1: CLI-level helper returning resolved selection

Add to `glazed/pkg/cli`:

```go
type ResolveProfileSelectionOptions struct {
    AppName string // required to compute env prefix + default profile path

    // Optional: config file resolver (low -> high). If nil, use ResolveAppConfigPath + --config-file.
    ConfigFilesFunc func(parsedCommandLayers *values.Values, cmd *cobra.Command, args []string) ([]string, error)

    // Optional: config mapper used by LoadParametersFromFiles for non-standard config structures
    ConfigMapper middlewares.ConfigFileMapper
}

func ResolveProfileSelection(
    cmd *cobra.Command,
    args []string,
    opts ResolveProfileSelectionOptions,
) (profile cli.ProfileSettings, configFiles []string, defaultProfileFile string, err error)
```

Then apps build their chain as:

- compute `profileSettings := ResolveProfileSelection(...)`
- append `GatherFlagsFromProfiles(defaultProfileFile, profileSettings.ProfileFile, profileSettings.Profile, ...)`

Pros:

- Simple mental model
- Easy to test

Cons:

- Caller still has to wire middleware ordering and metadata consistently

### Option 2: middleware factory helper (most ergonomic)

Add to `glazed/pkg/cmds/middlewares`:

```go
type ProfileBootstrapOptions struct {
    AppName string

    // Optional: config file resolver. If nil, use ResolveAppConfigPath + --config-file.
    ConfigFilesFunc func(parsed *values.Values, cmd *cobra.Command, args []string) ([]string, error)

    // Optional: config mapper for non-standard config structure
    ConfigMapper ConfigFileMapper
}

func ProfileMiddlewareFromBootstrap(
    cmd *cobra.Command,
    args []string,
    opts ProfileBootstrapOptions,
    parseOpts ...parameters.ParseStepOption,
) (mw Middleware, configFiles []string, err error)
```

This helper would:

1. Bootstrap-parse `command-settings` (to discover config-file overrides if needed)
2. Compute config file list (low → high)
3. Bootstrap-parse `profile-settings` from defaults + config + env + cobra
4. Return:
   - `mw`: `GatherFlagsFromProfiles(defaultProfileFile, resolvedProfileFile, resolvedProfileName, ...)`
   - `configFiles`: the list it used (so caller can reuse it in the main chain)

Pros:

- Hard to misuse
- Central place to enforce “circularity-safe” behavior
- Makes apps smaller and consistent

Cons:

- Needs careful documentation to avoid “hidden behavior”

### Option 3: extend `CobraParser.Parse()` to support “pre-middlewares”

This would add a formal phase-0 hook. It’s invasive; prefer Option 1/2 first.

## What Geppetto would look like after refactor (sketch)

Today: `GetCobraCommandGeppettoMiddlewares` contains ~100+ lines of bootstrap logic.

After:

```go
profileMw, configFiles, err := middlewares.ProfileMiddlewareFromBootstrap(cmd, args, middlewares.ProfileBootstrapOptions{
    AppName:      "pinocchio",
    ConfigMapper: configMapper, // keep repositories filtered
})
if err != nil { return nil, err }

middlewares_ := []middlewares.Middleware{
    sources.FromCobra(cmd),
    sources.FromArgs(args),
    sources.FromEnv("PINOCCHIO"),
    sources.FromFiles(configFiles, ...),
    profileMw,
    sources.FromDefaults(),
}
```

## Docs impact / update plan

If we add helpers, update:

- `glazed/pkg/doc/topics/15-profiles.md`: “recommended implementation” section showing helper usage
- `glazed/pkg/doc/topics/12-profiles-use-code.md`: update example to use helper instead of ad-hoc bootstrap
- Geppetto docs: mention that profile selection is circularity-safe by resolving `profile-settings` first

## Recommendation

Implement **Option 2** (middleware factory) in Glazed because it prevents incorrect usage and makes apps smaller.
Option 1 can exist as a lower-level helper if advanced customization needs it.

