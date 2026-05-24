---
Title: Glazed CLI linting rules analysis and implementation guide
Ticket: GLZ-CLI-LINT
Status: active
Topics:
    - glazed
    - linting
    - cli
    - cobra
    - intern-onboarding
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: geppetto/cmd/tools/geppetto-lint/main.go
      Note: Multichecker packaging precedent for bundled custom vettools
    - Path: geppetto/cmd/tools/turnsdatalint/main.go
      Note: Singlechecker packaging precedent for focused analyzer debugging
    - Path: geppetto/pkg/analysis/turnsdatalint/analyzer.go
      Note: Primary go/analysis analyzer precedent for traversal
    - Path: geppetto/pkg/doc/topics/12-turnsdatalint.md
      Note: Contributor documentation precedent for custom project linters
    - Path: glazed/pkg/cli/cobra-parser.go
      Note: Defines parser middleware chain and env/config/default value sources
    - Path: glazed/pkg/cli/cobra.go
      Note: Defines Cobra build/execution path and Glaze mode dispatch
    - Path: glazed/pkg/cmds/cmds.go
      Note: Defines Glazed command descriptions
    - Path: glazed/pkg/settings/glazed_section.go
      Note: Defines the Glazed output section that the proposed analyzer must recognize
    - Path: pinocchio/Makefile
      Note: Downstream vettool build and version-pinning pattern
    - Path: pinocchio/cmd/pinocchio/cmds/clip.go
      Note: Representative os.Getenv usage to be flagged or migrated
    - Path: pinocchio/cmd/pinocchio/cmds/js.go
    - Path: pinocchio/cmd/pinocchio/cmds/serve.go
      Note: Representative raw Cobra flag definition to be flagged by the proposed linter
    - Path: pinocchio/cmd/pinocchio/main.go
ExternalSources:
    - https://pkg.go.dev/golang.org/x/tools/go/analysis
    - https://pkg.go.dev/golang.org/x/tools/go/analysis/analysistest
Summary: Design for a Glazed-specific go/analysis vettool that flags raw environment reads, misplaced Glazed output sections, and raw Cobra/go flag definitions in CLI verbs.
LastUpdated: 2026-05-24T12:35:00-04:00
WhatFor: Use this to implement a custom Glazed CLI policy linter from scratch.
WhenToUse: When adding pkg/analysis/glazedclilint, wiring a glazed-lint vettool, or explaining the rule semantics to a new contributor.
---


# Glazed CLI linting rules analysis and implementation guide

## Executive summary

This document designs a new Glazed-specific static-analysis tool, tentatively named `glazedclilint`, that encodes three CLI policy rules that normal Go type checking cannot enforce:

1. **No direct `os.Getenv` in CLI command code.** Command configuration should flow through Glazed sections, parser middlewares, explicit config plans, or narrowly documented framework adapters, not ad-hoc process environment reads.
2. **Do not attach the Glazed output section to commands that do not emit structured rows through the Glazed framework.** `--output`, `--fields`, `--jq`, sorting, templating, and related flags are meaningful only when a command implements `cmds.GlazeCommand` or a deliberate dual-mode wrapper that still uses `settings.SetupTableProcessor` and a `middlewares.Processor`.
3. **Do not define raw Cobra, pflag, or standard-library `flag` flags in CLI verbs.** User-visible command flags should be declared as `fields.Definition` values inside `cmds.NewCommandDescription` so they participate in Glazed parsing, defaults, env/config layering, command schema printing, alias creation, help rendering, and generated documentation.

The implementation should follow the proven Geppetto pattern: put reusable analyzers under `pkg/analysis/<name>`, provide a single-analyzer `singlechecker` command for debugging, provide a bundled `multichecker` command for normal `go vet -vettool=...` use, test with `analysistest`, document the analyzer, and wire it into `make lint` and downstream repos. Geppetto already does this for `turnsdatalint`: the reusable analyzer is in `geppetto/pkg/analysis/turnsdatalint/analyzer.go`, the single checker is `geppetto/cmd/tools/turnsdatalint/main.go`, and the bundle is `geppetto/cmd/tools/geppetto-lint/main.go`.

The most important design choice is to make the linter **type-aware and policy-aware**, not just a grep wrapper. The rule should identify calls by resolved package/object identity (`os.Getenv`, `github.com/spf13/cobra.Command.Flags`, `github.com/spf13/pflag`, `flag`) and should understand Glazed command shapes (`cmds.CommandDescription`, `cmds.WithSections`, `settings.NewGlazedSection`, `cmds.GlazeCommand`). It should also have small, explicit allowlists for Glazed framework internals, tests, and legacy migrations so the linter can be enabled incrementally without becoming noisy.

