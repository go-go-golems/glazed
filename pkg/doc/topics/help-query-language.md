---
Title: Help Query Language
Slug: help-query-language
SubTitle: Complete guide to the SQLite-backed help query system
Short: Query language for help sections
SectionType: GeneralTopic
Topics: [help, query, dsl, sqlite]
Flags: [help]
Commands: [help]
IsTopLevel: true
ShowPerDefault: true
Order: 100
---

# Help Query Language

This document provides a comprehensive guide to the new SQLite-backed help query system, including the predicate-based Domain Specific Language (DSL) for building complex queries.

## Overview

The help query system replaces the previous in-memory slice-based approach with a powerful SQLite backend that supports:

- **Complex Boolean Logic**: AND, OR, NOT operations with proper precedence
- **Full-Text Search**: FTS5-powered content searching
- **Relational Queries**: Join-based filtering on topics, flags, and commands
- **Type Safety**: Compile-time validated predicates
- **SQL Generation**: Automatic conversion to optimized SQL queries

## Quick Start

```go
import (
    "context"
    "github.com/go-go-golems/glazed/pkg/help/store"
    "github.com/go-go-golems/glazed/pkg/help/query"
    "github.com/go-go-golems/glazed/pkg/help/model"
)

// Create store
helpStore, err := store.NewStore("help.db")
if err != nil {
    log.Fatal(err)
}
defer helpStore.Close()

// Simple query: Find all examples
examples, err := helpStore.Find(ctx, query.IsType(model.SectionExample))

// Complex query: Find top-level tutorials about "getting-started"
results, err := helpStore.Find(ctx, query.And(
    query.IsType(model.SectionTutorial),
    query.HasTopic("getting-started"),
    query.IsTopLevel(),
))
```

## Basic Predicates

### Type Filtering

```go
// Find sections by type
query.IsType(model.SectionExample)      // Examples
query.IsType(model.SectionTutorial)     // Tutorials  
query.IsType(model.SectionApplication)  // Applications
query.IsType(model.SectionGeneralTopic) // General topics
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s WHERE s.sectionType = ? ORDER BY s.ord
-- Args: ["Example"]
```

### Metadata Filtering

```go
// Filter by visibility and organization
query.IsTopLevel()      // Sections shown at top level
query.ShownByDefault()  // Sections shown by default
query.SlugEquals("getting-started") // Specific section by slug
```

### Relationship Filtering

```go
// Filter by associated topics, flags, or commands
query.HasTopic("authentication")   // Sections tagged with topic
query.HasFlag("verbose")          // Sections mentioning flag
query.HasCommand("deploy")        // Sections about command
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
JOIN section_topics st ON st.section_id = s.id 
WHERE st.topic = ? ORDER BY s.ord
-- Args: ["authentication"]
```

### Full-Text Search

```go
// Search in content using FTS5
query.TextSearch("database connection")
query.TextSearch("\"exact phrase\"")
query.TextSearch("deploy OR build")
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
JOIN section_fts fts ON fts.rowid = s.id 
WHERE section_fts MATCH ? ORDER BY s.ord
-- Args: ["database connection"]
```

## Boolean Combinators

### AND Logic

```go
// All conditions must be true
query.And(
    query.IsType(model.SectionExample),
    query.HasTopic("api"),
    query.IsTopLevel(),
)
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
JOIN section_topics st ON st.section_id = s.id 
WHERE s.sectionType = ? AND st.topic = ? AND s.isTopLevel = 1 ORDER BY s.ord
-- Args: ["Example", "api"]
```

### OR Logic

```go
// Any condition can be true
query.Or(
    query.IsType(model.SectionExample),
    query.IsType(model.SectionTutorial),
)
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
WHERE (s.sectionType = ? OR s.sectionType = ?) ORDER BY s.ord
-- Args: ["Example", "Tutorial"]
```

### NOT Logic

```go
// Exclude matching sections
query.Not(query.IsTopLevel())
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
WHERE NOT (s.isTopLevel = 1) ORDER BY s.ord
```

## Advanced Query Patterns

### Nested Boolean Logic

```go
// Complex nested conditions
query.And(
    query.Or(
        query.IsType(model.SectionExample),
        query.IsType(model.SectionTutorial),
    ),
    query.Or(
        query.HasTopic("getting-started"),
        query.HasTopic("advanced"),
    ),
    query.Not(query.IsTopLevel()),
)
```

**Generated SQL:**
```sql
SELECT DISTINCT s.* FROM sections s 
JOIN section_topics st ON st.section_id = s.id 
JOIN section_topics st2 ON st2.section_id = s.id 
WHERE (s.sectionType = ? OR s.sectionType = ?) 
  AND (st.topic = ? OR st2.topic = ?) 
  AND NOT (s.isTopLevel = 1) ORDER BY s.ord
-- Args: ["Example", "Tutorial", "getting-started", "advanced"]
```

