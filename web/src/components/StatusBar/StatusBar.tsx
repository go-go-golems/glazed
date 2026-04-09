// StatusBar.tsx — bottom bar showing section count and version.
import { StatusBarParts } from './parts';
import './styles/statusbar.css';

interface StatusBarProps {
  count: number;
  version?: string;
}

export function StatusBar({ count, version = 'v0.1' }: StatusBarProps) {
  return (
    <div data-part={StatusBarParts.root} aria-label="Status bar">
      <span data-part={StatusBarParts.count}>
        {count} section{count !== 1 ? 's' : ''}
      </span>
      <span data-part={StatusBarParts.version}>{version}</span>
    </div>
  );
}