## Problem statement and scope

Glazed CLIs use Cobra as their command tree substrate, but Glazed adds a schema layer on top of Cobra. A proper Glazed command describes user input in a `cmds.CommandDescription`; the command builder converts schema sections into Cobra flags and parses runtime values through middlewares. The framework then chooses classic mode (`BareCommand` / `WriterCommand`) or structured Glaze mode (`GlazeCommand`) when the command runs.

That layered model is visible in the code:

- `glazed/pkg/cmds/cmds.go` defines `CommandDescription`, schema sections, `WithFlags`, `WithArguments`, and command interfaces. `WithSections` stores sections in the schema at lines 77-82, `WithFlags` creates or reuses the default schema section at lines 127-147, `WithArguments` does the same for positional arguments at lines 150-174, and `GlazeCommand` is the interface with `RunIntoGlazeProcessor` at the bottom of the command interface block.
- `glazed/pkg/cli/cobra.go` builds Cobra commands from Glazed commands. It parses sections first at lines 49-57, detects whether the command should run in Glaze mode at lines 140-154, then creates a table processor and calls `RunIntoGlazeProcessor` at lines 155-184. It also automatically adds a Glazed section for commands that implement `cmds.GlazeCommand` if the section is missing at lines 224-240.
- `glazed/pkg/cli/cobra-parser.go` defines the middleware chain that resolves values from Cobra flags, positional args, env, config, and defaults. The default chain reads Cobra and args at lines 44-56. The env/config-aware built-in path can add `cmd_sources.FromEnv` and config plan loading at lines 143-186.
- `glazed/pkg/settings/glazed_section.go` defines the `settings.GlazedSlug` section and exposes all Glazed output/formatting flags; `NewGlazedSchema` wraps `NewGlazedSection` at lines 39-43, and `AddSectionToCobraCommand` adds child output/filter sections to Cobra at lines 113-125.

Three recurring mistakes bypass this model.

### Mistake 1: raw `os.Getenv`

Direct environment reads hide configuration from Glazed. They do not show up in `--print-schema`, they bypass value source tracking, they do not participate in env/config/default precedence, and they are hard to audit for secrets or reproducibility. A current Pinocchio example is `pinocchio/cmd/pinocchio/cmds/clip.go:135`, where `previewInPager` reads `$PAGER` directly. Glazed also has framework/internal examples, such as `glazed/pkg/cmds/sources/vault.go` reading `VAULT_TOKEN`, that should be considered separately because they implement a configuration source rather than a user-facing CLI verb.

The linter should make these differences explicit:

- **Default finding:** `os.Getenv` in command packages and general application code.
- **Allowed by default:** `_test.go` files and Glazed framework packages whose job is to implement env sources, if listed in an allowlist.
- **Migration recommendation:** replace direct env reads with Glazed parser env sources (`cmd_sources.FromEnv`) or fields/defaults that are surfaced in the command schema. For home directory lookup, use `os.UserHomeDir` rather than `os.Getenv("HOME")` when the intent is not command configuration.

### Mistake 2: Glazed output flags on non-structured commands

The Glazed output section contains user-facing flags such as output format, fields filtering, templating, jq, sorting, and skip/limit. Those flags imply that output is a stream of structured rows handled by the Glazed processor pipeline. If a command attaches that section but only writes text, starts a server, mutates state, or reads/writes files without calling `RunIntoGlazeProcessor`, users get flags that appear to work but are ignored.

The intended structured path is explicit in `glazed/pkg/cli/cobra.go`: when Glaze mode is active, the builder loads `settings.GlazedSlug`, constructs a table processor, wires output, and calls `RunIntoGlazeProcessor` (lines 155-184). If a command only implements `WriterCommand` or `BareCommand`, `BuildCobraCommandFromCommand` routes it to classic mode (lines 352-382), where Glazed output flags are meaningless.

Good examples in Pinocchio are `profiles list`, `profiles show`, and token listing commands:

- `pinocchio/cmd/pinocchio/cmds/profiles/list.go` creates a Glazed section, declares `var _ cmds.GlazeCommand = (*ListCommand)(nil)`, and emits profile rows via `gp.AddRow`.
- `pinocchio/cmd/pinocchio/cmds/profiles/show.go` does the same for a single profile row.
- `pinocchio/cmd/pinocchio/cmds/tokens/list.go` lists models/codecs as rows.

