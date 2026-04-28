---
Title: 'Design: Serve External Help Sources - Multi-Source Help Browser'
Ticket: GLAZE-HELP-EXPORT
Status: active
Topics:
    - glazed
    - help-system
    - server
    - cli
    - import
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/doc/topics/29-serve-external-help-sources.md
      Note: User-facing help topic for external sources
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/cmd/export.go
      Note: Produces JSON/SQLite consumed by serve loaders
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/loader/paths.go
      Note: Existing markdown loader to reuse/wrap
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/loader/sources.go
      Note: ContentLoader implementations for external sources
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/loader/sources_test.go
      Note: Unit tests for JSON/SQLite/command loaders
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/model/section.go
      Note: Section and SectionType conversion details
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go
      Note: ServeCommand to extend with external source flags
    - Path: ../../../../../../../../../../code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go
      Note: Store.List, Store.Upsert, Store.Clear used by loaders
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-28T07:35:00-04:00
WhatFor: Design implementation of multi-source loading for glaze serve, including --from-glazed-cmd convenience loading from other Glazed binaries.
WhenToUse: Use when implementing or reviewing external-source support for the Glazed help browser server.
---




# Design: Serve External Help Sources — Multi-Source Help Browser

## Executive Summary

This document designs an extension to `glaze serve` so it can serve help pages from **other tools and external snapshots**, not only from the help system embedded in the `glaze` binary. The command should accept several source types and merge them into one in-memory `HelpSystem` before starting the existing HTTP API and React help browser.

The most important new source is a convenience flag:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

`--from-glazed-cmd` treats each value as the name or path of a Glazed-compatible binary. For each binary, `glaze serve` automatically runs:

```bash
<binary> help export --output json
```

Then it parses the JSON output and inserts the exported help sections into the server's store. This makes cross-package help browsing easy: users do not need to remember the full `help export --output json` invocation for every tool.

The more general form remains available:

```bash
glaze serve --from-cmd "pinocchio help export --output json"
```

The final design supports five source families:

1. Embedded docs already loaded into the current binary's `HelpSystem`
2. Markdown files or directories from positional `paths`
3. JSON files from `--from-json`
4. SQLite databases from `--from-sqlite`
5. Command outputs from `--from-cmd` and the convenience `--from-glazed-cmd`

All list-like inputs must use `fields.TypeStringList` and `[]string` settings fields. The implementation should accept both repeated flags and comma-separated values where Glazed's string-list parsing permits it.

---

## Terminology

| Term | Meaning |
|------|---------|
| Help section | A `model.Section` record with slug, title, markdown content, type, topics, commands, flags, and display metadata. |
| Embedded docs | Help sections loaded at process startup from `pkg/doc` via Go `embed`. |
| External source | Any help data loaded after startup from a file, database, or command. |
| Export JSON row | The JSON object produced by `glaze help export --output json`; it is a Glazed row, not necessarily a direct `model.Section` serialization. |
| Glazed command source | A binary that supports `help export --output json`, loaded through `--from-glazed-cmd`. |

---

## Current-State Architecture

`glaze serve` is implemented by `ServeCommand` in:

```text
/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go
```

The command currently has one flag and one positional argument group:

```go
type ServeSettings struct {
    Address string   `glazed:"address"`
    Paths   []string `glazed:"paths"`
}
```

The current run flow is:

1. `cmd/glaze/main.go` creates a `help.HelpSystem`.
2. `doc.AddDocToHelpSystem(helpSystem)` loads embedded Glazed help docs.
3. `server.NewServeCommand(helpSystem, spaHandler)` receives that already-loaded help system.
4. `ServeCommand.Run` decodes `--address` and positional `paths`.
5. If `paths` are provided, it calls `helploader.ReplaceStoreWithPaths(ctx, hs, s.Paths)`, which clears the store and loads markdown from disk.
6. If no `paths` are provided, it serves the embedded docs already in the store.
7. `NewServeHandler` exposes `/api/health`, `/api/sections`, and `/api/sections/{slug}`.

The HTTP layer is source-agnostic. Once sections are in `hs.Store`, the API and SPA do not care whether those sections came from embedded markdown, JSON, SQLite, or another binary.

---

## Design Review of the Previous Draft

The previous draft was directionally correct but had several issues that this revision fixes.

### Issue 1: It contradicted itself about embedded docs

