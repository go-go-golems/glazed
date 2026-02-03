#!/usr/bin/env python3
from pathlib import Path
import re

path = Path('pkg/cmds/sources/custom-profiles_test.go')
text = path.read_text()

replacements = [
    (re.compile(r'\bparameterLayers\b'), 'schema_'),
    (re.compile(r'\bparsedLayers\b'), 'parsedValues'),
    (re.compile(r'\bparsedLayer\b'), 'sectionValues'),
    (re.compile(r'\bLayers\b'), 'Sections'),
    (re.compile(r'\blayers\b'), 'sections'),
    (re.compile(r'\bLayer\b'), 'Section'),
    (re.compile(r'\blayer\b'), 'section'),
]

new_text = text
for pattern, repl in replacements:
    new_text = pattern.sub(repl, new_text)

if new_text != text:
    path.write_text(new_text)
    print('Updated', path)
else:
    print('No changes')
