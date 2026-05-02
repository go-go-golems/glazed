import { NavigationModeToggleParts } from './parts';
import './styles/navigation-mode-toggle.css';

export type NavigationMode = 'tree' | 'search';

interface NavigationModeToggleProps {
  value: NavigationMode;
  onChange: (mode: NavigationMode) => void;
}

export function NavigationModeToggle({ value, onChange }: NavigationModeToggleProps) {
  return (
    <div data-part={NavigationModeToggleParts.root}>
      <span data-part={NavigationModeToggleParts.label}>Navigation</span>
      <div data-part={NavigationModeToggleParts.buttons} role="group" aria-label="Navigation mode">
        <button
          data-part={NavigationModeToggleParts.button}
          aria-pressed={value === 'tree'}
          onClick={() => onChange('tree')}
          type="button"
        >
          Tree
        </button>
        <button
          data-part={NavigationModeToggleParts.button}
          aria-pressed={value === 'search'}
          onClick={() => onChange('search')}
          type="button"
        >
          Search
        </button>
      </div>
    </div>
  );
}
