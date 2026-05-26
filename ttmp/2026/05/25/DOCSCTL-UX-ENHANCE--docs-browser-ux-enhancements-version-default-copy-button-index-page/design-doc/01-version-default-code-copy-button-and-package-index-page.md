---
Title: ""
Ticket: ""
Status: ""
Topics: []
DocType: ""
Intent: ""
Owners: []
RelatedFiles:
    - Path: glazed/pkg/help/server/handlers.go
      Note: handleListPackages version sorting
    - Path: glazed/pkg/help/server/types.go
      Note: PackageSummary and ListPackagesResponse types
    - Path: glazed/web/src/App.tsx
      Note: handlePackageChange
    - Path: glazed/web/src/components/DocumentationTree/tree.ts
      Note: buildDocumentationTree grouping by type
    - Path: glazed/web/src/components/EmptyState/EmptyState.tsx
      Note: Current placeholder to replace
    - Path: glazed/web/src/components/Markdown/MarkdownContent.tsx
      Note: Markdown renderer with custom components
    - Path: glazed/web/src/components/Markdown/styles/markdown.css
      Note: Styles for pre/code blocks
    - Path: glazed/web/src/types/index.ts
      Note: PackageSummary TypeScript type
ExternalSources: []
Summary: ""
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---


# Version Default, Code Copy Button, and Package Index Page

## 1. Executive Summary

Three independent UX improvements for the docs browser SPA:

1. **Latest version selection** — When the user switches to a different package in the dropdown, the version selector should default to the latest version (semver-highest or most recently published). The current code already picks `versions[0]`, which the Go API returns in reverse-sorted order, so this works for semver-style versions. But there is no explicit "latest" semantic — if the API ever reorders, the frontend would break. We should verify the contract and add a defensive comment.

2. **Copy-to-clipboard on code blocks** — Every fenced code block in the rendered Markdown should have a small "copy" button (typically top-right corner) that copies the code content to the clipboard. This is a standard docs-site feature that all developer documentation tools provide.

3. **Package index page** — When the user navigates to `/{package}/{version}` without a section slug, the content area currently shows a generic "Select a section from the list." empty state. Instead, it should show a structured index of all sections in that package/version, grouped by type (Topics, Examples, Applications, Tutorials), with links to each section.

---

## 2. Problem Statement

### 2.1 Version selection

**Current behavior**: When the user changes the package dropdown, `handlePackageChange` picks `nextVersions[0]`:

```tsx
// App.tsx line ~92
const handlePackageChange = (value: string) => {
    const nextPackage = packages.find((pkg) => pkg.name === value);
    const nextVersions = nextPackage?.versions ?? [];
    const newVersion = nextVersions[0] || '';
    navigate(`/${value}/${versionToUrl(newVersion)}`);
};
```

The Go API sorts versions in reverse order (`sort.Reverse(sort.StringSlice(...))` in `handlers.go:147`), so `versions[0]` is the lexicographically highest version string. For semver versions like `v1.2.15`, this happens to be the latest. But:

- Lexicographic sort is not semver sort. `v1.10.0` sorts before `v1.9.0` lexicographically (because `"1.10" > "1.9"` string-compares `"1.1" < "1.9"`). Wait — actually `sort.Reverse(sort.StringSlice(...))` sorts descending, so `"v1.9.0"` comes before `"v1.10.0"` because `"9" > "1"` in string comparison. For standard semver with `v` prefix and same-length segments, this usually works. For mixed-length segments, it doesn't.

- The `ListPackagesResponse` has a `defaultVersion` field that the server already computes as `packages[0].Versions[0]`. The frontend already uses this for the initial redirect but not for package changes.

**What needs fixing**: The version-selection logic is fragile but functionally correct for the current data. The fix is to add a comment documenting the assumption, and optionally add a `latestVersion` field to the API response that the server computes with proper semver sorting.

### 2.2 Copy button on code blocks

**Current behavior**: The `MarkdownContent` component renders fenced code blocks as `<pre><code>...</code></pre>` with basic styling (background, border, monospace font). There is no copy button.

**File**: `web/src/components/Markdown/MarkdownContent.tsx`

The component uses `react-markdown` with `remarkGfm`. It defines custom components for headings (`h1`–`h4`) but uses the default renderer for `pre` and `code`. The CSS (`markdown.css`) styles `pre` blocks with padding, background, and a left border.

**What's needed**: Override the `pre` component in `MarkdownContent` to wrap it in a container that includes a copy button positioned in the top-right corner. The button should copy the text content of the `<code>` child to the clipboard using `navigator.clipboard.writeText()`.

### 2.3 Package index page

**Current behavior**: When the URL is `/{package}/{version}` (no slug), the `EmptyState` component shows "Select a section from the list." This is unhelpful — the user has already selected a package and version but gets no overview of what's available.

