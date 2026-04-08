// services/api.ts
// RTK Query API slice — mirrors the Go HTTP endpoints defined in pkg/help/server/handlers.go.
//
// Base URL: /api  (proxied by Vite in dev, served same-origin in production).
// The proxy in vite.config.ts forwards /api/* → localhost:8088.

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { ListSectionsResponse, SectionDetail, HealthResponse } from '../types';

// Injected base URL — in dev Vite proxies /api to :8088; in prod it's same-origin.
const baseUrl = '/api';

export const helpApi = createApi({
  reducerPath: 'helpApi',
  baseQuery: fetchBaseQuery({ baseUrl }),
  tagTypes: ['Section'],
  endpoints: (builder) => ({
    // GET /api/health
    healthCheck: builder.query<HealthResponse, void>({
      query: () => ({ url: '/health' }),
    }),

    // GET /api/sections
    // Optional query params: type, topic, command, flag, q (search), limit, offset
    listSections: builder.query<ListSectionsResponse, string | void>({
      query: (q) => ({
        url: '/sections',
        params: q ? { q } : undefined,
      }),
      providesTags: (result) =>
        result
          ? [
              ...result.sections.map(({ slug }) => ({ type: 'Section' as const, id: slug })),
              { type: 'Section' as const, id: 'LIST' },
            ]
          : [{ type: 'Section' as const, id: 'LIST' }],
    }),

    // GET /api/sections/:slug
    getSection: builder.query<SectionDetail, string>({
      query: (slug) => ({ url: `/sections/${encodeURIComponent(slug)}` }),
      providesTags: (_result, _error, slug) => [{ type: 'Section' as const, id: slug }],
    }),
  }),
});

// Auto-generated React hooks — use these in components instead of the raw endpoints.
export const {
  useHealthCheckQuery,
  useListSectionsQuery,
  useGetSectionQuery,
} = helpApi;
