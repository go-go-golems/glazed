# Changelog

## 2026-02-25

- Initial workspace created
- Reproduced parsed-fields metadata leak from `pinocchio` with profile registry path.
- Captured sanitized evidence showing `map-value` propagating secret-looking data across multiple fields.
- Root cause identified in glazed metadata handling:
  - `GatherFieldsFromMap` adds raw `map-value`.
  - `WithMetadata` aliases metadata maps (no defensive copy).
  - `printParsedFields` emits metadata verbatim.
- Added design-doc bug report with remediation plan and test strategy.
- Added investigation diary with commands, outputs, and evidence.
- Implemented metadata-copy fix in `glazed/pkg/cmds/fields/field-value.go` to remove parse-step metadata aliasing while preserving `map-value`.
- Added regression tests:
  - `pkg/cmds/fields/field-value_test.go` for metadata map copying and no cross-step aliasing.
  - `pkg/cmds/fields/gather-fields_test.go` for per-field `map-value` stability under shared parse metadata options.
- Verified with pinocchio repro that config `map-value` now matches each field value (no cross-field contamination).

## 2026-02-25

Investigated parsed-fields map-value leak; confirmed metadata aliasing root cause in glazed and documented remediation plan.

### Related Files

- /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/field-value.go — metadata map aliasing in WithMetadata
- /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/pkg/cmds/fields/gather-fields.go — raw map-value attached to parse-step metadata
- /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/ttmp/2026/02/25/GL-003-PARSED-FIELDS-METADATA-LEAK--parsed-fields-metadata-leak-via-shared-map-value/design-doc/01-bug-report-map-value-metadata-leaks-secrets-in-print-parsed-fields.md — bug report deliverable


## 2026-02-25

Validated ticket with docmgr doctor (clean) and uploaded bug report bundle to reMarkable at /ai/2026/02/25/GL-003-PARSED-FIELDS-METADATA-LEAK.

### Related Files

- /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/ttmp/2026/02/25/GL-003-PARSED-FIELDS-METADATA-LEAK--parsed-fields-metadata-leak-via-shared-map-value/design-doc/01-bug-report-map-value-metadata-leaks-secrets-in-print-parsed-fields.md — included in uploaded bundle
- /home/manuel/workspaces/2026-02-24/geppetto-profile-registry-js/glazed/ttmp/2026/02/25/GL-003-PARSED-FIELDS-METADATA-LEAK--parsed-fields-metadata-leak-via-shared-map-value/reference/01-investigation-diary-parsed-fields-metadata-leak.md — included in uploaded bundle