A risky pattern is a command that creates `settings.NewGlazedSection` and adds it through `cmds.WithSections(...)` but does not implement `RunIntoGlazeProcessor`. Another risky pattern is a command that is technically a `GlazeCommand` but only sometimes uses `gp` based on a custom boolean like `--glazed`. `pinocchio/cmd/pinocchio/cmds/catter/cmds/print.go` uses that conditional shape: it always adds the Glazed section at lines 39-42 and 131-134, but only attaches the processor when `s.Glazed` is true at lines 191-193. That may be deliberate dual-mode behavior, but it should be clearly annotated or eventually migrated to the framework-level dual-mode support (`cli.WithDualMode`).

### Mistake 3: raw Cobra, pflag, or `flag` definitions in CLI verbs

Raw Cobra flags fragment the command schema. A simple example is `pinocchio/cmd/pinocchio/cmds/serve.go`: it creates `*cobra.Command` manually and calls `cmd.Flags().StringVar(&address, "address", ":8088", ...)` at line 46. That is idiomatic Cobra, but it bypasses Glazed fields, value sources, command settings, schema printing, and consistent help behavior.

This is not always wrong inside Glazed itself. `glazed/pkg/cli/cobra.go` adds framework-controlled dual-mode toggle flags at lines 243-249, and `glazed/pkg/help/cmd/cobra.go` defines flags for the framework help command. Those are framework integration points. The rule should focus on **CLI verbs** rather than framework internals.

## Current-state architecture

### How Go analyzers are packaged today

Geppetto provides the reference implementation pattern:

```text
pkg/analysis/turnsdatalint/analyzer.go  reusable Analyzer value
cmd/tools/turnsdatalint/main.go         singlechecker wrapper
cmd/tools/geppetto-lint/main.go         multichecker bundle
Makefile linttool target                go vet -vettool=/tmp/geppetto-lint ./...
Pinocchio Makefile                      build Geppetto vettool and run it downstream
```

Concrete evidence:

- `geppetto/cmd/tools/turnsdatalint/main.go` imports the analyzer package and calls `singlechecker.Main(turnsdatalint.Analyzer)`.
- `geppetto/cmd/tools/geppetto-lint/main.go` imports `multichecker` and calls `multichecker.Main(turnsdatalint.Analyzer)` at lines 14-17. Its comment at lines 8-13 explicitly says this is the preferred long-term packaging shape: add analyzers under `pkg/analysis/<name>`, then register them in the bundle.
- `geppetto/pkg/analysis/turnsdatalint/analyzer.go` declares `var Analyzer = &analysis.Analyzer{...}` at lines 45-50, requires `inspect.Analyzer`, and uses `pass.Reportf` for diagnostics.
- `geppetto/pkg/analysis/turnsdatalint/analyzer_test.go` runs `analysistest.Run(t, analysistest.TestData(), Analyzer, "a")` at lines 9-10.
- `geppetto/pkg/analysis/turnsdatalint/testdata/src/a/a.go` demonstrates `// want` comments for expected diagnostics, such as raw run metadata keys at line 19 and raw payload literals at line 38.
- `geppetto/pkg/doc/topics/12-turnsdatalint.md` explains the model for contributors: custom analyzers are project rules implemented as `go vet` plugins, packaged under `pkg/analysis/<name>`, and run through `make lint`.

Pinocchio demonstrates downstream use:

- `pinocchio/Makefile` defines `GEPPETTO_LINT_BIN`, `GEPPETTO_LINT_PKG`, and `GEPPETTO_VERSION` at lines 19-21.
- `geppetto-lint-build` installs the vettool using the module version, with a workspace fallback for `(devel)`, at lines 23-34.
- `lint` and `lintmax` run `go vet -vettool=$(GEPPETTO_LINT_BIN) ./...` after `golangci-lint` at lines 46-52.

The Glazed linter should copy this packaging shape.

### How Glazed commands are represented

A Glazed command is a Go type with an embedded or otherwise exposed `*cmds.CommandDescription`. It implements one of the command interfaces in `glazed/pkg/cmds/cmds.go`:

- `BareCommand`: command performs side effects or manages its own output.
- `WriterCommand`: command writes classic text/bytes to an `io.Writer`.
- `GlazeCommand`: command emits structured rows through `RunIntoGlazeProcessor(ctx, parsedValues, gp)`.

