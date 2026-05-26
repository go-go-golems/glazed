// entry-server.tsx — SSR entry point for the Node.js sidecar.
//
// The SSR server (server.mjs) calls renderApp() after pre-fetching data
// from the Go API. The pre-fetched data is inserted into RTK Query's
// cache before renderToString(), so React components render the same
// documentation tree and article body on the server that the browser will
// hydrate on the client.

import React from 'react';
import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router-dom/server';
import { Provider } from 'react-redux';
import { makeStore } from './store';
import { helpApi } from './services/api';
import { AppRoutes } from './AppRoutes';
import type { ListPackagesResponse, ListSectionsResponse, SectionDetail } from './types';

export interface SSRData {
  packages?: ListPackagesResponse | null;
  sections?: ListSectionsResponse | null;
  section?: SectionDetail | null;
}

export interface SSRResult {
  html: string;
  preloadedState: unknown;
}

interface DocsRoute {
  packageName: string;
  version: string;
  slug: string | null;
}

const NO_VERSION = '_';

function versionFromUrl(urlVersion: string | undefined): string {
  if (!urlVersion || urlVersion === NO_VERSION) return '';
  return urlVersion;
}

export function parseDocsRoute(url: string): DocsRoute {
  const pathname = url.split('#')[0]?.split('?')[0] || '/';
  const parts = pathname.replace(/^\/+/, '').replace(/\/+$/, '').split('/').filter(Boolean);

  if (parts.length >= 4 && parts[2] === 'sections') {
    return {
      packageName: parts[0] ?? '',
      version: versionFromUrl(parts[1]),
      slug: parts[3] ?? null,
    };
  }

  if (parts.length >= 2) {
    return {
      packageName: parts[0] ?? '',
      version: versionFromUrl(parts[1]),
      slug: null,
    };
  }

  if (parts.length >= 1) {
    return {
      packageName: parts[0] ?? '',
      version: '',
      slug: null,
    };
  }

  return { packageName: '', version: '', slug: null };
}

async function preloadRTKQueryCache(store: ReturnType<typeof makeStore>, url: string, data: SSRData) {
  const route = parseDocsRoute(url);
  const preloadActions: Array<Promise<unknown>> = [];

  if (data.packages) {
    preloadActions.push(store.dispatch(helpApi.util.upsertQueryData('listPackages', undefined, data.packages)) as unknown as Promise<unknown>);
  }

  if (route.packageName && data.sections) {
    preloadActions.push(store.dispatch(helpApi.util.upsertQueryData('listSections', {
      packageName: route.packageName,
      version: route.version,
    }, data.sections)) as unknown as Promise<unknown>);
  }

  if (route.packageName && route.slug && data.section) {
    preloadActions.push(store.dispatch(helpApi.util.upsertQueryData('getSection', {
      slug: route.slug,
      packageName: route.packageName,
      version: route.version,
    }, data.section)) as unknown as Promise<unknown>);
  }

  await Promise.all(preloadActions);
}

/**
 * Render the React app to an HTML string for the given URL.
 */
export async function renderApp(url: string, data: SSRData): Promise<SSRResult> {
  const store = makeStore();
  await preloadRTKQueryCache(store, url, data);

  const html = renderToString(
    <React.StrictMode>
      <Provider store={store}>
        <StaticRouter location={url}>
          <AppRoutes />
        </StaticRouter>
      </Provider>
    </React.StrictMode>,
  );

  return { html, preloadedState: store.getState() };
}
