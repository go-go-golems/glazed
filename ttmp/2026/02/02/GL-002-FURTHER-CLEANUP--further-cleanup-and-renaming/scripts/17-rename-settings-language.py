#!/usr/bin/env python3
from pathlib import Path
import re

ROOT = Path('pkg/settings')

REPLACEMENTS = [
    (re.compile(r'\bParameters\b'), 'Fields'),
    (re.compile(r'\bparameters\b'), 'fields'),
    (re.compile(r'\bParameter\b'), 'Field'),
    (re.compile(r'\bparameter\b'), 'field'),
    (re.compile(r'\bLayers\b'), 'Sections'),
    (re.compile(r'\blayers\b'), 'sections'),
    (re.compile(r'\bLayer\b'), 'Section'),
    (re.compile(r'\blayer\b'), 'section'),
]

changed = []
for path in ROOT.glob('*.go'):
    text = path.read_text()
    new_text = text
    for pattern, repl in REPLACEMENTS:
        new_text = pattern.sub(repl, new_text)
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} file(s)")
