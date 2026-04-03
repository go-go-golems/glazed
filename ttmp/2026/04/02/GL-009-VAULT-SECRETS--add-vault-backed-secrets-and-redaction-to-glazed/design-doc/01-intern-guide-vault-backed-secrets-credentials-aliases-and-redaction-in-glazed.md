---
Title: 'Intern guide: vault-backed secrets and redaction in Glazed'
Ticket: GL-009-VAULT-SECRETS
Status: active
Topics:
    - glazed
    - security
    - config
    - vault
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: ../../../../../../../vault-envrc-generator/pkg/glazed/middleware.go
      Note: Reference vault overlay middleware shape to port into Glazed
    - Path: ../../../../../../../vault-envrc-generator/pkg/vaultlayer/layer.go
      Note: Reference vault settings section and decoding helper
    - Path: pkg/appconfig/options.go
      Note: Existing bootstrap parse pattern to reuse for vault-settings
    - Path: pkg/cli/helpers.go
      Note: printParsedFields currently emits raw field values and metadata
    - Path: pkg/cmds/fields/field-value_test.go
      Note: Shows metadata aliasing is already fixed in current tree
    - Path: pkg/cmds/fields/serialize.go
      Note: Current serialization path leaks raw secret values and parse logs
    - Path: pkg/cmds/sources/middlewares.go
      Note: Defines middleware execution and precedence semantics
ExternalSources: []
Summary: ""
LastUpdated: 2026-04-02T19:20:42.494806726-04:00
WhatFor: ""
WhenToUse: ""
---


# Intern guide: vault-backed secrets and redaction in Glazed

## Executive Summary

Glazed already has the beginnings of the feature set we want. It has `fields.TypeSecret`, it already treats secret values as strings during parsing, and it already has a working model for precedence-aware bootstrap parsing in `appconfig.WithProfile(...)`. The two big missing pieces are:

1. Output redaction is incomplete. Secret fields are masked by `RenderValue`, but raw secret values still leak through parsed-field serialization and `--print-parsed-fields`.
2. Vault-backed secret population exists as a migrated proof-of-concept in `vault-envrc-generator`, but Glazed core does not yet provide a first-class section + middleware pattern for it.

After reading the imported notes and the actual code, my recommendation is:

1. Keep one internal "sensitive string" semantic for the first implementation.
   Use `TypeSecret` as the only behavior-bearing type for now.
2. Keep `TypeSecret` as the only behavior-bearing sensitive-string semantic in the first pass.
3. Do not add a `credentials` alias in this pass.
4. Centralize redaction in `pkg/cmds/fields`, then make all human-readable/debug output flow through those helpers.
5. Port only the small Vault overlay pattern first.
   Do not start with the larger generalized `SecretResolver` registry / `SecretRef` / bootstrap-flag framework from the clean patch unless a second provider is already a real requirement.
6. Reuse the profile bootstrap pattern as the conceptual model for provider bootstrap parsing.

This guide is written for a new intern. It explains the current system, the actual gaps, the recommended first-pass design, the tradeoffs, and the exact file-level work plan.

### Implementation Update: 2026-04-02

The implementation that landed after this investigation stayed smaller than the early alias-oriented sketch:

1. `TypeSecret` remains the only sensitivity semantic.
2. No `credentials` alias was added.
3. Vault support landed as reusable source-layer APIs in `pkg/cmds/sources`:
   - `NewVaultSettingsSection()`
   - `GetVaultSettings(...)`
   - `FromVaultSettings(...)`
   - `BootstrapVaultSettings(...)`
4. The `appconfig.WithProfile(...)` flow remains the conceptual bootstrap model, but the first merge did not add a separate `appconfig.WithVault(...)` wrapper.

## Problem Statement And Scope

We want Glazed to support two related capabilities:

1. Sensitive field handling that is safe and predictable.
2. Vault-backed secret loading that composes with Glazed's existing precedence model.

The user-facing request mentions "vault support" and "secrets/credentials". Those sound like one problem because they share the same semantic center:

- some fields are sensitive,
- those fields should not leak in help/debug/serialization output,
- and some of those fields may be resolved from an external secret store.

The imported notes are useful inputs, but they are not automatically the correct design. The codebase already contains important facts that change the situation:

1. `TypeSecret` already exists in Glazed and is already treated as a string type in multiple parsers.
2. The old metadata aliasing bug has already been fixed in current Glazed.
3. A robust bootstrap-parse pattern already exists in `appconfig.WithProfile(...)`.
4. The migrated `vault-envrc-generator` implementation demonstrates the right "overlay after next" middleware shape, but it is still broader and looser than what Glazed core should ship first.

### Scope Of This Design

In scope for the first implementation:

1. Make `TypeSecret` the single first-pass sensitivity semantic.
2. Centralize redaction for parsed-field serialization, parsed-field printing, and other debug-facing representations.
3. Add a small Vault settings section and a Vault overlay middleware that hydrates only sensitive fields.
4. Add a bootstrap parsing recipe or helper for `vault-settings` so provider settings can come from config/env/flags while final app fields still allow env/flags to win.

Out of scope for the first pass:

1. A generalized secret-provider registry.
2. A `credentials` alias or any second sensitivity spelling.
3. A field-level `SecretRef` model.
4. Non-string sensitive field semantics.
5. A KMS abstraction covering multiple providers.

Those are reasonable future directions, but they are not needed to land a clean first implementation.

## Current-State Analysis

### 1. Glazed already has a secret field type

Observed behavior:

1. `TypeSecret` exists in the core type enum in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-type.go:7-10`.
2. Secret fields are parsed as single strings in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/parse.go:103-113`.
3. Secret fields are rendered in masked form by `RenderValue` in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/render.go:27-39`.
4. Cobra and Viper parsing already special-case `TypeSecret` alongside strings in:
   - `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go:183-197`
   - `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go:454-472`
   - `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/viper.go:57-63`
5. Definition validation, initialization, and reflect assignment already treat `TypeSecret` like a string in:
   - `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/definitions.go:151-204`
   - `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/definitions.go:773-849`

Implication:

Glazed does not need a brand-new secret field system. It already has one. The first pass should build on that instead of introducing a second parallel semantic branch.

### 2. The real leak is in serialization and parsed-field printing

Observed behavior:

1. `ToSerializableFieldValue` currently copies raw `FieldValue.Value` and raw `[]ParseStep` directly in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go:16-22`.
2. `printParsedFields` prints raw `fieldValue.Value`, raw `ParseStep.Value`, and raw `ParseStep.Metadata` in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go:68-99`.
3. Config/map-based parsing still records the raw source payload under `map-value` metadata in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/gather-fields.go:56-59`.
4. Config file loading adds file/index metadata and then calls `updateFromMap(...)` in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/load-fields-from-config.go:65-79`.

Implication:

`TypeSecret` currently protects only one presentation path: `RenderValue`.

That means the public documentation claim that secrets are "masked in output" is too broad. The actual code still leaks in at least these paths:

1. YAML/JSON serialization of parsed values.
2. `--print-parsed-fields`.
3. Cobra help/default display when a secret field has a non-empty default.

### 3. Metadata aliasing has already been fixed

Observed behavior:

1. `WithMetadata` now copies the incoming metadata map when first assigning it in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-value.go:21-45`.
2. The regression tests for this already exist in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-value_test.go:9-44`.

Implication:

An older design concern is no longer current. The remaining problem is not "shared metadata aliasing leaks secrets across unrelated fields". The remaining problem is simpler:

1. secret fields still retain raw values in memory and logs,
2. sanctioned debug/serialization outputs still expose them verbatim.

That correction matters because it means the first fix should target output sanitization first.

### 4. Cobra help can still leak secret defaults

Observed behavior:

1. When a secret field has a default, `AddFieldsToCobraCommand` passes the raw default string to Cobra in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go:183-197`.
2. Cobra stores a display default string separately from the actual flag value.

Implication:

