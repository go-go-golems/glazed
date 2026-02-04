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
      Note: Migration playbook referenced in tooling blueprint
    - Path: glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/12-rename-symbols.yaml
      Note: Rename map referenced for tool-assisted refactor design
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Refactor Infrastructure Blueprint: Data, Tools, Human Oversight

## Purpose

This document answers four practical questions that emerge after a large, no‑compat refactor:

1) **What structured database of code/files/symbols/strings would have been useful?**
2) **What refactoring tools would have helped (self‑contained or framework‑based)?**
3) **Which refactor steps require human supervision?**
4) **What would a tool‑assisted refactor look like if we started from scratch?**

The intent is to provide a concrete engineering blueprint: a database schema, a tooling architecture, a human‑in‑the‑loop model, and an end‑to‑end workflow with pseudocode.

---

## 1) The structured database we should have had

### 1.1 Goals for the database

The database should allow you to answer, with minimal ad‑hoc scripting:

- **Where does a term appear?** (symbol name, string literal, CLI flag, doc snippet)
- **What is the semantic role?** (API symbol vs. local variable vs. prose)
- **How is it used?** (imports, interfaces, method calls, serialization, tags)
- **What will break if we rename it?** (compile dependencies, doc references, test fixtures)
- **Which docs/examples must be updated together?** (code‑doc coupling)
- **What is the delta vs. upstream?** (commit, file rename, symbol rename)

### 1.2 Minimal schema (SQLite)

Below is a minimal but highly actionable schema for refactor‑grade introspection.

```sql
-- Files and paths
CREATE TABLE files (
  file_id INTEGER PRIMARY KEY,
  path TEXT UNIQUE NOT NULL,
  kind TEXT NOT NULL,              -- go, md, yaml, json, sh, etc.
  checksum TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  last_modified TEXT NOT NULL
);

-- Git metadata
CREATE TABLE commits (
  commit_id INTEGER PRIMARY KEY,
  sha TEXT UNIQUE NOT NULL,
  subject TEXT NOT NULL,
  author TEXT,
  date TEXT
);

CREATE TABLE file_commits (
  file_id INTEGER NOT NULL,
  commit_id INTEGER NOT NULL,
  change_type TEXT NOT NULL,       -- A/M/D/R
  old_path TEXT,                   -- for renames
  PRIMARY KEY (file_id, commit_id)
);

-- Parsed symbols (Go AST)
CREATE TABLE symbols (
  symbol_id INTEGER PRIMARY KEY,
  file_id INTEGER NOT NULL,
  package TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL,              -- type, func, method, var, const, field
  receiver TEXT,                   -- for methods
  signature TEXT,                  -- for funcs/methods
  exported INTEGER NOT NULL,
  line INTEGER NOT NULL,
  column INTEGER NOT NULL
);

-- Symbol references (Go AST)
CREATE TABLE references (
  ref_id INTEGER PRIMARY KEY,
  file_id INTEGER NOT NULL,
  symbol_name TEXT NOT NULL,
  context TEXT NOT NULL,            -- call, type-ref, selector, import, etc.
  line INTEGER NOT NULL,
  column INTEGER NOT NULL
);

-- String literals (for flags, error messages, doc keys)
CREATE TABLE strings (
  string_id INTEGER PRIMARY KEY,
  file_id INTEGER NOT NULL,
  value TEXT NOT NULL,
  context TEXT NOT NULL,            -- string, raw, tag, yaml-key, json-key, doc
  line INTEGER NOT NULL,
  column INTEGER NOT NULL
);

-- Struct tags
CREATE TABLE struct_tags (
  tag_id INTEGER PRIMARY KEY,
  file_id INTEGER NOT NULL,
  struct_name TEXT,
  field_name TEXT,
  tag_key TEXT NOT NULL,            -- e.g. glazed
  tag_value TEXT NOT NULL,
  line INTEGER NOT NULL
);

-- Docs: block‑level structure and code blocks
CREATE TABLE doc_blocks (
  block_id INTEGER PRIMARY KEY,
  file_id INTEGER NOT NULL,
  block_type TEXT NOT NULL,         -- heading, paragraph, code, list
  language TEXT,                    -- for code blocks
  content TEXT NOT NULL,
  start_line INTEGER NOT NULL,
  end_line INTEGER NOT NULL
);

-- Doc ↔ code reference index (explicit or inferred)
CREATE TABLE doc_refs (
  doc_ref_id INTEGER PRIMARY KEY,
  doc_file_id INTEGER NOT NULL,
  target TEXT NOT NULL,             -- symbol name, file path, or command string
  ref_type TEXT NOT NULL,           -- symbol, file, cli-flag, code-snippet
  confidence REAL NOT NULL
);

-- Rename maps used in refactors
CREATE TABLE rename_maps (
  map_id INTEGER PRIMARY KEY,
  scope TEXT NOT NULL,              -- go-ast, regex, yaml, docs
  old TEXT NOT NULL,
  new TEXT NOT NULL,
  rationale TEXT
);
```

