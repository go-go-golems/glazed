// components/TypographyPalette/FontSizeStepper.tsx
// Stepper control for font size (px or em) and line-height.

import { TypographyPaletteParts } from './parts';

interface FontSizeStepperProps {
  value: number;
  unit?: string;     // 'px', 'em', or '' for unitless
  step?: number;     // step increment (default 1 for px, 0.1 for em)
  min?: number;
  max?: number;
  onChange: (value: number) => void;
}

export function FontSizeStepper({
  value,
  unit = 'px',
  step,
  min,
  max,
  onChange,
}: FontSizeStepperProps) {
  const effectiveStep = step ?? (unit === 'em' || unit === '' ? 0.1 : 1);
  const effectiveMin = min ?? (unit === 'em' ? 0.5 : 6);
  const effectiveMax = max ?? (unit === '' ? 4 : 72);  // '' = line-height

  const decrement = () => onChange(Math.max(effectiveMin, +(value - effectiveStep).toFixed(2)));
  const increment = () => onChange(Math.min(effectiveMax, +(value + effectiveStep).toFixed(2)));

  const displayUnit = unit === 'em' ? 'em' : unit === '' ? '' : 'px';

  return (
    <div data-part={TypographyPaletteParts.stepper}>
      <button
        data-part={TypographyPaletteParts.stepperBtn}
        onClick={decrement}
        tabIndex={-1}
      >
        −
      </button>
      <span data-part={TypographyPaletteParts.stepperValue}>
        {value}{displayUnit}
      </span>
      <button
        data-part={TypographyPaletteParts.stepperBtn}
        onClick={increment}
        tabIndex={-1}
      >
        +
      </button>
    </div>
  );
}