Even if runtime parse logs are redacted, help text can still leak a secret default. This is not the main gap discussed in the imported redaction note, but it is a real sensitivity leak and should be fixed in the same hardening phase.

### 5. Precedence and bootstrap parsing are already modeled elsewhere in Glazed

Observed behavior:

1. `sources.Execute(...)` reverses the middleware list and then runs middlewares that usually call `next` first, as documented in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/middlewares.go:37-67`.
2. The default Cobra parser builds a low-level source chain from config/defaults/env/args/flags in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/cobra-parser.go:140-210`.
3. `appconfig.WithProfile(...)` already performs a bootstrap parse of a single settings section (`profile-settings`) before applying the main profile middleware in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/options.go:174-314`.
4. The profile tests already prove the intended precedence and bootstrap behavior in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/profile_test.go:14-225`.

Implication:

Glazed does not need a new execution model for vault bootstrap parsing. It already has a working reference implementation for the exact kind of circularity we care about:

- bootstrap one small settings section first,
- decode it,
- then insert a derived middleware into the main chain at the correct precedence point.

### 6. There is already a migrated Vault proof-of-concept in vault-envrc-generator

Observed behavior:

1. `vault-envrc-generator/pkg/glazed/middleware.go` defines `UpdateFromVault(...)` as an overlay middleware that calls `next` first, then reads parsed settings, then updates values in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/glazed/middleware.go:35-105`.
2. `vault-envrc-generator/pkg/vaultlayer/layer.go` defines a reusable `VaultSettings` struct and a `NewVaultSection()` helper in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/vaultlayer/layer.go:12-73`.
3. The example command shows a bootstrap-only vault section followed by a Vault overlay in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/cmd/examples/vault-glaze-example/main.go:47-68`.

Observed limitations in that implementation:

1. It updates every matching field by name, not just sensitive fields.
2. The `vault-token` field is currently declared as `fields.TypeString`, not `fields.TypeSecret`, in `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/vaultlayer/layer.go:33-38`.
3. It is provider-specific and tightly coupled to the envrc-generator repo.

Implication:

The middleware shape is good. The field selection policy is too loose for Glazed core. The settings section also needs a sensitivity pass.

## System Map For A New Intern

If you are new to Glazed, the fastest way to understand this feature is to treat it as a five-layer pipeline:

### Layer A: field definitions

This is where semantics live.

- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/definitions.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-type.go`

Responsibilities:

1. declare field names,
2. declare types,
3. declare defaults,
4. drive validation and decoding.

### Layer B: field parsing and parse logs

- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/parse.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-value.go`

Responsibilities:

1. turn raw strings/maps/env/config inputs into typed values,
2. attach parse provenance (`source`, `metadata`),
3. keep a log of each applied step.

### Layer C: source middlewares

- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/update.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/load-fields-from-config.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/cobra.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/middlewares.go`

Responsibilities:

1. apply defaults,
2. apply config files,
3. apply env vars,
4. apply flags/args,
5. merge values into `values.Values`.

### Layer D: bootstrap and composition

- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/cobra-parser.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/options.go`

Responsibilities:

1. decide which sources run,
2. decide in what effective order they apply,
3. support "parse a tiny settings section first, then build the real chain".

### Layer E: output and debugging

- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/render.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go`
- File: `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/values/serialize_parsed.go`

Responsibilities:

1. present values safely to humans,
2. serialize parsed values safely,
3. preserve observability without leaking secrets.

### Mental model diagram

```text
Field Definition
    |
    v
Parsing + ParseStep Log
    |
    v
Source Middlewares
    |
    v
values.Values
    |
    +--> Decode into structs
    |
    +--> Debug print / YAML / JSON / help
```

Vault support is just an additional source middleware plus an optional bootstrap phase for the Vault settings section.

## Gap Analysis

### Gap 1: there is no single reusable sensitivity helper

Current state:

1. `RenderValue` knows how to mask `TypeSecret`.
2. Serialization does not.
3. Parsed-field printing does not.
4. Cobra display defaults do not.

