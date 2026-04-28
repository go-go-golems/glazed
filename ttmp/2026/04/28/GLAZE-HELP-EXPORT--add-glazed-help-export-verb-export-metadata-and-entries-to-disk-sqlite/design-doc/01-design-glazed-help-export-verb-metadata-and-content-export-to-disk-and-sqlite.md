---
Title: 'Design: Glazed Help Export Verb - Metadata and Content Export to Disk and SQLite'
Ticket: GLAZE-HELP-EXPORT
Status: active
Topics:
    - glazed
    - help-system
    - cli
    - export
    - sqlite
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go
      Note: Application root
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/doc/topics/28-export-help-entries.md
      Note: Help topic documenting the export verb
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/cmd/cobra.go
      Note: Cobra help command wiring
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/cmd/export.go
      Note: Implementation of ExportCommand
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/help.go
      Note: HelpSystem facade
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/model/section.go
      Note: Section struct definition
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/site/render.go
      Note: Prior art for export logic
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/store/query.go
      Note: Predicate query system
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go
      Note: SQLite store implementation
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-28T07:35:00-04:00
WhatFor: ""
WhenToUse: ""
---










# Design: Glazed Help Export Verb - Metadata and Content Export to Disk and SQLite

## Executive Summary

This document describes the design for adding an **export verb** to the Glazed help system. The goal is to let any binary built on Glazed export its embedded help entries - both their metadata and their full content - to external files on disk. This enables downstream tooling: indexing, searching offline, generating documentation sites, auditing coverage, and integrating help content into other applications.

The feature adds a single `glaze help export` verb that exports help sections from any binary built on Glazed. By default, it streams section metadata (and full content, since `--with-content` defaults to `true`) through the Glazed processor as JSON, CSV, or a table. With `--format files` or `--format sqlite`, it writes reconstructed markdown files or a portable SQLite database to disk instead.

This document is written for a new intern joining the team. Every concept is explained from first principles, with concrete file references, pseudocode, API call patterns, and step-by-step implementation guidance.

---

## Problem Statement and Scope

### What is the pain point today?

The Glazed help system stores documentation as **help sections** inside the binary, loaded from embedded markdown files at startup. Users can browse these sections interactively via:

- `glaze help <topic>` - view a single topic
- `glaze help --list` - list available topics
- `glaze help --query "..."` - search with the DSL
- `glaze help --ui` - browse in a TUI
- `glaze serve-help` - browse via HTTP
- `glaze render-site` - export as a static website

What is **missing** is a simple, scriptable way to get the help data *out* of the binary in a structured form. If a developer wants to:

- Audit which topics document which commands
- Build a custom search index in another tool
- Archive the documentation at a specific version
- Transform the markdown for another publishing pipeline
- Merge help entries from multiple Glazed binaries

...they currently have no CLI-native way to do so. The `render-site` command exists, but it produces a full React SPA with JSON payloads - overkill for many use cases.

### Scope of this ticket

1. Add `glaze help export` - a single verb that exports help section data.
2. Default behavior: stream all matching sections (metadata + content, since `--with-content` is `true` by default) as JSON/CSV/table via the Glaze processor.
3. Support filtering (by type, topic, command, flag, slug) so exports are targeted.
4. Support disk-export modes (`--format files` for markdown files, `--format sqlite` for a SQLite database).
5. Support `--with-content=false` to omit the `content` field from tabular output, producing lightweight metadata-only exports.
6. Keep the implementation idiomatic to the existing Glazed command framework.

### Out of scope

- Modifying the help section schema (no new fields).
- Changing how sections are loaded or rendered for interactive use.
- Replacing `render-site` (it remains the right tool for full static-site generation).

---

## Current-State Architecture

To understand where the export verb fits, we need to understand the help system's data flow. This section explains each layer from the bottom up.

### Layer 1: The Markdown Source Files

Help content lives in `.md` files inside the repository, typically under a `doc/` directory. Each file contains YAML frontmatter followed by markdown body.

**Example file:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/doc/help-system.md`

```yaml
---
Title: "Help System"
Slug: "help-system"
Short: "Overview of the glazed help system"
Topics:
- help
- documentation
Commands:
- help
SectionType: GeneralTopic
---

