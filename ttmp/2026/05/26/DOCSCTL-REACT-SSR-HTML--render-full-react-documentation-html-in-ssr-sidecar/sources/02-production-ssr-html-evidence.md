---
Title: Production SSR HTML Evidence
Ticket: DOCSCTL-REACT-SSR-HTML
Status: active
Topics:
    - docs-yolo
    - ssr
    - frontend
    - react
DocType: reference
Intent: short-term
Owners:
    - manuel
RelatedFiles: []
ExternalSources: []
Summary: "Production validation evidence after deploying data-backed React SSR HTML rendering."
LastUpdated: 2026-05-26T18:50:00-04:00
WhatFor: "Evidence for SSR implementation and validation."
WhenToUse: "Use to compare pre/post data-backed React SSR behavior."
---

# Production SSR HTML Evidence

Captured after deploying Glazed `981a6db` / k3s `2a27cdb`.

## Kubernetes deployment
```text
docs-browser => ghcr.io/go-go-golems/glazed:sha-981a6db
docs-registry => ghcr.io/go-go-golems/glazed:sha-981a6db
docs-ssr => ghcr.io/go-go-golems/glazed-ssr:sha-981a6db
```

## https://docs.yolo.scapegoat.dev/glazed/v1.2.15
Headers:
```text
HTTP/2 200 
content-type: text/html; charset=utf-8
```
Checks:
```text
__PRELOADED_STATE__: 1
RTK Query state marker (queries): 1
Documentation Index: 1
Known section title: 4
Known article body text: 0
```

## https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table
Headers:
```text
HTTP/2 200 
content-type: text/html; charset=utf-8
```
Checks:
```text
__PRELOADED_STATE__: 1
RTK Query state marker (queries): 1
Documentation Index: 0
Known section title: 8
Known article body text: 2
```

## Static assets and API checks
```text
--- https://docs.yolo.scapegoat.dev/assets/main-BZ3FdiSW.js
HTTP/2 200 
content-type: text/javascript; charset=utf-8
--- https://docs.yolo.scapegoat.dev/glazed/v1.2.15/assets/main-BZ3FdiSW.js
HTTP/2 200 
content-type: text/javascript; charset=utf-8
--- https://docs.yolo.scapegoat.dev/glazed/v1.2.15/sections/exposing-a-simple-sql-table.md
HTTP/2 200 
content-type: text/markdown; charset=utf-8
--- https://docs.yolo.scapegoat.dev/api/health
HTTP/2 200 
content-type: application/json
--- https://docs-registry.yolo.scapegoat.dev/healthz
HTTP/2 200 
content-type: application/json
```

## Playwright result

Direct section URL hydrated successfully. Console had no hydration/module MIME warnings; the only error was the known external Chicago font 404.
