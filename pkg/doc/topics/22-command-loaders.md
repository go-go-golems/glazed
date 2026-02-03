---
Title: Glazed Command Loaders
Slug: command-loaders
Short: Understanding how to load Glazed commands dynamically from various sources.
Topics:
- commands
- loaders
- yaml
- filesystem
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# Glazed Command Loaders

## Overview

While the [Using Commands](./15-using-commands.md) guide explains how to *create* commands programmatically in Go, many applications benefit from loading command definitions dynamically from external sources like files or directories. This allows for easier modification, addition of new commands without recompiling, and separation of command logic from the core application code.

Glazed provides a flexible **Command Loader** system to achieve this. Loaders are responsible for discovering command definitions in a specific format (like YAML) within a given filesystem (`io/fs.FS`) and translating them into runnable `cmds.Command` instances.

This guide explains the command loading architecture, how to use existing loaders, and how to create your own custom loaders.

## Architecture of Command Loading

The command loading process typically involves these components:

```
┌───────────────────┐      ┌──────────────────┐      ┌───────────────────────┐      ┌─────────────────┐
│ Filesystem (fs.FS)│────▶ │ LoadCommandsFromFS │────▶ │    CommandLoader      │────▶ │ cmds.Command(s) │
│ (e.g., os.DirFS)  │      │ (Helper function)  │      │ (Specific implementation)│      │ (Instances)     │
└───────────────────┘      └──────────────────┘      └─┬─────────────────────┘      └─────────────────┘
                                                      │ ▲
                                                      │ │ Uses
                                                      ▼ │
                                          ┌────────────────────────────┐
                                          │ LoadCommandOrAliasFromReader │
                                          │ (Helper function)          │
                                          └────────────────────────────┘
```

1.  **Filesystem (`io/fs.FS`)**: Represents the source of command definitions (e.g., a directory on disk, embedded files).
2.  **`LoadCommandsFromFS` Helper**: A utility function in `github.com/go-go-golems/glazed/pkg/cmds/loaders` that walks a filesystem directory structure, identifies potential command files, and delegates loading to a specific `CommandLoader`.
3.  **`CommandLoader` Interface**: Defined in `github.com/go-go-golems/glazed/pkg/cmds/loaders`, this interface standardizes how different command types are loaded. Implementations handle specific file formats or sources.
4.  **`LoadCommandOrAliasFromReader` Helper**: A utility function in `github.com/go-go-golems/glazed/pkg/cmds/loaders` often used by `CommandLoader` implementations. It takes an `io.Reader` and attempts to load either a `cmds.Command` or a `cmds.Alias` using provided loading functions, simplifying the process of handling both types from a single source.
5.  **`cmds.Command` Instances**: The final output – runnable command objects ready to be integrated into an application (e.g., added to a Cobra CLI).

## Key Components

### `CommandLoader` Interface (`github.com/go-go-golems/glazed/pkg/cmds/loaders`)

This is the core interface for all command loaders:

```go
package loaders

import (
	"io/fs"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
)

type CommandLoader interface {
	// LoadCommands reads an entry (file or directory) from the filesystem
	// and returns a list of commands or aliases found within.
	// options and aliasOptions are passed down to configure the
	// created CommandDescription or Alias.
	LoadCommands(
		f fs.FS,
		entryName string,
		options []cmds.CommandDescriptionOption,
		aliasOptions []alias.Option,
	) ([]cmds.Command, error)

	// IsFileSupported checks if a given file path within the filesystem
	// is potentially loadable by this specific loader. This is used by
	// helpers like LoadCommandsFromFS to decide which loader to use.
	IsFileSupported(f fs.FS, fileName string) bool
}

```

-   **`LoadCommands`**: The main method that performs the loading. It receives the filesystem, the path to the entry (file/directory), and options to configure the resulting command descriptions or aliases. Crucially, these options often include `cmds.WithSource` and `cmds.WithParents` provided by the calling context (like `LoadCommandsFromFS`) to track origin and hierarchy.
-   **`IsFileSupported`**: A quick check, usually based on file extension or content sniffing, to determine if the loader *might* handle the file.

### `LoadCommandsFromFS` Helper (`github.com/go-go-golems/glazed/pkg/cmds/loaders`)

This helper function simplifies loading commands from a directory structure:

```go
// Signature (simplified)
func LoadCommandsFromFS(
	f fs.FS,
	dir string,          // Directory to scan
	source string,       // Base source string for tracking
	loader CommandLoader, // The loader to use
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error)
```

It recursively walks the `dir` within the `f` filesystem. For each file, it calls `loader.IsFileSupported`. If supported, it invokes `loader.LoadCommands`, automatically adding `cmds.WithSource` and `cmds.WithParents` (derived from the directory structure) to the `options` and `aliasOptions` before passing them to the loader. This ensures loaded commands know their origin and place in the command hierarchy.