The help system allows...
```

**Key frontmatter fields:**

| Field | Type | Purpose |
|-------|------|---------|
| `Title` | string | Human-readable title |
| `Slug` | string | Unique identifier used for lookups |
| `Short` | string | One-line summary |
| `SectionType` | string | `GeneralTopic`, `Example`, `Application`, `Tutorial` |
| `Topics` | []string | Tags for cross-referencing |
| `Commands` | []string | Which CLI commands this section documents |
| `Flags` | []string | Which flags this section documents |
| `IsTopLevel` | bool | Show on the root help page? |
| `ShowPerDefault` | bool | Show by default in listings? |
| `Order` | int | Sort order |

**File references:**
- Parser: `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/model/parse.go`
- Section struct: `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/model/section.go`

### Layer 2: The `model.Section` Struct

The parser converts each markdown file into a `model.Section` value in memory.

```go
// From pkg/help/model/section.go
type Section struct {
    ID          int64
    Slug        string
    SectionType SectionType  // GeneralTopic, Example, Application, Tutorial
    Title       string
    SubTitle    string
    Short       string
    Content     string       // Full markdown body (after frontmatter)
    Topics      []string
    Flags       []string
    Commands    []string
    IsTopLevel  bool
    IsTemplate  bool
    ShowPerDefault bool
    Order       int
    CreatedAt   string
    UpdatedAt   string
}
```

This struct is the **canonical data shape** for everything downstream. Whether we are rendering to the terminal, serving over HTTP, or exporting to disk, we start from `[]*model.Section`.

### Layer 3: The `store.Store` - SQLite Backend

Every `HelpSystem` owns a `store.Store`, which wraps a SQLite database. When sections are loaded, they are upserted into this store. The store provides:

- `Insert(ctx, section)` - create new row
- `Update(ctx, section)` - modify existing row by ID
- `Upsert(ctx, section)` - insert or update by slug (the primary loading path)
- `GetBySlug(ctx, slug)` - retrieve one section
- `GetByID(ctx, id)` - retrieve one section by numeric ID
- `List(ctx, orderBy)` - list all sections with optional ordering
- `Find(ctx, predicate)` - query with structured predicates
- `Count(ctx)` - total section count
- `Clear(ctx)` - delete all sections

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go`

The SQLite schema is simple:

```sql
CREATE TABLE sections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT NOT NULL UNIQUE,
    section_type INTEGER NOT NULL,
    title TEXT NOT NULL,
    sub_title TEXT,
    short TEXT,
    content TEXT,
    topics TEXT,       -- comma-separated
    flags TEXT,        -- comma-separated
    commands TEXT,     -- comma-separated
    is_top_level BOOLEAN DEFAULT FALSE,
    is_template BOOLEAN DEFAULT FALSE,
    show_per_default BOOLEAN DEFAULT FALSE,
    order_num INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

The store is created in-memory by default (`:memory:`), but the `store.New(dbPath)` constructor can also target a file path.

### Layer 4: The `help.HelpSystem` Facade

The `HelpSystem` is the high-level object that most of the application interacts with. It owns the `Store` and provides file-loading conveniences.

```go
// From pkg/help/help.go
type HelpSystem struct {
    Store *store.Store
}

func NewHelpSystem() *HelpSystem        // Creates in-memory store
func NewHelpSystemWithStore(st *store.Store) *HelpSystem

func (hs *HelpSystem) LoadSectionsFromFS(f fs.FS, dir string) error
func (hs *HelpSystem) AddSection(section *model.Section)
func (hs *HelpSystem) GetSectionWithSlug(slug string) (*model.Section, error)
func (hs *HelpSystem) GetTopLevelHelpPage() *HelpPage
```

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/help.go`

### Layer 5: The Cobra Help Command

The `help_cmd` package wires the `HelpSystem` into Cobra. The key function is:

```go
// From pkg/help/cmd/cobra.go
func SetupCobraRootCommand(hs *help.HelpSystem, cmd *cobra.Command)
```

This function:
1. Overrides the root command's `HelpFunc` and `UsageFunc` to use Glazed rendering.
2. Creates a `help` subcommand (via `NewCobraHelpCommand`) that supports `--list`, `--topics`, `--query`, `--ui`, etc.
3. Adds the `help` subcommand to the root command tree.

The current `help` subcommand has a `Run` function that branches based on flags:
- `--ui` → launch TUI
- `--query` → DSL search
- `--list`, `--topics`, `--examples`, etc. → filtered listing
- positional args → lookup by slug or command path

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/cmd/cobra.go`

### Layer 6: The Application Root (`main.go`)

Each binary (e.g., `glaze`) initializes the help system in its `main`:

```go
// From cmd/glaze/main.go
helpSystem := help.NewHelpSystem()
err = doc.AddDocToHelpSystem(helpSystem)  // Load embedded docs
help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)
```

The `doc` package embeds the markdown files using `//go:embed *` and calls `hs.LoadSectionsFromFS(docFS, ".")`.

**File reference:** `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go`

### Architecture Diagram (Current State)