Consequence:

Every new output path can accidentally become a leak path.

### Gap 2: aliasing is not the real blocker

Current state:

1. Internal semantics already use `TypeSecret`.
2. The imported clean patch proposes a new `TypeCredentials` branch and alias handling.
3. None of that is required to land redaction or Vault hydration.

Consequence:

If we add a brand-new internal type or even a second accepted spelling too early, we must touch every switch and every user-facing explanation that already knows about `TypeSecret`:

- parsing,
- validation,
- reflection,
- cobra,
- viper,
- JSON schema,
- codegen,
- Lua integrations,
- tests,
- docs.

That is a large blast radius for very little new behavior.

### Gap 3: Glazed core has no built-in Vault middleware

Current state:

1. The proof-of-concept exists in `vault-envrc-generator`.
2. Glazed core does not expose an equivalent section and middleware.

Consequence:

Applications either re-implement the pattern or depend on a sidecar repo.

### Gap 4: field selection rules are not yet safe enough

Current state in the migrated envrc-generator middleware:

1. any field whose name matches a Vault key gets overwritten,
2. regardless of whether that field is meant to be sensitive.

Consequence:

This is too broad for a framework-level feature.

Framework code should follow an explicit invariant:

`Only fields declared as sensitive are eligible for secret hydration.`

### Gap 5: bootstrap parsing is needed, but should not become a new framework concept

Current state:

1. There is already a proven bootstrap pattern for profiles.
2. The imported clean patch grows a broader bootstrap framework (`Bootstrap`, `SecretRef`, resolver registry).

Consequence:

The general framework is likely premature for the first pass.

## Proposed Solution

## Recommendation Summary

Implement the feature in three layers:

1. `TypeSecret` remains the only first-pass sensitive behavior type.
2. Add centralized sensitivity helpers in `pkg/cmds/fields`.
3. Add a small Vault settings section plus a small overlay middleware that hydrates only sensitive fields.

Do not start with the generalized clean-patch architecture.

### Design Principle 1: keep one internal sensitive-string semantic

Recommendation:

1. Keep `TypeSecret` as the behavior-bearing type.
2. Do not add a second accepted spelling in the first pass.
3. Document that if we later need non-string sensitive values, sensitivity should become a property on `fields.Definition`, not a proliferation of pseudo-types.

Why:

1. The code already understands `TypeSecret`.
2. A new internal `TypeCredentials` or alias layer adds maintenance cost without adding first-pass behavior.
3. The user's requested semantics are "sensitive string", not "new structured type".

### Design Principle 2: centralize redaction in the fields package

Recommendation:

Add one small sensitivity API in `pkg/cmds/fields`, for example:

```go
func (t Type) IsSensitive() bool
func RedactValue(t Type, value interface{}) interface{}
func RedactMetadata(t Type, metadata map[string]interface{}) map[string]interface{}
func RedactParseStep(t Type, step ParseStep) ParseStep
```

Then wire all human-readable/debug paths through those helpers:

1. `ToSerializableFieldValue`
2. `ToSerializableFieldValues`
3. `printParsedFields`
4. Cobra default display values for sensitive flags

Why:

1. It keeps policy in one package.
2. It makes new output modes safer by default.
3. It avoids a long tail of one-off redaction logic.

### Design Principle 3: keep the Vault overlay middleware pure

Recommendation:

The middleware should:

1. call `next` first,
2. read already-parsed vault settings,
3. connect to Vault,
4. fetch one secret map,
5. overlay values only onto sensitive fields.

Pseudocode:

```go
func FromVaultSettings(vs *VaultSettings, options ...fields.ParseOption) sources.Middleware {
    return func(next sources.HandlerFunc) sources.HandlerFunc {
        return func(schema_ *schema.Schema, parsedValues *values.Values) error {
            if err := next(schema_, parsedValues); err != nil {
                return err
            }

            client, err := NewVaultClientFromSettings(vs)
            if err != nil {
                return err
            }

            secrets, err := client.ReadPath(vs.SecretPath)
            if err != nil {
                return err
            }

            return schema_.ForEachE(func(_ string, section schema.Section) error {
                sectionValues := parsedValues.GetOrCreate(section)
                return section.GetDefinitions().ForEachE(func(def *fields.Definition) error {
                    if def.Type != fields.TypeSecret {
                        return nil
                    }

                    raw, ok := secrets[def.Name]
                    if !ok {
                        return nil
                    }

                    opts := append([]fields.ParseOption{
                        fields.WithSource("vault"),
                        fields.WithMetadata(map[string]interface{}{
                            "provider": "vault",
                            "path":     vs.SecretPath,
                        }),
                    }, options...)

                    return sectionValues.Fields.UpdateValue(def.Name, def, raw, opts...)
                })
            })
        }
    }
}
```

Why:

1. This matches the existing middleware execution model.
2. It makes precedence explicit by placement in the chain.
3. It is much smaller than a general secret-framework rollout.

### Design Principle 4: bootstrap only when provider settings need dual precedence roles

Problem:

Sometimes Vault settings need to come from env/config/flags, while env/config/flags must still retain the right to override the final secret-backed application fields.

A single pass cannot satisfy both requirements cleanly.

Recommendation:

Reuse the `WithProfile(...)` pattern:

1. bootstrap-parse only `vault-settings`,
2. decode `VaultSettings`,
3. insert the Vault overlay middleware into the real chain between config and env/flags.

Diagram:

```text
Bootstrap chain for vault-settings only:
defaults -> config -> env -> cobra

Main chain for all sections:
defaults -> config -> vault -> env -> args -> cobra
```

This gives:

1. config/env/flags can configure the Vault client,
2. env/flags still override the final app field values.

### Design Principle 5: do not hydrate arbitrary matching fields

Recommendation:

For the first pass, only hydrate fields where:

1. `def.Type == fields.TypeSecret`

Do not hydrate:

1. ordinary strings,
2. integers,
3. fields that just happen to share a name with a Vault key.

Future extension:

If 1:1 field-name mapping becomes insufficient, add a field-level secret reference later, for example:

```go
type SecretRef struct {
    Provider string
    Path     string
    Key      string
}
```

But do not add that now unless a real caller needs it.

## Proposed APIs And File Layout

### A. Sensitivity helpers

Recommended files:

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-type.go`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/sensitive.go`
3. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go`
4. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go`
5. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go`

Recommended API surface:

```go
const RedactedPlaceholder = "***"

func (t Type) IsSensitive() bool
func RedactValue(t Type, value interface{}) interface{}
func RedactMetadata(t Type, metadata map[string]interface{}) map[string]interface{}
func RedactParseStep(t Type, step ParseStep) ParseStep
```

### B. Sensitivity naming

Recommendation for the first merge:

1. Keep the public and internal spelling on `secret` / `TypeSecret`.
2. Do not add `credentials` normalization yet.
3. Revisit aliasing only if a concrete downstream consumer still needs it after the Vault flow settles.

### C. Vault settings section

Recommended files:

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_settings.go`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault.go`

Recommended section:

```go
type VaultSettings struct {
    VaultAddr        string `glazed:"vault-addr"`
    VaultToken       string `glazed:"vault-token"`
    VaultTokenSource string `glazed:"vault-token-source"`
    VaultTokenFile   string `glazed:"vault-token-file"`
    SecretPath       string `glazed:"secret-path"`
}