### 1.3 Practical queries

**Find all API symbols that still use forbidden vocab:**

```sql
SELECT s.*
FROM symbols s
JOIN files f ON f.file_id = s.file_id
WHERE s.exported = 1
  AND (s.name LIKE '%Layer%' OR s.name LIKE '%Parameter%');
```

**Find doc blocks with stale terms but no matching code updates:**

```sql
SELECT f.path, d.start_line, d.content
FROM doc_blocks d
JOIN files f ON f.file_id = d.file_id
WHERE d.block_type = 'code'
  AND d.content LIKE '%AddFlags%';
```

**Find struct tags using old keys or values:**

```sql
SELECT f.path, t.struct_name, t.field_name, t.tag_key, t.tag_value
FROM struct_tags t
JOIN files f ON f.file_id = t.file_id
WHERE t.tag_key != 'glazed'
   OR t.tag_value LIKE '%parameter%';
```

### 1.4 Why this matters

During the refactor, we repeatedly built ad‑hoc inventories via `rg` and small scripts. A dedicated database would have:

- replaced one‑off scripts with reusable, testable queries;
- distinguished **API symbols** from **doc strings** and **test fixtures**;
- enabled reliable dependency analysis for “what breaks if I rename X”.

---

## 2) Tooling that would have helped

### 2.1 Core categories

1) **AST‑aware refactor engine** (Go AST rewrite)
   - Rename symbols across packages with import tracking.
   - Adjust method receivers, interface names, and method signatures.

2) **Doc refactor engine** (Markdown + code blocks)
   - Parse fenced code blocks, detect language, apply language‑aware refactors.
   - Track references to symbols and commands in prose.

3) **Config/fixture refactor engine** (YAML/JSON/TOML)
   - Parse configuration values and detect keys/values referencing symbols.
   - Support rewrites (e.g., `TargetParameter` -> `TargetField`).

4) **Refactor orchestration framework**
   - A pipeline that runs renames, then gofmt, then tests, then audits.

### 2.2 Useful self‑contained tools

- `refactor/rename-go`: AST rename with symbol resolution and import rewriting.
- `refactor/rename-docs`: Markdown parsing + fenced code transformation.
- `refactor/rename-yaml-json`: AST‑style parsing of data files to avoid text replacement errors.
- `refactor/audit`: run “forbidden terms” queries (ex: `Layer|Parameter`) by file category.
- `refactor/report`: generate a diff summary from git (renames + major symbol changes).

### 2.3 Useful refactor actions (API primitives)

Think of these as building blocks for a refactor framework:

- `RenameSymbol(old, new, scope=go, packages=[])`
- `RenameType(old, new)` (with method receiver updates)
- `RenameMethod(receiver, old, new)`
- `RenameTagKey(oldKey, newKey)`
- `RenameTagValue(oldValue, newValue)`
- `RenameStringLiteral(old, new, context=error|flag|doc)`
- `RewriteDocCodeBlocks(language, rewriteFn)`
- `RewriteConfigKeys(old, new, format=yaml|json)`
- `MoveFile(oldPath, newPath)`

### 2.4 A refactor framework (minimal architecture)

```text
+-------------------+   +--------------------+   +------------------+
| Scanner / Indexer |-->| Refactor Pipeline  |-->| Validator         |
+-------------------+   +--------------------+   +------------------+
          |                      |                         |
          v                      v                         v
  SQLite / Graph         AST + Doc + Config         gofmt + go test
  (symbols, refs,        rewrite passes              + audits
   strings, docs)
```

The key is **shared data**: the indexer populates the database, the pipeline queries it to decide what to rewrite, and the validator uses it to verify completeness.

---

## 3) What must be human‑supervised

Some refactor steps are fundamentally ambiguous and should be **reviewed or decided by humans**.

### 3.1 Naming decisions