```
┌─────────────────────────────────────────────────────────────┐
│                     Application (glaze)                      │
│  ┌─────────────┐    ┌─────────────────────────────────────┐ │
│  │ Cobra Root  │◄───│ help_cmd.SetupCobraRootCommand(...) │ │
│  │   Command   │    └─────────────────────────────────────┘ │
│  └──────┬──────┘                                           │
│         │                                                    │
│         ▼                                                    │
│  ┌────────────────────────────────────────┐                 │
│  │   "help" subcommand (cobra.Command)    │                 │
│  │  - --list, --topics, --examples, etc.  │                 │
│  │  - --query "DSL"                       │                 │
│  │  - --ui                                │                 │
│  └────────────────────────────────────────┘                 │
│                        │                                     │
│                        ▼                                     │
│              ┌──────────────────┐                           │
│              │  help.HelpSystem │                           │
│              │  - LoadSectionsFromFS()                      │
│              │  - GetSectionWithSlug()                      │
│              │  - ComputeRenderData()                       │
│              └────────┬─────────┘                           │
│                       │                                      │
│                       ▼                                      │
│              ┌──────────────────┐                           │
│              │   store.Store    │                           │
│              │  (SQLite backend)│                           │
│              │  - Upsert()      │                           │
│              │  - Find()        │                           │
│              │  - List()        │                           │
│              └──────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

---

## Gap Analysis

### What exists today for getting data out?

| Feature | Output | Granularity | Use Case |
|---------|--------|-------------|----------|
| `glaze help --list` | Terminal table | Slug + title | Human browsing |
| `glaze help --query "..."` | Terminal list | Slug + title | Human search |
| `glaze serve-help` | HTTP JSON API | Full sections | SPA browser |
| `glaze render-site` | Static SPA + JSON | Full sections | Publish website |
| `glaze docs <files>` | Glazed table | Frontmatter only | Analyze markdown files on disk |

### What is missing?

1. **No structured export from the live help system.**
   - `docs` only reads files you explicitly pass it; it does not talk to the `HelpSystem`.
   - There is no command that says "list every section in the current binary, with all metadata fields and content, as JSON/CSV."

2. **No disk-export to plain files or SQLite.**
   - `render-site` produces a React SPA - great for hosting, bad for piping into other tools.
   - There is no way to say "write each help section to a `.md` file" or "dump everything into a `.sqlite` file I can query with SQL."

3. **No filtering at export time.**
   - Even if we added export, we would want to support `--type`, `--topic`, `--command` filters so users do not have to post-process.

### Why this matters

- **Composability:** Unix philosophy says tools should output data in formats other tools can consume. JSON, CSV, and SQLite are lingua franca.
- **Offline access:** A SQLite file can be queried with any SQLite client, synced to mobile devices, or embedded into another application.
- **Auditing:** CI pipelines can export metadata and assert that every command has at least one help section.
- **Integration:** Documentation platforms, LLM context builders, and IDE plugins can consume structured help exports.

---

## Proposed Solution

### Overview

Add a single `export` subcommand under the existing `help` command:

```
glaze help export [flags]
```

The command operates on the `HelpSystem` already initialized by the binary. It uses the same `Store.Find()` and `Store.List()` APIs that the interactive help system uses. The behavior branches based on `--format`:

### `glaze help export` - Default Tabular Mode

**Purpose:** Export help sections as structured data via the Glaze processor.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--type` | string | "" | Filter by section type (`GeneralTopic`, `Example`, `Application`, `Tutorial`) |
| `--topic` | string | "" | Filter by topic tag |
| `--command` | string | "" | Filter by associated command |
| `--flag` | string | "" | Filter by associated flag |
| `--slug` | string | "" | Filter by exact slug (or comma-separated list) |
| `--with-content` | bool | `true` | Include the `content` field in tabular output |
| `--format` | string | `"glazed"` | Output mode: `glazed` (tabular), `files`, or `sqlite` |
| `--output` | string | `"-"` | Output file path ("-" for stdout; used by `files`/`sqlite`) |
| `--flatten` | bool | `false` | For `files` mode: flatten directory structure |

**Why `--with-content=true` by default?**
- The most common use case is "get everything out of the binary." Including content by default means one command gives a complete snapshot.
- Users who want lightweight metadata can opt out with `--with-content=false`.
- This avoids the confusion of having two subcommands that users must choose between.

**Output schema (JSON, `--with-content=true`):**

```json
[
  {
    "id": 1,
    "slug": "help-system",
    "type": "GeneralTopic",
    "title": "Help System",
    "short": "Overview of the glazed help system",
    "topics": ["help", "documentation"],
    "flags": [],
    "commands": ["help"],
    "is_top_level": true,
    "show_per_default": true,
    "order": 0,
    "content": "The help system allows...",
    "created_at": "2025-04-28T10:00:00Z",
    "updated_at": "2025-04-28T10:00:00Z"
  }
]
```

**Output schema (JSON, `--with-content=false`):**

```json
[
  {
    "id": 1,
    "slug": "help-system",
    "type": "GeneralTopic",
    "title": "Help System",
    "short": "Overview of the glazed help system",
    "topics": ["help", "documentation"],
    "flags": [],
    "commands": ["help"],
    "is_top_level": true,
    "show_per_default": true,
    "order": 0,
    "created_at": "2025-04-28T10:00:00Z",
    "updated_at": "2025-04-28T10:00:00Z"
  }
]
```

### Disk-Export Modes (`--format files` and `--format sqlite`)