### Multiple Topic Filtering

```go
// Sections that have ALL specified topics
query.And(
    query.HasTopic("docker"),
    query.HasTopic("deployment"),
    query.HasTopic("production"),
)
```

### Content and Metadata Combined

```go
// Combine full-text search with metadata filtering
query.And(
    query.TextSearch("kubernetes"),
    query.IsType(model.SectionTutorial),
    query.ShownByDefault(),
)
```

### Exclusion Patterns

```go
// Find beginner content (not advanced)
query.And(
    query.HasTopic("tutorial"),
    query.Not(query.HasTopic("advanced")),
    query.ShownByDefault(),
)
```

## Implementation Details

### Query Compilation Process

1. **Predicate Construction**: Build predicate tree using DSL functions
2. **Compiler Initialization**: Create compiler with alias tracking
3. **Tree Traversal**: Execute predicates to build SQL fragments
4. **Alias Generation**: Generate unique table aliases (`st`, `st2`, `sf`, etc.)
5. **SQL Assembly**: Combine fragments into final query
6. **Execution**: Run against SQLite with bound parameters

### Alias Management

The compiler automatically generates unique aliases to prevent conflicts:

```go
// This query uses multiple topic joins
query.And(
    query.HasTopic("foo"),  // Uses alias 'st'
    query.HasTopic("bar"),  // Uses alias 'st2'
    query.HasFlag("verbose"), // Uses alias 'sf'
)
```

### Join Optimization

The system optimizes JOINs by:
- Sharing JOINs between predicates when possible
- Using unique aliases when multiple joins to same table are needed
- Avoiding duplicate JOIN clauses through deduplication

### Performance Characteristics

- **Index Usage**: All relationship queries use proper indexes
- **FTS5 Performance**: Full-text search leverages SQLite's FTS5 engine
- **Query Planning**: SQLite's query planner optimizes generated SQL
- **Parameter Binding**: All queries use bound parameters for safety

## Extension Guide

### Adding New Predicates

1. **Define the Predicate Function**:
```go
func HasAuthor(author string) Predicate {
    return func(c *compiler) {
        alias := c.getUniqueAlias("sa")
        c.addJoin(fmt.Sprintf("JOIN section_authors %s ON %s.section_id = s.id", alias, alias))
        c.addWhere(fmt.Sprintf("%s.author = ?", alias), author)
    }
}
```

2. **Add Supporting Schema**:
```sql
CREATE TABLE section_authors (
    section_id INTEGER,
    author     TEXT,
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE
);
CREATE INDEX idx_authors_author ON section_authors(author);
```

3. **Update Section Model**:
```go
type Section struct {
    // ... existing fields
    Authors []string
}
```

4. **Add to UpsertSection Logic**:
```go
if err := s.insertRelations(ctx, tx, "section_authors", "author", sectionID, section.Authors); err != nil {
    return err
}
```

### Adding Custom Boolean Logic

```go
func AtLeastOneOf(preds ...Predicate) Predicate {
    return Or(preds...)  // Simple delegation to existing OR
}

func ExactlyOneOf(preds ...Predicate) Predicate {
    // More complex logic for mutual exclusion
    return func(c *compiler) {
        // Implementation for XOR logic
    }
}
```

## Database Schema

### Core Tables

```sql
-- Main sections table
CREATE TABLE sections (
    id          INTEGER PRIMARY KEY,
    slug        TEXT UNIQUE NOT NULL,
    title       TEXT,
    subtitle    TEXT,
    short       TEXT,
    content     TEXT,
    sectionType TEXT,
    isTopLevel  BOOLEAN,
    isTemplate  BOOLEAN,
    showDefault BOOLEAN,
    ord         INTEGER
);

-- Relationship tables
CREATE TABLE section_topics   (section_id INTEGER, topic   TEXT);
CREATE TABLE section_flags    (section_id INTEGER, flag    TEXT);
CREATE TABLE section_commands (section_id INTEGER, command TEXT);

-- Full-text search
CREATE VIRTUAL TABLE section_fts USING fts5(
    slug, title, subtitle, short, content, content='sections'
);
```

### Indexes

```sql
CREATE INDEX idx_topics_topic       ON section_topics(topic);
CREATE INDEX idx_flags_flag         ON section_flags(flag);
CREATE INDEX idx_commands_command   ON section_commands(command);
CREATE INDEX idx_sections_type      ON sections(sectionType);
CREATE INDEX idx_sections_toplevel  ON sections(isTopLevel);
CREATE INDEX idx_sections_slug      ON sections(slug);
```

## Best Practices

### Query Construction

