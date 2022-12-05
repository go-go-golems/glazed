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
- template-field
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: Example
---
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
