// components/TypographyPalette/element-registry.ts
// Defines every adjustable typography element in the docs browser,
// organized into accordion groups. Elements that participate in the
// design system baseline have `supportsScale: true` with default steps.

import type { TypographyGroup } from '../../types/typography-palette';

export const TYPOGRAPHY_GROUPS: TypographyGroup[] = [
  {
    id: 'root',
    label: 'Root / Body',
    elements: [
      {
        id: 'root.body',
        label: 'Body Text',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'lineHeight', 'letterSpacing', 'wordSpacing'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000', lineHeight: 1.6, letterSpacing: 0, wordSpacing: 0 },
        selector: '.app-root',
        supportsScale: true,
        defaultFontSizeStep: 0,
        defaultLineHeightStep: 0,
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'titlebar',
    label: 'Title Bar',
    elements: [
      {
        id: 'titlebar.title',
        label: 'Title Text',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='titlebar-title']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'display',
      },
    ],
  },
  {
    id: 'menubar',
    label: 'Menu Bar',
    elements: [
      {
        id: 'menubar.items',
        label: 'Menu Items',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='menubar']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'display',
      },
      {
        id: 'menubar.appname',
        label: 'App Name',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='menubar-title']",
        supportsScale: true,
        defaultFontSizeStep: -2,
        typefaceRole: 'display',
      },
    ],
  },
  {
    id: 'sidebar-controls',
    label: 'Sidebar Controls',
    elements: [
      {
        id: 'sidebar.search',
        label: 'Search Input',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='searchbar-input']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'display',
      },
      {
        id: 'sidebar.packageselector',
        label: 'Package Selector',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='package-selector-root']",
        supportsScale: true,
        defaultFontSizeStep: 0,
        typefaceRole: 'display',
      },
      {
        id: 'sidebar.navtoggle',
        label: 'Nav Mode Toggle',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='navigation-mode-toggle-root']",
        supportsScale: true,
        defaultFontSizeStep: 0,
        typefaceRole: 'display',
      },
      {
        id: 'sidebar.typefilter',
        label: 'Type Filter',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='typefilter-button']",
        supportsScale: true,
        defaultFontSizeStep: -3,
        typefaceRole: 'display',
      },
    ],
  },
  {
    id: 'sidebar-tree',
    label: 'Sidebar Tree',
    elements: [
      {
        id: 'tree.row',
        label: 'Document Row',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', color: '#111' },
        selector: "[data-part='documentation-tree-row']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'body',
      },
      {
        id: 'tree.heading',
        label: 'Heading Row',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', color: '#3f4b5a' },
        selector: "[data-part='documentation-tree-row'][data-kind='heading']",
        supportsScale: true,
        defaultFontSizeStep: -2,
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'sidebar-cards',
    label: 'Sidebar Cards',
    elements: [
      {
        id: 'cards.title',
        label: 'Card Title',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part~='section-card-title']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'body',
      },
      {
        id: 'cards.short',
        label: 'Card Description',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', color: '#777' },
        selector: "[data-part~='section-card-short']",
        supportsScale: true,
        defaultFontSizeStep: -3,
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'section-header',
    label: 'Section Header',
    elements: [
      {
        id: 'header.slug',
        label: 'Slug Label',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#999' },
        selector: "[data-part='section-header-slug']",
        supportsScale: true,
        defaultFontSizeStep: -3,
        typefaceRole: 'code',
      },
      {
        id: 'header.heading',
        label: 'Heading',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 24, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='section-header-heading']",
        supportsScale: true,
        defaultFontSizeStep: 4,
        typefaceRole: 'display',
      },
      {
        id: 'header.subtitle',
        label: 'Subtitle',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#555' },
        selector: "[data-part='section-header-subtitle']",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'markdown-prose',
    label: 'Markdown Prose',
    elements: [
      {
        id: 'prose.body',
        label: 'Body Text',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'lineHeight', 'letterSpacing', 'wordSpacing'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000', lineHeight: 1.6, letterSpacing: 0, wordSpacing: 0 },
        selector: "[data-part='markdown-content']",
        supportsScale: true,
        defaultFontSizeStep: 0,
        defaultLineHeightStep: 0,
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'markdown-headings',
    label: 'Markdown Headings',
    elements: [
      {
        id: 'headings.h1',
        label: 'H1',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'letterSpacing'],
        defaults: { fontSize: 1.6, fontSizeUnit: 'em', fontWeight: 700, color: '#000', letterSpacing: 0 },
        selector: "[data-part='markdown-content'] h1",
        supportsScale: true,
        defaultFontSizeStep: 4,
        typefaceRole: 'display',
      },
      {
        id: 'headings.h2',
        label: 'H2',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'letterSpacing'],
        defaults: { fontSize: 1.3, fontSizeUnit: 'em', fontWeight: 700, color: '#000', letterSpacing: 0 },
        selector: "[data-part='markdown-content'] h2",
        supportsScale: true,
        defaultFontSizeStep: 3,
        typefaceRole: 'display',
      },
      {
        id: 'headings.h3',
        label: 'H3',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'letterSpacing'],
        defaults: { fontSize: 1.1, fontSizeUnit: 'em', fontWeight: 700, color: '#000', letterSpacing: 0 },
        selector: "[data-part='markdown-content'] h3",
        supportsScale: true,
        defaultFontSizeStep: 2,
        typefaceRole: 'display',
      },
    ],
  },
  {
    id: 'markdown-code',
    label: 'Markdown Code',
    elements: [
      {
        id: 'code.inline',
        label: 'Inline Code',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 0.9, fontSizeUnit: 'em', fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] code",
        supportsScale: true,
        defaultFontSizeStep: -1,
        typefaceRole: 'code',
      },
      {
        id: 'code.block',
        label: 'Code Block',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'lineHeight', 'letterSpacing'],
        defaults: { fontFamily: 'mono', fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#000', lineHeight: 1.5, letterSpacing: 0 },
        selector: "[data-part='markdown-content'] pre",
        supportsScale: true,
        defaultFontSizeStep: -1,
        defaultLineHeightStep: 0,
        typefaceRole: 'code',
      },
    ],
  },
  {
    id: 'markdown-extras',
    label: 'Markdown Extras',
    elements: [
      {
        id: 'extras.blockquote',
        label: 'Blockquote',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#555' },
        selector: "[data-part='markdown-content'] blockquote",
        supportsScale: true,
        defaultFontSizeStep: 0,
        typefaceRole: 'body',
      },
      {
        id: 'extras.link',
        label: 'Link',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] a",
        typefaceRole: 'body',
      },
      {
        id: 'extras.table-header',
        label: 'Table Header',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] th",
        typefaceRole: 'body',
      },
    ],
  },
  {
    id: 'statusbar',
    label: 'Status Bar',
    elements: [
      {
        id: 'statusbar.text',
        label: 'Status Text',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', color: '#777' },
        selector: "[data-part='statusbar']",
        supportsScale: true,
        defaultFontSizeStep: -3,
        typefaceRole: 'display',
      },
    ],
  },
  {
    id: 'badges',
    label: 'Badges',
    elements: [
      {
        id: 'badges.badge',
        label: 'Badge',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='badge']",
        supportsScale: true,
        defaultFontSizeStep: -3,
        typefaceRole: 'display',
      },
    ],
  },
];

/** Build a flat map of element ID → element definition for quick lookup. */
export function buildElementMap(): Map<string, TypographyGroup['elements'][number]> {
  const map = new Map<string, TypographyGroup['elements'][number]>();
  for (const group of TYPOGRAPHY_GROUPS) {
    for (const elem of group.elements) {
      map.set(elem.id, elem);
    }
  }
  return map;
}
