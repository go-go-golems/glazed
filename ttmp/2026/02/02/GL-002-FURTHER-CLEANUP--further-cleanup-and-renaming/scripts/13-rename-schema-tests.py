#!/usr/bin/env python3
from pathlib import Path

TARGETS = [
    Path('pkg/cmds/schema/layer_test.go'),
]

REPLACEMENTS = {
    '.AppendLayers(': '.AppendSections(',
    '.PrependLayers(': '.PrependSections(',
}

changed = []
for path in TARGETS:
    text = path.read_text()
    new_text = text
    for old, new in REPLACEMENTS.items():
        new_text = new_text.replace(old, new)
    if new_text != text:
        path.write_text(new_text)
        changed.append(path)

print(f"Updated {len(changed)} file(s)")
