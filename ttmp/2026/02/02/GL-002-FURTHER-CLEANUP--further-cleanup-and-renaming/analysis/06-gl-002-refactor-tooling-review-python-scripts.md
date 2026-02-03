---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/01-exhaustive-parameter-layer-audit.py
      Note: Audit script analyzed
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/22-rename-doc-terms.py
      Note: Doc rename tool analyzed
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/25-sanitize-remarkable-bundle.py
      Note: Bundle sanitizer analyzed
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# GL-002 Refactor Tooling Review: Python Scripts

## Purpose

This document reviews all Python scripts created during the GL‑002 refactor. It evaluates their utility, correctness, limitations, and reusability, and identifies which should be retained as reusable tools vs. one‑off fixes. The goal is to extract engineering value from the tooling work, not just the refactor outcome.

## Evaluation criteria

Each script is assessed using the following criteria:

- **Scope clarity**: Are inputs/targets explicit and bounded?
- **Safety**: Does it avoid binary files, vendor/ttmp, or unexpected paths?
- **Idempotency**: Can it be re‑run safely?
- **Precision**: AST‑aware vs. regex/string replacement.
- **Configurability**: CLI flags, path overrides, dry run.
- **Traceability**: Emits a report/log of changes.
- **Reusability**: Likely useful in future refactors.

## Script catalog

### 1) Inventory and reporting

#### `01-exhaustive-parameter-layer-audit.py`
- **Purpose**: Full repo scan (all files) for `parameter`, `layer`, and legacy `glazed.parameter` tags; writes a detailed Markdown report with per‑file matches.
- **Strengths**: Comprehensive, binary‑aware, captures line‑level evidence, updates frontmatter timestamps.
- **Weaknesses**: Hard‑coded output path; scans everything including non‑code; no CLI.
- **Reusability**: **High**. With a CLI and configurable patterns, this is a solid general audit tool.

#### `11-layer-parameter-inventory.py`
- **Purpose**: Inventory files containing layer/parameter terms and extract Go identifiers or doc snippets. Outputs a structured Markdown summary.
- **Strengths**: Differentiates file types (go/doc/data/other); extracts identifiers from Go.
- **Weaknesses**: Regex‑only (no AST); can misclassify identifiers; default output path required via `--output`.
- **Reusability**: **Medium‑High**. Useful as a lightweight term‑inventory tool; could be upgraded with AST parsing.

#### `03-render-symbol-inventory.py`
- **Purpose**: Render a JSON symbol inventory into a readable Markdown report; includes top identifiers, per‑file inventory, and file index via `rg`.
- **Strengths**: Converts raw JSON into a structured report; includes rg file index.
- **Weaknesses**: Depends on a pre‑existing JSON file; tight coupling to one ticket path.
- **Reusability**: **Medium**. Good as a report renderer if generalized.

### 2) Code rename passes (regex/string based)

These are fast, targeted renames. They are **not AST‑aware**, so precision depends on careful patterns and scope.

#### `06-rename-parsed-fields.py`
- **Purpose**: Bulk rename ParsedParameter* → FieldValue* in Go files.
- **Strengths**: Explicit regex word boundaries, skips vendor/ttmp.
- **Weaknesses**: Regex could touch comments or string literals unintentionally.
- **Reusability**: **Medium**. Good template for targeted vocabulary renames.

#### `07-rename-decode-into.py`
- **Purpose**: Replace `.InitializeStruct(` with `.DecodeInto(` in Go code.
- **Strengths**: Tiny, focused; directory skipping.
- **Weaknesses**: Blind substitution in code; misses method name changes outside this pattern.
- **Reusability**: **Low‑Medium**. Useful as a one‑off for specific renames.

#### `08-rename-dot-parameters.py`
- **Purpose**: Replace `.Parameters` with `.Fields` in Go code.
- **Strengths**: Simple; uses word boundary.
- **Weaknesses**: Could hit unrelated `Parameters` structs if they exist.
- **Reusability**: **Low‑Medium**.

