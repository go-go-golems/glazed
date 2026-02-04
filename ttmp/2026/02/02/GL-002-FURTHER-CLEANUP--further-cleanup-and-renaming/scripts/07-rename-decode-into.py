#!/usr/bin/env python3
import re
from pathlib import Path

ROOT = Path('.')
SKIP_DIRS = {'.git', 'node_modules', 'vendor', 'ttmp'}

pattern = re.compile(r"\.InitializeStruct\(")

changed = []
for path in ROOT.glob('**/*.go'):
    if not path.is_file():
        continue
    if any(part in SKIP_DIRS for part in path.parts):
        continue
    text = path.read_text()
    new_text = pattern.sub('.DecodeInto(', text)
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} files")
