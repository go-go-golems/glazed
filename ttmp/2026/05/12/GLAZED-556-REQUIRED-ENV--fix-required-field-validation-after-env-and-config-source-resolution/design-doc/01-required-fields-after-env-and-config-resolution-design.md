---
Title: Required fields after env and config resolution design
Ticket: GLAZED-556-REQUIRED-ENV
Status: active
Topics:
    - glazed
    - cli
    - config
    - env
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cli/cobra-parser.go
      Note: Builds Cobra parser source middleware chain and should call final required validation after Execute
    - Path: pkg/cli/cobra_parser_config_test.go
      Note: Existing parser config tests and recommended home for regression tests
    - Path: pkg/cmds/fields/cobra.go
      Note: Contains current source-local required check that rejects missing Cobra flags before env/config can apply
    - Path: pkg/cmds/schema/section-impl.go
      Note: Calls GatherFlagsFromCobraCommand with ignoreRequired=false and contains TODO about moving required checks higher
    - Path: pkg/cmds/sources/load-fields-from-config.go
      Note: Implements config plan source loading that should also satisfy required final values
    - Path: pkg/cmds/sources/middlewares.go
      Note: Documents middleware execution ordering and precedence model
    - Path: pkg/cmds/sources/update.go
      Note: Implements FromEnv/updateFromEnv env key derivation and parse-step provenance
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/556
Summary: Design and implementation guide for fixing required field validation so env/config-backed values can satisfy fields.WithRequired(true).
LastUpdated: 2026-05-12T14:15:00-04:00
WhatFor: Use when implementing or reviewing GitHub issue 556 in Glazed's Cobra parser and source middleware pipeline.
WhenToUse: Read before changing required field semantics, Cobra parsing, environment variable loading, config-file loading, or parsed-field diagnostics.
---


# Required fields after env and config resolution design

## Executive summary

GitHub issue 556 reports that a Glazed field declared with `fields.WithRequired(true)` fails when the value is supplied only by an environment variable. The environment variable is valid and is parsed correctly for optional fields, but the Cobra source rejects the command before the env/config middlewares can merge their values into `values.Values`.

The core fix is to move required-field validation out of source-specific parsing and into an explicit final validation step that runs after the configured source chain has finished. Source readers such as Cobra, args, config, and env should answer the question "what values did this source provide?" A separate validator should answer "after all sources have been merged, which required values are still missing or empty?"

For a new intern: this is a parser pipeline ordering bug, not an environment variable naming bug. The env loader derives the right key and can parse the right value. The failure happens earlier, inside Cobra flag gathering, because `GatherFlagsFromCobraCommand` treats a required Glazed field as a required Cobra flag.

## Problem statement and scope

### User-visible problem

A command wants to declare a required value that may be supplied by any configured source:

```go
fields.New(
    "proc-file",
    fields.TypeString,
    fields.WithDefault(""),
    fields.WithRequired(true),
    fields.WithHelp("Process state JSON path; can also be set with DEVMUX_PROC_FILE"),
)
```

The parser is built with the env-aware path:

```go
cli.NewCobraParserFromSections(schema_, &cli.CobraParserConfig{
    ShortHelpSections: []string{schema.DefaultSlug},
    AppName:           "devmux",
})
```

Expected behavior:

```bash
DEVMUX_PROC_FILE=/tmp/procs.json devmux proc list --print-parsed-fields
```

should succeed and report the env source in the parsed-field log.

Actual behavior:

```text
Field proc-file is required
```

### Intended semantics

`fields.WithRequired(true)` should mean:

> after all configured value sources have been applied, the field must have a meaningful value.

It should not mean:

> the corresponding Cobra flag must have been explicitly typed on the command line.

This matters because Glazed already has a layered value-resolution model:

1. defaults,
2. config files,
3. environment variables,
4. positional arguments,
5. Cobra flags,
6. validation,
7. decode/run.

### In scope

- Cobra parser path in `pkg/cli/cobra-parser.go`.
- Cobra source middleware in `pkg/cmds/sources/cobra.go`.
- Required checks inside flag/argument gathering in `pkg/cmds/fields` and section parsing in `pkg/cmds/schema`.
- Final validation middleware or helper that checks required fields against final merged `values.Values`.
- Regression tests for env-backed and config-backed required fields.
- Documentation updates for required-value semantics.

