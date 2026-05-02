# Tasks

## TODO

- [x] Add PackageName and PackageVersion fields to model.Section plus Go/TypeScript API response types.
- [x] Migrate help store schema from globally unique slug to package/version/slug identity and add package-aware query predicates.
- [x] Implement SQLiteDirLoader and --from-sqlite-dir discovery for X/Y/X.db, X/X.db, and X.db layouts.
- [x] Assign package metadata for embedded Glazed docs, --from-glazed-cmd sources, explicit --from-sqlite files, JSON exports, and Markdown paths.
- [x] Add GET /api/packages and package/version filters for section list and detail endpoints.
- [x] Update the React web app with Package and conditional Version selectors matching the provided screenshot.
- [x] Add backend and frontend tests for duplicate slugs, directory discovery, package/version filters, and conditional version UI.
- [x] Validate with real pinocchio help export SQLite DB and resolve or document codebase-browser help export compatibility.
