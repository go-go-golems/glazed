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
- **Store:** The SQLite-backed storage and query engine (`store.Store`).
- **Query:** Predicate-based query language (see below).
- **Render:** Converts query results to markdown/terminal output (not changing much).
- **Glue:** Integration with Cobra CLI, etc.

**Interaction Flow:**
1. Markdown files are loaded and parsed into `Section` structs.
2. Sections are inserted/updated in SQLite.
3. Queries are built using Go predicates and compiled to SQL.
4. Results are rendered for the user.

## 4. SQLite Schema

```sql
CREATE TABLE sections (
  id          INTEGER PRIMARY KEY,
  slug        TEXT UNIQUE NOT NULL,
  title       TEXT,
  subtitle    TEXT,
  short       TEXT,
  content     TEXT,
  sectionType TEXT,          -- "GeneralTopic", "Example", ...
  isTopLevel  BOOLEAN,
  isTemplate  BOOLEAN,
  showDefault BOOLEAN,
  ord         INTEGER
);

CREATE TABLE section_topics   (section_id INTEGER, topic   TEXT);
CREATE TABLE section_flags    (section_id INTEGER, flag    TEXT);
CREATE TABLE section_commands (section_id INTEGER, command TEXT);

CREATE VIRTUAL TABLE section_fts USING fts5(
  slug, title, subtitle, short, content, content='sections'
);

CREATE INDEX idx_topics_topic       ON section_topics(topic);
CREATE INDEX idx_flags_flag         ON section_flags(flag);
CREATE INDEX idx_commands_command   ON section_commands(command);
CREATE INDEX idx_sections_type      ON sections(sectionType);
CREATE INDEX idx_sections_toplevel  ON sections(isTopLevel);
```

- **sections:** Main table for all help sections.
- **section_topics/flags/commands:** Many-to-many relationships for topics, flags, and commands.
- **section_fts:** Full-text search index for fast content lookup.
- **Indexes:** Speed up lookups and joins.

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

#### How Predicates Compile to SQL

- Each predicate adds WHERE clauses, JOINs, and arguments.
- The compiler builds a single SQL statement, e.g.:

```sql
SELECT DISTINCT s.*
FROM   sections s
JOIN   section_topics st ON st.section_id = s.id
WHERE  (s.sectionType = ? OR s.sectionType = ?)
  AND  st.topic = ?
  AND  s.isTopLevel = 1
ORDER BY s.ord
-- args: ["Tutorial", "Example", "foo"]
```

#### Adding New Predicates

- Create a new function returning `Predicate`.
- Add necessary JOINs and WHERE clauses.
- See `query/predicate.go` for examples.

## 11. Query Compiler Design & Example

### Purpose

The query compiler is responsible for taking a tree of predicates (built using the predicate DSL) and turning it into a single SQL statement with the correct JOINs, WHERE clauses, and arguments. This allows us to express complex queries in Go and have them efficiently executed by SQLite.

### Main Data Structure

The core of the compiler is a struct that accumulates SQL fragments:

```go
// query/compiler.go
package query

type compiler struct {
    joins  []string   // SQL JOIN clauses
    wheres []string   // SQL WHERE conditions
    args   []any      // Arguments for parameterized queries
}

func (c *compiler) addWhere(cond string, args ...any) {
    c.wheres = append(c.wheres, cond)
    c.args = append(c.args, args...)
}

func (c *compiler) addJoin(join string) {
    c.joins = append(c.joins, join)
}

func (c *compiler) SQL() (string, []any) {
    sql := "SELECT DISTINCT s.* FROM sections s"
    if len(c.joins) > 0 {
        sql += " " + strings.Join(c.joins, " ")
    }
    if len(c.wheres) > 0 {
        sql += " WHERE " + strings.Join(c.wheres, " AND ")
    }
    sql += " ORDER BY s.ord"
    return sql, c.args
}
```

### How Predicates Interact with the Compiler

Each predicate is a function that takes a pointer to a compiler and appends its own SQL logic:

```go
func IsType(t model.SectionType) Predicate {
    return func(c *compiler) {
        c.addWhere("s.sectionType = ?", t.String())
    }
}

func HasTopic(topic string) Predicate {
    return func(c *compiler) {
        c.addJoin("JOIN section_topics st ON st.section_id = s.id")
        c.addWhere("st.topic = ?", topic)
    }
}
```

Boolean combinators like `And`, `Or`, and `Not` recursively build up the query:

