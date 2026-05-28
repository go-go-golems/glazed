// components/TypographyPalette/FontWeightSelect.tsx

import type { FontWeight } from '../../types/typography-palette';
import { FONT_WEIGHTS } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';

interface FontWeightSelectProps {
  value: FontWeight;
  onChange: (value: FontWeight) => void;
}

const WEIGHT_LABELS: Record<FontWeight, string> = {
  100: 'Thin (100)',
  200: 'Extra Light (200)',
  300: 'Light (300)',
  400: 'Regular (400)',
  500: 'Medium (500)',
  600: 'Semi Bold (600)',
  700: 'Bold (700)',
  800: 'Extra Bold (800)',
  900: 'Black (900)',
};

export function FontWeightSelect({ value, onChange }: FontWeightSelectProps) {
  return (
    <select
      data-part={TypographyPaletteParts.select}
      value={value}
      onChange={(e) => onChange(Number(e.target.value) as FontWeight)}
    >
      {FONT_WEIGHTS.map((w) => (
        <option key={w} value={w}>{WEIGHT_LABELS[w]}</option>
      ))}
    </select>
  );
}