The earlier text first said external sources would replace embedded docs, then reconsidered and proposed `--with-embedded=true`. This document resolves that ambiguity:

- `glaze serve` with no sources serves embedded docs exactly as today.
- If external sources are provided, `--with-embedded=true` means keep embedded docs and merge external sources.
- If external sources are provided with `--with-embedded=false`, clear the store first and serve only explicit sources.

### Issue 2: It did not include `--from-glazed-cmd`

The earlier draft only had `--from-cmd`, requiring users to type the full export command repeatedly. The revised design adds `--from-glazed-cmd` as the ergonomic default for Glazed-compatible binaries.

### Issue 3: It assumed JSON uses `section_type`

The current `ExportCommand` emits Glazed rows. In `pkg/help/cmd/export.go`, the row key is currently `type`, not `section_type`:

```go
types.MRP("type", section.SectionType.String())
```

A robust JSON loader must accept both shapes:

- current export row shape: `type: "GeneralTopic"`
- direct model shape: `section_type: "GeneralTopic"` or numeric `section_type: 0`

### Issue 4: It treated several list inputs as single strings

The implementation must use `fields.TypeStringList` for all multi-source string inputs. Settings fields must be `[]string`, not `string`, for:

- `Paths`
- `FromJSON`
- `FromSQLite`
- `FromCmd`
- `FromGlazedCmd`

Loader structs may internally represent a single source or a grouped list, but the public command schema should expose list values.

### Issue 5: It used unsafe/underspecified command waiting pseudocode

The previous `CommandLoader` used `defer cmd.Wait()` while decoding stdout. A real implementation must always check the process exit status after decoding, otherwise a command that writes partial JSON then exits non-zero could be treated incorrectly. This revision spells out the safer flow.

---

## Goals and Non-Goals

### Goals

1. Make `glaze serve` a universal browser for Glazed help pages from many binaries.
2. Add `--from-glazed-cmd` for the common case where each input is a Glazed binary name.
3. Keep `--from-cmd` for advanced users who need a full command line.
4. Load JSON exports, SQLite exports, and markdown directories.
5. Let users combine several sources in one server instance.
6. Preserve current behavior when no external source flags or paths are provided.
7. Use `fields.TypeStringList` for every multi-source string input.
8. Fail before starting the HTTP server if any source cannot be loaded.

### Non-Goals

- No live reload/watch mode in the MVP.
- No HTTP URL source in the MVP.
- No authentication changes to the help server.
- No schema redesign for `model.Section`.
- No shell pipeline support inside `--from-glazed-cmd`.

---

## Proposed CLI Surface

### Existing behavior remains valid

```bash
# Serve embedded docs from the current glaze binary
glaze serve

# Serve markdown files or directories
glaze serve ./pkg/doc ./extra-docs

# Serve on another address
glaze serve --address :18100 ./pkg/doc
```

### New convenience source: `--from-glazed-cmd`

Use this when each value is a binary that supports the Glazed help export verb.

```bash
# Comma-separated list, as requested
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx

# Repeated flag form should also work
glaze serve \
  --from-glazed-cmd pinocchio \
  --from-glazed-cmd sqleton \
  --from-glazed-cmd xxx
```

For each binary name, the loader runs:

```bash
<binary> help export --output json
```

So the first example expands conceptually to:

```bash
pinocchio help export --output json
sqleton help export --output json
xxx help export --output json
```

### General command source: `--from-cmd`

Use this when the command is not exactly `<binary> help export --output json`, or when you need extra flags.

```bash
glaze serve --from-cmd "pinocchio help export --output json --topic database"
```

This flag also uses `fields.TypeStringList`, so users can pass several commands:

```bash
glaze serve \
  --from-cmd "pinocchio help export --output json" \
  --from-cmd "sqleton help export --output json --with-content=true"
```

### JSON source: `--from-json`

```bash
# Load from one JSON file
glaze serve --from-json ./pinocchio-help.json

# Load from several JSON files
glaze serve --from-json ./pinocchio-help.json,./sqleton-help.json

# Read from stdin
glaze help export --output json | glaze serve --from-json -
```

### SQLite source: `--from-sqlite`

```bash
# Load one exported DB
glaze serve --from-sqlite ./pinocchio-help.db

# Load several DBs
glaze serve --from-sqlite ./pinocchio-help.db,./sqleton-help.db
```