Flags are not supposed to be added to the Cobra command by hand in normal verbs. Instead:

```go
cmds.NewCommandDescription(
    "profiles",
    cmds.WithFlags(
        fields.New("verbosity", fields.TypeString, fields.WithDefault("default")),
    ),
    cmds.WithArguments(
        fields.New("profile-ref", fields.TypeString),
    ),
    cmds.WithSections(profileSettingsSection),
)
```

The builder then converts sections to Cobra flags through `CobraParser.AddToCobraCommand` in `glazed/pkg/cli/cobra-parser.go` at lines 221-230 and beyond. At runtime it parses values through the middleware chain rather than through `cmd.Flags().GetString(...)` in the command body.

### ASCII architecture diagram

```text
                           source code
                               |
                               v
                   +-------------------------+
                   | pkg/analysis/glazedclilint |
                   | go/analysis Analyzer    |
                   +-----------+-------------+
                               |
           +-------------------+--------------------+
           |                   |                    |
           v                   v                    v
   os.Getenv calls     Glazed section misuse   raw flag definitions
   (CallExpr)          (NewGlazedSection +     (Cobra/pflag/flag calls)
                       WithSections + type)
           |                   |                    |
           +-------------------+--------------------+
                               |
                               v
                      pass.Reportf diagnostics
                               |
                               v
                go vet -vettool=/tmp/glazed-lint ./...
                               |
                               v
                    make lint / CI / downstream repos
```

## Proposed solution

Create a new analyzer package in the Glazed repository:

```text
glazed/pkg/analysis/glazedclilint/analyzer.go
glazed/pkg/analysis/glazedclilint/analyzer_test.go
glazed/pkg/analysis/glazedclilint/testdata/src/a/a.go
glazed/pkg/analysis/glazedclilint/testdata/src/github.com/go-go-golems/glazed/...
glazed/cmd/tools/glazedclilint/main.go      # singlechecker
glazed/cmd/tools/glazed-lint/main.go        # multichecker bundle
glazed/pkg/doc/topics/NN-glazedclilint.md   # contributor docs
```

The analyzer should export:

```go
var Analyzer = &analysis.Analyzer{
    Name: "glazedclilint",
    Doc:  "enforce Glazed CLI command policy: no raw env reads, no raw CLI flags, no Glazed output sections on non-Glaze commands",
    Requires: []*analysis.Analyzer{inspect.Analyzer},
    Run: run,
}
```

The bundled vettool should initially register only this analyzer:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/analysis/glazedclilint"
    "golang.org/x/tools/go/analysis/multichecker"
)

func main() {
    multichecker.Main(glazedclilint.Analyzer)
}
```

The single checker is equally small:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/analysis/glazedclilint"
    "golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
    singlechecker.Main(glazedclilint.Analyzer)
}
```

### Rule A: `os.Getenv`

**Diagnostic ID:** `glazedclilint/raw-env`

**Message:** `use Glazed config/env middleware or an explicit command field instead of os.Getenv in CLI code`

**Detection algorithm:**

- Visit `*ast.CallExpr` nodes.
- Resolve the called function through `pass.TypesInfo.Uses` on the selector.
- Match only a function named `Getenv` from package path `os`.
- Skip generated files, `_test.go`, and configured package/path allowlist entries.
- Optionally limit v1 to command roots (`/cmd/`, `/cmds/`) to avoid flagging framework env-source implementations until the repo is ready.

Pseudocode:

```text
for each CallExpr call:
  sel = selector(call.Fun)
  if sel == nil: continue
  fn = pass.TypesInfo.Uses[sel.Sel]
  if fn is not *types.Func: continue
  if fn.Pkg.Path != "os" or fn.Name != "Getenv": continue
  if file is generated or test or allowed path: continue
  if !isCommandCode(file) and command_roots_only: continue
  report(call.Pos, "use Glazed config/env middleware ...")
```

Implementation notes:

- Do not match by import alias text (`os.Getenv`) only. A file can import `os` as another name. Type identity is more reliable.
- For `os.LookupEnv`, decide explicitly. The user asked for `os.Getenv`; start with `Getenv`, but add a flag `-glazedclilint.env-functions=Getenv,LookupEnv` if the team wants both.
- Provide a suppression comment only if absolutely needed. The Geppetto suppression work showed that inline exceptions can become a policy escape hatch. Prefer path allowlists first.

Suggested alternatives in diagnostics or docs:

- Use `cmd_sources.FromEnv("APP", fields.WithSource("env"))` in a parser middleware, as Pinocchio's JS command does at `pinocchio/cmd/pinocchio/cmds/js.go:114-123`.
- Use `CobraParserConfig{AppName: "app"}` to enable built-in env loading when no custom middleware is supplied; this is supported in `glazed/pkg/cli/cobra-parser.go:91-104` and implemented at lines 162-168.
- Use `os.UserHomeDir()` for home directories where the value is not part of command configuration.

### Rule B: Glazed output section on non-structured command

**Diagnostic ID:** `glazedclilint/glazed-section-without-glaze-command`

**Message:** `this command exposes Glazed output flags but does not implement cmds.GlazeCommand; remove settings.NewGlazedSection/NewGlazedSchema or implement RunIntoGlazeProcessor`

**Detection algorithm, v1:**

- Track local variables assigned from `settings.NewGlazedSection(...)` or `settings.NewGlazedSchema(...)`.
- Track direct calls to those constructors inside `cmds.WithSections(...)`.
- Detect `cmds.WithSections(...)` arguments that include a tracked Glazed section variable.
- Find the nearest enclosing function that constructs a command description, usually by returning `&T{CommandDescription: cmds.NewCommandDescription(...)}` or assigning that composite to a local.
- Resolve `T` and inspect its method set for `RunIntoGlazeProcessor`.
- If `T` does not implement the Glaze method, report the `WithSections` argument or constructor call.

Pseudocode:

```text
first pass inside each function:
  glazedVars = set[*types.Var]
  for AssignStmt lhs, rhs:
    if rhs is call to settings.NewGlazedSection/NewGlazedSchema:
      mark lhs var as glazed section

second pass inside same function:
  for CallExpr call:
    if call is cmds.WithSections(...):
      if any arg is direct glazed constructor or ident in glazedVars:
        cmdType = inferConstructedCommandType(enclosing function)
        if cmdType unknown: report weak warning or skip depending strictness
        else if !hasRunIntoGlazeProcessor(cmdType): report
```

The command type inference should support common patterns:

```go
return &FooCommand{CommandDescription: cmds.NewCommandDescription(...)}
```

```go
cmd := &FooCommand{CommandDescription: cmds.NewCommandDescription(...)}
return cmd, nil
```

```go
return &FooCommand{cmds.NewCommandDescription(...)}
```

The method check can be implemented with `types.NewMethodSet(types.NewPointer(namedType))` and a signature validation. For v1, method name plus arity is acceptable; for v2, validate the full signature:

```go
RunIntoGlazeProcessor(context.Context, *values.Values, middlewares.Processor) error
```

**Important nuance:** `glazed/pkg/cli/cobra.go` already adds a Glazed section automatically for `cmds.GlazeCommand` implementations at lines 224-240. That means a new proper Glaze command does not need to manually call `settings.NewGlazedSection` unless it wants non-default output options. The linter can optionally emit an informational note in the future: "explicit Glazed section is redundant for a plain GlazeCommand". Do not make that a hard error in v1 because many existing commands create the section explicitly.

**Dual-mode commands:**

If a command can produce either text or rows, prefer framework dual mode (`cli.WithDualMode`) over a custom `--glazed` field. If the command uses a custom boolean but still implements `RunIntoGlazeProcessor`, v1 should not fail it. A future rule may warn when a Glaze command has a custom field named `glazed` because framework dual mode already has `with-glaze-output` / `no-glaze-output` controls in `glazed/pkg/cli/cobra.go:243-249`.

### Rule C: raw Cobra, pflag, or Go `flag` definitions in CLI verbs

**Diagnostic ID:** `glazedclilint/raw-flags`

**Message:** `define CLI flags with cmds.WithFlags(fields.New(...)) instead of raw Cobra/pflag/flag APIs`

**Detection algorithm:**

- Visit `*ast.CallExpr` nodes.
- Match standard-library package `flag` functions such as `String`, `StringVar`, `Bool`, `BoolVar`, `Int`, `IntVar`, `Duration`, `Var`, `NewFlagSet`, and `Parse`.
- Match package `github.com/spf13/pflag` functions and `NewFlagSet`.
- Match method calls on Cobra/pflag flag sets:
  - `cmd.Flags().String(...)`
  - `cmd.Flags().StringVar(...)`
  - `cmd.PersistentFlags().Bool(...)`
  - `cmd.InheritedFlags().Var(...)` if ever used for definitions
