---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: glazed/pkg/doc/tutorials/migrating-to-facade-packages.md
      Note: Final migration guidance referenced
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/reference/01-diary.md
      Note: Primary source timeline
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Engineering Postmortem: GL-002 Refactor Execution

## Executive summary

GL‑002 completed a **no‑compat** migration of the Glazed API from the legacy layers/parameters vocabulary to the schema/fields/values/sources vocabulary. The effort touched core APIs, helpers, examples, docs, prompts, and configuration fixtures. The refactor succeeded technically (build/test/lint passed) and operationally (migration playbook, postmortem, and audit artifacts were generated and uploaded to reMarkable). The cost was high in tooling complexity: multiple one‑off scripts, repeated gofmt runs, and a reMarkable bundle pipeline that required a sanitizer to avoid pandoc LaTeX failures.

This postmortem is engineering‑oriented: it reconstructs the work by phase, identifies critical incidents, and outlines improvements for future refactors.

---

## Goals and constraints

**Goals**
- Remove all legacy layer/parameter naming from the public API (no backward compatibility).
- Update code, docs, and examples to a consistent schema/fields/values/sources vocabulary.
- Provide a deterministic migration playbook and an audit trail.

**Constraints**
- **No backward compatibility**: aliases were removed, not preserved.
- Documentation must be exhaustive and traceable (ticket diary, audits, postmortem).
- Scripts must live in the ticket `scripts/` folder.
- reMarkable uploads must contain the full ticket bundle.

---

## Timeline (phased execution)

### Phase 1: Planning & inventory (Steps 1–3)

**Work**
- Created GL‑002 ticket, imported cleanup notes, drafted refactor plan.
- Generated exhaustive audit of layer/parameter mentions; created symbol inventory.
- Switched struct tags to `glazed:` only and updated tooling to rewrite tags.

**Key artifacts**
- `analysis/01-exhaustive-parameter-layer-audit.md`
- `analysis/02-parameter-layer-symbol-inventory.md`
- `design-doc/01-further-cleanup-and-renaming-plan.md`

**Why it mattered**
- The inventory established the full scope and prevented blind spot regressions.
- Early tag changes forced tool updates and clarified the intended schema tag.

---

### Phase 2: Core API renames and tool-driven sweeps (Steps 6–12)

**Work**
- Expanded the AST rename tool and applied it across code paths.
- Renamed parsed parameter types to field values; updated values API.
- Renamed pattern mapper fields, appconfig language, and CLI flags.
- Updated Lua conversion APIs and logging section terminology.

**Key changes**
- `layers` → `schema`
- `parameters` → `fields`
- `ParsedLayers` → `Values`, `ParsedLayer` → `SectionValues`
- `InitializeStruct` → `DecodeSectionInto`
- `TargetParameter` → `TargetField`

**Why it mattered**
- These changes removed the last public API references to layers/parameters and aligned the internal and external mental model.

---

### Phase 3: Docs, examples, and migration guidance (Steps 13, 15–17)

**Work**
- Cleared doc references to `AddFlags` and legacy `credentials-param` examples.
- Updated the field types example and renamed example folder.
- Produced a detailed postmortem with appendices (commit list, renames, mapping).
- Rewrote the migration playbook into a full, no‑compat guide.

**Why it mattered**
- Documentation drift is the most common regression after refactors; the updated playbook and postmortem minimize that risk.

---

### Phase 4: Packaging and reMarkable delivery (Steps 14, 16, 19–20)

**Work**
- Built reMarkable bundle and discovered pandoc LaTeX failure due to inline `\n` and smart quotes.
- Added sanitizer script to normalize quotes and escape sequences.
- Re‑uploaded updated bundles as v2/v3/v4 with new artifacts (postmortem appendices, migration playbook, refactor blueprint).

**Why it mattered**
- The project required a stable, reviewable artifact on the tablet; the sanitizer became necessary infrastructure.

---

### Phase 5: PR finalization (Step 18)

**Work**
- Updated PR 524 description with a multi‑paragraph summary and testing info.
- `gh pr edit` failed due to classic Projects GraphQL API; REST patch used instead.

