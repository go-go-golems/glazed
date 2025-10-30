# Sober Assessment and Design Proposals: Pattern-Based Config Mapper (Phase 1)

Date: 2025-10-29  
Audience: Maintainers and contributors working on `pkg/cmds/middlewares/pattern-mapper.go`

---

## 1) Purpose and Scope

Provide a concise, implementation-focused assessment of the current pattern-based config mapper (Phase 1) and propose 12 concrete, actionable improvements. This document is derived from the debate in `11-debate-around-the-actual-mapper-implementation.md` and the implemented code/tests.

---

## 2) Confirmed Strengths (Keep as-is)

- Clear, debuggable manual traversal (`matchSegments`, `matchSegmentsRecursive`).
- Backward-compatible via `ConfigMapper` interface and `configFileMapperAdapter`.
- Early validation for pattern syntax, capture references, target layer existence.
- Nested rules with capture inheritance work and reduce duplication.
- Helpful runtime errors (esp. for missing required patterns, unknown parameters).

---

## 3) Key Risks and Pain Points (Observed)

- Non-determinism when wildcards match multiple keys (Go map iteration order).
- Silent overwrites when multiple matches/patterns map to the same parameter.
- Prefix handling obscures error messages and intent in some cases.
- Optional (non-required) patterns fail silently; hard to distinguish “missing vs typo”.
- Static target parameter names are not validated up-front.
- Potential capture shadowing in nested rules is not warned.
- Root config type must be `map[string]interface{}`; arrays not supported (by design).

---

## 4) Design Proposals (12 items)

Each proposal includes: Problem → Change → Rationale → Priority/Effort.

1. Deterministic wildcard matching order
   - Problem: `app.*.api_key` may produce different winners across runs.
   - Change: Sort map keys before iterating when segment is `*` or `{name}`.
   - Rationale: Determinism improves predictability and testability.
   - Priority/Effort: P1 / Low (local sorting in `matchSegments`).

2. Multi-match handling (strict by default)
   - Problem: Multiple matches silently overwrite (last one wins by iteration order).
   - Change: Treat multiple distinct values from a single rule mapping to the same target parameter as an error by default (no policy toggles). If values are identical, mapping succeeds. Prefer captures (e.g., `{env}`) to collect separate values.
   - Rationale: Eliminate nondeterminism and silent data loss; simpler model without configuration.
   - Priority/Effort: P1 / Medium.

3. Collision detection across rules (strict by default)
   - Problem: Different patterns can resolve to the same target parameter and overwrite silently.
   - Change: Detect and error when different rules write to the same target parameter (no policy toggles). If this is intended, refactor rules or target parameter names to avoid collisions.
   - Rationale: Prevents accidental overrides; errors are clearer than warnings.
   - Priority/Effort: P1 / Medium.

4. Prefix-aware error messages
   - Problem: Errors mention unprefixed name while actual lookup used prefixed name.
   - Change: Include resolved, prefix-adjusted parameter in error text (e.g., `api-key (checked as "demo-api-key")`).
   - Rationale: Clarity; aligns error with actual lookup behavior.
   - Priority/Effort: P1 / Low.

5. Early validation for static target parameters
   - Problem: Static `TargetParameter` (no captures) validated only at runtime.
   - Change: In `compileRule`, if no `{...}` in `TargetParameter`, derive prefix-adjusted name and validate existence immediately.
   - Rationale: Fail fast for common static mappings.
   - Priority/Effort: P1 / Low-Medium (implemented).

6. Warn on capture shadowing in nested rules
   - Problem: Child `{env}` can silently override parent `{env}`.
   - Change: During compile, detect duplicate capture names in parent+child; log a warning or require explicit override flag.
   - Rationale: Prevents subtle mapping bugs.
   - Priority/Effort: P2 / Medium (implemented).
   - Notes: Implemented as a compile-time warning to stderr; no runtime policy toggle.

7. Optional pattern diagnostics
   - Problem: Optional patterns that don’t match are silent; hard to tell typo vs missing.
   - Change: Add diagnostics collector or debug logging: record unmatched optional patterns; expose via return value or a debug hook.
   - Rationale: Better debuggability without breaking current semantics.
   - Priority/Effort: P2 / Medium.

