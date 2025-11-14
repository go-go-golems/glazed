---
Slug: simple-query-dsl
Title: Simple Query DSL for Help System
SectionType: GeneralTopic
Topics:
  - help-system
  - query
  - search
  - dsl
IsTopLevel: true
ShowPerDefault: true
Order: 15
---

# Simple Query DSL for Help System

The Glazed help system supports a simple Domain Specific Language (DSL) for querying help sections. This page is a **quick-start reference** for authors and users who just need the essentials. For the full language reference (including operator precedence tables, metadata fields, and troubleshooting tips) run `glaze help user-query-dsl`.

## Quick Reference

| Task | Syntax | Example |
| ---- | ------ | ------- |
| Filter by section type | `type:<value>` or shortcut word | `type:example`, `tutorials` |
| Filter by metadata | `toplevel:true`, `default:false`, `template:true` | `toplevel:true` |
| Match commands/flags | `command:<name>`, `flag:--<flag>` | `command:json`, `flag:--output` |
| Search topics/tags | `topic:<tag>` | `topic:database` |
| Text search | `"quoted phrase"` | `"SQLite database"` |
| Combine expressions | `AND`, `OR`, `NOT`, parentheses | `examples AND topic:database`, `(tutorials OR examples) AND NOT topic:advanced` |

Shortcuts (`examples`, `tutorials`, `topics`, `applications`, `toplevel`, `defaults`) are just readable aliases for the equivalent `field:value` queries.

## Core Patterns

### Field Filters
```
type:tutorial
topic:templates
command:json
flag:--output
slug:help-system
```

Mix multiple filters with boolean operators to narrow results:
```
type:example AND topic:database
(examples OR tutorials) AND command:json
flag:--output AND NOT topic:advanced
```

### Metadata Filters
```
toplevel:true        # Only sections shown on the root help screen
default:true         # Sections displayed without --all
template:true        # Template sections
```

### Text Search
```
"SQLite database"
"error handling" AND tutorials
performance OR "throughput tuning"
```

## Practical Examples

```bash
# Show beginner-friendly getting-started docs
glaze help --query "defaults AND tutorials"

# Find every example that mentions SQLite or SQL performance tips
glaze help --query "(examples) AND (\"SQLite\" OR topic:sql) AND NOT topic:advanced"

# List all sections that talk about the json command
glaze help --query "command:json"

# Search for CLI flags across docs
glaze help --query 'flag:--output AND "table"'
```

## CLI Usage

```bash
glaze help --query "examples AND topic:database"
glaze help --query "(tutorials OR examples) AND toplevel:true"
glaze help --query 'flag:--output AND "JSON"' --short
```

Combine `--query` with the standard help flags:
- `--short` for compact summaries
- `--all` to list every matching section
- `--print-query` / `--print-sql` (on the `help` command itself) when debugging complex expressions

## Next Steps

- Need every operator, metadata field, and error message documented? See `glaze help user-query-dsl`.
- Want to run queries from your own code or service? Check out `glaze help using-the-query-api`.
