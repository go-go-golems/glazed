// components/TypographyPalette/TypographyPaletteElement.tsx
// Renders the control rows for a single adjustable element.

import type { TypographyElement, TypographyProperties, FontFamily, GrayColor, FontWeight } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';
import { FontFamilySelect } from './FontFamilySelect';
import { FontSizeStepper } from './FontSizeStepper';
import { FontWeightSelect } from './FontWeightSelect';
import { ColorStepper } from './ColorStepper';

interface TypographyPaletteElementProps {
  element: TypographyElement;
  currentOverrides: TypographyProperties | undefined;
  onChange: (properties: TypographyProperties) => void;
}

export function TypographyPaletteElement({
  element,
  currentOverrides,
  onChange,
}: TypographyPaletteElementProps) {
  const effective: TypographyProperties = { ...element.defaults, ...currentOverrides };

  return (
    <div data-part={TypographyPaletteParts.element}>
      <span data-part={TypographyPaletteParts.elementLabel}>
        {element.label}
      </span>

      {element.adjustable.includes('fontFamily') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Font</span>
          <FontFamilySelect
            value={(effective.fontFamily ?? 'ui') as FontFamily}
            onChange={(v) => onChange({ fontFamily: v })}
          />
        </div>
      )}

      {element.adjustable.includes('fontSize') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Size</span>
          <FontSizeStepper
            value={effective.fontSize ?? 13}
            unit={effective.fontSizeUnit || 'px'}
            onChange={(v) => onChange({ fontSize: v })}
          />
        </div>
      )}

      {element.adjustable.includes('fontWeight') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Weight</span>
          <FontWeightSelect
            value={(effective.fontWeight ?? 400) as FontWeight}
            onChange={(v) => onChange({ fontWeight: v })}
          />
        </div>
      )}

      {element.adjustable.includes('color') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Color</span>
          <ColorStepper
            value={(effective.color ?? '#000') as GrayColor}
            onChange={(v) => onChange({ color: v })}
          />
        </div>
      )}

      {element.adjustable.includes('lineHeight') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Height</span>
          <FontSizeStepper
            value={effective.lineHeight ?? 1.6}
            unit=""
            step={0.1}
            min={0.8}
            max={3}
            onChange={(v) => onChange({ lineHeight: v })}
          />
        </div>
      )}
    </div>
  );
}
