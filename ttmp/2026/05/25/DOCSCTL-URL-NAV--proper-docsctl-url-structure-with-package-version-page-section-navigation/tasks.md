# Tasks

## TODO

- [x] 1. Fix site-config.js: add explicit apiBaseUrl
- [x] 2. Fix api.ts: prefer explicit apiBaseUrl over pathname-derived URL
- [x] 3. Switch main.tsx from HashRouter to BrowserRouter with /:package/:version routes
- [x] 4. Rewrite App.tsx: extract package/version/slug from URL params, navigate with full paths
- [x] 5. Update App.test.tsx: switch from HashRouter to BrowserRouter, update assertions
- [x] 6. Update api.test.ts: add test for explicit apiBaseUrl taking precedence
- [x] 7. Update vite.config.ts: remove HashRouter comment, note BrowserRouter requirements
- [x] 8. Add legacy hash URL redirect in index.html
- [x] 9. Manual testing with the live docs site
- [x] 10. Update diary and changelog
