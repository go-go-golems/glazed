#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path('pkg/cmds/fields')
TEST_DATA = ROOT / 'test-data'

EXPLICIT_REPLACEMENTS = [
    ('ParameterType', 'Type'),
    ('ParameterTypes', 'Types'),
    ('parameters_test.yaml', 'definitions_test.yaml'),
    ('parameters_validity_test.yaml', 'definitions_validity_test.yaml'),
    ('gatherParametersYAML', 'gatherFieldsYAML'),
    ('TestGatherParametersFromMap', 'TestGatherFieldsFromMap'),
    ('ParameterDefs', 'FieldDefs'),
    ('parameterDefs', 'fieldDefs'),
    ('parameterDefinitions', 'fieldDefinitions'),
    ('ParameterDefinition', 'FieldDefinition'),
    ('parameterDefinition', 'fieldDefinition'),
    ('parameters_', 'fields_'),
    ('parameterType', 'fieldType'),
]

REGEX_REPLACEMENTS = [
    (re.compile(r'\bParameters\b'), 'Fields'),
    (re.compile(r'\bparameters\b'), 'fields'),
    (re.compile(r'\bParameter\b'), 'Field'),
    (re.compile(r'\bparameter\b'), 'field'),
    (re.compile(r'Parameter'), 'Field'),
]

paths = list(ROOT.glob('*.go')) + list(TEST_DATA.glob('*.yaml'))
changed = []
for path in paths:
    text = path.read_text()
    new_text = text
    for old, new in EXPLICIT_REPLACEMENTS:
        new_text = new_text.replace(old, new)
    for pattern, repl in REGEX_REPLACEMENTS:
        new_text = pattern.sub(repl, new_text)
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} file(s)")
