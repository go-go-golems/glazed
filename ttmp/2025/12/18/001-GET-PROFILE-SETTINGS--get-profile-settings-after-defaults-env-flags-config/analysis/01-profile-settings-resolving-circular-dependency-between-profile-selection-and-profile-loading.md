---
Title: 'Profile settings: resolving circular dependency between profile selection and profile loading'
Ticket: 001-GET-PROFILE-SETTINGS
Status: active
Topics:
    - glazed
    - cobra
    - profiles
    - config
DocType: analysis
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2025-12-18T12:54:53.298143865-05:00
---

# Profile settings: resolving circular dependency between profile selection and profile loading

## Executive summary

Glazed can **load profiles** via `middlewares.GatherFlagsFromProfiles(...)`, but *selecting* which profile to load
(`profile-settings.profile` + `profile-settings.profile-file`) can itself be influenced by **defaults/env/config/flags**.

The current “bootstrap” mechanism (`cli.ParseCommandSettingsLayer`) only parses these settings from **Cobra flags**,
so profile selection from env/config is impossible during middleware construction. Any code that instantiates the profile
middleware with values read “too early” (like `geppetto/pkg/layers/layers.go`) will end up capturing defaults and loading
the wrong profile.

**Key takeaway:** middlewares don’t receive context; `*cobra.Command` is captured in closures. To resolve the circularity,
we need a **bootstrap parse** (mini middleware chain) that resolves `profile-settings` (and usually also `command-settings`)
*before* the profile-loading middleware is created or executed.

## Glossary / structures involved

- **`ParameterLayers`**: layer definitions (what flags exist, types, defaults).
- **`ParsedLayers`**: parsed values and their provenance (parse-step logs).
- **`ProfileSettings`**: a small struct with `Profile` and `ProfileFile` defined in `glazed/pkg/cli/cli.go`.
- **Profile file format**: YAML mapping `profileName -> layerSlug -> paramName -> value` (see `middlewares/profiles.go`).

## How Cobra flag registration works in Glazed

### Where flags get registered

The registration flow is:

1. `CobraParser.AddToCobraCommand(cmd)` iterates every layer and requires it to implement `layers.CobraParameterLayer`.
2. Each `ParameterLayerImpl.AddLayerToCobraCommand(cmd)` calls:
   - `ParameterDefinitions.AddParametersToCobraCommand(cmd, layerPrefix)`
3. `AddParametersToCobraCommand`:
   - Registers **flags** using `cmd.Flags()` (not persistent flags)
   - Registers **argument constraints** by setting `cmd.Args` and rewriting `cmd.Use` for argument-style parameters

Important implications:

- **Unknown flags are rejected by Cobra** before Glazed parsing ever runs.
- **Profile flags (`--profile`, `--profile-file`) only work if the ProfileSettingsLayer is actually registered** on the command
  (i.e., the layer exists in the `ParameterLayers` used for `AddToCobraCommand`).

### Defaults and help text

Flags are registered with typed default values (string/int/bool slices, …). But Glazed largely avoids relying on Cobra defaults:
Glazed’s parsing layer uses `Flags().Changed(name)` to decide if a value should be considered “provided”.

This is intentional because Glazed has its own precedence model (defaults/config/env/flags), and Cobra defaults would confuse it.

## How Cobra flag parsing works in Glazed

### Two distinct phases: Cobra parses first, Glazed parses second

Glazed does **not** drive Cobra’s parsing. Cobra parses CLI input as part of normal command execution, populating:

- `cmd.Flags()` values
- `cmd.Flags().Changed(flagName)` indicators

Only afterwards does Glazed parse those already-populated flags into `ParsedLayers` via middleware:

- `middlewares.ParseFromCobraCommand(cmd)` iterates layers and calls `ParseLayerFromCobraCommand(cmd, ...)`
- `ParameterLayerImpl.ParseLayerFromCobraCommand(...)` calls
  `ParameterDefinitions.GatherFlagsFromCobraCommand(cmd, onlyProvided=true, ...)`
- `GatherFlagsFromCobraCommand` uses `cmd.Flags().Changed(flagName)` to skip non-provided flags

### “Is cobra.Command carried through ctx?”

No. The middleware signature is:

```go
type HandlerFunc func(layers *layers.ParameterLayers, parsedLayers *layers.ParsedLayers) error
```

So there’s **no context** and no `*cobra.Command` passed through the middleware pipeline. Cobra is accessed via closure capture:
`ParseFromCobraCommand(cmd)` stores `cmd` in the returned middleware closure.

If you need to access flags “early”, you do it the same way: **read from `cmd.Flags()` directly** (or capture `cmd`).

## Middleware ordering / precedence (why it’s easy to get confused)

`ExecuteMiddlewares(layers, parsedLayers, m1, m2, m3)` calls the *outermost* middleware first, but “effective precedence”
depends on whether a middleware calls `next(...)` **first** (common for “override parsed values”) or **last**
(for “modify layers before parsing”).

Most of the value-setting middlewares (`SetFromDefaults`, `LoadParametersFromFiles`, `UpdateFromEnv`, `ParseFromCobraCommand`)
call `next(...)` first, so the *actual* value application order becomes:

```
defaults -> ... -> flags
```

even if the slice is written in the opposite order (flags first).

## Why profile loading becomes circular

### What we want (semantically)

We want:

1. **Resolve profile selection** (`profile`, `profile-file`) with precedence:
   - flags > env > config > defaults
2. **Load the selected profile’s values** as a mid-precedence source:
   - overrides defaults
   - overridden by config/env/flags

### What tends to happen today

