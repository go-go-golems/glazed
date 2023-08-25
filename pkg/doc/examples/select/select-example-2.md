---
Title: Select a single column and output raw
Slug: select-single-column
Short: |
  ```
  glaze json misc/test-data/[123].json --select-template '{{.a}}-{{.b}}'
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
In order to output more complex values, you can specific a single `select` template.

This is really just a shortcut for `--output tsv --with-headers=false --template=`.

```
❯ glaze json misc/test-data/[123].json --select-template '{{.a}}-{{.b}}'
1-2
10-20
100-200
```