### Out of scope

- Redesigning the entire field type system.
- Replacing Cobra.
- Changing environment variable derivation rules.
- Adding application-specific post-decode validation helpers as the primary fix.

## Current-state architecture with evidence

### The parser builds a source middleware chain

`pkg/cli/cobra-parser.go` defines `CobraParserConfig`. `AppName` controls environment loading in the built-in parser path; `ConfigPlanBuilder` controls explicit config-file loading. The code comments state that a custom `MiddlewaresFunc` replaces the built-in chain entirely, so callers using a custom chain must add any sources they need themselves.

Evidence:

- `pkg/cli/cobra-parser.go:91-104` defines `CobraParserConfig`, including `MiddlewaresFunc`, `AppName`, and `ConfigPlanBuilder`.
- `pkg/cli/cobra-parser.go:143-185` builds the built-in chain when `MiddlewaresFunc == nil`.
- `pkg/cli/cobra-parser.go:149-183` appends middlewares in reverse precedence order: Cobra, args, env, config, defaults.
- `pkg/cli/cobra-parser.go:253-272` runs `cmd_sources.Execute(c.Sections, parsedSections, middlewares_...)`.

Simplified architecture:

```text
Cobra command execution
        |
        v
CobraParser.Parse(cmd,args)
        |
        +--> ParseCommandSettingsSection(cmd)
        |
        +--> c.middlewaresFunc(...)
        |       returns source middlewares
        |
        +--> sources.Execute(schema, values, middlewares...)
                |
                +--> defaults
                +--> config plan files
                +--> env vars
                +--> positional args
                +--> cobra flags
                |
                v
             final values.Values
```

The order can be confusing because the code appends highest precedence first, then `Execute` reverses the middleware slice. `pkg/cmds/sources/middlewares.go:37-56` documents this convention: when middlewares call `next` first, values are modified in source-precedence order. With `[FromCobra, FromArgs, FromEnv, FromDefaults]`, defaults run before env, env before args, and args before Cobra.

### Environment loading is source-specific and works independently

`pkg/cmds/sources/update.go:143-211` implements `updateFromEnv`. For every schema section and field, it derives an env key from section prefix plus field name, converts hyphens to underscores, uppercases the key, prepends the app prefix, parses the string through the field parser, and updates `sectionValues.Fields` with the parsed value and parse log.

Important details:

- `pkg/cmds/sources/update.go:156-160` derives the env key.
- `pkg/cmds/sources/update.go:162-168` records `env_key` metadata and `source: env`.
- `pkg/cmds/sources/update.go:193-199` parses with `p.ParseField` and preserves parse logs through `UpdateWithLog`.
- `pkg/cmds/sources/update_test.go:14-75` verifies typed env parsing for bools, integers, floats, dates, choices, lists, key-values, and strings.

For issue 556, this means env parsing is not the failing component. The issue reproduction and the optional-field test both show that `REQ_ENV_TEST_REQUIRED_NAME` is derivable and parseable. The required check prevents the env middleware from running far enough to apply the value.

### Config file loading has the same conceptual issue

`pkg/cmds/sources/load-fields-from-config.go:123-159` defines `FromConfigPlanBuilder`, which runs after lower-precedence values exist and loads resolved config files into `values.Values`. `pkg/cli/cobra_parser_config_test.go:52-75` verifies that `ConfigPlanBuilder` can load a file into a section.

Because the current failure is raised by the Cobra source before later source middlewares can finish, config-backed required fields are vulnerable to the same bug as env-backed required fields.

### Cobra flag gathering currently enforces required too early

`pkg/cmds/fields/cobra.go:413-441` is the key failure site. When a flag was not changed on the command line, and the field definition is required, `GatherFlagsFromCobraCommand` returns `Field %s is required` unless `ignoreRequired` is true.

```text
if !cmd.Flags().Changed(flagName) {
    if pd.Required {
        if ignoreRequired {
            return nil
        }
        return errors.Errorf("Field %s is required", pd.Name)
    }
    ...
}
```

`pkg/cmds/schema/section-impl.go:236-249` calls this function with `ignoreRequired=false`. There is even a TODO at `pkg/cmds/schema/section-impl.go:241-242` noting that required checks probably need to move to a higher-level middleware because Glazed no longer relies only on Cobra.