**File**: `web/src/App.tsx` renders `<EmptyState />` when `!section`.

**What's needed**: Replace the `EmptyState` with a `PackageIndex` component that displays all sections for the selected package/version, grouped by type, with links. This is essentially the same data as the sidebar tree but presented as a browsable list in the main content area.

---

## 3. Proposed Architecture

### 3.1 Version selection — document + harden

**No code change needed for the basic case.** The current logic already picks the first version from the API's reverse-sorted list. What we should do:

1. Add a code comment in `App.tsx` documenting the assumption that the API returns versions in reverse-sorted (latest-first) order.
2. Add a `latestVersion` field to the `ListPackagesResponse` Go type and the `PackageSummary` type, computed explicitly by the server. This makes the contract explicit rather than relying on array ordering.

**Go server change** (`pkg/help/server/types.go` + `handlers.go`):

```go
// In PackageSummary, add:
LatestVersion string   `json:"latestVersion,omitempty"`

// In handleListPackages, after building pkg:
if len(pkg.Versions) > 0 {
    pkg.LatestVersion = pkg.Versions[0]  // Already reverse-sorted
}
```

**Frontend change** (`App.tsx`):

```tsx
const handlePackageChange = (value: string) => {
    const nextPackage = packages.find((pkg) => pkg.name === value);
    // Prefer explicit latestVersion from API; fall back to versions[0]
    const newVersion = nextPackage?.latestVersion
        || nextPackage?.versions?.[0]
        || '';
    navigate(`/${value}/${versionToUrl(newVersion)}`);
};
```

### 3.2 Copy button on code blocks

Create a `CodeBlock` component that wraps the default `<pre>` rendering with a copy button.

**Component**: `web/src/components/Markdown/CodeBlock.tsx`

```tsx
// Pseudocode for CodeBlock
import { useState, useRef, type ComponentPropsWithoutRef, type ReactNode } from 'react';

interface CodeBlockProps extends ComponentPropsWithoutRef<'pre'> {
  children?: ReactNode;
}

export function CodeBlock({ children, ...props }: CodeBlockProps) {
  const [copied, setCopied] = useState(false);
  const preRef = useRef<HTMLPreElement>(null);

  const handleCopy = async () => {
    // Extract text from the <code> child inside the <pre>
    const codeEl = preRef.current?.querySelector('code');
    const text = codeEl?.textContent ?? preRef.current?.textContent ?? '';
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div style={{ position: 'relative' }}>
      <pre ref={preRef} {...props}>{children}</pre>
      <button
        onClick={handleCopy}
        aria-label="Copy code"
        style={{
          position: 'absolute',
          top: 8,
          right: 8,
          // ... styling: small, subtle, hover-visible
        }}
      >
        {copied ? '✓' : '📋'}
      </button>
    </div>
  );
}
```

**Integration**: In `MarkdownContent.tsx`, add `pre: CodeBlock` to the `markdownComponents` object:

```tsx
const markdownComponents = {
  h1: createHeading('h1', seenIDs),
  h2: createHeading('h2', seenIDs),
  h3: createHeading('h3', seenIDs),
  h4: createHeading('h4', seenIDs),
  pre: CodeBlock,  // <-- new
};
```

**CSS**: Add styles for the copy button in `markdown.css`:

```css
[data-part='markdown-content'] .code-block-wrapper {
  position: relative;
}

[data-part='markdown-content'] .code-block-wrapper pre {
  /* existing pre styles remain */
}

[data-part='markdown-content'] .copy-button {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(255,255,255,0.8);
  border: 1px solid #ccc;
  border-radius: 3px;
  padding: 2px 6px;
  font-size: 12px;
  cursor: pointer;
  opacity: 0;
  transition: opacity 0.2s;
}

[data-part='markdown-content'] .code-block-wrapper:hover .copy-button {
  opacity: 1;
}
```

### 3.3 Package index page

Create a `PackageIndex` component that renders when no section is selected.

**Component**: `web/src/components/PackageIndex/PackageIndex.tsx`

```tsx
// Pseudocode for PackageIndex
import type { SectionSummary } from '../../types';

interface PackageIndexProps {
  sections: SectionSummary[];
  onSelect: (slug: string) => void;
}

// Group sections by type, same categories as DocumentationTree
const GROUPS = [
  { type: 'GeneralTopic', label: 'Topics', icon: '📖' },
  { type: 'Example', label: 'Examples', icon: '📄' },
  { type: 'Application', label: 'Applications', icon: '📦' },
  { type: 'Tutorial', label: 'Tutorials', icon: '🎓' },
];

export function PackageIndex({ sections, onSelect }: PackageIndexProps) {
  // Group sections by type
  const grouped = GROUPS.map(group => ({
    ...group,
    sections: sections.filter(s => s.type === group.type),
  })).filter(g => g.sections.length > 0);

  return (
    <div>
      <h1>Documentation Index</h1>
      <p>{sections.length} sections available</p>
      {grouped.map(group => (
        <section key={group.type}>
          <h2>{group.icon} {group.label}</h2>
          <ul>
            {group.sections.map(section => (
              <li key={section.slug}>
                <a href="#" onClick={(e) => { e.preventDefault(); onSelect(section.slug); }}>
                  {section.title}
                </a>
                {section.short && <p>{section.short}</p>}
              </li>
            ))}
          </ul>
        </section>
      ))}
    </div>
  );
}
```

