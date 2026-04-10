// services/api.ts
// RTK Query API slice — mirrors the Go HTTP endpoints defined in pkg/help/server/handlers.go.
//
// Base URL: derived from the current pathname so mounted deployments like /help
// call /help/api instead of /api. In dev, the app runs at / and this resolves to /api,
// which Vite proxies to localhost:8088.

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { ListSectionsResponse, SectionDetail, HealthResponse } from '../types';

declare global {
  interface Window {
    __GLAZE_SITE_CONFIG__?: {
      mode?: 'server' | 'static';
      apiBaseUrl?: string;
      dataBasePath?: string;
      siteTitle?: string;
    };
  }
}

function resolveApiBaseUrl(pathname: string): string {
  if (!pathname || pathname === '/') {
    return '/api';
  }

  const mountPrefix = pathname.replace(/\/+$/, '');
  return `${mountPrefix}/api`;
}

const runtimeConfig = window.__GLAZE_SITE_CONFIG__ ?? {};
const isStaticMode = runtimeConfig.mode === 'static';
const baseUrl = isStaticMode
  ? (runtimeConfig.dataBasePath ?? './site-data')
  : (runtimeConfig.apiBaseUrl ?? resolveApiBaseUrl(window.location.pathname));

export const helpApi = createApi({
  reducerPath: 'helpApi',
  baseQuery: fetchBaseQuery({ baseUrl }),
  tagTypes: ['Section'],
  endpoints: (builder) => ({
    // GET /api/health
    healthCheck: builder.query<HealthResponse, void>({
      query: () => ({ url: isStaticMode ? '/health.json' : '/health' }),
    }),

    // GET /api/sections
    // Optional query params: type, topic, command, flag, q (search), limit, offset
    listSections: builder.query<ListSectionsResponse, string | void>({
      query: (q) => ({
        url: isStaticMode ? '/sections.json' : '/sections',
        params: isStaticMode ? undefined : (q ? { q } : undefined),
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
      query: (slug) => ({
        url: isStaticMode
          ? `/sections/${encodeURIComponent(slug)}.json`
          : `/sections/${encodeURIComponent(slug)}`,
      }),
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