That TODO matches the issue exactly.

### Reproduction script confirms the behavior

The ticket script `scripts/02-reproduce-required-env-parser.sh` injects a temporary `pkg/cli/required_env_repro_test.go` and runs only the reproduction tests. It creates two fields:

- `required-name`: `fields.WithRequired(true)`, env var set, no flag.
- `optional-name`: no required marker, env var set, no flag.

Observed result stored in `scripts/evidence/02-reproduce-required-env-parser.log`:

```text
=== RUN   TestReproIssue556RequiredEnvBackedField
    required_env_repro_test.go:37: BUG REPRODUCED: required env-backed field failed before env could satisfy it: Field required-name is required
--- FAIL: TestReproIssue556RequiredEnvBackedField (0.00s)
=== RUN   TestReproIssue556OptionalEnvBackedField
--- PASS: TestReproIssue556OptionalEnvBackedField (0.00s)
```

This isolates the bug to the interaction between `Required` and source ordering.

## Core concepts for interns

### Schema

A `schema.Schema` is the collection of command sections. Each section has field definitions. A section can represent default command flags, output options, profile options, or custom grouped settings.

API references:

- `schema.NewSchema(schema.WithSections(...))`
- `schema.NewSection(slug, name, schema.WithPrefix(...), schema.WithFields(...))`
- `schema.Section.GetDefinitions()`
- `schema.CobraSection.ParseSectionFromCobraCommand(...)`

### Field definition

A `fields.Definition` describes one logical setting:

- name,
- type,
- default,
- required marker,
- help text,
- choices,
- whether it is a positional argument.

API references:

- `fields.New("name", fields.TypeString, ...)`
- `fields.WithDefault(value)`
- `fields.WithRequired(true)`
- `fields.Definition.ParseField(inputs, options...)`
- `fields.Definition.CheckValueValidity(value)`

Important rule for this ticket: `Required` is metadata on the logical value, not on a single source.

### Source middleware

A source middleware applies one source of values into `values.Values`:

- `sources.FromDefaults(...)`
- `sources.FromConfigPlanBuilder(...)`
- `sources.FromEnv(prefix, ...)`
- `sources.FromArgs(args, ...)`
- `sources.FromCobra(cmd, ...)`

Each source should be allowed to provide zero or more values. Missing values in one source are normal because a later or earlier source may provide them.

### Parsed values and provenance

`values.Values` stores sections. Each section stores `fields.FieldValues`. Each field value carries a `Log []ParseStep` so `--print-parsed-fields` can show how the final value was obtained.

API references:

- `values.New()`
- `values.Values.GetOrCreate(section)`
- `values.Values.GetField(sectionSlug, key)`
- `fields.FieldValues.Merge(...)`
- `fields.FieldValues.UpdateWithLog(...)`
- `cli.PrintParsedFields(writer, parsedValues)`

The fix must preserve provenance. If env supplies a required value, its log should still show `source: env` and `env_key: ...`.

## Gap analysis

### Gap 1: required validation is source-local

The current code checks required-ness in `GatherFlagsFromCobraCommand`, which only knows whether the Cobra flag was explicitly changed. It cannot know whether config or env will supply the value.

### Gap 2: required validation is not represented as a middleware

The source chain has middlewares for defaults, config, env, args, and flags. It lacks an explicit final `ValidateRequired` middleware or equivalent post-execute function.

### Gap 3: custom middleware chains need a compatible story

`CobraCommandDefaultMiddlewares` is a public helper. Some callers may use it directly. If the built-in parser path adds validation but custom chains do not, then behavior diverges. The safest design is to provide a reusable middleware/helper and use it in the standard chains.

### Gap 4: empty-string defaults need a clear policy

The issue example has `fields.WithDefault("")` plus `fields.WithRequired(true)`. A final validator that merely checks field presence would incorrectly accept the empty default. The desired behavior is "non-empty value" for strings and probably non-empty collections for list-like fields.

## Proposed architecture

### Design principle

Separate source collection from final validation.

```text
Before
------
Cobra source:
  if required flag not explicitly changed:
      error
  parse flag
Env source:
  never reached when Cobra source errors

After
-----
Cobra source:
  if flag explicitly changed:
      parse flag
  else:
      provide nothing
Env/config/defaults/args:
  provide whatever they have
Final required validator:
  for every required definition:
      inspect final values.Values
      error if missing or empty
```

