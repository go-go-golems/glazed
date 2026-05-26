---
Title: Diary
Ticket: DOCSCTL-A14Y
Status: active
Topics:
    - docsctl
    - a14y
    - agent-readability
    - spa
    - go
    - seo
DocType: reference
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/server/wellknown.go
    - /home/manuel/workspaces/2026-05-25/docsctl-cicd-deploy/glazed/pkg/help/server/serve.go
ExternalSources: []
Summary: "Diary of a14y agent readability improvement work: from 42/100 to 76/100 (page mode)"
LastUpdated: 2026-05-25T22:05:00.000000000-04:00
WhatFor: ""
WhenToUse: ""
---

# Diary

## Goal

Improve the a14y (agent readability) score for docs.yolo.scapegoat.dev from 42/100 to 80+/100. The root cause is that the Go server's SPA fallback serves `index.html` for every non-API path, including well-known agent files like `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml`, and `/sitemap.md`.

## Step 1: Baseline audit (42/100 page mode)

The initial a14y audit scored 42/100 in page mode. 17 checks failed. The SPA returned an empty `<div id="root"></div>` for all URLs, including agent-discovery files. No metadata, no content, no structure.

### Prompt Context

**User prompt (verbatim):** (see session summary — part of the original DOCSCTL-SSR work)

**Assistant interpretation:** Run a14y audit, identify root causes, create design doc with 4-phase implementation plan.

**Inferred user intent:** Understand the a14y score, plan improvements, and create a tracking ticket.

### What I did
- Ran `npx a14y check http://localhost:8088 --mode page` — scored 42/100
- Identified root cause: SPA fallback serves index.html for agent files
- Created DOCSCTL-A14Y ticket with 47KB design doc (4 phases)
- Created 4 tasks, uploaded to remarkable

### Why
Without understanding the specific failing checks, we can't prioritize fixes. The design doc provides a complete implementation guide.

### What worked
- The audit clearly identified 17 specific failing checks
- Grouping failures by category (discoverability, metadata, content, markdown) revealed Phase 1 as highest-impact

### What didn't work
- (none)

### What I learned
- a14y is agent-readability, not WCAG accessibility — it measures how well AI agents can discover/parse/comprehend a site
- SPA architecture is fundamentally hostile to agent readability — the SPA shell has no content until JavaScript runs

### What was tricky to build
- N/A — this was an audit/analysis step, not implementation

### What warrants a second pair of eyes
- The 4-phase prioritization: is Phase 1 (well-known files) really the highest impact? Yes — it fixes the worst failures (serving HTML for .txt/.xml files)

### What should be done in the future
- Re-audit after each phase to measure improvement

### Code review instructions
- Read `ttmp/2026/05/25/DOCSCTL-A14Y.../design-doc/01-a14y-agent-readability-improvement-design-and-implementation-guide.md`

---

## Step 2: SSR sidecar adds metadata (42 → 55/100 page mode, SSR metadata)

The SSR sidecar implementation (DOCSCTL-SSR) added `<meta>` tags, Open Graph tags, canonical links, and `<html lang="en">` to the server-rendered HTML. This was a side effect of the SSR work, not a dedicated a14y fix.

### Prompt Context

**User prompt (verbatim):** (see DOCSCTL-SSR work)

**Commit (code):** dab2fc9 — "DOCSCTL-SSR: add devctl plugin for docs server + SSR sidecar, fix SSR runtime issues"

### What I did
- The SSR server.mjs injects meta description, og:title, og:description, and canonical link into the HTML
- Re-ran a14y: page-mode score went from 42 to 55
- New passing checks: html.canonical-link, html.meta-description, html.og-title, html.og-description, html.lang-attribute

### Why
SSR produces meaningful HTML metadata that the SPA shell lacked. This was a free improvement from the SSR work.

### What worked
- The SSR server correctly generates per-page metadata (title, description vary by section)
- The canonical link uses the production base URL

### What didn't work
- The SSR HTML still has 0 headings and low text ratio — React renders the skeleton, not full content

### What I learned
- SSR metadata alone bumps the page score by 13 points (42 → 55)
- The biggest remaining gaps are: no headings, low text ratio, no JSON-LD, no markdown negotiation

### What was tricky to build
- N/A — this was an observation of the SSR side effect

### What warrants a second pair of eyes
- The SSR metadata is correct but limited — og:title shows "Glazed Help Browser" for all pages, not the section-specific title

### What should be done in the future
- Pre-populate RTK Query cache so SSR renders actual content (headings, text)
- Add JSON-LD structured data

### Code review instructions
- Check `web/server.mjs` — the metadata injection in the catch-all handler

---

## Step 3: Well-known file handler (55 → 76/100 page mode)

Created `pkg/help/server/wellknown.go` with a `WellKnownHandler` that generates `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml`, `/sitemap.md`, and `/index.md` dynamically from the help database. Wired it into the routing chain in `serve.go` before the SPA fallback.