- Skip framework packages that build the bridge (`glazed/pkg/cli`, `glazed/pkg/cmds/fields`, `glazed/pkg/help/cmd`, logging early parsing), generated files, tests, and allowlisted paths.
- Focus default enforcement on command roots (`cmd/`, app `cmds/`) so the first rollout catches user-facing CLI verbs.

Pseudocode:

```text
for each CallExpr call:
  if isPackageFunction(call, "flag", definitionNames): report unless allowed
  if isPackageFunction(call, "github.com/spf13/pflag", definitionNames): report unless allowed
  if isRawCobraFlagDefinition(call):
      if isAllowedFrameworkPath(file): continue
      report(call.Pos, "define CLI flags with cmds.WithFlags...")

function isRawCobraFlagDefinition(call):
  method = selected method name of call.Fun
  if method not in flagDefinitionNames: return false
  recv = receiver of selector
  if recv is CallExpr whose selector is Flags/PersistentFlags/LocalFlags:
      recvType = type of recv
      return recvType package is github.com/spf13/pflag and named type FlagSet
  return false
```

This rule should flag `pinocchio/cmd/pinocchio/cmds/serve.go:46`, where `cmd.Flags().StringVar` defines an address flag directly. The migration shape is:

```go
type ServeSettings struct {
    Address string `glazed:"address"`
}

type ServeCommand struct { *cmds.CommandDescription }
var _ cmds.BareCommand = (*ServeCommand)(nil)

func NewServeCommand(hs *help.HelpSystem) (*ServeCommand, error) {
    return &ServeCommand{CommandDescription: cmds.NewCommandDescription(
        "serve",
        cmds.WithShort("Serve pinocchio help documentation as a web application"),
        cmds.WithFlags(fields.New(
            "address",
            fields.TypeString,
            fields.WithDefault(":8088"),
            fields.WithHelp("Address to listen on"),
        )),
    )}, nil
}

func (c *ServeCommand) Run(ctx context.Context, parsed *values.Values) error {
    s := &ServeSettings{}
    if err := parsed.DecodeSectionInto(schema.DefaultSlug, s); err != nil { return err }
    return runServe(ctx, s.Address, c.helpSystem)
}
```

The exact struct needs a field for `*help.HelpSystem`, but the principle is: the CLI flag is a Glazed field, and the command implements `BareCommand` or `WriterCommand` instead of returning a hand-built `*cobra.Command`.

## API reference for the intern

### `golang.org/x/tools/go/analysis`

Key types:

- `analysis.Analyzer`: declares name, documentation, required analyzers, flags, and the `Run` function.
- `analysis.Pass`: gives access to files, file set, package type information, imported packages, and diagnostic reporting.
- `pass.TypesInfo`: maps expressions/idents/selectors to types and objects.
- `pass.Reportf(pos, format, args...)`: emits a vet diagnostic.

Use `inspect.Analyzer` for efficient AST traversal:

```go
insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
nodeFilter := []ast.Node{(*ast.CallExpr)(nil), (*ast.ReturnStmt)(nil)}
insp.Preorder(nodeFilter, func(n ast.Node) { ... })
```

### Glazed APIs the linter must understand

- `cmds.NewCommandDescription`: creates the schema-carrying command description.
- `cmds.WithFlags`: declares command-local flags as `fields.Definition` values.
- `cmds.WithArguments`: declares positional arguments.
- `cmds.WithSections`: adds additional sections such as profile settings or Glazed output settings.
- `settings.NewGlazedSection` / `settings.NewGlazedSchema`: create the Glazed output section.
- `cmds.GlazeCommand`: marker interface for commands that emit rows through a `middlewares.Processor`.
- `cli.BuildCobraCommandFromCommand`: converts a Glazed command into a Cobra command and chooses classic vs Glaze execution.

### Cobra and flag APIs to flag

- `(*cobra.Command).Flags().String`, `StringVar`, `Bool`, `BoolVar`, `Int`, `IntVar`, `StringSlice`, `StringArray`, `Duration`, `Var`, and analogous methods.
- `(*cobra.Command).PersistentFlags().<definition method>`.
- Standard library `flag.String`, `flag.StringVar`, `flag.Bool`, `flag.Parse`, `flag.NewFlagSet`, etc.
- `github.com/spf13/pflag` package-level constructors and flag definition functions.

## Implementation plan

### Phase 1: Scaffold and run a no-op analyzer

