# SQLite Help System

This package implements a SQLite-backed help system for the Glazed CLI and related tools. It replaces the previous in-memory slice-based system with a performant, searchable, and extensible SQLite database.

## Features

- **Performance**: Fast lookups even with hundreds or thousands of help sections
- **Full-text search**: Advanced search capabilities using SQLite FTS5 (optional)
- **Predicate-based querying**: Composable query language for complex filtering
- **Boolean logic**: Support for AND, OR, and NOT operations
- **Flexible filtering**: Filter by section type, topic, flag, command, slug, etc.
- **Backward compatibility**: Drop-in replacement for existing help system
- **Build tag support**: Optional FTS5 features via `sqlite_fts5` build tag

## Architecture

The system consists of several key components:

- **Model**: Data structures for help sections (`model.Section`, `model.SectionType`)
- **Store**: SQLite-backed storage with CRUD operations
- **Query**: Predicate-based query language with SQL compilation
- **Loader**: Markdown file parsing and data synchronization
- **Compat**: Compatibility layer for existing help system interface

## Quick Start

### Basic Usage

```go
package main

import (
    "log"
    "github.com/go-go-golems/glazed/pkg/help/store"
    "github.com/go-go-golems/glazed/pkg/help/model"
)

func main() {
    // Create an in-memory help system
    hs, err := store.NewInMemoryHelpSystem()
    if err != nil {
        log.Fatal(err)
    }
    defer hs.Close()

    // Add a section
    section := &model.Section{
        Slug:        "getting-started",
        Title:       "Getting Started",
        SectionType: model.SectionGeneralTopic,
        Content:     "This section covers the basics...",
        Topics:      []string{"basics", "introduction"},
        IsTopLevel:  true,
    }

    if err := hs.AddSection(section); err != nil {
        log.Fatal(err)
    }

    // Query sections
    topLevel, err := hs.Find(store.IsTopLevel())
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d top-level sections", len(topLevel))
}
```

### Predicate-based Queries

The system uses a powerful predicate-based query language:

```go
// Simple queries
examples := hs.Find(store.IsExample())
topLevel := hs.Find(store.IsTopLevel())
withTopic := hs.Find(store.HasTopic("configuration"))

// Boolean combinations
complexQuery := hs.Find(store.And(
    store.Or(store.IsExample(), store.IsTutorial()),
    store.HasTopic("basics"),
    store.ShownByDefault(),
))

// Text search (uses FTS5 if available, LIKE fallback otherwise)
searchResults := hs.Find(store.TextSearch("hello world"))

// Ordering and pagination
ordered := hs.Find(store.And(
    store.IsExample(),
    store.OrderByOrder(),
    store.Limit(10),
    store.Offset(20),
))
```

### Available Predicates

#### Type Filters
- `IsType(sectionType)` - Filter by section type
- `IsExample()` - Filter for examples
- `IsTutorial()` - Filter for tutorials
- `IsGeneralTopic()` - Filter for general topics
- `IsApplication()` - Filter for applications

#### Content Filters
- `HasTopic(topic)` - Sections with specific topic
- `HasFlag(flag)` - Sections with specific flag
- `HasCommand(command)` - Sections with specific command
- `SlugEquals(slug)` - Exact slug match
- `SlugIn(slugs)` - Slug in list
- `TitleContains(term)` - Title contains term
- `ContentContains(term)` - Content contains term
- `TextSearch(term)` - Full-text search

#### Metadata Filters
- `IsTopLevel()` - Top-level sections
- `ShownByDefault()` / `NotShownByDefault()` - Default display
- `IsTemplate()` - Template sections

#### Ordering and Pagination
- `OrderByOrder()` - Sort by order field
- `OrderByTitle()` - Sort by title
- `OrderByCreatedAt()` - Sort by creation time
- `Limit(n)` - Limit results
- `Offset(n)` - Skip results

#### Boolean Combinators
- `And(predicates...)` - All predicates must match
- `Or(predicates...)` - Any predicate must match
- `Not(predicate)` - Predicate must not match

## FTS5 Support

Full-text search is available when building with the `sqlite_fts5` tag:

```bash
# Build with FTS5 support
go build -tags sqlite_fts5

# Test with FTS5 support
go test -tags sqlite_fts5
```

Without the tag, `TextSearch()` falls back to LIKE queries.

## Loading Data

### From Filesystem

```go
import "embed"

//go:embed help/**/*.md
var helpFS embed.FS

// Load all markdown files from embedded filesystem
err := hs.LoadSectionsFromFS(helpFS, "help")

// Or sync (clears existing data first)
err := hs.SyncFromFS(helpFS, "help")
```

### Markdown Format

Sections are defined in markdown files with YAML frontmatter:

```markdown
---
Slug: getting-started
Title: Getting Started Guide
SectionType: Tutorial
Topics: [basics, setup]
Commands: [init, start]
Flags: [--verbose, --help]
IsTopLevel: true
ShowPerDefault: true
Order: 1
---

# Getting Started

This is the content of the help section...
```

## Database Schema

The system creates the following tables:

- `sections` - Main sections table with all metadata
- `sections_fts` - FTS5 virtual table (when enabled)

Indexes are automatically created for performance on common query patterns.

## Compatibility Layer

The `HelpSystem` type provides backward compatibility:

```go
// Drop-in replacement for existing help system
hs, err := store.NewInMemoryHelpSystem()

// Use existing interface methods
section, err := hs.GetSectionWithSlug("example")
sections, err := hs.GetSections()
examples, err := hs.GetExamplesForTopic("basics")
```

## Performance

The SQLite-backed system provides significant performance improvements:

- Fast indexed lookups for common queries
- Full-text search with ranking
- Efficient pagination for large result sets
- Minimal memory usage (data stored on disk)

## Testing

```bash
# Run all tests
go test ./pkg/help/store/...

# Run with FTS5
go test -tags sqlite_fts5 ./pkg/help/store/...

# Run examples
go test -run Example ./pkg/help/store/...
```

## Migration from Existing System

The new system is designed as a drop-in replacement:

1. Replace `help.NewHelpSystem()` with `store.NewInMemoryHelpSystem()`
2. Use existing methods or migrate to predicate-based queries
3. Optionally enable FTS5 for enhanced search capabilities

The predicate system is more powerful than the existing `SectionQuery`, but the compatibility layer ensures existing code continues to work.
