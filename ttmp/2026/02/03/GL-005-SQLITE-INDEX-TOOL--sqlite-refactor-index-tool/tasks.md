# Tasks

## TODO

- [x] Add tasks here

- [x] Review GL-005 design doc and define minimal scoped CLI + schema for refactor-index
- [ ] Scaffold refactor-index CLI with Glazed command patterns (root + subcommands)
- [ ] Implement SQLite schema + migrations + DB helpers for runs/diff ingestion
- [ ] Implement ingest diff command (git diff --name-status/-U0) + structured output rows
- [ ] Implement query/report command to list diff files from DB
- [ ] Add golden smoke tests that create a git repo, run refactor-index, and assert expected SQLite rows
