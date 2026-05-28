// components/TypographyPalette/TypographyPaletteGroup.tsx
// Accordion group that expands to show element controls.

import { useDispatch } from 'react-redux';
import type { TypographyGroup, TypographyOverrides, TypographyProperties } from '../../types/typography-palette';
import { setOverride } from '../../store/typographyPaletteSlice';
import { TypographyPaletteParts } from './parts';
import { TypographyPaletteElement } from './TypographyPaletteElement';

interface TypographyPaletteGroupProps {
  group: TypographyGroup;
  isExpanded: boolean;
  overrides: TypographyOverrides;
  onToggle: () => void;
}

export function TypographyPaletteGroup({
  group,
  isExpanded,
  overrides,
  onToggle,
}: TypographyPaletteGroupProps) {
  const dispatch = useDispatch();

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
              currentOverrides={overrides[elem.id]}
              onChange={(properties: TypographyProperties) =>
                dispatch(setOverride({ elementId: elem.id, properties }))
              }
            />
          ))}
        </div>
      )}
    </div>
  );
}
