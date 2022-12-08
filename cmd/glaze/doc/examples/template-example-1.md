---
Title: Use a single template for single field output
Slug: templates-example-1
Short: "glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d.f}}'"
Topics:
- templates
Commands:
- json
- yaml
Flags:
- template
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Example
---
You can use go templates to either create a new field (called _0 per default).
Per default, the templates are applied at the input level, when rows
are actually still full blown objects (if reading in from JSON for example).

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
