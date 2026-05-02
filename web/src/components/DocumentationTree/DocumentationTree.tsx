import { useEffect, useMemo, useState } from 'react';
import type { SectionSummary } from '../../types';
import { DocumentationTreeParts } from './parts';
import { buildDocumentationTree, defaultExpandedGroupIds, filterDocumentationTree } from './tree';
import './styles/documentation-tree.css';

interface DocumentationTreeProps {
  sections: SectionSummary[];
  search: string;
  activeSlug: string | null;
  activeHeadingId?: string;
  onSelectDocument: (slug: string) => void;
  onSelectHeading: (slug: string, headingId: string) => void;
}

export function DocumentationTree({
  sections,
  search,
  activeSlug,
  activeHeadingId,
  onSelectDocument,
  onSelectHeading,
}: DocumentationTreeProps) {
  const baseTree = useMemo(() => buildDocumentationTree(sections), [sections]);
  const tree = useMemo(() => filterDocumentationTree(baseTree, search), [baseTree, search]);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(() => new Set(defaultExpandedGroupIds(baseTree)));
  const [expandedSections, setExpandedSections] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    setExpandedGroups(new Set(defaultExpandedGroupIds(baseTree)));
  }, [baseTree]);

  useEffect(() => {
    if (!activeSlug) return;
    setExpandedSections((prev) => new Set(prev).add(activeSlug));
  }, [activeSlug]);

  if (tree.length === 0) {
    return <div data-part={DocumentationTreeParts.empty}>No documentation matches.</div>;
  }

  const toggleGroup = (id: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const toggleSection = (slug: string) => {
    setExpandedSections((prev) => {
      const next = new Set(prev);
      next.has(slug) ? next.delete(slug) : next.add(slug);
      return next;
    });
  };

  return (
    <div data-part={DocumentationTreeParts.root} role="tree" aria-label="Documentation tree">
      {tree.map((group) => {
        const groupExpanded = expandedGroups.has(group.id);
        return (
          <div key={group.id} data-part={DocumentationTreeParts.group}>
            <button
              type="button"
              data-part={DocumentationTreeParts.row}
              data-kind="group"
              role="treeitem"
              aria-expanded={groupExpanded}
              onClick={() => toggleGroup(group.id)}
            >
              <span data-part={DocumentationTreeParts.disclosure}>{groupExpanded ? '▾' : '▸'}</span>
              <span data-part={DocumentationTreeParts.icon}>{group.icon}</span>
              <span data-part={DocumentationTreeParts.label}>{group.label}</span>
            </button>
            {groupExpanded && (
              <div data-part={DocumentationTreeParts.children} role="group">
                {group.sections.map((section) => {
                  const sectionExpanded = expandedSections.has(section.slug);
                  const hasHeadings = section.headings.length > 0;
                  return (
                    <div key={section.slug}>
                      <button
                        type="button"
                        data-part={DocumentationTreeParts.row}
                        data-kind="document"
                        role="treeitem"
                        aria-expanded={hasHeadings ? sectionExpanded : undefined}
                        aria-selected={activeSlug === section.slug}
                        onClick={() => onSelectDocument(section.slug)}
                        onDoubleClick={() => hasHeadings && toggleSection(section.slug)}
                      >
                        <span
                          data-part={DocumentationTreeParts.disclosure}
                          onClick={(event) => {
                            if (!hasHeadings) return;
                            event.stopPropagation();
                            toggleSection(section.slug);
                          }}
                        >
                          {hasHeadings ? (sectionExpanded ? '▾' : '▸') : ''}
                        </span>
                        <span data-part={DocumentationTreeParts.icon}>📄</span>
                        <span data-part={DocumentationTreeParts.label}>{section.title}</span>
                      </button>
                      {sectionExpanded && hasHeadings && (
                        <div data-part={DocumentationTreeParts.children} role="group">
                          {section.headings.map((heading) => (
                            <button
                              type="button"
                              key={heading.id}
                              data-part={DocumentationTreeParts.row}
                              data-kind="heading"
                              data-level={heading.level}
                              role="treeitem"
                              aria-selected={activeSlug === section.slug && activeHeadingId === heading.id}
                              onClick={() => onSelectHeading(section.slug, heading.id)}
                            >
                              <span data-part={DocumentationTreeParts.disclosure} aria-hidden="true" />
                              <span data-part={DocumentationTreeParts.icon} aria-hidden="true">#</span>
                              <span data-part={DocumentationTreeParts.label}>{heading.text}</span>
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
