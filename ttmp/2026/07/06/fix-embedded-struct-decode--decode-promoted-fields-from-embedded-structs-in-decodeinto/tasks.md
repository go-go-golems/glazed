# Tasks

## Done

- [x] 1. Fix `DecodeInto` to recurse into anonymous (embedded) struct fields
- [x] 2. Fix `StructToDataMap` symmetrically (same silent-skip bug)
- [x] 3. Add regression tests for embedded-struct decode (value + pointer) and StructToDataMap
- [x] 4. Run `go test ./... -count=1` (acceptance gate)
- [x] 5. Commit fix + tests (`7bd852f`)
- [ ] 6. Open PR referencing #597
- [ ] 7. Post Bluesky update via `goat`