#### `09-rename-values-decode.py`
- **Purpose**: Replace `vals.DecodeInto(...)` with `vals.DecodeSectionInto(...)` for common variable names.
- **Strengths**: Pattern‑based to reduce over‑matching.
- **Weaknesses**: Only matches a known set of variable names; may miss cases.
- **Reusability**: **Low** unless generalized.

#### `10-rename-yaml-parameters.py`
- **Purpose**: Replace YAML key `parameters:` → `fields:` in sources tests.
- **Strengths**: Narrow scope; minimal risk.
- **Weaknesses**: Extremely specific and hard‑coded path.
- **Reusability**: **Low** (ticket‑specific).

#### `13-rename-schema-tests.py`
- **Purpose**: Replace `.AppendLayers`/`.PrependLayers` in old schema tests.
- **Strengths**: Targeted file list; safe.
- **Weaknesses**: Hard‑coded path to legacy test files.
- **Reusability**: **Low**.

#### `14-rename-sources-test-yaml.py`
- **Purpose**: Rewrite YAML test fixtures under `pkg/cmds/sources/tests` with section/field vocabulary.
- **Strengths**: Regex‑based replacements include pluralization and numeric suffixes.
- **Weaknesses**: Regex language is complex and can cause unintended replacements; no YAML parsing.
- **Reusability**: **Medium** if generalized to YAML parsing.

#### `15-rename-custom-profiles-test.py`
- **Purpose**: Update a single Go test for new naming in sources custom profiles.
- **Strengths**: Specific and safe.
- **Weaknesses**: Hard‑coded path; not reusable.
- **Reusability**: **Low**.

#### `16-rename-schema-tests.py`
- **Purpose**: Update schema test names and variables (layer → section) in specific files.
- **Strengths**: Explicit target files; safe.
- **Weaknesses**: Regex over‑reach if reused broadly.
- **Reusability**: **Low**.

#### `17-rename-settings-language.py`
- **Purpose**: Update settings package vocabulary (Parameter/Layer → Field/Section).
- **Strengths**: Focused to package; consistent word boundary rules.
- **Weaknesses**: String‑based; may touch prose or error messages without intent.
- **Reusability**: **Medium** within a bounded package.

#### `19-rename-fields-language.py`
- **Purpose**: Fix field‑package vocabulary; rename test data file names and internal identifiers.
- **Strengths**: Covers both code and file name references; high impact.
- **Weaknesses**: Hard‑coded replacements; not AST‑aware.
- **Reusability**: **Medium**.

#### `20-rename-values-tests.py`
- **Purpose**: Update values test names (layer → section) and parsed values naming.
- **Strengths**: Focused to a single file.
- **Weaknesses**: Includes broad regex replacements (`Layer`/`layer`), risk of over‑matching if applied elsewhere.
- **Reusability**: **Low**.

#### `21-rename-field-types-example.py`
- **Purpose**: Rename the parameter‑types example into field‑types (identifiers, tag names, strings).
- **Strengths**: Ordered replacements (longest first) to avoid partial collisions.
- **Weaknesses**: Hard‑coded to a single path; not generalized.
- **Reusability**: **Low‑Medium** (good pattern for careful replace ordering).

#### `22-rename-doc-terms.py`
- **Purpose**: Large doc rewrite across markdown, renaming many terms and paths.
- **Strengths**: Centralized replacement table, broad coverage.
- **Weaknesses**: Regex‑free string replace can change semantics; no fenced‑code awareness; high risk of false positives.
- **Reusability**: **Medium** if augmented with doc parsing.

#### `23-rename-initialize-struct-docs.py`
- **Purpose**: Replace `InitializeStruct` -> `DecodeSectionInto` in docs while avoiding `InitializeStructFrom...`.
- **Strengths**: Uses negative lookahead to avoid incorrect replacements.
- **Weaknesses**: Still global string replacement; no code‑block awareness.
- **Reusability**: **Medium**.

