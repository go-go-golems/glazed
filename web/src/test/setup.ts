import { cleanup } from '@testing-library/react';
import { afterEach, vi } from 'vitest';

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  delete window.__GLAZE_SITE_CONFIG__;
  window.history.replaceState({}, '', '/');
  window.location.hash = '';
});
