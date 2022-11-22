#!/bin/bash

set -x

glaze json misc/test-data/*.json
glaze json misc/test-data/*.json --output csv
glaze json misc/test-data/*.json --table-format markdown | glow -
glaze json misc/test-data/2.json --output json
glaze json misc/test-data/2.json --output json --flatten
glaze json misc/test-data/*.json --fields c,b,a --table-format markdown | glow -
glaze json misc/test-data/*.json --filter d.e