**Why it mattered**
- The PR description is the primary reviewer entry point; it needed a clear narrative after a large refactor.

---

## Critical incidents and root cause analysis

### Incident: alias removal broke sources example
- **Symptom**: `cmd/examples/sources-example/main.go: undefined: cmds.NewCommandDefinition`
- **Cause**: Removing alias without running tests early.
- **Fix**: Replaced `NewCommandDefinition` with `NewCommandDescription`.
- **Lesson**: Run `go test ./...` immediately after alias removal.

### Incident: lint failures (gofmt)
- **Symptom**: golangci‑lint reported gofmt failures across renamed files.
- **Cause**: AST + regex scripts created formatting changes without gofmt sweep.
- **Fix**: Explicit `gofmt -w` on touched files.
- **Lesson**: Always gofmt after automated code rewrites.

### Incident: reMarkable bundle pandoc failure
- **Symptom**: `Undefined control sequence` for `\n` in inline code.
- **Cause**: LaTeX sensitivity to raw `\n` and smart quotes in Markdown.
- **Fix**: Sanitizer script to escape and normalize before bundling.
- **Lesson**: Make sanitization a standard step for large bundles.

### Incident: PR update tooling failure
- **Symptom**: `gh pr edit` failed due to Projects classic deprecation.
- **Cause**: GitHub GraphQL API field removal.
- **Fix**: REST patch via `gh api -X PATCH`.
- **Lesson**: Keep a REST fallback for GH CLI operations.

---

## Tooling impact analysis

### What worked
- AST renames (Go) for safe symbol changes.
- Regex‑based scripts for scoped fixes.
- Reusable audit scripts for evidence generation.

### What hurt
- Hard‑coded script paths and repo root assumptions.
- Regex replacement in docs (risk of semantic drift).
- Multiple one‑off scripts without a shared CLI framework.

### Mitigation for future
- Build a shared refactor toolkit with:
  - root discovery
  - dry runs + diff summaries
  - AST + doc parsing passes
  - centralized rename mappings

---

## Engineering outcomes

### Codebase
- Legacy vocabulary removed from non‑ttmp code.
- API surface consistently uses schema/fields/values/sources.

### Docs and migration
- Exhaustive migration playbook created and kept in sync with rename map.
- Postmortem and appendices provide traceability (renames, commits, scripts).

### Tooling assets
- Inventory and audit scripts now exist for future refactors.
- Sanitizer and appendix builder are reusable for documentation workflows.

---

## Key commits (from diary)

- `2425fdd` Docs: add exhaustive parameter/layer audit
- `bf850ec` Docs: add GL‑002 renaming plan and notes
- `1480995` Refactor: inventory symbols and rename AddLayerToCobraCommand
- `5766e46` Refactor: update appconfig and CLI section naming
- `7874112` Refactor: rename Lua section/field conversion
- `6844cbf` Refactor: finish section/field cleanup
- `8a9c26e` Docs: add remarkable bundle sanitizer
- `37b787d` Docs: expand GL‑002 postmortem examples
- `e455738` Docs: add postmortem appendix + update bundle script
- `cebc160` Docs: expand migration playbook
- `c76ce17` Docs: include migration playbook in bundle sanitizer
- `7ad70d8` Docs: add refactor infrastructure blueprint

---

## Recommendations for future refactors

1) **Create a refactor index database** before touching code (symbols, strings, docs, tags).
2) **Run gofmt automatically** after each automated code pass.
3) **Use structured doc parsing** instead of raw string replacements.
4) **Package scripts into a CLI** with consistent flags and dry‑run support.
5) **Standardize bundle sanitization** for reMarkable workflows.

---

## Appendix: Referenced artifacts

- Diary: `ttmp/.../reference/01-diary.md`
- Migration playbook: `pkg/doc/tutorials/migrating-to-facade-packages.md`
- Postmortem: `analysis/04-postmortem-gl-002-refactor-and-tooling.md`
- Refactor blueprint: `analysis/05-refactor-infrastructure-blueprint-data-tools-human-oversight.md`
