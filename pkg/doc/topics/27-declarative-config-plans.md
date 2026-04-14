---
Title: Declarative Config Plans
Slug: declarative-config-plans
Short: Define layered config discovery as explicit source plans and preserve config provenance in parsed field history, either by resolving files yourself or by loading plans directly through source middlewares.
Topics:
- configuration
- middlewares
- tracing
- overlays
- api-design
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Declarative Config Plans

Glazed now supports a declarative config-plan API for applications that want to express config discovery as explicit data instead of hidden helper logic. A config plan lets you define which config sources exist, which semantic layer each source belongs to, and in what order those layers should be applied. The result is easier to read in code, easier to debug, and easier to reuse across multiple applications.

This page explains the generic Glazed API, not any one application’s policy. It is the right reference when you want to build your own layered config discovery model and feed the result into Glazed’s field/value system with parse-step provenance intact.

## Why use a config plan

Traditional config discovery helpers are convenient for simple cases, but they hide important policy decisions in control flow. A reviewer often has to reverse-engineer the answers to questions like:

- which files are considered?
- which source wins when multiple files exist?
- is a repo-local file allowed to override user config?
- how do I tell which layer wrote the final value?

A config plan makes those decisions explicit.

Instead of “ask a helper for a path,” the application builds a plan such as:

```go
plan := config.NewPlan(
    config.WithLayerOrder(
        config.LayerSystem,
        config.LayerUser,
        config.LayerRepo,
        config.LayerCWD,
        config.LayerExplicit,
    ),
    config.WithDedupePaths(),
).Add(
    config.SystemAppConfig("myapp").Named("system-app-config"),
    config.XDGAppConfig("myapp").Named("xdg-app-config"),
    config.HomeAppConfig("myapp").Named("home-app-config"),
    config.GitRootFile(".myapp.local.yaml").Named("git-root-local"),
    config.WorkingDirFile(".myapp.local.yaml").Named("cwd-local"),
    config.ExplicitFile(explicitPath).Named("explicit-config-file"),
)
```

That one block tells the reader almost everything they need to know.

## Core types

The declarative config-plan API lives in `glazed/pkg/config`.

### `ConfigLayer`

A config layer is a semantic precedence bucket.

Built-in layers currently include:

- `LayerSystem`
- `LayerUser`
- `LayerRepo`
- `LayerCWD`
- `LayerExplicit`

These are more readable than raw priority numbers because they communicate intent directly.

### `SourceSpec`

A source spec describes one config discovery rule.

It includes:

- source name
- layer
- source kind
- discovery function
- optional condition
- optional stop-if-found behavior

You usually construct source specs through helper functions and then customize them fluently:

```go
config.GitRootFile(".myapp.local.yaml").
    Named("git-root-local").
    InLayer(config.LayerRepo).
    Kind("project-config")
```

### `ResolvedConfigFile`

A resolved config file is the output of the plan after discovery.

It carries:

- file path
- config layer
- source name
- source kind
- file index in the final low→high precedence list

This is the object you pass to `sources.FromResolvedFiles(...)` when you want provenance-aware loading and also want to inspect or reuse the explicit resolved file list.

### `PlanReport`

A plan report summarizes what happened during resolution:

- which sources were found
- which were skipped
- which were deduped
- what the final ordered file list is

This is useful for debugging, tests, and CLI explain output.

## Built-in source constructors

Glazed ships with a small set of generic source constructors that cover common application patterns.

### Conventional app config locations

- `SystemAppConfig(appName)`
- `HomeAppConfig(appName)`
- `XDGAppConfig(appName)`
- `ExplicitFile(path)`

These cover the common “system / user / explicit override” flow.

### Local/project-oriented sources

- `GitRootFile(name)`
- `WorkingDirFile(name)`

These are especially useful for tools that want repository-local or working-directory-local config conventions.

## Example: resolving and loading a plan

Once you have a plan, you can either resolve it explicitly and feed the result into Glazed sources, or let the middleware do that for you.

```go
files, report, err := plan.Resolve(context.Background())
if err != nil {
    return err
}

fmt.Println(report.String())

parsed := values.New()
err = sources.Execute(
    schema_,
    parsed,
    sources.FromResolvedFiles(files),
    sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
)
```

This gives you both:

- deterministic config ordering
- rich parse-step provenance in `parsed`
- access to the explicit `PlanReport` and `[]ResolvedConfigFile` values before loading

