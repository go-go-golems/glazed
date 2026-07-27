---
Title: JSON Output Example
Slug: json-example
Short: Format command output as JSON
SectionType: Example
Topics:
- json
- formatting
- output
Commands:
- json
Flags:
- --format
- --output-fields
- --max-output-rows
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
Order: 2
---

# JSON Output Example

Use `--format json` for a JSON array or `--format jsonl` for one compact object per line.

```bash
json --input=data.csv --format=json
json --input=data.csv --format=jsonl --max-output-rows=100
```

Project fields before serialization when only a subset is needed:

```bash
json --input=data.csv --format=jsonl --output-fields=id,status
```

Ad hoc filtering stays in the caller:

```bash
json --input=data.csv --format=json | jq '.[] | select(.status == "active")'
```
