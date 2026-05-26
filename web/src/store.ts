// store.ts
// Redux store configuration. RTK Query is the only piece of state management needed;
// no additional slices are required for v1.

import { configureStore } from '@reduxjs/toolkit';
import { helpApi } from './services/api';

const reducer = {
  // RTK Query auto-generates a reducer and a middleware from helpApi.
  [helpApi.reducerPath]: helpApi.reducer,
};

export function makeStore(preloadedState?: unknown) {
  const config = {
    reducer,
    middleware: (getDefaultMiddleware: any) =>
      getDefaultMiddleware().concat(helpApi.middleware),
    ...(preloadedState ? { preloadedState } : {}),
  };

  return configureStore(config as any);
}

// Browser/dev singleton. SSR must call makeStore() per request instead.
export const store = makeStore();

// Infer the RootState and AppDispatch types from the store itself.
export type AppStore = ReturnType<typeof makeStore>;
export type RootState = ReturnType<AppStore['getState']>;
export type AppDispatch = AppStore['dispatch'];