- **Vocabulary convergence**: choosing “field” vs “parameter”, “section” vs “layer”.
- **Semantic boundaries**: e.g., `values.Section` vs `schema.Section` naming to avoid confusion.

### 3.2 Public API ergonomics

- Determining if a rename makes the API *clearer* or just *different*.
- Deciding whether to keep or remove wrappers when alias removal is complete.

### 3.3 Documentation semantics

- Code snippets in docs often encode meaning beyond symbol names.
- Even if a rename is correct syntactically, examples may become misleading.

### 3.4 Release and migration posture

- No‑compat vs compat, explicit deprecations, user communication.
- Whether to ship tooling (e.g., an AST migrator) to downstream users.

### 3.5 Invariants and domain rules

- Things like precedence order (defaults < config < env < flags) are domain semantics.
- A tool can rename, but a human must verify the semantics remain intact.

**Bottom line:** a fully automated refactor is not reliable for public API changes. The human‑supervised layers are *not optional*.

---

## 4) What a tool‑assisted refactor would look like (from scratch)

### 4.1 High‑level workflow

1) **Index**: Build a full DB of symbols, references, strings, and docs.
2) **Plan**: Codify renames and rules into a structured refactor plan.
3) **Execute**: Apply AST and structured rewrites in deterministic passes.
4) **Validate**: Compile + run targeted audits + regenerate reports.
5) **Review**: Human review of API naming and doc examples.
6) **Publish**: Update playbooks and postmortems with appendices.

### 4.2 Pseudocode: index + plan + execute

```pseudo
function refactor_project(root):
    db = build_index(root)

    plan = build_plan(db)
    # plan includes rename_map, file_renames, doc_rewrites, config_rewrites

    for step in plan.steps:
        apply(step)
        if step.requires_format:
            gofmt(step.affected_files)

    validate(root, db)
    emit_reports(db, plan)

function build_index(root):
    db = sqlite_open()
    scan_files(db, root)
    parse_go_symbols(db)
    parse_strings(db)
    parse_docs(db)
    parse_struct_tags(db)
    return db

function build_plan(db):
    rename_map = derive_rename_map(db)
    file_moves = derive_file_renames(db)
    doc_rules  = derive_doc_rewrites(db)
    config_rules = derive_config_rewrites(db)
    return Plan(steps=[...])
```

### 4.3 Pseudocode: validation and audits

```pseudo
function validate(root, db):
    run("gofmt -w ...")
    run("go test ./...")
    run("golangci-lint run ...")

    forbidden = ["Layer", "Parameter"]
    for term in forbidden:
        assert db.query("SELECT ... WHERE name LIKE %term%") == empty

    ensure_docs_match_code(db)
```

### 4.4 Refactor plan structure (YAML)

```yaml
refactor:
  name: "glazed-api-rename"
  renames:
    - scope: go
      old: ParameterLayer
      new: Section
      packages: ["github.com/go-go-golems/glazed/pkg/cmds/schema"]
    - scope: go
      old: ParsedParameters
      new: FieldValues
  files:
    - move: "cmd/examples/parameter-types" -> "cmd/examples/field-types"
  docs:
    - rewrite: "AddFlags" -> "AddFields"
      where: "code-blocks"  # not prose
    - rewrite: "credentials-param" -> "credentials-field"
      where: "docs/examples"
  config:
    - rewrite: "TargetParameter" -> "TargetField"
```

---

## 5) Engineering recommendations (next refactor)

### 5.1 Build the refactor database first

Create the SQLite database as the single source of truth for naming, file moves, and doc coupling. This reduces the reliance on ad‑hoc scripts and makes the refactor more deterministic.

### 5.2 Use structured rewrite passes

Perform renames in discrete phases:

1) Go AST renames
2) Structured config rewrites
3) Doc code block rewrites
4) Prose terminology cleanup

Each phase can be validated separately before the next begins.

### 5.3 Keep human oversight in the loop

Use tools for scale, but keep humans in the loop for:

- Naming decisions
- Semantic correctness
- Documentation clarity

---

## 6) Summary

A large API refactor cannot be successful with raw search‑replace. It needs:

- a **structured index** to reason about symbols, strings, and doc references;
- **AST‑aware rewrites** for code, structured rewrites for data files;
- **explicit human supervision** for naming and semantic correctness;
- and an orchestrated pipeline that makes the process reproducible.

This blueprint provides a concrete foundation for the next refactor project: a database schema, a tool taxonomy, and a deterministic workflow.
