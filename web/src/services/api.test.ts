import { describe, expect, it } from 'vitest';
import {
  getRuntimeConfig,
  resolveApiBaseUrl,
  resolveDataBasePath,
  resolveRuntimeBaseUrl,
  resolveRuntimeMode,
  type GlazeSiteConfig,
} from './api';

describe('helpApi runtime configuration', () => {
  it('uses site-data paths in static mode', () => {
    const config: GlazeSiteConfig = {
      mode: 'static',
      dataBasePath: './site-data',
    };

    expect(resolveRuntimeMode(config)).toBe('static');
    expect(resolveDataBasePath(config)).toBe('./site-data');
    expect(resolveRuntimeBaseUrl('/docs', config)).toBe('./site-data');
  });

  it('derives mounted api paths in server mode when no explicit apiBaseUrl is set', () => {
    const config: GlazeSiteConfig = {
      mode: 'server',
    };

    expect(resolveRuntimeMode(config)).toBe('server');
    expect(resolveApiBaseUrl('/help')).toBe('/help/api');
    expect(resolveRuntimeBaseUrl('/help', config)).toBe('/help/api');
  });

  it('reads the runtime config from window when present', () => {
    window.__GLAZE_SITE_CONFIG__ = {
      mode: 'static',
      dataBasePath: './exported-data',
      siteTitle: 'Static Help',
    };

    expect(getRuntimeConfig()).toEqual({
      mode: 'static',
      dataBasePath: './exported-data',
      siteTitle: 'Static Help',
    });
  });
});
