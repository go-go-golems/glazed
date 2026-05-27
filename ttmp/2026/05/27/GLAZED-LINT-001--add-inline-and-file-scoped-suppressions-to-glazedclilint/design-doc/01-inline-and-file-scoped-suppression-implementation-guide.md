---
Title: Inline and file scoped suppression implementation guide
Ticket: GLAZED-LINT-001
Status: active
Topics:
    - cli
    - automation
    - linting
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/analysis/glazedclilint/analyzer.go
      Note: |-
        Analyzer implementation that will parse and honor suppressions
        Analyzer implementation to extend with suppression parsing
    - Path: pkg/analysis/glazedclilint/analyzer_test.go
      Note: |-
        analysistest entry point for suppression behavior
        Existing analysistest entry point
    - Path: pkg/analysis/glazedclilint/testdata/src/a/a.go
      Note: |-
        Fixture file where inline and file scoped suppressions are exercised
        Fixture file for suppression cases
ExternalSources: []
Summary: Design and implementation guide for adding reasoned inline and file scoped suppressions to glazedclilint.
LastUpdated: 2026-05-27T15:10:00-04:00
WhatFor: Use this to implement and review comment-based suppression support before reducing Makefile allow-path scopes across downstream packages.
WhenToUse: Before changing glazedclilint suppression semantics or reducing broad allow-path usage in downstream repositories.
---


# Inline and file scoped suppression implementation guide

## Executive summary

`glazedclilint` currently supports broad suppression through the `-glazedclilint.allow-paths` flag. That works for rollout bootstrap, but it is too coarse for long-term policy enforcement. A directory allow path can hide future violations in production command packages. The analyzer needs comment-based suppression so maintainers can exempt a specific legacy line, declaration, or file while keeping the rest of the package under policy.

The implementation should add two comment forms:

```go
//glazedclilint:ignore <reason>
```

and:

```go
//glazedclilint:file-ignore <reason>
```

The reason must be non-empty. A bare suppression is not honored and should produce an analyzer diagnostic. Inline `ignore` applies to the next syntax node or the same line. File `file-ignore` applies to the entire file. This gives downstream repositories a way to replace broad Makefile allow paths with file-level or line-level documented exceptions.

## Problem statement

The INFRA-002 rollout found a real issue in `discord-bot`: an allow path for `cmd/discord-bot/` skipped every current and future file in the primary command package. That path made the rollout pass, but it also weakened future enforcement. If a new file under `cmd/discord-bot/` introduced raw pflag or raw environment access, CI would not catch it.

Exact file paths are better than broad directories, but sometimes even a whole file is too much. A mostly compliant file may contain one legacy raw flag definition or one bootstrap-time environment read. Today the only available choices are:

1. migrate the code immediately;
2. allow-list the containing file or directory;
3. accept the diagnostic and fail CI.

The missing option is a local, reviewed suppression with a reason.

## Proposed syntax

### Same-line or next-node suppression

```go
value := os.Getenv("DISCORD_TOKEN") //glazedclilint:ignore legacy bootstrap before command settings exist
```

```go
//glazedclilint:ignore legacy pflag adapter; migrate after INFRA-002
cmd.Flags().StringVar(&path, "config", "", "config path")
```

The suppression applies to the diagnostic whose position is on the same line as the comment or inside the next AST node after the comment.

### File scoped suppression

```go
//glazedclilint:file-ignore generated legacy Cobra bridge; tracked by INFRA follow-up
```

This suppresses all diagnostics in the file. It should be used sparingly. It is still better than a Makefile directory allow path because the exception is attached to the file and contains a reason.

### Invalid suppression

```go
//glazedclilint:ignore
```

This should not suppress anything. The analyzer should report:

```text
glazedclilint suppression requires a reason
```

## Design decisions

### Suppressions require a reason

Suppressions without reasons turn the comment mechanism into an invisible allow path. The reason is part of the review contract. It explains why the code is still exempt and gives future maintainers a starting point for migration.

### Suppression matching is position-based

The analyzer already reports diagnostics by AST position. The suppression logic should answer one question: should a diagnostic at this position be skipped? It does not need to change the existing rule logic.

### File-level suppressions are explicit

A file-level suppression should use `file-ignore`, not an `ignore` comment at the top of the file. This prevents accidental whole-file exemptions from comments that were intended for one declaration.

### Path allow lists remain supported

The existing `allow-paths` flag should remain. It is still useful for framework internals, generated compatibility areas, and initial rollouts. Comment suppressions are an additional narrower mechanism, not a replacement in this ticket.

## Implementation plan

1. Extend `fileMeta` to include suppression metadata.
2. Parse comments in `buildFileInfo`.
3. Add `suppressionSet` with:
   - file-wide boolean and reason;
   - ignored lines;
   - next-node suppression positions;
   - invalid suppression comment positions.
4. Add `suppressed(pass, fileInfo, pos)` helper.
5. Change diagnostic emission from `pass.Reportf` to a helper that checks suppressions first.
6. Report invalid bare suppression comments once per file.
7. Add analysistest fixtures for:
   - same-line ignore;
   - previous-line ignore;
   - file-ignore;
   - invalid ignore with expected diagnostic;
   - invalid file-ignore with expected diagnostic.
8. Run focused and full tests:

```bash
go test ./pkg/analysis/glazedclilint -count=1
go test ./... 
```

## Pseudocode

```go
func report(pass, fileInfo, pos, message) {
    if shouldSuppress(pass, fileInfo, pos) {
        return
    }
    pass.Reportf(pos, message)
}

func shouldSuppress(pass, fileInfo, pos) bool {
    filename := pass.Fset.Position(pos).Filename
    line := pass.Fset.Position(pos).Line
    meta := fileInfoByFilename[filename]
    if meta.suppressions.fileIgnore {
        return true
    }
    if meta.suppressions.ignoredLines[line] {
        return true
    }
    if meta.suppressions.ignoredLines[line-1] && pos belongs to next node {
        return true
    }
    return false
}
```

The actual implementation can avoid complex ownership logic by mapping an ignore comment to the line of the next AST node using `ast.Inspect`.

## Tasks

- [ ] Create ticket and docs.
- [ ] Design suppression syntax.
- [ ] Implement suppression parsing.
- [ ] Add report helper and wire existing diagnostics through it.
- [ ] Add invalid suppression diagnostics.
- [ ] Add analysistest fixtures.
- [ ] Run focused tests.
- [ ] Run full tests.
- [ ] Commit implementation.
- [ ] Update diary and changelog.
- [ ] Run docmgr doctor.

## Review instructions

Start with `pkg/analysis/glazedclilint/analyzer.go`. Review suppression parsing before reviewing rule changes. The main correctness question is whether suppression matching is narrow enough: same-line, next-node, or explicit file-wide only.

Then review `pkg/analysis/glazedclilint/testdata/src/a/a.go`. The fixtures should make the intended semantics obvious without relying on prose.
