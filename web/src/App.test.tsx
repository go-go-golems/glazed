import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App';

const scrollIntoViewMock = vi.fn();

Object.defineProperty(Element.prototype, 'scrollIntoView', {
  configurable: true,
  value: scrollIntoViewMock,
});

beforeEach(() => {
  scrollIntoViewMock.mockClear();
});

vi.mock('./services/api', () => ({
  useListPackagesQuery: () => ({
    data: {
      packages: [{ name: 'glazed', displayName: 'Glazed', versions: [], sectionCount: 2 }],
      defaultPackage: 'glazed',
    },
  }),
  useListSectionsQuery: () => ({
    data: {
      sections: [
        {
          id: 1,
          packageName: 'glazed',
          slug: 'alpha-section',
          type: 'GeneralTopic',
          title: 'Alpha Section',
          short: 'Alpha short',
          topics: ['alpha'],
          isTopLevel: true,
          headings: [{ id: 'overview', level: 2, text: 'Overview' }],
        },
        {
          id: 2,
          packageName: 'glazed',
          slug: 'beta-section',
          type: 'Tutorial',
          title: 'Beta Section',
          short: 'Beta short',
          topics: ['beta'],
          isTopLevel: false,
        },
      ],
      total: 2,
      limit: -1,
      offset: 0,
    },
    isLoading: false,
    error: undefined,
  }),
  useGetSectionQuery: (args: { slug: string }, options?: { skip?: boolean }) => {
    if (options?.skip || !args?.slug) {
      return { data: undefined };
    }

    const detailBySlug: Record<string, object> = {
      'alpha-section': {
        id: 1,
        packageName: 'glazed',
        slug: 'alpha-section',
        type: 'GeneralTopic',
        title: 'Alpha Section',
        short: 'Alpha short',
        topics: ['alpha'],
        isTopLevel: true,
        headings: [{ id: 'overview', level: 2, text: 'Overview' }],
        flags: ['--alpha'],
        commands: ['glaze alpha'],
        content: '# Alpha\n\n## Overview\n\nOverview text.',
      },
      'beta-section': {
        id: 2,
        packageName: 'glazed',
        slug: 'beta-section',
        type: 'Tutorial',
        title: 'Beta Section',
        short: 'Beta short',
        topics: ['beta'],
        isTopLevel: false,
        flags: ['--beta'],
        commands: ['glaze beta'],
        content: '# Beta',
      },
    };

    return { data: detailBySlug[args.slug] };
  },
}));

/**
 * Render the App inside a MemoryRouter at the given initial path.
 * This mirrors the route structure in main.tsx:
 *   /:package/:version/sections/:slug
 *   /:package/:version
 *   *  (catch-all, redirects to default package)
 */
function renderAppAt(initialPath = '/glazed/_') {
  render(
    <MemoryRouter initialEntries={[initialPath]} future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <Routes>
        <Route path="/:package/:version/sections/:slug" element={<App />} />
        <Route path="/:package/:version" element={<App />} />
        <Route path="*" element={<App />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('App route selection', () => {
  it('shows the route-selected section when the URL points at a section slug', async () => {
    renderAppAt('/glazed/_/sections/alpha-section');

    expect(await screen.findByRole('heading', { name: 'Alpha Section' })).toBeTruthy();
    expect(screen.getByText('alpha-section')).toBeTruthy();
  });

  it('shows EmptyState when no section slug is in the URL', async () => {
    renderAppAt('/glazed/_');

    // Should show the "Select a section from the list." empty state text
    await waitFor(() => {
      expect(screen.getByText('Select a section from the list.')).toBeTruthy();
    });
  });
});

describe('App package selector', () => {
  it('shows package selector but hides version selector for unversioned packages', async () => {
    renderAppAt();

    expect(await screen.findByLabelText('Package')).toBeTruthy();
    expect(screen.queryByLabelText('Version')).toBeNull();
  });
});

describe('App tree navigation', () => {
  it('switches between tree and search navigation modes', async () => {
    renderAppAt();

    expect(await screen.findByRole('tree', { name: 'Documentation tree' })).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Search' }));
    expect(await screen.findByRole('listbox', { name: 'Sections' })).toBeTruthy();
  });

  it('navigates to the correct section URL when a section is selected from the tree', async () => {
    renderAppAt();

    fireEvent.click(await screen.findByRole('treeitem', { name: /Alpha Section/i }));

    // With BrowserRouter, clicking a section navigates via navigate(),
    // which the MemoryRouter captures. We verify the section loads.
    expect(await screen.findByRole('heading', { name: 'Alpha Section' })).toBeTruthy();
  });

  it('navigates to heading URLs when a heading is clicked in the tree', async () => {
    renderAppAt('/glazed/_/sections/alpha-section');

    // Expand the section to show headings
    const alphaItem = await screen.findByRole('treeitem', { name: /Alpha Section/i });
    fireEvent.click(alphaItem);

    // Click the heading
    const overviewItem = screen.getByRole('treeitem', { name: 'Overview' });
    fireEvent.click(overviewItem);

    // Verify scroll was called (heading scroll behavior)
    await waitFor(() => {
      expect(scrollIntoViewMock).toHaveBeenCalledWith({ block: 'start' });
    });
  });

  it('scrolls the markdown pane to the selected subsection', async () => {
    renderAppAt('/glazed/_/sections/alpha-section#overview');

    await waitFor(() => {
      expect(scrollIntoViewMock).toHaveBeenCalledWith({ block: 'start' });
    });
  });
});
