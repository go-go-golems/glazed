import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  base: './',

  // API proxy: during development, forward /api requests to the Go server.
  // In production the SPA is embedded in the Go binary and served same-origin.
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8088',
        changeOrigin: true,
        // No rewrite needed: Go server serves /api/* directly.
      },
    },
  },

  // Build output goes to dist/, which is embedded into the Go binary.
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },

  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },

  // Use HashRouter in the app so no server-side route fallback is needed.
  // SPA routes will be #/sections/database etc.
});