func NewVaultSettingsSection() (schema.Section, error)
func GetVaultSettings(parsed *values.Values) (*VaultSettings, error)
```

Important details:

1. `vault-token` should be `fields.TypeSecret`, not `fields.TypeString`.
2. `secret-path` can remain a plain string.
3. If the existing envrc-generator templated path feature is desired, preserve it behind a small helper rather than spreading template logic through the middleware.

### D. Optional bootstrap helper

This is optional for the first merge, but recommended if the code will be reused by more than one app.

Recommended shape:

```go
func BootstrapVaultSettings(
    cmd *cobra.Command,
    args []string,
    configFiles []string,
    envPrefix string,
) (*VaultSettings, error)
```

or a middleware-builder form:

```go
func BuildVaultAwareMiddlewares(
    cmd *cobra.Command,
    args []string,
    configFiles []string,
    envPrefix string,
) ([]sources.Middleware, error)
```

The simplest valid first pass is not a public helper at all. It can just be an example pattern plus tests.

## Precedence Guide

This is the part most likely to confuse a new intern.

### How source ordering works in practice

Observed behavior:

1. `sources.Execute(...)` reverses the middleware slice before building the handler chain.
2. Most middlewares call `next` first, then apply their values.
3. Therefore, the effective application order is usually low precedence first, high precedence last.

Rule of thumb:

If you want this precedence:

```text
defaults < config < vault < env < args < cobra
```

then the values must be applied in exactly that order.

### Recommended main-chain order

```text
defaults -> config files -> vault -> environment -> arguments -> cobra flags
```

Recommended semantics:

1. defaults establish baseline values,
2. config files provide file-backed application config,
3. vault hydrates secret fields from the chosen path,
4. environment variables can override both config and vault,
5. args and flags win last.

### Recommended bootstrap order for vault-settings

```text
defaults -> config files -> env -> cobra
```

Why:

1. provider connection settings often live in config,
2. CI or developers often override them with env vars,
3. operators may still want an explicit flag to win.

### ASCII flow diagram

```text
                    +-----------------------+
                    | Parse vault-settings  |
                    | defaults/config/env   |
                    | /cobra only           |
                    +-----------+-----------+
                                |
                                v
                       decode VaultSettings
                                |
                                v
defaults -> config -> FromVaultSettings(vs) -> env -> args -> cobra
                                |
                                v
                       update TypeSecret fields
```

## Detailed Implementation Plan

### Phase 1: Fix sensitivity and redaction in core field output

Goal:

Make all secret-bearing debug/serialization output safe before adding new Vault population behavior.

Files to change:

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-type.go`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go`
3. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go`
4. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go`
5. new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/sensitive.go`

Tasks:

1. Add `Type.IsSensitive()` with only `TypeSecret == true` for now.
2. Add redaction helpers for values, parse steps, and metadata.
3. Make `ToSerializableFieldValue(...)` return redacted values/logs.
4. Make `printParsedFields(...)` build its output from `ToSerializableFieldValue(...)` instead of duplicating formatting logic.
5. Redact Cobra `DefValue` display strings for sensitive defaults after registering the flag.
6. Add regression tests.

Recommended tests:

1. secret value is redacted in `ToSerializableFieldValue`,
2. secret parse-step metadata is redacted,
3. JSON marshalling of parsed field values does not contain the raw secret,
4. Cobra help/default display does not expose the raw secret default,
5. non-secret values remain unchanged.

Notes for the intern:

1. Do not build a redaction helper that assumes strings only if the API claims to be generic.
2. The first implementation can still return `***` for unknown/non-string sensitive values.
3. Keep the public behavior deterministic and easy to test.

### Phase 2: Keep the first pass on `TypeSecret` only

Goal:

Avoid introducing a second sensitive-string spelling while the core Vault and redaction behavior is still stabilizing.

Tasks:

1. Do not add `credentials` decoding aliases in the first merge.
2. Keep all behavior keyed on `TypeSecret`.
3. Revisit aliasing only if a concrete downstream caller still needs it after the Vault flow settles.

### Phase 3: Port the minimal Vault settings section into Glazed

Goal:

Make Vault support available as a reusable Glazed building block.

Files to change:

1. new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_settings.go`
2. maybe new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_settings_test.go`

Tasks:

1. Define `VaultSettings`.
2. Define `NewVaultSettingsSection()`.
3. Define `GetVaultSettings(parsed *values.Values)`.
4. Mark `vault-token` as `TypeSecret`.
5. Add tests for decoding the section into the struct.

Potential fields:

1. `vault-addr`
2. `vault-token`
3. `vault-token-source`
4. `vault-token-file`
5. `secret-path`