### Combined source example

```bash
glaze serve \
  --from-glazed-cmd pinocchio,sqleton \
  --from-json ./team-overrides.json \
  --from-sqlite ./legacy-help.db \
  ./local-markdown-docs
```

This produces one help browser containing:

- embedded Glazed docs, only when `--with-embedded=true` is provided,
- live exports from `pinocchio` and `sqleton`,
- JSON override docs,
- archived SQLite docs,
- local markdown docs.

---

## ServeSettings and Flag Definitions

All string source settings are lists. This is non-negotiable because the main value of this feature is serving multiple packages together.

```go
type ServeSettings struct {
    Address       string   `glazed:"address"`
    Paths         []string `glazed:"paths"`
    FromJSON      []string `glazed:"from-json"`
    FromSQLite    []string `glazed:"from-sqlite"`
    FromCmd       []string `glazed:"from-cmd"`
    FromGlazedCmd []string `glazed:"from-glazed-cmd"`
    WithEmbedded  bool     `glazed:"with-embedded"`
}
```

Command schema additions:

```go
cmds.WithFlags(
    fields.New(
        "address",
        fields.TypeString,
        fields.WithHelp("Address to listen on"),
        fields.WithDefault(DefaultAddr),
    ),
    fields.New(
        "from-json",
        fields.TypeStringList,
        fields.WithHelp("JSON help export files to load; use - for stdin"),
    ),
    fields.New(
        "from-sqlite",
        fields.TypeStringList,
        fields.WithHelp("SQLite help export databases to load"),
    ),
    fields.New(
        "from-cmd",
        fields.TypeStringList,
        fields.WithHelp("Commands to run; stdout must be a JSON help export"),
    ),
    fields.New(
        "from-glazed-cmd",
        fields.TypeStringList,
        fields.WithHelp("Glazed binaries to load by running '<binary> help export --output json'"),
    ),
    fields.New(
        "with-embedded",
        fields.TypeBool,
        fields.WithHelp("Include already-loaded embedded docs when external sources are provided"),
        fields.WithDefault(false),
    ),
)
```

The positional `paths` argument remains a string list:

```go
cmds.WithArguments(
    fields.New(
        "paths",
        fields.TypeStringList,
        fields.WithHelp("Markdown files or directories to load"),
    ),
)
```

---

## Source Loading Semantics

### Default behavior

If no external source flags and no positional paths are provided:

```bash
glaze serve
```

The command serves the embedded docs already loaded by `cmd/glaze/main.go`. This is exactly the current behavior.

### External source behavior without embedded docs (default)

If any external source is provided and `--with-embedded=false` (the default), the command should clear the store before loading explicit sources. This keeps `glaze serve --from-glazed-cmd pinocchio` focused on Pinocchio docs instead of mixing in Glazed framework docs.

```bash
glaze serve --from-glazed-cmd pinocchio
```

Result:

1. Clear embedded Glazed docs.
2. Load Pinocchio docs.
3. Serve only Pinocchio docs.

### External source behavior with embedded docs

If any external source is provided and `--with-embedded=true`, the command should **keep** the currently loaded embedded docs and then merge the external sources into the same store.

```bash
glaze serve --from-glazed-cmd pinocchio
```

Result:

1. Start with embedded Glazed docs.
2. Load Pinocchio docs from `pinocchio help export --output json`.
3. Serve both sets together.

### Loading order

Loaders should run in a deterministic order. Later sources overwrite earlier sources when slugs collide.

Recommended order:

1. Existing embedded docs (only if `--with-embedded=true`, already loaded before `Run`)
2. Positional markdown `paths`
3. `--from-json` sources, in normalized list order
4. `--from-sqlite` sources, in normalized list order
5. `--from-cmd` sources, in normalized list order
6. `--from-glazed-cmd` sources, in normalized list order

Why put `--from-glazed-cmd` last? It is the most live/authoritative source for another binary. If a JSON snapshot and a live binary contain the same slug, the live binary should win.

### Collision handling

Use `Store.Upsert(ctx, section)`. If two sources contain the same slug, the later source replaces the earlier section.

The implementation should log collisions, but collisions should not fail the load:

```go
existing, err := hs.Store.GetBySlug(ctx, section.Slug)
if err == nil && existing != nil {
    log.Warn().Str("slug", section.Slug).Str("source", loader.String()).Msg("Replacing existing help section")
}
```

