---
Title: Templating Helpers in Glazed
Slug: templating-helpers
Short: Guide to using the Glazed templating package.
Topics:
- templating
- helpers
- go-templates
- sprig
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Using Templating Helpers in Glazed

## Overview

The `glazed` framework includes a powerful templating helper package located at `github.com/go-go-golems/glazed/pkg/helpers/templating`. This package extends Go's standard `text/template` library by integrating the popular [Sprig](http://masterminds.github.io/sprig/) function library and adding several custom functions tailored for common data manipulation and formatting tasks within Glazed applications.

This guide provides a comprehensive overview of how to use the `templating` package, create templates, leverage the available functions, and integrate templating into your Glazed projects.

## Core Concepts

The `templating` package builds upon Go's standard templating system. Key concepts include:

1.  **Templates**: Text documents with embedded actions (`{{ }}`) that are evaluated to produce output.
2.  **`template.Template`**: The core Go type representing a parsed template.
3.  **`template.FuncMap`**: A map associating names with functions that can be called from within a template.
4.  **Execution**: The process of applying a parsed template to data, executing the embedded actions, and writing the output.

The Glazed `templating` package simplifies this by:

-   Providing a convenient constructor (`CreateTemplate`) that automatically includes standard Go template functions, Sprig functions, and Glazed-specific custom functions.
-   Offering a curated set of useful custom functions for tasks like string manipulation, date formatting, random data generation, and basic arithmetic.

## Creating and Using Templates

The primary way to create a template instance enriched with Glazed helpers is using `templating.CreateTemplate`.

### Basic Template Creation

```go
import "github.com/go-go-golems/glazed/pkg/helpers/templating"

// Create a new template named "myTemplate"
// It automatically includes Go, Sprig, and Glazed functions.
tmpl := templating.CreateTemplate("myTemplate")

// Parse the template string
parsedTmpl, err := tmpl.Parse("Hello, {{ .Name | upper }}! Your lucky number is {{ randomInt 1 100 }}.")
if err != nil {
    // Handle error
}

// Prepare data
data := map[string]interface{}{
    "Name": "World",
}

// Execute the template
var output strings.Builder
err = parsedTmpl.Execute(&output, data)
if err != nil {
    // Handle error
}

fmt.Println(output.String())
// Example Output: Hello, WORLD! Your lucky number is 42.
```

### Adding Custom Functions

You can extend the default function set by adding your own `template.FuncMap`:

```go
import (
    "text/template"
    "github.com/go-go-golems/glazed/pkg/helpers/templating"
)

// Define a custom function
func customGreeting(name string) string {
    return "Aloha, " + name + "!"
}

// Create the FuncMap
customFuncs := template.FuncMap{
    "greet": customGreeting,
}

// Create a template and add the custom functions
tmpl := templating.CreateTemplate("customTemplate").Funcs(customFuncs)

// Parse and execute as before
parsedTmpl, err := tmpl.Parse("{{ greet .Name }}")
// ... execute ...
```

### Example: Using Templates in `geppetto`

The `geppetto/pkg/conversation/builder/builder.go` file demonstrates using templates to construct prompts and messages dynamically:

```go
// Simplified example from geppetto/pkg/conversation/builder/builder.go

import (
    "github.com/go-go-golems/glazed/pkg/helpers/templating"
    "strings"
)

func renderTemplate(templateString string, variables map[string]interface{}) (string, error) {
    tmpl, err := templating.CreateTemplate("prompt").Parse(templateString)
    if err != nil {
        return "", errors.Wrap(err, "failed to parse prompt template")
    }

    var buffer strings.Builder
    err = tmpl.Execute(&buffer, variables)
    if err != nil {
        return "", errors.Wrap(err, "failed to execute prompt template")
    }
    return buffer.String(), nil
}

// Usage:
// systemPromptTemplate := "Be a helpful assistant focusing on {{ .topic }}."
// vars := map[string]interface{}{"topic": "Go programming"}
// renderedPrompt, _ := renderTemplate(systemPromptTemplate, vars)
// -> "Be a helpful assistant focusing on Go programming."
```

### Example: Using Templates in `clay` (SQL Generation)

The `clay/pkg/sql/template.go` file showcases a more advanced use case where templates generate SQL queries. It adds specific SQL-related functions to the template:

```go
// Simplified example from clay/pkg/sql/template.go
import (
	"context"
	"text/template"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
	"github.com/jmoiron/sqlx"
)

// SQL-specific template functions (examples)
func sqlStringIn(values interface{}) (string, error) { /* ... implementation ... */ }
func sqlDate(date interface{}) (string, error) { /* ... implementation ... */ }
func subQuery(name string, subQueries map[string]string) (string, error) { /* ... */ }
// ... other SQL functions

func CreateSQLTemplate(
	ctx context.Context,
	subQueries map[string]string,
	ps map[string]interface{}, // Fields for the template
	db *sqlx.DB, // Database connection (used by some functions like sqlColumn)
) *template.Template {
	// Start with the Glazed base template
	tmpl := templating.CreateTemplate("query").
		Funcs(templating.TemplateFuncs) // Ensure Glazed funcs are included

	// Add SQL-specific functions
	tmpl.Funcs(template.FuncMap{
		"sqlStringIn":    sqlStringIn,
		"sqlDate":        sqlDate,
		"subQuery": func(name string) (string, error) { // Closure to capture subQueries
			s, ok := subQueries[name]
			if !ok { return "", errors.Errorf("Subquery %s not found", name) }
			return s, nil
		},
		"sqlColumn": func(query string, args ...interface{}) ([]interface{}, error) {
			// Function implementation that runs a query using db and ps...
		},
		// ... other SQL functions
	})

	return tmpl
}

// Usage:
// sqlTmpl := CreateSQLTemplate(ctx, subQueries, params, db)
// parsedTmpl, _ := sqlTmpl.Parse("SELECT * FROM users WHERE id = {{ .userId }} AND status IN ({{ sqlStringIn .statuses }})")
// vars := map[string]interface{}{"userId": 123, "statuses": []string{"active", "pending"}}
// var queryBuf strings.Builder
// parsedTmpl.Execute(&queryBuf, vars)
// -> SELECT * FROM users WHERE