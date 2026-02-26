# Tasks

## TODO

- [x] Reproduce `pinocchio --print-parsed-fields` metadata leak with profile registries.
- [x] Collect sanitized evidence that `map-value` carries API-key-like values across unrelated fields.
- [x] Trace root cause to glazed metadata handling and map ingest path.
- [x] Write primary bug report (`design-doc/01-...`).
- [x] Write chronological investigation diary (`reference/01-...`).
- [x] Run `docmgr doctor` for ticket validation.
- [x] Upload bug report bundle to reMarkable and verify remote listing.

## Follow-up (Implementation)

- [ ] Implement fix in `glazed` to preserve `map-value` but eliminate cross-field metadata aliasing.
- [ ] Add unit/integration regression tests for stable per-field `map-value` metadata.
- [ ] Re-run pinocchio repro to verify `ai-chat.ai-engine` no longer shows mismatched config `map-value`.
- [ ] Commit implementation changes in incremental commits.
