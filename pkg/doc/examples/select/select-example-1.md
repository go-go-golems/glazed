---
Title: Select a single column and output raw
Slug: select-single-column
Short: |
  ```
  glaze json misc/test-data/[123].json --select a
  ```
Commands:
- json
- yaml
Flags:
- select
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can quickly select a field and output it in raw format for use in shell scripts.

This is really just a shortcut for `--output tsv --fields a --with-headers=false`.

```
❯ glaze json misc/test-data/[123].json --select a           
1
10
100
```
