// components/TypographyPalette/useTypographyOverrides.ts
// React hook that syncs Redux typography overrides to the DOM.
// Whenever the overrides state changes, this hook re-generates and
// injects the CSS rules into the page.

import { useEffect } from 'react';
import { useSelector } from 'react-redux';
import type { RootState } from '../../store';
import { applyOverrides, clearOverrides } from './css-override-engine';

export function useTypographyOverrides(): void {
  const overrides = useSelector((state: RootState) => state.typographyPalette.overrides);

  useEffect(() => {
    if (Object.keys(overrides).length === 0) {
      clearOverrides();
    } else {
      applyOverrides(overrides);
    }
  }, [overrides]);
}