```go
func And(preds ...Predicate) Predicate {
    return func(c *compiler) {
        var subWheres []string
        var subJoins []string
        var subArgs []any
        for _, p := range preds {
            sub := &compiler{}
            p(sub)
            subWheres = append(subWheres, strings.Join(sub.wheres, " AND "))
            subJoins = append(subJoins, sub.joins...)
            subArgs = append(subArgs, sub.args...)
        }
        c.joins = append(c.joins, subJoins...)
        c.wheres = append(c.wheres, "("+strings.Join(subWheres, " AND ")+")")
        c.args = append(c.args, subArgs...)
    }
}

func Or(preds ...Predicate) Predicate {
    return func(c *compiler) {
        var subWheres []string
        var subJoins []string
        var subArgs []any
        for _, p := range preds {
            sub := &compiler{}
            p(sub)
            subWheres = append(subWheres, strings.Join(sub.wheres, " AND "))
            subJoins = append(subJoins, sub.joins...)
            subArgs = append(subArgs, sub.args...)
        }
        c.joins = append(c.joins, subJoins...)
        c.wheres = append(c.wheres, "("+strings.Join(subWheres, " OR ")+")")
        c.args = append(c.args, subArgs...)
    }
}

func Not(pred Predicate) Predicate {
    return func(c *compiler) {
        sub := &compiler{}
        pred(sub)
        c.joins = append(c.joins, sub.joins...)
        c.wheres = append(c.wheres, "NOT ("+strings.Join(sub.wheres, " AND ")+")")
        c.args = append(c.args, sub.args...)
    }
}
```

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
SELECT DISTINCT s.*
FROM   sections s
JOIN   section_topics st ON st.section_id = s.id
WHERE  (s.sectionType = ? OR s.sectionType = ?)
  AND  st.topic = ?
  AND  s.isTopLevel = 1
ORDER BY s.ord
```

- **Args:** `["Tutorial", "Example", "foo"]`

### Tips for Extending the Compiler

- **Deduping JOINs:**
  - If multiple predicates add the same JOIN, consider using a map or a helper to deduplicate before building the final SQL.
- **Nested AND/OR/NOT:**
  - Always wrap subclauses in parentheses to ensure correct SQL logic.
  - Each combinator should recursively compile its children.
- **Adding New Predicates:**
  - Just add a new function that appends the necessary JOINs and WHEREs.
  - If you need to join a new table, add the JOIN in the predicate.
- **Debugging:**
  - Print the generated SQL and args when testing new predicates to ensure correctness.

### Pseudocode: Compiler Walk

1. Start with an empty compiler.
2. Call the root predicate with the compiler.
3. Each predicate (leaf or combinator) appends JOINs, WHEREs, and args.
4. After the walk, call `compiler.SQL()` to get the final SQL and args.
5. Execute the query with `db.QueryContext(ctx, sql, args...)`.

This design makes it easy to add new query features and ensures all queries are safe and efficient.

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

- Walk the markdown directory tree.
- For each file:
  - Parse front-matter and content into a `Section` struct.
  - Insert or update the section in SQLite (`INSERT OR REPLACE`).
  - Insert topics, flags, commands into their tables.
- After all inserts, run `INSERT INTO section_fts(section_fts) VALUES ('rebuild');` to update the FTS index.
- For updates, check file timestamps or always upsert for now.

## 8. Testing and Validation

- Write unit tests for each predicate (see `query/predicate_test.go`).
- Write integration tests for the store (see `store/store_test.go`).
- Compare results with the old in-memory system for validation.
- Use golden files for output rendering tests.

## 9. File/Directory References

- **New code should go in:**
  - `glazed/pkg/help/model/` (data structs)
  - `glazed/pkg/help/store/` (SQLite store, loader)
  - `glazed/pkg/help/query/` (predicates, compiler)
- **Existing helpers:**
  - `glazed/pkg/help/` (for reference, will be refactored)
  - `glazed/pkg/helpers/` (string helpers, etc.)

## 10. Next Steps and Resources

- Start by implementing the schema and loading pipeline.
- Implement the core predicates and the compiler.
- Write tests for each part as you go.
- Ask Manuel or the team for help if you get stuck.
- See the README and code comments for more details.
- Useful docs:
  - [SQLite FTS5](https://www.sqlite.org/fts5.html)
  - [Go database/sql](https://pkg.go.dev/database/sql)
  - [Go SQLite drivers](https://github.com/mattn/go-sqlite3, https://pkg.go.dev/modernc.org/sqlite)

Welcome aboard, and happy hacking! 