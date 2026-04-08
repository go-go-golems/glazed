#!/bin/bash
# 02-preset-summary.sh
# Run the framework-summary preset on the converted session
set -euo pipefail

go-minitrace query duckdb \
  --archive-glob './analysis/pi-help-browser/active/*/*.minitrace.json' \
  --preset framework-summary
