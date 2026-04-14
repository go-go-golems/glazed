---
Title: Declarative Config Plan Example
Slug: declarative-config-plan-example
Short: Use `config.Plan` and `sources.FromResolvedFiles(...)` to build layered config loading with provenance.
Topics:
- configuration
- examples
- tracing
Commands:
- help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---

This example demonstrates the Glazed config-plan API in a small runnable program.

It shows how to:

- declare a layered config plan
- resolve repo, cwd, and explicit config files in low → high precedence order
- print a human-readable plan report
- load the resolved files with `sources.FromResolvedFiles(...)`
- inspect the resulting parsed field history and config provenance metadata

## Run the example

```bash
cd cmd/examples/config-plan

go run . show
```

This resolves:

- `repo.yaml` from the repository root via `GitRootFile(...)`
- `local.yaml` from the current working directory via `WorkingDirFile(...)`

Then it prints:

1. the resolved plan report
2. the final decoded settings
3. the parsed field history including config metadata

## Add an explicit override

```bash
cd cmd/examples/config-plan

go run . show --explicit explicit.yaml
```

The explicit file is applied last, so it overrides both the repo-level and cwd-level files.

## What to inspect

The most important part of the output is the parsed field history. Config-derived writes will include metadata like:

- `config_file`
- `config_index`
- `config_layer`
- `config_source_name`
- `config_source_kind`

That makes it easy to explain exactly why a final value won.

## Source files

See:

- `cmd/examples/config-plan/main.go`
- `cmd/examples/config-plan/repo.yaml`
- `cmd/examples/config-plan/local.yaml`
- `cmd/examples/config-plan/explicit.yaml`
- `cmd/examples/config-plan/README.md`

## See Also

- [Declarative Config Plans](../../topics/27-declarative-config-plans.md)
- [Config Files and Overlays](../../topics/24-config-files.md)