### Proposed pipeline diagram

```text
           lower precedence                              higher precedence
┌──────────┐   ┌─────────────┐   ┌──────────┐   ┌──────┐   ┌──────────┐   ┌───────────────┐
│ defaults │-->| config file │-->| env vars │-->| args │-->| cobra CLI │-->| validate req. │
└──────────┘   └─────────────┘   └──────────┘   └──────┘   └──────────┘   └───────────────┘
      |               |                |            |            |                 |
      v               v                v            v            v                 v
                         merged values.Values with parse-step provenance
```

### New API: required-value validation helper

Add a reusable helper in `pkg/cmds/sources` or `pkg/cmds/values`.

Recommended location: `pkg/cmds/sources/validate.go`, because it can be exposed as a source middleware and used by parser chains.

Sketch:

```go
package sources

func ValidateRequired(options ...ValidateRequiredOption) Middleware {
    cfg := defaultValidateRequiredConfig(options...)
    return func(next HandlerFunc) HandlerFunc {
        return func(schema_ *schema.Schema, parsedValues *values.Values) error {
            if err := next(schema_, parsedValues); err != nil {
                return err
            }
            return ValidateRequiredValues(schema_, parsedValues, cfg)
        }
    }
}

func ValidateRequiredValues(schema_ *schema.Schema, parsed *values.Values, cfg validateRequiredConfig) error {
    var missing []string
    err := schema_.ForEachE(func(sectionKey string, section schema.Section) error {
        sv, sectionExists := parsed.Get(section.GetSlug())
        return section.GetDefinitions().ForEachE(func(def *fields.Definition) error {
            if !def.Required {
                return nil
            }
            if def.IsArgument && cfg.SkipArguments { // optional policy knob; see below
                return nil
            }
            if !sectionExists {
                missing = append(missing, displayName(section, def))
                return nil
            }
            fv, ok := sv.Fields.Get(def.Name)
            if !ok || isRequiredValueEmpty(def, fv.Value) {
                missing = append(missing, displayName(section, def))
            }
            return nil
        })
    })
    if err != nil {
        return err
    }
    if len(missing) > 0 {
        return fmt.Errorf("missing required field(s): %s", strings.Join(missing, ", "))
    }
    return nil
}
```

### Empty-value policy

The validator needs a field-type-aware emptiness function.

Sketch:

```go
func isRequiredValueEmpty(def *fields.Definition, value interface{}) bool {
    if value == nil {
        return true
    }

    switch def.Type {
    case fields.TypeString, fields.TypeSecret, fields.TypeFile,
         fields.TypeStringFromFile, fields.TypeChoice:
        s, ok := value.(string)
        return ok && strings.TrimSpace(s) == ""

    default:
        rv := reflect.ValueOf(value)
        switch rv.Kind() {
        case reflect.Slice, reflect.Array, reflect.Map:
            return rv.Len() == 0
        case reflect.Pointer, reflect.Interface:
            return rv.IsNil()
        default:
            return false
        }
    }
}
```

The design deliberately treats numeric zero and boolean false as present values. A required integer may legitimately be `0`, and a required boolean may legitimately be `false`. If a command needs stricter semantic validation, it should use command-specific validation after decode.

### Cobra source should ignore required while collecting source values

Change `SectionImpl.ParseSectionFromCobraCommand` so source parsing does not enforce required values:

Current call:

```go
ps, err := p.Definitions.GatherFlagsFromCobraCommand(
    cmd, true, false, p.Prefix,
    options...,
)
```

Proposed call:

```go
ps, err := p.Definitions.GatherFlagsFromCobraCommand(
    cmd, true, true, p.Prefix,
    options...,
)
```

This is the smallest local change that stops Cobra from failing before env/config. It should be paired with final validation; otherwise missing required flags might silently pass.


### Control and diagnostic paths skip required validation

Required validation should run only when Glazed is about to execute the command's normal business logic. It should not block help or parser diagnostics. Two cases matter:

