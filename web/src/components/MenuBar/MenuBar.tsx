// MenuBar.tsx — retro menu bar: Apple logo, File, Edit, View, Help, app title.
import { MenuBarParts } from './parts';
import './styles/menubar.css';

const ITEMS = ['File', 'Edit', 'View', 'Help'];

interface MenuBarProps {
  title?: string;
}

export function MenuBar({ title = 'Glazed Help Browser' }: MenuBarProps) {
  return (
    <div data-part={MenuBarParts.root} role="menubar" aria-label="Main menu">
      <span data-part="menubar-apple" aria-hidden="true">
        &#xF8FF;{/* Apple logo */}
      </span>
      {ITEMS.map((item) => (
        <span key={item} data-part={MenuBarParts.item} role="menuitem">
          {item}
        </span>
      ))}
      <span data-part="menubar-title">{title}</span>
    </div>
  );
}