When `--format` is `files` or `sqlite`, the command switches from tabular output to file-writing mode and behaves as a `BareCommand`.

#### `files` mode (`--format files --output ./help-export`)

Each matching section is written as an individual `.md` file. The file contents are reconstructed frontmatter + body, so they are valid Glazed help sections that can be re-imported.

**Directory layout (default):**

```
help-export/
├── general-topics/
│   ├── help-system.md
│   └── markdown-style.md
├── examples/
│   └── help-example-1.md
├── applications/
│   └── declarative-config-plan-example.md
└── tutorials/
    └── writing-help-entries.md
```

**Directory layout with `--flatten`:**

```
help-export/
├── help-system.md
├── markdown-style.md
├── help-example-1.md
└── ...
```

**Reconstructed file format:**

Each file is written with YAML frontmatter reconstructed from the `model.Section` fields, followed by the original `Content`.

```yaml
---
Title: "Help System"
Slug: "help-system"
Short: "Overview of the glazed help system"
Topics:
- help
- documentation
Commands:
- help
SectionType: GeneralTopic
IsTopLevel: true
ShowPerDefault: true
Order: 0
---

The help system allows...
```

#### `sqlite` mode (`--format sqlite --output ./my-help.sqlite`)

A single SQLite database file is created with the full `sections` schema, including the `content` column. This is a **portable, queryable snapshot** of the help system.

```bash
glaze help export --format sqlite --output ./my-help.sqlite
sqlite3 ./my-help.sqlite "SELECT slug, title FROM sections WHERE section_type = 0;"
```

The SQLite schema matches the existing store schema exactly, so the exported file can be opened with `store.New("./my-help.sqlite")`.

### Filtering Architecture

Both commands reuse the existing **predicate system** in `pkg/help/store/query.go`.

A predicate is a function `func(*QueryCompiler)` that adds WHERE clauses, JOINs, ORDER BY, LIMIT, and OFFSET to a SQL query. The store provides many built-in predicates:

```go
// From pkg/help/store/query.go
store.IsType(model.SectionGeneralTopic)
store.HasTopic("help")
store.HasCommand("json")
store.HasFlag("verbose")
store.IsTopLevel()
store.SlugEquals("help-system")
store.SlugIn([]string{"a", "b"})
store.OrderByTitle()
store.Limit(10)
```

Predicates can be combined with boolean logic:

```go
store.And(pred1, pred2, pred3)
store.Or(pred1, pred2)
store.Not(pred)
```

**Pseudocode for building the export predicate:**

```go
func buildExportPredicate(flags ExportFlags) store.Predicate {
    var preds []store.Predicate

    if flags.Type != "" {
        st, _ := model.SectionTypeFromString(flags.Type)
        preds = append(preds, store.IsType(st))
    }
    if flags.Topic != "" {
        preds = append(preds, store.HasTopic(flags.Topic))
    }
    if flags.Command != "" {
        preds = append(preds, store.HasCommand(flags.Command))
    }
    if flags.Flag != "" {
        preds = append(preds, store.HasFlag(flags.Flag))
    }
    if flags.Slug != "" {
        slugs := strings.Split(flags.Slug, ",")
        if len(slugs) == 1 {
            preds = append(preds, store.SlugEquals(flags.Slug))
        } else {
            preds = append(preds, store.SlugIn(slugs))
        }
    }

    // Always order predictably
    base := store.OrderByOrder()
    if len(preds) == 0 {
        return base
    }
    return store.And(store.And(preds...), base)
}
```

This predicate is passed to `hs.Store.Find(ctx, predicate)` to retrieve the matching sections.

---

## Detailed Design: Command Implementations

### How Glazed Commands Work (Primer for Interns)

Before diving into the export commands, we need to understand how Glazed commands are structured. Glazed provides its own command framework on top of Cobra.

#### The `cmds.Command` interface

```go
// From pkg/cmds/cmds.go
type Command interface {
    Description() *CommandDescription
    ToYAML(w io.Writer) error
}
```

Every command must be able to describe itself (name, flags, arguments, help text) and serialize to YAML.

#### The `cmds.BareCommand` interface

```go
// From pkg/cmds/cmds.go
type BareCommand interface {
    Command
    Run(ctx context.Context, parsedValues *values.Values) error
}
```

A `BareCommand` receives parsed flag/argument values and runs its logic. This is the simplest kind of Glazed command.

#### The `cmds.GlazeCommand` interface

```go
// From pkg/cmds/cmds.go
type GlazeCommand interface {
    Command
    RunIntoGlazeProcessor(ctx context.Context, parsedValues *values.Values, gp middlewares.Processor) error
}
```

A `GlazeCommand` streams rows into a `GlazeProcessor`, which handles output formatting (JSON, CSV, table, YAML, etc.) automatically. This is ideal for commands that produce tabular data.

#### Building a Cobra command from a Glazed command