- `--help` / `-h`: the user is asking Cobra to explain the command, not to execute it. Cobra usually handles this before `RunE`, but `CobraParser.Parse` should still be defensive for tests and custom wiring.
- `--print-parsed-fields`: the user is asking Glazed to show the partial merged parser state and provenance. This mode is specifically useful when a required field is missing or coming from the wrong source, so required validation must not prevent output.

The final policy is:

```text
if help flag was requested:
    skip required validation
else if command settings say print-parsed-fields:
    skip required validation
else:
    validate required fields against final merged values
```

Recommended helper in `pkg/cli/cobra-parser.go`:

```go
func shouldValidateRequiredFields(cmd *cobra.Command, parsedCommandSections *values.Values) (bool, error) {
    if isHelpRequested(cmd) {
        return false, nil
    }

    commandSettings := &CommandSettings{}
    if err := parsedCommandSections.DecodeSectionInto(CommandSettingsSlug, commandSettings); err != nil {
        return false, err
    }
    if commandSettings.PrintParsedFields {
        return false, nil
    }
    return true, nil
}
```

Use this helper after source execution and before `ValidateRequiredValues`.

### Add final validation to standard parser chains

Update both standard middleware constructors in `pkg/cli/cobra-parser.go`:

1. `CobraCommandDefaultMiddlewares`
2. the built-in env/config-aware chain inside `NewCobraParserFromSections`

Because middlewares call `next` first and then update values, the final validator must be arranged so it runs after all sources have updated `parsedValues`. There are two safe implementation options.

Option A: execute validation after `cmd_sources.Execute` inside `CobraParser.Parse`:

```go
err = cmd_sources.Execute(c.Sections, parsedSections, middlewares_...)
if err != nil {
    return nil, err
}
if err := cmd_sources.ValidateRequiredValues(c.Sections, parsedSections, cmd_sources.DefaultRequiredValidation()); err != nil {
    return nil, err
}
return parsedSections, nil
```

Option B: add a final middleware whose execution ordering is tested carefully:

```go
middlewares_ := []cmd_sources.Middleware{
    cmd_sources.ValidateRequired(),
    cmd_sources.FromCobra(cmd, fields.WithSource("cobra")),
    cmd_sources.FromArgs(args, fields.WithSource("arguments")),
    cmd_sources.FromEnv(envPrefix, fields.WithSource("env")),
    cmd_sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
}
```

Option A is easier to reason about because it happens after `Execute` returns. Option B is elegant but easier to mis-order because of `Execute`'s reverse wrapping behavior. For an intern implementation, start with Option A.

### Config-backed required value flow

A config-backed required field should follow the same final validation path:

```text
required field missing from CLI flag
        |
        v
FromCobra provides no value and does not error
        |
        v
FromConfigPlanBuilder loads config file
        |
        v
ValidateRequiredValues sees final field exists and is non-empty
        |
        v
parse succeeds
```

### Env-backed required value flow

```text
REQ_ENV_TEST_REQUIRED_NAME=from-env
        |
        v
FromDefaults may provide empty default
        |
        v
FromEnv derives REQ_ENV_TEST_REQUIRED_NAME and parses string
        |
        v
FromCobra sees flag not changed, provides no value, does not error
        |
        v
ValidateRequiredValues sees final value "from-env"
        |
        v
parse succeeds and parsed-field log includes source env
```

## Implementation guide

### Phase 1: add failing regression tests

Add tests in `pkg/cli/cobra_parser_config_test.go` because this bug is about the public parser configuration path.

Test 1: env satisfies required field.

```go
func TestCobraParserRequiredFieldCanBeSatisfiedFromEnv(t *testing.T) {
    section, err := schema.NewSection(
        schema.DefaultSlug,
        "Default",
        schema.WithFields(fields.New(
            "required-name",
            fields.TypeString,
            fields.WithRequired(true),
        )),
    )
    require.NoError(t, err)

    parser, err := NewCobraParserFromSections(schema.NewSchema(schema.WithSections(section)), &CobraParserConfig{
        ShortHelpSections:          []string{schema.DefaultSlug},
        SkipCommandSettingsSection: true,
        AppName:                    "REQ_ENV_TEST",
    })
    require.NoError(t, err)

    cmd := &cobra.Command{Use: "probe"}
    require.NoError(t, parser.AddToCobraCommand(cmd))
    t.Setenv("REQ_ENV_TEST_REQUIRED_NAME", "from-env")

    parsed, err := parser.Parse(cmd, nil)
    require.NoError(t, err)

    fv, ok := parsed.GetField(schema.DefaultSlug, "required-name")
    require.True(t, ok)
    require.Equal(t, "from-env", fv.Value)
    require.Contains(t, fv.Log[len(fv.Log)-1].Source, "env")
}
```

