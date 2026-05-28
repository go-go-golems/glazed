// components/TypographyPalette/TypographyPaletteGroup.tsx
// Accordion group that expands to show element controls.

import { useDispatch } from 'react-redux';
import type { TypographyGroup } from '../../types/typography-palette';
import { setActiveGroup } from '../../store/typographyPaletteSlice';
import { TypographyPaletteParts } from './parts';
import { TypographyPaletteElement } from './TypographyPaletteElement';

interface TypographyPaletteGroupProps {
  group: TypographyGroup;
  isExpanded: boolean;
  onToggle: () => void;
}

export function TypographyPaletteGroup({
  group,
  isExpanded,
  onToggle,
}: TypographyPaletteGroupProps) {
  return (
    <div data-part={TypographyPaletteParts.group}>
      <button
        data-part={TypographyPaletteParts.groupHeader}
        onClick={onToggle}
      >
        <span style={{ marginRight: 6, fontSize: 10 }}>
          {isExpanded ? '▾' : '▸'}
        </span>
        {group.label}
      </button>
      {isExpanded && (
        <div data-part={TypographyPaletteParts.groupBody}>
          {group.elements.map((elem) => (
            <TypographyPaletteElement
              key={elem.id}
              element={elem}
              currentOverrides={undefined}
            />
          ))}
        </div>
      )}
    </div>
  );
}