---

## ContentLoader Interface

All external sources should be represented by a small common interface.

```go
type ContentLoader interface {
    Load(ctx context.Context, hs *help.HelpSystem) error
    String() string
}
```

`Load` inserts sections into the provided help system. It should not clear the store; clearing is orchestrated once by `ServeCommand.Run` based on `--with-embedded`.

`String` returns a human-readable source description for logs and errors.

Recommended package location:

```text
pkg/help/loader/sources.go
```

This keeps source loading near the existing markdown path loader while avoiding direct dependency from `pkg/help/server` on low-level parsing details.

---

## Loader Designs

### MarkdownPathLoader

`MarkdownPathLoader` wraps the existing `loader.LoadPaths` function.

```go
type MarkdownPathLoader struct {
    Paths []string
}

func (l *MarkdownPathLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
    return loader.LoadPaths(ctx, hs, l.Paths)
}

func (l *MarkdownPathLoader) String() string {
    return "markdown paths: " + strings.Join(l.Paths, ", ")
}
```

Use one grouped loader for all positional paths. The input is already a `[]string`.

### JSONFileLoader

`JSONFileLoader` loads one or more JSON files. It must accept the actual `glaze help export --output json` row format and a direct `model.Section` JSON format.

```go
type JSONFileLoader struct {
    Paths []string
}
```

The loader should support `-` for stdin. Because stdin can only be consumed once, reject multiple `-` values:

```go
if count(paths, "-") > 1 {
    return errors.New("--from-json - may only be used once")
}
```

#### JSON input shapes to support

Current export row shape:

```json
{
  "slug": "help-system",
  "type": "GeneralTopic",
  "title": "Help System",
  "short": "...",
  "content": "...",
  "topics": ["help"],
  "flags": ["topic"],
  "commands": ["help"],
  "is_top_level": true,
  "show_per_default": true,
  "order": 0
}
```

Direct model shape:

```json
{
  "slug": "help-system",
  "section_type": "GeneralTopic",
  "title": "Help System",
  "short": "...",
  "content": "..."
}
```

The implementation should define a small import DTO instead of unmarshaling directly into `model.Section`:

```go
type sectionImportRow struct {
    ID             int64           `json:"id,omitempty"`
    Slug           string          `json:"slug"`
    Type           json.RawMessage `json:"type,omitempty"`
    SectionType    json.RawMessage `json:"section_type,omitempty"`
    Title          string          `json:"title"`
    SubTitle       string          `json:"sub_title"`
    Short          string          `json:"short"`
    Content        string          `json:"content"`
    Topics         []string        `json:"topics"`
    Flags          []string        `json:"flags"`
    Commands       []string        `json:"commands"`
    IsTopLevel     bool            `json:"is_top_level"`
    IsTemplate     bool            `json:"is_template"`
    ShowPerDefault bool            `json:"show_per_default"`
    Order          int             `json:"order"`
    CreatedAt      string          `json:"created_at"`
    UpdatedAt      string          `json:"updated_at"`
}
```

Then convert it:

```go
func (r sectionImportRow) ToSection() (*model.Section, error) {
    st, err := parseSectionType(r.Type, r.SectionType)
    if err != nil {
        return nil, err
    }
    return &model.Section{
        ID:             r.ID,
        Slug:           r.Slug,
        SectionType:    st,
        Title:          r.Title,
        SubTitle:       r.SubTitle,
        Short:          r.Short,
        Content:        r.Content,
        Topics:         r.Topics,
        Flags:          r.Flags,
        Commands:       r.Commands,
        IsTopLevel:     r.IsTopLevel,
        IsTemplate:     r.IsTemplate,
        ShowPerDefault: r.ShowPerDefault,
        Order:          r.Order,
        CreatedAt:      r.CreatedAt,
        UpdatedAt:      r.UpdatedAt,
    }, nil
}
```

`parseSectionType` should accept:

- string values: `"GeneralTopic"`, `"Example"`, `"Application"`, `"Tutorial"`
- numeric values: `0`, `1`, `2`, `3`
- missing value: default to `GeneralTopic` only if that is acceptable; otherwise fail clearly. Recommendation: fail for missing type unless compatibility demands a default.

### SQLiteLoader

`SQLiteLoader` loads one or more SQLite databases previously produced by:

