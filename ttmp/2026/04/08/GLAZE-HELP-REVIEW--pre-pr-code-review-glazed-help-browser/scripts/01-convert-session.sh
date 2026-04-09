#!/bin/bash
# 01-convert-session.sh
# Convert the Pi session JSONL to minitrace format for analysis
set -euo pipefail

SESSION="/home/manuel/.pi/agent/sessions/--home-manuel-workspaces-2026-04-07-glaze-help-browser-glazed--/2026-04-08T00-21-48-462Z_8cea1965-7269-4c42-abd0-4c6bc82b66c6.jsonl"
OUTPUT_DIR="./analysis/pi-help-browser"

mkdir -p "$OUTPUT_DIR"

go-minitrace convert pi \
  --source-session "$SESSION" \
  --output-dir "$OUTPUT_DIR"

echo "Converted to $OUTPUT_DIR"
ls -la "$OUTPUT_DIR"/active/*/
