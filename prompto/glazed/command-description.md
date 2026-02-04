## 1. The `CommandDescription` struct

A **`CommandDescription`** is the primary metadata container for a command in this system. It includes:

- **Name** (short identifier, e.g., `"my-command"`)
- **Short** (a single-line description)
- **Long** (an optional, more-detailed description)
- **Sections** (contains field definitions, i.e. your command’s flags/arguments)
- **Parents** (if you have a hierarchy like `parent sub-command this-command`, you list `"parent", "sub-command"` as parents)
- **Source** (where the command definition was loaded from; optional)

### Minimal Example

```go
cd := cmds.NewCommandDescription("my-command",
    cmds.WithShort("Short description of my-command"),
    cmds.WithLong("A longer, more detailed explanation of my-command."),
)
```

---

## 2. Creating a CommandDescription

You create a new `CommandDescription` with:

```go
func NewCommandDescription(name string, options ...CommandDescriptionOption) *CommandDescription
```

- **`name`**: The string used to identify your command
- **`options...`**: Zero or more function options that configure aspects like short/long descriptions, flags, arguments, parents, etc.

Example usage:

```go
cd := cmds.NewCommandDescription("my-command",
    cmds.WithShort("Run my command"),
    cmds.WithLong("A longer help message describing what 'my-command' does."),
    // more options...
)
```

---

## 3. Adding Flags and Arguments

Your command’s fields (both flags and positional arguments) are grouped in a default “section.” You typically add them via the convenience functions:

- **`WithFlags(...)`**  
- **`WithArguments(...)`**

### 3.1 Defining Field Definitions

Fields themselves are described by `fields.Definition` from the `glazed/pkg/cmds/fields` package. For example:

```go
paramHost := fields.New(
    "host",
    fields.TypeString,
    fields.WithHelp("The host to connect to"),
    fields.WithDefault("localhost"),
    // fields.WithRequired(true), etc.
)
```

**Common field definition functions**:

- `fields.New(name string, paramType Type, opts ...DefinitionOption)`
- `fields.WithHelp("...")`
- `fields.WithDefault(...)`
- `fields.WithRequired(true|false)`
- etc.

### 3.2 Adding Flags

Flags are typically optional or named fields. You call `WithFlags(...)` with one or more `Definition`s:

```go
cd := cmds.NewCommandDescription("my-command",
    cmds.WithShort("Do something"),
    cmds.WithFlags(
        fields.New(
            "host",
            fields.TypeString,
            fields.WithHelp("The host to connect to"),
            fields.WithDefault("localhost"),
        ),
        fields.New(
            "verbose",
            fields.TypeBool,
            fields.WithHelp("Enable verbose output"),
            // no default => false by default
        ),
    ),
)
```

### 3.3 Adding Arguments

Positional arguments (like `my-command [ARGS ...]`) are also stored as fields but with `IsArgument = true`. You can use `WithArguments(...)`:

```go
cd := cmds.NewCommandDescription("my-command",
    cmds.WithShort("Process some files"),
    cmds.WithArguments(
        fields.New(
            "files",
            fields.TypeStringList,  // e.g. a list of filenames
            fields.WithHelp("The files to process"),
            fields.WithRequired(true),
        ),
    ),
)
```

Under the hood, `WithArguments(...)` sets `arg.IsArgument = true` on each definition.

---

## 4. CommandDescriptionOption Functions

You can pass various function options into `NewCommandDescription(...)`. The main ones:

1. **`WithName(s string)`**  
   Syntactic sugar to rename (rarely used because you typically pass the name in directly).

2. **`WithShort(s string)`**  
   Sets the single-line short description.

3. **`WithLong(s string)`**  
   Sets the longer help text.

4. **`WithFlags(flags ...*Definition)`**  
   Adds field definitions as **flags** to the default section.

5. **`WithArguments(arguments ...*Definition)`**  
   Adds field definitions as **positional arguments** to the default section.

6. **`WithSchema(schema *schema.Schema)`** or **`WithSections(sections ...schema.Section)`**  
   Used if you already have a custom `Schema` object or multiple field sections. Typically more advanced usage.

7. **`WithReplaceSections(sections ...schema.Section)`**  
   Replaces any existing sections with the ones you provide.

8. **`WithParents(p ...string)`**  
   If you want a hierarchical CLI structure, specify the chain of parent commands. Example:

   ```go
   cmds.NewCommandDescription("child",
       cmds.WithParents("top-level", "middle-level"),
   )
   ```
   Now `FullPath()` might produce `"top-level middle-level child"`.

9. **`WithStripParentsPrefix(prefixes []string)`**  
   If you loaded a command from somewhere but want to remove certain leading parent nodes, you can do that. E.g. if the command is originally in `[ "my", "prefix", "subcommand" ]`, you can strip `[ "my", "prefix" ]`.

10. **`WithSource(s string)`** and **`WithPrependSource(s string)`**  
    For debugging or record-keeping. Mark where the command was loaded from.

---

## 5. Inspecting Fields at Runtime

Once your `CommandDescription` is built, you can retrieve field definitions in code:

- **`GetDefaultFlags()`**: Returns a `Definitions` object of all flags in the default section.  
- **`GetDefaultArguments()`**: Returns all arguments (where `IsArgument = true`) from the default section.  
- **`Sections`**: The entire `Schema` object if you need advanced usage.

---

## 6. Example Putting It All Together

Below is a complete snippet that shows how you might define a command called `"my-command"` with some flags and arguments:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/glazed/pkg/cmds/fields"
    // ... other imports
)

func makeMyCommand() *cmds.CommandDescription {
    cd := cmds.NewCommandDescription("my-command",
        cmds.WithShort("Process some input"),
        cmds.WithLong("A command to process input with optional flags and arguments."),
        cmds.WithParents("toolkit", "utils"), // hierarchy: `toolkit utils my-command`
        
        // Add some flags
        cmds.WithFlags(
            fields.New(
                "host",
                fields.TypeString,
                fields.WithHelp("Hostname to connect"),
                fields.WithDefault("localhost"),
            ),
            fields.New(
                "port",
                fields.TypeInteger,
                fields.WithHelp("Port to use"),
                fields.WithDefault(8080),
            ),
        ),

        // Add some arguments
        cmds.WithArguments(
            fields.New(
                "paths",
                fields.TypeStringList,
                fields.WithHelp("Paths to process"),
                fields.WithRequired(true),
            ),
        ),

        // Maybe store info about source
        cmds.WithSource("my-commands.yaml"),
    )

    return cd
}
```

- The **`host`** and **`port`** are optional flags with defaults.  
- The **`paths`** argument is required and is a list of strings.  
- The command can be referred to as `toolkit utils my-command` in a hierarchical CLI.

---

## 7. Summary

1. **Create the command** with `cmds.NewCommandDescription("name", ...)`.  
2. **Give it short/long descriptions** via `WithShort(...)`, `WithLong(...)`.  
3. **Attach flags** with `WithFlags(...)` and **arguments** with `WithArguments(...)`.  
4. **Organize parents** if needed (`WithParents(...)`).  
5. **Optionally** set the source string or manipulate advanced schema composition.  
