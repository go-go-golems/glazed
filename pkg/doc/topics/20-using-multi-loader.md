---
Title: Using the MultiLoader
Slug: using-multi-loader
Short: Learn how to use the MultiLoader to handle different types of command files with type-based dispatch
Topics:
  - Commands
  - YAML
  - Configuration
  - Loaders
Commands:
  - LoadCommands
Flags:
  - none
IsTopLevel: true
ShowPerDefault: false
SectionType: GeneralTopic
---

# Using the MultiLoader in Glazed

The MultiLoader provides a flexible way to load commands from different file types by dispatching to appropriate loaders based on a `type` field or file content. This guide explains how to use and configure the MultiLoader for your applications.

## Basic Structure

The MultiLoader works by first checking a file's `type` field (if present) and then dispatching to the appropriate registered loader. Here's how a typical command file with a type field looks:

```yaml
type: sqleton
name: query-users
short: Query the users table
flags:
  - name: limit
    type: int
    help: Maximum number of users to return
```

## Setting Up the MultiLoader

### Basic Setup

Create and configure a MultiLoader with specific type handlers:

```go
// Create a new MultiLoader
loader := loaders.NewMultiLoader()

// Register loaders for different types
loader.RegisterLoader("sql", sqlLoader)
loader.RegisterLoader("http", httpLoader)
loader.RegisterLoader("shell", shellLoader)

// Set a default loader for files without a type field
loader.SetDefaultLoader(defaultLoader)
```

### Using the MultiLoader

Use the configured MultiLoader like any other CommandLoader:

```go
commands, err := loader.LoadCommands(fs, "commands/query.yaml", options, aliasOptions)
if err != nil {
    // Handle error
}
```

## Loader Selection Process

The MultiLoader follows this process to select the appropriate loader:

1. First, attempts to parse the file as YAML and check for a `type` field
2. If a type is found, uses the registered loader for that type
3. If no type is found or parsing fails:
   - Uses the default loader if one is set
   - Tries each registered loader to find one that supports the file
4. Returns an error if no suitable loader is found

## Supported Features

### Type-Based Dispatch

Register loaders for specific command types:

```go
// SQL command loader
loader.RegisterLoader("sql", &SQLCommandLoader{})

// HTTP command loader
loader.RegisterLoader("http", &HTTPCommandLoader{})
```

### Default Loader

Set a fallback loader for files without a type field:

```go
loader.SetDefaultLoader(&YAMLCommandLoader{})
```

### Automatic Loader Detection

The MultiLoader can automatically detect the right loader even without a type field:

```go
// Will try each registered loader's IsFileSupported method
commands, err := loader.LoadCommands(fs, "commands/unknown.txt", options, aliasOptions)
```

## Example Command Files

### SQL Command
```yaml
type: sqleton
name: list-users
short: List all users from database
flags:
  - name: order-by
    type: string
    help: Column to sort by
    default: id
```

### HTTP Command
```yaml
type: http
name: get-weather
short: Get weather information
flags:
  - name: city
    type: string
    help: City name
    required: true
```

### Default YAML Command
```yaml
# No type field - will use default loader
name: simple-command
short: A simple command
flags:
  - name: verbose
    type: bool
    help: Enable verbose output
```

## Common Patterns

### Registering Multiple Loaders

```go
func RegisterLoaders(ml *loaders.MultiLoader) {
    ml.RegisterLoader("sql", NewSQLLoader())
    ml.RegisterLoader("http", NewHTTPLoader())
    ml.RegisterLoader("shell", NewShellLoader())
    
    // Set default loader last
    ml.SetDefaultLoader(NewYAMLLoader())
}
```

### Conditional Loading

```go
func LoadCommandsWithMultiLoader(fs fs.FS, path string) ([]cmds.Command, error) {
    loader := loaders.NewMultiLoader()
    
    // Register loaders based on configuration
    if config.SQLEnabled {
        loader.RegisterLoader("sql", NewSQLLoader())
    }
    if config.HTTPEnabled {
        loader.RegisterLoader("http", NewHTTPLoader())
    }
    
    return loader.LoadCommands(fs, path, nil, nil)
}
```

## Conclusion

The MultiLoader provides a flexible and extensible way to handle different types of command files in your application. By following these guidelines and patterns, you can create a robust command loading system that supports multiple file formats and command types while maintaining clean, maintainable code. 