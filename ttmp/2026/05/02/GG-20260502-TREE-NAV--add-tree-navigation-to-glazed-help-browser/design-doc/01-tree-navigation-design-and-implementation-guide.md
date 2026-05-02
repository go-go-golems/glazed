---
Title: Tree navigation design and implementation guide
Ticket: GG-20260502-TREE-NAV
Status: active
Topics:
    - glazed
    - help
    - frontend
    - react
    - ui
DocType: design-doc
Intent: long-term
Owners:
    - manuel
RelatedFiles:
    - Path: pkg/help/server/handlers.go
      Note: List endpoint constructs section summaries for tree data.
    - Path: pkg/help/server/types.go
      Note: Go response contract needs SectionHeading metadata.
    - Path: web/src/App.tsx
      Note: Owns sidebar state and needs Tree/Search mode wiring.
    - Path: web/src/components/SectionList/SectionList.tsx
      Note: Existing flat Search-mode list to preserve.
    - Path: web/src/services/api.ts
      Note: RTK Query section list contracts need heading metadata.
    - Path: web/src/types/index.ts
      Note: TypeScript section summary type needs subsection heading fields.
ExternalSources:
    - /tmp/pi-clipboard-f9adfa63-c8b4-464e-b207-454e290c9bbb.png
Summary: Design and implementation guide for adding a tree/search navigation mode to the Glazed help browser.
LastUpdated: 2026-05-02T16:35:00-04:00
WhatFor: Use when implementing the Documentation sidebar tree view and Tree/Search navigation toggle.
WhenToUse: Before changing the Glazed help browser sidebar, navigation state, or section grouping logic.
---


# Tree navigation design and implementation guide

## Executive summary

The current Glazed help browser sidebar is a flat list of section cards. After the recent multi-package work, users can choose a package and optional version, then filter the flat list by search text and section type. The next UI step is to add a tree navigation mode matching the attached screenshot at `/tmp/pi-clipboard-f9adfa63-c8b4-464e-b207-454e290c9bbb.png`.

For the first implementation, the tree does not need to understand source-code APIs, packages, functions, or arbitrary heading hierarchies deeply. The requested initial tree is deliberately simple:

1. Four top-level document type groups:
   - Topics / GeneralTopic
   - Examples / Example
   - Applications / Application
   - Tutorials / Tutorial
2. Under each group, list the documents for the selected package/version.
3. Under each document, list its subsections parsed from Markdown headings in that document content.
4. Add a `Tree` / `Search` segmented toggle in the sidebar.
5. Tree mode shows grouped navigation.
6. Search mode shows the existing card list/search result behavior.

This should be implemented mostly in the React frontend. The current API already fetches section summaries for the selected package/version and can fetch a section detail by slug. However, subsection titles are only present in `content`, which is omitted from section summaries. Therefore there are two reasonable implementation options:

- **Phase 1 frontend-only/simple approach**: build a tree from section summaries and only show document nodes. Add subsection parsing later.
- **Recommended approach for this ticket**: add lightweight heading metadata to the list endpoint so the tree can render document subsections without fetching every full document individually.

This guide recommends the heading-metadata approach because the user explicitly asked for “documents and their subsections,” and because fetching every detail document just to build the navigation tree would be wasteful and harder to reason about.

## Problem statement

### Current behavior

The sidebar currently shows:

- title bar `📁 Sections`,
- search input,
- package/version selectors,
- type filter buttons,
- flat `SectionList`,
- status bar.

This is implemented in `web/src/App.tsx`. The app state and data loading are concentrated at lines 16-80:

- `search` and `filter` state at `web/src/App.tsx:17-18`.
- selected package/version state at `web/src/App.tsx:19-20`.
- package list query at `web/src/App.tsx:29`.
- section list query at `web/src/App.tsx:51-53`.
- detail query at `web/src/App.tsx:54-60`.
- client-side search/type filtering at `web/src/App.tsx:67-80`.

The sidebar render path is at `web/src/App.tsx:84-107`. It always renders `SectionList` at `web/src/App.tsx:101-105`; there is no navigation mode switch.

