---
ticket: DOCSCTL-SSR
doc_type: design-doc
title: "SSR Sidecar: Analysis and Implementation Guide"
status: active
intent: long-term
topics: [docsctl, ssr, react, node, kubernetes, sidecar]
created: 2026-05-25
---

# SSR Sidecar: Analysis and Implementation Guide

## 1. Executive Summary

Add a Node.js SSR sidecar to the docs browser deployment. The Go server proxies
SPA page requests to the sidecar, which renders React to full HTML using
`renderToString`. The client hydrates the server-rendered HTML — no flash, no
double-render. Local dev (`glaze serve` without sidecar) falls back to current
client-side rendering.

**Why:** The SPA shell currently returns an empty `<div id="root"></div>` to
crawlers and AI agents. This tanks our a14y score (42/100) and makes the site
invisible to search engines. SSR gives every page real headings, text content,
and metadata before JavaScript runs.

## 2. Architecture

```
Production (k3s pod):
┌──────────────────────────────────────────────┐
│  Pod                                          │
│  ┌──────────────┐     ┌────────────────────┐ │
│  │  Go server    │────▶│  Node SSR sidecar   │ │
│  │  (:8088)     │     │  (:8089)            │ │
│  │              │◀────│                     │ │
│  │  /api/* → own│     │  Calls /api/* for   │ │
│  │  /* → proxy  │     │  data, renders      │ │
│  │    to :8089  │     │  React → full HTML   │ │
│  └──────────────┘     └────────────────────┘ │
└──────────────────────────────────────────────┘

Local dev (fallback):
┌──────────────────────────┐
│  Go server (:8088)       │
│  /api/* → API handler    │
│  /*     → SPA fallback   │
│          (index.html)    │
└──────────────────────────┘
```

## 3. Request Flow

```
1. Request hits Go server
2. Path starts with /api/ → Go API (unchanged)
3. Path matches a static asset → Go file server (unchanged)
4. Go server checks --ssr-url flag:
   a. SET: reverse-proxy the request to Node SSR on :8089
      - Node fetches data from Go API on localhost:8088
      - Node renders React to HTML
      - Node returns complete HTML with preloaded state
      - Client hydrates
   b. NOT SET: serve index.html (current SPA behavior, unchanged)
```

## 4. Key Design Decisions

### 4.1 SSR via sidecar, not embedded Node

The Go binary stays pure Go. No embedded Node.js runtime, no CGO, no WASM.
The sidecar is a separate container in the same pod, communicating over
localhost. Local dev works without it.

### 4.2 SSR entry point reuses the same App component

The SSR server imports the same `<App />` component that the client uses.
This ensures the server-rendered HTML and the client-rendered HTML are
structurally identical, which is required for correct React hydration.

### 4.3 Preloaded state via window.__PRELOADED_STATE__

The SSR server fetches data from the Go API and injects it into the HTML as
`window.__PRELOADED_STATE__`. The client reads this on mount and uses it to
pre-populate RTK Query's cache. This avoids duplicate API calls during
hydration.

### 4.4 BrowserRouter on server: use StaticRouter

React Router's `BrowserRouter` doesn't work on the server (it reads
`window.location`). The SSR entry point uses `StaticRouter` instead, which
takes a `location` prop from the request URL.

### 4.5 Hydration strategy: hydrateRoot

The client switches from `createRoot().render()` to `hydrateRoot()`. React
reuses the server-rendered DOM nodes instead of replacing them. This gives
zero flash of content and instant interactivity.

## 5. File Layout

New files:
```
web/
  src/
    entry-client.tsx    # Client entry (hydrateRoot instead of createRoot)
    entry-server.tsx    # Server entry (StaticRouter, renderToString)
  server.mjs            # Node.js SSR HTTP server (Express)
  ssr.Dockerfile        # Docker image for the sidecar
```

Modified files:
```
web/src/main.tsx        # → entry-client.tsx (rename + hydration)
web/vite.config.ts      # Add SSR build config
web/package.json        # Add SSR scripts, express dependency
pkg/help/server/serve.go # Add --ssr-url flag + proxy logic
```

## 6. Implementation Tasks

1. Create `entry-server.tsx` — SSR entry point with StaticRouter
2. Create `server.mjs` — Express server that renders React
3. Create `entry-client.tsx` — hydration entry point
4. Add SSR build to `vite.config.ts`
5. Add SSR scripts to `package.json`
6. Add `--ssr-url` flag to Go server + reverse proxy
7. Test SSR locally (Go server + Node SSR sidecar)
8. Create `ssr.Dockerfile` for k3s deployment
9. Run a14y audit to verify score improvement

## 7. Expected Outcome

- Every page returns full HTML with headings, text, and metadata
- a14y score jumps from 42 to ~75+ (combined with DOCSCTL-A14Y fixes)
- SEO: search engines can index all section pages
- Hydration: zero flash, instant interactivity