8. Improve required pattern error context
   - Problem: Error doesn’t point to nearest existing path or reason.
   - Change: When failing a required pattern, include nearest matched prefix and indicate the missing segment.
   - Rationale: Faster troubleshooting.
   - Priority/Effort: P2 / Medium (implemented).

9. Explicit helper for canonical parameter name resolution
   - Problem: Prefix handling scattered and implicit.
   - Change: Introduce `resolveCanonicalParameterName(layer, target)` helper (or method) centralizing prefix logic.
    - Rationale: Single source of truth; easier to test and reuse.
    - Priority/Effort: P2 / Low (implemented early).

10. Document wildcard semantics and guidance
    - Problem: Docs understate nondeterminism of wildcards and collisions.
    - Change: Update `pattern-based-config-mapping.md` with: (a) sorting/determinism policy, (b) recommendation to prefer captures over wildcards when multiple values are expected, (c) ambiguity and collision behavior (both error by default) and mitigation.
    - Rationale: Set correct expectations; reduce misuse.
    - Priority/Effort: P1 / Low (implemented).

11. Add tests for collisions, deterministic wildcard order, and prefix+captures interplay
    - Problem: Gaps for collisions and ordering behaviors.
    - Change: Add table-driven tests covering: (a) multiple wildcard matches (deterministic; identical values allowed, distinct values error), (b) rule collisions (error), (c) `{env}` capture combined with layer prefix.
    - Rationale: Guard rails for new behaviors.
    - Priority/Effort: P1 / Low-Medium (implemented).

12. Keep regex compilation out of hot path; remove or gate if unused
    - Problem: Regex compiled but unused; minor memory/complexity overhead.
    - Change: Either (a) remove compilation until used, or (b) guard behind build tag / constructor option. Keep validation via simple syntax checks.
    - Rationale: Reduce technical debt; clearer intent.
    - Priority/Effort: P3 / Low.

---

## 5) Proposed Acceptance Criteria (per proposal)

- P1 items: deterministic iteration (1), multi-match handling (2), collision detection (3), prefix-aware errors (4), static param early validation (5), docs update (10), tests (11).  
  - All implemented with linters/tests passing.  
  - New behaviors covered by table-driven tests.  
  - Simplified behavior: ambiguous cases (multi-match with distinct values, collisions) error by default; no runtime policy toggles.
- P2 items: shadowing warning (6), improved required error context (8), canonical param helper (9) implemented; optional diagnostics (7) pending.  
  - (6) warning-only (stderr) with no behavior change; (8) improves error message context only.
- P3 item: regex compilation gating/removal (12).  
  - No functional change; code clarity improvement.

---

## 6) Rollout Plan (Incremental)

- MR 1 (P1):
  - Deterministic key sorting for wildcards/captures.
  - Prefix-aware error messaging.
  - Early validation for static targets.
  - Docs update + tests.
- MR 2 (P1) — simplified strict behavior:
  - Multi-match (distinct values) → error; collisions across rules → error.
  - Remove runtime policy toggles; adjust docs and tests accordingly.
- MR 3 (P2):
  - Capture shadowing warnings (done); improved required error context (done); canonical param-name helper (done early); optional diagnostics collector (pending).
- MR 4 (P3):
  - Regex compilation removal or gating.

---

## 7) Code Review Checklist (to apply during MRs)

- Deterministic traversal: keys sorted only at wildcard/capture segments; avoid global perf regressions.
- No breaking changes to `ConfigMapper` interface or `LoadParametersFromFile` behavior by default.
- Error messages include resolved (prefixed) parameter names where applicable.
- Static target parameters validated at compile time; dynamic ones at runtime.
- Ambiguities error by default: multi-match with distinct values, cross-rule collisions; no runtime policy toggles.
- New tests cover ordering, collisions, prefix+captures, and error messaging.
- Documentation updated and examples aligned with new strict semantics.

---

## 8) Summary

We adopted strict, deterministic semantics for Phase 1: ambiguous wildcard multi-matches (with distinct values) and cross-rule collisions now error by default, removing policy toggles and last-wins behavior. Prefix-aware errors and a canonical parameter resolver improve clarity. Documentation and tests reflect these changes. This simplifies user expectations while preserving the `ConfigMapper` API.
