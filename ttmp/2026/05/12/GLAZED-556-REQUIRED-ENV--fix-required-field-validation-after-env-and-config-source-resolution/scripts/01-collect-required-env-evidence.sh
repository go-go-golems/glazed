#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"
OUT="ttmp/2026/05/12/GLAZED-556-REQUIRED-ENV--fix-required-field-validation-after-env-and-config-source-resolution/scripts/evidence"
mkdir -p "$OUT"

files=(
  pkg/cli/cobra-parser.go
  pkg/cmds/fields/cobra.go
  pkg/cmds/schema/section-impl.go
  pkg/cmds/sources/update.go
  pkg/cmds/sources/middlewares.go
  pkg/cli/cobra_parser_config_test.go
  pkg/cmds/sources/update_test.go
  pkg/doc/topics/24-config-files.md
)
for f in "${files[@]}"; do
  safe="${f//\//__}"
  nl -ba "$f" > "$OUT/$safe.nl.txt"
done

rg -n "GatherFlagsFromCobraCommand|ignoreRequired|Required|AppName|FromEnv|updateFromEnv|NewCobraParserFromSections|CobraCommandDefaultMiddlewares|ConfigPlanBuilder" pkg/cli pkg/cmds pkg/doc/topics/24-config-files.md -S > "$OUT/rg-required-env.txt"
