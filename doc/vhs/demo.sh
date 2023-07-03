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
glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d.f}}'
glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}' \
  --use-row-templates --fields a,_0 \
  --output csv

glaze json misc/test-data/[123].json \
    --template-field 'foo:{{.a}}-{{.b}},bar:{{.d_f}}' \
    --use-row-templates --fields a,foo,bar
glaze json misc/test-data/[123].json \
    --template-field '@misc/template-field-row.yaml' \
    --use-row-templates  --output markdown

glaze json misc/test-data/[123].json \
    --template-field '@misc/template-field-object.yaml' \
    --output json

glaze json misc/test-data/[123].json --select a

glaze json misc/test-data/[123].json \
    --select-template '{{.a}}-{{.b}}'

glaze yaml misc/test-data/test.yaml --input-is-array --rename baz:blop

glaze yaml misc/test-data/test.yaml --input-is-array \
    --rename-regexp '^(.*)bar:${1}blop'

glaze yaml misc/test-data/test.yaml --input-is-array \
    --rename-yaml misc/rename.yaml

 glaze json misc/test-data/[123].json --output csv \
    --output-file /tmp/test.csv