In Geppetto’s middleware builder (`geppetto/pkg/layers/layers.go`), profile selection is read from `parsedCommandLayers`
*before* the main middleware chain executes.

But `parsedCommandLayers` is produced by `cli.ParseCommandSettingsLayer(cmd)` which currently parses from **Cobra only**.

Therefore:

- env/config cannot influence `profile-settings`
- profile middleware gets instantiated with default/empty selection values
- later env/flags may update `profile-settings` in the *real* `parsedLayers`,
  but the already-instantiated profile middleware continues to load the wrong profile

This is the same bug class documented in `pinocchio/ttmp/.../IMPROVE-PROFILES-001` (previous analysis).

## How to properly load profiles as part of the middleware stack

There are two viable “Glazed-shaped” solutions. Both rely on the same idea:
**run a mini parse of selection layers** (profile + command settings) using the real precedence sources,
then load profile values using the resolved selection, then run the main parse.

### Option A (recommended): explicit bootstrap parse before building the main chain

**Idea:** Before you construct the real middleware slice (which includes `GatherFlagsFromProfiles`),
execute a mini-chain against a restricted layer set containing:

- `command-settings` (for config file selection, if needed)
- `profile-settings`

Then instantiate `GatherFlagsFromProfiles` using those resolved values.

Pseudocode:

```go
bootstrapLayers := layers.NewParameterLayers(layers.WithLayers(
  NewCommandSettingsLayer(),
  NewProfileSettingsLayer(),
))
bootstrapParsed := layers.NewParsedLayers()

ExecuteMiddlewares(bootstrapLayers, bootstrapParsed,
  ParseFromCobraCommand(cmd),
  GatherArguments(args),
  UpdateFromEnv(APP_PREFIX),
  LoadParametersFromResolvedFilesForCobra(cmd, args, resolver),
  SetFromDefaults(),
)

// read resolved selection
cs := &cli.CommandSettings{}; bootstrapParsed.InitializeStruct(cli.CommandSettingsSlug, cs)
ps := &cli.ProfileSettings{}; bootstrapParsed.InitializeStruct(cli.ProfileSettingsSlug, ps)

// now build main middleware list, inserting profile middleware between defaults and config/env/flags
mainMiddlewares := []Middleware{
  ParseFromCobraCommand(cmd),
  GatherArguments(args),
  UpdateFromEnv(APP_PREFIX),
  LoadParametersFromResolvedFilesForCobra(cmd, args, resolver),
  GatherFlagsFromProfiles(defaultProfileFile, ps.ProfileFile, ps.Profile),
  SetFromDefaults(),
}
ExecuteMiddlewares(allLayers, allParsed, mainMiddlewares...)
```

Why this works:

- profile selection is resolved with “real” precedence sources
- profile values are applied as a mid-precedence update
- higher precedence sources (config/env/flags) still override profile values

How to keep it local to some layers:

- `WrapWithWhitelistedLayers([]string{cli.ProfileSettingsSlug, cli.CommandSettingsSlug}, ...)`
  can be used if you want to run the same middleware list but restricted to only selection layers.

### Option B: a single “resolve + load profiles” middleware that runs a nested bootstrap chain internally

**Idea:** Keep a single middleware in the main chain (placed just above defaults), but inside it:

1. Execute a nested “bootstrap” `ExecuteMiddlewares(...)` against a restricted layer set
2. Read `ProfileSettings` from that nested result
3. Load the profile YAML and merge the selected profile map into the *outer* `parsedLayers`

This satisfies the desire to “load profiles as part of the middleware stack itself” while still solving the circularity.

Pros:

- Single “unit” to reuse across apps
- Can be inserted into existing stacks without restructuring `CobraParser.Parse()`

Cons:

- More complex to reason about (nested middleware chains)
- Needs careful metadata/logging (so parse provenance remains understandable)

### Option C: change middleware signatures to carry context/cmd

Not recommended as the first move: it’s invasive and would cascade across middlewares. The current design prefers
closure capture (`ParseFromCobraCommand(cmd)`), and solving the circularity doesn’t *require* signature changes.

## Recommended path forward

**Recommendation:** Implement Option A (explicit bootstrap parse) first because it is simpler, testable, and
matches existing prior design work in Pinocchio.

If we want a reusable building block for multiple apps (Pinocchio, Geppetto, etc.), wrap Option A into a helper:

- `ResolveBootstrapSettings(cmd, args, appName, configResolver) (CommandSettings, ProfileSettings, error)`

and then have application middleware builders use it to instantiate:

- `GatherFlagsFromProfiles(defaultProfileFile, resolved.ProfileFile, resolved.Profile, ...)`

## Open questions / risks

- **Required flags**: `GatherFlagsFromCobraCommand` enforces `pd.Required` at the Cobra layer. In a world where required
  values may come from config/env/profile, this likely needs adjustment (there are TODOs in the code).
- **Config file selection**: if config file paths themselves can be influenced by env/flags, the bootstrap chain should
  include `command-settings` and apply env/flags when resolving config file lists.
- **Tests**: we need tests for precedence across defaults/profile/config/env/flags and for “profile selection sourced from config”.

## References

- Glazed Cobra integration:
  - `glazed/pkg/cli/cobra-parser.go`
  - `glazed/pkg/cmds/parameters/cobra.go`
  - `glazed/pkg/cmds/middlewares/cobra.go`
- Profile middleware:
  - `glazed/pkg/cmds/middlewares/profiles.go`
- Prior broader analysis (Pinocchio/Geppetto integration):
  - `pinocchio/ttmp/2025/12/15/IMPROVE-PROFILES-001--fix-profile-system-interdependency-issues/analysis/01-profile-system-interdependency-health-inspection.md`
