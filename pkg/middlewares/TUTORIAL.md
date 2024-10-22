# Tutorial: Working with Rows, Tables, and Middlewares in Glazed

The glazed middleware package is a powerful tool for processing and transforming data in a flexible and extensible manner. This tutorial will guide you through the basics of working with rows, tables, and middlewares, and how to use them together to process data effectively.

Use "github.com/go-go-golems/glazed/pkg/middlewares" to import the package.

## 1. Understanding Rows and Tables

### 1.1 Rows

In Glazed, a `Row` represents a single record of data. It's essentially a collection of key-value pairs, where keys are field names and values can be of any type.

```go
// Creating a new row
row := types.NewRow(
    types.MRP("name", "John Doe"),
    types.MRP("age", 30),
    types.MRP("city", "New York")
)

// Accessing values
name, _ := row.Get("name")
fmt.Println(name) // Output: John Doe

// Setting values
row.Set("country", "USA")
```

### 1.2 Tables

A `Table` is a collection of rows with a defined set of columns.

```go
// Creating a new table
table := types.NewTable()
table.Columns = []types.FieldName{"name", "age", "city"}

// Adding rows to the table
table.AddRows(
    types.NewRow(
        types.MRP("name", "John Doe"),
        types.MRP("age", 30),
        types.MRP("city", "New York")
    ),
    types.NewRow(
        types.MRP("name", "Jane Smith"),
        types.MRP("age", 25),
        types.MRP("city", "London")
    )
)
```

## 2. The middlewares.Processor

The `Processor` interface is the core of the middleware system in Glazed. It defines two methods:


```8:24:glazed/pkg/middlewares/mod.go
type TableMiddleware interface {
	// Process transforms a full table
	Process(ctx context.Context, table *types.Table) (*types.Table, error)
	Close(ctx context.Context) error
}

type ObjectMiddleware interface {
	// Process transforms each individual object. Each object can return multiple
	// objects which will get processed individually downstream.
	Process(ctx context.Context, object types.Row) ([]types.Row, error)
	Close(ctx context.Context) error
}

type RowMiddleware interface {
	Process(ctx context.Context, row types.Row) ([]types.Row, error)
	Close(ctx context.Context) error
}
```


The main implementation of this interface is the `TableProcessor`, which manages a collection of middlewares and applies them to rows as they are added.

You can create a new `TableProcessor` and add middlewares to it:


```21:41:glazed/pkg/middlewares/processor.go
var _ Processor = (*TableProcessor)(nil)

type TableProcessorOption func(*TableProcessor)

func WithTableMiddleware(tm ...TableMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.TableMiddlewares = append(p.TableMiddlewares, tm...)
	}
}

func WIthPrependTableMiddleware(tm ...TableMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.TableMiddlewares = append(tm, p.TableMiddlewares...)
	}
}

func WithObjectMiddleware(om ...ObjectMiddleware) TableProcessorOption {
	return func(p *TableProcessor) {
		p.ObjectMiddlewares = append(p.ObjectMiddlewares, om...)
	}
}
```


## 3. How Middlewares Work

Middlewares in Glazed are categorized into three types:

1. Object Middlewares: Process individual objects before they become rows.
2. Row Middlewares: Process rows after they've been created from objects.
3. Table Middlewares: Process the entire table after all rows have been added.

Each middleware type has a specific interface it must implement:


```8:24:glazed/pkg/middlewares/mod.go
type TableMiddleware interface {
	// Process transforms a full table
	Process(ctx context.Context, table *types.Table) (*types.Table, error)
	Close(ctx context.Context) error
}

type ObjectMiddleware interface {
	// Process transforms each individual object. Each object can return multiple
	// objects which will get processed individually downstream.
	Process(ctx context.Context, object types.Row) ([]types.Row, error)
	Close(ctx context.Context) error
}

type RowMiddleware interface {
	Process(ctx context.Context, row types.Row) ([]types.Row, error)
	Close(ctx context.Context) error
}
```


The `TableProcessor` applies these middlewares in the following order:

1. Object Middlewares
2. Row Middlewares
3. Table Middlewares (during Close())

Here's how the `AddRow` method processes a row through the middlewares:


