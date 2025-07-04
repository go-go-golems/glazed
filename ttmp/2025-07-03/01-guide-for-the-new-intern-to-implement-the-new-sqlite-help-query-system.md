## 1. Project Overview

Welcome! This project powers the help/documentation system for our CLI and related tools. The goal is to make help topics, examples, tutorials, and applications easily searchable and filterable. We are moving from an in-memory Go slice-based system to a SQLite-backed system for better performance, full-text search, and future extensibility.

## 2. Goals and Requirements

- **Performance:** Fast lookups, even with hundreds or thousands of help sections.
- **Expressiveness:** Support complex queries (AND/OR/NOT, by type, topic, flag, etc.).
- **Full-text search:** Allow searching in titles, content, etc.
- **Maintainability:** Clean, testable, and extensible code.
- **CLI and programmatic access:** Should work for both command-line and internal API use.

The new system must support:
- Filtering by section type (GeneralTopic, Example, Application, Tutorial)
- Filtering by topic, flag, command, slug
- Filtering by "shown by default", "top level", etc.
- Full-text search in content, title, etc.
- Boolean logic (AND, OR, NOT) in queries

## 3. Architecture Overview

- **Model:** Data structures for help sections (see `model.Section`, `model.SectionType`).
- **Store:** The SQLite-backed storage and query engine (`store.Store`), in memory (so db must be shared across helpsystems or repopulated each time).
- **Query:** Predicate-based query language (see below).
- **Render:** Converts query results to markdown/terminal output (not changing much).
- **Glue:** Integration with Cobra CLI, etc.

**Interaction Flow:**
1. Markdown files are loaded and parsed into `Section` structs.
2. Sections are inserted/updated in SQLite.
3. Queries are built using Go predicates and compiled to SQL.
4. Results are rendered for the user.

## 5. Query Language Specification

### Predicate-based DSL (in Go)

- Queries are built from composable predicates (functions).
- Each predicate appends SQL fragments and arguments to a query compiler.
- Predicates can be combined with `And`, `Or`, and `Not`.

#### Example Predicates

```go
func IsType(t model.SectionType) Predicate
func HasTopic(topic string) Predicate
func HasFlag(flag string) Predicate
func HasCommand(cmd string) Predicate
func IsTopLevel() Predicate
func ShownByDefault() Predicate
func SlugEquals(slug string) Predicate
func TextSearch(term string) Predicate // FTS5
```

#### Boolean Combinators

```go
func And(preds ...Predicate) Predicate
func Or(preds ...Predicate) Predicate
func Not(pred Predicate) Predicate
```

#### Example Query

```go
q := And(
    Or(IsType(SectionTutorial), IsType(SectionExample)),
    HasTopic("foo"),
    IsTopLevel(),
)
secs, err := hs.Find(q)
```

## 11. Query Compiler Design & Example

### Purpose

The query compiler is responsible for taking a tree of predicates (built using the predicate DSL) and turning it into a single SQL statement with the correct JOINs, WHERE clauses, and arguments. This allows us to express complex queries in Go and have them efficiently executed by SQLite.

### Example: From Predicate Tree to SQL

Suppose you write:

```go
q := And(
    Or(IsType(SectionTutorial), IsType(SectionExample)),
    HasTopic("foo"),
    IsTopLevel(),
)
```

The compiler will produce:

- **SQL:**

```sql
... FILL IN YOUR OWN SQL
```

- **Args:** `["Tutorial", "Example", "foo"]`

## 6. API Design

### Main Interfaces

```go
type Predicate func(*compiler)

type Store struct {
    db *sql.DB
}

func (s *Store) Find(ctx context.Context, pred Predicate) ([]*model.Section, error)
```

### Example Usage

```go
q := And(HasTopic("templates"), IsType(SectionExample))
results, err := store.Find(ctx, q)
```

## 7. Loading and Syncing Data

- Walk the markdown directory tree (from embed.FS)
- For each file:
  - Parse front-matter and content into a `Section` struct.
  - Insert or update the section in SQLite (`INSERT OR REPLACE`).
  - Insert topics, flags, commands into their tables.
