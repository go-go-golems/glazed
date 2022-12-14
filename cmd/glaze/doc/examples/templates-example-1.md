---
Title: Use a single template for single field output
Slug: templates-single-field
Short: |
  ```
  glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d.f}}'
  ```
Topics:
- templates
Commands:
- json
- yaml
Flags:
- template
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can use go templates to either create a new field (called _0 per default).
Per default, the templates are applied at the input level, when rows
are actually still full blown objects (if reading in from JSON for example).

By default, templates are executed at the "object" level, that is before
the input data has been converted to flattened rows. This means that the entire
input object will be replaced by the result of the template, creating a single
column with the name _0.

```
❯ glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d.f}}'

+---------------------+
| _0                  |
+---------------------+
| 1-2: 7              |
| 10-20: 70           |
| 100-200: <no value> |
+---------------------+
```
