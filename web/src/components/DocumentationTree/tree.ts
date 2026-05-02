import type { SectionHeading, SectionSummary, SectionType } from '../../types';

export interface DocumentationTreeSection {
  slug: string;
  title: string;
  type: SectionType;
  headings: SectionHeading[];
}

export interface DocumentationTreeGroup {
  id: SectionType;
  label: string;
  icon: string;
  sections: DocumentationTreeSection[];
}

const GROUPS: Array<Omit<DocumentationTreeGroup, 'sections'>> = [
  { id: 'GeneralTopic', label: 'Topics', icon: '📖' },
  { id: 'Example', label: 'Examples', icon: '📄' },
  { id: 'Application', label: 'Applications', icon: '📦' },
  { id: 'Tutorial', label: 'Tutorials', icon: '🎓' },
];

export function buildDocumentationTree(sections: SectionSummary[]): DocumentationTreeGroup[] {
  const byType = new Map<SectionType, DocumentationTreeGroup>();
  for (const group of GROUPS) {
    byType.set(group.id, { ...group, sections: [] });
  }

  for (const section of sections) {
    const type = section.type as SectionType;
    const group = byType.get(type);
    if (!group) continue;
    group.sections.push({
      slug: section.slug,
      title: section.title,
      type,
      headings: section.headings ?? [],
    });
  }

  for (const group of byType.values()) {
    group.sections.sort((a, b) => a.title.localeCompare(b.title));
  }

  return GROUPS
    .map((group) => byType.get(group.id)!)
    .filter((group) => group.sections.length > 0);
}

export function filterDocumentationTree(groups: DocumentationTreeGroup[], query: string): DocumentationTreeGroup[] {
  const q = query.trim().toLowerCase();
  if (!q) return groups;

  return groups
    .map((group) => ({
      ...group,
      sections: group.sections
        .map((section) => {
          const titleMatches = section.title.toLowerCase().includes(q);
          const headings = titleMatches
            ? section.headings
            : section.headings.filter((heading) => heading.text.toLowerCase().includes(q));
          return { ...section, headings };
        })
        .filter((section) => section.title.toLowerCase().includes(q) || section.headings.length > 0),
    }))
    .filter((group) => group.sections.length > 0);
}

export function defaultExpandedGroupIds(groups: DocumentationTreeGroup[]): string[] {
  return groups.map((group) => group.id);
}
