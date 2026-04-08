// types/index.ts
// TypeScript interfaces that mirror the Go response types in pkg/help/server/types.go.
// Keep these in sync with the server types; the JSON field tags are authoritative.

/** Short summary shape returned in list/search results. */
export interface SectionSummary {
  id: number;
  slug: string;
  /** "GeneralTopic" | "Example" | "Application" | "Tutorial" */
  type: string;
  title: string;
  short: string;
  topics: string[];
  isTopLevel: boolean;
}

/** Full shape returned by GET /api/sections/:slug. */
export interface SectionDetail extends SectionSummary {
  short: string;
  topics: string[];
  flags: string[];
  commands: string[];
  /** Rendered Markdown body */
  content: string;
}

/** Shape of GET /api/sections and GET /api/sections/search. */
export interface ListSectionsResponse {
  sections: SectionSummary[];
  /** Total number of matches before pagination. */
  total: number;
  /** Requested limit, or -1 if no limit. */
  limit: number;
  /** Requested offset. */
  offset: number;
}

/** Shape of GET /api/health. */
export interface HealthResponse {
  ok: boolean;
  sections: number;
}

/** Shape of all error responses (4xx/5xx). */
export interface ErrorResponse {
  error: string; // machine-readable code, e.g. "not_found"
  message: string; // human-readable description
}

/** Section type values, mirrored from model.SectionType in Go. */
export type SectionType = 'GeneralTopic' | 'Example' | 'Application' | 'Tutorial';

/** Filter state for the sidebar section list. */
export interface SectionFilter {
  search: string;
  type: SectionType | 'All';
}
