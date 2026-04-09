// SectionHeader.tsx — title, slug, short description, and tag badges.
import type { SectionDetail } from '../../types';
import { Badge } from '../Badge/Badge';
import { SectionHeaderParts } from './parts';
import './styles/section-view.css';

interface SectionHeaderProps {
  section: SectionDetail;
}

export function SectionHeader({ section }: SectionHeaderProps) {
  const visibleFlags = section.flags.slice(0, 4);
  const extraFlags = section.flags.length - visibleFlags.length;

  return (
    <div data-part={SectionHeaderParts.root}>
      <div>
        <span data-part={SectionHeaderParts.slug}>{section.slug}</span>
      </div>
      <h1 data-part={SectionHeaderParts.heading}>{section.title}</h1>
      <p data-part={SectionHeaderParts.subtitle}>{section.short}</p>
      <div data-part={SectionHeaderParts.tags}>
        <Badge text={section.type} variant="type" />
        {section.topics.map((t) => (
          <Badge key={t} text={t} variant="topic" />
        ))}
        {section.commands.map((c) => (
          <Badge key={c} text={c} variant="command" />
        ))}
        {visibleFlags.map((f) => (
          <Badge key={f} text={f} variant="flag" />
        ))}
        {extraFlags > 0 && (
          <span style={{ fontSize: 10, color: '#999' }}>+{extraFlags}</span>
        )}
      </div>
    </div>
  );
}
