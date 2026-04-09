// store.ts
// Redux store configuration. RTK Query is the only piece of state management needed;
// no additional slices are required for v1.

import { configureStore } from '@reduxjs/toolkit';
import { helpApi } from './services/api';

export const store = configureStore({
  reducer: {
    // RTK Query auto-generates a reducer and a middleware from helpApi.
    [helpApi.reducerPath]: helpApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(helpApi.middleware),
});

// Infer the RootState and AppDispatch types from the store itself.
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
