#!/bin/bash

set -x

glaze json misc/test-data/[123].json
glaze json misc/test-data/[123].json --output csv
glaze json misc/test-data/[123].json --table-format markdown | glow -
glaze json misc/test-data/2.json --output json
glaze json misc/test-data/2.json --output json --flatten
glaze json misc/test-data/[123].json --fields c,b,a --table-format markdown | glow -
glaze json misc/test-data/[123].json --filter d.e

glaze json --input-is-array misc/test-data/rows.json --output yaml
glaze yaml misc/test-data/[123].yaml
glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}'
glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}' \
  --use-row-templates --fields a,_0 \
  --output csv