Test 2: missing required still fails.

```go
func TestCobraParserRequiredFieldMissingStillFails(t *testing.T) {
    // same section, same parser, no env, no flag
    _, err := parser.Parse(cmd, nil)
    require.Error(t, err)
    require.Contains(t, err.Error(), "required-name")
}
```

Test 3: empty-string default does not satisfy required.

```go
fields.New(
    "required-name",
    fields.TypeString,
    fields.WithDefault(""),
    fields.WithRequired(true),
)
```

Expected: parse fails unless env/config/flag supplies a non-empty value.

Test 4: explicit flag still satisfies required and overrides env.

```go
t.Setenv("REQ_ENV_TEST_REQUIRED_NAME", "from-env")
cmd.SetArgs([]string{"--required-name", "from-flag"})
// execute cobra or parse after flags are parsed
// expect final value from-flag and source cobra after env in log
```

Test 5: config satisfies required field.

Reuse the existing `ConfigPlanBuilder` pattern from `pkg/cli/cobra_parser_config_test.go:52-75`, but make `api-key` required and omit any CLI flag.

### Phase 2: make source parsing non-fatal for required flags

Change `pkg/cmds/schema/section-impl.go`:

```diff
- cmd, true, false, p.Prefix,
+ cmd, true, true, p.Prefix,
```

Then run the regression tests. At this point env/config tests may pass, but the "missing required still fails" test will probably fail because no final validator exists yet. That is expected.

### Phase 3: implement final required validator

Add `pkg/cmds/sources/validate_required.go`.

Minimum API:

```go
func ValidateRequiredValues(schema_ *schema.Schema, parsedValues *values.Values) error
```

Optional middleware API:

```go
func ValidateRequired() Middleware
```

Recommended error format:

```text
required field required-name is missing
```

or for multiple fields:

```text
required fields are missing: default.required-name, auth.token
```

The exact message is less important than preserving field names and making tests stable.

Pseudocode:

```go
func ValidateRequiredValues(schema_ *schema.Schema, parsed *values.Values) error {
    missing := []string{}
    schema_.ForEachE(func(sectionKey string, section schema.Section) error {
        sv, _ := parsed.Get(section.GetSlug())
        section.GetDefinitions().ForEachE(func(def *fields.Definition) error {
            if !def.Required { return nil }
            if sv == nil {
                missing = append(missing, qualified(section, def))
                return nil
            }
            fv, ok := sv.Fields.Get(def.Name)
            if !ok || requiredValueEmpty(def, fv.Value) {
                missing = append(missing, qualified(section, def))
            }
            return nil
        })
        return nil
    })
    if len(missing) > 0 { return fmt.Errorf(...) }
    return nil
}
```

### Phase 4: call the validator after source execution

Modify `pkg/cli/cobra-parser.go` in `CobraParser.Parse`:

```go
parsedSections := values.New()
err = cmd_sources.Execute(c.Sections, parsedSections, middlewares_...)
if err != nil {
    return nil, err
}

validateRequired, err := shouldValidateRequiredFields(cmd, parsedCommandSections)
if err != nil {
    return nil, err
}
if validateRequired {
    if err := cmd_sources.ValidateRequiredValues(c.Sections, parsedSections); err != nil {
        return nil, err
    }
}

return parsedSections, nil
```

This makes the standard `CobraParser` path correct while still allowing help and `--print-parsed-fields` to inspect commands with missing required values. If there are direct callers of `cmd_sources.Execute` outside `pkg/cli`, they can opt into validation by calling the helper explicitly.

### Phase 5: audit other early required checks

Search results show related functions:

- `pkg/cmds/fields/gather-arguments.go` checks required positional arguments.
- `pkg/cmds/fields/gather-fields.go` checks required map fields when `onlyProvided` is false.
- `pkg/cmds/fields/parse.go` checks required when no raw inputs are passed.
- `pkg/cmds/fields/strings.go` checks required for string-list parsing.

