#!/usr/bin/env bash
set -euo pipefail

mv pkg/cmds/fields/parameters.go pkg/cmds/fields/definitions.go
mv pkg/cmds/fields/parameters_test.go pkg/cmds/fields/definitions_test.go
mv pkg/cmds/fields/parameters_from_defaults_test.go pkg/cmds/fields/definitions_from_defaults_test.go
mv pkg/cmds/fields/parsed-parameter.go pkg/cmds/fields/field-value.go
mv pkg/cmds/fields/gather-parameters.go pkg/cmds/fields/gather-fields.go
mv pkg/cmds/fields/gather-parameters_test.go pkg/cmds/fields/gather-fields_test.go
mv pkg/cmds/fields/parameter-type.go pkg/cmds/fields/field-type.go

mv pkg/cmds/fields/test-data/parameters_test.yaml pkg/cmds/fields/test-data/definitions_test.yaml
mv pkg/cmds/fields/test-data/parameters_validity_test.yaml pkg/cmds/fields/test-data/definitions_validity_test.yaml
