---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: glazed/pkg/help/model/section.go
      Note: Section model with PackageName/PackageVersion
    - Path: glazed/pkg/help/server/handlers.go
      Note: API handler routes
    - Path: glazed/pkg/help/server/serve.go
      Note: Go HTTP server routing and SPA fallback
    - Path: glazed/pkg/help/server/types.go
      Note: Response/request types
    - Path: glazed/pkg/help/store/query.go
      Note: Query predicates including InPackageVersion
    - Path: glazed/pkg/help/store/store.go
      Note: SQLite store with package/version/slug composite key
    - Path: glazed/web/public/site-config.js
      Note: Runtime config for SPA
    - Path: glazed/web/src/App.tsx
      Note: Root component with routing and package/section state
    - Path: glazed/web/src/main.tsx
      Note: SPA entry point using HashRouter
    - Path: glazed/web/src/services/api.ts
      Note: RTK Query API layer with resolveApiBaseUrl
    - Path: glazed/web/src/types/index.ts
      Note: TypeScript types mirroring Go response types
    - Path: pinocchio/pkg/spa/spa.go
      Note: SPA handler with fallback routing
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Deep Linking Design: Semantic URLs for docsctl

## 1. Executive Summary

The docs.yolo.scapegoat.dev documentation site currently uses hash-based routing (`/#/sections/slug`) that prevents deep linking, bookmarking, and search engine indexing. This document describes the current architecture in full detail, identifies every gap preventing proper URL-based navigation, and proposes a phased implementation that introduces semantic URLs of the form:

```
/{package}/{version}/sections/{slug}#{heading-id}
```

so that navigating to `https://docs.yolo.scapegoat.dev/pinocchio/v1.2.15/sections/build-your-first-glazed-command#prerequisites` loads the correct package, version, and page, scrolls to the heading, and keeps all of that in the URL bar.

The design touches three layers: the Go HTTP server (SPA fallback + API routing), the React SPA (client-side router, URL synchronization, heading scroll), and the docsctl publish/registry pipeline (no changes needed). The implementation is structured so that each phase delivers working, testable value.

---

## 2. Problem Statement and Scope

### 2.1 What is broken today

When you open `https://docs.yolo.scapegoat.dev` and click on a section in the sidebar, the URL changes to something like:

```
https://docs.yolo.scapegoat.dev/#/sections/adding-field-types
```

This URL has three fundamental problems:

1. **No package or version in the URL.** The page shows the "Glazed" package at version "vtest" by default, but the URL does not encode this. If you switch to Pinocchio v1.2.15 and then share the URL, the recipient will see Glazed instead.

2. **Hash routing prevents server-side routing.** Because the SPA uses `HashRouter`, everything after `#` is client-only. The Go server never sees it. If you navigate directly to `https://docs.yolo.scapegoat.dev/pinocchio/v1.2.15/sections/some-slug`, the Go server returns `index.html` (SPA fallback), but the React router has no route that matches `/pinocchio/v1.2.15/sections/some-slug`, so the page is blank.

3. **Heading anchors don't survive navigation.** Clicking a heading in the documentation tree uses `navigate('/sections/slug#heading-id')`, but React Router's hash-based router already consumes the `#` for its own routing, so heading anchors are unreliable.

### 2.2 Scope

- **In scope**: Changing the React SPA from `HashRouter` to `BrowserRouter`, adding URL parameters for package/version/page/section to both the client routes and the Go server's SPA fallback, making heading scroll work from URLs, and ensuring existing API routes are unaffected.
- **Out of scope**: Search engine optimization (sitemap, SSR), authentication, static-site generation mode, changes to the docsctl `publish` or `validate` commands, and changes to the registry handler.

---

## 3. Current-State Architecture (Evidence-Based)

### 3.1 System overview

The docs system consists of four major components that work together:

```
┌─────────────────────────────────────────────────────────────┐
│  docs.yolo.scapegoat.dev                                    │
│                                                             │
│  ┌───────────────────┐    ┌──────────────────────────────┐ │
│  │  Go HTTP Server   │    │  React SPA (HashRouter)       │ │
│  │  (serve.go)       │    │  (App.tsx, main.tsx)         │ │
│  │                   │    │                              │ │
│  │  /api/*  ─────────┼───▶│  RTK Query → /api/*         │ │
│  │  /*     ──────────┼───▶│  HashRouter: /#/sections/:slug│ │
│  └───────────────────┘    └──────────────────────────────┘ │
│           │                                                 │
│           ▼                                                 │
│  ┌───────────────────┐                                     │
│  │  SQLite Store     │                                     │
│  │  (store.go)       │                                     │
│  │                   │                                     │
│  │  sections table   │                                     │
│  │  FTS5 search      │                                     │
│  └───────────────────┘                                     │
└─────────────────────────────────────────────────────────────┘

         ┌──────────────────────────────────┐
         │  docsctl publish CLI             │
         │  (publish.go)                    │
         │                                  │
         │  PUT /v1/packages/{pkg}/versions │
         │      /{ver}/sqlite               │
         └──────────────────────────────────┘
                  │
                  ▼
         ┌──────────────────────────────────┐
         │  docs-registry server            │
         │  (docs-registry/main.go)        │
         │                                  │
         │  Writes .db files to package-root│
         │  /{pkg}/{ver}/{pkg}.db          │
         └──────────────────────────────────┘
```

### 3.2 Go HTTP server

**File**: `glazed/pkg/help/server/serve.go`

The `ServeCommand` starts an HTTP server that combines an API handler and an optional SPA handler. The key routing logic lives in `NewServeHandler`:

```go
// glazed/pkg/help/server/serve.go, NewServeHandler
func NewServeHandler(deps HandlerDeps, spaHandler http.Handler) http.Handler {
    // ...auto-assign default package...
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

The server splits on `/api` prefix: if the path starts with `/api`, it goes to the API handler; otherwise, it falls through to the SPA handler. The SPA handler (`pinocchio/pkg/spa/spa.go`) implements the standard SPA fallback pattern: serve `index.html` for any path that doesn't match a static asset.

**API routes** (defined in `glazed/pkg/help/server/handlers.go`):

| Route | Method | Purpose |
|---|---|---|
| `/api/health` | GET | Health check |
| `/api/packages` | GET | List packages with versions |
| `/api/sections` | GET | List/filter sections (query params: `package`, `version`, `type`, `topic`, `q`, `limit`, `offset`) |
| `/api/sections/search` | GET | Same as `/api/sections` with search |
| `/api/sections/{slug}` | GET | Get full section by slug (query params: `package`, `version`) |

**Important**: The `/api/sections/{slug}` route accepts `package` and `version` as query parameters, not as path segments. This is fine — the API is stable and we don't need to change it.

### 3.3 React SPA

**File**: `glazed/web/src/main.tsx`

The SPA entry point uses `HashRouter`:

```tsx
// glazed/web/src/main.tsx
ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <HashRouter>
        <ErrorBoundary>
          <App />
        </ErrorBoundary>
      </HashRouter>
    </Provider>
  </React.StrictMode>,
);
```

**File**: `glazed/web/src/App.tsx`

The `App` component reads the current route to extract the active slug:

```tsx
const activeSlug = useMemo(() => {
    const match = matchPath('/sections/:slug', location.pathname);
    return match?.params.slug ?? null;
}, [location.pathname]);
```

The route `/sections/:slug` is the only client-side route. When a section is selected, the code calls:

```tsx
const handleSelect = (slug: string) => {
    navigate(`/sections/${slug}`);
};
```

And heading navigation uses:

```tsx
const handleSelectHeading = (slug: string, headingId: string) => {
    navigate(`/sections/${slug}#${headingId}`);
};
```

**File**: `glazed/web/src/services/api.ts`

The API layer uses RTK Query and resolves its base URL from the current pathname:

```tsx
export function resolveApiBaseUrl(pathname: string): string {
    if (!pathname || pathname === '/') {
        return '/api';
    }
    const mountPrefix = pathname.replace(/\/+$/, '');
    return `${mountPrefix}/api`;
}
```

This function is designed to work when the app is mounted under a prefix like `/help`. With `HashRouter`, `window.location.pathname` is always `/` (the hash part isn't part of the pathname), so `resolveApiBaseUrl` always returns `/api`. With `BrowserRouter`, the pathname will include the package/version segments, and this function will break. This is a critical bug to fix.

**File**: `glazed/web/src/types/index.ts`

The TypeScript types mirror the Go response types exactly. Key types:

- `SectionSummary` — list shape with `packageName`, `packageVersion`, `slug`, `type`, `title`, `headings`
- `SectionDetail` — full shape with `content`
- `PackageSummary` — package with `name`, `versions`, `sectionCount`
- `ListPackagesResponse` — includes `defaultPackage`, `defaultVersion`

### 3.4 Data model

**File**: `glazed/pkg/help/model/section.go`

Each `Section` has a composite identity: `(package_name, package_version, slug)`. The unique index in the SQLite store is:

```sql
CREATE UNIQUE INDEX idx_sections_package_version_slug
    ON sections(package_name, package_version, slug);