### `LoadCommandOrAliasFromReader` Helper (`github.com/go-go-golems/glazed/pkg/cmds/loaders`)

This helper is commonly used inside `LoadCommands` implementations:

```go
// Signature (simplified)
type LoadReaderCommandFunc func(
	r io.Reader,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error)


func LoadCommandOrAliasFromReader(
	r io.Reader,
	rawLoadCommand LoadReaderCommandFunc, // Your function to load a command
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error)
```

It reads the content from `r`, first tries to load it as a command using `rawLoadCommand`. If that fails, it attempts to load it as a command alias using `LoadCommandAliasFromYAML`. This abstracts away the need to handle command vs. alias loading logic explicitly in many loaders.

## Implementing a Custom `CommandLoader`

Let's create a hypothetical loader for commands defined in a simple custom format.

### Step 1: Define the Loader Struct

```go
package myloader

import (
	"io"
	"io/fs"
	"strings"
	// ... other necessary imports ...
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/pkg/errors"
)

type MyCustomLoader struct {
	// Any configuration needed for this loader
}

// Ensure interface is implemented
var _ loaders.CommandLoader = (*MyCustomLoader)(nil)
```

### Step 2: Implement `IsFileSupported`

Check if the file matches the expected format (e.g., by extension).

```go
func (mcl *MyCustomLoader) IsFileSupported(f fs.FS, fileName string) bool {
	// Example: Only support files ending in ".mycmd"
	return strings.HasSuffix(fileName, ".mycmd")
}
```

### Step 3: Implement `LoadCommands`

This is the core logic. It typically involves:
1.  Opening the file specified by `entryName` from the `f` filesystem.
2.  Passing the file reader and the provided `options` and `aliasOptions` to `LoadCommandOrAliasFromReader`, along with a function to handle the actual parsing (`loadMyCommandFromReader` in this case).

```go
func (mcl *MyCustomLoader) LoadCommands(
	f fs.FS,
	entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	r, err := f.Open(entryName)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file %s", entryName)
	}
	defer func(r fs.File) {
		_ = r.Close()
	}(r)

	// Add source tracking before passing options down
	// Note: LoadCommandsFromFS might have already added a source,
	// this prepends the file name for more specificity.
	// Use WithSource if LoadCommandsFromFS isn't the caller.
	sourceOption := cmds.WithPrependSource("file:" + entryName)
	allOptions := append(options, sourceOption)
	allAliasOptions := append(aliasOptions, alias.WithPrependSource("file:"+entryName))


	return loaders.LoadCommandOrAliasFromReader(
		r,
		mcl.loadMyCommandFromReader, // Function to parse our specific format
		allOptions,
		allAliasOptions,
	)
}
```

### Step 4: Implement the Format-Specific Loading Function

This function, passed to `LoadCommandOrAliasFromReader`, handles parsing the specific file format and creating the command object.

```go
// loadMyCommandFromReader parses the custom format from an io.Reader
func (mcl *MyCustomLoader) loadMyCommandFromReader(
	r io.Reader,
	options []cmds.CommandDescriptionOption, // These include parent/source info
	_ []alias.Option, // Ignored as we are loading a command here
) ([]cmds.Command, error) {
	// 1. Read and parse the custom format from 'r'
	//    (Implementation depends heavily on the format)
	//    Let's assume parsing yields name, shortDesc, and a list of params.
	contentBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read command file")
	}
	
	// Hypothetical parsing logic
	// commandName, shortDesc, paramDefs := parseMyCustomFormat(contentBytes)
	// if err != nil { ... handle error ... }
	commandName := "parsed-name" // Replace with actual parsed value
	shortDesc := "parsed-short" // Replace with actual parsed value
	paramDefs := []*fields.Definition{ /* ... parsed params ... */ }


	// 2. Create the CommandDescription, applying passed-in options FIRST
	description := cmds.NewCommandDescription(
		commandName,
		options..., // Apply options from caller (includes parents, source)
	)

	// 3. Apply options derived from the file content itself
	cmds.WithShort(shortDesc)(description)
	cmds.WithFlags(paramDefs...)(description)
	// Add other options like WithLong, WithArguments, WithSections etc. as needed

	// 4. Create the actual command instance (assuming a specific command type)
	//    This depends on the type of command being loaded (Bare, Writer, Glaze)
	//    Let's assume we are loading a simple BareCommand implementation.
	myCmd := NewMySpecificCommand(description /*, other args */)

	return []cmds.Command{myCmd}, nil
}

// Placeholder for the actual command implementation
type MySpecificCommand struct {
	*cmds.CommandDescription
	// ... other fields ...
}
func (msc *MySpecificCommand) Run(ctx context.Context, parsedSections *values.Values) error { /* ... */}
func NewMySpecificCommand(desc *cmds.CommandDescription) *MySpecificCommand { /* ... */ }

```