Do not blindly remove all of them. Some APIs are lower-level parsers where callers may still expect strict source-local behavior. For this issue, the key public path is `CobraParser.Parse`; change only the code path needed for source middleware composition, then add tests to prevent regressions.

### Phase 6: update docs

Update required-value semantics in documentation:

- `pkg/doc/topics/24-config-files.md` because it describes `AppName`, env overrides, and config loading.
- `pkg/doc/topics/16-parsing-fields.md` because it explains required fields.
- possibly `pkg/doc/topics/21-cmds-middlewares.md` because source middleware ordering is central to the behavior.

Suggested wording:

> In the Cobra parser path, `fields.WithRequired(true)` validates the final merged value after defaults, config files, environment variables, positional args, and flags have been applied. It does not require the value to be supplied as a Cobra flag. Use command-specific validation if you need a stricter business rule than non-empty final value.

## API references

### Parser construction

```go
parser, err := cli.NewCobraParserFromSections(schema_, &cli.CobraParserConfig{
    ShortHelpSections:          []string{schema.DefaultSlug},
    SkipCommandSettingsSection: true,
    AppName:                    "REQ_ENV_TEST",
    ConfigPlanBuilder:          optionalBuilder,
})
```

Key files:

- `pkg/cli/cobra-parser.go`
- `pkg/cli/cobra_parser_config_test.go`

### Field declaration

```go
fields.New(
    "required-name",
    fields.TypeString,
    fields.WithRequired(true),
    fields.WithHelp("Can be supplied by --required-name or REQ_ENV_TEST_REQUIRED_NAME"),
)
```

Key files:

- `pkg/cmds/fields/definitions.go`
- `pkg/cmds/fields/parse.go`
- `pkg/cmds/fields/field-value.go`

### Source middlewares

```go
[]sources.Middleware{
    sources.FromCobra(cmd, fields.WithSource("cobra")),
    sources.FromArgs(args, fields.WithSource("arguments")),
    sources.FromEnv("REQ_ENV_TEST", fields.WithSource("env")),
    sources.FromDefaults(fields.WithSource(fields.SourceDefaults)),
}
```

Key files:

- `pkg/cmds/sources/cobra.go`
- `pkg/cmds/sources/update.go`
- `pkg/cmds/sources/load-fields-from-config.go`
- `pkg/cmds/sources/middlewares.go`

### Final validation

Proposed:

```go
if err := sources.ValidateRequiredValues(schema_, parsedValues); err != nil {
    return nil, err
}
```

Key new file:

- `pkg/cmds/sources/validate_required.go`

## Testing and validation strategy

### Unit tests

Run targeted tests first:

```bash
go test ./pkg/cli -run 'TestCobraParser.*Required|TestReproIssue556' -count=1 -v
```

Then run source package tests:

```bash
go test ./pkg/cmds/sources ./pkg/cmds/schema ./pkg/cmds/fields -count=1
```

Then run the full repository tests if time allows:

```bash
go test ./... -count=1
```

### Manual smoke test with parsed-field output

Create or use a small command with:

```go
fields.New("required-name", fields.TypeString, fields.WithRequired(true))
```

Run:

```bash
REQ_ENV_TEST_REQUIRED_NAME=from-env probe --print-parsed-fields
```

Expected YAML shape:

```yaml
default:
  required-name:
    log:
      - metadata:
          env_key: REQ_ENV_TEST_REQUIRED_NAME
        source: env
        value: from-env
    value: from-env
```

### Regression matrix

| Case | Sources configured | Value supplied by | Expected |
|---|---|---|---|
| required missing | defaults only or no sources | nowhere | error |
| required missing + print-parsed-fields | command settings | nowhere | success; print partial parsed values |
| required missing + help | Cobra help | nowhere | success; print help |
| required empty default | defaults | `""` | error |
| required env | AppName/env | env var | success |
| required config | ConfigPlanBuilder | config file | success |
| required flag | Cobra | `--required-name` | success |
| env + flag | AppName/env + Cobra | both | flag wins |
| optional env | AppName/env | env var | success |
| invalid env value | AppName/env | parse-invalid value | parse error before validation |

## Risks and tradeoffs

### Risk: behavior change for required Cobra flags

