---
Title: 'Bug Report: map-value metadata leaks secrets in print-parsed-fields'
Ticket: GL-003-PARSED-FIELDS-METADATA-LEAK
Status: active
Topics:
    - glazed
    - security
    - metadata
    - config
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: glazed/pkg/cli/helpers.go
      Note: printParsedFields emits metadata verbatim
    - Path: glazed/pkg/cmds/fields/field-value.go
      Note: WithMetadata aliases caller map when parse step metadata is nil
    - Path: glazed/pkg/cmds/fields/gather-fields.go
      Note: injects map-value metadata during map ingestion
    - Path: glazed/pkg/cmds/sources/load-fields-from-config.go
      Note: reuses parse metadata option across many field updates
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-25T20:03:32.444807122-05:00
WhatFor: ""
WhenToUse: ""
---


# Bug Report: map-value metadata leaks secrets in print-parsed-fields

## Executive Summary

`--print-parsed-fields` can expose sensitive values (including API keys) in parse-step metadata. The immediate trigger is `map-value` metadata added during config map ingestion. In practice, this leaks into unrelated fields due to shared metadata map aliasing, so benign fields can display secret-looking values.

This is a security bug and also a correctness bug in parse logs.

## Problem Statement

When loading config/profile values via map-based sources, Glazed records parse log metadata under `map-value`. That metadata currently stores raw source values and is printed verbatim by `printParsedFields`.

Observed behavior in `pinocchio`:

1. Non-secret fields such as `ai-chat.ai-engine` show `metadata.map-value` containing a secret value.
2. Multiple fields show the same `map-value`, indicating metadata mutation/aliasing across parse steps.

Sanitized reproduction snippet:

```text
ai-chat.ai-api-type            value=openai         map=sk-REDACTED
ai-chat.ai-engine              value=gpt-4o-mini    map=sk-REDACTED
ai-chat.ai-max-response-tokens value=4096           map=sk-REDACTED
```

Scope of this report:

1. Root-cause analysis in `glazed` parse/metadata plumbing.
2. Proposed fixes and test coverage guidance.
3. No code change included in this ticket write-up.

## Current-State Analysis (Evidence)

### 1) `map-value` metadata is explicitly attached during map ingest

`GatherFieldsFromMap` appends metadata that includes raw `v_`:

- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/gather-fields.go:56`
- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/gather-fields.go:58`

### 2) Parse options include shared config metadata map

`FromFiles` creates one metadata map per config file (`config_file`, `index`) and passes that parse option through all fields in the file application step:

- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/sources/load-fields-from-config.go:70`
- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/sources/load-fields-from-config.go:72`

### 3) `WithMetadata` stores map pointers without copying

When parse step metadata is nil, `WithMetadata` assigns the incoming map directly:

- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/field-value.go:21`
- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/field-value.go:24`

This means later writes can mutate metadata seen by earlier steps if the same map is reused.

### 4) Parsed field output prints metadata verbatim

`printParsedFields` emits `l.Metadata` directly:

- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cli/helpers.go:82`
- `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cli/helpers.go:83`

## Gap Analysis

Expected behavior:

1. Parse logs should never expose secrets by default.
2. Parse-step metadata should be immutable per step (no cross-step mutation).
3. `map-value` should either be absent or safely redacted.

Actual behavior:

1. Raw sensitive values can be emitted.
2. Metadata from one field can contaminate others.
3. Observability output is both insecure and semantically wrong.

## Proposed Solution

Implement a two-part fix:

1. Metadata safety: stop raw value logging for `map-value` in default paths.
2. Metadata isolation: copy metadata maps when applying parse options.

### API/behavior sketch

```go
// fields.WithMetadata
func WithMetadata(metadata map[string]interface{}) ParseOption {
  return func(p *ParseStep) {
    if p.Metadata == nil {
      p.Metadata = shallowCopyMap(metadata)
      return
    }
    for k, v := range metadata {
      p.Metadata[k] = v
    }
  }
}
```

```go
// GatherFieldsFromMap: remove or sanitize "map-value"
options_ := append(options /* no raw map-value by default */)
```

If keeping a trace token is needed, prefer non-sensitive hints (for example `map-key`, `value-type`, or `redacted: true`) rather than raw values.

## Design Decisions

1. Prioritize secure defaults over rich raw debug metadata.
2. Preserve parse-step traceability (`source`, `config_file`, `index`) while removing sensitive payload leakage.
3. Fix aliasing at the metadata utility level to protect all current and future call sites.

## Alternatives Considered

1. Only filter at print time:
Rejected because secrets would still exist in in-memory logs and other renderers/serializers.

2. Keep `map-value` and rely on field-name based redaction:
Rejected because it is fragile (new fields/providers), and non-secret fields already receive secret values due to aliasing.

3. Fix aliasing only and keep raw values:
Rejected because direct secret exposure risk remains.

## Implementation Plan

Phase 1: Safety and correctness

1. Update `fields.WithMetadata` to copy metadata on initial assignment.
2. Remove `map-value` from `GatherFieldsFromMap` parse metadata, or hard-redact before writing.

Phase 2: Test coverage

1. Add unit test: metadata maps are not aliased across parse steps.
2. Add regression test: config parse logs for non-secret fields never include raw API-key-like value.
3. Add integration test for `--print-parsed-fields` output in a CLI fixture with secret inputs.

Phase 3: Docs and communication

1. Update parsed-fields docs to clarify metadata redaction policy.
2. Add changelog note for security-sensitive behavior change.

## Testing and Validation Strategy

1. Unit tests in `glazed/pkg/cmds/fields`:
   - map aliasing regression.
   - metadata immutability across parse steps.
2. Middleware tests in `glazed/pkg/cmds/sources`:
   - config file parse does not propagate raw map values.
3. End-to-end smoke:
   - run `pinocchio ... --print-parsed-fields` with known secret and assert no secret substring in output.

## Risks, Alternatives, Open Questions

1. Risk: removing `map-value` may reduce debug visibility.
Mitigation: retain non-sensitive metadata (`config_file`, `index`, `source`) and optionally add explicit redacted marker.

2. Open question: should secret redaction be centralized in one sanitizer used by all output modes?
Recommendation: yes, but still remove raw secret capture at source.

## References

Code references:

1. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/gather-fields.go`
2. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/field-value.go`
3. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/sources/load-fields-from-config.go`
4. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cli/helpers.go`

Related ticket docs:

1. `reference/01-investigation-diary-parsed-fields-metadata-leak.md`