```bash
<binary> help export --format sqlite --output-path ./help.db
```

```go
type SQLiteLoader struct {
    Paths []string
}

func (l *SQLiteLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
    for _, path := range l.Paths {
        sourceStore, err := store.New(path)
        if err != nil {
            return fmt.Errorf("open sqlite %s: %w", path, err)
        }
        sections, listErr := sourceStore.List(ctx, "")
        closeErr := sourceStore.Close()
        if listErr != nil {
            return fmt.Errorf("list sections from %s: %w", path, listErr)
        }
        if closeErr != nil {
            return fmt.Errorf("close sqlite %s: %w", path, closeErr)
        }
        for _, section := range sections {
            if err := upsertWithCollisionLog(ctx, hs, section, l.String()); err != nil {
                return err
            }
        }
    }
    return nil
}
```

### CommandJSONLoader

`CommandJSONLoader` runs arbitrary user-provided commands whose stdout is expected to be JSON help export data.

```go
type CommandJSONLoader struct {
    Commands []string
}
```

This is the advanced form. It should be used when users need custom flags:

```bash
glaze serve --from-cmd "pinocchio help export --output json --topic orm"
```

A safe implementation must:

1. Tokenize the command without invoking a shell.
2. Start the process with `exec.CommandContext`.
3. Capture stdout and stderr separately.
4. Decode stdout as JSON.
5. Wait for the process and check the exit status.
6. Return stderr in the error message if the command fails.

Pseudocode:

```go
func runCommandForJSON(ctx context.Context, command string) ([]byte, error) {
    args, err := tokenizeCommand(command)
    if err != nil {
        return nil, err
    }
    if len(args) == 0 {
        return nil, errors.New("empty command")
    }

    cmd := exec.CommandContext(ctx, args[0], args[1:]...)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err = cmd.Run()
    if err != nil {
        return nil, fmt.Errorf("command %q failed: %w; stderr: %s", command, err, stderr.String())
    }
    return stdout.Bytes(), nil
}
```

This simpler `cmd.Run()` form is preferable to manual `StdoutPipe` unless streaming is required. Help exports are finite and should fit in memory for MVP.

### GlazedCommandLoader

`GlazedCommandLoader` is the ergonomic loader requested by the user. It accepts binary names or executable paths and constructs the export command automatically.

```go
type GlazedCommandLoader struct {
    Binaries []string
}

func (l *GlazedCommandLoader) Load(ctx context.Context, hs *help.HelpSystem) error {
    for _, binary := range l.Binaries {
        data, err := runGlazedHelpExport(ctx, binary)
        if err != nil {
            return err
        }
        if err := importJSONBytes(ctx, hs, data, "glazed command: "+binary); err != nil {
            return err
        }
    }
    return nil
}
```

`runGlazedHelpExport` should not tokenize a full shell command. It should call the binary directly:

```go
func runGlazedHelpExport(ctx context.Context, binary string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, binary, "help", "export", "--output", "json")
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("%s help export failed: %w; stderr: %s", binary, err, stderr.String())
    }
    return stdout.Bytes(), nil
}
```

This approach avoids shell injection entirely and gives users a short CLI:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

---

## Loader Construction

`ServeCommand.Run` should build loaders from normalized settings values.

```go
func buildLoaders(s *ServeSettings) []loader.ContentLoader {
    var loaders []loader.ContentLoader

    if len(s.Paths) > 0 {
        loaders = append(loaders, &loader.MarkdownPathLoader{Paths: normalizeStringList(s.Paths)})
    }
    if len(s.FromJSON) > 0 {
        loaders = append(loaders, &loader.JSONFileLoader{Paths: normalizeStringList(s.FromJSON)})
    }
    if len(s.FromSQLite) > 0 {
        loaders = append(loaders, &loader.SQLiteLoader{Paths: normalizeStringList(s.FromSQLite)})
    }
    if len(s.FromCmd) > 0 {
        loaders = append(loaders, &loader.CommandJSONLoader{Commands: normalizeStringList(s.FromCmd)})
    }
    if len(s.FromGlazedCmd) > 0 {
        loaders = append(loaders, &loader.GlazedCommandLoader{Binaries: normalizeStringList(s.FromGlazedCmd)})
    }

    return loaders
}
```

`normalizeStringList` should remove empty values, trim whitespace, and split comma-separated entries if the Glazed field parsing layer does not already do so.