1. **Start Simple**: Begin with basic predicates and add complexity gradually
2. **Use AND Sparingly**: Each AND condition narrows results significantly
3. **Leverage OR for Flexibility**: Use OR to broaden search scope
4. **Combine Text and Metadata**: Mix full-text search with structured filtering

### Performance Optimization

1. **Index-Friendly Queries**: Structure queries to use available indexes
2. **Limit Result Sets**: Use specific predicates to reduce result size
3. **Avoid Complex NOT**: NOT operations can be expensive
4. **Profile Query Performance**: Use EXPLAIN QUERY PLAN for optimization

### Error Handling

```go
results, err := helpStore.Find(ctx, query)
if err != nil {
    // Check for SQL syntax errors
    if strings.Contains(err.Error(), "SQL logic error") {
        log.Error().Err(err).Msg("Invalid query generated")
        // Handle malformed query
    }
    // Handle other database errors
    return err
}
```

## Common Patterns

### Finding Related Content

```go
// Find content related to a specific section
func findRelatedContent(store *Store, section *model.Section) ([]*model.Section, error) {
    var predicates []query.Predicate
    
    // Same type
    if len(section.Topics) > 0 {
        topicPreds := make([]query.Predicate, len(section.Topics))
        for i, topic := range section.Topics {
            topicPreds[i] = query.HasTopic(topic)
        }
        predicates = append(predicates, query.Or(topicPreds...))
    }
    
    // Exclude self
    predicates = append(predicates, query.Not(query.SlugEquals(section.Slug)))
    
    return store.Find(ctx, query.And(predicates...))
}
```

### Building Search Interfaces

```go
// Multi-faceted search
func searchHelp(store *Store, searchParams SearchParams) ([]*model.Section, error) {
    var predicates []query.Predicate
    
    // Text search
    if searchParams.Query != "" {
        predicates = append(predicates, query.TextSearch(searchParams.Query))
    }
    
    // Type filter
    if len(searchParams.Types) > 0 {
        typePreds := make([]query.Predicate, len(searchParams.Types))
        for i, t := range searchParams.Types {
            typePreds[i] = query.IsType(t)
        }
        predicates = append(predicates, query.Or(typePreds...))
    }
    
    // Topic filter
    if len(searchParams.Topics) > 0 {
        for _, topic := range searchParams.Topics {
            predicates = append(predicates, query.HasTopic(topic))
        }
    }
    
    return store.Find(ctx, query.And(predicates...))
}
```

## Troubleshooting

### Common Errors

1. **"ambiguous column name"**: Multiple joins with same alias
   - **Fix**: Ensure unique alias generation is working
   - **Debug**: Print generated SQL to identify conflicts

2. **"no such table"**: Schema not created properly
   - **Fix**: Verify database initialization
   - **Debug**: Check table existence with `.schema` in sqlite3

3. **"constraint failed"**: Foreign key violations
   - **Fix**: Ensure proper transaction handling in UpsertSection
   - **Debug**: Check foreign key constraints with `PRAGMA foreign_keys`

### Debugging Queries

```go
// Print generated SQL for debugging
sql, args := query.Compile(predicate)
fmt.Printf("SQL: %s\n", sql)
fmt.Printf("Args: %v\n", args)

// Test query compilation without execution
results, err := store.Find(ctx, predicate)
```

### Performance Analysis

```sql
-- Analyze query performance
EXPLAIN QUERY PLAN SELECT DISTINCT s.* FROM sections s 
JOIN section_topics st ON st.section_id = s.id 
WHERE st.topic = ? ORDER BY s.ord;

-- Check index usage
.indices sections
.indices section_topics
```

## Migration Guide

### From Legacy System

If migrating from the old slice-based system:

1. **Update Import Paths**:
```go
// Old
import "github.com/go-go-golems/glazed/pkg/help"

// New
import (
    "github.com/go-go-golems/glazed/pkg/help/store"
    "github.com/go-go-golems/glazed/pkg/help/query"
    "github.com/go-go-golems/glazed/pkg/help/model"
)
```

2. **Replace Query Logic**:
```go
// Old
sections := helpSystem.FindSections(query.NewSectionQuery().
    ReturnExamples().
    ReturnOnlyTopics("api"))

// New
sections, err := helpStore.Find(ctx, query.And(
    query.IsType(model.SectionExample),
    query.HasTopic("api"),
))
```

3. **Update Initialization**:
```go
// Old
helpSystem := help.NewHelpSystem()
helpSystem.LoadSectionsFromFS(fs, "doc")

// New
helpStore, err := store.NewStore("help.db")
if err != nil { /* handle error */ }
err = helpStore.LoadSectionsFromFS(ctx, fs, "doc")
```

This query language provides a powerful, type-safe, and performant way to search and filter help content. The predicate-based approach makes complex queries readable and maintainable while the SQLite backend ensures excellent performance even with large help databases.
