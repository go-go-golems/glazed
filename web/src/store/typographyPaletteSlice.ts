// store/typographyPaletteSlice.ts
// Redux slice for the Typography Debug Palette.
// Manages override state, baseline parameters, scale modes, presets,
// visibility, and localStorage persistence.

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type {
  TypographyOverrides, TypographyProperties, TypographyPreset,
  BaselineParameters, ElementSizeModeMap, ElementScaleSteps,
} from '../types/typography-palette';
import { DEFAULT_BASELINE } from '../types/typography-palette';
import { loadPaletteState, persistPaletteState, clearPaletteState } from '../components/TypographyPalette/persistence';

interface TypographyPaletteState {
  isOpen: boolean;
  activeGroup: string | null;
  activePreset: string | null;
  overrides: TypographyOverrides;
  customPresets: TypographyPreset[];
  copiedFeedback: string | null;
  // Design system baseline
  baseline: BaselineParameters;
  // Per-element mode: 'custom' = absolute values, 'scale' = derived from baseline
  elementModes: ElementSizeModeMap;
  // Per-element scale steps (when in 'scale' mode)
  elementScaleSteps: Record<string, ElementScaleSteps>;
}

/** Load initial state from localStorage if available. */
function loadInitialState(): TypographyPaletteState {
  const persisted = loadPaletteState();
  if (persisted) {
    return {
      isOpen: false,
      activeGroup: null,
      activePreset: persisted.activePreset,
      overrides: persisted.overrides ?? {},
      customPresets: persisted.customPresets ?? [],
      copiedFeedback: null,
      baseline: persisted.baseline ?? { ...DEFAULT_BASELINE },
      elementModes: persisted.elementModes ?? {},
      elementScaleSteps: persisted.elementScaleSteps ?? {},
    };
  }
  return {
    isOpen: false,
    activeGroup: null,
    activePreset: null,
    overrides: {},
    customPresets: [],
    copiedFeedback: null,
    baseline: { ...DEFAULT_BASELINE },
    elementModes: {},
    elementScaleSteps: {},
  };
}

/** Persist after any state change that affects overrides or presets. */
function persistAfterChange(state: TypographyPaletteState) {
  persistPaletteState(
    state.overrides,
    state.activePreset,
    state.customPresets,
    state.baseline,
    state.elementModes,
    state.elementScaleSteps,
  );
}

const typographyPaletteSlice = createSlice({
  name: 'typographyPalette',
  initialState: loadInitialState(),
  reducers: {
    togglePalette(state) {
      state.isOpen = !state.isOpen;
    },
    openPalette(state) {
      state.isOpen = true;
    },
    closePalette(state) {
      state.isOpen = false;
    },
    setActiveGroup(state, action: PayloadAction<string | null>) {
      state.activeGroup = action.payload;
    },
    setPreset(state, action: PayloadAction<{
      presetId: string;
      overrides: TypographyOverrides;
      baseline?: BaselineParameters;
      elementModes?: ElementSizeModeMap;
      elementScaleSteps?: Record<string, ElementScaleSteps>;
    }>) {
      state.activePreset = action.payload.presetId;
      state.overrides = { ...action.payload.overrides };
      if (action.payload.baseline) state.baseline = { ...action.payload.baseline };
      if (action.payload.elementModes) state.elementModes = { ...action.payload.elementModes };
      if (action.payload.elementScaleSteps) state.elementScaleSteps = { ...action.payload.elementScaleSteps };
      persistAfterChange(state);
    },
    setOverride(state, action: PayloadAction<{ elementId: string; properties: TypographyProperties }>) {
      const { elementId, properties } = action.payload;
      state.overrides[elementId] = {
        ...state.overrides[elementId],
        ...properties,
      };
      state.activePreset = null;
      persistAfterChange(state);
    },
    removeOverride(state, action: PayloadAction<string>) {
      delete state.overrides[action.payload];
      if (Object.keys(state.overrides).length === 0) {
        state.activePreset = null;
      }
      persistAfterChange(state);
    },
    resetAllOverrides(state) {
      state.overrides = {};
      state.activePreset = null;
      state.activeGroup = null;
      state.elementModes = {};
      state.elementScaleSteps = {};
      state.baseline = { ...DEFAULT_BASELINE };
      persistAfterChange(state);
    },
    /** Update a baseline parameter. */
    setBaseline(state, action: PayloadAction<Partial<BaselineParameters>>) {
      state.baseline = { ...state.baseline, ...action.payload };
      state.activePreset = null;
      persistAfterChange(state);
    },
    /** Set an element's size mode ('custom' or 'scale'). */
    setElementMode(state, action: PayloadAction<{ elementId: string; mode: ElementSizeModeMap[string] }>) {
      state.elementModes[action.payload.elementId] = action.payload.mode;
      state.activePreset = null;
      persistAfterChange(state);
    },
    /** Set an element's scale steps (when in 'scale' mode). */
    setElementScaleSteps(state, action: PayloadAction<{ elementId: string; steps: ElementScaleSteps }>) {
      state.elementScaleSteps[action.payload.elementId] = {
        ...state.elementScaleSteps[action.payload.elementId],
        ...action.payload.steps,
      };
      state.activePreset = null;
      persistAfterChange(state);
    },
    /** Save current overrides as a new custom preset. */
    saveAsPreset(state, action: PayloadAction<{ label: string; id: string }>) {
      const newPreset: TypographyPreset = {
        id: action.payload.id,
        label: action.payload.label,
        isBuiltIn: false,
        overrides: { ...state.overrides },
        baseline: { ...state.baseline },
        elementModes: { ...state.elementModes },
        elementScaleSteps: { ...state.elementScaleSteps },
      };
      state.customPresets.push(newPreset);
      state.activePreset = newPreset.id;
      persistAfterChange(state);
    },
    /** Delete a custom preset by ID. */
    deleteCustomPreset(state, action: PayloadAction<string>) {
      state.customPresets = state.customPresets.filter(p => p.id !== action.payload);
      if (state.activePreset === action.payload) {
        state.activePreset = null;
      }
      persistAfterChange(state);
    },
    /** Set the "Copied!" feedback message. */
    setCopiedFeedback(state, action: PayloadAction<string | null>) {
      state.copiedFeedback = action.payload;
    },
    /** Clear persisted state from localStorage. */
    clearPersistence() {
      clearPaletteState();
    },
  },
});

export const {
  togglePalette,
  openPalette,
  closePalette,
  setActiveGroup,
  setPreset,
  setOverride,
  removeOverride,
  resetAllOverrides,
  setBaseline,
  setElementMode,
  setElementScaleSteps,
  saveAsPreset,
  deleteCustomPreset,
  setCopiedFeedback,
  clearPersistence,
} = typographyPaletteSlice.actions;

export default typographyPaletteSlice.reducer;