**Key points in `loadMyCommandFromReader`:**
- It receives `options` which already contain crucial context like `WithParents` and `WithSource` from the caller (e.g., `LoadCommandsFromFS`). Apply these first when creating the `CommandDescription`.
- Then, apply options derived from the parsed file content (`WithShort`, `WithFlags`, etc.).
- Finally, instantiate the specific `Command` implementation using the configured `CommandDescription`.

## Built-in and Example Loaders

Glazed and related projects use several loaders:

-   **`YAMLCommandLoader` (`github.com/go-go-golems/glazed/pkg/cmds/loaders`)**: A generic loader for commands defined purely by their `CommandDescription` structure in YAML. Less common now, as specific loaders are preferred.
-   **`SqlCommandLoader` (`github.com/go-go-golems/sqleton/pkg/cmds`)**: Loads SQL execution commands for the `sqleton` tool from YAML files containing SQL queries and field definitions. Uses `loaders.CheckYamlFileType(f, fileName, "sqleton")` in `IsFileSupported`.
-   **`PinocchioCommandLoader` (`github.com/go-go-golems/pinocchio/pkg/cmds`)**: Loads AI prompt commands for the `pinocchio` tool from YAML files defining prompts, messages, and AI settings. Checks for `type: pinocchio` in the YAML.
-   **`AgentCommandLoader` (`github.com/go-go-golems/goagent/pkg/cmds`)**: Loads LLM Agent commands for the `goagent` framework from YAML files. Checks for `.yaml`/`.yml` suffix.

These examples demonstrate how loaders are tailored to specific command types and configuration formats.

## Handling Multiple Command Types: `MultiLoader`

What if you need to load different *types* of commands (e.g., SQL commands and AI commands) from the same directory structure? The `MultiLoader` (`github.com/go-go-golems/glazed/pkg/cmds/loaders`) solves this.

```go
package loaders

// MultiLoader dispatches to registered loaders based on a 'type' field.
type MultiLoader struct {
    loaders map[string]CommandLoader // Map type name -> specific loader
    defaultLoader CommandLoader      // Optional loader if no type field
}

func NewMultiLoader() *MultiLoader
func (m *MultiLoader) RegisterLoader(typeName string, loader CommandLoader)
func (m *MultiLoader) SetDefaultLoader(loader CommandLoader)

// Implements CommandLoader interface
func (m *MultiLoader) LoadCommands(...) error
func (m *MultiLoader) IsFileSupported(...) bool
```

1.  Create a `MultiLoader` instance.
2.  `RegisterLoader` specific loaders (like `SqlCommandLoader`, `PinocchioCommandLoader`) associated with a unique `typeName` string.
3.  Optionally `SetDefaultLoader` for files without a type hint.
4.  Pass the `MultiLoader` instance to `LoadCommandsFromFS`.

When `MultiLoader.LoadCommands` is called, it first peeks into the file (usually YAML) to find a `type:` field. It then uses the loader registered for that type. If no type field is found, it uses the default loader or tries to find any registered loader whose `IsFileSupported` returns true.

This allows a single loading mechanism to handle diverse command definitions differentiated by a type hint in their source file.

## Best Practices for Loaders

1.  **Clear `IsFileSupported`**: Make it efficient. Prefer checking file extensions or simple magic bytes/keywords over full parsing if possible.
2.  **Leverage Helpers**: Use `LoadCommandOrAliasFromReader` when applicable to handle commands and aliases uniformly.
3.  **Pass Options Correctly**: Ensure `options` and `aliasOptions` received in `LoadCommands` are passed down to `NewCommandDescription` and `NewCommandAlias` (often via `LoadCommandOrAliasFromReader`) to preserve source and parent information.
4.  **Add Source Information**: Use `cmds.WithSource` or `cmds.WithPrependSource` within your loader if the caller hasn't already provided sufficient detail.
5.  **Handle Errors Gracefully**: Use `errors.Wrap` or `errors.Wrapf` to provide context when file reading or parsing fails.
6.  **Focus `LoadCommands`**: Keep `LoadCommands` focused on opening the file and delegating parsing (often via `LoadCommandOrAliasFromReader`). Put the format-specific logic in the dedicated reader function.

## Summary

Command loaders are essential for building flexible Glazed applications where commands can be defined externally. By implementing the `CommandLoader` interface, you can teach Glazed how to understand different command definition formats and sources. Using helpers like `LoadCommandsFromFS` and `LoadCommandOrAliasFromReader`, along with the `MultiLoader` for complex scenarios, provides a robust system for dynamic command loading. 