```go
func normalizeStringList(values []string) []string {
    var ret []string
    for _, value := range values {
        for _, part := range strings.Split(value, ",") {
            part = strings.TrimSpace(part)
            if part != "" {
                ret = append(ret, part)
            }
        }
    }
    return ret
}
```

This explicitly supports the desired syntax:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

---

## Updated ServeCommand.Run

The server command should orchestrate clearing/merging once, then delegate all source-specific work to loaders.

```go
func (sc *ServeCommand) Run(ctx context.Context, parsedValues *values.Values) error {
    s := &ServeSettings{}
    if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
        return fmt.Errorf("failed to decode serve settings: %w", err)
    }

    hs := sc.helpSystem
    if hs.Store == nil {
        return errors.New("HelpSystem.Store is nil")
    }

    loaders := buildLoaders(s)
    hasExplicitSources := len(loaders) > 0

    if hasExplicitSources && !s.WithEmbedded {
        if err := hs.Store.Clear(ctx); err != nil {
            return fmt.Errorf("failed to clear embedded help sections: %w", err)
        }
    }

    for _, l := range loaders {
        log.Info().Str("source", l.String()).Msg("Loading help source")
        if err := l.Load(ctx, hs); err != nil {
            return fmt.Errorf("failed to load %s: %w", l.String(), err)
        }
    }

    count, err := hs.Store.Count(ctx)
    if err != nil {
        return fmt.Errorf("failed to count help sections: %w", err)
    }
    log.Info().Int64("sections", count).Msg("Loaded help sections")

    deps := HandlerDeps{Store: hs.Store}
    handler := NewServeHandler(deps, sc.spaHandler)
    return serveHTTP(s.Address, handler)
}
```

Notice the important behavior: explicit sources clear embedded docs by default because `--with-embedded=false`. If the user sets `--with-embedded=true`, we do **not** clear and simply add sources to the store that already contains embedded docs.

---

## JSON Compatibility Design

The import side must support the JSON format actually emitted by `glaze help export --output json`.

Current export rows use these keys:

```go
types.MRP("id", section.ID)
types.MRP("slug", section.Slug)
types.MRP("type", section.SectionType.String())
types.MRP("title", section.Title)
types.MRP("short", section.Short)
types.MRP("topics", section.Topics)
types.MRP("flags", section.Flags)
types.MRP("commands", section.Commands)
types.MRP("is_top_level", section.IsTopLevel)
types.MRP("show_per_default", section.ShowPerDefault)
types.MRP("order", section.Order)
types.MRP("created_at", section.CreatedAt)
types.MRP("updated_at", section.UpdatedAt)
```

The loader must therefore not require `section_type`. It should prefer `type` when present, accept `section_type` for future/direct model JSON, and return a clear error if neither can be parsed.

### Optional improvement to export

A future cleanup could rename export's row field from `type` to `section_type` for consistency with `model.Section`. That would be a breaking-ish output change for scripts, so this design does not require it. Instead, the importer handles both.

---

## Implementation Phases

### Phase 1: Add import DTO and JSON parser tests

1. Add an internal `sectionImportRow` DTO in `pkg/help/loader`.
2. Implement `parseSectionType` accepting `type`, `section_type`, strings, and ints.
3. Implement `DecodeSectionsJSON(r io.Reader) ([]*model.Section, error)`.
4. Add tests for:
   - current export row with `type: "GeneralTopic"`
   - model-like row with `section_type: "Example"`
   - model-like row with `section_type: 2`
   - missing/invalid type
   - missing slug/title validation

This phase is safer than changing `model.SectionType.MarshalJSON` first, because it avoids changing exported JSON behavior before the importer exists.

### Phase 2: Add ContentLoader implementations

1. Add `ContentLoader` interface.
2. Add `MarkdownPathLoader`.
3. Add `JSONFileLoader` with `Paths []string`.
4. Add `SQLiteLoader` with `Paths []string`.
5. Add `CommandJSONLoader` with `Commands []string`.
6. Add `GlazedCommandLoader` with `Binaries []string`.
7. Add unit tests for each loader.

### Phase 3: Extend ServeCommand

