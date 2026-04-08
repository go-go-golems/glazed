// EmptyState.tsx — placeholder shown when no section is selected.
import { EmptyStateParts } from './parts';
import './styles/empty-state.css';

interface EmptyStateProps {
  label?: string;
}

export function EmptyState({ label = 'Select a section from the list.' }: EmptyStateProps) {
  return (
    <div data-part={EmptyStateParts.root}>
      <span data-part="empty-state-icon" aria-hidden="true">
        &#x1F4D6;{/* open book */}
      </span>
      <span data-part={EmptyStateParts.label}>{label}</span>
    </div>
  );
}
