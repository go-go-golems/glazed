# Refactor New Packages Example

This example demonstrates the new wrapper packages (`schema`, `fields`, `values`, `sources`) introduced in ticket `001-REFACTOR-NEW-PACKAGES`.

## Purpose

Show how to use the new vocabulary packages to:
- Define schema sections using `schema.NewSection()`
- Define field definitions using `fields.New()`
- Decode resolved values into structs using `values.DecodeSectionInto()`
- Demonstrate env + cobra parsing with proper precedence

## Usage

### Basic usage with defaults

```bash
go run ./cmd/examples/refactor-new-packages refactor-demo input.txt
```

Output shows default values:
- `app-verbose`: false
- `app-port`: 8080
- `app-host`: localhost
- `output-format`: table
- `output-pretty`: true

### Environment variable parsing

Set environment variables with the `DEMO_` prefix:

```bash
DEMO_APP_VERBOSE=true DEMO_APP_PORT=9090 go run ./cmd/examples/refactor-new-packages refactor-demo input.txt
```

Environment variable format:
- Global prefix: `DEMO_` (from `AppName: "demo"`)
- Section prefix: `APP_` (from `schema.WithPrefix("app-")`)
- Field name: `VERBOSE` (from field name `verbose`)
- Full key: `DEMO_APP_VERBOSE=true`

### Cobra flags override env (precedence)

Cobra flags have highest precedence and override environment variables:

```bash
DEMO_APP_VERBOSE=true go run ./cmd/examples/refactor-new-packages refactor-demo --app-verbose=false --app-port=3000 input.txt
```

Result: `app-verbose=false` (flag overrides env), `app-port=3000` (flag overrides env)

## Precedence Order

1. **Defaults** (lowest precedence) - from field definitions
2. **Environment variables** - `DEMO_<SECTION>_<FIELD>=value`
3. **Cobra flags** (highest precedence) - `--section-field=value`

## Schema Sections

The example defines three schema sections:

1. **Default section** (`default`) - for positional arguments
   - `input-file` (required positional arg)

2. **App section** (`app`) - with prefix `app-`
   - `verbose` (bool, default: false)
   - `port` (int, default: 8080)
   - `host` (string, default: localhost)

3. **Output section** (`output`) - with prefix `output-`
   - `format` (choice: json/yaml/table, default: table)
   - `pretty` (bool, default: true)

## Code Structure

- Uses `schema.NewSection()` to create schema sections
- Uses `fields.New()` to define fields with types and options
- Uses `values.DecodeSectionInto()` to decode resolved values into structs
- Uses `cli.BuildCobraCommand()` with `CobraParserConfig.AppName` for env parsing

## Related

- Design doc: `glazed/ttmp/2025/12/17/001-REFACTOR-NEW-PACKAGES--refactor-add-schema-fields-values-sources-wrapper-packages-example-program/design-doc/01-design-wrapper-packages-schema-fields-values-sources.md`
- Implementation plan: `glazed/ttmp/2025/12/17/001-REFACTOR-NEW-PACKAGES--refactor-add-schema-fields-values-sources-wrapper-packages-example-program/planning/01-implementation-plan-wrapper-packages-example-program.md`

