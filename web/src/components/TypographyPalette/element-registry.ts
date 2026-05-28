// components/TypographyPalette/element-registry.ts
// Defines every adjustable typography element in the docs browser,
// organized into accordion groups. This is the single source of truth
// for what the palette can override.

import type { TypographyGroup } from '../../types/typography-palette';

export const TYPOGRAPHY_GROUPS: TypographyGroup[] = [
  {
    id: 'root',
    label: 'Root / Body',
    elements: [
      {
        id: 'root.body',
        label: 'Body Text',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: '.app-root',
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
      },
      {
        id: 'menubar.appname',
        label: 'App Name',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='menubar-title']",
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
      },
      {
        id: 'sidebar.packageselector',
        label: 'Package Selector',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='package-selector-root']",
      },
      {
        id: 'sidebar.navtoggle',
        label: 'Nav Mode Toggle',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 13, fontSizeUnit: 'px', color: '#000' },
        selector: "[data-part='navigation-mode-toggle-root']",
      },
      {
        id: 'sidebar.typefilter',
        label: 'Type Filter',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='typefilter-button']",
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
      },
      {
        id: 'tree.heading',
        label: 'Heading Row',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 11, fontSizeUnit: 'px', color: '#3f4b5a' },
        selector: "[data-part='documentation-tree-row'][data-kind='heading']",
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
      },
      {
        id: 'cards.short',
        label: 'Card Description',
        adjustable: ['fontSize', 'color'],
        defaults: { fontSize: 10, fontSizeUnit: 'px', color: '#777' },
        selector: "[data-part~='section-card-short']",
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
      },
      {
        id: 'header.heading',
        label: 'Heading',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 24, fontSizeUnit: 'px', fontWeight: 700, color: '#000' },
        selector: "[data-part='section-header-heading']",
      },
      {
        id: 'header.subtitle',
        label: 'Subtitle',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#555' },
        selector: "[data-part='section-header-subtitle']",
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
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color', 'lineHeight'],
        defaults: { fontFamily: 'ui', fontSize: 13, fontSizeUnit: 'px', fontWeight: 400, color: '#000', lineHeight: 1.6 },
        selector: "[data-part='markdown-content']",
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
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.6, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h1",
      },
      {
        id: 'headings.h2',
        label: 'H2',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.3, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h2",
      },
      {
        id: 'headings.h3',
        label: 'H3',
        adjustable: ['fontSize', 'fontWeight', 'color'],
        defaults: { fontSize: 1.1, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] h3",
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
      },
      {
        id: 'code.block',
        label: 'Code Block',
        adjustable: ['fontFamily', 'fontSize', 'fontWeight', 'color'],
        defaults: { fontFamily: 'mono', fontSize: 12, fontSizeUnit: 'px', fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] pre",
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
      },
      {
        id: 'extras.link',
        label: 'Link',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 400, color: '#000' },
        selector: "[data-part='markdown-content'] a",
      },
      {
        id: 'extras.table-header',
        label: 'Table Header',
        adjustable: ['fontWeight', 'color'],
        defaults: { fontWeight: 700, color: '#000' },
        selector: "[data-part='markdown-content'] th",
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
