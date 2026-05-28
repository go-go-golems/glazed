// types/typography-palette.ts
// TypeScript type definitions for the Typography Debug Palette.
// These types define the data model for typography overrides, presets,
// the design system baseline, and the element registry.

/** A single grayscale color value used in the monochrome palette. */
export type GrayColor =
  | '#000' | '#111' | '#222' | '#333' | '#444'
  | '#555' | '#666' | '#777' | '#888' | '#999'
  | '#aaa' | '#bbb' | '#ccc' | '#ddd' | '#eee' | '#fff';

/** Ordered list of gray shades for the ColorStepper. */
export const GRAY_SHADES: GrayColor[] = [
  '#000', '#111', '#222', '#333', '#444',
  '#555', '#666', '#777', '#888', '#999',
  '#aaa', '#bbb', '#ccc', '#ddd', '#eee', '#fff',
];

/** Available font families for the dropdown. */
export type FontFamily = 'ui' | 'mono' | 'serif';

/** Human-readable labels for font families. */
export const FONT_FAMILY_LABELS: Record<FontFamily, string> = {
  ui: 'Chicago_',
  mono: 'Monaco',
  serif: 'EB Garamond',
};

/** CSS font stacks that correspond to each FontFamily value. */
export const FONT_STACKS: Record<FontFamily, string> = {
  ui: "'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif",
  mono: "'Monaco', 'Courier New', monospace",
  serif: "'EB Garamond', 'Garamond', 'Georgia', 'Palatino', 'Times New Roman', serif",
};

/** Standard font weight values. */
export type FontWeight = 100 | 200 | 300 | 400 | 500 | 600 | 700 | 800 | 900;

/** All available font weight values for dropdowns. */
export const FONT_WEIGHTS: FontWeight[] = [100, 200, 300, 400, 500, 600, 700, 800, 900];

/** Typography properties that can be overridden for any element. */
export interface TypographyProperties {
  fontFamily?: FontFamily;
  fontSize?: number;         // in px (or em for relative elements)
  fontSizeUnit?: 'px' | 'em';
  fontWeight?: FontWeight;
  color?: GrayColor;
  lineHeight?: number;      // unitless multiplier (e.g., 1.6)
  letterSpacing?: number;    // in em (e.g., 0.05)
  wordSpacing?: number;     // in em (e.g., 0.1)
}

/** Map of element ID → override properties. */
export type TypographyOverrides = Record<string, TypographyProperties>;

// ---------------------------------------------------------------------------
// Design System Baseline
// ---------------------------------------------------------------------------

/** Named scale ratios for the baseline design system. */
export type ScaleRatioName = 'minor-second' | 'major-second' | 'minor-third' | 'major-third' | 'perfect-fourth' | 'aug-fourth' | 'perfect-fifth' | 'golden';

/** Numeric values for named scale ratios. */
export const SCALE_RATIOS: Record<ScaleRatioName, number> = {
  'minor-second':  1.067,
  'major-second':  1.125,
  'minor-third':   1.200,
  'major-third':   1.250,
  'perfect-fourth': 1.333,
  'aug-fourth':    1.414,
  'perfect-fifth': 1.500,
  'golden':        1.618,
};

/** Human-readable labels for scale ratios. */
export const SCALE_RATIO_LABELS: Record<ScaleRatioName, string> = {
  'minor-second':  '1.067 — Minor Second',
  'major-second':  '1.125 — Major Second',
  'minor-third':   '1.200 — Minor Third',
  'major-third':   '1.250 — Major Third',
  'perfect-fourth': '1.333 — Perfect Fourth',
  'aug-fourth':    '1.414 — Aug Fourth',
  'perfect-fifth': '1.500 — Perfect Fifth',
  'golden':        '1.618 — Golden Ratio',
};

/** All scale ratio names for dropdowns. */
export const SCALE_RATIO_NAMES: ScaleRatioName[] = [
  'minor-second', 'major-second', 'minor-third', 'major-third',
  'perfect-fourth', 'aug-fourth', 'perfect-fifth', 'golden',
];

/** Scale step labels for the step selector. */
export const SCALE_STEPS = [-3, -2, -1, 0, 1, 2, 3, 4, 5, 6] as const;
export type ScaleStep = typeof SCALE_STEPS[number];