#### `24-update-docs-addfields-credentials.py`
- **Purpose**: Fix specific doc issues: `AddFlags` → `AddFields`, `credentials-param` → `credentials-field`, plus identifier changes.
- **Strengths**: Targeted and safe; ideal for cleanup after broad refactors.
- **Weaknesses**: Hard‑coded paths and terms.
- **Reusability**: **Low**.

### 3) Bundle / publishing scripts

#### `25-sanitize-remarkable-bundle.py`
- **Purpose**: Prepare sanitized Markdown for reMarkable PDF conversion (escape `\n`, `\t`, normalize smart quotes). Also includes external docs in the bundle.
- **Strengths**: Solves pandoc LaTeX failures; handles non‑ticket docs via `extras`.
- **Weaknesses**: Hard‑coded input list; no CLI; no dry run.
- **Reusability**: **High** if generalized into a standard “sanitize + bundle” step.

#### `26-build-postmortem-appendix.py`
- **Purpose**: Generate appendices for postmortem (commit list, rename list, symbol map, scripts inventory).
- **Strengths**: Uses git introspection + rename map to make postmortem traceable.
- **Weaknesses**: Repo‑root path was brittle (required fixes); no CLI.
- **Reusability**: **Medium‑High** if turned into a CLI tool for refactor reporting.

---

## What was good overall

- **Speed and specificity**: Many scripts were small and quick to execute.
- **Focused scope**: Most scripts limited scope to a package or file group.
- **Cumulative power**: The scripts provided a pipeline without building a complex framework.

## What was bad or risky

- **Hard‑coded paths**: Almost every script hard‑coded repo or ticket paths.
- **Regex over AST**: For Go code, regex substitutions risk unintended edits (comments/strings).
- **No dry‑run / diff output**: Most scripts overwrite files without reporting diffs.
- **Idempotency not guaranteed**: Some replacements could repeatedly mutate text (rare, but possible).

## What should be reused

- **Inventory scripts** (`01`, `03`, `11`) as reusable analysis tools once parameterized.
- **Bundle sanitizer** (`25`) as a standard reMarkable workflow step.
- **Postmortem appendix builder** (`26`) as part of a standardized refactor documentation pipeline.

## What should be rebuilt as proper tools

- **Go renames**: Replace regex scripts with a unified AST rename tool.
- **Doc rewrites**: Parse Markdown, rewrite fenced code blocks with language‑aware transforms.
- **YAML/JSON rewrites**: Use structured parsers, not string replacement.

## Interesting patterns worth keeping

- **Ordered replacement strategy** (script 21) to avoid partial replacements.
- **Negative lookahead in regex** (script 23) to avoid renaming similar identifiers.
- **Per‑package scope** (scripts 17, 19) to keep refactors bounded.

## Recommendations

1) **Create a shared refactor toolkit** with CLI parameters for root paths, output files, and dry runs.
2) **Centralize replacements** into a single YAML mapping plus AST rewrite for Go code.
3) **Add structured doc and config rewriting** to reduce risk of semantic breakage.
4) **Require a report** from every script (files changed, replacements count).

---

## Appendix: Script list (GL‑002)

- 01-exhaustive-parameter-layer-audit.py
- 03-render-symbol-inventory.py
- 04-rename-add-layer-to-section.py
- 06-rename-parsed-fields.py
- 07-rename-decode-into.py
- 08-rename-dot-parameters.py
- 09-rename-values-decode.py
- 10-rename-yaml-parameters.py
- 11-layer-parameter-inventory.py
- 13-rename-schema-tests.py
- 14-rename-sources-test-yaml.py
- 15-rename-custom-profiles-test.py
- 16-rename-schema-tests.py
- 17-rename-settings-language.py
- 19-rename-fields-language.py
- 20-rename-values-tests.py
- 21-rename-field-types-example.py
- 22-rename-doc-terms.py
- 23-rename-initialize-struct-docs.py
- 24-update-docs-addfields-credentials.py
- 25-sanitize-remarkable-bundle.py
- 26-build-postmortem-appendix.py
