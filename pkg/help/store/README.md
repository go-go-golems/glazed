# SQLite Help Query System

This package implements a SQLite-based help query system for the glazed CLI, providing fast, expressive, and scalable help section management with full-text search capabilities.

## Features

- **Fast Performance**: SQLite backend with proper indexing for quick lookups
- **Expressive Queries**: Predicate-based DSL with boolean logic (AND, OR, NOT)
- **Full-Text Search**: FTS5-powered search across titles, content, and metadata
- **Flexible Filtering**: Filter by section type, topic, flag, command, and more
- **Maintainable Code**: Clean separation of concerns with testable components

## Architecture

### Core Components

- **Store**: SQLite backend for data persistence and querying
- **HelpSystem**: High-level interface for help section management
- **Query Compiler**: Converts predicate trees to optimized SQL
- **Predicates**: Composable query building blocks

### Database Schema

```sql
-- Main sections table
sections(id, slug, section_type, title, sub_title, short, content, 
         is_top_level, is_template, show_per_default, order_index, 
         created_at, updated_at)

-- Normalized relationship tables
topics(id, name)
flags(id, name) 
commands(id, name)

-- Junction tables
section_topics(section_id, topic_id)
section_flags(section_id, flag_id)
section_commands(section_id, command_id)

-- FTS5 virtual table for full-text search
sections_fts(slug, title, sub_title, short, content)
```

## Usage

### Basic Setup

```go
import "github.com/go-go-golems/glazed/pkg/help/store"

// Create an in-memory help system
hs, err := store.NewInMemoryHelpSystem()
if err != nil {
    log.Fatal(err)
}
defer hs.Close()

// Or create a file-based system
hs, err := store.NewHelpSystem("/path/to/help.db")
```

### Adding Sections

```go
section := &help.Section{
    Slug:        "example-section",
    SectionType: help.SectionExample,
    Title:       "Example Section",
    Content:     "This shows how to use our CLI",
    Topics:      []string{"examples", "cli"},
    Flags:       []string{"--verbose", "--output"},
    Commands:    []string{"run", "test"},
    IsTopLevel:  true,
    ShowPerDefault: true,
}

err := hs.AddSection(ctx, section)
```

### Querying with Predicates

```go
// Simple queries
examples, err := hs.Find(ctx, store.IsExample())
topLevel, err := hs.Find(ctx, store.IsTopLevel())
dbSections, err := hs.Find(ctx, store.HasTopic("database"))

// Complex boolean queries
results, err := hs.Find(ctx, store.And(
    store.Or(
        store.IsExample(),
        store.IsTutorial(),
    ),
    store.HasTopic("database"),
    store.IsTopLevel(),
))

// Full-text search
searchResults, err := hs.Find(ctx, store.TextSearch("authentication"))

// Combining search with filters
filtered, err := hs.Find(ctx, store.And(
    store.TextSearch("database"),
    store.IsExample(),
    store.ShownByDefault(),
))
```

### Available Predicates

#### Type Filters
- `IsType(sectionType)` - Filter by section type
- `IsExample()` - Example sections only
- `IsTutorial()` - Tutorial sections only  
- `IsApplication()` - Application sections only
- `IsGeneralTopic()` - General topic sections only

#### Relationship Filters
- `HasTopic(topic)` - Sections with specific topic
- `HasFlag(flag)` - Sections with specific flag
- `HasCommand(command)` - Sections with specific command

#### Metadata Filters
- `IsTopLevel()` - Top-level sections
- `ShownByDefault()` - Sections shown by default
- `NotShownByDefault()` - Sections not shown by default
- `IsTemplate()` - Template sections
- `SlugEquals(slug)` - Exact slug match

#### Search
- `TextSearch(term)` - Full-text search using FTS5

#### Boolean Combinators
- `And(predicates...)` - All predicates must match
- `Or(predicates...)` - Any predicate must match
- `Not(predicate)` - Negate a predicate

## Build Requirements

To use FTS5 full-text search, build with the `sqlite_fts5` tag:

```bash
go build -tags sqlite_fts5 ./...
go test -tags sqlite_fts5 ./...
```

## Example Query SQL Generation

The predicate DSL generates optimized SQL:

```go
// Query
q := And(
    Or(IsExample(), IsTutorial()),
    HasTopic("database"),
    IsTopLevel(),
)

// Generated SQL
SELECT DISTINCT s.id, s.slug, s.section_type, s.title, s.sub_title, s.short, s.content,
    s.is_top_level, s.is_template, s.show_per_default, s.order_index, s.created_at, s.updated_at 
FROM sections s 
JOIN section_topics st ON s.id = st.section_id 
JOIN topics t ON st.topic_id = t.id 
WHERE (((s.section_type = ?) OR (s.section_type = ?)) AND (t.name = ?) AND (s.is_top_level = ?)) 
ORDER BY s.order_index ASC, s.title ASC

// Arguments: ["Example", "Tutorial", "database", true]
```

## Migration from In-Memory System

The new SQLite system is designed to be a drop-in replacement for the existing in-memory help system:

1. Replace `help.NewHelpSystem()` with `store.NewInMemoryHelpSystem()`
2. Update query calls to use the new predicate-based API
3. Add context parameters to method calls
4. Build with `-tags sqlite_fts5` for full-text search

## Performance Benefits

- **Indexed Lookups**: Fast retrieval even with thousands of sections
- **Efficient Joins**: Optimized SQL for complex relationship queries
- **Memory Efficient**: SQLite handles large datasets without loading everything into memory
- **Full-Text Search**: FTS5 provides fast, relevance-ranked search results

## Testing

Run tests with FTS5 support:

```bash
go test -tags sqlite_fts5 ./pkg/help/store/...
```

The test suite includes comprehensive coverage of:
- Basic CRUD operations
- Complex predicate queries
- Boolean logic combinations
- Full-text search functionality
- Error handling and edge cases
