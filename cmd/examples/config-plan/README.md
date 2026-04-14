# Config Plan Example

This example demonstrates Glazed's declarative config-plan API and provenance-aware config loading.

It shows how to:

- define a layered config plan with named sources
- resolve config files in low → high precedence order
- load resolved files through `sources.FromResolvedFiles(...)`
- inspect parse history with config metadata such as:
  - `config_file`
  - `config_index`
  - `config_layer`
  - `config_source_name`
  - `config_source_kind`

## What the example does

The example defines a config plan with three layers:

1. `repo` — a file discovered from the git repository root
2. `cwd` — a file discovered in the current working directory
3. `explicit` — an optional file passed with `--explicit`

The repo source uses:

- `GitRootFile("cmd/examples/config-plan/repo.yaml")`

The cwd source uses:

- `WorkingDirFile("local.yaml")`

So you should run the example from this directory.

## Run

```bash
cd cmd/examples/config-plan

go run . show
```

Expected behavior:

- `repo.yaml` is discovered from the repository root
- `local.yaml` is discovered from the current directory
- `local.yaml` overrides overlapping repo values

## Run with an explicit override

```bash
cd cmd/examples/config-plan

go run . show --explicit explicit.yaml
```

Expected behavior:

- `explicit.yaml` is applied last
- its values override both `repo.yaml` and `local.yaml`

## What to look for

The example prints three sections:

1. **Resolved config plan**
   - shows which sources were found and in what order
2. **Final settings**
   - shows the final merged values
3. **Parsed fields with provenance**
   - shows the field history and config metadata attached to each write

## Why this example matters

This is the reusable pattern intended for Glazed-based applications that want:

- explicit config-layer ordering
- project-local config conventions
- traceable precedence in parsed field history
- clean handoff into higher-level bootstrap systems such as Geppetto
