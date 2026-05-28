// components/TypographyPalette/useTypographyOverrides.ts
// React hook that syncs Redux typography overrides + scale-mode computed
// values to the DOM. Whenever overrides, baseline, or element modes change,
// this hook resolves all scale-mode elements to concrete CSS properties
// and injects the rules into the page.

import { useEffect, useMemo } from 'react';
import { useSelector } from 'react-redux';
import type { RootState } from '../../store';
import type { TypographyOverrides, TypographyProperties, ScaleStep, ElementScaleSteps, TypefaceRoleMap } from '../../types/typography-palette';
import { SCALE_RATIOS, computeScaledValue } from '../../types/typography-palette';
import { applyOverrides, clearOverrides } from './css-override-engine';
import { buildElementMap } from './element-registry';

/**
 * Resolve scale-mode elements and typeface role assignments.
 * Priority (highest to lowest):
 *   1. Per-element custom overrides (explicit fontFamily set by user)
 *   2. Typeface role assignment (element inherits from its role)
 *   3. Scale-mode computed values (size, line-height from baseline)
 *   4. Element defaults
 */
function resolveScaleOverrides(
  baseline: { baseFontSize: number; scaleRatioName: string; baseLineHeight: number; baseLetterSpacing: number; baseWordSpacing: number },
  elementModes: Record<string, string>,
  elementScaleSteps: Record<string, ElementScaleSteps>,
  customOverrides: TypographyOverrides,
  typefaceRoles: TypefaceRoleMap,
): TypographyOverrides {
  const elementMap = buildElementMap();
  const ratio = SCALE_RATIOS[baseline.scaleRatioName];
  const resolved: TypographyOverrides = {};

  for (const [elementId, mode] of Object.entries(elementModes)) {
    if (mode !== 'scale') continue;

    const elem = elementMap.get(elementId);
    if (!elem) continue;

    const steps = elementScaleSteps[elementId] ?? {};
    const props: TypographyProperties = {};

    // Font size: compute from baseline × ratio^step
    if (elem.adjustable.includes('fontSize')) {
      const step: ScaleStep = steps.fontSizeStep ?? elem.defaultFontSizeStep ?? 0;
      const unit = elem.defaults.fontSizeUnit || 'px';
      if (unit === 'em') {
        // For em-based elements, compute as multiplier
        // We store the em value as the ratio step result divided by baseFontSize
        props.fontSize = computeScaledValue(1, ratio, step);
        props.fontSizeUnit = 'em';
      } else {
        props.fontSize = computeScaledValue(baseline.baseFontSize, ratio, step);
        props.fontSizeUnit = 'px';
      }
    }

    // Line height: offset from base line height
    if (elem.adjustable.includes('lineHeight')) {
      const lhStep: ScaleStep = steps.lineHeightStep ?? elem.defaultLineHeightStep ?? 0;
      props.lineHeight = +(baseline.baseLineHeight + lhStep * 0.1).toFixed(2);
    }

    // Letter spacing: use baseline value
    if (elem.adjustable.includes('letterSpacing') && baseline.baseLetterSpacing !== 0) {
      props.letterSpacing = baseline.baseLetterSpacing;
    }

    // Word spacing: use baseline value
    if (elem.adjustable.includes('wordSpacing') && baseline.baseWordSpacing !== 0) {
      props.wordSpacing = baseline.baseWordSpacing;
    }

    if (Object.keys(props).length > 0) {
      resolved[elementId] = props;
    }
  }

  // Apply typeface role fontFamily to every element that has a typefaceRole.
  // This runs regardless of whether fontFamily is in the adjustable list,
  // because the role system is a higher-level mechanism than per-element control.
  // Per-element custom overrides still win when set.
  for (const [elementId, elem] of elementMap.entries()) {
    const roleFont = typefaceRoles[elem.typefaceRole];
    if (!roleFont) continue;
    // If custom override already sets fontFamily, skip
    if (customOverrides[elementId]?.fontFamily) continue;
    if (!resolved[elementId]) resolved[elementId] = {};
    resolved[elementId].fontFamily = roleFont;
  }

  // Merge custom overrides on top (custom takes precedence for properties it sets)
  for (const [elementId, customProps] of Object.entries(customOverrides)) {
    if (resolved[elementId]) {
      resolved[elementId] = { ...resolved[elementId], ...customProps };
    } else {
      resolved[elementId] = { ...customProps };
    }
  }

  return resolved;
}

export function useTypographyOverrides(): void {
  const customOverrides = useSelector((s: RootState) => s.typographyPalette.overrides);
  const baseline = useSelector((s: RootState) => s.typographyPalette.baseline);
  const elementModes = useSelector((s: RootState) => s.typographyPalette.elementModes);
  const elementScaleSteps = useSelector((s: RootState) => s.typographyPalette.elementScaleSteps);
  const typefaceRoles = useSelector((s: RootState) => s.typographyPalette.typefaceRoles);

  // Resolve scale-mode elements + merge with custom overrides + apply typeface roles
  const resolvedOverrides = useMemo(() => {
    const hasScaleModes = Object.values(elementModes).some(m => m === 'scale');
    const hasRoleOverrides = true; // typeface roles always apply
    if (!hasScaleModes && !hasRoleOverrides) return customOverrides;
    return resolveScaleOverrides(baseline, elementModes, elementScaleSteps, customOverrides, typefaceRoles);
  }, [customOverrides, baseline, elementModes, elementScaleSteps, typefaceRoles]);

  useEffect(() => {
    if (Object.keys(resolvedOverrides).length === 0) {
      clearOverrides();
    } else {
      applyOverrides(resolvedOverrides);
    }
  }, [resolvedOverrides]);
}
