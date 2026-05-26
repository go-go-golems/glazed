import { describe, expect, it } from 'vitest';
import { renderApp, parseDocsRoute } from './entry-server';
import type { ListPackagesResponse, ListSectionsResponse, SectionDetail } from './types';

const packages: ListPackagesResponse = {
  packages: [{
    name: 'glazed',
    displayName: 'Glazed',
    versions: ['v1.2.15'],
    latestVersion: 'v1.2.15',
    sectionCount: 1,
  }],
  defaultPackage: 'glazed',
  defaultVersion: 'v1.2.15',
};

const sections: ListSectionsResponse = {
  sections: [{
    id: 1,
    packageName: 'glazed',
    packageVersion: 'v1.2.15',
    slug: 'exposing-a-simple-sql-table',
    type: 'GeneralTopic',
    title: 'Exposing a simple SQL table using glaze',
    short: 'Placeholder for what will be a small application overview',
    topics: ['sqlite'],
    isTopLevel: true,
    headings: [{ id: 'overview', level: 2, text: 'Overview' }],
  }],
  total: 1,
  limit: -1,
  offset: 0,
};

const section: SectionDetail = {
  ...sections.sections[0],
  flags: [],
  commands: ['glaze help exposing-a-simple-sql-table'],
  content: '# Exposing a simple SQL table using glaze\n\nTODO(manuel, 2022-12-10): Make a little tutorial showing how to query a SQL table.',
};

describe('parseDocsRoute', () => {
  it('parses package/version URLs', () => {
    expect(parseDocsRoute('/glazed/v1.2.15')).toEqual({
      packageName: 'glazed',
      version: 'v1.2.15',
      slug: null,
    });
  });

  it('parses section URLs and strips query/hash', () => {
    expect(parseDocsRoute('/glazed/v1.2.15/sections/exposing-a-simple-sql-table?x=1#overview')).toEqual({
      packageName: 'glazed',
      version: 'v1.2.15',
      slug: 'exposing-a-simple-sql-table',
    });
  });

  it('normalizes underscore versions to the API empty version', () => {
    expect(parseDocsRoute('/glazed/_/sections/intro')).toEqual({
      packageName: 'glazed',
      version: '',
      slug: 'intro',
    });
  });
});

describe('renderApp', () => {
  it('renders a package index from preloaded RTK Query data', async () => {
    const { html, preloadedState } = await renderApp('/glazed/v1.2.15', { packages, sections });

    expect(html).toContain('Documentation Index');
    expect(html).toContain('Exposing a simple SQL table using glaze');
    expect(JSON.stringify(preloadedState)).toContain('listSections');
  });

  it('renders a section article body from preloaded RTK Query data', async () => {
    const { html, preloadedState } = await renderApp('/glazed/v1.2.15/sections/exposing-a-simple-sql-table', {
      packages,
      sections,
      section,
    });

    expect(html).toContain('Exposing a simple SQL table using glaze');
    expect(html).toContain('TODO(manuel, 2022-12-10)');
    expect(JSON.stringify(preloadedState)).toContain('getSection');
  });
});
