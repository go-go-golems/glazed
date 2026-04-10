// services/api.ts
// RTK Query API slice — mirrors the Go HTTP endpoints defined in pkg/help/server/handlers.go.
//
// Base URL: derived from the current pathname so mounted deployments like /help
// call /help/api instead of /api. In dev, the app runs at / and this resolves to /api,
// which Vite proxies to localhost:8088.

import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { ListSectionsResponse, SectionDetail, HealthResponse } from '../types';

export interface GlazeSiteConfig {
  mode?: 'server' | 'static';
  apiBaseUrl?: string;
  dataBasePath?: string;
  siteTitle?: string;
}

declare global {
  interface Window {
    __GLAZE_SITE_CONFIG__?: GlazeSiteConfig;
  }
}

export function resolveApiBaseUrl(pathname: string): string {
  if (!pathname || pathname === '/') {
    return '/api';
  }

  const mountPrefix = pathname.replace(/\/+$/, '');
  return `${mountPrefix}/api`;
}

export function getRuntimeConfig(win: Window = window): GlazeSiteConfig {
  return win.__GLAZE_SITE_CONFIG__ ?? {};
}

export function resolveRuntimeMode(config: GlazeSiteConfig): 'server' | 'static' {
  return config.mode === 'static' ? 'static' : 'server';
}

export function resolveDataBasePath(config: GlazeSiteConfig): string {
  return config.dataBasePath ?? './site-data';
}

export function resolveRuntimeBaseUrl(
  pathname: string,
  config: GlazeSiteConfig,
): string {
  if (resolveRuntimeMode(config) === 'static') {
    return resolveDataBasePath(config);
  }

  return config.apiBaseUrl ?? resolveApiBaseUrl(pathname);
}

const runtimeConfig = getRuntimeConfig();
const runtimeMode = resolveRuntimeMode(runtimeConfig);
const isStaticMode = runtimeMode === 'static';
const baseUrl = resolveRuntimeBaseUrl(window.location.pathname, runtimeConfig);

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