```

This means two packages can have the same slug, and even the same package at different versions can have the same slug. The slug alone is not globally unique — it must be qualified by package and version.

**File**: `glazed/pkg/help/store/store.go`

The `GetByPackageSlug` method retrieves a section by the full composite key:

```go
func (s *Store) GetByPackageSlug(ctx context.Context, packageName, packageVersion, slug string) (*model.Section, error) {
    query := `SELECT ... FROM sections WHERE package_name = ? AND package_version = ? AND slug = ?`
    // ...
}
```

### 3.5 SPA handler (fallback routing)

**File**: `pinocchio/pkg/spa/spa.go`

```go
func NewHandler() (http.Handler, error) {
    // ...
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...try static assets first...
        // SPA fallback: serve index.html for all unknown paths
        serveSPAIndex(w, r, indexBytes)
    }), nil
}
```

This already does the right thing: for any path that doesn't match a static asset file, it returns `index.html`. The problem isn't on the server side — the server already falls back correctly. The problem is that the client-side `HashRouter` doesn't parse paths like `/pinocchio/v1.2.15/sections/slug`.

### 3.6 Package loading and directory layout

**File**: `glazed/pkg/help/loader/sources.go`

The `SQLiteDirLoader` discovers packages from a directory tree using this convention:

```
X.db           → package X, no version
X/X.db         → package X, no version  
X/Y/X.db       → package X, version Y
```

**File**: `glazed/pkg/help/publish/directory_store.go`

When the registry publishes a package, it writes to:

```
{package-root}/{packageName}/{version}/{packageName}.db
```

This is the layout that `--from-sqlite-dir` scans when the serve command loads packages.

### 3.7 docsctl publish flow

**File**: `glazed/cmd/docsctl/publish.go`

The `docsctl publish` command validates a SQLite help database locally, then uploads it to the registry:

```go
url := strings.TrimRight(opts.Server, "/") +
    fmt.Sprintf("/v1/packages/%s/versions/%s/sqlite", opts.PackageName, opts.Version)
