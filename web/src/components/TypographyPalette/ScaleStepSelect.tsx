// components/TypographyPalette/ScaleStepSelect.tsx
// Dropdown selector for scale steps (−3 to +6).
// Shows the computed value next to each step label.

import { SCALE_STEPS, SCALE_STEP_LABELS, computeScaledValue } from '../../types/typography-palette';
import type { ScaleStep, ScaleRatioName } from '../../types/typography-palette';
import { SCALE_RATIOS } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';

interface ScaleStepSelectProps {
  value: ScaleStep;
  baseValue: number;
  ratioName: ScaleRatioName;
  unit?: string;  // 'px', 'em', or '' for line-height
  onChange: (step: ScaleStep) => void;
}

export function ScaleStepSelect({
  value,
  baseValue,
  ratioName,
  unit = 'px',
  onChange,
}: ScaleStepSelectProps) {
  const ratio = SCALE_RATIOS[ratioName];
  const computed = computeScaledValue(baseValue, ratio, value);

  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
      <select
        data-part={TypographyPaletteParts.select}
        value={value}
        onChange={(e) => onChange(Number(e.target.value) as ScaleStep)}
        style={{ flex: 1 }}
      >
        {SCALE_STEPS.map(step => (
          <option key={step} value={step}>
            {SCALE_STEP_LABELS[step]} → {computeScaledValue(baseValue, ratio, step)}{unit}
          </option>
        ))}
      </select>
      <span style={{ fontSize: 9, color: '#999', fontFamily: 'var(--font-mono)', minWidth: 36, textAlign: 'right' }}>
        {computed}{unit}
      </span>
    </div>
  );
}
