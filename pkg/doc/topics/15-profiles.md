---
Title: Profiles (profiles.yaml)
Slug: profiles
Short: |
  Use profiles.yaml to apply named configuration bundles across parameter layers, with predictable precedence and debugging.
Topics:
- configuration
- middleware
- profiles
Commands:
Flags:
- profile
- profile-file
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Profiles (profiles.yaml)

Profiles are a **named bundle of parameter overrides** stored in a YAML file (typically `profiles.yaml`).
They are applied via Glazed middleware and are designed to be **mid-precedence defaults**:

- Profiles override **defaults**
- But can be overridden by **config files**, **environment variables**, and **CLI flags**

This makes profiles ideal for “environment presets” (dev/staging/prod), model presets, or organization defaults.

## File format

The profile file is a YAML map:

- **Top level**: profile name
- **Second level**: layer slug
- **Third level**: parameter name/value pairs for that layer

Example:

```yaml
development:
  api:
    base-url: http://localhost:8080
  auth:
    token: XXX

production:
  api:
    base-url: https://api.example.com
```

## Default file location

Glazed’s profile middleware (`middlewares.GatherFlagsFromProfiles`) takes **two paths** and a default profile name:

- `defaultProfileFile`: the “well-known default” (commonly `os.UserConfigDir() + "/<app>/profiles.yaml"`)
- `profileFile`: the actual file to load (can be overridden)
- `defaultProfileName`: the profile name treated as optional when the default file is missing (defaults to `default`)

Recommended convention for apps:

- `~/.config/<app>/profiles.yaml` (via `os.UserConfigDir()` / `$XDG_CONFIG_HOME`)

## Selecting a profile

In a Cobra CLI built with Glazed, profile selection typically comes from the **ProfileSettings layer**:

- `--profile <name>`
- `--profile-file <path>`

and the matching environment variables (for app `PINOCCHIO`):

- `PINOCCHIO_PROFILE=<name>`
- `PINOCCHIO_PROFILE_FILE=<path>`

### Enable the flags (important)

These flags only exist if you enable the ProfileSettings layer when building your Cobra command:

```go
command, err := cli.BuildCobraCommand(myCmd,
  cli.WithProfileSettingsLayer(),
)
```

Without that layer:

- Cobra won’t accept `--profile/--profile-file`
- Glazed won’t parse `*_PROFILE/_PROFILE_FILE` into `profile-settings`
- Config files won’t be able to set `profile-settings` either

## Setting profile selection in config.yaml

Glazed config files are parsed in **layer-slug form**:

```yaml
<layer-slug>:
  <param-name>: <value>
```

So profile selection in a config file must be expressed under the `profile-settings` layer:

```yaml
profile-settings:
  profile: mistral
  profile-file: /etc/pinocchio/profiles.yaml
```

## Precedence (what overrides what)

When profiles are integrated into a full CLI pipeline, the effective precedence should be:

1. **CLI flags** (highest)
2. **Environment variables**
3. **Config files** (low → high overlays)
4. **Profiles**
5. **Defaults** (lowest)

This means:

- profiles are a convenient baseline
- but “local overrides” win (config/env/flags)

## Error behavior

Recommended (and implemented in Glazed’s `GatherFlagsFromProfiles`):

- If `profileFile` does not exist **and it is not the defaultProfileFile** → **error**
- If `profileFile` is the defaultProfileFile:
  - profile == `defaultProfileName` → skip silently
  - profile != `defaultProfileName` → **error** (so `APP_PROFILE=foobar` fails fast)
- If the file exists but the named profile is missing:
  - profile != `defaultProfileName` → **error**
  - profile == `defaultProfileName` → skip silently

## Avoiding “profile selection circularity” (important)

If you want profile selection (`profile-settings.profile` / `profile-settings.profile-file`) to be set via:

- env vars, and/or
- config files,

you cannot instantiate the profile middleware with “early” (default) values and expect it to pick up later overrides.
The common fix is a **bootstrap parse**:

1. Parse `profile-settings` (and usually `command-settings`) from defaults + config + env + flags.
2. Instantiate `GatherFlagsFromProfiles(...)` using the resolved values.
3. Run the main middleware chain for all layers, with the profiles middleware inserted between defaults and higher-precedence sources.

If you need a concrete example, see `topics/12-profiles-use-code.md` and the Geppetto reference implementation.

## Debugging

Use `--print-parsed-parameters` to inspect parse provenance per parameter, including which values came from:

- `defaults`
- `profiles` (with metadata like `{ profile, profileFile }`)
- `config` (with metadata like `{ config_file, index }`)
- `env`
- `cobra` (flags)

