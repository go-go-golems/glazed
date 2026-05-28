// components/TypographyPalette/BaselineParameters.tsx
// Controls for the design system baseline parameters.
// These global values drive all scale-mode elements.

import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { setBaseline, setTypefaceRole } from '../../store/typographyPaletteSlice';
import type { BaselineParameters, ScaleRatioName, TypefaceRole, FontFamily } from '../../types/typography-palette';
import { SCALE_RATIOS, SCALE_RATIO_NAMES, SCALE_RATIO_LABELS, computeScaledValue, TYPEFACE_ROLES, TYPEFACE_ROLE_LABELS, FONT_FAMILY_LABELS } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';
import { FontSizeStepper } from './FontSizeStepper';

/** Compute and display the size scale preview. */
function ScalePreview({ baseline }: { baseline: BaselineParameters }) {
  const ratio = SCALE_RATIOS[baseline.scaleRatioName];
  const steps = [-2, -1, 0, 1, 2, 3, 4];
  return (
    <div style={{ fontSize: 10, color: '#777', padding: '4px 0 2px', fontFamily: 'var(--font-mono)' }}>
      {steps.map(step => (
        <span key={step} style={{ marginRight: 8 }}>
          {step === 0 ? '→' : ' '}{computeScaledValue(baseline.baseFontSize, ratio, step)}px
        </span>
      ))}
    </div>
  );
}

export function BaselineParametersPanel() {
  const dispatch = useDispatch();
  const baseline = useSelector((s: RootState) => s.typographyPalette.baseline);
  const typefaceRoles = useSelector((s: RootState) => s.typographyPalette.typefaceRoles);

  const handleChange = (partial: Partial<BaselineParameters>) => {
    dispatch(setBaseline(partial));
  };

  const handleRoleChange = (role: TypefaceRole, fontFamily: FontFamily) => {
    dispatch(setTypefaceRole({ role, fontFamily }));
  };

  const ratio = SCALE_RATIOS[baseline.scaleRatioName];

  return (
    <div style={{ padding: '6px 10px 8px', borderBottom: '2px solid #000', background: '#f8f6f0' }}>
      <div style={{ fontWeight: 700, fontSize: 11, marginBottom: 6, color: '#444' }}>
        📐 Baseline Design System
      </div>

      {/* Base font size */}
      <div data-part={TypographyPaletteParts.controlRow}>
        <span data-part={TypographyPaletteParts.controlLabel}>Base</span>
        <FontSizeStepper
          value={baseline.baseFontSize}
          unit="px"
          min={8}
          max={32}
          onChange={(v) => handleChange({ baseFontSize: v })}
        />
      </div>

      {/* Scale ratio */}
      <div data-part={TypographyPaletteParts.controlRow}>
        <span data-part={TypographyPaletteParts.controlLabel}>Scale</span>
        <select
          data-part={TypographyPaletteParts.select}
          value={baseline.scaleRatioName}
          onChange={(e) => handleChange({ scaleRatioName: e.target.value as ScaleRatioName })}
        >
          {SCALE_RATIO_NAMES.map(name => (
            <option key={name} value={name}>
              {SCALE_RATIO_LABELS[name]}
            </option>
          ))}
        </select>
      </div>

      {/* Line height */}
      <div data-part={TypographyPaletteParts.controlRow}>
        <span data-part={TypographyPaletteParts.controlLabel}>L-Height</span>
        <FontSizeStepper
          value={baseline.baseLineHeight}
          unit=""
          step={0.1}
          min={0.8}
          max={3}
          onChange={(v) => handleChange({ baseLineHeight: v })}
        />
      </div>

      {/* Letter spacing */}
      <div data-part={TypographyPaletteParts.controlRow}>
        <span data-part={TypographyPaletteParts.controlLabel}>L-Space</span>
        <FontSizeStepper
          value={baseline.baseLetterSpacing}
          unit="em"
          step={0.01}
          min={-0.1}
          max={0.5}
          onChange={(v) => handleChange({ baseLetterSpacing: v })}
        />
      </div>

      {/* Word spacing */}
      <div data-part={TypographyPaletteParts.controlRow}>
        <span data-part={TypographyPaletteParts.controlLabel}>W-Space</span>
        <FontSizeStepper
          value={baseline.baseWordSpacing}
          unit="em"
          step={0.01}
          min={-0.2}
          max={1}
          onChange={(v) => handleChange({ baseWordSpacing: v })}
        />
      </div>

      {/* Scale preview */}
      <ScalePreview baseline={baseline} />

      {/* Typeface Roles */}
      <div style={{ fontWeight: 700, fontSize: 11, margin: '8px 0 4px', color: '#444' }}>
        🔤 Typeface Roles
      </div>
      {TYPEFACE_ROLES.map(role => (
        <div key={role} data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>
            {TYPEFACE_ROLE_LABELS[role].split(' /')[0]}
          </span>
          <select
            data-part={TypographyPaletteParts.select}
            value={typefaceRoles[role]}
            onChange={(e) => handleRoleChange(role, e.target.value as FontFamily)}
          >
            {Object.entries(FONT_FAMILY_LABELS).map(([value, label]) => (
              <option key={value} value={value}>{label}</option>
            ))}
          </select>
        </div>
      ))}
    </div>
  );
}
