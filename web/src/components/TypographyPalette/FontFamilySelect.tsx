// components/TypographyPalette/FontFamilySelect.tsx
import type { FontFamily } from '../../types/typography-palette';
import { FONT_FAMILY_LABELS } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';

interface FontFamilySelectProps {
  value: FontFamily;
  onChange: (value: FontFamily) => void;
}

export function FontFamilySelect({ value, onChange }: FontFamilySelectProps) {
  return (
    <select
      data-part={TypographyPaletteParts.select}
      value={value}
      onChange={(e) => onChange(e.target.value as FontFamily)}
    >
      {Object.entries(FONT_FAMILY_LABELS).map(([key, label]) => (
        <option key={key} value={key}>{label}</option>
      ))}
    </select>
  );
}