### Prompt Context

**User prompt (verbatim):** "go ahead" (continuing from the session summary's next steps)

**Assistant interpretation:** Implement DOCSCTL-A14Y Phase 1: server-side well-known file serving.

**Inferred user intent:** Fix the most impactful a14y failures by serving real content for agent files instead of the SPA shell.

**Commit (code):** (pending)

### What I did
- Created `pkg/help/server/wellknown.go` with:
  - `WellKnownHandler` struct with `CanHandle()` and `ServeHTTP()` methods
  - `generateLlmsTxt()` — packages + sections with links
  - `generateRobotsTxt()` — allows all AI bots, references sitemap
  - `generateAgentsMd()` — install, usage, API, packages sections
  - `generateSitemapXML()` — urlset with all section URLs + lastmod
  - `generateSitemapMd()` — headings by package, links to sections
  - `generateIndexMd()` — YAML frontmatter + package list + sitemap section
  - Helper functions: `deriveBaseURL()`, `groupPackages()`, `sortedKeys()`
- Updated `serve.go`: added well-known handler interception before SPA fallback and SSR proxy
- Built and tested: all handlers return correct Content-Type and real content
- Re-ran a14y: page-mode score went from 55 to 76/100

### Why
The SPA fallback served index.html for agent files, which is invalid content. Agents need proper text/plain for llms.txt, text/markdown for AGENTS.md, and application/xml for sitemap.xml. The well-known handler intercepts these paths and generates content from the live database.

### What worked
- All 6 well-known files served with correct Content-Type
- llms-txt.content-type: now passing (was failing with text/html)
- sitemap-xml.valid: now passing (was HTML, now valid urlset)
- sitemap-md.has-structure: now passing (3 headings, 98 links)
- agents-md.has-min-sections: now passing (found: installation, usage)
- sitemap-xml.has-lastmod: new passing check (129 entries with lastmod)
- markdown.frontmatter: now passing (index.md has title, description, doc_version, last_updated)
- markdown.sitemap-section: now passing (index.md has ## Sitemap heading)
- Page-mode score: 55 → 76 (+21 points)

### What didn't work
- llms-txt.md-extensions still fails — all links in llms.txt use the path format (e.g. `/glazed/_/sections/foo`) without `.md` extension
- discovery.indexed shows "no site index available" in page mode — the root page IS in the sitemap but the page-mode check seems to not find it

### What I learned
- The well-known handler is the single highest-impact a14y fix: +21 points from one file
- The routing order matters: well-known handler must be checked BEFORE the SSR proxy, because the SSR proxy would also handle these paths (returning HTML)
- `deriveBaseURL()` needs to handle both local dev (http) and production (https with X-Forwarded-Proto)

### What was tricky to build
- The routing insertion point: well-known files must be intercepted before BOTH the SPA fallback AND the SSR proxy. If the SSR proxy catches `/llms.txt`, it would render the React app for that path (which is wrong).
- The `llms-txt.md-extensions` check: it expects links to end in `.md` or `.mdx`. Our URLs use the path scheme `/{package}/{version}/sections/{slug}`. Adding `.md` suffix to these would break the web routing. This check may require a separate approach.

### What warrants a second pair of eyes
- The well-known handler generates content on every request. For production, caching the output would reduce database load. The 5-minute Cache-Control header helps but doesn't prevent re-generation.
- The `llms-txt.md-extensions` failure: should we add `.md` URL variants that redirect to the section page? Or is this a check that doesn't apply to our URL scheme?

### What should be done in the future
- Add markdown content negotiation (Accept: text/markdown → return section content as markdown)
- Add `<link rel="alternate" type="text/markdown">` header
- Add JSON-LD structured data
- Improve SSR to render headings and more text content
- Add Link header with canonical URL for markdown mirror

### Code review instructions
- Start with `pkg/help/server/wellknown.go` — all 6 generators
- Check routing insertion in `pkg/help/server/serve.go` (search "wellKnownHandler")
- Test: `curl -I http://localhost:8088/llms.txt && curl -I http://localhost:8088/sitemap.xml && curl -I http://localhost:8088/AGENTS.md`

---

## Step 4: Markdown content negotiation and alternate links (IN PROGRESS)

### What remains (8 failing checks)

1. `llms-txt.md-extensions` — links don't use .md/.mdx extensions
2. `html.json-ld` — no JSON-LD structured data
3. `html.headings` — 0 headings in SSR HTML
4. `html.text-ratio` — 12.9% (need >15%)
5. `html.glossary-link` — no glossary/terminology link
6. `markdown.alternate-link` — no `<link rel="alternate" type="text/markdown">`
7. `markdown.canonical-header` — no Link header on markdown mirror
8. `markdown.content-negotiation` — no Accept-based negotiation
