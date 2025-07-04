---
Title: Help Search Syntax
Slug: help-search-syntax
SubTitle: Complete guide to the help system query language
Short: Advanced search syntax for finding help sections efficiently
SectionType: GeneralTopic
Topics: [help, search, query, syntax]
Flags: [help]
Commands: [help]
IsTopLevel: true
ShowPerDefault: true
Order: 10
---

The glazed help system provides a powerful query language that allows you to quickly find relevant documentation using simple text-based queries. This guide covers the complete syntax and provides examples for effective searching.

## Quick Start

The simplest way to search is by typing keywords:

```bash
glaze help search "docker deployment"
```

For more specific results, use field filters:

```bash
glaze help search "type:example topic:api"
```

## Basic Syntax

### Text Search

Search for text in titles and content:

```
docker
"exact phrase"
getting started
```

- **Unquoted words**: Match anywhere in title or content
- **Quoted phrases**: Match exact phrase
- **Multiple words**: Implicitly combined with AND

### Field Filters

Filter by specific attributes using `field:value` syntax:

```
type:example
topic:api
flag:verbose
command:deploy
toplevel:true
default:false
slug:getting-started
```

#### Supported Fields

| Field | Description | Example Values |
|-------|-------------|----------------|
| `type` | Section type | `example`, `tutorial`, `application`, `topic` |
| `topic` | Topic tags | `api`, `getting-started`, `deployment` |
| `flag` | Command flags | `verbose`, `debug`, `output` |
| `command` | Associated commands | `deploy`, `build`, `run` |
| `toplevel` | Top-level sections | `true`, `false` |
| `default` | Default visibility | `true`, `false` |
| `slug` | Exact slug match | `getting-started`, `api-tutorial` |
| `title` | Title text search | Any text |
| `content` | Content text search | Any text |

#### Field Operators

- `:` - Exact match (default)
- `=` - Exact match (alternative)
- `~` - Contains match

```
type:example        # Exact type match
topic~getting       # Topic contains "getting"
title="API Guide"   # Exact title match
```

#### Supported Types

- `topic` or `generaltopic` - General documentation topics
- `example` - Code examples and samples
- `application` or `app` - Full application guides
- `tutorial` or `tut` - Step-by-step tutorials

#### Boolean Values

For `toplevel` and `default` fields:

- **True**: `true`, `yes`, `1`, `on`
- **False**: `false`, `no`, `0`, `off`

## Advanced Syntax

### Negation

Use `-` or `NOT` to exclude results:

```
-type:tutorial              # Exclude tutorials
NOT flag:debug             # Exclude debug-related content
-"advanced configuration"   # Exclude specific phrases
```

### Boolean Logic

Combine filters with logical operators:

```
type:example AND topic:api
type:example OR type:tutorial
NOT type:application
```

#### Operator Precedence

1. `NOT` (highest)
2. `AND`
3. `OR` (lowest)

### Grouping

Use parentheses to control evaluation order:

```
(type:example OR type:tutorial) AND topic:getting-started
NOT (type:tutorial AND flag:advanced)
```

### Complex Queries

Combine all features for powerful searches:

```
(type:example OR type:tutorial) AND topic:api AND NOT flag:deprecated
docker AND (type:example OR type:tutorial) AND -flag:verbose
"kubernetes deployment" type:tutorial -topic:advanced
```

## Query Examples

### Finding Examples

```bash
# All examples
type:example

# API examples
type:example topic:api

# Docker-related examples
type:example docker

# Examples without debug flags
type:example -flag:debug
```

### Finding Tutorials

```bash
# Getting started tutorials
type:tutorial topic:getting-started

# Beginner tutorials (no advanced topics)
type:tutorial -topic:advanced

# Short tutorials
type:tutorial "quick start"
```

### Finding by Commands

```bash
# Documentation for deploy command
command:deploy

# Help for build or deploy commands
command:build OR command:deploy

# Advanced deployment guides
command:deploy topic:advanced
```

### Finding by Flags

```bash
# Documentation about verbose flag
flag:verbose

# Output-related flags
flag:output OR flag:format

# All flags except debug
flag:* -flag:debug
```

### Complex Searches

```bash
# Docker examples for beginners
docker type:example -topic:advanced

# API documentation (any type)
topic:api (type:example OR type:tutorial OR type:topic)

# Top-level getting started content
toplevel:true topic:getting-started

# Hidden advanced tutorials
type:tutorial topic:advanced default:false
```

## Search Tips

### Effective Strategies

1. **Start broad, then narrow**: Begin with general terms, then add filters
2. **Use type filters early**: Specify `type:` to focus on relevant content
3. **Combine text and filters**: Mix keywords with field filters for precision
4. **Use negation wisely**: Exclude irrelevant content with `-` or `NOT`

