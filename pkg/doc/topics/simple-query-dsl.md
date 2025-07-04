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

The glazed help system supports a simple Domain Specific Language (DSL) for querying help sections. This allows users to quickly find relevant documentation using intuitive search expressions.

## Basic Syntax

### Field Queries

Query specific fields using the `field:value` syntax:

```
type:example          # Find example sections
topic:database        # Find sections about database
flag:--output         # Find sections mentioning --output flag
command:json          # Find sections for json command
```

### Field Types

- **type:** - Section type (`example`, `tutorial`, `topic`, `application`)
- **topic:** - Topic tags assigned to sections
- **flag:** - Flag names (with or without --)
- **command:** - Command names
- **slug:** - Exact slug match

### Metadata Queries

Query section metadata:

```
toplevel:true         # Only top-level sections
default:true          # Only sections shown by default
template:true         # Only template sections
```

### Text Search

Use quotes for full-text search:

```
"SQLite database"     # Search content for these words
"full text search"    # Multi-word phrases
```

## Boolean Operations

Combine queries using boolean operators:

### AND Operation
```
type:example AND topic:database
flag:--output AND command:json
"SQLite" AND type:tutorial
```

### OR Operation
```
type:example OR type:tutorial
topic:database OR topic:sqlite
command:json OR command:yaml
```

### NOT Operation
```
NOT type:application
type:example AND NOT topic:advanced
"database" AND NOT type:tutorial
```

### Grouping with Parentheses
```
(type:example OR type:tutorial) AND topic:database
type:example AND (topic:database OR topic:sqlite)
(toplevel:true AND default:true) OR type:example
```

## Shortcuts

Common queries have convenient shortcuts:

```
examples              # Equivalent to type:example
tutorials             # Equivalent to type:tutorial
topics                # Equivalent to type:topic
applications          # Equivalent to type:application
toplevel              # Equivalent to toplevel:true
defaults              # Equivalent to default:true
```

## Query Examples

### Find Database Examples
```
examples AND topic:database
type:example AND "SQLite"
```

### Find All Documentation for JSON Command
```
command:json
command:json OR "json command"
```

### Find Beginner-Friendly Content
```
defaults AND (examples OR tutorials)
toplevel AND default:true
```

### Advanced Searches
```
(type:example OR type:tutorial) AND (topic:database OR topic:sql) AND NOT "advanced"
flag:--output AND (command:json OR command:yaml)
```

### Content-Based Searches
```
"full text search" AND type:tutorial
"performance" OR "optimization"
"error handling" AND examples
```

## Case Sensitivity

- Field names are case-insensitive: `Type:Example` = `type:example`
- Values are case-insensitive: `topic:Database` = `topic:database`
- Text search is case-insensitive: `"SQLite"` matches "sqlite"
- Boolean operators can be uppercase or lowercase: `AND` = `and`

## Special Characters

- Quotes: Use `"` for multi-word text searches
- Colons: Used to separate field from value (`field:value`)
- Parentheses: Group expressions `(expr1 OR expr2)`
- Spaces: Separate tokens and operators

## Error Handling

Invalid queries will show helpful error messages:

```
type:invalid          # Error: Unknown section type 'invalid'
field:value          # Error: Unknown field 'field'
unclosed "quote      # Error: Unclosed quote
```

## Usage in CLI

The query DSL can be used with the help system:

```bash
# Search using the DSL
glaze help --query "examples AND topic:database"
glaze help --query "type:tutorial OR type:example"
glaze help --query 'flag:--output AND "JSON"'

# Combine with other flags
glaze help --query "defaults" --short
glaze help --query "toplevel AND examples" --all
```

## Implementation Notes

The DSL is implemented as a parser that converts user queries into the internal predicate system. This provides a user-friendly interface while maintaining the power and flexibility of the underlying SQLite-based query engine.

The parser supports:
- Lexical analysis with token recognition
- Recursive descent parsing for boolean expressions
- Error recovery and meaningful error messages
- Query optimization and validation
