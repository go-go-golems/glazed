// SectionCard.tsx — renders a section summary in the sidebar list.
import type { SectionSummary } from '../../types';
import { Badge } from '../Badge/Badge';
import { SectionCardParts, SectionListParts } from './parts';
import './styles/section-list.css';

interface SectionCardProps {
  section: SectionSummary;
  isActive: boolean;
  onClick: () => void;
}

export function SectionCard({ section, isActive, onClick }: SectionCardProps) {
  return (
    <button
      data-part={`${SectionListParts.item} ${SectionCardParts.root}`}
      aria-selected={isActive}
      onClick={onClick}
    >
      <div data-part={SectionCardParts.meta}>
        {!isActive && <Badge text={section.type} variant="type" />}
        {isActive && (
          <span style={{ fontSize: 10, fontWeight: 700, color: '#aaa' }}>
            {section.type === 'GeneralTopic' ? 'Topic'
              : section.type === 'Example' ? 'Example'
              : section.type === 'Application' ? 'App'
              : 'Tutorial'}
          </span>
        )}
        {section.topics.length > 0 && (
          <span data-part={SectionCardParts.topBadge}>&#9670; TOP</span>
        )}
      </div>
      <div data-part={SectionCardParts.title}>{section.title}</div>
      <div data-part={SectionCardParts.short}>{section.short}</div>
    </button>
  );
}
