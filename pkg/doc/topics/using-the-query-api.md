---
Slug: using-the-query-api
Title: Developer Guide - Using the Query API and DSL
SectionType: GeneralTopic
Topics:
  - help-system
  - api
  - query
  - dsl
  - development
IsTopLevel: true
ShowPerDefault: false
Order: 20
---

# Developer Guide: Using the Query API and DSL

This guide provides comprehensive documentation for developers who want to integrate with, extend, or use the glazed help system's query capabilities programmatically.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Query API Basics](#query-api-basics)
3. [DSL Integration](#dsl-integration)
4. [Predicate System](#predicate-system)
5. [Query Compilation](#query-compilation)
6. [Custom Extensions](#custom-extensions)
7. [Performance Optimization](#performance-optimization)
8. [Integration Patterns](#integration-patterns)
9. [Debugging and Troubleshooting](#debugging-and-troubleshooting)
10. [API Reference](#api-reference)

## Architecture Overview

### System Components

The query system consists of several composed components:

```
User Query DSL → Parser → AST → Compiler → Predicates → Execution
     ↓              ↓       ↓        ↓          ↓          ↓
  "examples"    Lexer   Expression  SQL Gen   Filters   Results
```

### Package Structure

```
pkg/help/
├── dsl/                    # DSL parser and compiler
│   ├── lexer.go           # Tokenization
│   ├── parser.go          # AST building
│   ├── compiler.go        # Predicate generation
│   └── dsl.go             # Main API
├── store/                  # Storage backend (optional)
│   ├── query.go           # Store-specific predicates
│   └── store.go           # SQLite implementation
├── dsl_bridge.go          # Integration section
├── help.go                # Core help system
└── cobra.go               # CLI integration
```

### Data Flow

1. **Input**: User provides query string
2. **Parsing**: DSL parser converts to AST
3. **Compilation**: AST converts to predicate functions
4. **Execution**: Predicates filter help sections
5. **Output**: Filtered results returned

## Query API Basics

### Simple Programmatic Usage

```go
package main

import (
    "context"
    "fmt"
    "github.com/go-go-golems/glazed/pkg/help"
)

func main() {
    // Create help system
    hs := help.NewHelpSystem()
    
    // Load documentation
    err := doc.AddDocToHelpSystem(hs)
    if err != nil {
        panic(err)
    }
    
    // Execute simple query
    results, err := hs.QuerySections("examples")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d examples\n", len(results))
    for _, section := range results {
        fmt.Printf("- %s: %s\n", section.Slug, section.Title)
    }
}
```

### Boolean Query Usage

```go
// Complex boolean queries
queries := []string{
    "type:example AND topic:database",
    "examples OR tutorials",
    "NOT type:application",
    "(examples OR tutorials) AND topic:templates",
}

for _, query := range queries {
    results, err := hs.QuerySections(query)
    if err != nil {
        fmt.Printf("Query error: %s\n", err)
        continue
    }
    
    fmt.Printf("Query: %s → %d results\n", query, len(results))
}
```

### Error Handling

```go
results, err := hs.QuerySections("invalid:syntax AND")
if err != nil {
    // Handle different error types
    switch {
    case strings.Contains(err.Error(), "unknown field"):
        fmt.Println("Invalid field name")
    case strings.Contains(err.Error(), "syntax error"):
        fmt.Println("Query syntax error")
    default:
        fmt.Printf("Query error: %s\n", err)
    }
}
```

## DSL Integration

### Using the DSL Parser Directly

```go
import "github.com/go-go-golems/glazed/pkg/help/dsl"

// Parse query to AST
expr, err := dsl.Parse("type:example AND topic:database")
if err != nil {
    return err
}

// Compile to predicate (if using store backend)
predicate, err := dsl.ParseQuery("type:example AND topic:database")
if err != nil {
    return err
}

// Use predicate with store
ctx := context.Background()
results, err := store.Find(ctx, predicate)
```

### Query Information and Validation

```go
// Get DSL information
info := dsl.GetQueryInfo()
fmt.Printf("Valid fields: %v\n", info.ValidFields)
fmt.Printf("Valid types: %v\n", info.ValidTypes)
fmt.Printf("Valid shortcuts: %v\n", info.ValidShortcuts)

// Validate queries
validQueries := []string{
    "examples",
    "type:tutorial",
    "topic:database AND examples",
}

for _, query := range validQueries {
    if err := dsl.ValidateQuery(query); err != nil {
        fmt.Printf("Invalid query '%s': %s\n", query, err)
    } else {
        fmt.Printf("Valid query: %s\n", query)
    }
}
```

### Debug Information

```go
// Get debug information about a query
debugInfo, err := dsl.GetDebugInfo("(examples OR tutorials) AND NOT default:true")
if err != nil {
    return err
}

fmt.Printf("Query: %s\n", debugInfo.Query)
fmt.Printf("AST: %s\n", debugInfo.AST)
fmt.Printf("SQL: %s\n", debugInfo.SQL)
fmt.Printf("Fields: %v\n", debugInfo.Fields)
```

## Predicate System

### Understanding Predicates

Predicates are functions that filter help sections based on criteria:

```go
type Predicate func(*Section) bool

// Example predicate implementations
func isExample(section *Section) bool {
    return section.SectionType == SectionExample
}

func hasTopic(topic string) func(*Section) bool {
    return func(section *Section) bool {
        for _, t := range section.Topics {
            if strings.EqualFold(t, topic) {
                return true
            }
        }
        return false
    }
}
```

### Creating Custom Predicates

```go
// Custom predicate for recent sections
func CreatedAfter(date time.Time) func(*Section) bool {
    return func(section *Section) bool {
        return section.CreatedAt.After(date)
    }
}

// Custom predicate for content length
func MinContentLength(minLength int) func(*Section) bool {
    return func(section *Section) bool {
        return len(section.Content) >= minLength
    }
}

// Usage
recentSections := filterSections(allSections, CreatedAfter(time.Now().AddDate(0, -1, 0)))
longSections := filterSections(allSections, MinContentLength(1000))
```

### Combining Predicates

```go
// Predicate combinators
func And(predicates ...func(*Section) bool) func(*Section) bool {
    return func(section *Section) bool {
        for _, pred := range predicates {
            if !pred(section) {
                return false
            }
        }
        return true
    }
}

func Or(predicates ...func(*Section) bool) func(*Section) bool {
    return func(section *Section) bool {
        for _, pred := range predicates {
            if pred(section) {
                return true
            }
        }
        return false
    }
}

func Not(predicate func(*Section) bool) func(*Section) bool {
    return func(section *Section) bool {
        return !predicate(section)
    }
}

// Usage
complexPredicate := And(
    isExample,
    hasTopic("database"),
    Not(func(s *Section) bool { return s.IsTopLevel }),
)
```

## Query Compilation

### Store Backend Integration

If using the SQLite store backend, queries compile to SQL:

```go
import "github.com/go-go-golems/glazed/pkg/help/store"

// Create store
store, err := store.NewInMemory()
if err != nil {
    return err
}
defer store.Close()

// Load sections
ctx := context.Background()
for _, section := range sections {
    err := store.AddSection(ctx, section)
    if err != nil {
        return err
    }
}

// Query with predicates
predicate := store.And(
    store.IsType("example"),
    store.HasTopic("database"),
    store.Not(store.IsTopLevel()),
)

results, err := store.Find(ctx, predicate)
if err != nil {
    return err
}
```

### Custom Query Compilation

```go
// Implement custom query compiler
type CustomCompiler struct {
    filters []func(*Section) bool
}

func (c *CustomCompiler) AddFilter(filter func(*Section) bool) {
    c.filters = append(c.filters, filter)
}

func (c *CustomCompiler) Execute(sections []*Section) []*Section {
    var results []*Section
    
    for _, section := range sections {
        matches := true
        for _, filter := range c.filters {
            if !filter(section) {
                matches = false
                break
            }
        }
        if matches {
            results = append(results, section)
        }
    }
    
    return results
}

// Usage
compiler := &CustomCompiler{}
compiler.AddFilter(isExample)
compiler.AddFilter(hasTopic("database"))
results := compiler.Execute(allSections)
```

## Custom Extensions

### Adding Custom Fields

```go
// Extend Section with custom fields
type ExtendedSection struct {
    *help.Section
    Priority    int               `json:"priority"`
    Tags        []string          `json:"tags"`
    Metadata    map[string]string `json:"metadata"`
}

// Custom field predicates
func HasPriority(priority int) func(*ExtendedSection) bool {
    return func(section *ExtendedSection) bool {
        return section.Priority >= priority
    }
}

func HasTag(tag string) func(*ExtendedSection) bool {
    return func(section *ExtendedSection) bool {
        for _, t := range section.Tags {
            if strings.EqualFold(t, tag) {
                return true
            }
        }
        return false
    }
}
```

### Custom DSL Extensions

```go
// Extend the DSL parser for custom syntax
type ExtendedParser struct {
    *dsl.Parser
}

func (p *ExtendedParser) parseCustomExpression() (dsl.Expression, error) {
    // Parse custom syntax like "priority:>=5" or "tag:important"
    switch p.currentToken().Type {
    case dsl.PRIORITY:
        return p.parsePriorityExpression()
    case dsl.TAG:
        return p.parseTagExpression()
    default:
        return p.Parser.ParseExpression()
    }
}

// Register custom field handlers
func init() {
    dsl.RegisterField("priority", handlePriorityField)
    dsl.RegisterField("tag", handleTagField)
}
```

### Custom Storage Backends

```go
// Implement custom storage backend
type CustomStore struct {
    sections []*help.Section
    index    map[string][]*help.Section // field -> sections index
}

func (s *CustomStore) Find(ctx context.Context, pred store.Predicate) ([]*help.Section, error) {
    // Custom query execution logic
    var results []*help.Section
    
    for _, section := range s.sections {
        if s.evaluatePredicate(pred, section) {
            results = append(results, section)
        }
    }
    
    return results, nil
}

func (s *CustomStore) AddSection(ctx context.Context, section *help.Section) error {
    s.sections = append(s.sections, section)
    s.updateIndex(section)
    return nil
}
```

## Performance Optimization

### Query Optimization Tips

```go
// 1. Use indexed fields first
// Good: Fast field lookup first
"type:example AND \"complex search\""

// Less optimal: Text search first
"\"complex search\" AND type:example"

// 2. Limit result sets early
// Good: Narrow scope quickly
"type:example AND topic:specific"

// Less optimal: Broad scope
"\"common word\" OR type:example"

// 3. Use shortcuts when possible
// Good: Optimized shortcut
"examples"

// Verbose: Field query
"type:example"
```

### Caching Strategies

```go
type CachedHelpSystem struct {
    *help.HelpSystem
    queryCache map[string][]*help.Section
    cacheMutex sync.RWMutex
}

func (chs *CachedHelpSystem) QuerySections(query string) ([]*help.Section, error) {
    chs.cacheMutex.RLock()
    if results, found := chs.queryCache[query]; found {
        chs.cacheMutex.RUnlock()
        return results, nil
    }
    chs.cacheMutex.RUnlock()
    
    // Execute query
    results, err := chs.HelpSystem.QuerySections(query)
    if err != nil {
        return nil, err
    }
    
    // Cache results
    chs.cacheMutex.Lock()
    chs.queryCache[query] = results
    chs.cacheMutex.Unlock()
    
    return results, nil
}
```

### Indexing for Fast Lookups

```go
type IndexedHelpSystem struct {
    sections []*help.Section
    
    // Pre-built indexes
    typeIndex    map[help.SectionType][]*help.Section
    topicIndex   map[string][]*help.Section
    commandIndex map[string][]*help.Section
    flagIndex    map[string][]*help.Section
}

func (ihs *IndexedHelpSystem) buildIndexes() {
    ihs.typeIndex = make(map[help.SectionType][]*help.Section)
    ihs.topicIndex = make(map[string][]*help.Section)
    ihs.commandIndex = make(map[string][]*help.Section)
    ihs.flagIndex = make(map[string][]*help.Section)
    
    for _, section := range ihs.sections {
        // Index by type
        ihs.typeIndex[section.SectionType] = append(
            ihs.typeIndex[section.SectionType], section)
        
        // Index by topics
        for _, topic := range section.Topics {
            ihs.topicIndex[topic] = append(ihs.topicIndex[topic], section)
        }
        
        // Index by commands and flags
        for _, command := range section.Commands {
            ihs.commandIndex[command] = append(ihs.commandIndex[command], section)
        }
        for _, flag := range section.Flags {
            ihs.flagIndex[flag] = append(ihs.flagIndex[flag], section)
        }
    }
}
```

## Integration Patterns

### HTTP API Integration

```go
func setupQueryAPI(router *mux.Router, hs *help.HelpSystem) {
    router.HandleFunc("/api/help/search", func(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        if query == "" {
            http.Error(w, "Missing query field", http.StatusBadRequest)
            return
        }
        
        results, err := hs.QuerySections(query)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        
        response := struct {
            Query   string           `json:"query"`
            Results []*help.Section  `json:"results"`
            Count   int              `json:"count"`
        }{
            Query:   query,
            Results: results,
            Count:   len(results),
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })
}
```

### GraphQL Integration

```go
type Resolver struct {
    helpSystem *help.HelpSystem
}

func (r *Resolver) SearchHelp(ctx context.Context, args struct {
    Query string
    Limit *int32
}) ([]*help.Section, error) {
    results, err := r.helpSystem.QuerySections(args.Query)
    if err != nil {
        return nil, err
    }
    
    if args.Limit != nil && len(results) > int(*args.Limit) {
        results = results[:*args.Limit]
    }
    
    return results, nil
}

// GraphQL schema
const schema = `
    type Section {
        slug: String!
        title: String!
        content: String!
        sectionType: String!
        topics: [String!]!
        commands: [String!]!
        flags: [String!]!
    }
    
    type Query {
        searchHelp(query: String!, limit: Int): [Section!]!
    }
`
```

### CLI Plugin Integration

```go
func CreateQueryCommand(hs *help.HelpSystem) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "query [query]",
        Short: "Search help sections using query DSL",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            query := args[0]
            
            // Get flags
            format, _ := cmd.Flags().GetString("format")
            printQuery, _ := cmd.Flags().GetBool("print-query")
            printSQL, _ := cmd.Flags().GetBool("print-sql")
            
            // Debug output
            if printQuery {
                ast, err := dsl.ParseToAST(query)
                if err != nil {
                    return err
                }
                fmt.Printf("AST: %s\n", ast.String())
            }
            
            // Execute query
            results, err := hs.QuerySections(query)
            if err != nil {
                return err
            }
            
            // Format output
            switch format {
            case "json":
                return json.NewEncoder(os.Stdout).Encode(results)
            case "table":
                return printTable(results)
            default:
                return printList(results)
            }
        },
    }
    
    cmd.Flags().String("format", "list", "Output format: list, table, json")
    cmd.Flags().Bool("print-query", false, "Print parsed query AST")
    cmd.Flags().Bool("print-sql", false, "Print generated SQL")
    
    return cmd
}
```

## Debugging and Troubleshooting

### Query Analysis

```go
func analyzeQuery(query string) {
    fmt.Printf("Analyzing query: %s\n", query)
    
    // Parse to AST
    ast, err := dsl.ParseToAST(query)
    if err != nil {
        fmt.Printf("Parse error: %s\n", err)
        return
    }
    
    fmt.Printf("AST structure:\n%s\n", ast.String())
    
    // Analyze complexity
    complexity := ast.Complexity()
    fmt.Printf("Query complexity: %d\n", complexity)
    
    // Estimate performance
    if complexity > 10 {
        fmt.Println("Warning: High complexity query, consider optimization")
    }
    
    // Show field usage
    fields := ast.GetUsedFields()
    fmt.Printf("Fields used: %v\n", fields)
}
```

### Performance Profiling

```go
func profileQuery(hs *help.HelpSystem, query string) {
    start := time.Now()
    
    results, err := hs.QuerySections(query)
    
    duration := time.Since(start)
    
    if err != nil {
        fmt.Printf("Query failed after %v: %s\n", duration, err)
        return
    }
    
    fmt.Printf("Query completed in %v\n", duration)
    fmt.Printf("Results: %d sections\n", len(results))
    fmt.Printf("Performance: %.2f sections/ms\n", 
               float64(len(results))/float64(duration.Nanoseconds())*1e6)
    
    // Performance warnings
    if duration > 100*time.Millisecond {
        fmt.Println("Warning: Slow query detected")
    }
}
```

### Error Recovery

```go
func robustQuery(hs *help.HelpSystem, query string) ([]*help.Section, error) {
    results, err := hs.QuerySections(query)
    if err != nil {
        // Try to recover from common errors
        if strings.Contains(err.Error(), "unknown field") {
            // Suggest similar fields
            suggestions := suggestFields(query)
            return nil, fmt.Errorf("%s. Did you mean: %v", err, suggestions)
        }
        
        if strings.Contains(err.Error(), "syntax error") {
            // Try simplified query
            simplified := simplifyQuery(query)
            if simplified != query {
                fmt.Printf("Trying simplified query: %s\n", simplified)
                return hs.QuerySections(simplified)
            }
        }
        
        return nil, err
    }
    
    return results, nil
}
```

## API Reference

### Core Interfaces

```go
// Main help system interface
type HelpSystem interface {
    QuerySections(query string) ([]*Section, error)
    AddSection(section *Section)
    GetSectionWithSlug(slug string) (*Section, error)
}

// DSL parser interface
type Parser interface {
    Parse(query string) (Expression, error)
    ParseToAST(query string) (*AST, error)
    Validate(query string) error
}

// Query compiler interface
type Compiler interface {
    Compile(expr Expression) (Predicate, error)
    CompileToSQL(expr Expression) (string, []interface{}, error)
}
```

### Key Functions

```go
// DSL package
func Parse(query string) (Expression, error)
func ParseQuery(query string) (store.Predicate, error)
func GetQueryInfo() *QueryInfo
func ValidateQuery(query string) error

// Help system
func NewHelpSystem() *HelpSystem
func (hs *HelpSystem) QuerySections(query string) ([]*Section, error)

// Store backend (optional)
func NewInMemoryStore() (*Store, error)
func (s *Store) Find(ctx context.Context, pred Predicate) ([]*Section, error)
```

### Configuration Options

```go
type QueryConfig struct {
    MaxResults      int           // Maximum results to return
    Timeout         time.Duration // Query timeout
    CacheEnabled    bool          // Enable result caching
    CacheTTL        time.Duration // Cache time-to-live
    DebugMode       bool          // Enable debug output
    IndexingEnabled bool          // Enable field indexing
}

func ConfigureQuery(hs *HelpSystem, config QueryConfig) {
    // Apply configuration to help system
}
```

### Error Types

```go
type QueryError struct {
    Type    ErrorType
    Message string
    Query   string
    Position int
}

type ErrorType int

const (
    SyntaxError ErrorType = iota
    UnknownFieldError
    InvalidValueError
    TimeoutError
    InternalError
)
```

## Best Practices

### Query Design

1. **Start Simple**: Begin with basic field queries, add complexity gradually
2. **Use Indexes**: Leverage indexed fields (type, topic, command, flag) for performance
3. **Minimize Text Search**: Use field filters before text search when possible
4. **Cache Results**: Cache frequently used queries for better performance
5. **Validate Early**: Validate queries on the client side when possible

### Error Handling

1. **Graceful Degradation**: Fall back to simpler queries on errors
2. **User-Friendly Messages**: Provide helpful error messages with suggestions
3. **Logging**: Log query errors for debugging and optimization
4. **Timeout Protection**: Set reasonable timeouts for complex queries

### Testing

```go
func TestQuerySystem(t *testing.T) {
    hs := createTestHelpSystem()
    
    testCases := []struct {
        query    string
        expected int
        wantErr  bool
    }{
        {"examples", 5, false},
        {"type:tutorial", 3, false},
        {"invalid:field", 0, true},
        {"(examples OR tutorials) AND topic:database", 2, false},
    }
    
    for _, tc := range testCases {
        results, err := hs.QuerySections(tc.query)
        
        if tc.wantErr {
            assert.Error(t, err)
        } else {
            assert.NoError(t, err)
            assert.Len(t, results, tc.expected)
        }
    }
}
```

This guide provides comprehensive coverage of the query system's capabilities for developers. Use it as a reference for building applications that leverage the powerful help system query functionality.
