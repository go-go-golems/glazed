# Tasks

## TODO

- [x] **Option A bootstrap parse (Geppetto)**: Implement “bootstrap parse” to resolve `profile-settings` using defaults + config + env + flags *before* instantiating `GatherFlagsFromProfiles` in `geppetto/pkg/layers/layers.go`.
- [x] **Config file list consistency**: Ensure the bootstrap parse uses the same config file list + mapper as the main middleware chain (avoid “profile resolved from config A but main chain loads config B”).
- [x] **Profile settings layer enabled in example**: Update `geppetto/cmd/examples/simple-inference/main.go` to enable the Glazed `ProfileSettings` layer so `--profile/--profile-file` flags exist and env parsing has the layer available.
- [x] **Failure semantics**: Make `PINOCCHIO_PROFILE=foobar` fail reliably (including when the default profiles file is missing) rather than silently skipping.
- [x] **Example profiles**: Extend `geppetto/misc/profiles.yaml` to include profiles:
  - [x] `gemini-2.5-pro`
  - [x] `gemini-2.5-flash`
  - [x] `sonnet-4-5`
- [x] **Smoke test script**: Add a script under `scripts/` that runs the example in `--debug` mode and validates:
  - [x] three known profiles succeed
  - [x] `PINOCCHIO_PROFILE=foobar` fails
  - [x] supports overriding profile file via `PINOCCHIO_PROFILE_FILE` (and documents the default `~/.config/pinocchio/profiles.yaml`)
- [x] **Smoke test**: Run `simple-inference` with:
  - [x] `PINOCCHIO_PROFILE=gemini-2.5-pro`
  - [x] `PINOCCHIO_PROFILE=gemini-2.5-flash`
  - [x] `PINOCCHIO_PROFILE=sonnet-4-5`
  - [x] `PINOCCHIO_PROFILE=foobar` (expect non-zero / error)

