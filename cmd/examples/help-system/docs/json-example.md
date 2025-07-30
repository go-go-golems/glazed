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
- format
Flags:
- --output
- --pretty
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
Order: 2
---

# JSON Output Example

Use the json command to format data as JSON output.

## Basic Usage

```bash
json --input=data.csv --output=json
```

## Pretty Printing

For human-readable output:

```bash
json --input=data.csv --output=json --pretty
```

## Integration with Other Tools

Pipe JSON output to other commands:

```bash
json --input=data.csv | jq '.[] | select(.status == "active")'
```
