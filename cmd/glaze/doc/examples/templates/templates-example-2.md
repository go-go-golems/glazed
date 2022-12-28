---
Title: Apply templates at the row level
Slug: templates-row-level
Short: |
  ```
  glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}' \
  --use-row-templates --fields a,_0 \
  --output csv
  ```
Topics:
- templates
Commands:
- json
- yaml
Flags:
- template
- use-row-templates
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---
You can also apply templates at the row level, once the input has been flattened.
In this case, because flattened columns contain the symbol `.`, fields get renamed
to use the symbol `_` as a separator.

The new column is output alongside all the other columns.

``` 
❯ glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}' \
  --use-row-templates --fields a,_0 \
  --output csv
a,_0
1,1-2: 7
10,10-20: 70
100,100-200: <no value>
```