`SectionList` itself is simple. It maps section summaries to `SectionCard` components at `web/src/components/SectionList/SectionList.tsx:13-25`. `SectionCard` renders a type badge, optional top-level badge, title, and short description at `web/src/components/SectionList/SectionCard.tsx:14-30`.

### Desired behavior

The new design in the screenshot changes the left pane from “Sections” to “Documentation” and introduces a navigation area with a mode toggle:

```text
Documentation

Package  [ Gepetto v ]
Version  [ v5.1.0  v ]

Search documentation...        ⌘K

Navigation        [ Tree | Search ]

▾ 📖 Topics
  📄 What is Gepetto?
    # Purpose
    # Install
  📄 Quick Start
▾ 📁 Examples
  📄 Streaming CLI
▾ 📦 Applications
▾ 🎓 Tutorials
```

For now the top-level tree categories should be the four Glazed document types. The screenshot shows richer categories such as `Introduction`, `Guides`, and `API Reference`, but the user explicitly scoped this first pass to the four document types plus documents and subsections.

### Success criteria

A successful implementation should let a user:

1. Select a package/version as before.
2. Use the sidebar toggle to switch between `Tree` and `Search` modes.
3. In `Tree` mode, see four type groups even if some are empty or, preferably, only groups with documents.
4. Expand/collapse document type groups.
5. Expand/collapse document nodes.
6. Click a document node to navigate to that document.
7. Click a subsection node to navigate to the document and scroll to or highlight that subsection.
8. Switch to `Search` mode and still use the current flat search/card list behavior.
9. Preserve the Classic Mac visual style.

## Current-state architecture

### Frontend data flow

The current RTK Query API slice is `web/src/services/api.ts`.

Important current endpoints:

```ts
listPackages: builder.query<ListPackagesResponse, void>({...})
listSections: builder.query<ListSectionsResponse, ListSectionsQueryArgs | void>({...})
getSection: builder.query<SectionDetail, GetSectionQueryArgs>({...})
```

These are defined at `web/src/services/api.ts:83-126`. `listSections` supports package/version query parameters at `web/src/services/api.ts:90-100`. `getSection` supports package/version query parameters at `web/src/services/api.ts:111-122`.

Current section summary type is `web/src/types/index.ts:5-17`:

```ts
export interface SectionSummary {
  id: number;
  packageName?: string;
  packageVersion?: string;
  slug: string;
  type: string;
  title: string;
  short: string;
  topics: string[];
  isTopLevel: boolean;
}
```

Current detail type extends the summary and adds content at `web/src/types/index.ts:19-25`.

Important implication: the flat list only has summary metadata. The tree needs subsection headings, which are only in Markdown content unless the server exposes them as metadata.

### Backend response contracts

Server response types live in `pkg/help/server/types.go`:

- `SectionSummary` currently omits content to keep list responses small.
- `SectionDetail` includes `Content`.
- `ListSectionsResponse` returns only `[]SectionSummary`.

The tree feature should avoid embedding full Markdown `content` into every list response. Instead it can expose a compact heading list.

### Layout and styling

The app uses a classic two-pane layout:

- `AppLayout` renders sidebar and content at `web/src/components/AppLayout/AppLayout.tsx:11-17`.
- Sidebar width is fixed at 280px in `web/src/components/AppLayout/styles/app-layout.css:11-20`.
- Existing section-list styling is in `web/src/components/SectionList/styles/section-list.css`.

The tree will need its own styles, but should follow the same data-part CSS pattern used by other components. For example, `PackageSelector` uses:

- component file: `web/src/components/PackageSelector/PackageSelector.tsx`
- parts file: `web/src/components/PackageSelector/parts.ts`
- styles file: `web/src/components/PackageSelector/styles/package-selector.css`

The tree should follow this same pattern.

## Proposed UX and interaction model

### Sidebar structure

Replace the sidebar title `📁 Sections` with `📖 Documentation` or `📚 Documentation`.

