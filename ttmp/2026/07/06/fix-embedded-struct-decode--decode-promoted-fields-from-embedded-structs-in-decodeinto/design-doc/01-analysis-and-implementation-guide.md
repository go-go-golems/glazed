---
Title: Analysis and Implementation Guide
Ticket: fix-embedded-struct-decode
Status: active
Topics:
    - bug
    - fields
    - parsing
    - reflection
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cmds/fields/initialize-struct.go
      Note: DecodeInto and StructToDataMap both skip anonymous (embedded) struct fields
    - Path: pkg/cmds/fields/initialize-struct_test.go
      Note: |-
        existing DecodeInto / StructToDataMap tests (patterns to mirror)
        existing DecodeInto/StructToDataMap tests to mirror
    - Path: ttmp/2026/07/06/fix-embedded-struct-decode--decode-promoted-fields-from-embedded-structs-in-decodeinto/scripts/02-visible-fields-probe.go
      Note: proves VisibleFields includes promoted-from-unexported-embed and excludes shadowed fields (basis for the Step 4 fix)
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/597
Summary: DecodeInto silently skips glazed-tagged fields on embedded structs. Walk anonymous struct fields recursively (and apply the same fix to the symmetric StructToDataMap).
LastUpdated: 2026-07-06T00:00:00Z
WhatFor: 'Fixing the embedded-struct silent-skip bug reported in issue #597'
WhenToUse: When a glazed-tagged field on an embedded struct decodes to its zero value with no error
---



# Analysis and Implementation Guide

## Executive Summary

`FieldValues.DecodeInto` iterates only **direct** struct fields. An embedded
(anonymous) struct field has no `glazed` tag of its own, so the loop `continue`s
past it and never visits its promoted fields. A `glazed:`-tagged field on an
embedded struct is therefore silently skipped during decoding — with no error —
leaving settings at their zero value. The fix is to recurse into anonymous
struct fields (value or pointer, allocating nil pointers) and decode their
tagged fields against the same `FieldValues`. The symmetric `StructToDataMap`
has the identical silent-skip bug and gets the same treatment.

## Problem Statement

`pkg/cmds/fields/initialize-struct.go` (`DecodeInto`):

```go
for i := 0; i < st.NumField(); i++ {
    structField := st.Field(i)
    tag, ok := structField.Tag.Lookup("glazed")
    if !ok {
        continue          // <-- embedded struct's promoted fields never reached
    }
    ...
}
```

`st.Field(i)` for an embedded struct returns the anonymous field itself; it has
no `glazed` tag, so the loop `continue`s and the embedded struct's tagged fields
are never decoded.

### Reproduction (from issue #597)

```go
type commonSettings struct {
    DB string `glazed:"db"`
}

type ServeSettings struct {
    commonSettings   // embedded for reuse
    Listen string `glazed:"listen"`
}
```

Decoding `ServeSettings` with `--db /tmp/x.db --listen :8080` sets `Listen`
correctly but leaves `DB == ""` with no error. In llm-proxy this caused an empty
DB path to be passed to `sql.Open("sqlite3", ...)`, which created a phantom file
literally named `?_foreign_keys=on&_busy_timeout=5000`.

## Current-State Architecture (evidence)

- `DecodeInto` (line ~146) iterates `st.NumField()`, looks up each field's
  `glazed` tag, and on `!ok` does `continue`. It already recurses for **named**
  tagged struct fields via `setTargetValue` (`if dst.Kind() == reflect.Struct {
  return p.DecodeInto(dst.Addr().Interface()) }`), so the recursion machinery
  exists — only the anonymous-field entry point is missing.
- `StructToDataMap` (line ~450) has the same iteration pattern and the same
  silent skip for anonymous fields. It is the symmetric counterpart
  (struct → `map[string]interface{}`) and is exercised by tests but has no
  internal callers.
