## Workflow: docmgr + diary + git commits (single source of truth)

This document consolidates **docmgr instructions**, **diary instructions**, and **git commit instructions** into one consistent, unambiguous workflow.

It’s written to be copy/paste friendly and to minimize “where do I put this?” ambiguity.

---

## Goals

- **Make work traceable**: every meaningful investigation step has a diary entry and (when applicable) a linked commit hash.
- **Make tickets navigable**: every ticket has an index, tasks, changelog, and related files kept current.
- **Keep git history clean**: avoid committing noise and make small, reviewable commits.

---

## Terminology and conventions

- **Repo root**: wherever you run `git status` for the relevant module.
- **docmgr root**: the module folder containing `.ttmp.yaml` and `ttmp/` (ex: `.../glazed/`).
- **Ticket directory**: `ttmp/YYYY/MM/DD/<TICKET-ID>--<slug>/`
- **Doc types**: `analysis`, `design`, `reference`, `playbooks`, `log`, etc (docmgr-defined types).
- **Diary**: a living document under the ticket, usually `reference/01-diary.md`, updated frequently while you work.

**Absolute paths** are preferred in commands to avoid ambiguity.

---

## The canonical workflow (end-to-end)

### 1) Start / resume work (docmgr + git sanity)

From the relevant module root (example: `.../glazed/`):

```bash
git status --porcelain && docmgr status --summary-only
```

If this is new work:

1. Create a ticket (see next section).
2. Create at least:
   - **analysis doc** (to hold the “deep analysis”)
   - **diary** (to log your investigation steps)
3. Relate key files to the ticket index.

If this is existing work:

- open the ticket index and diary, skim latest entries, and continue.

---

## docmgr: tickets, docs, and relationships

### 2) Create a ticket (docmgr)

Run from the docmgr root (example: `.../glazed/`):

```bash
docmgr ticket create-ticket \
  --ticket <TICKET-ID> \
  --title "<human title>" \
  --topics topic1,topic2,topic3
```

Notes:

- Ticket ID format is project-specific; keep it consistent within the repo (examples: `001-FOO-BAR`, `PIN-1234`, etc).
- The ticket directory is created under `ttmp/YYYY/MM/DD/`.

### 3) Create docs in the ticket

At minimum, create:

- **Analysis doc**: where the “final answer” and technical reasoning live
- **Diary doc**: where step-by-step progress is logged

```bash
docmgr doc add --ticket <TICKET-ID> --doc-type analysis   --title "<analysis title>" && \
docmgr doc add --ticket <TICKET-ID> --doc-type reference --title "Diary"
```

To see available doc types and their required structure:

```bash
docmgr doc guidelines --list && docmgr doc guidelines --doc-type <type>
```

### 4) Keep the ticket index useful (RelateFiles)

Relate code/docs you touched or that are central to the issue:

```bash
docmgr doc relate --ticket <TICKET-ID> \
  --file-note "/abs/path/to/key-file.go:Why this file matters" \
  --file-note "/abs/path/to/another.md:Relevant background"
```

Good `--file-note` style:

- Include the **exact path** and a **one-line reason** (“where parsing happens”, “contains current bug”, “defines schema”, etc).

### 5) Update tasks + changelog as you work

Use `tasks.md` for “what remains”. Use `changelog.md` for “what happened / decisions”.

When you finish a meaningful step (investigation milestone, decision, implementation change):

```bash
docmgr changelog update --ticket <TICKET-ID> --entry "<what changed / what we decided>"
```

When you want a quick overview:

```bash
docmgr ticket list --ticket <TICKET-ID>
```

---

## Diary: how to log investigations (high frequency, high signal)

### 6) Diary purpose and structure

Your diary is not a blog; it’s an **audit trail** that answers:

- What did I look at?
- Why did I look at it?
- What did I learn / decide?
- What should I do next?
- What commit(s) correspond to changes?

Recommended format per step (copy/paste template):

```md
## Step N: <short title>

### What I did
- ...

### Why
- ...

### What I learned
- ...

### Open questions / next steps
- ...

### Commits (if any)
- <hash> - <message>
```

### 7) How often to write diary entries

Write an entry whenever any of these happens:

- You completed a search/reading pass across multiple files and now have a clearer model.
- You discovered a key constraint or unexpected behavior.
- You made a decision (even a tentative one).
- You finished a patch / commit.
- You changed direction (“we thought X, but it’s actually Y”).

Aim: **many small entries**, not one giant entry at the end.

### 8) Where to store “future research”

If you’re doing additional research that isn’t directly a ticket doc yet, store it under:

- `ttmp/YYYY-MM-DD/0X-XXX.md`

This keeps ad-hoc research centralized and discoverable.

---

## Git: clean, reviewable commits with no noise

### 9) Before every commit

```bash
git status --porcelain && \
git diff --stat && \
git diff --cached --stat
```

### 10) Stage changes (preferred: explicit paths)

```bash
git add /abs/path/to/file1.md /abs/path/to/file2.go
```

Or (careful) stage everything:

```bash
git add -A
```

### 11) Review what will be committed

```bash
git diff --cached --name-only && git diff --cached
```

### 12) Commit

```bash
git commit -m "<short summary of change>" && git rev-parse HEAD
```

Then paste the hash into the **Diary** under the current step.

### 13) Never commit (common noise)

Do not commit (unless the repo explicitly requires it):

- `node_modules/`
- `vendor/` (often)
- binaries (`*.exe`, `*.bin`, etc)
- `.env`, `.env.local`
- logs (`*.log`)
- build output (`dist/`, `build/`, `out/`)
- OS junk (`.DS_Store`, `Thumbs.db`)
- python caches (`*.pyc`, `__pycache__/`)
- IDE config (`.idea/`, `.vscode/`) unless intentionally shared
- coverage output

### 14) If you accidentally staged noise

Unstage one file:

```bash
git reset HEAD path/to/noise
```

Unstage everything:

```bash
git reset HEAD
```

### 15) If noise got committed

Remove from git but keep on disk:

```bash
git rm --cached path/to/noise
```

Ensure it’s ignored (append to `.gitignore`), then amend if it was the last commit:

```bash
echo "path/to/noise" >> .gitignore && git add .gitignore && git commit --amend --no-edit
```

### 16) Debugging “why is this file showing up?”

See ignored files:

```bash
git status --ignored
```

Check ignore rules for a path:

```bash
git check-ignore -v path/to/file
```

---

## Putting it together: a minimal, unambiguous “happy path”

From docmgr root (example: `.../glazed/`), start a new ticket:

```bash
git status --porcelain && docmgr status --summary-only && \
docmgr ticket create-ticket --ticket <TICKET-ID> --title "<title>" --topics topic1,topic2 && \
docmgr doc add --ticket <TICKET-ID> --doc-type analysis --title "<analysis title>" && \
docmgr doc add --ticket <TICKET-ID> --doc-type reference --title "Diary"
```

During investigation:

- Update diary every time you learn something important.
- Add `docmgr doc relate` file notes as soon as you identify “key files”.

When you’re ready to checkpoint changes:

```bash
git status --porcelain && git diff --stat && \
git add <paths...> && \
git diff --cached --name-only && \
git commit -m "<message>" && \
git rev-parse HEAD
```

Then:

- paste commit hash into diary
- add a ticket changelog entry summarizing the milestone

```bash
docmgr changelog update --ticket <TICKET-ID> --entry "<milestone summary (mention commit hash)>"
```