### Phase 4: Port the minimal Vault overlay middleware

Goal:

Hydrate only sensitive application fields from Vault in a way that composes with current middleware precedence.

Files to change:

1. new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault.go`
2. new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_test.go`

Tasks:

1. Port the "call next first, then overlay" pattern from envrc-generator.
2. Restrict hydration to `TypeSecret` fields only.
3. Keep the initial field-to-secret mapping rule simple: `definition.Name -> secrets[definition.Name]`.
4. Attach parse metadata such as:
   - `provider: vault`
   - `path: <effective path>`
5. Ensure update failures wrap errors with clear context.

Recommended test strategy:

Do not require a live Vault server in unit tests.

Instead:

1. define a tiny internal client interface,
2. inject a fake client in tests,
3. assert that:
   - only secret fields are updated,
   - plain string fields are left alone,
   - parse logs record `source=vault`,
   - precedence works when middleware is placed before or after env/cobra.

### Phase 5: Add a bootstrap recipe or helper

Goal:

Solve the precedence cycle cleanly when Vault connection settings themselves come from config/env/flags.

Files to change:

1. maybe new `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/vault_bootstrap.go`
2. or only examples/docs/tests if you want to keep the first merge smaller

Tasks:

1. Bootstrap-parse only `vault-settings`.
2. Decode `VaultSettings`.
3. Build the real source chain with Vault inserted between config and env/flags.

The intern should study:

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/options.go:174-314`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/profile_test.go:14-225`

Those files are the best in-repo model for this problem.

### Phase 6: Documentation and examples

Goal:

Make the behavior obvious to downstream app authors.

Files to change:

1. a new doc topic or tutorial in `pkg/doc/topics/` or `pkg/doc/tutorials/`
2. an example command under `cmd/examples/`

Documentation must explain:

1. what `TypeSecret` actually guarantees,
2. how to place Vault in the precedence chain,
3. when bootstrap parsing is required,
4. that only sensitive fields are eligible for secret hydration.

## Testing And Validation Strategy

### Unit tests

Recommended packages:

1. `go test ./pkg/cmds/fields`
2. `go test ./pkg/cmds/sources`
3. `go test ./pkg/appconfig`

Required coverage:

1. type alias normalization,
2. redaction helpers,
3. serializable parsed values do not leak raw secrets,
4. Cobra default display redaction,
5. Vault settings section decode,
6. Vault overlay only touches secret fields,
7. bootstrap parse precedence.

### Integration-style tests

Add one or both:

1. a small example command test that prints parsed fields and asserts the raw secret never appears,
2. a source-chain test that applies defaults/config/vault/env/cobra and asserts the final winning value is correct.

### Human review checklist

When reviewing the implementation, explicitly inspect:

1. every output path that touches `FieldValue.Value` or `ParseStep.Value`,
2. every switch over `fields.Type`,
3. the exact middleware order in examples/tests,
4. whether any new Vault settings field should itself be marked `TypeSecret`.

## Alternatives Considered

### Alternative A: add a full `TypeCredentials` internal type

This is the direction taken in the imported clean patch.

Why I do not recommend it for the first pass:

1. it duplicates the existing `TypeSecret` semantic,
2. it increases switch-branch count across the codebase,
3. it creates more documentation and maintenance work than behavior value,
4. the actual problem is not "missing second sensitive string type".

When it might become reasonable:

If `credentials` eventually means a structured object rather than a sensitive string. That is not the requirement today.

### Alternative B: start with the full generalized secret resolver framework

The clean patch also proposes:

1. `SecretRef`,
2. `SecretResolver`,
3. provider registry,
4. bootstrap-only fields,
5. fill-if-unset vs override modes.

Why I do not recommend it first:

1. it is a much larger conceptual surface,
2. Glazed currently has one concrete provider/use case,
3. the repo already has a smaller working Vault pattern to port,
4. it is harder for a new contributor to land and review safely.

When it might become reasonable:

When a second provider or per-field mapping requirement is real and active.

### Alternative C: redact only in `printParsedFields`

Rejected because:

1. YAML/JSON serialization would still leak,
2. future output paths would still be vulnerable,
3. the policy would be scattered.

### Alternative D: remove `map-value` metadata entirely

This is a valid hardening option and was previously attractive when metadata aliasing was broken.

Current assessment:

1. It is less urgent now because metadata aliasing has already been fixed.
2. Central output redaction is still required either way.
3. Removing `map-value` entirely is still a reasonable follow-up if maintainers decide the debug value is not worth the in-memory risk.

Recommendation:

Treat this as an optional hardening follow-up, not the first blocking step.

### Alternative E: keep Vault integration outside Glazed entirely

Why this is still tempting:

1. it keeps provider dependencies out of the core framework,
2. it keeps Glazed narrower.

Why I still recommend a small Glazed-side integration:

1. the request is for first-class Vault support in Glazed,
2. the middleware/section abstraction is already a natural fit for the framework,
3. the code can remain small if the design avoids premature generalization.

## Design Decisions

### Decision 1: `TypeSecret` remains canonical

Status: recommended

Rationale:

1. existing code already supports it,
2. it minimizes blast radius,
3. it aligns with the actual semantics we need today.

### Decision 2: keep `TypeSecret` as the only first-pass semantic

Status: recommended

Rationale:

1. the runtime behavior already exists on `TypeSecret`,
2. a second spelling does not unlock Vault support or redaction safety by itself,
3. avoiding the alias keeps the implementation surface smaller.

### Decision 3: redaction lives in `pkg/cmds/fields`

Status: recommended

Rationale:

1. one policy point,
2. fewer accidental leaks,
3. cleaner reuse by debug/serialization layers.

### Decision 4: only `TypeSecret` fields are eligible for Vault hydration

Status: recommended

Rationale:

1. explicit beats implicit,
2. field-name coincidence is too broad,
3. this aligns the hydration rule with the sensitivity rule.

### Decision 5: bootstrap parsing should copy the profile pattern, not invent a new framework

Status: recommended

Rationale:

1. proven in current code,
2. easier for new contributors to reason about,
3. keeps first-pass scope small.

## Open Questions

1. Does the Glazed repo want to take on the Vault dependency directly, or should the client implementation sit behind a tiny interface and provider-specific package?
2. Should `map-value` metadata remain in memory for sensitive fields once output redaction is fixed, or should Glazed also stop recording raw map values at source?
3. Should the first Glazed Vault section support templated secret paths immediately, or should path templating stay in a follow-up after the base middleware lands?
4. Should the bootstrap helper be public API in the first merge, or should the first merge ship only a tested recipe/example?
5. Long term, should sensitivity move from "type" to a `Definition` property such as `Sensitive bool`?

## References

### Primary Glazed code

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-type.go`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/definitions.go`
3. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/parse.go`
4. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/render.go`
5. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-value.go`
6. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/serialize.go`
7. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/gather-fields.go`
8. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/cobra.go`
9. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/viper.go`
10. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/load-fields-from-config.go`
11. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/update.go`
12. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/sources/middlewares.go`
13. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/helpers.go`
14. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cli/cobra-parser.go`
15. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/options.go`
16. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/appconfig/profile_test.go`

### Existing regression proof

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/glazed/pkg/cmds/fields/field-value_test.go`

### Vault reference implementation

1. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/glazed/middleware.go`
2. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/vaultlayer/layer.go`
3. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/cmd/examples/vault-glaze-example/main.go`
4. `/home/manuel/workspaces/2025-09-10/add-vault-middleware-to-glazed/vault-envrc-generator/pkg/doc/06-vault-glazed-middleware.md`

### Imported artifacts for this ticket

1. `sources/local/01-glazed-vault-bootstrap-example.md`
2. `sources/local/glazed-secret-redaction.patch`
3. `sources/local/glazed-implemented-clean.patch`

## Problem Statement

<!-- Describe the problem this design addresses -->

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