### Common Patterns

```bash
# Find beginner content
topic:getting-started OR topic:beginner OR "quick start"

# Find troubleshooting help
"troubleshooting" OR "error" OR "problem" OR "fix"

# Find configuration guides
"configuration" OR "config" OR "setup" OR "install"

# Find API documentation
topic:api OR "API" OR "endpoint" OR "request"
```

### Performance Tips

- Use specific type filters to reduce search scope
- Prefer quoted phrases over multiple separate words
- Use field filters instead of text search when possible
- Avoid overly complex boolean expressions

## Error Messages

### Common Syntax Errors

```
Error: Invalid field name 'typo'
Fix: Use valid field names (type, topic, flag, command, etc.)

Error: Unterminated string at line 1
Fix: Close quoted strings with matching quotes

Error: Expected ':' after field name
Fix: Use proper field:value syntax
```

### Validation Errors

```
Error: Empty value for field 'type'
Fix: Provide a value after the colon (type:example)

Error: Invalid section type 'unknown'
Fix: Use valid types (example, tutorial, application, topic)

Error: Invalid boolean value 'maybe'
Fix: Use true/false, yes/no, 1/0, on/off
```

## Integration Examples

### Command Line Usage

```bash
# Search from command line
glaze help search "type:example docker"

# List all tutorials
glaze help search "type:tutorial" --list

# Find API examples with output
glaze help search "type:example topic:api flag:output"
```

### Programmatic Usage

```go
import "github.com/go-go-golems/glazed/pkg/help/integration"

// Create search service
service := integration.NewSearchService(store)

// Simple search
sections, err := service.Search(ctx, "type:example docker")

// Search with metadata
result, err := service.SearchWithMetadata(ctx, 
    "type:tutorial topic:getting-started", 
    integration.SearchOptions{Limit: 10})

// Validate query
err := service.ValidateQuery("type:example AND topic:api")

// Build query programmatically
sections, err := service.ExecuteBuilder(ctx,
    service.BuildQuery().
        WithType(model.SectionExample).
        WithTopic("api").
        WithTextSearch("docker"))
```

## Troubleshooting

### Query Not Finding Results

1. **Check spelling**: Verify field names and values
2. **Simplify query**: Remove complex boolean logic
3. **Use broader terms**: Try less specific keywords
4. **Check available content**: Use `--list` to see all sections

### Unexpected Results

1. **Use quotes**: Wrap exact phrases in quotes
2. **Add field filters**: Narrow scope with `type:` or `topic:`
3. **Check negation**: Ensure `-` syntax is correct
4. **Verify precedence**: Use parentheses for clarity

### Performance Issues

1. **Add type filters**: Reduce search scope early
2. **Avoid wildcards**: Use specific terms instead
3. **Limit results**: Use pagination for large result sets
4. **Cache queries**: Reuse compiled predicates when possible

## Advanced Features

### Query Optimization

The system automatically optimizes queries for better performance:

- Removes unnecessary grouping
- Flattens nested boolean operations
- Reorders predicates for optimal execution

### Full-Text Search

Text searches use SQLite FTS (Full-Text Search) for fast results:

- Automatic stemming and ranking
- Support for phrase queries
- Boolean operators within text searches

### Extensibility

The query system is designed for extension:

- Add custom field types
- Implement custom operators
- Create domain-specific query builders

## Reference

### Complete Field List

- `type` - Section type filter
- `topic` - Topic tag filter  
- `flag` - Command flag filter
- `command` - Command name filter
- `toplevel` - Top-level section filter
- `default` - Default visibility filter
- `slug` - Exact slug match
- `title` - Title text search
- `content` - Content text search

### Type Values

- `topic`, `generaltopic` - General topics
- `example` - Examples
- `application`, `app` - Applications  
- `tutorial`, `tut` - Tutorials

### Boolean Values

- **True**: `true`, `yes`, `1`, `on`
- **False**: `false`, `no`, `0`, `off`

### Operators

- `:` - Field assignment (exact match)
- `=` - Field assignment (alternative)
- `~` - Field assignment (contains)
- `-` - Negation prefix
- `AND` - Logical AND
- `OR` - Logical OR
- `NOT` - Logical NOT
- `()` - Grouping

### Reserved Keywords

- `AND`, `OR`, `NOT` - Boolean operators
- `true`, `false` - Boolean values
- `yes`, `no` - Boolean values
- `on`, `off` - Boolean values

---

This query language provides powerful and intuitive search capabilities for the glazed help system. Start with simple queries and gradually incorporate advanced features as needed.
