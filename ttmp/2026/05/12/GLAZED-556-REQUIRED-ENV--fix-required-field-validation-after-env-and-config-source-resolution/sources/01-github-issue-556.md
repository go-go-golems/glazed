---
Title: GitHub issue 556 source summary
Ticket: GLAZED-556-REQUIRED-ENV
Status: active
Topics:
    - glazed
    - cli
    - config
    - env
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/556
Summary: "Source summary for issue 556 about required fields validating before env/config source resolution."
LastUpdated: 2026-05-12T13:50:00-04:00
WhatFor: "Use as compact source context for the issue 556 design doc."
WhenToUse: "Read before implementing or reviewing the required-env fix."
---

# GitHub issue 556: Required field validation runs before env/config sources can satisfy fields

Source: https://github.com/go-go-golems/glazed/issues/556

## Summary

`fields.WithRequired(true)` currently fails for env-backed values when the field is not explicitly provided as a Cobra flag. The issue reports that Glazed's source model should allow defaults, config, environment variables, arguments, and Cobra flags to merge first, then validate required fields against the final `values.Values` result.

## Reproduction shape

- Field: `required-name` / `proc-file`, type string, `fields.WithRequired(true)`.
- Parser: `cli.NewCobraParserFromSections` or `cli.BuildCobraCommandFromCommand` with `CobraParserConfig{AppName: "..."}` and no custom middleware override.
- Env var: app prefix + field name, for example `REQ_ENV_TEST_REQUIRED_NAME=from-env` or `DEVMUX_PROC_FILE=/tmp/procs.json`.
- No explicit Cobra flag.

## Actual behavior

Parsing fails before env can apply:

```text
Field required-name is required
```

## Expected behavior

Parsing succeeds and parsed-field provenance records `source: env` for the required field. Missing required values should still fail, but only after all configured sources have run.

## Additional comment evidence

A downstream `go-go-host` reproduction confirmed the same behavior and pointed to `pkg/cmds/fields/cobra.go`, where `GatherFlagsFromCobraCommand` checks `cmd.Flags().Changed(flagName)` and returns `Field %s is required` before the env middleware runs.