1. Create `glazed/pkg/analysis/glazedclilint/analyzer.go` with an analyzer that requires `inspect.Analyzer` and returns no diagnostics.
2. Create `glazed/cmd/tools/glazedclilint/main.go` using `singlechecker.Main`.
3. Create `glazed/cmd/tools/glazed-lint/main.go` using `multichecker.Main`.
4. Add Makefile variables and targets:

```make
GLAZED_LINT_BIN ?= /tmp/glazed-lint

linttool-build:
	go build -o $(GLAZED_LINT_BIN) ./cmd/tools/glazed-lint

linttool: linttool-build
	go vet -vettool=$(GLAZED_LINT_BIN) ./cmd/... ./pkg/...
```

5. Keep this in a separate commit so reviewers can verify packaging before rule behavior.

### Phase 2: Implement Rule A (`os.Getenv`)

1. Implement helpers:
   - `func calledFunction(pass *analysis.Pass, call *ast.CallExpr) *types.Func`
   - `func isFunction(fn *types.Func, pkgPath, name string) bool`
   - `func filePath(pass *analysis.Pass, pos token.Pos) string`
   - `func shouldSkipFile(pass *analysis.Pass, filename string) bool`
2. Add flags:
   - `-allow-tests` default `true`
   - `-command-roots-only` default `true`
   - `-allow-path` repeatable or comma-separated allowlist
3. Add testdata cases:
   - raw `os.Getenv("PAGER")` in command file: diagnostic
   - aliased import `o "os"; o.Getenv("X")`: diagnostic
   - `_test.go`: no diagnostic by default
   - allowed framework source path: no diagnostic

### Phase 3: Implement Rule C (raw flags)

Do raw flags before the Glazed-section rule because it is mostly local call detection and will quickly pay off.

1. Implement package-function matching for `flag` and `pflag`.
2. Implement method-chain matching for `cmd.Flags().StringVar(...)` and `cmd.PersistentFlags().Bool(...)`.
3. Add testdata with a small fake command constructor:

```go
func NewServeCommand() *cobra.Command {
    var address string
    cmd := &cobra.Command{Use: "serve"}
    cmd.Flags().StringVar(&address, "address", ":8088", "Address") // want `define CLI flags with cmds.WithFlags`
    return cmd
}
```

4. Add allowed testdata under a fake `github.com/go-go-golems/glazed/pkg/cli` path so framework bridge code can keep raw Cobra operations.

### Phase 4: Implement Rule B (Glazed section misuse)

This is the only rule that needs light intra-function dataflow.

1. Build a per-function visitor that records variables initialized from Glazed section constructors.
2. Detect `cmds.WithSections` calls that contain those variables.
3. Infer command type from the composite literal that contains `cmds.NewCommandDescription`.
4. Check method set for `RunIntoGlazeProcessor`.
5. Report if the method is absent.
6. Add testdata:
   - `GlazeCommand` with `RunIntoGlazeProcessor`: no diagnostic.
   - `WriterCommand` with Glazed section: diagnostic.
   - `BareCommand` without Glazed section: no diagnostic.
   - Direct `cmds.WithSections(settings.NewGlazedSection())`: diagnostic when non-Glaze.
   - Unknown command type: decide whether to report a weaker diagnostic or skip. Prefer skip in v1 to avoid false positives.

### Phase 5: Documentation and rollout

1. Add a Glazed help topic in `glazed/pkg/doc/topics/` mirroring the style of `geppetto/pkg/doc/topics/12-turnsdatalint.md`.
2. Add Makefile lint integration after the analyzer passes on Glazed itself.
3. Run:

```bash
cd glazed
go test ./pkg/analysis/glazedclilint -count=1
go build -o /tmp/glazed-lint ./cmd/tools/glazed-lint
go vet -vettool=/tmp/glazed-lint ./cmd/... ./pkg/...
make lint
```

4. For downstream repos, follow the Pinocchio `geppetto-lint` shape:

```make
GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)

glazed-lint-build:
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) ./...
```

The version/workspace fallback matters because Pinocchio already learned this lesson for Geppetto: installing `@latest` causes drift, while installing without a version in CI can fail with readonly module mode. Pinocchio's Makefile handles both cases at lines 23-34.

## Suggested diagnostic wording and fixes

| Rule | Bad pattern | Suggested fix |
|---|---|---|
| raw env | `os.Getenv("PAGER")` | Add a Glazed field such as `pager` and resolve env through parser middleware, or use a documented framework env source. |
| raw flags | `cmd.Flags().StringVar(&address, "address", ":8088", ...)` | Convert the command to `cmds.BareCommand` / `WriterCommand` and declare `fields.New("address", fields.TypeString, fields.WithDefault(":8088"))`. |
| Glazed section misuse | `cmds.WithSections(glazedSection)` on a `WriterCommand` | Remove the output section or implement `RunIntoGlazeProcessor` and emit rows through `gp.AddRow`. |

