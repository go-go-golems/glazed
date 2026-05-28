// components/TypographyPalette/TypographyPaletteElement.tsx
// Renders the control rows for a single adjustable element.
// Supports both 'custom' mode (absolute values) and 'scale' mode
// (derived from the baseline design system).

import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { setOverride, setElementMode, setElementScaleSteps } from '../../store/typographyPaletteSlice';
import type {
  TypographyElement, TypographyProperties, FontFamily, GrayColor, FontWeight,
  ElementSizeMode, ScaleStep, ElementScaleSteps,
} from '../../types/typography-palette';
import { SCALE_RATIOS, computeScaledValue } from '../../types/typography-palette';
import { TypographyPaletteParts } from './parts';
import { FontFamilySelect } from './FontFamilySelect';
import { FontSizeStepper } from './FontSizeStepper';
import { FontWeightSelect } from './FontWeightSelect';
import { ColorStepper } from './ColorStepper';
import { ScaleStepSelect } from './ScaleStepSelect';
import { useHighlightToggle } from './useHighlightElement';

interface TypographyPaletteElementProps {
  element: TypographyElement;
  currentOverrides: TypographyProperties | undefined;
}

export function TypographyPaletteElement({
  element,
  currentOverrides,
}: TypographyPaletteElementProps) {
  const dispatch = useDispatch();
  const baseline = useSelector((s: RootState) => s.typographyPalette.baseline);
  const elementMode = useSelector(
    (s: RootState) => s.typographyPalette.elementModes[element.id] ?? 'custom'
  ) as ElementSizeMode;
  const elementSteps = useSelector(
    (s: RootState) => s.typographyPalette.elementScaleSteps[element.id] ?? {}
  ) as ElementScaleSteps;
  const highlightedElementId = useSelector(
    (s: RootState) => s.typographyPalette.highlightedElementId
  );
  const toggleHighlight = useHighlightToggle();

  const effective: TypographyProperties = { ...element.defaults, ...currentOverrides };

  const handleChange = (properties: TypographyProperties) => {
    dispatch(setOverride({ elementId: element.id, properties }));
  };

  const handleModeChange = (mode: ElementSizeMode) => {
    dispatch(setElementMode({ elementId: element.id, mode }));
  };

  const handleStepsChange = (steps: ElementScaleSteps) => {
    dispatch(setElementScaleSteps({ elementId: element.id, steps }));
  };

  const isScaleMode = element.supportsScale && elementMode === 'scale';
  const ratio = SCALE_RATIOS[baseline.scaleRatioName];

  return (
    <div data-part={TypographyPaletteParts.element}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 2 }}>
        <span data-part={TypographyPaletteParts.elementLabel}>
          {element.label}
        </span>
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          {/* Highlight button */}
          <button
            title="Highlight this element on the page"
            onClick={() => toggleHighlight(element.id)}
            style={{
              border: 'none',
              padding: '1px 4px',
              fontSize: 9,
              cursor: 'pointer',
              background: highlightedElementId === element.id ? '#ff6600' : 'transparent',
              color: highlightedElementId === element.id ? '#fff' : '#888',
              borderRadius: 2,
              fontFamily: 'inherit',
            }}
          >
            🔍
          </button>
          {element.supportsScale && (
          <div style={{ display: 'inline-flex', border: '1px solid #aaa', borderRadius: 2, overflow: 'hidden' }}>
            <button
              style={{
                border: 'none',
                padding: '1px 6px',
                fontSize: 9,
                fontWeight: elementMode === 'custom' ? 700 : 400,
                background: elementMode === 'custom' ? '#000' : '#fff',
                color: elementMode === 'custom' ? '#fff' : '#000',
                cursor: 'pointer',
                fontFamily: 'inherit',
              }}
              onClick={() => handleModeChange('custom')}
            >
              Custom
            </button>
            <button
              style={{
                border: 'none',
                borderLeft: '1px solid #aaa',
                padding: '1px 6px',
                fontSize: 9,
                fontWeight: elementMode === 'scale' ? 700 : 400,
                background: elementMode === 'scale' ? '#000' : '#fff',
                color: elementMode === 'scale' ? '#fff' : '#000',
                cursor: 'pointer',
                fontFamily: 'inherit',
              }}
              onClick={() => handleModeChange('scale')}
            >
              Scale
            </button>
          </div>
        )}
        </div>
      </div>

      {/* Font family — always custom (not scale-derived) */}
      {element.adjustable.includes('fontFamily') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Font</span>
          <FontFamilySelect
            value={(effective.fontFamily ?? 'ui') as FontFamily}
            onChange={(v) => handleChange({ fontFamily: v })}
          />
        </div>
      )}

      {/* Font size — custom or scale mode */}
      {element.adjustable.includes('fontSize') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Size</span>
          {isScaleMode ? (
            <ScaleStepSelect
              value={elementSteps.fontSizeStep ?? element.defaultFontSizeStep ?? 0}
              baseValue={baseline.baseFontSize}
              ratioName={baseline.scaleRatioName}
              unit={effective.fontSizeUnit || 'px'}
              onChange={(step) => handleStepsChange({ fontSizeStep: step })}
            />
          ) : (
            <FontSizeStepper
              value={effective.fontSize ?? 13}
              unit={effective.fontSizeUnit || 'px'}
              onChange={(v) => handleChange({ fontSize: v })}
            />
          )}
        </div>
      )}

      {/* Font weight — always custom */}
      {element.adjustable.includes('fontWeight') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Weight</span>
          <FontWeightSelect
            value={(effective.fontWeight ?? 400) as FontWeight}
            onChange={(v) => handleChange({ fontWeight: v })}
          />
        </div>
      )}

      {/* Color — always custom */}
      {element.adjustable.includes('color') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>Color</span>
          <ColorStepper
            value={(effective.color ?? '#000') as GrayColor}
            onChange={(v) => handleChange({ color: v })}
          />
        </div>
      )}

      {/* Line height — custom or scale mode */}
      {element.adjustable.includes('lineHeight') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>L-Height</span>
          {isScaleMode ? (
            <FontSizeStepper
              value={baseline.baseLineHeight + (elementSteps.lineHeightStep ?? element.defaultLineHeightStep ?? 0) * 0.1}
              unit=""
              step={0.1}
              min={0.5}
              max={3}
              onChange={(v) => {
                // Store as offset from base line height
                const offset = Math.round((v - baseline.baseLineHeight) * 10);
                handleStepsChange({ lineHeightStep: offset as ScaleStep });
              }}
            />
          ) : (
            <FontSizeStepper
              value={effective.lineHeight ?? 1.6}
              unit=""
              step={0.1}
              min={0.8}
              max={3}
              onChange={(v) => handleChange({ lineHeight: v })}
            />
          )}
        </div>
      )}

      {/* Letter spacing — custom or scale mode */}
      {element.adjustable.includes('letterSpacing') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>L-Space</span>
          <FontSizeStepper
            value={effective.letterSpacing ?? 0}
            unit="em"
            step={0.01}
            min={-0.1}
            max={0.5}
            onChange={(v) => handleChange({ letterSpacing: v })}
          />
        </div>
      )}

      {/* Word spacing — custom or scale mode */}
      {element.adjustable.includes('wordSpacing') && (
        <div data-part={TypographyPaletteParts.controlRow}>
          <span data-part={TypographyPaletteParts.controlLabel}>W-Space</span>
          <FontSizeStepper
            value={effective.wordSpacing ?? 0}
            unit="em"
            step={0.01}
            min={-0.2}
            max={1}
            onChange={(v) => handleChange({ wordSpacing: v })}
          />
        </div>
      )}
    </div>
  );
}
