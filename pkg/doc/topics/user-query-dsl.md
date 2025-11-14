---
Slug: user-query-dsl
Title: Complete User Query DSL Reference
SectionType: GeneralTopic
Topics:
  - help-system
  - query
  - search
  - dsl
  - boolean
IsTopLevel: true
ShowPerDefault: true
Order: 10
---

# Complete User Query DSL Reference

The glazed help system features a powerful Domain Specific Language (DSL) for querying documentation. This DSL allows users to construct sophisticated queries using field filters, boolean logic, and text search to quickly find relevant help content.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Syntax](#basic-syntax)
3. [Boolean Operations](#boolean-operations)
4. [Advanced Features](#advanced-features)
5. [Query Examples](#query-examples)
6. [Debugging Tools](#debugging-tools)
7. [Performance Guide](#performance-guide)
8. [Error Reference](#error-reference)
9. [Complete API Reference](#complete-api-reference)

## Introduction

### What is the Query DSL?

The Query DSL is a user-friendly query language that allows you to search through help documentation using structured queries instead of browsing through lists. It supports:

- **Field-based filtering**: Find sections by type, topic, command, or flag
- **Boolean logic**: Combine conditions with AND, OR, and NOT
- **Text search**: Full-text search through content, titles, and descriptions
- **Metadata queries**: Filter by section properties like top-level status
- **Grouping**: Use parentheses to control query precedence

### Why Use the Query DSL?

- **Speed**: Find relevant documentation instantly without browsing
- **Precision**: Narrow down results with specific criteria
- **Flexibility**: Combine multiple conditions to find exactly what you need
- **Discoverability**: Learn about features through targeted searches

> Looking for a lightweight cheat sheet? Run `glaze help simple-query-dsl`. That page highlights the most common patterns, while this document dives into every operator, metadata field, and failure mode.

## Basic Syntax

### Simple Shortcuts

Common queries have convenient shortcut words:

```bash
# Section type shortcuts
glaze help --query "examples"      # All example sections
glaze help --query "tutorials"     # All tutorial sections  
glaze help --query "topics"        # All general topic sections
glaze help --query "applications"  # All application sections

# Metadata shortcuts
glaze help --query "toplevel"      # Top-level sections only
glaze help --query "defaults"      # Sections shown by default
```

### Field:Value Syntax

Query specific fields using the `field:value` format:

```bash
# Section types
glaze help --query "type:example"
glaze help --query "type:tutorial"
glaze help --query "type:topic"
glaze help --query "type:application"

# Content filters
glaze help --query "topic:database"
glaze help --query "command:json"
glaze help --query "flag:--output"
glaze help --query "slug:help-system"

# Metadata filters
glaze help --query "toplevel:true"
glaze help --query "default:false"
```

### Text Search

Search through content using quoted strings:

```bash
# Simple text search
glaze help --query "\"SQLite database\""
glaze help --query "\"error handling\""

# Single word search (quotes optional)
glaze help --query "performance"
glaze help --query "\"optimization\""
```

## Boolean Operations

### Operator Precedence

The DSL follows standard boolean logic precedence:
1. **NOT** (highest precedence)
2. **AND** (medium precedence)  
3. **OR** (lowest precedence)

Use parentheses to override default precedence.

### AND Operations

Find sections that match ALL conditions:

```bash
# Basic AND
glaze help --query "type:example AND topic:database"
glaze help --query "examples AND topic:templates"

# Multiple conditions
glaze help --query "type:tutorial AND command:json AND toplevel:true"
```

### OR Operations

Find sections that match ANY condition:

```bash
# Basic OR
glaze help --query "type:example OR type:tutorial"
glaze help --query "examples OR tutorials"

# Multiple alternatives
glaze help --query "topic:database OR topic:sql OR topic:sqlite"
```

### NOT Operations

Exclude sections that match a condition:

```bash
# Basic NOT
glaze help --query "NOT type:application"
glaze help --query "NOT toplevel:true"

# NOT with other operations
glaze help --query "examples AND NOT topic:advanced"
glaze help --query "type:tutorial OR NOT default:true"
```

### Parentheses for Grouping

Control evaluation order with parentheses:

```bash
# Without parentheses (AND has higher precedence than OR)
glaze help --query "examples AND topic:database OR tutorials"
# Equivalent to: (examples AND topic:database) OR tutorials

# With parentheses (explicit grouping)
glaze help --query "examples AND (topic:database OR topic:sql)"
glaze help --query "(examples OR tutorials) AND topic:templates"

# Complex grouping
glaze help --query "(type:example OR type:tutorial) AND (topic:database OR topic:sql) AND NOT toplevel:false"
```

## Advanced Features

### Multiple Values

Query multiple values for the same field (OR semantics):

```bash
# Multiple types
glaze help --query "type:example,tutorial"
# Equivalent to: type:example OR type:tutorial

# Multiple topics
glaze help --query "topic:database,sql,sqlite"
# Equivalent to: topic:database OR topic:sql OR topic:sqlite

# Multiple commands
glaze help --query "command:json,yaml,csv"
```

### Metadata Queries

Filter by section metadata properties:

```bash
# Boolean metadata
glaze help --query "toplevel:true"          # Top-level sections
glaze help --query "default:false"          # Non-default sections
glaze help --query "template:true"          # Template sections

# Combining metadata
glaze help --query "toplevel:true AND default:true"
glaze help --query "type:example AND template:false"
```

### Case Sensitivity

All queries are case-insensitive:

```bash
# These are all equivalent
glaze help --query "Type:Example"
glaze help --query "type:example"
glaze help --query "TYPE:EXAMPLE"

# Boolean operators are also case-insensitive
glaze help --query "examples AND tutorials"
glaze help --query "examples and tutorials"
glaze help --query "EXAMPLES AND TUTORIALS"
```

### Flag Normalization

Flags are automatically normalized:

```bash
# These all find the same flag
glaze help --query "flag:output"
glaze help --query "flag:-output"
glaze help --query "flag:--output"
```

## Query Examples

### Finding Examples and Tutorials

```bash
# All examples about templates
glaze help --query "type:example AND topic:templates"

# Examples or tutorials about databases
glaze help --query "(examples OR tutorials) AND topic:database"

# Examples that aren't shown by default
glaze help --query "examples AND default:false"
```

### Command-Specific Help

```bash
# All documentation for the json command
glaze help --query "command:json"

# Examples specifically for json command
glaze help --query "type:example AND command:json"

# Help for json OR yaml commands
glaze help --query "command:json OR command:yaml"
```

### Flag Documentation

```bash
# All sections mentioning --output flag
glaze help --query "flag:--output"

# Examples using specific flags
glaze help --query "type:example AND (flag:--output OR flag:--format)"

# Flag documentation for json command
glaze help --query "command:json AND flag:--output"
```

### Content Discovery

```bash
# Find performance-related content
glaze help --query "\"performance\" OR \"optimization\""

# Database-related tutorials and examples
glaze help --query "(tutorials OR examples) AND (topic:database OR \"SQLite\")"

# Advanced features (not shown by default)
glaze help --query "default:false AND (topic:advanced OR \"advanced\")"
```

### Troubleshooting Queries

```bash
# Error handling documentation
glaze help --query "\"error\" OR \"handling\" OR \"troubleshoot\""

# Debugging information
glaze help --query "\"debug\" OR \"troubleshoot\" OR \"problem\""

# Configuration help
glaze help --query "\"config\" OR \"configuration\" OR \"setup\""
```

## Debugging Tools

When a query behaves unexpectedly, surface the parser's view of the expression before guessing.

```bash
glaze help --query "(examples OR tutorials) AND topic:database" \
  --print-query \
  --print-sql
```

- `--print-query` dumps the normalized AST so you can check operator precedence, shortcut expansion, and field normalization.
- `--print-sql` (when the SQLite store is active) prints the generated SQL, which is useful when profiling or when you implement a custom store backend.

For programmatic debugging, the DSL package exposes the same building blocks:

```go
ast, err := dsl.ParseToAST(query)
if err != nil {
    log.Fatal(err)
}

info, err := dsl.GetDebugInfo(query) // includes SQL, used fields, detected shortcuts
fmt.Println(info.SQL)
```

> Tip: Help output now targets stdout. If you prefer to keep debug output on stderr, call `help_cmd.SetHelpWriter(os.Stderr)` once during startup and both Glamour styling and debug streams will follow that writer.

## Performance Guide

### Query Optimization Tips

1. **Use specific fields first**: `type:example AND topic:database` is faster than text search
2. **Limit text searches**: Use field filters before text search when possible
3. **Use shortcuts**: `examples` is faster than `type:example`
4. **Combine efficiently**: Use AND to narrow results, OR to expand them

### Best Practices

```bash
# Good: Specific field filter first
glaze help --query "type:example AND \"SQLite\""

# Less efficient: Text search first  
glaze help --query "\"example\" AND topic:database"

# Good: Use shortcuts when possible
glaze help --query "examples AND topic:templates"

# Verbose: Explicit field queries
glaze help --query "type:example AND topic:templates"
```

### Query Complexity

- Simple field queries: Near-instant
- Boolean combinations: Very fast
- Text search: Fast (full-text indexed)
- Complex nested queries: Fast (optimized boolean evaluation)

## Error Reference

### Common Syntax Errors

#### Invalid Field Names
```bash
# Error
glaze help --query "invalid:value"
# Fix: Use valid fields: type, topic, flag, command, slug
```

#### Invalid Boolean Syntax
```bash
# Error: Missing operand
glaze help --query "examples AND"
# Fix: Complete the expression
glaze help --query "examples AND tutorials"

# Error: Invalid operator
glaze help --query "examples XOR tutorials" 
# Fix: Use valid operators: AND, OR, NOT
```

#### Unmatched Parentheses
```bash
# Error: Unclosed parenthesis
glaze help --query "(examples AND tutorials"
# Fix: Close all parentheses
glaze help --query "(examples AND tutorials)"
```

#### Invalid Values
```bash
# Error: Unknown section type
glaze help --query "type:invalid"
# Fix: Use valid types: example, tutorial, topic, application

# Error: Boolean field with non-boolean value
glaze help --query "toplevel:maybe"
# Fix: Use boolean values: true, false
```

### Error Recovery Tips

1. **Check field names**: Use `type`, `topic`, `flag`, `command`, `slug`
2. **Verify boolean values**: Use `true` or `false` for metadata fields
3. **Balance parentheses**: Ensure every `(` has a matching `)`
4. **Complete expressions**: Don't end with operators like `AND`

## Complete API Reference

### Valid Fields

| Field | Description | Values | Examples |
|-------|-------------|---------|-----------|
| `type` | Section type | `example`, `tutorial`, `topic`, `application` | `type:example` |
| `topic` | Topic tags | Any topic name | `topic:database` |
| `flag` | Flag names | Flag names (with/without dashes) | `flag:--output` |
| `command` | Command names | Command names | `command:json` |
| `slug` | Section slug | Exact slug match | `slug:help-system` |
| `toplevel` | Top-level status | `true`, `false` | `toplevel:true` |
| `default` | Default display | `true`, `false` | `default:false` |
| `template` | Template status | `true`, `false` | `template:true` |

### Boolean Operators

| Operator | Description | Precedence | Examples |
|----------|-------------|------------|-----------|
| `NOT` | Logical negation | 1 (highest) | `NOT type:application` |
| `AND` | Logical conjunction | 2 (medium) | `examples AND topic:database` |
| `OR` | Logical disjunction | 3 (lowest) | `examples OR tutorials` |

### Shortcuts

| Shortcut | Equivalent | Description |
|----------|------------|-------------|
| `examples` | `type:example` | All example sections |
| `tutorials` | `type:tutorial` | All tutorial sections |
| `topics` | `type:topic` | All general topic sections |
| `applications` | `type:application` | All application sections |
| `toplevel` | `toplevel:true` | Top-level sections |
| `defaults` | `default:true` | Default sections |

### Special Characters

| Character | Purpose | Example |
|-----------|---------|---------|
| `:` | Field separator | `type:example` |
| `"` | Text search delimiter | `"search text"` |
| `(` `)` | Grouping | `(examples OR tutorials)` |
| `,` | Multiple values | `type:example,tutorial` |

### Query Limits

- **Maximum query length**: 1000 characters
- **Maximum nested parentheses**: 10 levels
- **Maximum field values**: 20 per field
- **Result limit**: 1000 sections (pagination available)

## Usage Examples

### CLI Integration

```bash
# Basic usage
glaze help --query "examples"

# With other help flags
glaze help --query "tutorials" --short
glaze help --query "type:example" --all

# Combining with traditional flags
glaze help json --query "type:example"
```

### Common Workflows

```bash
# Learning workflow
glaze help --query "tutorials AND toplevel:true"     # Start here
glaze help --query "examples AND default:true"       # Try examples  
glaze help --query "topic:advanced AND NOT default:true"  # Advanced topics

# Problem-solving workflow
glaze help --query "\"error\" OR \"troubleshoot\""   # Find error docs
glaze help --query "command:json AND \"problem\""    # Command-specific issues
glaze help --query "flag:--debug OR \"debugging\""   # Debug information

# Feature discovery
glaze help --query "default:false"                   # Hidden features
glaze help --query "type:application"                # Real-world examples
glaze help --query "topic:advanced"                  # Advanced techniques
```

The Query DSL provides a powerful and intuitive way to navigate the glazed help system. Start with simple queries and gradually incorporate boolean logic and advanced features as you become more comfortable with the syntax.
