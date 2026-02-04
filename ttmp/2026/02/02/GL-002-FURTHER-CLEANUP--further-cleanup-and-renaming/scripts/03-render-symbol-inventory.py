#!/usr/bin/env python3
import json
import os
import subprocess
from datetime import datetime
from pathlib import Path

root = Path('.')
json_path = Path('ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/02-parameter-layer-symbol-inventory.json')
doc_path = Path('ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/02-parameter-layer-symbol-inventory.md')

report = json.loads(json_path.read_text())
files = report.get('files', [])

symbol_counts = {}
for f in files:
    for name in f.get('idents', []):
        symbol_counts[name] = symbol_counts.get(name, 0) + 1

sorted_symbols = sorted(symbol_counts.items(), key=lambda x: (-x[1], x[0].lower()))

rg_cmd = ["rg", "-l", "-i", "parameter|layer", "-g", "!ttmp/**"]
rg_proc = subprocess.run(rg_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
rg_output = rg_proc.stdout.splitlines()
rg_output = [line.strip() for line in rg_output if line.strip()]

files_go = [p for p in rg_output if p.endswith('.go')]
files_docs = [p for p in rg_output if p.endswith('.md')]
files_other = [p for p in rg_output if p not in files_go and p not in files_docs]

now = datetime.now().isoformat(timespec='seconds')

lines = []
lines.append('# Parameter/Layer Symbol Inventory')
lines.append('')
lines.append(f'Generated: `{now}`')
lines.append('')
lines.append('## Scope')
lines.append('This report lists Go identifiers (symbols) containing `parameter` or `layer` (case-insensitive) across non-`ttmp` Go files, plus an index of all non-`ttmp` files where the words appear in contents.')
lines.append('')
lines.append('## Summary')
lines.append(f'- Go files with matching identifiers: **{len(files)}**')
lines.append(f'- Unique identifiers found: **{len(symbol_counts)}**')
lines.append(f'- All files with parameter/layer mentions (non-ttmp): **{len(rg_output)}**')
lines.append(f'  - Go files: **{len(files_go)}**')
lines.append(f'  - Markdown files: **{len(files_docs)}**')
lines.append(f'  - Other files: **{len(files_other)}**')
lines.append('')

lines.append('## Top identifiers by file coverage')
for name, count in sorted_symbols[:80]:
    lines.append(f'- `{name}` — {count} file(s)')
lines.append('')

lines.append('## All identifiers (alphabetical)')
for name in sorted(symbol_counts, key=lambda x: x.lower()):
    lines.append(f'- `{name}`')
lines.append('')

lines.append('## Per-file identifier inventory (Go)')
for f in files:
    rel_path = os.path.relpath(f['path'], root)
    lines.append(f'### `{rel_path}`')
    for name in f.get('idents', []):
        lines.append(f'- `{name}`')
    lines.append('')

lines.append('## File index with parameter/layer mentions (non-ttmp)')
lines.append('### Go files')
for path in files_go:
    lines.append(f'- `{path}`')
lines.append('')
lines.append('### Markdown files')
for path in files_docs:
    lines.append(f'- `{path}`')
lines.append('')
lines.append('### Other files')
for path in files_other:
    lines.append(f'- `{path}`')
lines.append('')

content = '\n'.join(lines) + '\n'

raw = doc_path.read_text()
parts = raw.split('---', 2)
if len(parts) >= 3:
    head = parts[1]
    updated_head_lines = []
    for line in head.splitlines():
        if line.startswith('LastUpdated: '):
            updated_head_lines.append(f'LastUpdated: {now}')
        else:
            updated_head_lines.append(line)
    new_head = '\n'.join(updated_head_lines)
    new_raw = '---' + new_head + '\n---\n\n' + content
else:
    new_raw = content

doc_path.write_text(new_raw)
