// types/typography-palette.ts
// TypeScript type definitions for the Typography Debug Palette.
// These types define the data model for typography overrides, presets,
// and the element registry.

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
export type FontFamily = 'ui' | 'mono';

/** Human-readable labels for font families. */
export const FONT_FAMILY_LABELS: Record<FontFamily, string> = {
  ui: 'Chicago_',
  mono: 'Monaco',
};

/** CSS font stacks that correspond to each FontFamily value. */
export const FONT_STACKS: Record<FontFamily, string> = {
  ui: "'Chicago_', 'Geneva', 'Charcoal', 'Lucida Grande', 'Helvetica Neue', sans-serif",
  mono: "'Monaco', 'Courier New', monospace",
};

/** Standard font weight values. */
export type FontWeight = 100 | 200 | 300 | 400 | 500 | 600 | 700 | 800 | 900;

/** All available font weight values for dropdowns. */
export const FONT_WEIGHTS: FontWeight[] = [100, 200, 300, 400, 500, 600, 700, 800, 900];

/** Typography properties that can be overridden for any element. */
export interface TypographyProperties {
  fontFamily?: FontFamily;
  fontSize?: number;       // in px (or em for relative elements)
  fontSizeUnit?: 'px' | 'em';
  fontWeight?: FontWeight;
  color?: GrayColor;
  lineHeight?: number;    // unitless multiplier (e.g., 1.6)
}

/** Map of element ID → override properties. */
export type TypographyOverrides = Record<string, TypographyProperties>;

/** A preset is a named collection of overrides. */
export interface TypographyPreset {
  id: string;
  label: string;
  /** true if this is a built-in preset (cannot be deleted). */
  isBuiltIn?: boolean;
  overrides: TypographyOverrides;
}

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
  adjustable: ('fontFamily' | 'fontSize' | 'fontWeight' | 'color' | 'lineHeight')[];
  /** Default values (from the current CSS). */
  defaults: TypographyProperties;
  /** CSS selector to target this element. */
  selector: string;
}

/** localStorage key for persisted palette state. */
export const PALETTE_STORAGE_KEY = 'glazed-typography-palette';

/** Shape of persisted palette state in localStorage. */
export interface PersistedPaletteState {
  overrides: TypographyOverrides;
  activePreset: string | null;
  customPresets: TypographyPreset[];
}
