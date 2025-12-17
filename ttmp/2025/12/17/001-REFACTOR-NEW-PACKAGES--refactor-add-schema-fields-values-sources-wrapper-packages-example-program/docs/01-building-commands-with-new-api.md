---
Title: Building Commands with the New API
Slug: building-commands-with-new-api
Short: Guide to building Glazed commands using the new schema/fields/values/sources vocabulary
Topics:
- tutorial
- commands
- schema
- fields
- values
- api
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# Building Commands with the New API

Glazed's new API vocabulary (`schema`, `fields`, `values`, `sources`) provides clearer, more intuitive names for building commands. Instead of working with "parameter layers" and "parsed layers," you now work with **schemas** (collections of sections), **fields** (parameter definitions), **values** (resolved parameters), and **sources** (where values come from). This guide demonstrates how to build a complete command using these new packages, showing how schema sections organize parameters, how fields define individual parameters, and how values decode into type-safe structs.

**Learning objectives:**
- Understand the new API vocabulary and its relationship to the underlying system
- Create commands with multiple schema sections using `schema.NewSection()`
- Define fields using `fields.New()` with type-safe options
- Decode resolved values into structs using `values.DecodeSectionInto()`
- Configure environment variable parsing and Cobra flag integration

## Prerequisites

- Go 1.19+ installed
- Basic familiarity with Go and command-line tools
- Understanding of Glazed's core concepts (see `glaze help build-first-command`)

## Understanding the New Vocabulary

The new API packages provide clearer names that align with common developer mental models. A **Schema** is a collection of **Sections**, where each section groups related **Fields**. When a command runs, these fields are resolved from various **Sources** (defaults, environment variables, config files, Cobra flags) into **Values** that can be decoded into Go structs.

**Key concepts:**
- **Schema**: A collection of sections (`schema.Schema` = `layers.ParameterLayers`)
- **Section**: A named group of fields (`schema.Section` = `layers.ParameterLayer`)
- **Field**: A parameter definition (`fields.Definition` = `parameters.ParameterDefinition`)
- **Values**: Resolved parameters (`values.Values` = `layers.ParsedLayers`)
- **Sources**: Where values come from (`sources` package provides middleware wrappers)

This vocabulary makes it easier to think about command structure: you define a schema with sections, each section contains fields, and at runtime you decode values from those sections into your application structs.

## Step 1: Define Your Command Structure

Every Glazed command follows a consistent pattern: a command struct embeds `*cmds.CommandDefinition` for metadata, and settings structs map command-line parameters to Go fields using struct tags. The new API makes this pattern clearer by using `CommandDefinition` instead of `CommandDescription`.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// Command struct embeds CommandDefinition
type AppCommand struct {
	*cmds.CommandDefinition
}

// Settings structs map to schema sections
type AppSettings struct {
	Verbose bool   `glazed.parameter:"verbose"`
	Port    int    `glazed.parameter:"port"`
	Host    string `glazed.parameter:"host"`
}

