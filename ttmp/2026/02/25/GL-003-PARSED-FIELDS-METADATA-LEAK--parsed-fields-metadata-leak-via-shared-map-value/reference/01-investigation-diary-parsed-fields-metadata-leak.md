---
Title: 'Investigation diary: parsed-fields metadata leak'
Ticket: GL-003-PARSED-FIELDS-METADATA-LEAK
Status: active
Topics:
    - glazed
    - security
    - metadata
    - config
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: geppetto/pkg/sections/profile_registry_source.go
      Note: profile registry middleware in runtime parse chain
    - Path: pinocchio/pkg/cmds/helpers/parse-helpers.go
      Note: pinocchio helper wiring config/profile middlewares used in repro
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-25T20:03:32.683513191-05:00
WhatFor: ""
WhenToUse: ""
---


# Investigation diary: parsed-fields metadata leak

## Goal

Investigate why `pinocchio --print-parsed-fields` shows API-key-like values under unrelated parsed-field metadata, identify the root cause, and produce a bug report ticket for `glazed`.

## Context

User report (paraphrased): running `go run ./cmd/pinocchio ... --print-parsed-fields` showed secrets in parsed field logs, likely from Glazed matching/metadata.

Repository layout note: command had to be run from `pinocchio/` submodule, not monorepo root.

## Quick Reference

### Reproduction commands used

```bash
cd /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/pinocchio
go run ./cmd/pinocchio code professional \
  --profile-registries /tmp/foo.yaml \
  --profile gpt-5-mini \
  hello --print-parsed-fields \
  | yq '."ai-chat"."ai-engine"'
```

Observed (sanitized):

```yaml
log:
  - source: defaults
    value: gpt-4
  - source: config
    value: gpt-4o-mini
    metadata:
      config_file: /home/manuel/.pinocchio/config.yaml
      index: 0
      map-value: sk-REDACTED
  - source: profiles
    value: gpt-5-mini
```

Cross-field leakage check:

```bash
go run ./cmd/pinocchio code professional \
  --profile-registries /tmp/foo.yaml \
  --profile gpt-5-mini \
  hello --print-parsed-fields \
  | yq -o=json '.' \
  | jq -r 'to_entries[] as $s | ($s.value | to_entries[]?) as $f | (($f.value.log // [])[]? | select(.source == "config")) as $l | "\($s.key).\($f.key)\tvalue=\($l.value|tostring)\tmap=\($l.metadata["map-value"]|tostring)"' \
  | sed -E 's/sk-[A-Za-z0-9_-]+/sk-REDACTED/g'
```

Observed (sanitized):

```text
ai-chat.ai-api-type            value=openai         map=sk-REDACTED
ai-chat.ai-engine              value=gpt-4o-mini    map=sk-REDACTED
ai-chat.ai-max-response-tokens value=4096           map=sk-REDACTED
claude-chat.claude-api-key     value=sk-REDACTED    map=sk-REDACTED
openai-chat.openai-api-key     value=sk-REDACTED    map=sk-REDACTED
```

### Root-cause chain (code)

1. `glazed/pkg/cmds/fields/gather-fields.go` appends `metadata["map-value"] = v_`.
2. `glazed/pkg/cmds/sources/load-fields-from-config.go` reuses one metadata map (`config_file`, `index`) across many field updates.
3. `glazed/pkg/cmds/fields/field-value.go` `WithMetadata` assigns map pointer directly when metadata is nil, so metadata is shared/mutable across parse steps.
4. `glazed/pkg/cli/helpers.go` prints metadata verbatim in `printParsedFields`.

Conclusion: shared mutable metadata map + raw `map-value` capture causes secret leakage and cross-field contamination.

## Usage Examples

Use this diary when:

1. implementing the fix in `glazed`,
2. writing regression tests for parsed-fields metadata,
3. validating that CLI debug output no longer contains raw secrets.

Minimal validation command after fix:

```bash
go run ./cmd/pinocchio code professional \
  --profile-registries /tmp/foo.yaml \
  --profile gpt-5-mini \
  hello --print-parsed-fields \
  | rg -n 'sk-[A-Za-z0-9_-]+'
```

Expected result: no matches.

## Related

1. `design-doc/01-bug-report-map-value-metadata-leaks-secrets-in-print-parsed-fields.md`
2. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/gather-fields.go`
3. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/field-value.go`
4. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/sources/load-fields-from-config.go`
5. `/home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cli/helpers.go`
