#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path('pkg/cmds/sources/tests')

KEY_REPLACEMENTS = {
    'parameterLayers': 'sections',
    'parsedLayers': 'values',
    'expectedLayers': 'expectedSections',
}

WORD_REPLACEMENTS = [
    (re.compile(r'\bParameterLayers\b'), 'Schema'),
    (re.compile(r'\bParameterLayer\b'), 'Section'),
    (re.compile(r'\bParsedLayers\b'), 'Values'),
    (re.compile(r'\bParameters\b'), 'Fields'),
    (re.compile(r'\bparameters\b'), 'fields'),
    (re.compile(r'\bParameter\b'), 'Field'),
    (re.compile(r'\bparameter\b'), 'field'),
    (re.compile(r'\bLayers\b'), 'Sections'),
    (re.compile(r'\bLayer\b'), 'Section'),
    (re.compile(r'\blayers\b'), 'sections'),
    (re.compile(r'\blayer\b'), 'section'),
    (re.compile(r'\blayers_\b'), 'schema_'),
    (re.compile(r'\blayer(\d+)\b'), r'section\1'),
]

changed = []
for path in ROOT.glob('*.yaml'):
    text = path.read_text()
    new_text = text
    for old, new in KEY_REPLACEMENTS.items():
        new_text = new_text.replace(old, new)
    for pattern, repl in WORD_REPLACEMENTS:
        new_text = pattern.sub(repl, new_text)
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} file(s)")
