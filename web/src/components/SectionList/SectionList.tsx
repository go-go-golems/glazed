// SectionList.tsx — scrollable list of SectionCard items.
import type { SectionSummary } from '../../types';
import { SectionCard } from './SectionCard';
import { SectionListParts } from './parts';
import './styles/section-list.css';

interface SectionListProps {
  sections: SectionSummary[];
  activeSlug: string | null;
  onSelect: (slug: string) => void;
}

export function SectionList({ sections, activeSlug, onSelect }: SectionListProps) {
  return (
    <div data-part={SectionListParts.root} role="listbox" aria-label="Sections">
      {sections.map((section) => (
        <SectionCard
          key={section.slug}
          section={section}
          isActive={activeSlug === section.slug}
          onClick={() => onSelect(section.slug)}
        />
      ))}
    </div>
  );
}
