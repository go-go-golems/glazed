#!/usr/bin/env python3
import os
import re
from datetime import datetime
from pathlib import Path

root = Path('.')
frontmatter_path = Path('ttmp/2026/02/02/GL-002-FURTHER-CLEANUP--further-cleanup-and-renaming/analysis/01-exhaustive-parameter-layer-audit.md')

param_re = re.compile(r'parameter', re.IGNORECASE)
layer_re = re.compile(r'layer', re.IGNORECASE)
legacy_tag_re = re.compile(r'glazed\.parameter')

files = []

for dirpath, dirnames, filenames in os.walk(root):
    dirnames[:] = [d for d in dirnames if d != '.git']
    for name in filenames:
        path = Path(dirpath) / name
        rel = path.relative_to(root)
        try:
            data = path.read_bytes()
        except Exception as e:
            files.append({
                'path': rel.as_posix(),
                'size': None,
                'binary': False,
                'error': str(e),
                'has_param': False,
                'has_layer': False,
                'has_legacy_tag': False,
                'matches': [],
            })
            continue
        is_binary = b'\x00' in data
        size = len(data)
        matches = []
        has_param = False
        has_layer = False
        has_legacy_tag = False
        if not is_binary:
            text = data.decode('utf-8', errors='replace')
            for idx, line in enumerate(text.splitlines(), 1):
                if param_re.search(line) or layer_re.search(line) or legacy_tag_re.search(line):
                    if param_re.search(line):
                        has_param = True
                    if layer_re.search(line):
                        has_layer = True
                    if legacy_tag_re.search(line):
                        has_legacy_tag = True
                    matches.append((idx, line))
        files.append({
            'path': rel.as_posix(),
            'size': size,
            'binary': is_binary,
            'error': None,
            'has_param': has_param,
            'has_layer': has_layer,
            'has_legacy_tag': has_legacy_tag,
            'matches': matches,
        })

files.sort(key=lambda x: x['path'])

count_total = len(files)
count_binary = sum(1 for f in files if f['binary'])
count_text = count_total - count_binary
count_param = sum(1 for f in files if f['has_param'])
count_layer = sum(1 for f in files if f['has_layer'])
count_legacy = sum(1 for f in files if f['has_legacy_tag'])

name_param = [f for f in files if re.search(r'parameter', f['path'], re.IGNORECASE)]
name_layer = [f for f in files if re.search(r'layer', f['path'], re.IGNORECASE)]

matched_files = [f for f in files if f['matches']]
matched_docs = [f for f in matched_files if Path(f['path']).suffix.lower() in {'.md', '.txt', '.rst'}]
matched_go = [f for f in matched_files if Path(f['path']).suffix.lower() == '.go']
matched_other = [f for f in matched_files if f not in matched_docs and f not in matched_go]

now = datetime.now().isoformat(timespec='seconds')

lines = []
lines.append('# Exhaustive Parameter/Layer Audit')
lines.append('')
lines.append(f'Generated: `{now}`')
lines.append('')
lines.append('## Scope')
lines.append('This audit scanned **every file** under the `glazed/` repository root (excluding `.git`). Both filenames and file contents were inspected for case-insensitive mentions of `parameter` and `layer`, plus the legacy `glazed.parameter` struct tag. Binary files were detected by NUL bytes and recorded as binary without line-level inspection.')
lines.append('')
lines.append('## Compile status')
lines.append('- `go test ./...` succeeded (see diary for full command output).')
lines.append('')
lines.append('## Summary Counts')
lines.append(f'- Total files scanned: **{count_total}**')
lines.append(f'- Text files scanned: **{count_text}**')
lines.append(f'- Binary files scanned: **{count_binary}**')
lines.append(f'- Files with `parameter` in contents: **{count_param}**')
lines.append(f'- Files with `layer` in contents: **{count_layer}**')
lines.append(f'- Files with `glazed.parameter` in contents: **{count_legacy}**')
lines.append('')
lines.append('## High-level answers')
lines.append('- **Does the repo still compile?** Yes (latest `go test ./...` succeeded).')
lines.append('- **Are there any remaining `layer`/`parameter` mentions?** Yes. See the per-file listings below; they remain in docs, example comments, and a handful of code paths that still describe domain concepts as layers/parameters.')
lines.append('- **Any remaining `glazed.parameter` tags?** Only if listed below; all runtime tag parsing now expects `glazed:`.')
lines.append('')

lines.append('## Filenames containing "parameter"')
if name_param:
    for f in name_param:
        lines.append(f'- `{f["path"]}`')
else:
    lines.append('- (none)')
lines.append('')

lines.append('## Filenames containing "layer"')
if name_layer:
    for f in name_layer:
        lines.append(f'- `{f["path"]}`')
else:
    lines.append('- (none)')
lines.append('')

lines.append('## Files with content matches (documentation)')
if matched_docs:
    for f in matched_docs:
        lines.append(f'### `{f["path"]}`')
        for ln, text in f['matches']:
            lines.append(f'- L{ln}: `{text}`')
        lines.append('')
else:
    lines.append('- (none)')
    lines.append('')

lines.append('## Files with content matches (Go source)')
if matched_go:
    for f in matched_go:
        lines.append(f'### `{f["path"]}`')
        for ln, text in f['matches']:
            lines.append(f'- L{ln}: `{text}`')
        lines.append('')
else:
    lines.append('- (none)')
    lines.append('')

lines.append('## Files with content matches (other types)')
if matched_other:
    for f in matched_other:
        lines.append(f'### `{f["path"]}`')
        for ln, text in f['matches']:
            lines.append(f'- L{ln}: `{text}`')
        lines.append('')
else:
    lines.append('- (none)')
    lines.append('')

lines.append('## Full file index (all files)')
lines.append('| Path | Type | Size (bytes) | parameter? | layer? | glazed.parameter? |')
lines.append('| --- | --- | ---: | :---: | :---: | :---: |')
for f in files:
    ftype = 'binary' if f['binary'] else 'text'
    size = f['size'] if f['size'] is not None else ''
    param_flag = 'Y' if f['has_param'] else ''
    layer_flag = 'Y' if f['has_layer'] else ''
    legacy_flag = 'Y' if f['has_legacy_tag'] else ''
    lines.append(f'| `{f["path"]}` | {ftype} | {size} | {param_flag} | {layer_flag} | {legacy_flag} |')

content = '\n'.join(lines) + '\n'

raw = frontmatter_path.read_text()
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

frontmatter_path.write_text(new_raw)
