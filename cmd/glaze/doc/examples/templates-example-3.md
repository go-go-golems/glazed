---
Title: Specifying multiple templates with --template-field
Slug: templates-example-3
Short: |
  ```
  glaze json misc/test-data/[123].json \
  --template-field 'foo:{{.a}}-{{.b}},bar:{{.d_f}}' \
  --use-row-templates --fields a,foo,bar
  ```
Topics:
- templates
Commands:
- json
- yaml
Flags:
- template
- use-row-templates
- template-field
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---
Instead of just adding / replacing everything with a single field `_0`, you
can also specify multiple templates using the `--template-field` argument, which has
the form `COLNAME:TEMPLATE`.

``` 
❯ glaze json misc/test-data/[123].json \
    --template-field 'foo:{{.a}}-{{.b}},bar:{{.d_f}}' \
    --use-row-templates --fields a,foo,bar
+-----+---------+------------+
| a   | foo     | bar        |
+-----+---------+------------+
| 1   | 1-2     | 7          |
| 10  | 10-20   | 70         |
| 100 | 100-200 | <no value> |
+-----+---------+------------+
```
