---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: ../../../../../../../pinocchio/pkg/spa/spa.go
      Note: SPA fallback handler — no change needed but important to understand
    - Path: pkg/help/server/handlers.go
      Note: API handlers — reference for new well-known handlers
    - Path: pkg/help/server/serve.go
      Note: NewServeHandler routing chain — where well-known handler must be inserted
    - Path: pkg/help/server/types.go
      Note: API response types — reference for sitemap generation
    - Path: pkg/help/store/store.go
      Note: Store interface — ListPackages needed for sitemap generation
    - Path: web/index.html
      Note: SPA shell — needs meta tags
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---







# a14y Agent Readability Improvement: Design and Implementation Guide

**Target:** docs.yolo.scapegoat.dev (production) and local dev (`localhost:5173`)  
**Current score:** 42/100 (scorecard v0.2.0)  
**Target score:** 80+/100  
**Failing checks:** 17 (site-wide + page-level)  
**Passing checks:** 14

---

## 1. Executive Summary

The docs.yolo.scapegoat.dev documentation browser scores **42/100** on the a14y agent-readability scorecard. This is a failing grade for AI agent discoverability. The root cause is architectural: the Go server's SPA fallback handler serves `index.html` for **every** non-`/api` path, including well-known agent files like `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml`, and `/sitemap.md`. Instead of returning the actual content, agents get the React SPA shell — an empty `<div id="root"></div>` that only renders in a browser.

This guide provides a complete, intern-ready design and implementation plan to fix all 17 failing a14y checks. The work breaks into four phases:

1. **Server-side well-known file serving** — Teach the Go server to intercept requests for `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml`, and `/sitemap.md` before they hit the SPA fallback, and generate the correct content dynamically from the help database.
2. **HTML metadata enrichment** — Add `<meta>` tags, Open Graph tags, canonical links, and JSON-LD structured data to the SPA's `index.html` so that crawlers can understand the page even before JavaScript hydrates.
3. **Markdown content negotiation** — Add a Go handler that returns the raw Markdown content of a section when an agent sends `Accept: text/markdown`, plus `<link rel="alternate">` headers in the HTML response.
4. **AGENTS.md and llms.txt content quality** — Populate these files with real, structured content that documents how to install, configure, and use the docs browser as an agent.

Each phase includes the exact files to modify, the API contracts, pseudocode, diagrams, and testing instructions.

---

## 2. System Architecture Overview

Before diving into the fixes, let me explain how the docs browser works end-to-end. If you're a new intern picking up this ticket, read this section carefully — it will give you the mental model you need to understand every change.

### 2.1 The Two-Process Architecture

The docs browser is a **Go backend** + **React/TypeScript SPA frontend**, deployed as a single binary with embedded frontend assets.