```113:149:glazed/pkg/middlewares/processor.go
func (p *TableProcessor) AddRow(ctx context.Context, row types.Row) error {
	rows := []types.Row{row}

	for _, ow := range p.ObjectMiddlewares {
		newRows := []types.Row{}
		for _, row_ := range rows {
			rows_, err := ow.Process(ctx, row_)
			if err != nil {
				return err
			}
			newRows = append(newRows, rows_...)
		}

		rows = newRows
	}

	for _, mw := range p.RowMiddlewares {
		newRows := []types.Row{}
		for _, row_ := range rows {
			rows_, err := mw.Process(ctx, row_)
			if err != nil {
				return err
			}
			newRows = append(newRows, rows_...)
		}

		rows = newRows
	}

	// Only collect table rows if we have table middlewares to actually process them,
	// otherwise discard the row so that we don't waste memory.
	if len(p.TableMiddlewares) > 0 {
		p.Table.AddRows(rows...)
	}

	return nil
}
```


## 4. Important Middlewares and Usage Examples

### 4.1 Row Middlewares

#### FlattenObjectMiddleware

Flattens nested objects in a row.

```go
flattenMiddleware := row.NewFlattenObjectMiddleware()
processor.AddRowMiddleware(flattenMiddleware)
```

#### TemplateMiddleware

Applies Go templates to row fields.


```18:22:glazed/pkg/middlewares/row/template.go
	// that contain a ".", which is frequent due to flattening.
	RenameSeparator string
	funcMaps        []template.FuncMap

	renamedColumns map[types.FieldName]types.FieldName
```


#### SkipLimitMiddleware

Skips a certain number of rows and limits the total number of rows processed.


```164:164:glazed/pkg/middlewares/row/skip-limit_test.go
			skipLimitMiddleware := &SkipLimitMiddleware{Skip: tt.skip, Limit: tt.limit}
```


### 4.2 Table Middlewares

#### SortByMiddleware

Sorts the table rows based on specified columns.


```33:57:glazed/pkg/middlewares/table/sortby.go
func NewSortByMiddlewareFromColumns(columns ...string) *SortByMiddleware {
	ret := &SortByMiddleware{
		columns: make([]columnOrder, 0),
	}

	for _, column := range columns {
		if len(column) == 0 {
			continue
		}

		isAsc := true

		if column[0] == '-' {
			column = column[1:]
			isAsc = false
		}

		ret.columns = append(ret.columns, columnOrder{
			name: column,
			asc:  isAsc,
		})
	}

	return ret
}
```


#### OutputMiddleware

Outputs the table using a specified formatter.


```20:25:glazed/pkg/middlewares/table/output.go
func NewOutputMiddleware(formatter formatters.TableOutputFormatter, writer io.Writer) *OutputMiddleware {
	return &OutputMiddleware{
		formatter: formatter,
		writer:    writer,
	}
}
```


### 4.3 Object Middlewares

#### JqObjectMiddleware

Applies jq-like transformations to objects before they become rows.


```13:17:glazed/pkg/middlewares/jq_test.go
	ret, err := NewJqObjectMiddleware(e)
	require.NoError(t, err)

	return ret
}
```


## 5. Putting It All Together

Here's an example that demonstrates how to use the `TableProcessor` with various middlewares:

```go
import (
    "context"
    "github.com/go-go-golems/glazed/pkg/middlewares"
    "github.com/go-go-golems/glazed/pkg/middlewares/row"
    "github.com/go-go-golems/glazed/pkg/middlewares/table"
    "github.com/go-go-golems/glazed/pkg/types"
)

func processData(data []map[string]interface{}) error {
    ctx := context.Background()

    processor := middlewares.NewTableProcessor(
        middlewares.WithRowMiddleware(
            row.NewFlattenObjectMiddleware(),
            row.NewTemplateMiddleware(
                map[types.FieldName]string{
                    "fullName": "{{.firstName}} {{.lastName}}",
                },
                "",
            ),
        ),
        middlewares.WithTableMiddleware(
            table.NewSortByMiddlewareFromColumns("age"),
            table.NewOutputMiddleware(
                formatters.NewCSVFormatter(),
                os.Stdout,
            ),
        ),
    )

    for _, item := range data {
        if err := processor.AddRow(ctx, types.NewRowFromMap(item)); err != nil {
            return err
        }
    }

    return processor.Close(ctx)
}
```

This example creates a `TableProcessor` with row and table middlewares. It processes a slice of maps, flattens nested objects, applies a template to create a "fullName" field, sorts the resulting table by age, and outputs the result as CSV to stdout.

Often, a function processDataIntoProcessor is used that gets passed a processor that has already been initialized, often when using a glazed command.

By using middlewares, you can easily create flexible data processing pipelines that transform and output structured data in various formats.