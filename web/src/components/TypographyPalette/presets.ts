// components/TypographyPalette/presets.ts
// Built-in typography presets and utilities for managing custom presets.

import type { TypographyPreset, TypographyOverrides } from '../../types/typography-palette';

const CLASSIC_MAC_DEFAULTS: TypographyOverrides = {};

const CLEAN_MODERN: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222' },
  'titlebar.title':      { fontSize: 13, fontWeight: 600, color: '#111' },
  'header.heading':      { fontSize: 28, fontWeight: 700, color: '#111' },
  'header.subtitle':     { fontSize: 14, fontWeight: 400, color: '#666' },
  'prose.body':          { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222', lineHeight: 1.7 },
  'headings.h1':         { fontSize: 2.0, fontSizeUnit: 'em', fontWeight: 700, color: '#111' },
  'headings.h2':         { fontSize: 1.5, fontSizeUnit: 'em', fontWeight: 600, color: '#222' },
  'headings.h3':         { fontSize: 1.25, fontSizeUnit: 'em', fontWeight: 600, color: '#333' },
  'code.inline':         { fontSize: 0.85, fontSizeUnit: 'em', color: '#333' },
  'code.block':          { fontSize: 14, color: '#222' },
  'extras.link':         { color: '#333' },
  'extras.blockquote':  { color: '#666' },
  'statusbar.text':      { fontSize: 11, color: '#888' },
  'badges.badge':        { fontSize: 11, fontWeight: 500, color: '#333' },
};

const DENSE_TERMINAL: TypographyOverrides = {
  'root.body':           { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111' },
  'titlebar.title':      { fontSize: 11, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 16, fontWeight: 700, color: '#000' },
  'header.subtitle':     { fontSize: 11, fontWeight: 400, color: '#888' },
  'prose.body':          { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111', lineHeight: 1.4 },
  'headings.h1':         { fontSize: 1.4, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h2':         { fontSize: 1.2, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.0, fontSizeUnit: 'em', fontWeight: 700, color: '#222' },
  'code.inline':         { fontFamily: 'mono', fontSize: 1.0, fontSizeUnit: 'em', color: '#111' },
  'code.block':          { fontFamily: 'mono', fontSize: 11, color: '#111' },
  'statusbar.text':      { fontSize: 9, color: '#999' },
  'badges.badge':        { fontSize: 9, fontWeight: 700, color: '#000' },
};

const LARGE_PRINT: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000' },
  'titlebar.title':      { fontSize: 15, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 32, fontWeight: 700, color: '#000' },
  'header.subtitle':     { fontSize: 16, fontWeight: 400, color: '#333' },
  'prose.body':          { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000', lineHeight: 1.8 },
  'headings.h1':         { fontSize: 2.0, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h2':         { fontSize: 1.5, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.3, fontSizeUnit: 'em', fontWeight: 700, color: '#111' },
  'code.inline':         { fontSize: 0.85, fontSizeUnit: 'em', color: '#000' },
  'code.block':          { fontSize: 16, color: '#000' },
  'statusbar.text':      { fontSize: 12, color: '#555' },
  'badges.badge':        { fontSize: 12, fontWeight: 700, color: '#000' },
};

export const BUILT_IN_PRESETS: TypographyPreset[] = [
  { id: 'classic-mac',    label: 'Classic Mac (default)', isBuiltIn: true, overrides: CLASSIC_MAC_DEFAULTS },
  { id: 'clean-modern',   label: 'Clean Modern',          isBuiltIn: true, overrides: CLEAN_MODERN },
  { id: 'dense-terminal', label: 'Dense Terminal',       isBuiltIn: true, overrides: DENSE_TERMINAL },
  { id: 'large-print',    label: 'Large Print',           isBuiltIn: true, overrides: LARGE_PRINT },
];

/** Find a built-in preset by ID. */
export function getBuiltInPreset(id: string): TypographyPreset | undefined {
  return BUILT_IN_PRESETS.find(p => p.id === id);
}

/** Merge built-in and custom presets into a single list. */
export function getAllPresets(customPresets: TypographyPreset[]): TypographyPreset[] {
  return [...BUILT_IN_PRESETS, ...customPresets];
}

/** Generate a unique ID for a new custom preset. */
export function newCustomPresetId(): string {
  return `custom-${Date.now()}`;
}