```
┌─────────────────────────────────────────────────────┐
│                    Browser                           │
│                                                     │
│  ┌──────────────┐     ┌─────────────────────────┐ │
│  │  React SPA    │────▶│  Go HTTP Server (:8088)  │ │
│  │  (BrowserRouter)│   │                           │ │
│  │               │◀────│  /api/*  → JSON API       │ │
│  │  Renders docs │     │  /*      → SPA fallback  │ │
│  └──────────────┘     └─────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

The Go server serves two things:

1. **`/api/*`** — REST API endpoints that return JSON (section lists, section content, package metadata).
2. **`/*`** (everything else) — The SPA fallback: it serves `index.html` for any path that doesn't match a static asset. This is what makes client-side routing work — navigating to `/glazed/_/sections/build-first-command` returns `index.html`, and React Router handles the URL client-side.

**The critical problem:** The SPA fallback catches *everything* that isn't `/api/*` or a static asset file. This means `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml` — they all get `index.html` instead of their actual content.

### 2.2 The Request Routing Chain

When a request hits the Go server, here's the exact decision tree:

```
Incoming request
  │
  ├─ Path starts with /api/ ?
  │    YES → API handler (JSON responses)
  │    NO  ↓
  │
  ├─ Path matches a static asset file?
  │    (e.g. /assets/index-C9ROnpts.js, /assets/index-C3NMyfra.css)
  │    YES → Serve the static file (http.FileServer)
  │    NO  ↓
  │
  └─ SPA fallback: serve index.html
       (Content-Type: text/html; charset=utf-8)
```

**What needs to change:** We must insert a new decision *before* the SPA fallback that intercepts well-known agent files:

```
Incoming request
  │
  ├─ Path starts with /api/ ?
  │    YES → API handler
  │    NO  ↓
  │
  ├─ [NEW] Path is a well-known agent file?
  │    (/llms.txt, /robots.txt, /AGENTS.md, /sitemap.xml, /sitemap.md)
  │    YES → Generate and serve the file dynamically
  │    NO  ↓
  │
  ├─ Path matches a static asset file?
  │    YES → Serve the static file
  │    NO  ↓
  │
  └─ SPA fallback: serve index.html
```

### 2.3 Key Files and Their Roles

Here are the files you'll be working with, organized by layer:

#### Go Backend (the server)

| File | Role | What you'll change |
|------|------|--------------------|
| `pkg/help/server/serve.go` | Main server setup; `NewServeHandler` builds the routing chain | Add well-known file interception before SPA fallback |
| `pkg/help/server/handlers.go` | API handlers: `/api/health`, `/api/packages`, `/api/sections`, etc. | Add new handlers: `handleLlmsTxt`, `handleRobotsTxt`, `handleSitemapXML`, `handleSitemapMd`, `handleAgentsMd`, `handleMarkdownMirror` |
| `pkg/help/server/types.go` | API response types (`PackageSummary`, `SectionSummary`, etc.) | Add new types for sitemap/agents-md generation |
| `pkg/help/server/middleware.go` | CORS middleware | No change needed |
| `pinocchio/pkg/spa/spa.go` | SPA fallback handler (serves `index.html` for unknown paths) | No change — the interception happens before the SPA handler |

#### React/TypeScript Frontend (the SPA)

| File | Role | What you'll change |
|------|------|--------------------|
| `web/index.html` | The HTML shell that React mounts into | Add `<meta>` tags, OG tags, canonical link, JSON-LD |
| `web/public/site-config.js` | Runtime configuration (API base URL, etc.) | No change for this ticket |
| `web/src/App.tsx` | Main app component with React Router | No change for this ticket |

#### Data Sources

| Source | Description |
|--------|-------------|
| SQLite help database | Contains all sections with their Markdown content, metadata, headings, topics |
| `helploader.SQLiteDirLoader` | Discovers databases at `X/Y/X.db` → package X, version Y |
| `store.Store` | In-memory query layer over SQLite; powers all API handlers |

### 2.4 The API Surface

The existing REST API that the SPA calls:

| Endpoint | Method | Description | Response shape |
|----------|--------|-------------|----------------|
| `/api/health` | GET | Health check | `{ ok: true, sections: 128 }` |
| `/api/packages` | GET | List all packages with versions | `{ packages: [...], defaultPackage: "glazed" }` |
| `/api/sections` | GET | List sections (filterable) | `{ sections: [...], total: 74, limit: -1, offset: 0 }` |
| `/api/sections/{slug}` | GET | Get full section content | `{ id, slug, title, content, ... }` |

Query parameters for `/api/sections`:
- `package` — filter by package name
- `version` — filter by package version
- `type` — filter by section type (GeneralTopic, Example, Application, Tutorial)
- `topic` — filter by topic tag
- `q` — full-text search

### 2.5 How Section Data Is Stored

Each section in the SQLite database has these fields:

```go
type Section struct {
    ID             int64
    PackageName    string    // "glazed" or "pinocchio"
    PackageVersion string    // "v1.2.15" or "" (unversioned)
    Slug           string    // "build-first-command"
    SectionType    SectionType  // GeneralTopic | Example | Application | Tutorial
    Title          string    // "Build Your First Glazed Command"
    Short          string    // First paragraph or manual summary
    Topics         []string  // ["help-system", "api"]
    Flags          []string  // CLI flags referenced
    Commands       []string  // CLI commands referenced
    IsTopLevel     bool      // Whether this is a top-level section
    Content        string    // Full Markdown body
}
```

The `Content` field contains the raw Markdown that gets rendered by the React frontend. This is also what we'll serve to agents via the Markdown mirror.

---

## 3. Problem Statement: The 17 Failing Checks

The a14y audit found 17 failing checks, all caused by the SPA fallback serving `index.html` for agent-discoverable files. Here they are grouped by category:

### 3.1 Discoverability (3 failures)

These check whether an AI agent can *find* the site's content through standard discovery mechanisms.

| Check ID | What it checks | Current behavior | Required behavior |
|----------|---------------|-----------------|-------------------|
| `agents-md.has-min-sections` | AGENTS.md documents at least 2 of install/config/usage | AGENTS.md is HTML (SPA shell), has no usable sections | AGENTS.md must be real Markdown with install, config, usage sections |
| `sitemap-md.has-structure` | sitemap.md has headings and links | sitemap.md is HTML, no structure | sitemap.md must be Markdown with `##` headings and `[links](url)` |
| `sitemap-xml.valid` | sitemap.xml parses as urlset/sitemapindex | sitemap.xml is HTML | sitemap.xml must be valid XML with `<url><loc>` entries |

### 3.2 Page Discovery (1 failure)

| Check ID | What it checks | Current behavior | Required behavior |
|----------|---------------|-----------------|-------------------|
| `discovery.indexed` | Page is announced by sitemap, llms.txt, or sitemap.md | The root page is orphaned — not listed in any index | Add root URL to sitemap.xml and sitemap.md |

### 3.3 HTML Metadata (4 failures)

These check whether the HTML `<head>` contains machine-readable metadata that helps agents understand the page before JavaScript renders.

| Check ID | What it checks | Current behavior |
|----------|---------------|-----------------|
| `html.canonical-link` | Has `<link rel="canonical">` | Missing |
| `html.meta-description` | Has `<meta name="description">` (≥50 chars) | Missing |
| `html.og-title` | Has `<meta property="og:title">` | Missing |
| `html.og-description` | Has `<meta property="og:description">` | Missing |

### 3.4 Content Structure (2 failures)

| Check ID | What it checks | Current behavior |
|----------|---------------|-----------------|
| `html.headings` | Has at least 3 section headings (`<h2>`/`<h3>`) | 0 headings in the initial HTML (React renders them client-side) |
| `html.text-ratio` | Text-to-HTML ratio above 15% | 0% (only `<div id="root"></div>`) |
| `html.glossary-link` | Links to a glossary/terminology page | No glossary link |

### 3.5 Structured Data (1 failure)

| Check ID | What it checks | Current behavior |
|----------|---------------|-----------------|
| `html.json-ld` | Has parseable JSON-LD block | Missing |

### 3.6 Markdown Mirror (5 failures)

These check whether the site provides Markdown alternatives for agents that prefer plain text over HTML.

| Check ID | What it checks | Current behavior |
|----------|---------------|-----------------|
| `markdown.alternate-link` | HTML declares `<link rel="alternate" type="text/markdown">` | Missing |
| `markdown.content-negotiation` | Server returns Markdown for `Accept: text/markdown` | Returns HTML |
| `markdown.canonical-header` | Markdown mirror sends `Link: <url>; rel="canonical"` header | Missing |
| `markdown.frontmatter` | Markdown mirror has required frontmatter (title, description, etc.) | Missing |
| `markdown.sitemap-section` | Markdown mirror includes `## Sitemap` heading | Missing |

### 3.7 Warnings (2 — not failures but should fix)

| Check ID | What it checks | Current behavior |
|----------|---------------|-----------------|
| `llms-txt.content-type` | llms.txt served as `text/plain` | Served as `text/html` |
| `llms-txt.md-extensions` | llms.txt contains links to evaluate | No links (content is HTML, not real text) |

---

## 4. Proposed Architecture

### 4.1 Phase 1: Server-Side Well-Known File Serving

**Goal:** Intercept requests for `/llms.txt`, `/robots.txt`, `/AGENTS.md`, `/sitemap.xml`, `/sitemap.md` before they reach the SPA fallback, and generate the correct content dynamically from the help database.

#### 4.1.1 Routing Change

The change goes in `pkg/help/server/serve.go`, in the `NewServeHandler` function. Currently:

```go
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    // ... auto-assign default package ...
    apiHandler := NewHandler(deps)
    if spaHandler == nil {
        return apiHandler
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cleanPath := stdpath.Clean("/" + r.URL.Path)
        if cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/") {
            apiHandler.ServeHTTP(w, r)
            return
        }
        spaHandler.ServeHTTP(w, r)
    })
}
```

After the change:

```go
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    // ... auto-assign default package ...
    apiHandler := NewHandler(deps)
    wellKnownHandler := NewWellKnownHandler(deps)
    if spaHandler == nil {
        return apiHandler
    }
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cleanPath := stdpath.Clean("/" + r.URL.Path)
        if cleanPath == "/api" || strings.HasPrefix(cleanPath, "/api/") {
            apiHandler.ServeHTTP(w, r)
            return
        }
        // NEW: intercept well-known agent files before SPA fallback
        if wellKnownHandler.CanHandle(cleanPath) {
            wellKnownHandler.ServeHTTP(w, r)
            return
        }
        spaHandler.ServeHTTP(w, r)
    })
}
```

#### 4.1.2 Well-Known Handler

Create a new file `pkg/help/server/wellknown.go`:

```go
package server

import (
    "fmt"
    "net/http"
    "strings"
    "time"
)

// wellKnownPaths lists the paths that the well-known handler intercepts
// before the SPA fallback. These must be served with correct Content-Type
// and content, not as index.html.
var wellKnownPaths = map[string]string{
    "/llms.txt":    "text/plain; charset=utf-8",
    "/robots.txt":  "text/plain; charset=utf-8",
    "/AGENTS.md":   "text/markdown; charset=utf-8",
    "/sitemap.xml": "application/xml; charset=utf-8",
    "/sitemap.md":  "text/markdown; charset=utf-8",
}

// WellKnownHandler generates well-known agent files dynamically from the
// help database. It is used by NewServeHandler to intercept these paths
// before the SPA fallback.
type WellKnownHandler struct {
    deps HandlerDeps
}

func NewWellKnownHandler(deps HandlerDeps) *WellKnownHandler {
    return &WellKnownHandler{deps: deps}
}

func (h *WellKnownHandler) CanHandle(path string) bool {
    _, ok := wellKnownPaths[path]
    return ok
}

func (h *WellKnownHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    cleanPath := stdpath.Clean("/" + r.URL.Path)
    contentType, ok := wellKnownPaths[cleanPath]
    if !ok {
        // Not a well-known path; let the SPA fallback handle it
        http.NotFound(w, r)
        return
    }
    
    // Cache for 5 minutes — these files are generated from the database
    // and don't change frequently.
    w.Header().Set("Content-Type", contentType)
    w.Header().Set("Cache-Control", "public, max-age=300")
    
    var body string
    switch cleanPath {
    case "/llms.txt":
        body = h.generateLlmsTxt(r)
    case "/robots.txt":
        body = h.generateRobotsTxt()
    case "/AGENTS.md":
        body = h.generateAgentsMd(r)
    case "/sitemap.xml":
        body = h.generateSitemapXML(r)
    case "/sitemap.md":
        body = h.generateSitemapMd(r)
    }
    
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(body))
}
```

#### 4.1.3 Content Generation Functions

Each well-known file needs a generator. Here are the specifications:

##### `/llms.txt` (resolves `llms-txt.content-type`, improves `llms-txt.md-extensions`)

The [llms.txt](https://llmstxt.org/) convention is a Markdown-formatted plain-text file at the root of a website that gives LLMs a quick overview of the site's content. Format:

```
# [Site Name]

> [One-line description of the site]

[Optional: more paragraphs about the site]

## Sections

- [Section title](url): Short description
- [Another title](url): Another description
```

Pseudocode for `generateLlmsTxt`:

```
func generateLlmsTxt(r *http.Request) string:
    baseURL := deriveBaseURL(r)  // e.g. "https://docs.yolo.scapegoat.dev"
    
    // Fetch all packages
    packages := deps.Store.ListPackages(ctx)
    
    // Build the header
    output = "# Glazed Help Browser\n\n"
    output += "> Documentation browser for Glazed CLI framework and related tools.\n\n"
    output += "This site provides structured help documentation for Go-Go-Golems CLI tools. "
    output += "Browse by package and version, or search across all sections.\n\n"
    
    // Build the sections list
    output += "## Packages\n\n"
    for each package in packages:
        output += "- [{displayName}]({baseURL}/{name}/_): {sectionCount} sections\n"
        for each version in package.versions:
            sections := Store.Find(InPackageVersion(name, version), IsTopLevel)
            output += "  - [{name} {version}]({baseURL}/{name}/{version}): {len(sections)} sections\n"
    
    // Build the full section index
    output += "\n## Sections\n\n"
    for each package in packages:
        sections := Store.Find(InPackageVersion(name, latestVersion))
        for each section in sections where section.isTopLevel:
            url := baseURL + "/" + name + "/" + version + "/sections/" + section.slug
            output += "- [" + section.title + "](" + url + "): " + section.short + "\n"
    
    return output
```

Key considerations:
- The base URL must be derived from the request (use `r.Host` + `r.TLS` check, or a configurable base URL). For local dev, it will be `http://localhost:8088`; for production, `https://docs.yolo.scapegoat.dev`.
- Links in llms.txt should use the semantic URL format `/{package}/{version}/sections/{slug}`.
- The `short` field of each section is the description; if empty, omit the colon.

##### `/robots.txt` (already passing, but content is HTML — needs real content)

```
User-agent: *
Allow: /

User-agent: GPTBot
Allow: /

User-agent: ChatGPT-User
Allow: /

User-agent: CCBot
Allow: /

User-agent: Google-Extended
Allow: /

Sitemap: https://docs.yolo.scapegoat.dev/sitemap.xml
```

Pseudocode:

```
func generateRobotsTxt() string:
    return fixed string above
```

This is a static response since robots.txt doesn't change based on the database content. The `Sitemap:` line should use the production URL. For local dev, we can use `http://localhost:8088`.

##### `/AGENTS.md` (resolves `agents-md.has-min-sections`)

The [AGENTS.md](https://agents.md/) convention is a Markdown file at the root that tells AI agents how to interact with the site. It must document at least 2 of: install, configure, usage.

```markdown
# Glazed Help Browser — AGENTS.md

## Install

The Glazed Help Browser runs as a Go binary with an embedded React SPA:

```bash
go install github.com/go-go-golems/glazed/cmd/glaze@latest
glaze serve --from-sqlite-dir /path/to/help-dbs --address :8088
```

Or via Docker (if applicable).

## Configure

The server accepts these flags:
- `--address` — TCP address (default `:8088`)
- `--from-sqlite-dir` — Recursively load SQLite help databases
- `--from-json` — Load JSON help exports
- `--from-glazed-cmd` — Load help from other Glazed binaries
- `--reload-interval` — Periodically reload sources (e.g. `30s`)
- `--with-embedded` — Include built-in docs alongside external sources

The frontend reads `site-config.js` for the `apiBaseUrl` runtime setting.

## Usage

### API Endpoints

| Endpoint | Method | Parameters | Description |
|----------|--------|------------|-------------|
| `/api/health` | GET | — | Health check |
| `/api/packages` | GET | — | List packages with versions |
| `/api/sections` | GET | `package`, `version`, `type`, `topic`, `q`, `limit`, `offset` | List/filter sections |
| `/api/sections/{slug}` | GET | `package`, `version` (query params) | Get section content |

### URL Scheme

The SPA uses semantic URLs: `/{package}/{version}/sections/{slug}`
- Example: `/glazed/_/sections/build-first-command`
- Unversioned packages use `_` as the version segment
- Direct navigation to any URL works (SPA fallback)

### Content Types

- HTML: request normally (default)
- Markdown: send `Accept: text/markdown` header
- JSON: use `/api/*` endpoints

## Sitemap

See [sitemap.md](/sitemap.md) for a full list of pages.
```

Pseudocode:

```
func generateAgentsMd(r *http.Request) string:
    baseURL := deriveBaseURL(r)
    
    output = "# Glazed Help Browser — AGENTS.md\n\n"
    output += "## Install\n\n"
    output += "```bash\n"
    output += "go install github.com/go-go-golems/glazed/cmd/glaze@latest\n"
    output += "glaze serve --from-sqlite-dir /path/to/help-dbs --address :8088\n"
    output += "```\n\n"
    output += "## Configure\n\n"
    output += "The server accepts these flags:\n"
    // ... list all flags from ServeSettings ...
    output += "## Usage\n\n"
    output += "### API Endpoints\n\n"
    // ... generate table from the actual route definitions ...
    output += "### URL Scheme\n\n"
    // ... explain the semantic URL format ...
    output += "## Sitemap\n\n"
    output += "See [sitemap.md](/sitemap.md) for a full list of pages.\n"
    
    return output
```

##### `/sitemap.xml` (resolves `sitemap-xml.valid`, `discovery.indexed`)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://docs.yolo.scapegoat.dev/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://docs.yolo.scapegoat.dev/glazed/_</loc>
    <changefreq>daily</changefreq>
    <priority>0.9</priority>
  </url>
  <url>
    <loc>https://docs.yolo.scapegoat.dev/glazed/_/sections/build-first-command</loc>
    <changefreq>weekly</changefreq>
    <priority>0.7</priority>
  </url>
  <!-- ... one entry per section ... -->
</urlset>
```

Pseudocode:

```
func generateSitemapXML(r *http.Request) string:
    baseURL := deriveBaseURL(r)
    packages := Store.ListPackages(ctx)
    
    xml  = '<?xml version="1.0" encoding="UTF-8"?>\n'
    xml += '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n'
    
    // Root URL
    xml += '  <url><loc>' + baseURL + '/</loc>'
    xml += '<changefreq>daily</changefreq><priority>1.0</priority></url>\n'
    
    for each package in packages:
        // Package index page
        version := package.latestVersion || "_"
        xml += '  <url><loc>' + baseURL + '/' + name + '/' + version + '</loc>'
        xml += '<changefreq>daily</changefreq><priority>0.9</priority></url>\n'
        
        // Each section
        sections := Store.Find(InPackageVersion(name, version))
        for each section:
            url := baseURL + '/' + name + '/' + version + '/sections/' + section.slug
            xml += '  <url><loc>' + url + '</loc>'
            xml += '<changefreq>weekly</changefreq><priority>0.7</priority></url>\n'
    
    xml += '</urlset>\n'
    return xml
```

**Important:** XML requires proper escaping of special characters (`&`, `<`, `>`, `"`, `'`). Use `xml.EscapeText` or manually replace `&` → `&amp;`, `<` → `&lt;`, etc. in URLs (though URLs typically don't contain these characters).

##### `/sitemap.md` (resolves `sitemap-md.has-structure`)

```markdown
# Sitemap — Glazed Help Browser

## Packages

- [Glazed](/glazed/_): 74 sections
- [Pinocchio](/pinocchio/v1.2.15): 54 sections

## Glazed

- [Adding New Field Types to Glazed](/glazed/_/sections/adding-new-field-types)
- [Build Your First Glazed Command](/glazed/_/sections/build-first-command)
- ...

## Pinocchio (v1.2.15)

- [Pinocchio Overview](/pinocchio/v1.2.15/sections/overview)
- ...
```

Pseudocode:

```
func generateSitemapMd(r *http.Request) string:
    baseURL := deriveBaseURL(r)
    packages := Store.ListPackages(ctx)
    
    output = "# Sitemap — Glazed Help Browser\n\n"
    output += "## Packages\n\n"
    for each package in packages:
        version := package.latestVersion || "_"
        output += "- [" + package.displayName + "](" + baseURL + "/" + name + "/" + version + "): "
        output += strconv.Itoa(package.sectionCount) + " sections\n"
    
    for each package in packages:
        version := package.latestVersion || "_"
        sections := Store.Find(InPackageVersion(name, version))
        output += "\n## " + package.displayName
        if version != "_":
            output += " (" + version + ")"
        output += "\n\n"
        for each section:
            url := baseURL + "/" + name + "/" + version + "/sections/" + section.slug
            output += "- [" + section.title + "](" + url + ")\n"
    
    return output
```

#### 4.1.4 Base URL Derivation

The well-known files need to know the site's base URL for generating absolute links. Two approaches:

**Option A (recommended):** Add a `--base-url` flag to the serve command. Default to `http://localhost:8088`. In production, set `--base-url https://docs.yolo.scapegoat.dev`.

**Option B:** Derive from the request's `Host` header. This works for single-host deployments but breaks behind reverse proxies (the Host header might be `localhost` even though the public URL is `https://docs.yolo.scapegoat.dev`).

I recommend **Option A** because:
- It's explicit and predictable
- It works behind reverse proxies (Caddy, nginx, Cloud Run)
- It's a single flag in the deployment command

Implementation:

```go
// In serve.go, add to ServeSettings:
type ServeSettings struct {
    // ... existing fields ...
    BaseURL string `glazed:"base-url"` // NEW
}

// In NewServeHandler, pass BaseURL to WellKnownHandler:
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler, opts ...ServeOption) http.Handler {
    // ...
}
```

If `--base-url` is not set, derive from the request:

```go
func deriveBaseURL(r *http.Request) string {
    // Check X-Forwarded-Host / X-Forwarded-Proto for reverse proxy scenarios
    proto := "http"
    host := r.Host
    if r.TLS != nil {
        proto = "https"
    }
    if fwd := r.Header.Get("X-Forwarded-Proto"); fwd != "" {
        proto = fwd
    }
    if fwd := r.Header.Get("X-Forwarded-Host"); fwd != "" {
        host = fwd
    }
    return proto + "://" + host
}
```

---

### 4.2 Phase 2: HTML Metadata Enrichment

**Goal:** Add machine-readable metadata to `index.html` so that crawlers and agents can understand the page even before JavaScript hydrates.

#### 4.2.1 Changes to `web/index.html`

The current `index.html` is minimal:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Glazed Help Browser</title>
  </head>
  <body>
    <div id="root"></div>
    <!-- legacy hash redirect script -->
    <script src="./site-config.js"></script>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

After the change:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>Glazed Help Browser</title>
    
    <!-- SEO & Agent Metadata -->
    <meta name="description" content="Documentation browser for the Glazed CLI framework and Go-Go-Golems tools. Browse structured help sections, search by topic, and navigate by package and version." />
    
    <!-- Open Graph -->
    <meta property="og:title" content="Glazed Help Browser" />
    <meta property="og:description" content="Documentation browser for the Glazed CLI framework and Go-Go-Golems tools. Browse structured help sections, search by topic, and navigate by package and version." />
    <meta property="og:type" content="website" />
    <meta property="og:url" content="https://docs.yolo.scapegoat.dev/" />
    
    <!-- Canonical -->
    <link rel="canonical" href="https://docs.yolo.scapegoat.dev/" />
    
    <!-- Markdown alternate (resolves markdown.alternate-link) -->
    <link rel="alternate" type="text/markdown" href="/llms.txt" />
    
    <!-- JSON-LD Structured Data -->
    <script type="application/ld+json">
    {
      "@context": "https://schema.org",
      "@type": "WebApplication",
      "name": "Glazed Help Browser",
      "description": "Documentation browser for the Glazed CLI framework and Go-Go-Golems tools",
      "url": "https://docs.yolo.scapegoat.dev/",
      "applicationCategory": "DeveloperApplication",
      "operatingSystem": "Any"
    }
    </script>
    
    <!-- Pre-rendered headings for agent readability (resolves html.headings) -->
    <noscript>
      <h1>Glazed Help Browser</h1>
      <h2>Documentation</h2>
      <h2>Packages</h2>
      <h2>Sections</h2>
    </noscript>
  </head>
  <body>
    <div id="root"></div>
    <!-- legacy hash redirect script -->
    <script src="./site-config.js"></script>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

**Why `<noscript>` for headings?** The `html.headings` check looks for heading elements in the HTML source. Since React renders headings client-side, the initial HTML has zero headings. We can't add visible headings that React would then duplicate. The `<noscript>` approach provides headings for agents and crawlers that don't execute JavaScript, while being invisible to browsers that do. This is a pragmatic compromise — a fully server-rendered solution would be better but requires a much larger architectural change (SSR / hydration).

**Why hardcoded OG/canonical URLs?** The `index.html` is a static file — it can't dynamically read the current URL. We'll use the production URL as the canonical. For local dev, the OG tags will point to the wrong URL, which is acceptable since OG tags are only consumed by social media crawlers and agents, not by the SPA itself.

**Alternative:** Generate `index.html` dynamically in the Go server (inject the correct canonical/OG URLs per request). This would add complexity to the SPA handler but give correct URLs for all environments. For Phase 2, the static approach is sufficient.

#### 4.2.2 Resolving `html.text-ratio`

The text-ratio check measures visible text vs. HTML markup. With our `<noscript>` headings, the ratio will improve slightly but likely still fail. The fundamental issue is that the SPA renders content client-side.

Two approaches:

- **Quick fix (recommended):** Add a meaningful `<noscript>` block with descriptive text. This won't make the ratio high enough on its own, but it provides agent-readable content.
- **Full fix (future):** Server-side rendering (SSR) with a framework like Next.js, or pre-rendering the SPA to static HTML. This is a large architectural change beyond the scope of this ticket.

For now, add a `<noscript>` block with the key information:

```html
<noscript>
  <h1>Glazed Help Browser</h1>
  <p>Documentation browser for the Glazed CLI framework and Go-Go-Golems tools.</p>
  <h2>Packages</h2>
  <p>Available packages: Glazed, Pinocchio. Use the API at /api/packages for a full list.</p>
  <h2>API Endpoints</h2>
  <p>GET /api/packages — list all packages</p>
  <p>GET /api/sections — list sections with filters</p>
  <p>GET /api/sections/{slug} — get section content</p>
  <h2>URL Scheme</h2>
  <p>/{package}/{version}/sections/{slug} — navigate to a specific section</p>
</noscript>
```

#### 4.2.3 Resolving `html.glossary-link`

Add a link to a glossary page. The simplest approach is to add a glossary section to the AGENTS.md file (which we're already generating) and link to it from the `<noscript>` block:

```html
<a href="/AGENTS.md">Glossary &amp; Agent Reference</a>
```

This is a lightweight solution — we don't need a full separate glossary page. The AGENTS.md serves as both the agent reference and the glossary.

---

### 4.3 Phase 3: Markdown Content Negotiation

**Goal:** When an agent sends `Accept: text/markdown`, return the raw Markdown content of the requested section instead of the HTML SPA.

#### 4.3.1 New Handler: Markdown Mirror

Add a handler in `pkg/help/server/wellknown.go` (or a new file `pkg/help/server/markdown_mirror.go`) that intercepts requests with `Accept: text/markdown`:

```
In NewServeHandler's routing function:
    cleanPath := stdpath.Clean("/" + r.URL.Path)
    
    // ... /api/ check ...
    // ... well-known file check ...
    
    // NEW: Markdown content negotiation
    if acceptHeaderWantsMarkdown(r) && looksLikeSectionURL(cleanPath):
        markdownHandler.ServeHTTP(w, r)
        return
    
    // ... SPA fallback ...
```

Pseudocode for `acceptHeaderWantsMarkdown`:

```go
func acceptHeaderWantsMarkdown(r *http.Request) bool {
    accept := r.Header.Get("Accept")
    return strings.Contains(accept, "text/markdown")
}
```

Pseudocode for `looksLikeSectionURL`:

```go
func looksLikeSectionURL(path string) bool {
    // Match: /{package}/{version}/sections/{slug}
    // or: /{package}/_/sections/{slug}
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    return len(parts) >= 4 && parts[2] == "sections"
}
```

Pseudocode for the markdown mirror handler:

```go
func (h *MarkdownMirrorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Parse: /{package}/{version}/sections/{slug}
    path := stdpath.Clean("/" + r.URL.Path)
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    // parts[0] = package, parts[1] = version, parts[2] = "sections", parts[3] = slug
    
    packageName := parts[0]
    packageVersion := parts[1]
    if packageVersion == "_" {
        packageVersion = ""
    }
    slug := parts[3]
    
    // Look up section
    section, err := h.deps.Store.GetByPackageSlug(ctx, packageName, packageVersion, slug)
    if err != nil:
        http.NotFound(w, r)
        return
    
    // Build Markdown response with frontmatter
    var body strings.Builder
    body.WriteString("---\n")
    body.WriteString("title: " + section.Title + "\n")
    body.WriteString("description: " + section.Short + "\n")
    body.WriteString("doc_version: " + packageVersion + "\n")
    body.WriteString("last_updated: " + time.Now().UTC().Format(time.RFC3339) + "\n")
    body.WriteString("---\n\n")
    body.WriteString(section.Content)
    
    // Add sitemap section at the bottom
    body.WriteString("\n\n## Sitemap\n\n")
    // List sibling sections
    siblings := Store.Find(InPackageVersion(packageName, packageVersion))
    for each sibling:
        body.WriteString("- [" + sibling.Title + "](/" + packageName + "/" + version + "/sections/" + sibling.Slug + ")\n")
    
    w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
    w.Header().Set("Link", "<"+baseURL+path+">; rel=\"canonical\"")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(body.String()))
}
```

This resolves:
- `markdown.content-negotiation` — returns Markdown for `Accept: text/markdown`
- `markdown.frontmatter` — includes required frontmatter (title, description, doc_version, last_updated)
- `markdown.canonical-header` — sends `Link: <url>; rel="canonical"` header
- `markdown.sitemap-section` — includes `## Sitemap` section

---

### 4.4 Phase 4: AGENTS.md and llms.txt Content Quality

This phase is about making the *content* of these files genuinely useful, not just structurally correct. Phase 1 gets them served with the right Content-Type; Phase 4 makes them worth reading.

The content templates are already detailed in sections 4.1.3 above. The key additions:

1. **AGENTS.md** must document at least 2 of: install, configure, usage (resolves `agents-md.has-min-sections`). The template in 4.1.3 covers all three.

2. **llms.txt** must include links to section pages (resolves `llms-txt.md-extensions`). The template in 4.1.3 includes links in the format `[Title](/package/version/sections/slug)`.

3. **Both files** should be generated dynamically from the database, so they always reflect the current state of loaded packages.

---

## 5. Implementation Plan

### Phase 1: Server-Side Well-Known File Serving (3-4 hours)

| Step | Description | Files changed |
|------|-------------|---------------|
| 1.1 | Add `--base-url` flag to `ServeSettings` and `ServeCommand` | `pkg/help/server/serve.go` |
| 1.2 | Create `wellknown.go` with `WellKnownHandler` and `CanHandle` method | `pkg/help/server/wellknown.go` (new) |
| 1.3 | Implement `generateLlmsTxt` | `pkg/help/server/wellknown.go` |
| 1.4 | Implement `generateRobotsTxt` | `pkg/help/server/wellknown.go` |
| 1.5 | Implement `generateAgentsMd` | `pkg/help/server/wellknown.go` |
| 1.6 | Implement `generateSitemapXML` | `pkg/help/server/wellknown.go` |
| 1.7 | Implement `generateSitemapMd` | `pkg/help/server/wellknown.go` |
| 1.8 | Wire `WellKnownHandler` into `NewServeHandler` before SPA fallback | `pkg/help/server/serve.go` |
| 1.9 | Add `Store.ListPackages` method (if not already available) | `pkg/help/store/store.go` |
| 1.10 | Add Go tests for each generator | `pkg/help/server/wellknown_test.go` (new) |
| 1.11 | Test with `curl` and re-run a14y | — |

### Phase 2: HTML Metadata Enrichment (1-2 hours)

| Step | Description | Files changed |
|------|-------------|---------------|
| 2.1 | Add `<meta description>`, OG tags, canonical link to `index.html` | `web/index.html` |
| 2.2 | Add JSON-LD structured data to `index.html` | `web/index.html` |
| 2.3 | Add `<link rel="alternate" type="text/markdown">` to `index.html` | `web/index.html` |
| 2.4 | Add `<noscript>` block with headings and text | `web/index.html` |
| 2.5 | Add glossary link in `<noscript>` | `web/index.html` |
| 2.6 | Rebuild SPA assets, re-run a14y | — |

### Phase 3: Markdown Content Negotiation (2-3 hours)

| Step | Description | Files changed |
|------|-------------|---------------|
| 3.1 | Create `markdown_mirror.go` with `MarkdownMirrorHandler` | `pkg/help/server/markdown_mirror.go` (new) |
| 3.2 | Wire into `NewServeHandler` before SPA fallback | `pkg/help/server/serve.go` |
| 3.3 | Add `acceptHeaderWantsMarkdown` helper | `pkg/help/server/markdown_mirror.go` |
| 3.4 | Implement frontmatter generation | `pkg/help/server/markdown_mirror.go` |
| 3.5 | Implement `## Sitemap` section in Markdown output | `pkg/help/server/markdown_mirror.go` |
| 3.6 | Implement `Link: <url>; rel="canonical"` header | `pkg/help/server/markdown_mirror.go` |
| 3.7 | Test with `curl -H 'Accept: text/markdown'` | — |
| 3.8 | Re-run a14y | — |

### Phase 4: Content Quality and Polish (1-2 hours)

| Step | Description | Files changed |
|------|-------------|---------------|
| 4.1 | Refine AGENTS.md template with complete install/config/usage | `pkg/help/server/wellknown.go` |
| 4.2 | Refine llms.txt with package links and section summaries | `pkg/help/server/wellknown.go` |
| 4.3 | Add dynamic section counts and version lists to sitemap.md | `pkg/help/server/wellknown.go` |
| 4.4 | Final a14y re-run, verify 80+ score | — |

---

## 6. Expected Score Improvement

| Phase | Checks resolved | Estimated score |
|-------|----------------|-----------------|
| Baseline | — | 42 |
| Phase 1 | `sitemap-xml.valid`, `sitemap-md.has-structure`, `agents-md.has-min-sections`, `llms-txt.content-type`, `llms-txt.md-extensions`, `discovery.indexed` | ~60 |
| Phase 2 | `html.canonical-link`, `html.meta-description`, `html.og-title`, `html.og-description`, `html.json-ld`, `html.glossary-link`, `markdown.alternate-link`, `html.headings` (partial) | ~72 |
| Phase 3 | `markdown.content-negotiation`, `markdown.canonical-header`, `markdown.frontmatter`, `markdown.sitemap-section` | ~82 |
| Phase 4 | Content quality improvements (may help `html.text-ratio` if `<noscript>` text is sufficient) | ~85 |

Checks that may still fail after all phases:
- `html.text-ratio` — The SPA shell is minimal HTML; the `<noscript>` block helps but may not reach 15%. Full fix requires SSR.
- `html.headings` — Depends on whether `<noscript>` headings count. Some a14y versions may skip `<noscript>` content.

---

## 7. Testing Strategy

### 7.1 Unit Tests (Go)

For each well-known file generator, write a test that:
1. Creates a `Store` with known test data
2. Calls the generator
3. Asserts the output contains expected content (headings, links, XML structure)
4. Asserts the Content-Type is correct

```go
func TestGenerateLlmsTxt(t *testing.T) {
    // Setup store with test data
    store := setupTestStore(t)
    deps := HandlerDeps{Store: store}
    handler := NewWellKnownHandler(deps)
    
    req := httptest.NewRequest("GET", "/llms.txt", nil)
    req.Host = "docs.yolo.scapegoat.dev"
    
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Header().Get("Content-Type"), "text/plain")
    
    body := rec.Body.String()
    assert.Contains(t, body, "# Glazed Help Browser")
    assert.Contains(t, body, "## Packages")
    assert.Contains(t, body, "](https://docs.yolo.scapegoat.dev/") // links present
}
```

### 7.2 Integration Tests (curl)

```bash
# Start server with --base-url
/tmp/glaze-test serve --from-sqlite-dir /tmp/help-dbs --address :8088 --base-url https://docs.yolo.scapegoat.dev

# Verify each well-known file
curl -s http://localhost:8088/llms.txt | head -5          # Should be plain text, not HTML
curl -s http://localhost:8088/robots.txt | head -5          # Should be robots.txt content
curl -s http://localhost:8088/AGENTS.md | head -5          # Should be Markdown
curl -s http://localhost:8088/sitemap.xml | head -5        # Should be XML
curl -s http://localhost:8088/sitemap.md | head -5         # Should be Markdown

# Verify content negotiation
curl -s -H 'Accept: text/markdown' http://localhost:8088/glazed/_/sections/build-first-command | head -10
# Should return Markdown with frontmatter, not HTML

# Verify canonical header
curl -sI -H 'Accept: text/markdown' http://localhost:8088/glazed/_/sections/build-first-command | grep Link
# Should contain: Link: <https://docs.yolo.scapegoat.dev/glazed/_/sections/build-first-command>; rel="canonical"
```

### 7.3 a14y Re-run

After each phase, re-run the audit against the local dev server:

```bash
# Start Vite dev server (for the SPA with updated index.html)
cd web && pnpm dev

# Start Go server
/tmp/glaze-test serve --from-sqlite-dir /tmp/help-dbs --address :8088 --base-url https://docs.yolo.scapegoat.dev

# Run a14y
npx -y a14y check http://localhost:5173 --mode site --max-pages 50
```

---

## 8. Design Decisions

### 8.1 Dynamic generation vs. static files

**Decision:** Generate well-known files dynamically from the database.

**Rationale:** The help database changes when new packages are loaded or when `--reload-interval` triggers a refresh. Static files would become stale. Dynamic generation ensures `/sitemap.xml` always lists every section that the API can serve.

**Alternative:** Write static files at startup and on each reload. This would be simpler (just `os.WriteFile`) but would require a reload hook and could serve stale content during the reload window.

### 8.2 `<noscript>` headings vs. SSR

**Decision:** Use `<noscript>` headings for agent readability rather than implementing full server-side rendering.

**Rationale:** SSR would require migrating from Vite + React to Next.js or adding a pre-rendering step. This is a massive architectural change for a documentation browser. The `<noscript>` approach provides heading elements for crawlers without changing the rendering pipeline.

**Alternative:** Pre-render the SPA to static HTML at build time using a headless browser (e.g., `react-snap`, `prerender-spa-plugin`). This would give real headings and content in the HTML but adds complexity to the build pipeline.

### 8.3 `--base-url` flag vs. request-derived URL

**Decision:** Add an explicit `--base-url` flag with a fallback to request-derived URL.

**Rationale:** Behind reverse proxies (Caddy, Cloud Run, nginx), the request's `Host` header and protocol may not match the public URL. An explicit flag is unambiguous. The fallback handles the local dev case where `--base-url` is not set.

### 8.4 Markdown mirror scope

**Decision:** Only mirror section content URLs (matching `/{package}/{version}/sections/{slug}`), not the root or package index pages.

**Rationale:** The root and package index pages are navigation/UI constructs, not content pages. They don't have a meaningful Markdown representation. Section pages have Markdown source in the database, making the mirror straightforward.

---

## 9. File Reference Summary

| File | Phase | Change type |
|------|-------|-------------|
| `pkg/help/server/wellknown.go` | 1, 4 | New file — well-known file handlers |
| `pkg/help/server/markdown_mirror.go` | 3 | New file — Markdown content negotiation |
| `pkg/help/server/serve.go` | 1, 3 | Modify — add `--base-url` flag, wire handlers |
| `pkg/help/server/types.go` | 1 | Possibly add `BaseURL` field to `ServeSettings` |
| `pkg/help/server/wellknown_test.go` | 1 | New file — tests for well-known file generators |
| `pkg/help/store/store.go` | 1 | Possibly add `ListPackages` method |
| `web/index.html` | 2 | Modify — add meta tags, OG tags, JSON-LD, noscript |

---

## 10. Glossary

| Term | Definition |
|------|-----------|
| **a14y** | Agent readability — how well an AI agent can discover, fetch, parse, and comprehend a website. Not to be confused with WCAG (disability accessibility). |
| **llms.txt** | A plain-text file at the root of a website that provides LLMs with a structured overview of the site's content. See [llmstxt.org](https://llmstxt.org/). |
| **AGENTS.md** | A Markdown file at the root of a website that tells AI agents how to install, configure, and use the site/tool. See [agents.md](https://agents.md/). |
| **SPA fallback** | The server pattern where all non-API, non-static-asset requests return `index.html` so that client-side routing can handle them. |
| **Content negotiation** | The HTTP pattern where the server returns different representations of the same URL based on the `Accept` header. |
| **JSON-LD** | JSON for Linked Data — a W3C standard for embedding structured data in HTML pages using `<script type="application/ld+json">`. |
| **Open Graph** | A protocol (originally by Facebook) for embedding metadata in HTML using `<meta property="og:*">` tags. Used by social media platforms and some crawlers. |
| **Canonical URL** | The preferred URL for a page, declared via `<link rel="canonical">`. Tells crawlers which URL to index when multiple URLs lead to the same content. |
| **Sitemap** | A file (`sitemap.xml` or `sitemap.md`) that lists all URLs on a site for crawlers and agents to discover. |
