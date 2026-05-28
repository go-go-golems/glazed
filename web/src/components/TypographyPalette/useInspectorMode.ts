// components/TypographyPalette/useInspectorMode.ts
// React hook that implements the inspector/dropper mode.
// When active, clicking on any element in the page:
//   1. Finds the matching palette element from the registry
//   2. Opens the relevant accordion group
//   3. Highlights the element on the page
//   4. Turns off inspector mode

import { useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { setActiveGroup, setHighlightedElement, toggleInspectorMode } from '../../store/typographyPaletteSlice';
import { TYPOGRAPHY_GROUPS } from './element-registry';

/**
 * Build a reverse lookup: CSS selector → { elementId, groupId }.
 * Computed once since the registry is static.
 */
function buildSelectorLookup(): Map<string, { elementId: string; groupId: string }> {
  const lookup = new Map<string, { elementId: string; groupId: string }>();
  for (const group of TYPOGRAPHY_GROUPS) {
    for (const elem of group.elements) {
      lookup.set(elem.selector, { elementId: elem.id, groupId: group.id });
    }
  }
  return lookup;
}

const SELECTOR_LOOKUP = buildSelectorLookup();

/**
 * Given a DOM element, find the matching palette element by walking up
 * the DOM tree and checking against registered selectors.
 */
function findMatchingElement(target: Element): { elementId: string; groupId: string } | null {
  // Walk from target up to root, testing each ancestor against all selectors
  let current: Element | null = target;
  while (current) {
    // First, try to match by data-part attribute
    const dataPart = current.getAttribute('data-part');
    if (dataPart) {
      // Check all multi-part selectors (e.g., [data-part~='section-card-title'])
      for (const [selector, info] of SELECTOR_LOOKUP.entries()) {
        if (current.matches(selector)) {
          return info;
        }
      }
    }
    // Also check class-based and compound selectors
    for (const [selector, info] of SELECTOR_LOOKUP.entries()) {
      if (current.matches(selector)) {
        return info;
      }
    }
    current = current.parentElement;
  }
  return null;
}

/** Hook that manages inspector mode click handling. */
export function useInspectorMode(): void {
  const dispatch = useDispatch();
  const inspectorMode = useSelector(
    (s: RootState) => s.typographyPalette.inspectorMode
  );

  useEffect(() => {
    if (!inspectorMode) return;

    const handleClick = (e: MouseEvent) => {
      // Don't capture clicks inside the palette itself
      const palette = (e.target as Element)?.closest?.('[data-part="typography-palette"]');
      if (palette) return;

      e.preventDefault();
      e.stopPropagation();

      const target = e.target as Element;
      const match = findMatchingElement(target);

      if (match) {
        dispatch(setActiveGroup(match.groupId));
        dispatch(setHighlightedElement(match.elementId));
      }

      // Turn off inspector mode after one click
      dispatch(toggleInspectorMode());
    };

    // Use capture phase to intercept before anything else
    document.addEventListener('click', handleClick, true);
    return () => document.removeEventListener('click', handleClick, true);
  }, [inspectorMode, dispatch]);
}

/** Hook that returns the inspector mode state and toggle. */
export function useInspectorToggle(): { active: boolean; toggle: () => void } {
  const dispatch = useDispatch();
  const active = useSelector(
    (s: RootState) => s.typographyPalette.inspectorMode
  );
  return {
    active,
    toggle: useCallback(() => dispatch(toggleInspectorMode()), [dispatch]),
  };
}
