// components/TypographyPalette/presets.ts
// Built-in typography presets and utilities for managing custom presets.

import type { TypographyPreset, TypographyOverrides, BaselineParameters, ElementSizeModeMap, ElementScaleSteps } from '../../types/typography-palette';
import { DEFAULT_BASELINE, SCALE_RATIOS } from '../../types/typography-palette';

const CLASSIC_MAC_DEFAULTS: TypographyOverrides = {};

const CLEAN_MODERN_OVERRIDES: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222', lineHeight: 1.7, letterSpacing: 0.01 },
  'titlebar.title':      { fontSize: 13, fontWeight: 600, color: '#111' },
  'header.heading':      { fontSize: 28, fontWeight: 700, color: '#111', letterSpacing: -0.02 },
  'header.subtitle':     { fontSize: 14, fontWeight: 400, color: '#666' },
  'prose.body':          { fontFamily: 'ui', fontSize: 16, fontWeight: 400, color: '#222', lineHeight: 1.7, letterSpacing: 0.01 },
  'headings.h1':         { fontSize: 2.0, fontSizeUnit: 'em', fontWeight: 700, color: '#111', letterSpacing: -0.02 },
  'headings.h2':         { fontSize: 1.5, fontSizeUnit: 'em', fontWeight: 600, color: '#222', letterSpacing: -0.01 },
  'headings.h3':         { fontSize: 1.25, fontSizeUnit: 'em', fontWeight: 600, color: '#333' },
  'code.inline':         { fontSize: 0.85, fontSizeUnit: 'em', color: '#333' },
  'code.block':          { fontSize: 14, color: '#222', lineHeight: 1.5 },
  'extras.link':         { color: '#333' },
  'extras.blockquote':   { color: '#666' },
  'statusbar.text':      { fontSize: 11, color: '#888' },
  'badges.badge':        { fontSize: 11, fontWeight: 500, color: '#333' },
};

const CLEAN_MODERN_BASELINE: BaselineParameters = {
  baseFontSize: 16,
  scaleRatioName: 'perfect-fourth',
  baseLineHeight: 1.7,
  baseLetterSpacing: 0.01,
  baseWordSpacing: 0,
};

