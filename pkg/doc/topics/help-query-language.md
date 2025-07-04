# Help Query Language: Implementation & Usage Guide

## Overview

The help query language powers advanced, expressive search and filtering for the Glazed CLI help/documentation system. It enables fast, composable queries over help topics, examples, tutorials, and applications, backed by a SQLite database with full-text search (FTS5).

This guide covers:
- Architecture and data flow
- Predicate and combinator API
- SQL generation and extensibility
- Usage patterns and examples
- Testing and debugging tips

---

## 1. Architecture & Data Flow

**Key components:**
- **Model:** `model.Section` struct represents a help/documentation section.
- **Store:** `store.Store` manages the SQLite database and exposes query APIs.
- **Query:** Predicate-based query language (Go functions) compiles to SQL.
- **Loader:** Parses markdown files with YAML front-matter into `Section` structs and syncs them to the DB.
- **Render:** Converts query results to markdown or terminal output.

**Flow:**
1. Markdown files are loaded and parsed into `Section` structs.
2. Sections are upserted into SQLite (with topics, flags, commands, and FTS index).
3. Queries are built using Go predicates and combinators.
4. The query compiler turns predicates into a single SQL statement.
5. Results are rendered for the user.

---

## 2. Data Model

```
type Section struct {
    ID          int64    `yaml:"id,omitempty"`
    Slug        string   `yaml:"slug,omitempty"`
    Title       string   `yaml:"title,omitempty"`
    Subtitle    string   `yaml:"subtitle,omitempty"`
    Short       string   `yaml:"short,omitempty"`
    Content     string   `yaml:"content,omitempty"`
    SectionType SectionType `yaml:"sectionType,omitempty"`
    IsTopLevel  bool     `yaml:"isTopLevel,omitempty"`
    IsTemplate  bool     `yaml:"isTemplate,omitempty"`
    ShowDefault bool     `yaml:"showDefault,omitempty"`
    Ord         int      `yaml:"ord,omitempty"`
    Topics      []string `yaml:"topics,omitempty"`
    Flags       []string `yaml:"flags,omitempty"`
    Commands    []string `yaml:"commands,omitempty"`
}
```

---

## 3. Predicate API

Predicates are Go functions that describe a filter or search condition. They are composable using boolean combinators.

### Core Predicates

```
IsType(t SectionType)           // Filter by section type (e.g., Example, Tutorial)
HasTopic(topic string)          // Section has a given topic
HasFlag(flag string)            // Section has a given flag
HasCommand(cmd string)          // Section has a given command
IsTopLevel()                    // Section is top-level
ShownByDefault()                // Section is shown by default
SlugEquals(slug string)         // Section slug matches
TextSearch(term string)         // Full-text search (FTS5)
```

### Boolean Combinators

```
And(preds ...Predicate)         // All predicates must match
Or(preds ...Predicate)          // Any predicate matches
Not(pred Predicate)             // Negate a predicate
```

### Special NOT Predicates (for correct SQL anti-join)

```
NotHasFlag(flag string)         // Section does NOT have a given flag
NotHasTopic(topic string)       // Section does NOT have a given topic
NotHasCommand(cmd string)       // Section does NOT have a given command
```

---

## 4. SQL Generation

The query compiler walks the predicate tree and generates a single SQL statement with the necessary JOINs, WHERE clauses, and arguments.

**Example:**

```go
q := And(
    Or(IsType(SectionTutorial), IsType(SectionExample)),
    HasTopic("foo"),
    IsTopLevel(),
)
results, err := store.Find(ctx, q)
```

**Generated SQL:**
```
SELECT DISTINCT s.*
FROM   sections s
JOIN   section_topics st1 ON st1.section_id = s.id
WHERE  (s.sectionType = ? OR s.sectionType = ?)
  AND  st1.topic = ?
  AND  s.isTopLevel = 1
ORDER BY s.ord
-- args: ["Tutorial", "Example", "foo"]
```

**NOT/anti-join example:**
```go
q := NotHasFlag("hidden")
```
**SQL:**
```
WHERE NOT EXISTS (SELECT 1 FROM section_flags nf WHERE nf.section_id = s.id AND nf.flag = ?)
```

---

## 5. Usage Patterns & Examples

### Find all tutorials or examples about "templates":
```go
q := And(
    Or(IsType(SectionTutorial), IsType(SectionExample)),
    HasTopic("templates"),
)
results, err := store.Find(ctx, q)
```

### Find all top-level sections not shown by default:
```go
q := And(IsTopLevel(), Not(ShownByDefault()))
```

### Full-text search for "csv import":
```go
q := TextSearch("csv import")
```

### Find all sections that do NOT have the flag "hidden":
```go
q := NotHasFlag("hidden")
```

### Find all examples with both topic "foo" and flag "bar":
```go
q := And(IsType(SectionExample), HasTopic("foo"), HasFlag("bar"))
```

### Find all sections with topic "foo" but NOT flag "bar":
```go
q := And(HasTopic("foo"), NotHasFlag("bar"))
```

---

## 6. Extending the Query Language

To add a new predicate:
1. Create a function returning `Predicate` in `query/predicate.go`.
2. In the function, add necessary JOINs and WHERE clauses to the compiler.
3. For anti-joins, use `NOT EXISTS` subqueries.
4. Add tests in `predicate_test.go` and integration tests in `store/advanced_query_test.go`.

**Example:**
```go
func HasTag(tag string) Predicate {
    return func(c *Compiler) {
        alias := c.nextAlias("tg")
        c.addJoin(fmt.Sprintf("JOIN section_tags %s ON %s.section_id = s.id", alias, alias))
        c.addWhere(fmt.Sprintf("%s.tag = ?", alias), tag)
    }
}
```

---

## 7. Testing & Debugging

- Use the provided test suites (`*_test.go`) to validate new predicates and queries.
- Print/log the generated SQL and args for debugging complex queries.
- Use SQLite tools to inspect the database and run raw queries if needed.
- For FTS5, remember to rebuild the index after bulk inserts:
  ```go
  db.Exec(`INSERT INTO section_fts(section_fts) VALUES ('rebuild')`)
  ```

---

## 8. Error Handling & Edge Cases

- Loader returns clear errors for malformed or missing front-matter.
- Upserts by slug ensure no duplicate sections.
- Empty arrays for topics/flags/commands are handled gracefully.
- All predicates are safe for use in combinators (AND/OR/NOT).

---

## 9. Resources & Further Reading

- [SQLite FTS5 Documentation](https://www.sqlite.org/fts5.html)
- [Go database/sql](https://pkg.go.dev/database/sql)
- [Go SQLite drivers](https://github.com/mattn/go-sqlite3)
- See `glazed/pkg/help/query/predicate.go` and `glazed/pkg/help/store/advanced_query_test.go` for real code examples.

---

## 10. FAQ

**Q: How do I add a new filter or search feature?**
- Add a new predicate function in `predicate.go` and test it.

**Q: How do I debug a query?**
- Print the SQL and args from the compiler before executing.

**Q: How do I ensure FTS5 works?**
- Use the `sqlite_fts5` build tag and rebuild the FTS index after inserts.

---

Happy hacking! For questions, see the code comments or ask the maintainers. 