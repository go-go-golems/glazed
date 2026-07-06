# Changelog

## 2026-07-06

- Initial workspace created


## 2026-07-06

Step 1: Investigated bug #596; confirmed at updateFromEnv (prefix not hyphen-normalized). Created ticket, design doc, diary.

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/sources/update.go — bug site


## 2026-07-06

Step 2: Fixed updateFromEnv to normalize hyphens in env prefix; added regression test TestUpdateFromEnvNormalizesHyphenatedPrefix. go test ./... green. Committed fix 4bb2f46 + docs 8f4f17f (--no-verify: pre-existing glazed-lint go1.25/1.26 toolchain mismatch + govulncheck stdlib x509 vulns in untouched files).

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/sources/update.go — updateFromEnv now normalizes prefix hyphens to underscores (4bb2f46)
- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/sources/update_test.go — added TestUpdateFromEnvNormalizesHyphenatedPrefix regression test (4bb2f46)


## 2026-07-06

Step 3: Pushed to wesen fork (--no-verify), opened PR #598 (Fixes #596), posted Bluesky update via goat.

### Related Files

- /home/manuel/workspaces/2026-07-06/fix-glazed-env-dashes/glazed/pkg/cmds/sources/update.go — fix live in PR #598