```go
// From pkg/cli/cobra.go
func BuildCobraCommand(cmd cmds.Command, options ...BuildOption) (*cobra.Command, error)
```

This function:
1. Reflects on the Glazed command's `CommandDescription`.
2. Creates a `cobra.Command` with matching flags and arguments.
3. Wires the `Run` method to parse Cobra flags into `*values.Values` and invoke the Glazed command.

**This is the standard pattern for adding new commands to `glaze`.**

### Single `ExportCommand` with Mode Switching

Instead of two separate commands, we implement one `ExportCommand` that behaves as either a `GlazeCommand` (tabular output) or a `BareCommand` (file output) depending on `--format`.

```go
// Pseudocode for the unified export command implementation

type ExportCommand struct {
    *cmds.CommandDescription
    helpSystem *help.HelpSystem
}

// ExportCommand implements both GlazeCommand and BareCommand behaviors
// depending on the --format flag. This is achieved by implementing
// RunIntoGlazeProcessor for tabular modes and Run for disk-export modes.

var _ cmds.GlazeCommand = (*ExportCommand)(nil)
var _ cmds.BareCommand = (*ExportCommand)(nil)

type ExportSettings struct {
    Type        string `glazed:"type"`
    Topic       string `glazed:"topic"`
    Command     string `glazed:"command"`
    Flag        string `glazed:"flag"`
    Slug        string `glazed:"slug"`
    WithContent bool   `glazed:"with-content"`
    Format      string `glazed:"format"`   // "glazed" | "files" | "sqlite"
    Output      string `glazed:"output"`
    Flatten     bool   `glazed:"flatten"`
}

func NewExportCommand(hs *help.HelpSystem) (*ExportCommand, error) {
    return &ExportCommand{
        CommandDescription: cmds.NewCommandDescription(
            "export",
            cmds.WithShort("Export help sections to external formats"),
            cmds.WithLong(`Export help sections as structured data or to disk.

By default, exports all matching sections as JSON/CSV/table via the Glazed
processor, including full markdown content (--with-content defaults to true).

Use --format files to write individual .md files, or --format sqlite to
produce a portable SQLite database. Use --with-content=false for lightweight
metadata-only exports.`),
            cmds.WithFlags(
                fields.New("type", fields.TypeString, fields.WithHelp("Filter by section type"), fields.WithDefault("")),
                fields.New("topic", fields.TypeString, fields.WithHelp("Filter by topic"), fields.WithDefault("")),
                fields.New("command", fields.TypeString, fields.WithHelp("Filter by command"), fields.WithDefault("")),
                fields.New("flag", fields.TypeString, fields.WithHelp("Filter by flag"), fields.WithDefault("")),
                fields.New("slug", fields.TypeString, fields.WithHelp("Filter by slug(s)"), fields.WithDefault("")),
                fields.New("with-content", fields.TypeBool, fields.WithHelp("Include content field in tabular output"), fields.WithDefault(true)),
                fields.New("format", fields.TypeString, fields.WithHelp("Export mode: glazed, files, sqlite"), fields.WithDefault("glazed")),
                fields.New("output", fields.TypeString, fields.WithHelp("Output path (- for stdout, or directory/file path)"), fields.WithDefault("-")),
                fields.New("flatten", fields.TypeBool, fields.WithHelp("Flatten directory structure in files mode"), fields.WithDefault(false)),
            ),
        ),
        helpSystem: hs,
    }, nil
}
```

#### Tabular mode (`--format glazed`) - implements `GlazeCommand`

```go
func (c *ExportCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedValues *values.Values,
    gp middlewares.Processor,
) error {
    settings := &ExportSettings{}
    if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
        return err
    }

    // Validate: this method should only be called for glazed format
    if settings.Format != "glazed" {
        return errors.New("RunIntoGlazeProcessor called with non-glazed format; use Run() for files/sqlite")
    }

    predicate := buildExportPredicate(settings)
    sections, err := c.helpSystem.Store.Find(ctx, predicate)
    if err != nil {
        return err
    }

    for _, section := range sections {
        row := types.NewRow(
            types.MRP("id", section.ID),
            types.MRP("slug", section.Slug),
            types.MRP("type", section.SectionType.String()),
            types.MRP("title", section.Title),
            types.MRP("short", section.Short),
            types.MRP("topics", section.Topics),
            types.MRP("flags", section.Flags),
            types.MRP("commands", section.Commands),
            types.MRP("is_top_level", section.IsTopLevel),
            types.MRP("show_per_default", section.ShowPerDefault),
            types.MRP("order", section.Order),
            types.MRP("created_at", section.CreatedAt),
            types.MRP("updated_at", section.UpdatedAt),
        )
        if settings.WithContent {
            row.Set("content", section.Content)
        }
        if err := gp.AddRow(ctx, row); err != nil {
            return err
        }
    }

    return nil
}
```

#### Disk-export mode (`--format files` / `--format sqlite`) - implements `BareCommand`

