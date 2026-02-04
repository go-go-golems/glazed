#!/usr/bin/env python3
from pathlib import Path

root = Path('.')
old = 'AddLayerToCobraCommand'
new = 'AddSectionToCobraCommand'

for path in root.rglob('*.go'):
    if '.git' in path.parts or 'ttmp' in path.parts:
        continue
    text = path.read_text()
    if old in text:
        path.write_text(text.replace(old, new))
