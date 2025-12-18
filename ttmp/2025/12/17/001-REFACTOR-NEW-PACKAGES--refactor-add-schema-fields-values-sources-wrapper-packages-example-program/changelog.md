# Changelog

## 2025-12-17

- Initial workspace created


## 2025-12-17

Seeded ticket docs (design/plan/diary), clarified wrapper package API surfaces, and linked key implementation files to ground the work.

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/ttmp/2025/12/17/001-REFACTOR-NEW-PACKAGES--refactor-add-schema-fields-values-sources-wrapper-packages-example-program/design-doc/01-design-wrapper-packages-schema-fields-values-sources.md — Design proposal for schema/fields/values/sources facade packages
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/ttmp/2025/12/17/001-REFACTOR-NEW-PACKAGES--refactor-add-schema-fields-values-sources-wrapper-packages-example-program/diary/01-diary.md — Diary capturing chronological work log
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/ttmp/2025/12/17/001-REFACTOR-NEW-PACKAGES--refactor-add-schema-fields-values-sources-wrapper-packages-example-program/planning/01-implementation-plan-wrapper-packages-example-program.md — Step-by-step plan and acceptance criteria


## 2025-12-17

Implemented all four wrapper packages (schema/fields/values/sources) with type aliases and wrapper functions. All packages compile successfully.

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/fields/fields.go — Fields wrapper package with Definition/Type aliases
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/schema/schema.go — Schema wrapper package with Section/Sections aliases
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/sources/sources.go — Sources wrapper package with middleware wrappers and Execute function
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/pkg/cmds/values/values.go — Values wrapper package with SectionValues/Values aliases and DecodeInto helpers


## 2025-12-17

Created example program demonstrating wrapper packages. Program successfully demonstrates multiple schema sections, env+cobra parsing, and struct decoding. All tests pass.

### Related Files

- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/cmd/examples/refactor-new-packages/README.md — Documentation for example program
- /home/manuel/workspaces/2025-11-18/fix-pinocchio-profiles/glazed/cmd/examples/refactor-new-packages/main.go — Example program using schema/fields/values packages


## 2025-12-18

Closed: wrapper packages + example were implemented and are in git; ticket tasks file was not kept in sync.