```go
func (c *ExportCommand) Run(ctx context.Context, parsedValues *values.Values) error {
    settings := &ExportSettings{}
    if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
        return err
    }

    predicate := buildExportPredicate(settings)
    sections, err := c.helpSystem.Store.Find(ctx, predicate)
    if err != nil {
        return err
    }

    switch settings.Format {
    case "glazed":
        // This should not happen if cli.BuildCobraCommand dispatches correctly,
        // but we handle it by delegating to the Glaze processor.
        return errors.New("use --output json/csv/table for glazed format, or call via GlazeCommand path")
    case "files":
        return exportToFiles(sections, settings)
    case "sqlite":
        return exportToSQLite(sections, settings)
    default:
        return fmt.Errorf("unknown format: %s", settings.Format)
    }
}
```

**Why this dual-interface approach?**
- `cli.BuildCobraCommand` inspects which interfaces a command implements. If `GlazeCommand` is present, it wires `--output json/csv/table/yaml` flags and creates a `GlazeProcessor`.
- For disk-export modes, we bypass the Glaze processor entirely and write files directly.
- At runtime, the `Run` method checks `--format` and delegates appropriately. The Cobra wiring may need a small adjustment to call `Run` instead of `RunIntoGlazeProcessor` when `--format` is not `glazed`.

#### `exportToFiles` pseudocode (unchanged from prior design)

```go
func exportToFiles(sections []*model.Section, settings *ExportContentSettings) error {
    outDir := settings.Output
    if err := os.MkdirAll(outDir, 0755); err != nil {
        return err
    }

    for _, section := range sections {
        dir := outDir
        if !settings.Flatten {
            dir = filepath.Join(outDir, slugify(section.SectionType.String()))
            if err := os.MkdirAll(dir, 0755); err != nil {
                return err
            }
        }

        path := filepath.Join(dir, section.Slug+".md")
        data, err := reconstructMarkdown(section)
        if err != nil {
            return err
        }
        if err := os.WriteFile(path, data, 0644); err != nil {
            return err
        }
    }
    return nil
}

func reconstructMarkdown(section *model.Section) ([]byte, error) {
    // Use a structured map and yaml.Marshal to reconstruct frontmatter
    frontmatter := map[string]interface{}{
        "Title":          section.Title,
        "Slug":           section.Slug,
        "Short":          section.Short,
        "SectionType":    section.SectionType.String(),
        "Topics":         section.Topics,
        "Flags":          section.Flags,
        "Commands":       section.Commands,
        "IsTopLevel":     section.IsTopLevel,
        "IsTemplate":     section.IsTemplate,
        "ShowPerDefault": section.ShowPerDefault,
        "Order":          section.Order,
    }
    if section.SubTitle != "" {
        frontmatter["SubTitle"] = section.SubTitle
    }

    var buf bytes.Buffer
    buf.WriteString("---\n")
    if err := yaml.NewEncoder(&buf).Encode(frontmatter); err != nil {
        return nil, err
    }
    buf.WriteString("---\n")
    buf.WriteString(section.Content)
    return buf.Bytes(), nil
}
```

#### `exportToSQLite` pseudocode (unchanged from prior design)

```go
func exportToSQLite(sections []*model.Section, settings *ExportContentSettings) error {
    dbPath := settings.Output
    if !strings.HasSuffix(dbPath, ".sqlite") && !strings.HasSuffix(dbPath, ".db") {
        dbPath += ".sqlite"
    }

    // Remove existing file to start fresh
    _ = os.Remove(dbPath)

    // Use the existing store.New constructor
    exportStore, err := store.New(dbPath)
    if err != nil {
        return err
    }
    defer exportStore.Close()

    for _, section := range sections {
        // Upsert creates or updates; since the DB is fresh, it always inserts
        if err := exportStore.Upsert(ctx, section); err != nil {
            return err
        }
    }

    return nil
}
```

**Why reuse `store.New`?**
- The schema and indexes are created automatically.
- `Upsert` handles both insert and update semantics.
- No raw SQL needed, so the implementation stays maintainable.

---

## Wiring the Commands into the Application

The export command must be registered as a subcommand of `glaze help`. The existing `help` command is created inside `help_cmd.SetupCobraRootCommand`. After that call, `rootCmd` has a `help` subcommand. We retrieve it and add the single `export` child.

### Wiring in `cmd/glaze/main.go`

```go
// In cmd/glaze/main.go, after SetupCobraRootCommand

// Find the help subcommand
helpCmd, _, err := rootCmd.Find([]string{"help"})
if err != nil {
    cobra.CheckErr(err)
}

// Build the export command
exportGlazedCmd, err := NewExportCommand(helpSystem)
cobra.CheckErr(err)

exportCobraCmd, err := cli.BuildCobraCommand(
    exportGlazedCmd,
    cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
)
cobra.CheckErr(err)

helpCmd.AddCommand(exportCobraCmd)
```

