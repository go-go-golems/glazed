// entry-server.tsx — SSR entry point for the Node.js sidecar.
//
// The SSR server (server.mjs) calls renderApp() after pre-fetching data
// from the Go API. The pre-fetched data is injected into the RTK Query
// cache so React components see cache hits during renderToString.

import React from 'react';
import { renderToString } from 'react-dom/server';
import { StaticRouter } from 'react-router-dom/server';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { helpApi } from './services/api';
import App from './App';
import type { ListPackagesResponse, ListSectionsResponse, SectionDetail } from './types';

export interface SSRResult {
  html: string;
}

/**
 * Render the React app to an HTML string for the given URL.
 */
export function renderApp(
  url: string,
  _data: {
    packages?: ListPackagesResponse | null;
    sections?: ListSectionsResponse | null;
    section?: SectionDetail | null;
  },
): SSRResult {
  // Create a fresh store for each SSR request.
  // We don't pre-populate the RTK Query cache here — instead, we rely
  // on the fact that renderToString will render the component tree and
  // the components will see "loading" state for their API hooks. This
  // means the SSR HTML will show the loading skeleton, not the actual
  // content. The real content comes from the SSR server injecting it
  // directly into the HTML shell.
  //
  // A future optimization would be to pre-populate the cache, but this
  // requires careful type handling with RTK Query's internal APIs.
  const store = configureStore({
    reducer: { [helpApi.reducerPath]: helpApi.reducer },
    middleware: (m: any) => m().concat(helpApi.middleware),
  });

  const html = renderToString(
    <React.StrictMode>
      <Provider store={store}>
        <StaticRouter location={url}>
          <App />
        </StaticRouter>
      </Provider>
    </React.StrictMode>,
  );

  return { html };
}
