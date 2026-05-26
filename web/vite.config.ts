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
    rollupOptions: {
      input: {
        main: './index.html',
      },
    },
  },

  // SSR build configuration.
  // Produces dist/ssr/entry-server.js — a CJS/ESM module that the
  // Node.js sidecar (server.mjs) imports to render React on the server.
  ssr: {
    noExternal: ['react-dom', 'react-router-dom', '@reduxjs/toolkit', 'react-redux'],
  },

  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
  },

  // Use BrowserRouter in the app. The Go server serves index.html for all
  // non-/api paths (SPA fallback), so no server-side route configuration is
  // needed. URL scheme: /{package}/{version}/sections/{slug}#{heading-id}
});
