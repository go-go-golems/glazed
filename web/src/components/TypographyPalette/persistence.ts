// components/TypographyPalette/persistence.ts
// localStorage persistence for typography palette state.
// Saves/loads overrides, active preset, custom presets, baseline, modes, and scale steps.

import type {
  PersistedPaletteState, TypographyOverrides, TypographyPreset,
  BaselineParameters, ElementSizeModeMap, ElementScaleSteps, TypefaceRoleMap,
} from '../../types/typography-palette';
import { PALETTE_STORAGE_KEY, DEFAULT_TYPEFACE_ROLES } from '../../types/typography-palette';

/** Save palette state to localStorage. */
export function persistPaletteState(
  overrides: TypographyOverrides,
  activePreset: string | null,
  customPresets: TypographyPreset[],
  baseline: BaselineParameters,
  elementModes: ElementSizeModeMap,
  elementScaleSteps: Record<string, ElementScaleSteps>,
  typefaceRoles: TypefaceRoleMap,
): void {
  const state: PersistedPaletteState = {
    overrides,
    activePreset,
    customPresets,
    baseline,
    elementModes,
    elementScaleSteps,
    typefaceRoles,
  };

  try {
    localStorage.setItem(PALETTE_STORAGE_KEY, JSON.stringify(state));
  } catch {
    // localStorage may be full or unavailable — silently ignore
  }
}

/** Load palette state from localStorage. Returns null if nothing saved. */
export function loadPaletteState(): PersistedPaletteState | null {
  try {
    const raw = localStorage.getItem(PALETTE_STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as PersistedPaletteState;
    // Basic validation
    if (typeof parsed !== 'object' || parsed === null) return null;
    if (typeof parsed.overrides !== 'object') return null;
    if (!Array.isArray(parsed.customPresets)) return null;
    // Gracefully handle missing typefaceRoles from older persisted state
    if (!parsed.typefaceRoles) {
      parsed.typefaceRoles = { ...DEFAULT_TYPEFACE_ROLES };
    }
    return parsed;
  } catch {
    return null;
  }
}

/** Clear persisted palette state. */
export function clearPaletteState(): void {
  try {
    localStorage.removeItem(PALETTE_STORAGE_KEY);
  } catch {
    // silently ignore
  }
}
