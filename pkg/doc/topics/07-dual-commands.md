---
Title: Dual Commands
Slug: dual-commands
Short: Build commands with both human-readable and structured output
Topics:
- commands
- dual-mode
- bare-command
- glaze-command
- structured-output
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Dual Commands

Dual commands implement both `BareCommand` and `GlazeCommand` interfaces, switching between human-readable and structured output based on the `--with-glaze-output` flag.

## When to Use

- **Interactive use** needs readable text output
- **Scripting/integration** needs machine-parseable data (JSON, YAML, CSV)
- Single command serves both use cases

## Minimal Example

```go
type StatusCommand struct {
    *cmds.CommandDescription
}

type StatusSettings struct {
    Verbose bool `glazed:"verbose"`
}

// BareCommand: human-readable output
func (c *StatusCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    fmt.Println("Status: Healthy")
    if s.Verbose {
        fmt.Println("Version: 1.0.0")
    }
    return nil
}

// GlazeCommand: structured output
func (c *StatusCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &StatusSettings{}
    if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
        return err
    }
    row := types.NewRow(
        types.MRP("status", "healthy"),
    )
    if s.Verbose {
        row.Set("version", "1.0.0")
    }
    return gp.AddRow(ctx, row)
}

var _ cmds.BareCommand = &StatusCommand{}
var _ cmds.GlazeCommand = &StatusCommand{}
```

## Builder Options

```go
cobraCmd, err := cli.BuildCobraCommand(cmd,
    cli.WithDualMode(true),                  // enable dual-mode toggle
    cli.WithGlazeToggleFlag("with-glaze-output"), // flag name
    cli.WithDefaultToGlaze(),                // default to glaze mode (optional)
    cli.WithHiddenGlazeFlags("template", "select"), // hide specific glaze flags
)
```

| Option | Description |
|--------|-------------|
| `WithDualMode(true)` | Enable toggle between BareCommand and GlazeCommand |
| `WithGlazeToggleFlag` | Set the flag name (default: `with-glaze-output`) |
| `WithDefaultToGlaze` | Start in glaze mode instead of classic mode |
| `WithHiddenGlazeFlags` | Hide specific glaze output flags |

## Default Output Format

Set the default output format for glaze mode using `WithOutputParameterLayerOptions`:

```go
glazedLayer, err := settings.NewGlazedParameterLayers(
    settings.WithOutputParameterLayerOptions(
        layers.WithDefaults(map[string]interface{}{
            "output": "json",  // default to JSON
        }),
    ),
)
```

## Common Patterns

### Field Consistency

When multiple commands output similar data, use consistent field names:

```go
// Common fields across cloud commands
types.MRP("id", node.Id()),
types.MRP("name", node.Name()),
types.MRP("type", node.Document.Type),
types.MRP("is_dir", node.IsDirectory()),
types.MRP("path", buildPathFromParents(node)),
```

### Error Handling in Callbacks

When iterating (e.g., WalkTree), `gp.AddRow()` returns `error`, not `bool`:

```go
// ❌ Wrong: returns error, not bool
Visit: func(node *model.Node, _ []string) bool {
    gp.AddRow(ctx, row)  // ignores error
    return err            // compile error
}

// ✅ Correct: handle or ignore error
Visit: func(node *model.Node, _ []string) bool {
    if err := gp.AddRow(ctx, row); err != nil {
        // log or handle
    }
    return false  // continue walking
}
```

### Settings Struct Pattern

Share the same settings struct between both interface implementations:

```go
type FindSettings struct {
    AuthSettings
    Compact bool `glazed.parameter:"compact"`
    Start   string `glazed.parameter:"start"`
    Pattern string `glazed.parameter:"pattern"`
}

// Used by both Run() and RunIntoGlazeProcessor()
func (c *FindCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
    s := &FindSettings{}
    parsedLayers.InitializeStruct(layers.DefaultSlug, s)
    // ...
}

func (c *FindCommand) RunIntoGlazeProcessor(ctx context.Context, parsedLayers *layers.ParsedLayers, gp middlewares.Processor) error {
    s := &FindSettings{}
    parsedLayers.InitializeStruct(layers.DefaultSlug, s)
    // ...
}
```

## Testing

```bash
# Classic mode (default)
./remarquee cloud ls

# Glaze mode - JSON (now default if configured)
./remarquee cloud ls --with-glaze-output

# Override format
./remarquee cloud ls --with-glaze-output --output yaml
./remarquee cloud ls --with-glaze-output --output csv --fields name,path
```

## Reference

- Tutorial: `glazed/pkg/doc/tutorials/05-build-first-command.md`
- Commands reference: `glazed/pkg/doc/topics/commands-reference.md`
- Example: `cmd/examples/new-api-dual-mode/main.go`