type OutputSettings struct {
	Format string `glazed.parameter:"format"`
	Pretty bool   `glazed.parameter:"pretty"`
}
```

**Key components:**

1. **Command Struct**: `AppCommand` embeds `*cmds.CommandDefinition`, which contains command metadata and schema configuration
2. **Settings Structs**: Each struct corresponds to a schema section, with fields tagged using `glazed.parameter:"field-name"` for automatic mapping
3. **Type Safety**: Struct fields provide compile-time type checking and automatic conversion from string inputs

## Step 2: Create Schema Sections

Schema sections organize related parameters into logical groups. Each section can have a prefix (like `app-` or `output-`) that affects how flags and environment variables are named. The `schema` package provides constructors that make section creation intuitive.

Attach fields to a section using `schema.WithFields(...)` (which wraps the historical `layers.WithParameterDefinitions(...)`).

```go
func NewAppCommand() (*AppCommand, error) {
	// Create glazed section for built-in output formatting options
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	// Create app section with prefix "app-"
	// Flags become --app-verbose, --app-port, --app-host
	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithPrefix("app-"),
		schema.WithDescription("Application configuration settings"),
		schema.WithFields(
			fields.New("verbose", fields.TypeBool,
				fields.WithHelp("Enable verbose logging"),
				fields.WithDefault(false),
			),
			fields.New("port", fields.TypeInteger,
				fields.WithHelp("Server port number"),
				fields.WithDefault(8080),
			),
			fields.New("host", fields.TypeString,
				fields.WithHelp("Server host address"),
				fields.WithDefault("localhost"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create output section with prefix "output-"
	outputSection, err := schema.NewSection(
		"output",
		"Output",
		schema.WithPrefix("output-"),
		schema.WithDescription("Output formatting settings"),
		schema.WithFields(
			fields.New("format", fields.TypeChoice,
				fields.WithHelp("Output format"),
				fields.WithChoices("json", "yaml", "table"),
				fields.WithDefault("table"),
			),
			fields.New("pretty", fields.TypeBool,
				fields.WithHelp("Pretty print output"),
				fields.WithDefault(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create schema collection from sections
	schema := schema.NewSchema(
		schema.WithSections(glazedSection, appSection, outputSection),
	)

	// Create command definition
	desc := cmds.NewCommandDefinition(
		"app",
		cmds.WithShort("Application command with multiple sections"),
		cmds.WithLong("Demonstrates schema sections, field definitions, and value decoding"),
		cmds.WithSchema(schema),
	)

	return &AppCommand{CommandDefinition: desc}, nil
}
```

**Section creation patterns:**

1. **Glazed Section**: `schema.NewGlazedSchema()` adds built-in output formatting options (`--output`, `--fields`, `--sort-columns`, etc.)
2. **Custom Sections**: `schema.NewSection()` creates sections with:
   - **Slug**: Internal identifier (e.g., `"app"`)
   - **Name**: Display name (e.g., `"App"`)
   - **Prefix**: Flag prefix (e.g., `"app-"` makes flags like `--app-verbose`)
   - **Fields**: Defined using `fields.New()` with type and options
3. **Schema Collection**: `schema.NewSchema()` combines sections into a complete schema
4. **Command Definition**: `cmds.NewCommandDefinition()` creates the command with the schema attached

**Field definition options:**

- **Type**: `fields.TypeBool`, `fields.TypeInteger`, `fields.TypeString`, `fields.TypeChoice`, etc.
- **Help Text**: `fields.WithHelp("description")`
- **Default Value**: `fields.WithDefault(value)`
- **Choices**: `fields.WithChoices("option1", "option2")` for choice fields
- **Required**: `fields.WithRequired(true)` for mandatory parameters
- **Short Flag**: `fields.WithShortFlag("v")` for single-letter flags

## Step 3: Implement Command Logic

The `GlazeCommand` interface requires implementing `RunIntoGlazeProcessor`, which receives resolved values and a processor for structured output. The new API uses `*values.Values` instead of `*layers.ParsedLayers`, and provides `values.DecodeSectionInto()` for type-safe struct population.

```go
// Ensure interface compliance
var _ cmds.GlazeCommand = &AppCommand{}

func (c *AppCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	// Decode app section into AppSettings struct
	appSettings := &AppSettings{}
	if err := values.DecodeSectionInto(vals, "app", appSettings); err != nil {
		return fmt.Errorf("failed to decode app settings: %w", err)
	}

	// Decode output section into OutputSettings struct
	outputSettings := &OutputSettings{}
	if err := values.DecodeSectionInto(vals, "output", outputSettings); err != nil {
		return fmt.Errorf("failed to decode output settings: %w", err)
	}

	// Business logic using decoded settings
	if appSettings.Verbose {
		fmt.Fprintf(os.Stderr, "Starting server on %s:%d\n", appSettings.Host, appSettings.Port)
	}

	// Create structured output row
	row := types.NewRow(
		types.MRP("host", appSettings.Host),
		types.MRP("port", appSettings.Port),
		types.MRP("verbose", appSettings.Verbose),
		types.MRP("output_format", outputSettings.Format),
		types.MRP("pretty", outputSettings.Pretty),
	)

	return gp.AddRow(ctx, row)
}
```

**Value decoding pattern:**

1. **Create Settings Struct**: Instantiate the struct that matches your section
2. **Decode Section**: Use `values.DecodeSectionInto(vals, "section-slug", &settings)` to populate the struct
3. **Type Safety**: The decoder automatically converts string inputs to the correct Go types
4. **Error Handling**: Always check for decoding errors and provide context

**Key advantages:**

- **Type Safety**: Struct fields ensure compile-time type checking
- **Clear Mapping**: Section slugs (`"app"`, `"output"`) clearly map to struct types
- **Automatic Conversion**: String inputs are converted to int, bool, etc. automatically
- **Validation**: Field definitions provide validation rules (choices, required, etc.)

## Step 4: Integrate with Cobra

Glazed commands integrate with Cobra through `cli.BuildCobraCommandFromCommand()`, which handles flag registration, environment variable parsing, and output processing. The `CobraParserConfig` controls how environment variables are parsed and which middlewares are applied.

```go
func main() {
	root := &cobra.Command{
		Use:   "myapp",
		Short: "Application using new Glazed API",
	}

	// Create command
	appCmd, err := NewAppCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
		os.Exit(1)
	}

	// Build Cobra command with parser configuration
	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		appCmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			// AppName enables environment variable parsing
			// Format: <APPNAME>_<SECTION_PREFIX>_<FIELD_NAME>
			// Example: MYAPP_APP_VERBOSE=true sets app.verbose
			AppName: "myapp",
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building cobra command: %v\n", err)
		os.Exit(1)
	}

	root.AddCommand(cobraCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Environment variable format:**

When `AppName` is set to `"myapp"`, environment variables follow this pattern:
- **Format**: `<APPNAME>_<SECTION_PREFIX>_<FIELD_NAME>`
- **Section prefix**: Converted from flag prefix (hyphens become underscores, uppercase)
- **Example**: `MYAPP_APP_VERBOSE=true` sets `app.verbose` when section has prefix `"app-"`

**Precedence order (lowest to highest):**

1. **Defaults**: From field definitions (`fields.WithDefault()`)
2. **Environment Variables**: `MYAPP_APP_VERBOSE=true`
3. **Cobra Flags**: `--app-verbose=true` (highest precedence)

## Optional: Manual value resolution with `sources`

Most commands should stick with `cli.BuildCobraCommandFromCommand()`. Use the `sources` package when you want to manually assemble and execute the value-resolution chain (defaults → config files → env → flags), or when you’re not using the CLI builder.

### What `sources` provides

The `sources` package is an additive façade over `pkg/cmds/middlewares`:

- **`sources.Middleware`**: alias for the underlying middleware type
- **Source middlewares**:
  - `sources.FromDefaults(...)`
  - `sources.FromFile(path, ...)` / `sources.FromFiles(paths, ...)`
  - `sources.FromMap(m, ...)`
  - `sources.FromEnv(prefix, ...)`
  - `sources.FromCobra(cmd, ...)` / `sources.FromArgs(args, ...)`
- **Parse-step helpers** (to label/debug where values came from):
  - `sources.WithSource("env"|"flags"|"config"|...)`
  - `sources.WithMetadata(map[string]any{...})`
  - `sources.WithParseOptions(...)` (for config-file loaders)

### Ordering and precedence (important)

`sources.Execute()` runs middlewares via Glazed’s middleware engine. For the common “call next first” style middlewares, **the first middleware you pass has the highest precedence** (it runs last and can override), and the last middleware you pass has the lowest precedence.

Example precedence (lowest → highest):

- Defaults < Config file < Programmatic map < Environment < Cobra flags

To get that precedence, pass middlewares in reverse order:

```go
ms := []sources.Middleware{
	sources.FromCobra(cmd, sources.WithSource("flags")),
	sources.FromEnv("APP", sources.WithSource("env")),
	sources.FromMap(map[string]map[string]interface{}{
		"config": {
			"api-key": "custom-map-key",
			"timeout": 60,
		},
	}, sources.WithSource("custom-map")),
	sources.FromFile("config.yaml",
		sources.WithParseOptions(sources.WithSource("config-file")),
	),
	sources.FromDefaults(sources.WithSource("defaults")),
}
```

### Runnable example

See `cmd/examples/sources-example/` for a tiny program that uses the new API (`schema`, `fields`, `values`, `sources`) and loads values from:

- A config file (`--config-file=...`)
- A programmatic map override
- Environment variables (`APP_CONFIG_API_KEY=...`)
- Cobra flags (`--config-api-key=...`)

## Step 5: Working with Positional Arguments

Positional arguments are handled through the default section (slug: `schema.DefaultSlug`). Use `schema.WithArguments()` to mark fields as positional arguments (as opposed to “regular” fields added via `schema.WithFields()`, which become flags/env/config inputs).

```go
// Create default section for positional arguments
defaultSection, err := schema.NewSection(
	schema.DefaultSlug,
	"Default",
	schema.WithDescription("Default parameters"),
	schema.WithArguments(
		fields.New("input-file", fields.TypeString,
			fields.WithHelp("Input file to process"),
			fields.WithRequired(true),
		),
	),
)
if err != nil {
	return nil, err
}

// Add to schema
schema := schema.NewSchema(
	schema.WithSections(glazedSection, appSection, outputSection, defaultSection),
)
```

**Decoding positional arguments:**

```go
type DefaultSettings struct {
	InputFile string `glazed.parameter:"input-file"`
}

func (c *AppCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	// Decode default section for positional arguments
	defaultSettings := &DefaultSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, defaultSettings); err != nil {
		return fmt.Errorf("failed to decode default settings: %w", err)
	}

	// Use inputFile in business logic
	fmt.Printf("Processing file: %s\n", defaultSettings.InputFile)

	return nil
}
```

## Complete Example

Here's a complete, runnable example that demonstrates all the concepts:

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
)

// Settings structs
type AppSettings struct {
	Verbose bool   `glazed.parameter:"verbose"`
	Port    int    `glazed.parameter:"port"`
	Host    string `glazed.parameter:"host"`
}

type OutputSettings struct {
	Format string `glazed.parameter:"format"`
	Pretty bool   `glazed.parameter:"pretty"`
}

type DefaultSettings struct {
	InputFile string `glazed.parameter:"input-file"`
}

// Command struct
type ExampleCommand struct {
	*cmds.CommandDefinition
}

func NewExampleCommand() (*ExampleCommand, error) {
	// Create glazed section
	glazedSection, err := schema.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	// Create app section
	appSection, err := schema.NewSection(
		"app",
		"App",
		schema.WithPrefix("app-"),
		schema.WithDescription("Application settings"),
		schema.WithFields(
			fields.New("verbose", fields.TypeBool,
				fields.WithHelp("Enable verbose logging"),
				fields.WithDefault(false),
			),
			fields.New("port", fields.TypeInteger,
				fields.WithHelp("Server port"),
				fields.WithDefault(8080),
			),
			fields.New("host", fields.TypeString,
				fields.WithHelp("Server host"),
				fields.WithDefault("localhost"),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create output section
	outputSection, err := schema.NewSection(
		"output",
		"Output",
		schema.WithPrefix("output-"),
		schema.WithDescription("Output settings"),
		schema.WithFields(
			fields.New("format", fields.TypeChoice,
				fields.WithHelp("Output format"),
				fields.WithChoices("json", "yaml", "table"),
				fields.WithDefault("table"),
			),
			fields.New("pretty", fields.TypeBool,
				fields.WithHelp("Pretty print"),
				fields.WithDefault(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create default section for positional args
	defaultSection, err := schema.NewSection(
		schema.DefaultSlug,
		"Default",
		schema.WithDescription("Default parameters"),
		schema.WithArguments(
			fields.New("input-file", fields.TypeString,
				fields.WithHelp("Input file"),
				fields.WithRequired(true),
			),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create schema
	schema := schema.NewSchema(
		schema.WithSections(glazedSection, appSection, outputSection, defaultSection),
	)

	// Create command definition
	desc := cmds.NewCommandDefinition(
		"example",
		cmds.WithShort("Example command using new API"),
		cmds.WithLong("Demonstrates schema sections, fields, and value decoding"),
		cmds.WithSchema(schema),
	)

	return &ExampleCommand{CommandDefinition: desc}, nil
}

var _ cmds.GlazeCommand = &ExampleCommand{}

func (c *ExampleCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	// Decode all sections
	appSettings := &AppSettings{}
	if err := values.DecodeSectionInto(vals, "app", appSettings); err != nil {
		return fmt.Errorf("failed to decode app settings: %w", err)
	}

	outputSettings := &OutputSettings{}
	if err := values.DecodeSectionInto(vals, "output", outputSettings); err != nil {
		return fmt.Errorf("failed to decode output settings: %w", err)
	}

	defaultSettings := &DefaultSettings{}
	if err := values.DecodeSectionInto(vals, schema.DefaultSlug, defaultSettings); err != nil {
		return fmt.Errorf("failed to decode default settings: %w", err)
	}

	// Create output row
	row := types.NewRow(
		types.MRP("app_verbose", appSettings.Verbose),
		types.MRP("app_port", appSettings.Port),
		types.MRP("app_host", appSettings.Host),
		types.MRP("output_format", outputSettings.Format),
		types.MRP("output_pretty", outputSettings.Pretty),
		types.MRP("input_file", defaultSettings.InputFile),
	)

	return gp.AddRow(ctx, row)
}

func main() {
	root := &cobra.Command{
		Use:   "example",
		Short: "Example using new Glazed API",
	}

	cmd, err := NewExampleCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cobraCmd, err := cli.BuildCobraCommandFromCommand(
		cmd,
		cli.WithParserConfig(cli.CobraParserConfig{
			AppName: "example",
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root.AddCommand(cobraCmd)

	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Testing Your Command

Test your command with various input sources to verify precedence and parsing:

```bash
# Build
go build -o example

# Test with defaults
./example example input.txt

# Test with environment variables
EXAMPLE_APP_VERBOSE=true EXAMPLE_APP_PORT=9090 ./example example input.txt

# Test with Cobra flags (overrides env)
EXAMPLE_APP_VERBOSE=true ./example example --app-verbose=false --app-port=3000 input.txt

# Test output formats
./example example input.txt --output json
./example example input.txt --output yaml
```

**Expected behavior:**

1. **Defaults**: Without flags, uses default values from field definitions
2. **Environment Variables**: `EXAMPLE_APP_VERBOSE=true` sets `app.verbose=true`
3. **Flag Override**: `--app-verbose=false` overrides environment variable
4. **Output Formats**: `--output json` produces JSON output automatically

## Appendix: Manual source chains with `sources.Execute`

Most commands should use `cli.BuildCobraCommandFromCommand()`, which wires parsing + output processing for you. Use `sources.Execute()` when you want to manually control the *value resolution chain* (defaults → config files → env → flags) or when you’re not using the CLI builder.

### Precedence ordering (important)

`sources.Execute()` uses Glazed’s middleware engine under the hood. For the common “call next first” middlewares, **the first middleware you pass has the highest precedence** (it runs last and can override), and the last middleware you pass has the lowest precedence.

Example precedence (lowest → highest):

- Defaults < Config file < Programmatic map < Environment < Cobra flags

To achieve that precedence, pass middlewares in the *reverse* order:

```go
ms := []sources.Middleware{
    sources.FromCobra(cmd, sources.WithSource("flags")),
    sources.FromEnv("APP", sources.WithSource("env")),
    sources.FromMap(map[string]map[string]interface{}{
        "config": {"api-key": "custom-map-key", "timeout": 60},
    }, sources.WithSource("custom-map")),
    sources.FromFile("config.yaml",
        sources.WithParseOptions(sources.WithSource("config-file")),
    ),
    sources.FromDefaults(sources.WithSource("defaults")),
}
```

### Tiny runnable example

See the example program at `cmd/examples/sources-example/`. It defines a small schema using the new API (`schema`, `fields`, `values`, `sources`), then resolves values from:

- A config file (`--config-file=...`)
- A programmatic map override
- Environment variables (`APP_CONFIG_API_KEY=...`)
- Cobra flags (`--config-api-key=...`)

Run it:

```bash
go run ./cmd/examples/sources-example --config-file=cmd/examples/sources-example/config.yaml
APP_CONFIG_API_KEY=env-key go run ./cmd/examples/sources-example --config-file=cmd/examples/sources-example/config.yaml
APP_CONFIG_API_KEY=env-key go run ./cmd/examples/sources-example --config-file=cmd/examples/sources-example/config.yaml --config-api-key=flag-key
```

## Best Practices

### Schema Organization

**Group Related Parameters**: Create sections that group logically related parameters. For example, all database connection settings belong in a `database` section, not scattered across multiple sections.

**Use Meaningful Prefixes**: Section prefixes should be short but descriptive. `app-` is better than `a-` or `application-configuration-`.

**Leverage Glazed Section**: Always include `schema.NewGlazedSchema()` to get built-in output formatting options without defining them yourself.

### Field Definitions

**Provide Defaults**: Always provide sensible defaults for optional parameters. This makes commands easier to use and reduces required flags.

**Use Appropriate Types**: Choose the right field type:
- `fields.TypeBool` for true/false flags
- `fields.TypeInteger` for numeric values
- `fields.TypeChoice` for limited options
- `fields.TypeString` for free-form text

**Write Clear Help Text**: Help text appears in `--help` output. Make it concise but informative.

### Value Decoding

**Decode Per Section**: Decode each section into its own struct. This keeps your code organized and makes it clear which parameters belong together.

**Handle Errors Gracefully**: Always check `DecodeSectionInto()` errors and provide context about which section failed.

**Use Type-Safe Structs**: Avoid accessing values as `map[string]interface{}`. Use structs with `glazed.parameter` tags for type safety and better IDE support.

## Migration from Old API

If you have existing commands using the old API (`layers.ParameterLayer`, `parameters.ParameterDefinition`), you can migrate gradually. In the new vocabulary, “parameter definitions” are called **fields** (hence `schema.WithFields(...)`).

**Old API:**
```go
layer, err := layers.NewParameterLayer("app", "App",
	layers.WithPrefix("app-"),
	layers.WithParameterDefinitions(
		parameters.NewParameterDefinition("verbose", parameters.ParameterTypeBool, ...),
	),
)
```

**New API:**
```go
section, err := schema.NewSection("app", "App",
	schema.WithPrefix("app-"),
	schema.WithFields(
		fields.New("verbose", fields.TypeBool, ...),
	),
)
```

**Key changes:**
- `layers.NewParameterLayer` → `schema.NewSection`
- `parameters.NewParameterDefinition` → `fields.New`
- `layers.ParameterLayers` → `schema.Schema`
- `layers.ParsedLayers` → `values.Values`
- `parsedLayers.InitializeStruct` → `values.DecodeSectionInto`

The underlying types are aliases, so old code continues to work. You can migrate incrementally, updating one command at a time.

## Next Steps

### Learn More

```
glaze help build-first-command
```

See the complete tutorial for building your first Glazed command with detailed examples.

```
glaze help layers-guide
```

Understand how parameter layers organize reusable configuration sets.

### Explore Examples

The `glazed/cmd/examples/refactor-new-packages/` directory contains a complete example demonstrating:
- Multiple schema sections
- Environment variable parsing
- Cobra flag integration
- Value decoding into structs
- Precedence demonstration

### Advanced Topics

- **Custom Field Types**: Create domain-specific field types for your application
- **Dynamic Schemas**: Build schemas programmatically based on configuration
- **Value Sources**: Use `sources` package to create custom value resolution middleware
- **Schema Validation**: Add validation rules beyond type checking

The new API vocabulary makes Glazed commands easier to understand and maintain, while maintaining full backward compatibility with existing code.

