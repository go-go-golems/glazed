---
Title: Building Custom Processors
Slug: building-custom-processors
Short: A comprehensive guide on how to build custom processors using the provided middleware and types system in Glazed.
Topics:
  - processors
  - middleware
  - data transformation
Commands:
  - process
Flags:
  - middleware
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Building Custom Processors: A Comprehensive Guide

This guide explains how to build custom processors using the provided middleware and types system. We'll cover the core concepts, type definitions, and provide examples of implementing various processors.

## Core Concepts

The processing system is built around three main concepts:
1. **Rows** - Individual data records with ordered fields
2. **Tables** - Collections of rows with defined column ordering
3. **Processors** - Components that process and transform data

## Type Definitions

### Basic Types
```go
type TableName = string
type FieldName = string
type GenericCellValue = interface{}
type Row = *orderedmap.OrderedMap[FieldName, GenericCellValue]
```

### Table Structure
```go
type Table struct {
    Columns   []FieldName
    Rows      []Row
    finalized bool
}
```

### Processor Interface
```go
type Processor interface {
    AddRow(ctx context.Context, obj Row) error
    Close(ctx context.Context) error
}
```

### Middleware Types
```go
type ObjectMiddleware interface {
    Process(ctx context.Context, row Row) ([]Row, error)
    Close(ctx context.Context) error
}

type RowMiddleware interface {
    Process(ctx context.Context, row Row) ([]Row, error)
    Close(ctx context.Context) error
}

type TableMiddleware interface {
    Process(ctx context.Context, table *Table) (*Table, error)
    Close(ctx context.Context) error
}
```

## Working with Rows

Rows are implemented using OrderedMap, which preserves field order. Here's how to work with them:

### Creating Rows

```go
// Create an empty row
row := types.NewRow()

// Create a row with initial data
row := types.NewRow(
    types.MRP("name", "John"),
    types.MRP("age", 30),
)

// Create from a map (fields will be sorted alphabetically)
data := map[string]interface{}{
    "age": 30,
    "name": "John",
}
row := types.NewRowFromMap(data)

// Create from a struct
type Person struct {
    Name string
    Age  int
}
person := Person{Name: "John", Age: 30}
row := types.NewRowFromStruct(person, true) // true = lowercase keys
```

### Manipulating Rows

```go
// Set values
row.Set("email", "john@example.com")

// Get values
value, exists := row.Get("name")

// Iterate over fields
for pair := row.Oldest(); pair != nil; pair = pair.Next() {
    fmt.Printf("%s: %v\n", pair.Key, pair.Value)
}
```

## Building Custom Processors

### Simple Object Middleware

Here's an example of a middleware that adds a timestamp to each row:

```go
type TimestampMiddleware struct{}

func (m *TimestampMiddleware) Process(ctx context.Context, row Row) ([]Row, error) {
    row.Set("timestamp", time.Now().Unix())
    return []Row{row}, nil
}

func (m *TimestampMiddleware) Close(ctx context.Context) error {
    return nil
}
```

### Row Filter Middleware

This middleware filters rows based on a condition:

```go
type AgeFilterMiddleware struct {
    minAge int
}

func (m *AgeFilterMiddleware) Process(ctx context.Context, row Row) ([]Row, error) {
    age, exists := row.Get("age")
    if !exists {
        return []Row{}, nil
    }
    
    if ageVal, ok := age.(int); ok && ageVal >= m.minAge {
        return []Row{row}, nil
    }
    
    return []Row{}, nil
}

func (m *AgeFilterMiddleware) Close(ctx context.Context) error {
    return nil
}
```

### Table Middleware

This middleware sorts rows by a specific column:

```go
type SortTableMiddleware struct {
    sortColumn string
}

func (m *SortTableMiddleware) Process(ctx context.Context, table *Table) (*Table, error) {
    sort.Slice(table.Rows, func(i, j int) bool {
        val1, _ := table.Rows[i].Get(m.sortColumn)
        val2, _ := table.Rows[j].Get(m.sortColumn)
        str1 := fmt.Sprintf("%v", val1)
        str2 := fmt.Sprintf("%v", val2)
        return str1 < str2
    })
    return table, nil
}

func (m *SortTableMiddleware) Close(ctx context.Context) error {
    return nil
}
```

## Using the TableProcessor

The TableProcessor combines multiple middlewares:

```go
// Create a new processor
processor := NewTableProcessor(
    WithObjectMiddleware(&TimestampMiddleware{}),
    WithRowMiddleware(&AgeFilterMiddleware{minAge: 18}),
    WithTableMiddleware(&SortTableMiddleware{sortColumn: "name"}),
)

// Process rows
row := types.NewRow(
    types.MRP("name", "John"),
    types.MRP("age", 25),
)
err := processor.AddRow(context.Background(), row)
if err != nil {
    // Handle error
}

// Close processor to apply table middlewares
err = processor.Close(context.Background())
if err != nil {
    // Handle error
}

// Get processed table
table := processor.GetTable()
```

## Processing Order

The processing pipeline follows this order:

1. ObjectMiddlewares process each row (can transform 1 row into many)
2. RowMiddlewares process the resulting rows (can filter or transform)
3. Rows are collected into the table
4. On Close(), TableMiddlewares process the entire table
5. Middlewares are closed in reverse order: Table -> Row -> Object

## Best Practices

1. **Error Handling**: Always return meaningful errors from Process() and Close()
2. **Context Usage**: Respect context cancellation in long-running operations
3. **Memory Management**: Consider using table middlewares only when necessary
4. **Immutability**: Create new rows instead of modifying existing ones when appropriate
5. **Cleanup**: Implement Close() properly to free resources

## Common Patterns

### Row Transformation
```go
func (m *MyMiddleware) Process(ctx context.Context, row Row) ([]Row, error) {
    newRow := types.NewRowFromRow(row)  // Create a copy
    // Modify newRow
    return []Row{newRow}, nil
}
```

### Row Splitting
```go
func (m *MyMiddleware) Process(ctx context.Context, row Row) ([]Row, error) {
    rows := []Row{}
    // Create multiple rows from input
    return rows, nil
}
```

### Row Filtering
```go
func (m *MyMiddleware) Process(ctx context.Context, row Row) ([]Row, error) {
    if shouldKeep(row) {
        return []Row{row}, nil
    }
    return []Row{}, nil
}
```
