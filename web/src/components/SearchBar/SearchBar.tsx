// SearchBar.tsx — search input with icon.
import { SearchBarParts } from './parts';
import './styles/searchbar.css';

interface SearchBarProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function SearchBar({ value, onChange, placeholder = 'Search\u2026' }: SearchBarProps) {
  return (
    <div data-part={SearchBarParts.root}>
      <span data-part="searchbar-icon" aria-hidden="true">
        &#x1F50D;{/* magnifying glass */}
      </span>
      <input
        data-part={SearchBarParts.input}
        type="search"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        aria-label="Search sections"
      />
    </div>
  );
}
