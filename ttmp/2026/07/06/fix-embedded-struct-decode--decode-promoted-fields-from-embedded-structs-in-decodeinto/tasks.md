# Tasks

## Done

- [x] 1. Fix `DecodeInto` to recurse into anonymous (embedded) struct fields
- [x] 2. Fix `StructToDataMap` symmetrically (same silent-skip bug)
- [x] 3. Add regression tests for embedded-struct decode (value + pointer) and StructToDataMap
- [x] 4. Run `go test ./... -count=1` (acceptance gate)
- [x] 5. Commit fix + tests (`7bd852f`)
- [x] 6. Open PR referencing #597 → PR #599
- [x] 7. Post Bluesky update via `goat`
- [x] 8. Address PR #599 review (shadowing) via reflect.VisibleFields + fieldByIndex (`78edb9d`)
- [x] 9. Add shadowing regression tests (verified meaningful via git stash round-trip)
- [x] 10. Store investigation probe scripts in ticket scripts/ dir (//go:build ignore)
- [x] 11. Reply to PR #599 review comment
- [x] 12. Post Bluesky update about the shadowing fix via `goat`
