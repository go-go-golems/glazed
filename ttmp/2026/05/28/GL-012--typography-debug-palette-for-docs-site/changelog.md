# Changelog

## 2026-05-28

- Initial workspace created


## 2026-05-28

Created GL-012 ticket, wrote full analysis/design/implementation guide (16 sections, 74KB), uploaded to reMarkable at /ai/2026/05/28/GL-012

### Related Files

- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/ttmp/2026/05/28/GL-012--typography-debug-palette-for-docs-site/design/01-typography-debug-palette-analysis-design-implementation-guide.md — Main design document
- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/ttmp/2026/05/28/GL-012--typography-debug-palette-for-docs-site/reference/01-diary.md — Implementation diary


## 2026-05-28

Implemented full Typography Debug Palette: 18 new files, 3 modified files. Features: 13 element groups, 4 presets, custom preset save/delete, localStorage persistence, CSS export to clipboard (rules or variables format), Ctrl+Shift+T shortcut, 𝒜a dev toggle. Commits: 117db61, 2d6700b, 391258c

### Related Files

- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/components/TypographyPalette/TypographyPalette.tsx — Main palette component with preset management
- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/components/TypographyPalette/css-override-engine.ts — CSS generation and clipboard export engine
- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/store/typographyPaletteSlice.ts — Redux slice with localStorage persistence


## 2026-05-28

Added baseline design system, scale mode, and spacing controls. Baseline panel with base font size, 8 scale ratios, line height, letter/word spacing. Custom/Scale toggle per element with scale step selectors. New 'Scale System (1.25)' preset. Letter-spacing and word-spacing CSS overrides. Commit: 8488851

### Related Files

- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/components/TypographyPalette/BaselineParameters.tsx — New baseline parameter controls component
- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/components/TypographyPalette/useTypographyOverrides.ts — Scale-mode resolution logic
- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/src/types/typography-palette.ts — New types: BaselineParameters


## 2026-05-28

Step 5: Add EB Garamond serif font and two serif presets (commit 11b25d9)

### Related Files

- /home/manuel/workspaces/2026-05-28/better-docs-fonts/glazed/web/public/fonts/eb-garamond-latin-400-normal.woff2 — Vendored serif font for editorial presets


## 2026-05-28

Step 6: Add typeface role system (Display/Body/Code) — three dropdowns cascade font to element groups (commit a1c83c0)


## 2026-05-28

Step 8: Fix Display typeface role (was skipped due to adjustable guard) + add Garamond Reading preset (commit 2197b2f)

