---
Title: Diary
Ticket: fix-env-prefix-dashes
Status: active
Topics:
    - env
    - bug
    - sources
    - parsing
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/cli/cobra-parser.go
      Note: |-
        built-in path passes strings.ToUpper(AppName) as the env prefix
        how AppName becomes the env prefix
    - Path: pkg/cmds/sources/update.go
      Note: |-
        updateFromEnv — the single shared env-key assembly helper (bug site)
        bug site
    - Path: pkg/cmds/sources/update_test.go
      Note: existing env source tests; mirror patterns for the regression test
ExternalSources:
    - https://github.com/go-go-golems/glazed/issues/596
Summary: Chronological investigation and fix for the env-prefix hyphen normalization bug (#596).
LastUpdated: 2026-07-06T00:00:00Z
WhatFor: Record how the env-source hyphen bug was reproduced, fixed, and validated
WhenToUse: When resuming work on fix-env-prefix-dashes or auditing the env source
---


# Diary

## Goal

Record the investigation, fix, and validation of the env-source prefix hyphen
normalization bug reported in glazed issue #596, so the change is reviewable and
continuation-friendly.

## Step 1: Investigate the bug and stand up the ticket/docs

I started by fetching issue #596 and locating the buggy code. The issue is precise:
`updateFromEnv` normalizes hyphens in the field-name portion of the env key but not in
the app-name prefix. I confirmed the code at `pkg/cmds/sources/update.go:156-160` and
traced how the prefix reaches it (the cobra built-in path computes
`strings.ToUpper(AppName)`). Then I created the docmgr ticket, the design doc, and this
diary.

### Prompt Context

**User prompt (verbatim):** "Createa new docmgr ticket to address https://github.com/go-go-golems/glazed/issues/596

Keep a detailed diary as you work, write an analysis/implementation guide (doesn't need to be very long at all).

Then fix it, commit at appropriate intervals, then open PR."

**Assistant interpretation:** Stand up a docmgr ticket for glazed #596, keep a detailed
diary, write a short analysis/implementation guide, then implement the fix, commit at
sensible checkpoints, and open a PR.

**Inferred user intent:** A reproducible, well-documented bug fix that lands as a PR,
with the usual ticket/diary/design-doc paper trail.

**Follow-up prompt (verbatim):** "After the PR is open, write a bluesky social post (use `goat`)"

### What I did
- Fetched issue #596 via `gh api` (`gh issue view` failed due to the classic-projects
  GraphQL deprecation).
- Read `pkg/cmds/sources/update.go` and confirmed the bug at `updateFromEnv`
  (lines ~156-160): `base` is hyphen-normalized, `prefix` is only uppercased.
- Traced callers of `FromEnv`/`updateFromEnv`: cobra-parser built-in path, runner,
  helpers, vault, and several `cmd/` binaries. Only `updateFromEnv` assembles the final
  env key, so it is the single choke point.
- Read `pkg/cmds/sources/update_test.go` to capture the test pattern
  (`TestUpdateFromEnvParsesTypedValues`, `TestUpdateFromEnvInvalidChoice`).
- Created docmgr ticket `fix-env-prefix-dashes` with topics `env,bug,sources,parsing`,
  added a design doc ("Analysis and Implementation Guide") and this diary.
- Confirmed we are on branch `task/fix-glazed-env-dashes` (already checked out).

### Why
The issue is well-scoped and the suggested fix is correct. Standing up the ticket/docs
first keeps code ↔ docs consistent per the working loop and gives reviewers a short
design rationale before reading the diff.

### What worked
- `gh api repos/go-go-golems/glazed/issues/596 --jq '{...}'` returned the full body
  with the exact buggy snippet and reproduction.
- The code matched the issue's line references almost exactly.

### What didn't work
- `gh issue view 596` failed: `GraphQL: Projects (classic) is being deprecated …`.
  Worked around it by using `gh api` to read the issue JSON directly.

### What I learned
- The env-key assembly is centralized in `updateFromEnv`, so a single fix there covers
  every caller — no need to touch each call site.
- The field-name half of the env key is *already* hyphen→underscore normalized; only the
  prefix half is inconsistent. The fix is to make the prefix half consistent.

### What was tricky to build
Nothing yet at this step — investigation only. The one sharp edge to keep in mind for
the fix: `ReplaceAll(prefix, "-", "_")` must be a no-op for already-underscored /
all-caps prefixes (e.g. `APP`, `DOCSCTL`, `BUILD_WEB`) so existing callers are
unaffected. It is, because there are no hyphens to replace.

### What warrants a second pair of eyes
- Confirm the fix is applied at `updateFromEnv` (covers all callers) and not only at
  the cobra built-in path.
- Confirm no real-world env var could have been matching via a hyphenated prefix
  before (it couldn't — shells can't export hyphenated names), so there is no
  backwards-compat surface to preserve.

### What should be done in the future
N/A

### Code review instructions
- Start at `pkg/cmds/sources/update.go` (`updateFromEnv`).
- Validate understanding by reading `pkg/cli/cobra-parser.go:161-167` (how `AppName`
  becomes the env prefix) and the existing tests in `update_test.go`.

### Technical details
Buggy snippet (`pkg/cmds/sources/update.go`, `updateFromEnv`):

```go
base := sectionPrefix + p.Name
envKey := strings.ToUpper(strings.ReplaceAll(base, "-", "_"))
if prefix != "" {
    envKey = strings.ToUpper(prefix) + "_" + envKey   // prefix NOT normalized
}
```

Reproduction from the issue (`AppName: "llm-proxy"`):

```sh
# Does NOT load (shell can't export the hyphenated name):
export LLM_PROXY_BYOK_SECRET=sk-test
./myapp --print-parsed-fields   # byok-secret is empty

# DOES load (quoted, bypasses the shell parser):
env 'LLM-PROXY_BYOK_SECRET=sk-test' ./myapp --print-parsed-fields
```