**Why this approach:**
- It keeps the `pkg/help/cmd` package focused on interactive help behavior.
- It follows the same pattern used for `serve-help` and `render-site` in `main.go`.
- It allows each binary to decide whether to expose export functionality.

### Reusable helper for other binaries

To make this available to *any* binary using the glazed help system, provide a helper function:

```go
// pkg/help/cmd/export.go

func AddExportCommand(helpCmd *cobra.Command, hs *help.HelpSystem) error {
    exportGlazedCmd, err := NewExportCommand(hs)
    if err != nil {
        return err
    }
    exportCobraCmd, err := cli.BuildCobraCommand(
        exportGlazedCmd,
        cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
    )
    if err != nil {
        return err
    }
    helpCmd.AddCommand(exportCobraCmd)
    return nil
}
```

Then each binary's `main.go` can call:

```go
helpCmd, _, _ := rootCmd.Find([]string{"help"})
help_cmd.AddExportCommand(helpCmd, helpSystem)
```

---

## Implementation Phases

### Phase 1: Add `export` command with tabular output

1. **Create `pkg/help/cmd/export.go`**
   - Define `ExportCommand` struct implementing both `cmds.GlazeCommand` and `cmds.BareCommand`.
   - Implement `buildExportPredicate` helper.
   - Implement `RunIntoGlazeProcessor` for `--format glazed` (default).
   - Wire `--with-content` flag (default `true`) to conditionally include the `content` field.

2. **Add unit tests in `pkg/help/cmd/export_test.go`**
   - Create an in-memory `HelpSystem`, seed it with test sections.
   - Run the command with `--output json` and assert `content` is present.
   - Run with `--with-content=false` and assert `content` is absent.
   - Run with `--type Example` and assert only example rows are emitted.

3. **Register in `cmd/glaze/main.go`**
   - Find the `help` subcommand.
   - Add `export` via `help_cmd.AddExportCommand`.

4. **Validate manually:**
   ```bash
   go run ./cmd/glaze help export --output json
   go run ./cmd/glaze help export --with-content=false --output csv
   go run ./cmd/glaze help export --type GeneralTopic --output yaml
   ```

### Phase 2: Add disk-export modes (`files` and `sqlite`)

1. **Extend `ExportCommand.Run`**
   - Implement `exportToFiles` with directory creation and markdown reconstruction.
   - Implement `exportToSQLite` using `store.New` and `Upsert`.
   - Ensure `Run` dispatches based on `--format`.

2. **Add unit tests**
   - Test `files` mode: assert directory structure and file contents round-trip correctly.
   - Test `sqlite` mode: assert exported DB can be opened and queried.

3. **Validate manually:**
   ```bash
   go run ./cmd/glaze help export --format files --output /tmp/help-files
   go run ./cmd/glaze help export --format sqlite --output /tmp/help.sqlite
   sqlite3 /tmp/help.sqlite "SELECT slug, title FROM sections;"
   ```

### Phase 3: Documentation and cross-binary validation

1. **Documentation**
   - Add a help section (`.md` file) documenting the export feature.
   - Update `pkg/doc/` with examples showing `--with-content`, `--format files`, and `--format sqlite`.

2. **Cross-binary validation**
   - Test in another binary (e.g., `pinocchio` or `parka` if available).
   - Verify that `AddExportCommand` works without code duplication.

---

## Testing Strategy

### Unit Tests

**For tabular output (`--format glazed`):**
- Seed an in-memory store with 3 sections of mixed types.
- Run with no filters and `--output json` → expect 3 rows, each with `content` field present.
- Run with `--with-content=false` → expect 3 rows, each without `content` field.
- Run with `--type Example` → expect 1 row.
- Run with `--topic help` → expect rows where `Topics` contains `"help"`.

**For disk-export (`--format files`):**
- Seed store, run `--format files --output /tmp/test`.
- Assert directory exists.
- Assert each section produces one `.md` file.
- Parse reconstructed markdown back into a `Section` and assert equality.

**For disk-export (`--format sqlite`):**
- Seed store, run `--format sqlite --output /tmp/test.sqlite`.
- Open the DB with `store.New("/tmp/test.sqlite")`.
- Query all sections and assert count and content match.

### Integration Tests

- Run the full `glaze help export metadata` and `glaze help export content` commands via `go run`.
- Verify exit code is 0.
- Verify output files are created and non-empty.
- Verify SQLite file can be queried with the `sqlite3` CLI.

### Manual Verification Checklist

```bash
# Export everything (metadata + content) as JSON
glaze help export --output json

# Export lightweight metadata only
glaze help export --with-content=false --output csv

# Filter and format
glaze help export --type Example --output yaml

# Export all content as markdown files
glaze help export --format files --output ./exported-help

# Export to SQLite
glaze help export --format sqlite --output ./help.db

# Query the SQLite file
sqlite3 ./help.db "SELECT slug, title FROM sections WHERE section_type = 0 ORDER BY order_num;"
```

