// entry-client.tsx — Client-side entry point with React hydration.
//
// Replaces main.tsx for the production build. Uses hydrateRoot() to
// reuse server-rendered DOM nodes from the SSR sidecar.
// When no SSR sidecar is present (local dev fallback), the <div id="root">
// will be empty and hydrateRoot still works — it creates the DOM from scratch.

import React from 'react';
import { hydrateRoot } from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import { store } from './store';
import App from './App';
import { ErrorBoundary } from './components/ErrorBoundary';
import './styles/global.css';

// Preloaded state injected by the SSR server (window.__PRELOADED_STATE__)
// If present, it's a serialized Redux state that we use to rehydrate the store.
declare global {
  interface Window {
    __PRELOADED_STATE__?: Record<string, unknown>;
  }
}

// If the SSR server injected preloaded state, we could use it to
// initialize the store. For now, the RTK Query cache will be populated
// by the client-side hooks on mount — the SSR-rendered HTML already has
// the content visible, so there's no flash even if the hooks re-fetch.
// Future optimization: parse __PRELOADED_STATE__ and pass as preloadedState.
delete window.__PRELOADED_STATE__;

hydrateRoot(
  document.getElementById('root')!,
  <React.StrictMode>
    <Provider store={store}>
      <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <ErrorBoundary>
          <App />
        </ErrorBoundary>
      </BrowserRouter>
    </Provider>
  </React.StrictMode>,
);
