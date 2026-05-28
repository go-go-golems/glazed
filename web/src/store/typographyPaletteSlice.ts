// store/typographyPaletteSlice.ts
// Redux slice for the Typography Debug Palette.
// Manages override state, presets, visibility, and localStorage persistence.

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { TypographyOverrides, TypographyProperties, TypographyPreset } from '../types/typography-palette';
import { loadPaletteState, persistPaletteState, clearPaletteState } from '../components/TypographyPalette/persistence';

interface TypographyPaletteState {
  isOpen: boolean;
  activeGroup: string | null;
  activePreset: string | null;
  overrides: TypographyOverrides;
  customPresets: TypographyPreset[];
  copiedFeedback: string | null;  // 'rules' | 'variables' | null — brief "Copied!" flash
}

/** Load initial state from localStorage if available. */
function loadInitialState(): TypographyPaletteState {
  const persisted = loadPaletteState();
  if (persisted) {
    return {
      isOpen: false,
      activeGroup: null,
      activePreset: persisted.activePreset,
      overrides: persisted.overrides,
      customPresets: persisted.customPresets,
      copiedFeedback: null,
    };
  }
  return {
    isOpen: false,
    activeGroup: null,
    activePreset: null,
    overrides: {},
    customPresets: [],
    copiedFeedback: null,
  };
}

/** Persist after any state change that affects overrides or presets. */
function persistAfterChange(state: TypographyPaletteState) {
  persistPaletteState(state.overrides, state.activePreset, state.customPresets);
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
    setPreset(state, action: PayloadAction<{ presetId: string; overrides: TypographyOverrides }>) {
      state.activePreset = action.payload.presetId;
      state.overrides = { ...action.payload.overrides };
      persistAfterChange(state);
    },
    setOverride(state, action: PayloadAction<{ elementId: string; properties: TypographyProperties }>) {
      const { elementId, properties } = action.payload;
      state.overrides[elementId] = {
        ...state.overrides[elementId],
        ...properties,
      };
      state.activePreset = null; // manual edit clears preset indicator
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
      persistAfterChange(state);
    },
    /** Save current overrides as a new custom preset. */
    saveAsPreset(state, action: PayloadAction<{ label: string; id: string }>) {
      const newPreset: TypographyPreset = {
        id: action.payload.id,
        label: action.payload.label,
        isBuiltIn: false,
        overrides: { ...state.overrides },
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
  saveAsPreset,
  deleteCustomPreset,
  setCopiedFeedback,
  clearPersistence,
} = typographyPaletteSlice.actions;

export default typographyPaletteSlice.reducer;