Recommended sidebar order:

1. Title bar.
2. Package selector.
3. Version selector, only if selected package has versions.
4. Search input.
5. Navigation label and mode toggle.
6. If mode is `Tree`, render `DocumentationTree`.
7. If mode is `Search`, render the existing type filters and flat `SectionList`.
8. Status bar.

Pseudocode render shape:

```tsx
<TitleBar title="📖 Documentation" />
<div className="sidebar-controls">
  <PackageSelector ... />
  <SearchBar
    value={search}
    onChange={setSearch}
    placeholder="Search documentation…"
  />
  <NavigationModeToggle value={mode} onChange={setMode} />
</div>
{mode === 'tree' ? (
  <DocumentationTree
    sections={treeSections}
    activeSlug={activeSlug}
    activeHeadingId={activeHeadingId}
    onSelectDocument={handleSelect}
    onSelectHeading={handleSelectHeading}
  />
) : (
  <>
    <TypeFilter value={filter} onChange={setFilter} />
    <SectionList sections={filtered} ... />
  </>
)}
<StatusBar count={...} />
```

### Navigation mode toggle

Add a small segmented control that looks like the screenshot:

```text
Navigation        [ Tree | Search ]
```

Behavior:

- `Tree` selected by default.
- `Search` switches to existing flat list behavior.
- The search input remains visible in both modes.
- In Tree mode, search filters visible documents and subsection labels inside the tree.
- In Search mode, search uses the current filtering behavior.

Suggested state:

```ts
type NavigationMode = 'tree' | 'search';
const [navigationMode, setNavigationMode] = useState<NavigationMode>('tree');
```

### Tree grouping

Initial tree top-level groups should map the four Glazed section types:

| Section type | Display label | Icon suggestion |
| --- | --- | --- |
| `GeneralTopic` | `Topics` | `📖` |
| `Example` | `Examples` | `📄` |
| `Application` | `Applications` | `📦` |
| `Tutorial` | `Tutorials` | `🎓` |

The screenshot uses folder/page icons and disclosure triangles. Unicode icons are acceptable for the first pass because existing UI already uses emoji in title bars. If visual consistency suffers, replace emoji with CSS-drawn icons later.

### Document nodes

A document node represents one `SectionSummary`.

Display:

```text
▸ 📄 Events, Streaming, and Watermill
```

Active state:

- If active document slug matches the node slug, render selected text in blue or inverted Classic Mac style.
- The screenshot shows the selected document as blue text with a yellow document icon, but exact colors can follow the app's current black/white selection styling if simpler.

Click behavior:

- Clicking a document navigates to `/sections/:slug` using the existing `handleSelect`.
- If a package/version is selected, the existing `getSection` query uses selected package/version, so no URL change is required in this first pass.

Important caveat: current routes do not include package/version. This means a copied URL only stores the slug, not the package/version context. This was already true after the multi-package work. Tree navigation can preserve the same limitation for now.

### Subsection nodes

A subsection node is parsed from Markdown headings in a document.

Display:

```text
  # First heading
  ## Nested heading
```

Better first-pass display:

```text
  § First heading
  § Nested heading
```

Click behavior:

- Navigate to the document.
- Add a hash fragment for the heading slug if possible:

```ts
navigate(`/sections/${section.slug}#${heading.id}`)
```

- Update `SectionView`/`MarkdownContent` so headings receive stable IDs.
- On route/hash changes, browser default hash scrolling should work if the element exists.

If stable heading IDs are not immediately implemented, subsection clicks can initially navigate to the document only. However, the implementation task should include stable heading IDs because otherwise subsection nodes are visually useful but not functionally useful.

## Proposed API changes

### Add heading metadata to section summaries

Extend Go server `SectionSummary` with optional headings:

```go
type SectionHeading struct {
    ID    string `json:"id"`
    Level int    `json:"level"`
    Text  string `json:"text"`
}

