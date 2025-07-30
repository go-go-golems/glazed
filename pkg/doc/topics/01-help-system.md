---
Title: Help System
Slug: help-system
Short: Glazed provides a powerful, queryable help system for creating rich CLI documentation with sections, metadata, and programmatic access.
Topics:
- help
- documentation
- cli
- sections
- query
Commands:
- help
Flags:
- flag
- topic
- command
- list
- topics
- examples
- applications
- tutorials
- help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Help System

## Overview

The Glazed help system provides a structured, queryable approach to CLI documentation that goes beyond basic command help. It organizes documentation into typed sections (topics, examples, applications, tutorials) with rich metadata for filtering and discovery. The system supports both human-readable help pages and programmatic querying through a simple DSL, making it easy to build comprehensive CLI documentation that users can explore efficiently.

The help system stores sections in an SQLite-backed store, enabling fast queries, text search, and metadata filtering. You typically load documentation from markdown files with YAML frontmatter at startup, creating a self-contained help database that you can query from both the command line and Go code.

## Section Types and Structure

Help sections follow a type-based classification system that separates conceptual documentation from practical examples. This separation enables precise filtering and contextual help display, allowing users to find exactly the type of information they need based on their current task.

### Section Types

- **GeneralTopic**: Conceptual documentation explaining how features work
- **Example**: Focused demonstrations of specific command usage
- **Application**: Real-world use cases combining multiple features
- **Tutorial**: Step-by-step guides for complex workflows

### Section Structure

Each section contains:

```go
type Section struct {
    // Core content
    Slug     string      // Unique identifier for referencing
    Title    string      // Display title
    Short    string      // Brief description
    Content  string      // Full markdown content
    
    // Section classification
    SectionType SectionType // GeneralTopic, Example, etc.
    
    // Searchable metadata
    Topics   []string    // Related topic tags
    Commands []string    // Relevant CLI commands
    Flags    []string    // Associated command flags
    
    // Display behavior
    IsTopLevel     bool // Show in main help listing
    ShowPerDefault bool // Include without --all flag
    Order          int  // Sort order within type
}
```

## Programmatic Usage

The help system exposes a programming model centered around the `HelpSystem` struct, which manages an SQLite-backed documentation store. You can initialize the system, load documentation from various sources, and query sections using both simple lookups and complex DSL expressions.

### Initializing the Help System

```go
// Create new help system with in-memory storage
hs := help.NewHelpSystem()

// Load documentation from embedded filesystem
//go:embed docs
var docsFS embed.FS
err := hs.LoadSectionsFromFS(docsFS, "docs")
if err != nil {
    log.Fatal(err)
}
```

### Loading Documentation

You typically load documentation from markdown files with YAML frontmatter:

```go
// Load from filesystem recursively
err := hs.LoadSectionsFromFS(filesystem, "documentation")

// Load individual section from markdown
markdownBytes := []byte(`---
Title: Example Command
Slug: json-example
SectionType: Example
Topics: [json, formatting]
Commands: [json]
---

# JSON Output Example

Use the json command to format data:
...`)

section, err := help.LoadSectionFromMarkdown(markdownBytes)
if err == nil {
    hs.AddSection(section)
}
```

## Query System and DSL

The help system treats documentation as structured data that can be queried using boolean logic and metadata filters. This approach transforms static help text into a searchable knowledge base where users can find relevant documentation by combining criteria like section type, topics, and command associations.

### Basic DSL Queries

```go
// Query by section type
examples, err := hs.QuerySections("type:example")

// Find sections about databases
dbSections, err := hs.QuerySections("topic:database")

// Search for specific commands
jsonHelp, err := hs.QuerySections("command:json")

// Full-text search
searchResults, err := hs.QuerySections(`"SQLite integration"`)
```

### Boolean Logic and Complex Queries

```go
// AND operations - sections that are both examples AND about databases
results, err := hs.QuerySections("type:example AND topic:database")

// OR operations - either examples or tutorials
results, err := hs.QuerySections("type:example OR type:tutorial")

// NOT operations - exclude advanced topics
results, err := hs.QuerySections("type:example AND NOT topic:advanced")

// Grouping with parentheses
results, err := hs.QuerySections("(type:example OR type:tutorial) AND topic:database")
```

### Metadata Queries

```go
// Query section display properties
topLevel, err := hs.QuerySections("toplevel:true")
defaults, err := hs.QuerySections("default:true")

// Flag and command associations
flagHelp, err := hs.QuerySections("flag:--output")
cmdHelp, err := hs.QuerySections("command:json OR command:yaml")
```

## Individual Section Retrieval

When you know the exact section you need, you can retrieve it directly using its unique slug identifier. This approach bypasses the query system for immediate access to specific documentation sections.