Today, a required field without an explicit Cobra flag fails inside Cobra source parsing. After the fix, it fails after the source chain completes. The user-facing error still occurs, but later. This is desirable for env/config-aware commands, but tests that assert the exact old error location/message may need updates.

Mitigation: keep the error message field-oriented and add tests for missing required values.

### Risk: positional arguments may have different semantics

Required positional arguments are different from required flags because positional arguments are naturally CLI-only in many commands. However, Glazed fields can live in the same default section. Avoid over-generalizing until tests clarify expected behavior.

Mitigation: for the first implementation, fix required flag/source validation in `CobraParser.Parse`; preserve lower-level argument parser behavior unless a test demonstrates it blocks env/config-backed non-argument fields.

### Risk: empty-value semantics can surprise callers

A field may have a valid zero value. Boolean `false` and integer `0` should be accepted as present values, but string `""` and empty lists should usually fail when required.

Mitigation: make emptiness rules explicit in tests and docs.

### Risk: custom middleware chains bypass validation

Callers that use `cmd_sources.Execute` directly may still need to call validation explicitly. Callers that use `CobraParser.Parse` will be covered.

Mitigation: provide `sources.ValidateRequiredValues` and document it in middleware docs.

## Alternatives considered

### Alternative 1: keep `WithRequired(true)` as "required Cobra flag"

This would preserve current behavior but contradict the layered source model documented in the issue and in Glazed config docs. It would force applications to remove `WithRequired(true)` and duplicate validation after decode.

Rejected because it makes env/config-backed required values impossible.

### Alternative 2: add a new field option such as `WithRequiredValue(true)`

This could avoid changing existing semantics. However, most users naturally read `WithRequired(true)` as final-value required, especially in a framework that supports config and env sources. Adding a second option creates confusion and leaves the current footgun in place.

Possible future enhancement: add a separate `RequiredFlag`/`CLIRequired` concept only if there is a real use case for "must be typed as a flag even when env/config has a value."

### Alternative 3: make env/config run before Cobra source

This does not solve the core issue. `FromCobra` would still error when the flag is not changed, regardless of whether env/config already wrote a value into `parsedValues`, because the current required check only inspects `cmd.Flags().Changed(flagName)`.

Rejected because source ordering alone cannot fix source-local required validation.

### Alternative 4: catch `Field X is required` errors and retry without required validation

This would be brittle and error-message-dependent. It would also mask genuine parse errors.

Rejected because an explicit final validator is cleaner and easier to test.

## File-level implementation checklist

- `pkg/cli/cobra_parser_config_test.go`
  - Add env-backed required regression test.
  - Add config-backed required regression test.
  - Add missing required still fails test.
  - Add missing required with `--print-parsed-fields` skips validation test.
  - Add help flag skips validation / Cobra help does not run parser test.
  - Add empty-string default still fails test.

- `pkg/cmds/schema/section-impl.go`
  - Change Cobra source parsing to ignore required while gathering changed flags.
  - Update the existing TODO comment to point to final validation.

- `pkg/cmds/sources/validate_required.go`
  - Add final required validator helper and optional middleware.
  - Implement field-type-aware emptiness rules.
  - Return stable, readable errors.

- `pkg/cli/cobra-parser.go`
  - Call final validation after `cmd_sources.Execute` in `CobraParser.Parse`.

- `pkg/doc/topics/24-config-files.md`
  - Document that env/config can satisfy required fields.

- `pkg/doc/topics/16-parsing-fields.md`
  - Clarify final-value required semantics.

- `pkg/doc/topics/21-cmds-middlewares.md`
  - Mention final validation after source chain when discussing middleware ordering.

## References

### Repository files

- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra-parser.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cli/cobra_parser_config_test.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/fields/cobra.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/schema/section-impl.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/cobra.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/update.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/load-fields-from-config.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/sources/middlewares.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/values/section-values.go`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/pkg/cmds/fields/field-value.go`

### Ticket files

- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/sources/01-github-issue-556.md`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/scripts/01-collect-required-env-evidence.sh`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/scripts/02-reproduce-required-env-parser.sh`
- `/home/manuel/workspaces/2026-05-12/fix-required-fields-env/glazed/ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/scripts/evidence/02-reproduce-required-env-parser.log`

### External source

- GitHub issue 556: https://github.com/go-go-golems/glazed/issues/556
