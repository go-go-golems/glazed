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
import { makeStore } from './store';
import { AppRoutes } from './AppRoutes';
import { ErrorBoundary } from './components/ErrorBoundary';
import './styles/global.css';

// Preloaded state injected by the SSR server (window.__PRELOADED_STATE__)
// If present, it's a serialized Redux state that we use to rehydrate the store.
declare global {
  interface Window {
    __PRELOADED_STATE__?: Record<string, unknown>;
  }
}

const preloadedState = window.__PRELOADED_STATE__;
delete window.__PRELOADED_STATE__;

const store = makeStore(preloadedState);

hydrateRoot(
  document.getElementById('root')!,
  <React.StrictMode>
    <Provider store={store}>
      <BrowserRouter future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
        <ErrorBoundary>
          <AppRoutes />
        </ErrorBoundary>
      </BrowserRouter>
    </Provider>
  </React.StrictMode>,
);