## Risks and mitigations

- **False positives in framework internals.** Glazed itself necessarily has packages that use raw Cobra or env reads to implement the framework. Mitigation: path allowlists and a default command-root focus.
- **Incomplete command type inference.** Go constructors can be written in many ways. Mitigation: support common patterns first, skip unknowns in v1, and add test cases for every real pattern encountered.
- **Dual-mode ambiguity.** Some commands intentionally support both text and structured output. Mitigation: treat `RunIntoGlazeProcessor` as sufficient in v1; document preferred `cli.WithDualMode` migration separately.
- **Downstream churn.** Pinocchio and other apps may have existing raw Cobra commands. Mitigation: introduce linter in warning/report mode first by running a focused target, then add to `make lint` after fixes or allowlists.
- **Suppressions hiding design problems.** Inline suppressions can undermine policy. Mitigation: do not add general `//nolint:glazedclilint` support in v1; rely on explicit path allowlists in analyzer flags.

## Alternatives considered

### Grep or ripgrep scripts

A shell script could find `os.Getenv` and `.Flags().StringVar`, but it would miss aliased imports and would not know whether a command implements `cmds.GlazeCommand`. It would also be hard to test across packages. Rejected for anything beyond an initial audit.

### golangci-lint custom plugin

A golangci-lint plugin could work, but the repository already uses `go vet -vettool` for custom project rules. The Geppetto pattern is simpler, versionable as a Go module command, and easy for downstream repos to reuse.

### Runtime validation in `BuildCobraCommandFromCommand`

The builder could detect some schema/runtime mismatches, but it would not catch raw `os.Getenv` or raw flag definitions in hand-built Cobra commands before they are registered. Static analysis catches the source pattern at review time.

## Open questions

1. Should `os.LookupEnv` be included from day one, or should v1 match only the user-requested `os.Getenv`?
2. Should the raw-env rule apply to all non-test packages, or only CLI command roots by default?
3. Should explicit `settings.NewGlazedSection` on a proper `cmds.GlazeCommand` be allowed silently, or should a later cleanup rule encourage relying on the builder's automatic Glazed section insertion?
4. What is the team-preferred annotation for intentional dual-mode commands? A metadata marker on `CommandDescription`, a Go comment near the constructor, or simply using `cli.WithDualMode`?
5. Should downstream repos run `glazed-lint` directly, or should Glazed rules eventually be bundled into each app's own multichecker together with app-specific analyzers?

## References

- `geppetto/pkg/analysis/turnsdatalint/analyzer.go`: type-aware analyzer implementation with `inspect.Analyzer`, `pass.TypesInfo`, flags, and `pass.Reportf`.
- `geppetto/cmd/tools/turnsdatalint/main.go`: minimal singlechecker wrapper.
- `geppetto/cmd/tools/geppetto-lint/main.go`: bundled multichecker wrapper and packaging comment.
- `geppetto/pkg/analysis/turnsdatalint/analyzer_test.go`: `analysistest` entry point.
- `geppetto/pkg/analysis/turnsdatalint/testdata/src/a/a.go`: examples of `// want` diagnostics.
- `geppetto/pkg/doc/topics/12-turnsdatalint.md`: contributor-facing documentation for custom vettools.
- `pinocchio/Makefile`: downstream vettool build/install pattern with version pinning and workspace fallback.
- `pinocchio/cmd/pinocchio/cmds/serve.go`: example of raw Cobra flag definition in a CLI verb.
- `pinocchio/cmd/pinocchio/cmds/clip.go`: example of direct `os.Getenv("PAGER")` in command code.
- `pinocchio/cmd/pinocchio/cmds/js.go`: example of a custom Glazed parser middleware using `cmd_sources.FromEnv` and config files instead of ad-hoc env reads.
- `glazed/pkg/cmds/cmds.go`: command descriptions, schema sections, and command interfaces.
- `glazed/pkg/cli/cobra.go`: command execution path, Glaze mode, automatic Glazed section insertion, and dual-mode flags.
- `glazed/pkg/cli/cobra-parser.go`: parser middleware chain and env/config-aware value resolution.
- `glazed/pkg/settings/glazed_section.go`: Glazed output section and its Cobra flag wiring.
