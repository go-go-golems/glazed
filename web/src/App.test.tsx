import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { HashRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import App from './App';

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
        flags: ['--alpha'],
        commands: ['glaze alpha'],
        content: '# Alpha',
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

function renderAppAt(hash = '') {
  window.history.replaceState({}, '', '/');
  window.location.hash = hash;

  render(
    <HashRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
      <App />
    </HashRouter>,
  );
}

describe('App hash-route selection', () => {
  it('shows the route-selected section when the hash points at a section slug', async () => {
    renderAppAt('#/sections/alpha-section');

    expect(await screen.findByRole('heading', { name: 'Alpha Section' })).toBeTruthy();
    expect(screen.getByText('alpha-section')).toBeTruthy();
  });

  it('updates the hash route when a section is selected from the list', async () => {
    renderAppAt();

    fireEvent.click(screen.getByRole('button', { name: /Beta Section/i }));

    await waitFor(() => {
      expect(window.location.hash).toBe('#/sections/beta-section');
    });

    expect(await screen.findByRole('heading', { name: 'Beta Section' })).toBeTruthy();
  });
});

describe('App package selector', () => {
  it('shows package selector but hides version selector for unversioned packages', async () => {
    renderAppAt();

    expect(await screen.findByLabelText('Package')).toBeTruthy();
    expect(screen.queryByLabelText('Version')).toBeNull();
  });
});