1. Add `FromJSON`, `FromSQLite`, `FromCmd`, `FromGlazedCmd`, and `WithEmbedded` to `ServeSettings`.
2. Add flags using `fields.TypeStringList` for every string-list source.
3. Add `normalizeStringList` and `buildLoaders`.
4. Update `Run` to clear when explicit sources exist and `--with-embedded=false` (the default).
5. Add integration tests around `/api/health` and `/api/sections`.

### Phase 4: Documentation and manual validation

1. Update `pkg/doc/topics/25-serving-help-over-http.md` or add a dedicated `serve-external-sources` topic.
2. Update `pkg/doc/topics/28-export-help-entries.md` to mention `--from-glazed-cmd` as the serve-side counterpart.
3. Run the manual verification checklist below.
4. Commit.

---

## Testing Strategy

### Unit tests for JSON import

```go
func TestDecodeSectionsJSON_CurrentExportTypeField(t *testing.T) {
    input := `[{
      "slug":"help-system",
      "title":"Help System",
      "type":"GeneralTopic",
      "content":"body"
    }]`
    sections, err := DecodeSectionsJSON(strings.NewReader(input))
    require.NoError(t, err)
    assert.Equal(t, model.SectionGeneralTopic, sections[0].SectionType)
}
```

Also test:

- `section_type: "Tutorial"`
- `section_type: 3`
- invalid `type`
- missing `type` and `section_type`
- list fields preserved (`topics`, `flags`, `commands`)

### Unit tests for `--from-glazed-cmd`

Use a tiny test helper binary or `go test` helper process pattern. The helper should print valid JSON when invoked as:

```bash
helper help export --output json
```

Test that `GlazedCommandLoader{Binaries: []string{helperPath}}` loads the expected section.

### Unit tests for command failure

Verify that stderr is included in the returned error:

```bash
helper help export --output json
# exits 1, stderr: "boom"
```

Expected error includes `boom`.

### Integration tests for `ServeCommand`

- `glaze serve` with no sources serves embedded docs.
- `glaze serve --with-embedded=false --from-json file.json` serves only file JSON.
- `glaze serve --from-json file.json` serves only JSON because `--with-embedded` defaults to false.
- `glaze serve --from-glazed-cmd helper1,helper2` serves both helpers.
- Repeated flags and comma-separated values produce the same source list.
- Slug collision uses last-write-wins and logs a warning.

### Manual verification checklist

```bash
# 1. Export and serve from JSON
glaze help export --output json > /tmp/glaze-help.json
glaze serve --from-json /tmp/glaze-help.json --address :18101 &
curl http://localhost:18101/api/health
kill %1

# 2. Export and serve from SQLite
glaze help export --format sqlite --output-path /tmp/glaze-help.db
glaze serve --from-sqlite /tmp/glaze-help.db --address :18102 &
curl http://localhost:18102/api/health
kill %1

# 3. Serve from full command
glaze serve --from-cmd "glaze help export --output json" --address :18103 &
curl http://localhost:18103/api/health
kill %1

# 4. Serve from Glazed command shorthand
glaze serve --from-glazed-cmd glaze --address :18104 &
curl http://localhost:18104/api/health
kill %1

# 5. Serve multiple Glazed commands using comma-separated syntax
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx --address :18105 &
curl http://localhost:18105/api/health
kill %1

# 6. Combine all source types
glaze serve \
  --from-glazed-cmd glaze \
  --from-json /tmp/glaze-help.json \
  --from-sqlite /tmp/glaze-help.db \
  ./pkg/doc/topics \
  --address :18106 &
curl http://localhost:18106/api/sections?limit=5
kill %1
```

---

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| `--from-cmd` command injection | Medium | High | Do not invoke a shell; tokenize and call `exec.CommandContext`. |
| `--from-glazed-cmd` binary not found | Medium | Low | Return a clear error naming the binary. |
| Export JSON shape mismatch | Medium | Medium | Import both `type` and `section_type`; add tests from real export output. |
| Slug collisions hide earlier docs | Medium | Low | Last-write-wins, with warning logs. |
| Multiple `--from-json -` values consume stdin twice | Low | Medium | Reject more than one stdin JSON source. |
| Very large exports consume memory | Low | Medium | Accept for MVP; streaming decode can be added later. |
| Command hangs | Low | High | Use `exec.CommandContext`; consider adding `--source-timeout` later. |

---

## Open Questions