const DENSE_TERMINAL_OVERRIDES: TypographyOverrides = {
  'root.body':           { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111', lineHeight: 1.4, letterSpacing: 0 },
  'titlebar.title':      { fontSize: 11, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 16, fontWeight: 700, color: '#000' },
  'header.subtitle':     { fontSize: 11, fontWeight: 400, color: '#888' },
  'prose.body':          { fontFamily: 'mono', fontSize: 12, fontWeight: 400, color: '#111', lineHeight: 1.4, letterSpacing: 0 },
  'headings.h1':         { fontSize: 1.4, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h2':         { fontSize: 1.2, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.0, fontSizeUnit: 'em', fontWeight: 700, color: '#222' },
  'code.inline':         { fontFamily: 'mono', fontSize: 1.0, fontSizeUnit: 'em', color: '#111' },
  'code.block':          { fontFamily: 'mono', fontSize: 11, color: '#111', lineHeight: 1.3 },
  'statusbar.text':      { fontSize: 9, color: '#999' },
  'badges.badge':        { fontSize: 9, fontWeight: 700, color: '#000' },
};

const DENSE_TERMINAL_BASELINE: BaselineParameters = {
  baseFontSize: 12,
  scaleRatioName: 'minor-third',
  baseLineHeight: 1.4,
  baseLetterSpacing: 0,
  baseWordSpacing: 0,
};

const LARGE_PRINT_OVERRIDES: TypographyOverrides = {
  'root.body':           { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000', lineHeight: 1.8, letterSpacing: 0.02 },
  'titlebar.title':      { fontSize: 15, fontWeight: 700, color: '#000' },
  'header.heading':      { fontSize: 32, fontWeight: 700, color: '#000', letterSpacing: -0.02 },
  'header.subtitle':     { fontSize: 16, fontWeight: 400, color: '#333' },
  'prose.body':          { fontFamily: 'ui', fontSize: 18, fontWeight: 400, color: '#000', lineHeight: 1.8, letterSpacing: 0.02, wordSpacing: 0.05 },
  'headings.h1':         { fontSize: 2.0, fontSizeUnit: 'em', fontWeight: 700, color: '#000', letterSpacing: -0.02 },
  'headings.h2':         { fontSize: 1.5, fontSizeUnit: 'em', fontWeight: 700, color: '#000' },
  'headings.h3':         { fontSize: 1.3, fontSizeUnit: 'em', fontWeight: 700, color: '#111' },
  'code.inline':         { fontSize: 0.85, fontSizeUnit: 'em', color: '#000' },
  'code.block':          { fontSize: 16, color: '#000', lineHeight: 1.6 },
  'statusbar.text':      { fontSize: 12, color: '#555' },
  'badges.badge':        { fontSize: 12, fontWeight: 700, color: '#000' },
};

const LARGE_PRINT_BASELINE: BaselineParameters = {
  baseFontSize: 18,
  scaleRatioName: 'perfect-fifth',
  baseLineHeight: 1.8,
  baseLetterSpacing: 0.02,
  baseWordSpacing: 0.05,
};

// Scale-mode presets: these use elementModes + elementScaleSteps instead of overrides
// This demonstrates the design system approach — baseline drives everything.

const CLEAN_MODERN_SCALE_MODES: ElementSizeModeMap = {};
const CLEAN_MODERN_SCALE_STEPS: Record<string, ElementScaleSteps> = {};

const DENSE_TERMINAL_SCALE_MODES: ElementSizeModeMap = {};
const DENSE_TERMINAL_SCALE_STEPS: Record<string, ElementScaleSteps> = {};

const LARGE_PRINT_SCALE_MODES: ElementSizeModeMap = {};
const LARGE_PRINT_SCALE_STEPS: Record<string, ElementScaleSteps> = {};

export const BUILT_IN_PRESETS: TypographyPreset[] = [
  {
    id: 'classic-mac',
    label: 'Classic Mac (default)',
    isBuiltIn: true,
    overrides: CLASSIC_MAC_DEFAULTS,
    baseline: { ...DEFAULT_BASELINE },
  },
  {
    id: 'clean-modern',
    label: 'Clean Modern',
    isBuiltIn: true,
    overrides: CLEAN_MODERN_OVERRIDES,
    baseline: CLEAN_MODERN_BASELINE,
    elementModes: CLEAN_MODERN_SCALE_MODES,
    elementScaleSteps: CLEAN_MODERN_SCALE_STEPS,
  },
  {
    id: 'dense-terminal',
    label: 'Dense Terminal',
    isBuiltIn: true,
    overrides: DENSE_TERMINAL_OVERRIDES,
    baseline: DENSE_TERMINAL_BASELINE,
    elementModes: DENSE_TERMINAL_SCALE_MODES,
    elementScaleSteps: DENSE_TERMINAL_SCALE_STEPS,
  },
  {
    id: 'large-print',
    label: 'Large Print',
    isBuiltIn: true,
    overrides: LARGE_PRINT_OVERRIDES,
    baseline: LARGE_PRINT_BASELINE,
    elementModes: LARGE_PRINT_SCALE_MODES,
    elementScaleSteps: LARGE_PRINT_SCALE_STEPS,
  },
  // Scale-driven preset: demonstrates the design system approach
  {
    id: 'scale-system',
    label: 'Scale System (1.25)',
    isBuiltIn: true,
    overrides: {},
    baseline: {
      baseFontSize: 16,
      scaleRatioName: 'major-third',
      baseLineHeight: 1.6,
      baseLetterSpacing: 0,
      baseWordSpacing: 0,
    },
    elementModes: {
      'root.body': 'scale',
      'titlebar.title': 'scale',
      'menubar.items': 'scale',
      'menubar.appname': 'scale',
      'sidebar.search': 'scale',
      'sidebar.packageselector': 'scale',
      'sidebar.navtoggle': 'scale',
      'sidebar.typefilter': 'scale',
      'tree.row': 'scale',
      'tree.heading': 'scale',
      'cards.title': 'scale',
      'cards.short': 'scale',
      'header.slug': 'scale',
      'header.heading': 'scale',
      'header.subtitle': 'scale',
      'prose.body': 'scale',
      'headings.h1': 'scale',
      'headings.h2': 'scale',
      'headings.h3': 'scale',
      'code.inline': 'scale',
      'code.block': 'scale',
      'extras.blockquote': 'scale',
      'statusbar.text': 'scale',
      'badges.badge': 'scale',
    },
    elementScaleSteps: {
      'root.body':          { fontSizeStep: 0, lineHeightStep: 0 },
      'titlebar.title':     { fontSizeStep: -1 },
      'menubar.items':      { fontSizeStep: -1 },
      'menubar.appname':    { fontSizeStep: -2 },
      'sidebar.search':     { fontSizeStep: -1 },
      'sidebar.packageselector': { fontSizeStep: 0 },
      'sidebar.navtoggle':  { fontSizeStep: 0 },
      'sidebar.typefilter': { fontSizeStep: -3 },
      'tree.row':           { fontSizeStep: -1 },
      'tree.heading':       { fontSizeStep: -2 },
      'cards.title':        { fontSizeStep: -1 },
      'cards.short':        { fontSizeStep: -3 },
      'header.slug':        { fontSizeStep: -3 },
      'header.heading':     { fontSizeStep: 4 },
      'header.subtitle':    { fontSizeStep: -1 },
      'prose.body':         { fontSizeStep: 0, lineHeightStep: 0 },
      'headings.h1':        { fontSizeStep: 4 },
      'headings.h2':        { fontSizeStep: 3 },
      'headings.h3':        { fontSizeStep: 2 },
      'code.inline':        { fontSizeStep: -1 },
      'code.block':         { fontSizeStep: -1, lineHeightStep: 0 },
      'extras.blockquote':  { fontSizeStep: 0 },
      'statusbar.text':     { fontSizeStep: -3 },
      'badges.badge':       { fontSizeStep: -3 },
    },
  },
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
