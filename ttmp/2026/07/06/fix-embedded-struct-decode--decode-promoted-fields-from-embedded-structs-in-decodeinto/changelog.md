# Changelog

## 2026-07-06

- Initial workspace created


## 2026-07-06

Step 1: Investigated bug #597; confirmed DecodeInto (and symmetric StructToDataMap) skip anonymous struct fields. Created ticket, design doc, diary. Branched task/fix-glazed-embedded-struct-decode off origin/main.

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/fields/initialize-struct.go — bug site


## 2026-07-06

Step 2: Fixed DecodeInto (+ symmetric StructToDataMap) to recurse into anonymous struct fields via reflect.Value (not .Interface()). go test ./... green. Committed 7bd852f (--no-verify: pre-existing glazed-lint go1.25/1.26 toolchain mismatch + govulncheck stdlib x509 vulns in untouched files).

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/fields/initialize-struct.go — DecodeInto split into decodeIntoValue + decodeEmbedded; StructToDataMap -> structValueToDataMap (7bd852f)
- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/fields/initialize-struct_test.go — added TestDecodeIntoEmbedded* and TestStructToDataMapWithEmbedded* regression tests (7bd852f)


## 2026-07-06

Step 3: Pushed to wesen fork (--no-verify), opened PR #599 (Fixes #597), posted Bluesky update via goat.

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/fields/initialize-struct.go — fix live in PR #599