- `setTargetValue` already handles pointer allocation (`reflect.New`) and
  dereferencing, so the embedded-field helper can reuse the same approach.

## Proposed Solution

In `DecodeInto`, before the `if !ok { continue }` check, detect anonymous
fields and recurse into them:

```go
if structField.Anonymous && !ok {
    if err := p.decodeEmbedded(v.Field(i)); err != nil {
        return errors.Wrapf(err, "failed to decode embedded field %s", structField.Name)
    }
    continue
}
```

where `decodeEmbedded` dereferences (and allocates) a pointer-to-struct, then
re-enters `DecodeInto` on the embedded struct so all of its tagged fields
(including nested pointers, wildcards, `from_json`) are decoded against the same
`FieldValues`:

```go
func (p *FieldValues) decodeEmbedded(field reflect.Value) error {
    if field.Kind() == reflect.Ptr {
        if field.IsNil() {
            field.Set(reflect.New(field.Type().Elem()))
        }
        field = field.Elem()
    }
    if field.Kind() != reflect.Struct {
        return nil
    }
    return p.DecodeInto(field.Addr().Interface())
}
```

Apply the symmetric fix to `StructToDataMap`: refactor the field walk into a
`structValueToDataMap(v reflect.Value)` helper, and for anonymous struct fields
recurse and merge the embedded map into the result.

## Design Decisions

- **Recurse only when `Anonymous && !ok`.** A tagged anonymous field is a
  degenerate case; preserving the existing tag-driven path for it keeps the
  change minimal and behavior-preserving. The reported bug is the *untagged*
  embedded struct, which is now recursed.
- **Reuse `DecodeInto` for recursion.** The existing pointer/wildcard/`from_json`
  handling is exercised unchanged on the embedded struct.
- **Allocate nil pointer-to-struct embedded fields** so a `*commonSettings`
  embedded field decodes correctly (mirrors `setTargetValue`).
- **Also fix `StructToDataMap`** (same bug class, symmetric pair). It has no
  internal callers, so the change is low-risk and keeps struct↔map round-trips
  consistent.
- **No backwards-compat shim.** Embedded fields were previously skipped with no
  error, so there is no working behavior to preserve; the new behavior is
  strictly more correct.

## Alternatives Considered

- **Return an error on unhandled embedded structs** (issue option 2). Rejected:
  promotes shared-settings structs are a common, ergonomic pattern; decoding
  them is preferable to forcing callers to inline fields (the current llm-proxy
  workaround).
- **Fix only `DecodeInto`.** Rejected: leaves the symmetric `StructToDataMap`
  silently lossy for embedded structs, so a decoded-then-reserialized struct
  would drop embedded fields.

## Implementation Plan

1. Add `decodeEmbedded` helper and the `Anonymous && !ok` branch in `DecodeInto`.
2. Refactor `StructToDataMap` to delegate to `structValueToDataMap`, recursing
   into anonymous struct fields and merging results.
3. Add regression tests in `initialize-struct_test.go`:
   - `DecodeInto` with an embedded struct (value and pointer) sets the promoted
     `glazed:` field.
   - `StructToDataMap` with an embedded struct includes the promoted field.
4. `gofmt`, `go test ./pkg/cmds/fields/... -count=1`, then `go test ./...`.
5. Commit fix + tests, open PR referencing #597, post Bluesky via `goat`.

## Testing and Validation

- New tests assert the promoted `db`/`listen` fields decode (value embed) and
  that a `*commonSettings` (pointer embed) is allocated and filled.
- Existing `DecodeInto` / `StructToDataMap` tests continue to pass (no
  anonymous fields in them).
- `go test ./... -count=1` is the acceptance gate.

## Open Questions

None. The issue's preferred option (decode promoted fields) is implemented.

## References

- Issue: https://github.com/go-go-golems/glazed/issues/597
- `pkg/cmds/fields/initialize-struct.go` (`DecodeInto`, `StructToDataMap`)