```

The upload goes to the docs-registry server, which writes the `.db` file into the directory store.

### 3.8 The docs.yolo.scapegoat.dev deployment

The production deployment runs `glaze serve --from-sqlite-dir <dir> --reload-interval 30s`, which:
1. Scans the directory for package/version `.db` files
2. Loads all sections into an in-memory SQLite store
3. Serves the API on `/api/*` and the SPA on `/*`
4. Periodically reloads the directory (every 30 seconds)

---

## 4. Gap Analysis

| # | Gap | Root Cause | Impact |
|---|---|---|---|
| G1 | URLs don't encode package/version | `HashRouter` with single route `/sections/:slug` | Sharing a URL shows wrong package; can't deep-link to a specific package version |
| G2 | `HashRouter` prevents semantic URLs | `main.tsx` uses `HashRouter` | All routes are behind `#`; server can't route; SEO impossible; blank page on direct navigation |
| G3 | API base URL resolution breaks with `BrowserRouter` | `resolveApiBaseUrl` uses `window.location.pathname` to derive API prefix | With real pathnames like `/pinocchio/v1.2.15/sections/slug`, the API URL would become `/pinocchio/v1.2.15/sections/api` |
| G4 | Heading anchors conflict with `HashRouter` | `#` is consumed by the router | `navigate('/sections/slug#heading')` doesn't work reliably |
| G5 | No URL-driven initial state | Package/version/section are set from `defaultPackage` response, not from URL | Direct navigation to a URL doesn't restore the correct state |
| G6 | SPA fallback doesn't set headers for caching | `serveSPAIndex` uses `no-cache` for `index.html` but doesn't differentiate between static assets and SPA routes | Not a blocking issue, but suboptimal for CDN caching |

---

## 5. Proposed Architecture and APIs

### 5.1 URL scheme

The new URL scheme encodes all navigation state in the path and fragment:

```
/{package}/{version}/sections/{slug}#{heading-id}
```

Examples:

| URL | Meaning |
|---|---|
| `/` | Redirect to default package/version |
| `/glazed/vtest` | Glazed package, version vtest, section list |
| `/pinocchio/v1.2.15` | Pinocchio package, version v1.2.15, section list |
| `/pinocchio/v1.2.15/sections/build-your-first-glazed-command` | Specific section |
| `/pinocchio/v1.2.15/sections/build-your-first-glazed-command#prerequisites` | Specific section + heading |

For unversioned packages (where version is empty in the data model), the URL uses `_` as a placeholder:

```
/glazed/_/sections/slug
```

Alternatively, unversioned packages can use a special version string that the server defines. The current data uses empty string for unversioned, but the `ListPackages` response may omit the version. We need a convention.

**Decision**: Use `_` as the URL segment for "no version". The frontend and the Go server will both understand this convention. When the user selects a package that has no version, the URL will show `/_/`.

### 5.2 React Router changes

Replace `HashRouter` with `BrowserRouter` and define routes that capture package, version, and slug:

```tsx
// Pseudocode for new router setup
<BrowserRouter>
  <Routes>
    <Route path="/" element={<Navigate to={defaultPath} replace />} />
    <Route path="/:package/:version" element={<App />} />
    <Route path="/:package/:version/sections/:slug" element={<App />} />
    <Route path="*" element={<Navigate to={defaultPath} replace />} />
  </Routes>
</BrowserRouter>
```

In the `App` component, the active package, version, and slug are all derived from the URL:

```tsx
// Pseudocode for URL-derived state
const { packageName, version, slug } = useParams();

// When user selects a package from the dropdown:
const handlePackageChange = (newPkg: string) => {
    const newVersion = getLatestVersion(newPkg);
    navigate(`/${newPkg}/${newVersion}`);
};

// When user selects a section:
const handleSelect = (slug: string) => {
    navigate(`/${packageName}/${version}/sections/${slug}`);
};

// When user clicks a heading:
const handleSelectHeading = (slug: string, headingId: string) => {
    navigate(`/${packageName}/${version}/sections/${slug}#${headingId}`);
};
```

### 5.3 API base URL resolution fix

The `resolveApiBaseUrl` function currently derives the API URL from the browser's pathname. This breaks with `BrowserRouter` because the pathname now includes `/pinocchio/v1.2.15/sections/slug`.

**Solution**: The API base URL should be resolved once at app startup from a fixed location, not from the dynamic pathname. There are two approaches:

**Option A (recommended): Use `site-config.js` to set the API base URL explicitly.**

```js
// web/public/site-config.js
window.__GLAZE_SITE_CONFIG__ = {
  mode: 'server',
  apiBaseUrl: '/api',
  siteTitle: 'Glazed Help Browser',
};
```

The `api.ts` already reads `window.__GLAZE_SITE_CONFIG__` and supports an explicit `apiBaseUrl` field. We just need to:
1. Set `apiBaseUrl: '/api'` in `site-config.js`
2. Change `resolveRuntimeBaseUrl` to prefer the explicit `apiBaseUrl` over the pathname-derived one

**Option B: Derive from the mount prefix, not the full pathname.**

This requires knowing the mount prefix at build time or runtime. It's more complex and fragile.

We go with **Option A** because it's a one-line change in config and a one-line change in the resolver.

### 5.4 Server-side changes

The Go server already does the right thing with SPA fallback: any non-`/api` path returns `index.html`. No server-side routing changes are needed.

However, we should add a small optimization: when the server receives a request for `/{package}/{version}/sections/{slug}`, and the `Accept` header is `application/json`, we could return the section JSON directly. This would allow curl/API clients to use semantic URLs too. But this is optional and not required for the core feature.

**Required server change**: Add a `Cache-Control: no-cache` header specifically for `index.html` responses, and add `Cache-Control: public, max-age=31536000, immutable` for hashed static assets (JS/CSS bundles). This is already partially done but should be validated.

### 5.5 Heading scroll with BrowserRouter

With `BrowserRouter`, the fragment (`#heading-id`) works natively because it's not consumed by the router. The scroll logic in `App.tsx` already handles this:

```tsx
useEffect(() => {
    if (!section) return;
    requestAnimationFrame(() => {
        if (activeHeadingId) {
            const heading = document.getElementById(activeHeadingId);
            if (heading) {
                heading.scrollIntoView({ block: 'start' });
                return;
            }
        }
        if (contentScrollRef.current) {
            contentScrollRef.current.scrollTop = 0;
        }
    });
}, [section, activeHeadingId]);
```

This code will work as-is with `BrowserRouter`. The `location.hash` parsing will also work correctly because the hash is no longer consumed by the router.

---

## 6. Pseudocode and Key Flows

### 6.1 Initial page load flow

```
Browser requests: GET /pinocchio/v1.2.15/sections/build-your-first-glazed-command#prerequisites

1. Go server receives request
2. Path does NOT start with /api → serve index.html (SPA fallback)
3. Browser loads React SPA
4. BrowserRouter matches route: /:package/:version/sections/:slug
5. App component extracts: packageName="pinocchio", version="v1.2.15", slug="build-your-first-glazed-command"
6. App reads location.hash → "prerequisites"
7. RTK Query fires:
   a. useListPackagesQuery() → GET /api/packages
   b. useListSectionsQuery({ packageName, version }) → GET /api/sections?package=pinocchio&version=v1.2.15
   c. useGetSectionQuery({ slug, packageName, version }) → GET /api/sections/build-your-first-glazed-command?package=pinocchio&version=v1.2.15
8. Section loads, heading "prerequisites" scrolls into view
9. Package selector shows "Pinocchio", version selector shows "v1.2.15"
```

### 6.2 Navigation flow (user clicks sidebar section)

```
User clicks "Build Your First Glazed Command" in sidebar

1. handleSelect("build-your-first-glazed-command") called
2. navigate(`/pinocchio/v1.2.15/sections/build-your-first-glazed-command`)
3. BrowserRouter updates URL without page reload
4. Route params change → slug changes → RTK Query refetches section
5. Section renders, scroll to top
```

### 6.3 Package/version change flow

```
User selects "Sqleton" from package dropdown

1. handlePackageChange("sqleton") called
2. Find latest version for sqleton (e.g., "v0.3.0")
3. navigate(`/sqleton/v0.3.0`)
4. BrowserRouter updates URL
5. Route params change → packageName changes → RTK Query refetches sections list
6. Section view shows EmptyState (no slug selected)
7. URL is now /sqleton/v0.3.0 — bookmarkable
```

### 6.4 API base URL resolution flow

```
SPA loads at /pinocchio/v1.2.15/sections/slug

1. api.ts reads window.__GLAZE_SITE_CONFIG__
2. apiBaseUrl is "/api" (from site-config.js)
3. All RTK Query requests go to /api/sections?package=pinocchio&...
   (NOT to /pinocchio/v1.2.15/api/sections)
```

### 6.5 Pseudocode: App.tsx route extraction

```tsx
// Current (broken with BrowserRouter)
const activeSlug = useMemo(() => {
    const match = matchPath('/sections/:slug', location.pathname);
    return match?.params.slug ?? null;
}, [location.pathname]);

// Proposed (works with BrowserRouter)
function App() {
    const { packageName, version, slug } = useParams();
    // packageName and version come from /:package/:version route
    // slug comes from /:package/:version/sections/:slug route (or undefined)

    const selectedPackage = packageName ?? '';
    const selectedVersion = version === '_' ? '' : (version ?? '');
    const activeSlug = slug ?? null;
    const activeHeadingId = location.hash.replace(/^#/, '');

    // ... rest of component
}
```

### 6.6 Pseudocode: PackageSelector URL sync

```tsx
const handlePackageChange = (newPkg: string) => {
    const nextPackage = packages.find(pkg => pkg.name === newPkg);
    const nextVersions = nextPackage?.versions ?? [];
    const newVersion = nextVersions[0] || '_';
    // Navigate to new package/version, clearing any selected section
    navigate(`/${newPkg}/${newVersion}`);
};

const handleVersionChange = (newVersion: string) => {
    navigate(`/${selectedPackage}/${newVersion}`);
};
```

### 6.7 Pseudocode: Redirect from root

```tsx
// When landing on /, redirect to default package/version
useEffect(() => {
    if (!packageData || packageName) return;
    const defaultPkg = packageData.defaultPackage || 'default';
    const defaultVer = packageData.defaultVersion || '_';
    navigate(`/${defaultPkg}/${defaultVer}`, { replace: true });
}, [packageData, packageName]);
```

---

## 7. Implementation Phases

### Phase 1: Switch to BrowserRouter with package/version routes

**Goal**: Replace `HashRouter` with `BrowserRouter` and add package/version to all routes.

**Files to change**:

1. **`glazed/web/src/main.tsx`** — Change `HashRouter` to `BrowserRouter`:
   ```tsx
   import { BrowserRouter, Routes, Route } from 'react-router-dom';
   // ...
   <BrowserRouter>
     <Routes>
       <Route path="/:package/:version" element={<App />} />
       <Route path="/:package/:version/sections/:slug" element={<App />} />
       <Route path="*" element={<App />} />
     </Routes>
   </BrowserRouter>
   ```

2. **`glazed/web/src/App.tsx`** — Extract package/version/slug from `useParams()` instead of `matchPath`. Update all `navigate()` calls to include `/${packageName}/${version}` prefix. Fix the `useEffect` that sets the initial package from `defaultPackage` to redirect to `/${defaultPkg}/${defaultVer}`.

3. **`glazed/web/public/site-config.js`** — Add `apiBaseUrl: '/api'` so the API base URL is explicit and not derived from the pathname.

4. **`glazed/web/src/services/api.ts`** — Ensure `resolveRuntimeBaseUrl` returns the explicit `apiBaseUrl` when available, without falling through to `resolveApiBaseUrl(pathname)`.

**Testing**:
- Navigate to `/` → should redirect to `/glazed/vtest`
- Click a section → URL should be `/glazed/vtest/sections/slug`
- Change package → URL should change to `/pinocchio/v1.2.15`
- Direct navigation to `/pinocchio/v1.2.15/sections/slug` → should load the correct page
- API calls should still go to `/api/*`

### Phase 2: Heading anchor support

**Goal**: Make heading anchors work correctly with the new URL scheme.

**Files to change**:

1. **`glazed/web/src/App.tsx`** — The existing `handleSelectHeading` already navigates with `#headingId`. With `BrowserRouter`, this works natively. Verify the `useEffect` scroll logic works with the new routing.

2. **`glazed/web/src/components/DocumentationTree/DocumentationTree.tsx`** — Ensure heading clicks call `onSelectHeading(slug, headingId)` which navigates to the full URL with fragment.

3. **`glazed/web/src/components/Markdown/MarkdownContent.tsx`** — If heading IDs are generated in the Markdown renderer, verify they match the IDs returned by the server's `ExtractHeadings` function (in `glazed/pkg/help/server/headings.go`).

**Testing**:
- Click a heading in the sidebar tree → URL should update with `#heading-id` and scroll to it
- Direct navigation to `/pkg/ver/sections/slug#heading-id` → should load page and scroll to heading
- Copy URL from browser → paste in new tab → should show correct page scrolled to heading

### Phase 3: Edge cases and polish

**Goal**: Handle edge cases like unversioned packages, 404 sections, and browser back/forward.

**Files to change**:

1. **`glazed/web/src/App.tsx`** — Handle the case where `useParams()` returns a version of `_` (map to empty string for API calls). Handle 404 responses from the API (section not found for package/version combination).

2. **`glazed/web/src/App.tsx`** — Ensure browser back/forward navigation correctly updates package, version, and section state from the URL.

3. **`glazed/pkg/help/server/serve.go`** — (Optional) Add a `Cache-Control` header for SPA index.html responses: `no-cache, no-store, must-revalidate`.

4. **`glazed/pkg/help/server/serve.go`** — (Optional) When the request path matches `/{package}/{version}/sections/{slug}` and the `Accept` header is `application/json`, proxy to the API handler. This makes semantic URLs work for API consumers too.

**Testing**:
- Navigate to unversioned package (`/_/`) → should work
- Navigate to non-existent section → should show error or empty state
- Browser back/forward → state should update correctly
- API calls with `Accept: application/json` → should return JSON (if Phase 3 optional feature implemented)

---

## 8. Testing and Validation Strategy

### 8.1 Manual testing checklist

- [ ] Load `/` → redirects to default package/version URL
- [ ] Click any section in sidebar → URL changes to `/{pkg}/{ver}/sections/{slug}`
- [ ] Change package dropdown → URL changes to `/{newPkg}/{newVer}`
- [ ] Change version dropdown → URL changes to `/{pkg}/{newVer}`
- [ ] Click heading in documentation tree → URL adds `#{headingId}`, scrolls to heading
- [ ] Copy full URL, paste in new tab → loads correct page with correct package/version/section
- [ ] Use browser back/forward → state updates correctly
- [ ] Direct navigation to `https://docs.yolo.scapegoat.dev/pinocchio/v1.2.15/sections/build-your-first-glazed-command#prerequisites` → loads correct page scrolled to heading
- [ ] API calls (check Network tab) still go to `/api/*`
- [ ] Static assets (JS/CSS) still load correctly

### 8.2 Automated testing

**Frontend unit tests** (`glazed/web/src/App.test.tsx`):
- Test that route params are extracted correctly
- Test that navigation calls `navigate()` with the correct URL pattern
- Test that the API base URL resolves to `/api` regardless of current pathname

**Go server tests** (`glazed/pkg/help/server/serve_test.go`):
- Existing tests should still pass (API routes unchanged)
- Add test: request to `/{pkg}/{ver}/sections/{slug}` returns `index.html` (SPA fallback)
- Add test: request to `/api/sections` still works

### 8.3 Integration testing

- Start the server with `glaze serve --from-sqlite-dir <dir>`
- Verify that navigating to `http://localhost:8088/pinocchio/v1.2.15/sections/slug` loads the SPA
- Verify that the SPA makes API calls to `http://localhost:8088/api/sections?package=pinocchio&version=v1.2.15`

---

## 9. Risks, Alternatives, and Open Questions

### 9.1 Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| BrowserRouter requires server-side SPA fallback for all routes | Already handled — the Go server already returns `index.html` for non-`/api` paths | No change needed |
| Existing bookmarks with `/#/sections/slug` break | Medium | Add a redirect from hash-based URLs to new URLs (a small `<script>` in `index.html` that detects `/#/` and redirects) |
| API base URL resolution breaks if `site-config.js` is missing | Low | Fall back to `/api` as default, which matches the common deployment |
| Version segment `_` for unversioned packages is confusing | Low | Could also use `latest` or omit version entirely with a two-segment route |

### 9.2 Alternatives considered

1. **Keep HashRouter but add package/version as query params**: `/#/sections/slug?package=pinocchio&version=v1.2.15`. This is easier but doesn't solve the bookmarking/SEO problem. The hash is still opaque to servers.

2. **Use query params with BrowserRouter**: `/sections/slug?package=pinocchio&version=v1.2.15`. This works but puts the package/version in a less prominent position. Path segments are more RESTful and more visible.

3. **SSR (Server-Side Rendering)**: Generate HTML on the server for each section. This is the "right" solution for SEO but is a much larger change and outside scope.

4. **Static site generation**: Pre-render all pages at build time. This is what the `mode: 'static'` config was designed for, but it's a separate feature.

### 9.3 Open questions

1. **Should unversioned packages use `_` or `latest` in the URL?** `_` is less ambiguous (it means "no version"), while `latest` implies a dynamic version. I recommend `_`.

2. **Should the root URL `/` redirect or render the default package?** Redirect is cleaner (the URL bar shows the actual package/version), but rendering avoids a network round-trip. I recommend redirect for clarity.

3. **Should we add a JSON API at the semantic URL path?** For example, `GET /pinocchio/v1.2.15/sections/build-your-first-glazed-command` with `Accept: application/json` returns the section JSON. This would make the semantic URLs work for API consumers too. I recommend deferring this to a later phase.

4. **How should the SPA handle 404s?** If a user navigates to a package/version/slug that doesn't exist, the API will return 404. The SPA should show a clear error message and offer a link back to the package's section list.

---

## 10. References

### Key files (absolute paths from workspace root)

| File | Role |
|---|---|
| `glazed/pkg/help/server/serve.go` | Go HTTP server, SPA/API routing, `NewServeHandler` |
| `glazed/pkg/help/server/handlers.go` | Go API handlers (`/api/sections`, `/api/packages`, etc.) |
| `glazed/pkg/help/server/types.go` | Go response/request types (`SectionSummary`, `SectionDetail`, etc.) |
| `glazed/pkg/help/server/middleware.go` | CORS middleware, `APIPathPrefix` constant |
| `glazed/pkg/help/server/headings.go` | Heading extraction from Markdown content, `SlugifyHeading` |
| `glazed/pkg/help/store/store.go` | SQLite store, `GetByPackageSlug`, `ListPackages`, `SetDefaultPackage` |
| `glazed/pkg/help/store/query.go` | Query compiler and predicates (`InPackageVersion`, `SlugEquals`, etc.) |
| `glazed/pkg/help/model/section.go` | `Section` struct with `PackageName`, `PackageVersion`, `Slug` |
| `glazed/pkg/help/loader/sources.go` | Content loaders (`SQLiteDirLoader`, `SQLiteLoader`, etc.) |
| `glazed/pkg/help/publish/registry.go` | Registry handler, `PUT /v1/packages/{pkg}/versions/{ver}/sqlite` |
| `glazed/pkg/help/publish/directory_store.go` | Directory-based package store, `Publish`, `PackageVersionDBPath` |
| `glazed/pkg/help/publish/validation.go` | Package/version validation, `ValidatePackageVersion` |
| `glazed/pkg/help/publish/sqlite_validator.go` | SQLite help DB validation |
| `glazed/cmd/docsctl/publish.go` | `docsctl publish` command |
| `glazed/cmd/docsctl/validate.go` | `docsctl validate` command |
| `glazed/cmd/docsctl/main.go` | `docsctl` CLI entry point |
| `glazed/cmd/docs-registry/main.go` | `docs-registry` server entry point |
| `glazed/web/src/main.tsx` | SPA entry, `HashRouter` → `BrowserRouter` |
| `glazed/web/src/App.tsx` | Root component, routing, package/section state |
| `glazed/web/src/services/api.ts` | RTK Query API slice, `resolveApiBaseUrl`, `resolveRuntimeBaseUrl` |
| `glazed/web/src/store.ts` | Redux store configuration |
| `glazed/web/src/types/index.ts` | TypeScript types mirroring Go response types |
| `glazed/web/src/components/DocumentationTree/tree.ts` | Tree builder for sidebar |
| `glazed/web/public/site-config.js` | Runtime configuration for SPA |
| `pinocchio/pkg/spa/spa.go` | SPA handler with fallback routing |

### External references

- React Router v6 documentation: https://reactrouter.com/en/main/routers/create-browser-router
- RTK Query: https://redux-toolkit.js.org/rtk-query/overview
- Go 1.22 `http.ServeMux` patterns: https://pkg.go.dev/net/http#ServeMux
