// components/TypographyPalette/usePaletteShortcut.ts
// Registers Ctrl+Shift+T (or Cmd+Shift+T on macOS) to toggle the palette.
// Only active in dev mode.

import { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { togglePalette } from '../../store/typographyPaletteSlice';

export function usePaletteShortcut(): void {
  const dispatch = useDispatch();

  useEffect(() => {
    if (!import.meta.env.DEV) return;

    const handler = (e: KeyboardEvent) => {
      if (e.key === 'T' && (e.ctrlKey || e.metaKey) && e.shiftKey) {
        e.preventDefault();
        dispatch(togglePalette());
      }
    };

    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [dispatch]);
}
