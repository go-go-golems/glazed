# Changelog

## 2026-02-02

- Initial workspace created


## 2026-02-02

Add renaming plan and imported notes

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/design-doc/01-further-cleanup-and-renaming-plan.md — No-compat cleanup plan
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/sources/01-glazed-cleanup-notes.md — Imported renaming notes


## 2026-02-02

Switch struct tags to `glazed` and update migration tooling

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/02-examples-rename-report.json — AST migration run report
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/various/01-parameter-layer-mentions.txt — Updated parameter/layer inventory (non-ttmp)
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-001-ADD-MIGRATION-DOCS--add-migration-docs-for-glazed-api-changes/scripts/rename_glazed_api.go — Migration tool tag rewrite support

## 2026-02-02

Add exhaustive parameter/layer audit report

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/01-exhaustive-parameter-layer-audit.md — Audit report covering all files


## 2026-02-02

Add parameter/layer symbol inventory and begin renames

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/schema/cobra.go — Renamed AddLayerToCobraCommand -> AddSectionToCobraCommand
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/02-parameter-layer-symbol-inventory.md — Symbol inventory report


## 2026-02-02

Store audit/rename scripts under ticket scripts

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/02-symbol-inventory.go — Symbol inventory tool moved from /tmp


## 2026-02-03

Rename parsed parameters to field values and update values/schema APIs

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/fields/parsed-parameter.go — ParsedParameter(s) → FieldValue(s)
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/values/parsed-layer.go — SectionValues uses Section/Fields, DecodeSectionInto
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/schema/serialize.go — SerializableSection/Schema with fields/sections tags
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/sources/tests/update-from-map.yaml — fixtures updated to fields
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/05-rename-glazed-api.go — Expanded rename tool


## 2026-02-03

Rename pattern mapper fields to sections/fields and refresh inventory

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/sources/patternmapper/pattern_mapper.go — TargetSection/TargetField rename and error/message updates
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/sources/patternmapper/pattern_mapper_builder.go — Builder API updated to section/field naming
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/sources/patternmapper/loader.go — target_section/target_field tags
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cmds/sources/patternmapper/*_test.go — Tests updated to section/field naming
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/cmd/examples/config-pattern-mapper/main.go — Example mappings updated
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/cmd/examples/config-pattern-mapper/README.md — Docs updated to TargetSection/TargetField
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/cmd/examples/config-pattern-mapper/mappings.yaml — YAML keys updated
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/03-layer-parameter-inventory.md — Inventory report

## 2026-02-03

Align CLI command settings with parsed-fields naming and remove legacy cobra builders (commit 5766e46)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cli/cli.go — Renamed parsed-fields flag and removed legacy load-from-file flag
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/cli/cobra.go — Cleaned cobra builder wrappers and section/values naming
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/cmd/examples/config-overlay/main.go — Updated parser config field names
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2025-08-05/signal-test/main.go — Updated parser config field names in ttmp app

## 2026-02-03

Rename Lua conversion helpers to section/field naming (commit 7874112)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/lua/lua.go — Renamed Lua parsing helpers and error strings
- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/lua/cmds.go — Updated Lua command bridge to parsed values and fields globals

## 2026-02-02

Step 13: finalize section/field doc cleanup and example fix (commit 6844cbf)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/cmd/examples/sources-example/main.go — Remove NewCommandDefinition alias usage


## 2026-02-02

Ticket closed


## 2026-02-03

Step 14: sanitize bundle for PDF upload + reMarkable upload (commit 8a9c26e)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/scripts/25-sanitize-remarkable-bundle.py — Sanitize markdown for PDF upload


## 2026-02-03

Step 15: expand postmortem with concrete examples (commit 37b787d)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/04-postmortem-gl-002-refactor-and-tooling.md — Add concrete examples and commits

## 2026-02-03

Step 16: add postmortem appendix + update bundle script (commit e455738)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/04-postmortem-gl-002-refactor-and-tooling.md — Add appendices


## 2026-02-03

Step 17: expand migration playbook (commit cebc160)

### Related Files

- /home/manuel/workspaces/2026-02-02/refactor-glazed-names/glazed/pkg/doc/tutorials/migrating-to-facade-packages.md — Exhaustive migration guide

