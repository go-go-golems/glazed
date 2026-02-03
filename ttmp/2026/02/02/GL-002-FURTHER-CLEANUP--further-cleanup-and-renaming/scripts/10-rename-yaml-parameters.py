#!/usr/bin/env python3
from pathlib import Path

ROOT = Path('pkg/cmds/sources/tests')
changed = []
for path in ROOT.glob('*.yaml'):
    text = path.read_text()
    new_text = text.replace('parameters:', 'fields:')
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} files")
