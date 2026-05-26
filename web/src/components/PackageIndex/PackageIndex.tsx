// PackageIndex.tsx — index page showing all sections for a package/version,
// grouped by type. Shown when the user navigates to /{package}/{version}
// without selecting a specific section.
import type { SectionSummary, SectionType } from '../../types';
import { PackageIndexParts } from './parts';
import './styles/package-index.css';

interface PackageIndexProps {
  sections: SectionSummary[];
  onSelect: (slug: string) => void;
}

const GROUPS: Array<{ type: SectionType; label: string; icon: string }> = [
  { type: 'GeneralTopic', label: 'Topics', icon: '📖' },
  { type: 'Example', label: 'Examples', icon: '📄' },
  { type: 'Application', label: 'Applications', icon: '📦' },
  { type: 'Tutorial', label: 'Tutorials', icon: '🎓' },
];

export function PackageIndex({ sections, onSelect }: PackageIndexProps) {
  const grouped = GROUPS.map((group) => ({
    ...group,
    sections: sections
      .filter((s) => s.type === group.type)
      .sort((a, b) => a.title.localeCompare(b.title)),
  })).filter((g) => g.sections.length > 0);

  return (
    <div data-part={PackageIndexParts.root}>
      <h1 data-part={PackageIndexParts.heading}>Documentation Index</h1>
      <p data-part={PackageIndexParts.count}>
        {sections.length} section{sections.length !== 1 ? 's' : ''} available
      </p>
      {grouped.map((group) => (
        <div key={group.type} data-part={PackageIndexParts.group}>
          <h2 data-part={PackageIndexParts.groupHeading}>
            {group.icon} {group.label}
          </h2>
          {group.sections.map((section) => (
            <div key={section.slug} data-part={PackageIndexParts.sectionItem}>
              <a
                data-part={PackageIndexParts.sectionLink}
                href="#"
                onClick={(e) => {
                  e.preventDefault();
                  onSelect(section.slug);
                }}
              >
                {section.title}
              </a>
              {section.short && (
                <p data-part={PackageIndexParts.sectionShort}>{section.short}</p>
              )}
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}