type SectionSummary struct {
    ID             int64            `json:"id"`
    PackageName    string           `json:"packageName"`
    PackageVersion string           `json:"packageVersion,omitempty"`
    Slug           string           `json:"slug"`
    Type           string           `json:"type"`
    Title          string           `json:"title"`
    Short          string           `json:"short"`
    Topics         []string         `json:"topics"`
    IsTopLevel     bool             `json:"isTopLevel"`
    Headings       []SectionHeading `json:"headings,omitempty"`
}
```

Add matching TypeScript type:

```ts
export interface SectionHeading {
  id: string;
  level: number;
  text: string;
}

export interface SectionSummary {
  // existing fields...
  headings?: SectionHeading[];
}
```

### Heading parser

Implement a small Markdown heading extractor in Go or TypeScript. Because the list endpoint already has access to `model.Section.Content`, parsing headings server-side keeps the frontend simple and avoids fetching every detail document.

Rules:

- Only parse ATX headings (`#`, `##`, `###`, etc.) initially.
- Ignore fenced code blocks.
- Include levels 2 through 4 by default, or levels 1 through 4 if content conventions need it.
- Ignore the first heading if it exactly equals the section title, to avoid duplicate document/document heading nodes.
- Generate stable IDs matching frontend heading IDs.

Pseudocode:

```go
func ExtractHeadings(markdown string, sectionTitle string) []SectionHeading {
    inFence := false
    var headings []SectionHeading
    scanner := bufio.NewScanner(strings.NewReader(markdown))
    for scanner.Scan() {
        line := scanner.Text()
        trimmed := strings.TrimSpace(line)
        if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
            inFence = !inFence
            continue
        }
        if inFence || !strings.HasPrefix(trimmed, "#") {
            continue
        }
        level := countLeadingHashes(trimmed)
        if level < 1 || level > 4 || !strings.HasPrefix(trimmed[level:], " ") {
            continue
        }
        text := strings.TrimSpace(trimmed[level:])
        text = strings.Trim(text, "# ")
        if text == "" || strings.EqualFold(text, sectionTitle) {
            continue
        }
        headings = append(headings, SectionHeading{
            ID: slugifyHeading(text),
            Level: level,
            Text: text,
        })
    }
    return headings
}
```

### Stable heading IDs in rendered Markdown

The frontend currently renders Markdown through `MarkdownContent` / `SectionView`. The implementation should inspect `web/src/components/Markdown/MarkdownContent.tsx` and add custom heading renderers that apply the same slugify algorithm used by the server.

Pseudocode:

```tsx
const components = {
  h1: Heading('h1'),
  h2: Heading('h2'),
  h3: Heading('h3'),
  h4: Heading('h4'),
};

function Heading(tag: keyof JSX.IntrinsicElements) {
  return function HeadingRenderer({ children, ...props }) {
    const text = extractText(children);
    const id = slugifyHeading(text);
    return React.createElement(tag, { id, ...props }, children);
  };
}
```

To avoid server/frontend slug drift, either:

1. duplicate a simple slug algorithm in Go and TS with tests, or
2. have the tree use the server-generated heading ID and have frontend heading renderer use an equivalent tested implementation.

## Proposed frontend component structure

Add these files:

```text
web/src/components/NavigationModeToggle/NavigationModeToggle.tsx
web/src/components/NavigationModeToggle/parts.ts
web/src/components/NavigationModeToggle/styles/navigation-mode-toggle.css

web/src/components/DocumentationTree/DocumentationTree.tsx
web/src/components/DocumentationTree/parts.ts
web/src/components/DocumentationTree/styles/documentation-tree.css
web/src/components/DocumentationTree/tree.ts
web/src/components/DocumentationTree/tree.test.ts
```

### Tree data model

Recommended TypeScript types:

```ts
type DocumentationTreeGroup = {
  id: SectionType;
  label: string;
  icon: string;
  sections: DocumentationTreeSection[];
};

type DocumentationTreeSection = {
  slug: string;
  title: string;
  type: SectionType;
  headings: SectionHeading[];
};
```

Build function:

```ts
export function buildDocumentationTree(sections: SectionSummary[]): DocumentationTreeGroup[] {
  const groups = makeTypeGroups();
  for (const section of sections) {
    const group = groups.get(section.type as SectionType);
    if (!group) continue;
    group.sections.push({
      slug: section.slug,
      title: section.title,
      type: section.type as SectionType,
      headings: section.headings ?? [],
    });
  }
  return GROUP_ORDER
    .map((type) => groups.get(type)!)
    .filter((group) => group.sections.length > 0);
}
```

Search filtering:

```ts
export function filterDocumentationTree(groups, query) {
  if (!query.trim()) return groups;
  const q = query.toLowerCase();
  return groups
    .map(group => ({
      ...group,
      sections: group.sections
        .map(section => ({
          ...section,
          headings: section.headings.filter(h => h.text.toLowerCase().includes(q)),
        }))
        .filter(section =>
          section.title.toLowerCase().includes(q) || section.headings.length > 0
        ),
    }))
    .filter(group => group.sections.length > 0);
}
```

### Expanded/collapsed state

Use local state in `DocumentationTree`:

```ts
const [expandedGroups, setExpandedGroups] = useState<Set<string>>(() => new Set(defaultGroupIds));
const [expandedSections, setExpandedSections] = useState<Set<string>>(() => new Set());
```

Default behavior:

- All non-empty type groups expanded initially.
- Active document expanded automatically.
- Document nodes collapsed initially unless active.

Implementation detail: React state should not mutate `Set` in place. Always copy:

```ts
setExpandedGroups(prev => {
  const next = new Set(prev);
  next.has(id) ? next.delete(id) : next.add(id);
  return next;
});
```

### Accessibility

Minimum accessibility requirements:

- Tree root uses `role="tree"`.
- Group/document nodes use `role="treeitem"`.
- Expanded nodes set `aria-expanded`.
- Active document uses `aria-current="page"` or `aria-selected="true"`.
- Buttons have text labels; do not rely only on icons.

Keyboard navigation can be deferred if necessary, but the component should be structured so it can be added later.

## Implementation phases

### Phase 1 — Backend heading metadata

Files:

- `pkg/help/server/types.go`
- `pkg/help/server/handlers.go`
- new `pkg/help/server/headings.go`
- new `pkg/help/server/headings_test.go`
- `pkg/help/server/server_test.go`

Steps:

1. Add `SectionHeading` response type.
2. Add `Headings []SectionHeading` to `SectionSummary`.
3. Implement `ExtractHeadings(content, title string) []SectionHeading`.
4. Update `SummaryFromModel` to populate headings.
5. Add tests for:
   - `#`, `##`, `###` headings,
   - fenced code blocks ignored,
   - duplicate title heading ignored,
   - stable slug generation,
   - list endpoint includes heading metadata.

### Phase 2 — Frontend types and heading IDs

Files:

- `web/src/types/index.ts`
- `web/src/components/Markdown/MarkdownContent.tsx`
- optional `web/src/utils/slugify.ts`
- tests if utilities are present

Steps:

1. Add `SectionHeading` type.
2. Add optional `headings?: SectionHeading[]` to `SectionSummary`.
3. Add heading ID rendering to Markdown headings.
4. Keep the slug algorithm aligned with the backend.
5. Verify hash navigation works by manually opening `/sections/<slug>#<heading-id>`.

### Phase 3 — Navigation mode toggle

Files:

- `web/src/components/NavigationModeToggle/*`
- `web/src/App.tsx`
- `web/src/App.test.tsx`

Steps:

1. Add `NavigationMode` type and component.
2. Add `navigationMode` state to `App`.
3. Render toggle next to `Navigation` label.
4. Default to `tree` mode.
5. In `search` mode, render existing `TypeFilter` and `SectionList` unchanged.

### Phase 4 — Documentation tree component

Files:

- `web/src/components/DocumentationTree/*`
- `web/src/App.tsx`
- `web/src/App.test.tsx`

Steps:

1. Add tree builder utility with deterministic type group order.
2. Add search filtering for tree labels.
3. Add `DocumentationTree` component.
4. Wire document click to existing `handleSelect`.
5. Wire heading click to a new `handleSelectHeading(slug, headingId)`.
6. Ensure active document is expanded.
7. Style to match screenshot: disclosure triangles, small icons, indentation, selected row.

### Phase 5 — Route/hash behavior

Files:

- `web/src/App.tsx`
- `web/src/components/SectionView/SectionView.tsx`
- `web/src/components/Markdown/MarkdownContent.tsx`

Steps:

1. Parse hash from `location.hash` or `window.location.hash` depending on current router setup.
2. When selecting a heading, navigate to `/sections/${slug}#${headingId}`.
3. After `SectionView` renders, scroll the heading into view if browser default hash scrolling is not sufficient.
4. Add regression test if practical.

### Phase 6 — Visual and manual validation

Commands:

```bash
cd glazed/web
pnpm test -- --run
pnpm exec tsc --noEmit
pnpm build

cd ..
go test ./pkg/help/...
GOWORK=off go generate ./pkg/web
```

Smoke server:

```bash
go run ./cmd/glaze serve --from-sqlite-dir /tmp/glazed-multi-help-smoke --address :8099
```

Manual checks:

- Package selector still works.
- Version selector still hides for unversioned packages.
- Tree mode is default.
- Search mode still shows flat cards.
- Four document types group documents correctly.
- Clicking document nodes navigates to the document.
- Clicking subsection nodes navigates to the document heading.
- Search input filters tree documents/subsections in tree mode.
- No 500 errors from `/api/packages`, `/api/sections`, or `/api/sections?package=...`.

## Risks and tradeoffs

### Risk: heading IDs differ between backend and frontend

If Go and TypeScript slugify headings differently, tree subsection links will not scroll. Mitigate with simple slug rules and tests on both sides.

### Risk: list endpoint becomes too heavy

Adding headings to summaries increases payload size, but headings are much smaller than full content. For the current help corpus size this is acceptable. If it becomes too large later, add a query parameter such as `?includeHeadings=true`.

### Risk: tree mode makes search semantics confusing

The screenshot shows both a search input and a `Tree/Search` toggle. The recommended behavior is:

- Search input always captures text.
- Tree mode filters visible tree nodes.
- Search mode shows card-style search results.

This should be explained in tests and maybe in placeholder text.

### Risk: package/version not represented in route

The current route is still `/sections/:slug`, while selected package/version are app state. This pre-existing limitation means deep links may resolve differently if the user changes package. Tree implementation can defer fixing this, but should avoid making it worse.

## File reference map

Primary frontend files:

- `web/src/App.tsx` — top-level state, queries, sidebar render, route navigation.
- `web/src/services/api.ts` — RTK Query endpoints for packages, section lists, and section detail.
- `web/src/types/index.ts` — TypeScript API contracts.
- `web/src/components/SectionList/SectionList.tsx` — existing flat list.
- `web/src/components/SectionList/SectionCard.tsx` — existing card display.
- `web/src/components/PackageSelector/PackageSelector.tsx` — component pattern to follow for new sidebar controls.
- `web/src/components/SearchBar/SearchBar.tsx` — search input component.
- `web/src/components/Markdown/MarkdownContent.tsx` — Markdown rendering and future heading IDs.

Primary backend files:

- `pkg/help/server/types.go` — API response contracts.
- `pkg/help/server/handlers.go` — response construction and list endpoint.
- `pkg/help/model/section.go` — section content and metadata model.

Reference artifact:

- `/tmp/pi-clipboard-f9adfa63-c8b4-464e-b207-454e290c9bbb.png` — target tree/search navigation UI sketch.

## Open questions

1. Should empty document type groups be shown, or should only non-empty groups be visible?
2. Should tree mode use the same search input or should search input only affect Search mode?
3. Should heading levels be capped at `h3` or `h4` for readability?
4. Should route URLs be upgraded to include package/version while implementing tree heading hashes?
5. Should the backend expose headings always, or only when `includeHeadings=true` is passed?
