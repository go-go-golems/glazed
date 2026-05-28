// StatusBar.tsx — bottom bar showing section count and version.
// Includes a dev-only typography palette toggle button.

import { useDispatch } from 'react-redux';
import { togglePalette } from '../../store/typographyPaletteSlice';
import { StatusBarParts } from './parts';
import './styles/statusbar.css';

interface StatusBarProps {
  count: number;
  version?: string;
}

export function StatusBar({ count, version = 'v0.1' }: StatusBarProps) {
  const dispatch = useDispatch();

  return (
    <div data-part={StatusBarParts.root} aria-label="Status bar">
      <span data-part={StatusBarParts.count}>
        {count} section{count !== 1 ? 's' : ''}
      </span>
      <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
        {import.meta.env.DEV && (
          <button
            onClick={() => dispatch(togglePalette())}
            title="Typography Debug Palette (Ctrl+Shift+T)"
            style={{
              background: 'none',
              border: '1px solid #999',
              padding: '0 4px',
              cursor: 'pointer',
              fontSize: 10,
              fontFamily: 'serif',
              lineHeight: 1.2,
              color: '#777',
            }}
          >
            𝒜a
          </button>
        )}
        <span data-part={StatusBarParts.version}>{version}</span>
      </span>
    </div>
  );
}