1. Should `--from-glazed-cmd` support extra arguments per binary? Recommendation: no for MVP. Use `--from-cmd` when extra arguments are needed.
2. Should `--from-glazed-cmd` run `--with-content=true` explicitly? Recommendation: yes if the export command supports it, but it is already the default. The explicit command could be `<binary> help export --with-content=true --output json` for clarity.
3. Should JSON import skip invalid sections or fail the whole source? Recommendation: fail the source for invalid JSON shape, but skip individual sections only if we explicitly log and count skips. For MVP, fail fast is easier to debug.
4. Should embedded docs be included by default with external sources? Decision: no. `--with-embedded` defaults to `false`, so explicit external sources replace embedded docs unless the user opts into merging them.

---

## API Reference Summary

### New/changed settings

```go
type ServeSettings struct {
    Address       string   `glazed:"address"`
    Paths         []string `glazed:"paths"`
    FromJSON      []string `glazed:"from-json"`
    FromSQLite    []string `glazed:"from-sqlite"`
    FromCmd       []string `glazed:"from-cmd"`
    FromGlazedCmd []string `glazed:"from-glazed-cmd"`
    WithEmbedded  bool     `glazed:"with-embedded"`
}
```

### New loaders

```go
type ContentLoader interface {
    Load(ctx context.Context, hs *help.HelpSystem) error
    String() string
}

type MarkdownPathLoader struct{ Paths []string }
type JSONFileLoader struct{ Paths []string }
type SQLiteLoader struct{ Paths []string }
type CommandJSONLoader struct{ Commands []string }
type GlazedCommandLoader struct{ Binaries []string }
```

### New helper functions

```go
func normalizeStringList(values []string) []string
func buildLoaders(s *ServeSettings) []loader.ContentLoader
func DecodeSectionsJSON(r io.Reader) ([]*model.Section, error)
func parseSectionType(typeRaw, sectionTypeRaw json.RawMessage) (model.SectionType, error)
func runGlazedHelpExport(ctx context.Context, binary string) ([]byte, error)
```

---

## File References

| File | Role |
|------|------|
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/serve.go` | `ServeCommand`; add flags, settings, loader orchestration. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/loader/paths.go` | Existing markdown loading; reuse in `MarkdownPathLoader`. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/loader/sources.go` | Proposed new file for `ContentLoader` and external source loaders. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/model/section.go` | Section model and `SectionType`; parser must convert strings/ints. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/store/store.go` | Store APIs used for SQLite loading and merging. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/cmd/export.go` | Current JSON row shape produced by `help export --output json`. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/pkg/help/server/handlers.go` | HTTP API remains unchanged once store is populated. |
| `/home/manuel/code/wesen/corporate-headquarters/glazed/cmd/glaze/main.go` | Root wiring that preloads embedded docs and installs `serve`. |

---

## Appendix: Full User Workflows

### Browse help from three Glazed binaries

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

This is the main target workflow. Users provide binary names, not full commands.

### Browse one Glazed binary without embedded Glaze docs

```bash
glaze serve --from-glazed-cmd pinocchio
```

### Use a custom export command

```bash
glaze serve --from-cmd "pinocchio help export --output json --topic database"
```

### Serve an archived SQLite snapshot

```bash
pinocchio help export --format sqlite --output-path ./pinocchio-help.db
glaze serve --from-sqlite ./pinocchio-help.db
```

### Serve local markdown overrides plus live command docs

```bash
glaze serve \
  --from-glazed-cmd pinocchio,sqleton \
  ./company-overrides/help
```

### Pipe filtered JSON from stdin

```bash
glaze help export --output json \
  | jq '[.[] | select(.type == "Example")]' \
  | glaze serve --from-json -
```

---

## Conclusion

`--from-glazed-cmd` is the missing ergonomic layer that makes the export/serve loop feel like one feature. Users can now say:

```bash
glaze serve --from-glazed-cmd pinocchio,sqleton,xxx
```

and get a single browser for several Glazed-based tools. The general loaders (`--from-json`, `--from-sqlite`, `--from-cmd`, and markdown paths) keep the design flexible, while `--from-glazed-cmd` makes the common case easy.

The implementation should be modest: add list-valued source flags to `ServeCommand`, introduce a `ContentLoader` interface, add robust JSON import for the real export row shape, and run Glazed binaries directly with `exec.CommandContext(binary, "help", "export", "--output", "json")`. The existing HTTP API and React frontend do not need to change.