/** Human-readable labels for scale steps. */
export const SCALE_STEP_LABELS: Record<number, string> = {
  [-3]: 'xs  (−3)',
  [-2]: 'sm  (−2)',
  [-1]: 'md  (−1)',
  0:   'base (0)',
  1:   'lg  (+1)',
  2:   'xl  (+2)',
  3:   '2xl (+3)',
  4:   '3xl (+4)',
  5:   '4xl (+5)',
  6:   '5xl (+6)',
};

/** Compute a scaled value: base × ratio^step. */
export function computeScaledValue(base: number, ratio: number, step: number): number {
  return +(base * Math.pow(ratio, step)).toFixed(2);
}

/** The design system baseline parameters. */
export interface BaselineParameters {
  /** Base font size in px (default 13). */
  baseFontSize: number;
  /** Scale ratio name (default 'major-third'). */
  scaleRatioName: ScaleRatioName;
  /** Base line height multiplier (default 1.6). */
  baseLineHeight: number;
  /** Base letter spacing in em (default 0). */
  baseLetterSpacing: number;
  /** Base word spacing in em (default 0). */
  baseWordSpacing: number;
}

/** Default baseline parameters matching the current Classic Mac aesthetic. */
export const DEFAULT_BASELINE: BaselineParameters = {
  baseFontSize: 13,
  scaleRatioName: 'major-third',
  baseLineHeight: 1.6,
  baseLetterSpacing: 0,
  baseWordSpacing: 0,
};

/** Element-level scale step overrides (step per property). */
export interface ElementScaleSteps {
  fontSizeStep?: ScaleStep;
  lineHeightStep?: ScaleStep;   // multiplier on baseLineHeight: 0 = use base, +1 = base*1.1, etc.
  letterSpacingStep?: ScaleStep; // multiplier on baseLetterSpacing
  wordSpacingStep?: ScaleStep;   // multiplier on baseWordSpacing
}

/** Per-element mode: 'custom' = absolute values, 'scale' = derived from baseline. */
export type ElementSizeMode = 'custom' | 'scale';

/** Per-element mode map (element ID → mode). */
export type ElementSizeModeMap = Record<string, ElementSizeMode>;

// ---------------------------------------------------------------------------
// Presets
// ---------------------------------------------------------------------------

/** A preset is a named collection of overrides. */
export interface TypographyPreset {
  id: string;
  label: string;
  /** true if this is a built-in preset (cannot be deleted). */
  isBuiltIn?: boolean;
  overrides: TypographyOverrides;
  /** Optional baseline parameters stored with the preset. */
  baseline?: BaselineParameters;
  /** Optional scale mode map stored with the preset. */
  elementModes?: ElementSizeModeMap;
  /** Optional scale steps stored with the preset. */
  elementScaleSteps?: Record<string, ElementScaleSteps>;
}

// ---------------------------------------------------------------------------
// Element Registry
// ---------------------------------------------------------------------------

/** Accordion group definition. */
export interface TypographyGroup {
  id: string;
  label: string;
  /** Sub-elements within this group. */
  elements: TypographyElement[];
}

/** A single adjustable element within a group. */
export interface TypographyElement {
  id: string;
  label: string;
  /** Which properties are adjustable for this element. */
  adjustable: ('fontFamily' | 'fontSize' | 'fontWeight' | 'color' | 'lineHeight' | 'letterSpacing' | 'wordSpacing')[];
  /** Default values (from the current CSS). */
  defaults: TypographyProperties;
  /** CSS selector to target this element. */
  selector: string;
  /** If true, this element supports the scale mode (design system derived values). */
  supportsScale?: boolean;
  /** Default scale step for font size (when in scale mode). */
  defaultFontSizeStep?: ScaleStep;
  /** Default scale step for line height offset (when in scale mode). */
  defaultLineHeightStep?: ScaleStep;
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------

/** localStorage key for persisted palette state. */
export const PALETTE_STORAGE_KEY = 'glazed-typography-palette';

/** Shape of persisted palette state in localStorage. */
export interface PersistedPaletteState {
  overrides: TypographyOverrides;
  activePreset: string | null;
  customPresets: TypographyPreset[];
  baseline: BaselineParameters;
  elementModes: ElementSizeModeMap;
  elementScaleSteps: Record<string, ElementScaleSteps>;
}