## Example: loading a plan directly through middleware

If you do not need to inspect the resolved file list first, use the higher-level source middlewares:

```go
parsed := values.New()
err := sources.Execute(
    schema_,
    parsed,
    sources.FromConfigPlan(plan),
    sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
)
```

For dynamic plans that depend on already-parsed lower-precedence values, use `FromConfigPlanBuilder(...)`:

```go
err := sources.Execute(
    schema_,
    parsed,
    sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
    sources.FromConfigPlanBuilder(func(ctx context.Context, parsed *values.Values) (*config.Plan, error) {
        settings := &MySelectorSettings{}
        if err := parsed.DecodeSectionInto("selector", settings); err != nil {
            return nil, err
        }

        return config.NewPlan(
            config.WithLayerOrder(config.LayerExplicit),
        ).Add(
            config.ExplicitFile(settings.ConfigFile).Named("selected-config"),
        ), nil
    }),
)
```

Use `FromConfigPlan(...)` / `FromConfigPlanBuilder(...)` when you want the middleware pipeline to own resolution and loading. Use `FromResolvedFiles(...)` when you want to debug, print, test, or otherwise reuse the explicit resolved results.

## Provenance metadata in parsed field history

The main reason to use `FromResolvedFiles(...)` instead of only `FromFiles(...)` is provenance.

When Glazed loads resolved config files, it attaches metadata like:

- `config_file`
- `config_index`
- `config_layer`
- `config_source_name`
- `config_source_kind`

That means a parsed field history can show not only that a value came from a config file, but also which layer and which source rule produced it.

Example shape:

```yaml
demo:
  threshold:
    value: 22
    log:
      - source: config
        value: 11
        metadata:
          config_file: /repo/repo.yaml
          config_layer: repo
          config_source_name: repo-example-config
      - source: config
        value: 22
        metadata:
          config_file: /repo/cmd/examples/config-plan/local.yaml
          config_layer: cwd
          config_source_name: cwd-example-config
```

That is much more useful than only seeing a final value with no history.

## Relationship to `FromFiles(...)` and `FromResolvedFiles(...)`

`FromFiles(...)` still exists and is still appropriate when you already have a simple ordered list of file paths.

Use `FromFiles(...)` when:

- your app already has a simple list of config files
- you do not need layer/source metadata
- you want a minimal API surface

Use `FromResolvedFiles(...)` when:

- you are building layered config discovery
- you want traceable provenance in parse history
- you want the loader to preserve layer/source information directly
- you want to inspect or reuse the resolved file list before loading

Use `FromConfigPlan(...)` or `FromConfigPlanBuilder(...)` when:

- you already have a plan and want the middleware pipeline to resolve it for you
- you want dynamic plan selection based on already-parsed lower-precedence values
- you still want the same provenance metadata without manually calling `plan.Resolve(...)`

## Relationship to higher-level bootstraps

The plan API is intentionally generic. Glazed does not decide what files your application should load. It provides reusable primitives so higher-level systems can define policy cleanly.

A common pattern looks like this:

- Glazed owns config discovery primitives and provenance-aware loading
- a higher-level bootstrap package builds an application-specific plan
- the application chooses actual file names and precedence semantics

That lets multiple tools share the same underlying config machinery without sharing the same policy.

## Troubleshooting

| Problem | Cause | Solution |
|---|---|---|
| A later file did not override an earlier one | The file order in the resolved plan is wrong | Print `report.String()` and verify the layer order and final path list |
| A config value shows `source: config` but no layer metadata | The loader used `FromFiles(...)` instead of a provenance-aware plan loader | Resolve the plan to `[]ResolvedConfigFile` and load through `FromResolvedFiles(...)`, or load the plan directly with `FromConfigPlan(...)` / `FromConfigPlanBuilder(...)` |
| The same file appears twice | Two sources discovered the same path | Enable `WithDedupePaths()` and inspect the deduped paths in the plan report |
| A git-root source never finds anything | The current process is not inside a git repository or the target file path is wrong | Verify the repo root and the path passed to `GitRootFile(...)` |

## See Also

- [Config Files and Overlays](24-config-files.md)
- [Pattern-Based Config Mapping](23-pattern-based-config-mapping.md)
- [Declarative Config Plan Example](../examples/config/01-declarative-config-plan.md)
- `cmd/examples/config-plan`
