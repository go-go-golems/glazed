#!/usr/bin/env python3
from pathlib import Path
import re

TARGETS = [
    Path('pkg/cmds/schema/schema_test.go'),
    Path('pkg/cmds/schema/section-impl_test.go'),
]

REPLACEMENTS = [
    (re.compile(r'\bLayers\b'), 'Sections'),
    (re.compile(r'\blayers\b'), 'sections'),
    (re.compile(r'\blayers(\d+)'), r'sections\1'),
    (re.compile(r'\bLayer\b'), 'Section'),
    (re.compile(r'\blayer\b'), 'section'),
    (re.compile(r'\blayer(\d+)'), r'section\1'),
]

for path in TARGETS:
    text = path.read_text()
    new_text = text
    for pattern, repl in REPLACEMENTS:
        new_text = pattern.sub(repl, new_text)
    if new_text != text:
        path.write_text(new_text)
        print('Updated', path)
    else:
        print('No changes', path)
