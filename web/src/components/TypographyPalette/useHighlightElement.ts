// components/TypographyPalette/useHighlightElement.ts
// React hook that highlights the DOM elements matching a given element ID
// by adding a temporary colored overlay. Used by the 🔍 highlight button
// and the inspector/dropper mode.

import { useEffect, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { setHighlightedElement } from '../../store/typographyPaletteSlice';
import { buildElementMap } from './element-registry';

const HIGHLIGHT_CLASS = 'typography-palette-highlight';
const HIGHLIGHT_STYLE_ID = 'typography-palette-highlight-style';

/** Inject the highlight CSS rule once. */
function ensureHighlightStyle() {
  if (document.getElementById(HIGHLIGHT_STYLE_ID)) return;
  const style = document.createElement('style');
  style.id = HIGHLIGHT_STYLE_ID;
  style.textContent = `
    .${HIGHLIGHT_CLASS} {
      outline: 2px solid #ff6600 !important;
      outline-offset: 2px !important;
      background-color: rgba(255, 102, 0, 0.08) !important;
      transition: outline 0.15s, background-color 0.15s;
    }
  `;
  document.head.appendChild(style);
}

/** Remove all highlight classes from the DOM. */
function clearAllHighlights() {
  document.querySelectorAll(`.${HIGHLIGHT_CLASS}`).forEach((el) => {
    el.classList.remove(HIGHLIGHT_CLASS);
  });
}

/** Add highlight class to all elements matching a selector. */
function applyHighlight(selector: string) {
  document.querySelectorAll(selector).forEach((el) => {
    el.classList.add(HIGHLIGHT_CLASS);
  });
}

/** Hook that syncs highlightedElementId to the DOM. */
export function useHighlightSync(): void {
  const highlightedElementId = useSelector(
    (s: RootState) => s.typographyPalette.highlightedElementId
  );

  useEffect(() => {
    ensureHighlightStyle();
    clearAllHighlights();

    if (highlightedElementId) {
      const elementMap = buildElementMap();
      const elem = elementMap.get(highlightedElementId);
      if (elem) {
        applyHighlight(elem.selector);
      }
    }

    return () => clearAllHighlights();
  }, [highlightedElementId]);
}

/** Hook that provides a toggle function for highlighting an element. */
export function useHighlightToggle(): (elementId: string) => void {
  const dispatch = useDispatch();
  const highlightedElementId = useSelector(
    (s: RootState) => s.typographyPalette.highlightedElementId
  );

  return useCallback(
    (elementId: string) => {
      dispatch(
        setHighlightedElement(
          highlightedElementId === elementId ? null : elementId
        )
      );
    },
    [dispatch, highlightedElementId]
  );
}
