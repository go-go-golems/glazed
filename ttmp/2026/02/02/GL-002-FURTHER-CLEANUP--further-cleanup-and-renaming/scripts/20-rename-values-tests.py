#!/usr/bin/env python3
from pathlib import Path
import re

path = Path('pkg/cmds/values/section-values_test.go')
text = path.read_text()

replacements = [
    (re.compile(r'\bLayers\b'), 'Sections'),
    (re.compile(r'\blayers\b'), 'sections'),
    (re.compile(r'\bLayer\b'), 'Section'),
    (re.compile(r'\blayer\b'), 'section'),
    (re.compile(r'Layer'), 'Section'),
    (re.compile(r'layer'), 'section'),
    (re.compile(r'\blayer(\d+)'), r'section\1'),
    (re.compile(r'\blayers(\d+)'), r'sections\1'),
    (re.compile(r'parsedLayers'), 'parsedValues'),
    (re.compile(r'parsedLayer'), 'sectionValues'),
]

new_text = text
for pattern, repl in replacements:
    new_text = pattern.sub(repl, new_text)

if new_text != text:
    path.write_text(new_text)
    print('Updated', path)
else:
    print('No changes')
