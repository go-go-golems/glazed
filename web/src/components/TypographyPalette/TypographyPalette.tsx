// components/TypographyPalette/TypographyPalette.tsx
// Main panel component for the Typography Debug Palette.
// Renders as a floating overlay with baseline parameters, preset selector,
// accordion groups, save-preset form, CSS export, and reset.

import { useState, useEffect, useCallback, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '../../store';
import {
  closePalette,
  setActiveGroup,
  setPreset,
  resetAllOverrides,
  saveAsPreset,
  deleteCustomPreset,
  setCopiedFeedback,
} from '../../store/typographyPaletteSlice';
import type { TypographyPreset } from '../../types/typography-palette';
import { newCustomPresetId } from './presets';
import { getAllPresets } from './presets';
import { copyCssToClipboard } from './css-override-engine';
import { TYPOGRAPHY_GROUPS } from './element-registry';
import { useTypographyOverrides } from './useTypographyOverrides';
import { useHighlightSync } from './useHighlightElement';
import { useInspectorMode, useInspectorToggle } from './useInspectorMode';
import { BaselineParametersPanel } from './BaselineParameters';
import { TypographyPaletteParts } from './parts';
import { TypographyPaletteGroup } from './TypographyPaletteGroup';
import './styles/typography-palette.css';

export function TypographyPalette() {
  const dispatch = useDispatch();
  const isOpen = useSelector((s: RootState) => s.typographyPalette.isOpen);
  const activeGroup = useSelector((s: RootState) => s.typographyPalette.activeGroup);
  const activePreset = useSelector((s: RootState) => s.typographyPalette.activePreset);
  const overrides = useSelector((s: RootState) => s.typographyPalette.overrides);
  const customPresets = useSelector((s: RootState) => s.typographyPalette.customPresets);
  const copiedFeedback = useSelector((s: RootState) => s.typographyPalette.copiedFeedback);
  const baseline = useSelector((s: RootState) => s.typographyPalette.baseline);
  const elementModes = useSelector((s: RootState) => s.typographyPalette.elementModes);
  const elementScaleSteps = useSelector((s: RootState) => s.typographyPalette.elementScaleSteps);

  const [saveFormVisible, setSaveFormVisible] = useState(false);
  const [presetName, setPresetName] = useState('');
  const [exportMenuOpen, setExportMenuOpen] = useState(false);
  const exportMenuRef = useRef<HTMLDivElement>(null);

  // Keep DOM in sync with overrides + scale-mode resolved values
  useTypographyOverrides();
  // Sync highlighted element overlay to DOM
  useHighlightSync();
  // Inspector mode click handler
  useInspectorMode();
  const inspector = useInspectorToggle();

  // Clear "Copied!" feedback after 2 seconds
  useEffect(() => {
    if (copiedFeedback) {
      const timer = setTimeout(() => dispatch(setCopiedFeedback(null)), 2000);
      return () => clearTimeout(timer);
    }
  }, [copiedFeedback, dispatch]);

  // Close export menu when clicking outside
  useEffect(() => {
    if (!exportMenuOpen) return;
    const handler = (e: MouseEvent) => {
      if (exportMenuRef.current && !exportMenuRef.current.contains(e.target as Node)) {
        setExportMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, [exportMenuOpen]);

  const allPresets = getAllPresets(customPresets);

  const handlePresetChange = useCallback(
    (presetId: string) => {
      const preset = allPresets.find((p) => p.id === presetId);
      if (preset) {
        dispatch(setPreset({
          presetId: preset.id,
          overrides: preset.overrides,
          baseline: preset.baseline,
          elementModes: preset.elementModes,
          elementScaleSteps: preset.elementScaleSteps,
          typefaceRoles: preset.typefaceRoles,
        }));
      }
    },
    [dispatch, allPresets],
  );

  const handleSavePreset = useCallback(() => {
    const trimmed = presetName.trim();
    if (!trimmed) return;
    dispatch(saveAsPreset({ label: trimmed, id: newCustomPresetId() }));
    setPresetName('');
    setSaveFormVisible(false);
  }, [dispatch, presetName]);

  const handleDeletePreset = useCallback(
    (id: string) => {
      dispatch(deleteCustomPreset(id));
    },
    [dispatch],
  );

  const handleExport = useCallback(
    async (format: 'rules' | 'variables') => {
      const ok = await copyCssToClipboard(overrides, format);
      dispatch(setCopiedFeedback(ok ? format : null));
      setExportMenuOpen(false);
    },
    [dispatch, overrides],
  );

  if (!isOpen) return null;

  const hasOverrides = Object.keys(overrides).length > 0 || Object.values(elementModes).some(m => m === 'scale');
  const hasScaleModes = Object.values(elementModes).some(m => m === 'scale');

  return (
    <div data-part={TypographyPaletteParts.root}>
      {/* Header */}
      <div data-part={TypographyPaletteParts.header}>
        <span data-part={TypographyPaletteParts.title}>𝒜a Typography</span>
        {/* Inspector/dropper toggle */}
        <button
          title={inspector.active ? 'Inspector active — click an element on the page' : 'Inspect: click to find element in palette'}
          onClick={inspector.toggle}
          style={{
            border: inspector.active ? '2px solid #ff6600' : 'none',
            padding: '0 4px',
            fontSize: 12,
            cursor: 'pointer',
            background: inspector.active ? '#fff3e0' : 'transparent',
            color: inspector.active ? '#ff6600' : '#888',
            borderRadius: 2,
            fontFamily: 'inherit',
            marginRight: 4,
          }}
        >
          💉
        </button>
        <button
          data-part={TypographyPaletteParts.closeBtn}
          onClick={() => dispatch(closePalette())}
        >
          ×
        </button>
      </div>

      {/* Baseline Design System parameters */}
      <BaselineParametersPanel />

      {/* Preset selector */}
      <div data-part={TypographyPaletteParts.presetRow}>
        <span data-part={TypographyPaletteParts.presetLabel}>Preset</span>
        <select
          data-part={TypographyPaletteParts.presetSelect}
          value={activePreset ?? ''}
          onChange={(e) => handlePresetChange(e.target.value)}
        >
          <option value="">— Custom —</option>
          {allPresets.map((p: TypographyPreset) => (
            <option key={p.id} value={p.id}>
              {p.label}
              {!p.isBuiltIn ? ' ★' : ''}
            </option>
          ))}
        </select>
      </div>

      {/* Custom preset delete buttons */}
      {customPresets.length > 0 && activePreset && !allPresets.find(p => p.id === activePreset)?.isBuiltIn && (
        <div style={{ padding: '2px 10px 4px', display: 'flex', justifyContent: 'flex-end' }}>
          <button
            className="palette-preset-delete-btn"
            onClick={() => handleDeletePreset(activePreset)}
            title="Delete this preset"
          >
            ✕ Delete preset
          </button>
        </div>
      )}

      {/* Save preset form */}
      {saveFormVisible ? (
        <div className="palette-save-preset-form">
          <input
            type="text"
            value={presetName}
            onChange={(e) => setPresetName(e.target.value)}
            placeholder="Preset name…"
            autoFocus
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleSavePreset();
              if (e.key === 'Escape') setSaveFormVisible(false);
            }}
          />
          <button onClick={handleSavePreset}>Save</button>
          <button onClick={() => setSaveFormVisible(false)}>Cancel</button>
        </div>
      ) : null}

      {/* Accordion groups */}
      <div data-part={TypographyPaletteParts.groups}>
        {TYPOGRAPHY_GROUPS.map((group) => (
          <TypographyPaletteGroup
            key={group.id}
            group={group}
            isExpanded={activeGroup === group.id}
            onToggle={() =>
              dispatch(setActiveGroup(activeGroup === group.id ? null : group.id))
            }
          />
        ))}
      </div>

      {/* Footer with actions */}
      <div data-part={TypographyPaletteParts.footer}>
        {hasOverrides && (
          <button
            data-part={TypographyPaletteParts.savePresetBtn}
            onClick={() => setSaveFormVisible(true)}
          >
            ★ Save
          </button>
        )}

        {/* Export menu */}
        <div style={{ position: 'relative' }} ref={exportMenuRef}>
          <button
            data-part={TypographyPaletteParts.exportBtn}
            onClick={() => setExportMenuOpen(!exportMenuOpen)}
          >
            {copiedFeedback ? '✓ Copied!' : '📋 Export'}
          </button>
          {exportMenuOpen && (
            <div data-part={TypographyPaletteParts.exportMenu}>
              <button onClick={() => handleExport('rules')}>
                Copy as CSS rules
              </button>
              <button onClick={() => handleExport('variables')}>
                Copy as CSS variables
              </button>
            </div>
          )}
        </div>

        {hasOverrides && (
          <button
            data-part={TypographyPaletteParts.resetBtn}
            onClick={() => dispatch(resetAllOverrides())}
          >
            Reset
          </button>
        )}
      </div>
    </div>
  );
}