```go
// Get specific section by slug
section, err := hs.GetSectionWithSlug("help-system")
if err != nil {
    if err == help.ErrSectionNotFound {
        // Handle missing section
    }
    return err
}

// Access section content and metadata
fmt.Printf("Title: %s\n", section.Title)
fmt.Printf("Type: %s\n", section.SectionType.String())
fmt.Printf("Topics: %v\n", section.Topics)
```

## Integration with Cobra Commands

The help system extends Cobra's built-in help functionality by automatically displaying relevant documentation sections when users request help for specific commands. This integration creates contextual help that shows both command syntax and related educational content.

### Basic Cobra Integration

```go
func SetupHelpSystem(rootCmd *cobra.Command, hs *help.HelpSystem) {
    // Add help command with query support
    helpCmd := &cobra.Command{
        Use:   "help [topic]",
        Short: "Help about any command or topic",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) == 0 {
                // Show top-level help page
                page := hs.GetTopLevelHelpPage()
                fmt.Print(page.Render())
                return
            }
            
            // Look up specific section
            section, err := hs.GetSectionWithSlug(args[0])
            if err != nil {
                fmt.Printf("Help topic '%s' not found\n", args[0])
                return
            }
            
            fmt.Print(section.Content)
        },
    }
    
    rootCmd.AddCommand(helpCmd)
}
```

### Enhanced Command Help

```go
// Augment command help with related sections
func AugmentCommandHelp(cmd *cobra.Command, hs *help.HelpSystem) {
    originalUsageFunc := cmd.UsageFunc()
    
    cmd.SetUsageFunc(func(c *cobra.Command) error {
        // Show standard command help
        if err := originalUsageFunc(c); err != nil {
            return err
        }
        
        // Find related help sections
        query := fmt.Sprintf("command:%s AND default:true", c.Name())
        sections, err := hs.QuerySections(query)
        if err != nil || len(sections) == 0 {
            return nil
        }
        
        // Display related sections
        fmt.Println("\n## Related Documentation")
        for _, section := range sections {
            fmt.Printf("  %s - %s\n", section.Slug, section.Short)
            fmt.Printf("    glaze help %s\n", section.Slug)
        }
        
        return nil
    })
}
```

## Working with Section Metadata

Section metadata transforms help content into a rich data structure that you can filter and organize based on user context. By leveraging metadata fields like topics, commands, and flags, you can build intelligent help systems that surface the most relevant documentation for any given situation.

### Filtering by Metadata

```go
// Get sections for specific contexts
func GetSectionsForCommand(hs *help.HelpSystem, commandName string) []*help.Section {
    sections, _ := hs.QuerySections(fmt.Sprintf("command:%s", commandName))
    return sections
}

func GetExamplesForTopic(hs *help.HelpSystem, topic string) []*help.Section {
    query := fmt.Sprintf("type:example AND topic:%s", topic)
    sections, _ := hs.QuerySections(query)
    return sections
}

func GetDefaultSections(hs *help.HelpSystem) []*help.Section {
    sections, _ := hs.QuerySections("default:true")
    return sections
}
```

### Dynamic Section Discovery

```go
// Build contextual help based on current command and flags
func BuildContextualHelp(hs *help.HelpSystem, cmdName string, flags []string) {
    var queries []string
    
    // Add command-specific sections
    queries = append(queries, fmt.Sprintf("command:%s", cmdName))
    
    // Add flag-specific sections
    for _, flag := range flags {
        queries = append(queries, fmt.Sprintf("flag:%s", flag))
    }
    
    // Combine with OR logic
    query := strings.Join(queries, " OR ")
    sections, err := hs.QuerySections(query)
    if err != nil {
        return
    }
    
    // Display relevant sections
    for _, section := range sections {
        if section.ShowPerDefault {
            fmt.Printf("📖 %s\n", section.Short)
        }
    }
}
```

## Advanced Features

When building complex documentation systems, you need visibility into how queries are parsed and executed. The help system includes debugging tools that show query AST generation and SQL translation, helping you understand and optimize query performance.

### Query Debugging

```go
// Debug query parsing and SQL generation
err := hs.PrintQueryDebug("type:example AND topic:database", true, true)
// Outputs:
// Query: type:example AND topic:database
// AST:
//   AND
//   ├── Field: type = "example"
//   └── Field: topic = "database"
// SQL Query:
//   SELECT * FROM sections WHERE section_type = ? AND ? = ANY(topics)
//   Parameters: [1 database]
```

### Performance Considerations

For high-frequency querying scenarios, you can implement caching by storing query results in a map. The help system's SQLite backend is already optimized for typical usage patterns.

For more information about the query DSL syntax and capabilities:

```
glaze help simple-query-dsl
```
