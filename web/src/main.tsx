// main.tsx — Application entry point (delegates to entry-client.tsx).
//
// This file is kept for backward compatibility. The actual client-side
// entry point is entry-client.tsx, which uses hydrateRoot() for SSR
// hydration support. This re-exports so that existing build configs
// that reference main.tsx still work.

export {};
// Re-export is not needed — Vite's build config now points to
// entry-client.tsx as the client entry. This file is kept as a
// thin wrapper for the dev server (vite dev uses main.tsx by default).

import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { BrowserRouter } from 'react-router-dom';
import { store } from './store';
import { AppRoutes } from './AppRoutes';
import { ErrorBoundary } from './components/ErrorBoundary';
import './styles/global.css';

// In dev mode, use createRoot (no SSR). In production, the build
// uses entry-client.tsx with hydrateRoot.
ReactDOM.createRoot(document.getElementById('root')!).render(
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
