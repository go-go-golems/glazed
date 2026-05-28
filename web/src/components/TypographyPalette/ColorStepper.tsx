// components/TypographyPalette/ColorStepper.tsx
// Stepper for grayscale color values (#000 through #fff).

import type { GrayColor } from '../../types/typography-palette';
import { GRAY_SHADES } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';

interface ColorStepperProps {
  value: GrayColor;
  onChange: (value: GrayColor) => void;
}

export function ColorStepper({ value, onChange }: ColorStepperProps) {
  const idx = GRAY_SHADES.indexOf(value);
  const safeIdx = idx === -1 ? 0 : idx;

  const decrement = () => {
    if (safeIdx > 0) onChange(GRAY_SHADES[safeIdx - 1]);
  };

  const increment = () => {
    if (safeIdx < GRAY_SHADES.length - 1) onChange(GRAY_SHADES[safeIdx + 1]);
  };

  return (
    <div data-part={TypographyPaletteParts.stepper}>
      <button
        data-part={TypographyPaletteParts.stepperBtn}
        onClick={decrement}
        disabled={safeIdx === 0}
        tabIndex={-1}
      >
        −
      </button>
      <span
        data-part={TypographyPaletteParts.stepperValue}
        style={{ color: value }}
      >
        {value}
      </span>
      <button
        data-part={TypographyPaletteParts.stepperBtn}
        onClick={increment}
        disabled={safeIdx === GRAY_SHADES.length - 1}
        tabIndex={-1}
      >
        +
      </button>
    </div>
  );
}
