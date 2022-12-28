---
Title: Specifying multiple templates with --template-field
Slug: templates-multiple-fields
Command: glaze
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

#### Multiple fields in row mode

Instead of just adding / replacing everything with a single field `_0`, you
can also specify multiple templates using the `--template-field` argument, which has
the form `COLNAME:TEMPLATE`.

In row mode, because the rows are flattened, we might encounter problems if we want
to output the value of a column with a ".". This is why we replace "." with "_" before
passing field values to the template engine.

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

#### Multiple fields in object mode

In object mode, there is separator replacement.

```
❯ glaze json misc/test-data/[123].json \
    --template-field 'foo:{{.a}}-{{.b}},bar:{{.d.f}}'
+------------+---------+
| bar        | foo     |
+------------+---------+
| 7          | 1-2     |
| 70         | 10-20   |
| <no value> | 100-200 |
+------------+---------+    
```