**Integration**: In `App.tsx`, replace `<EmptyState />` with `<PackageIndex />`:

```tsx
{!isLoading && !error && !section && (
  <PackageIndex
    sections={listData?.sections ?? []}
    onSelect={handleSelect}
  />
)}
```

The `EmptyState` component is still available for the true "nothing loaded" case (e.g., when `listData` is undefined), but the normal "no slug selected" case shows the index.

---

## 4. Implementation Plan

### Phase 1: Version selection hardening (small)

1. Add `latestVersion` field to `PackageSummary` in `pkg/help/server/types.go`
2. Populate it in `handleListPackages` in `pkg/help/server/handlers.go`
3. Update TypeScript type `PackageSummary` in `web/src/types/index.ts`
4. Use `latestVersion` in `App.tsx` `handlePackageChange`
5. Add comment documenting the version-sorting assumption
6. Add test for `latestVersion` in Go server tests

### Phase 2: Copy button on code blocks (medium)

1. Create `web/src/components/Markdown/CodeBlock.tsx`
2. Add `pre: CodeBlock` to `MarkdownContent.tsx` components
3. Add CSS for the copy button in `markdown.css`
4. Add Storybook story for `CodeBlock` (optional)
5. Test: verify copy button appears on code blocks, click copies to clipboard, "✓" feedback

### Phase 3: Package index page (medium)

1. Create `web/src/components/PackageIndex/PackageIndex.tsx`
2. Create `web/src/components/PackageIndex/parts.ts`
3. Create `web/src/components/PackageIndex/styles/package-index.css`
4. Wire into `App.tsx` (replace `EmptyState` with `PackageIndex` when sections are loaded)
5. Keep `EmptyState` as fallback when `listData` is undefined
6. Test: navigate to `/{pkg}/{ver}`, see index of all sections grouped by type

---

## 5. Testing

### Manual

- [ ] Switch packages in dropdown → version selector shows latest version
- [ ] View a section with code blocks → copy button visible on hover
- [ ] Click copy button → code copied to clipboard, "✓" feedback for 2 seconds
- [ ] Navigate to `/{pkg}/{ver}` → see index page with grouped sections
- [ ] Click a section in the index → navigates to that section

### Automated

- [ ] Go test: `latestVersion` field populated in `/api/packages` response
- [ ] Frontend test: `CodeBlock` renders button, click triggers clipboard write
- [ ] Frontend test: `PackageIndex` renders grouped sections with links

---

## 6. Risks and Open Questions

| # | Question | Recommendation |
|---|---|---|
| 1 | Should copy button use `navigator.clipboard.writeText` or a fallback for older browsers? | Use `navigator.clipboard.writeText` with a try/catch that silently fails. The docs browser targets modern browsers. |
| 2 | Should the PackageIndex show section headings (subsections) too? | No — keep it simple for v1. Just title + short description. Headings are in the sidebar tree. |
| 3 | Should the PackageIndex replace the sidebar tree or complement it? | Complement. The sidebar is for navigation; the index is for discovery/overview in the main content area. |
| 4 | Does the API need a proper semver sort instead of string sort? | Not for v1. The current lexicographic reverse sort works for standard `vX.Y.Z` versions. Add a comment noting the limitation. |

---

## 7. References

| File | Role |
|---|---|
| `glazed/pkg/help/server/handlers.go` | `handleListPackages` — version sorting, `defaultVersion` |
| `glazed/pkg/help/server/types.go` | `PackageSummary`, `ListPackagesResponse` types |
| `glazed/web/src/App.tsx` | `handlePackageChange`, section rendering, `EmptyState` |
| `glazed/web/src/types/index.ts` | `PackageSummary` TypeScript type |
| `glazed/web/src/components/Markdown/MarkdownContent.tsx` | Markdown renderer with custom heading components |
| `glazed/web/src/components/Markdown/styles/markdown.css` | Styles for `pre`, `code`, headings |
| `glazed/web/src/components/SectionView/SectionView.tsx` | Section header + Markdown body |
| `glazed/web/src/components/SectionView/SectionHeader.tsx` | Title, slug, badges |
| `glazed/web/src/components/EmptyState/EmptyState.tsx` | Current "select a section" placeholder |
| `glazed/web/src/components/DocumentationTree/tree.ts` | `buildDocumentationTree` — grouping by type |