---

## Risks, Alternatives, and Open Questions

### Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Reconstructed markdown differs slightly from source (whitespace, field ordering) | High | Low | Document that export is round-trippable but not byte-identical. Use canonical YAML ordering. |
| Large content export slows down or OOMs | Low | Medium | Add `--limit` and `--offset` flags. SQLite mode is naturally streaming. |
| Binary bloat from new command code | Low | Low | The commands are small; they reuse existing packages. |

### Alternatives Considered

1. **Extend `glaze docs` to read from HelpSystem instead of files.**
   - Rejected: `docs` is designed for analyzing arbitrary markdown files on disk, not for querying the embedded help system. Its flag surface is different.

2. **Add `--export` flag to existing `glaze help` instead of a subcommand.**
   - Rejected: The `help` command already has many flags. A subcommand tree (`help export ...`) is clearer and more discoverable.

3. **Use `render-site` for all export needs.**
   - Rejected: `render-site` produces a React SPA with hashed asset names. It is not suitable for data pipelines or simple file consumption.

4. **Export only to SQLite, not files.**
   - Rejected: Markdown files are the universal format for documentation pipelines. Users will want them for Git repos, static site generators, and editors.

### Open Questions

1. **Should `--with-content` default to `true` or `false`?**
   - Decision: Default to `true`. The primary use case is "get everything out of the binary." Users who want lightweight output opt out with `--with-content=false`.

2. **Should we support `--template` for customizing file output?**
   - Proposal: Allow Go templates for `files` mode naming.
   - Decision: Defer. Start with sensible defaults.

3. **Should exported files include `Source` or provenance metadata?**
   - Proposal: Add an `ExportedFrom` field to frontmatter.
   - Decision: No — keep files clean so they can be re-imported into any binary.

---

## API Reference Summary

### New Types

```go
// pkg/help/cmd/export.go

type ExportFlags struct {
    Type    string
    Topic   string
    Command string
    Flag    string
    Slug    string
}

func BuildExportPredicate(f ExportFlags) store.Predicate
```

### New Command

```go
// pkg/help/cmd/export.go

func NewExportCommand(hs *help.HelpSystem) (*ExportCommand, error)

// Implements both: cmds.GlazeCommand (for --format glazed)
//                  cmds.BareCommand (for --format files/sqlite)
```

### Registration Helper

```go
// pkg/help/cmd/export.go

func AddExportCommand(helpCmd *cobra.Command, hs *help.HelpSystem) error
```

---

## File References

| File | Role |
|------|------|
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/model/section.go` | `Section` struct definition |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/model/parse.go` | Markdown frontmatter parser |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go` | SQLite store implementation |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/query.go` | Predicate query system |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/help.go` | `HelpSystem` facade |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/cmd/cobra.go` | Cobra help command wiring |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/cmds/cmds.go` | Glazed command interfaces |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/cli/cobra.go` | `BuildCobraCommand` function |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go` | Application root |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/site/render.go` | Prior art for export logic |

---

## Appendix: Full Command Examples

### Export everything (metadata + content) as JSON

```bash
glaze help export --output json
```

Output:
```json
[
  {"id":1,"slug":"help-system","type":"GeneralTopic","title":"Help System","content":"The help system allows...",...},
  {"id":2,"slug":"markdown-style","type":"GeneralTopic","title":"Markdown Style","content":"Markdown is a lightweight...",...}
]
```

### Export lightweight metadata only, as CSV

```bash
glaze help export --with-content=false --type Example --output csv
```

Output:
```csv
id,slug,type,title,short,topics,flags,commands,is_top_level,show_per_default,order,created_at,updated_at
3,help-example-1,Example,Show the list of all toplevel topics,...,[],[],[help],false,false,0,...
```

### Export all content as markdown files

```bash
glaze help export --format files --output ./my-help
```

Result:
```
./my-help/
├── general-topics/
│   └── help-system.md
├── examples/
│   └── help-example-1.md
└── ...
```

### Export to SQLite and query

```bash
glaze help export --format sqlite --output ./my-help.sqlite
sqlite3 ./my-help.sqlite ".schema sections"
sqlite3 ./my-help.sqlite "SELECT slug, title FROM sections WHERE topics LIKE '%help%';"
```

### Filtered export: only `json` command documentation

```bash
glaze help export --command json --format files --output ./json-docs
glaze help export --command json --with-content=false --output json
```

---

## Conclusion

This design adds a small, focused, and reusable export capability to the Glazed help system. By leveraging the existing `Store` and predicate APIs, the implementation stays thin and idiomatic. The single `glaze help export` verb with `--with-content=true` by default gives users a complete snapshot with one command, while `--with-content=false`, `--format files`, and `--format sqlite` provide lightweight and archival alternatives. The `AddExportCommand` helper ensures that any binary built on Glazed can opt into this feature with a single function call.
