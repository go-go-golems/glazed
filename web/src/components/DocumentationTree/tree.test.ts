import { describe, expect, it } from 'vitest';
import { buildDocumentationTree, filterDocumentationTree } from './tree';
import type { SectionSummary } from '../../types';

const sections: SectionSummary[] = [
  { id: 1, slug: 'topic', type: 'GeneralTopic', title: 'Alpha Topic', short: '', topics: [], isTopLevel: true, headings: [{ id: 'install', level: 2, text: 'Install' }] },
  { id: 2, slug: 'example', type: 'Example', title: 'Beta Example', short: '', topics: [], isTopLevel: false, headings: [{ id: 'run-it', level: 2, text: 'Run It' }] },
  { id: 3, slug: 'tutorial', type: 'Tutorial', title: 'Gamma Tutorial', short: '', topics: [], isTopLevel: false },
];

describe('documentation tree', () => {
  it('groups sections by section type in deterministic order', () => {
    const tree = buildDocumentationTree(sections);
    expect(tree.map((group) => group.id)).toEqual(['GeneralTopic', 'Example', 'Tutorial']);
    expect(tree[0].sections[0].slug).toBe('topic');
  });

  it('filters by document title and heading text', () => {
    const tree = buildDocumentationTree(sections);
    expect(filterDocumentationTree(tree, 'beta')[0].sections[0].slug).toBe('example');
    const headingMatch = filterDocumentationTree(tree, 'install');
    expect(headingMatch).toHaveLength(1);
    expect(headingMatch[0].sections[0].headings[0].id).toBe('install');
  });
});
