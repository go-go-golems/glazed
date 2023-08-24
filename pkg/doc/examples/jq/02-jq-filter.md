---
Title: Use jq to filter out rows
Slug: jq-filter
Short: |
  ```
  glaze json misc/test-data/[123].json --jq '.select(.a > 10) // []'
  ```
Topics:
- jq
Commands:
- json
- yaml
- csv
Flags:
- jq
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---
 
Using `.select` as well as the alternative operator `//` you can filter out rows.
 
```
❯ glaze json misc/test-data/[123].json --jq '.select(.a > 10) // []'
+-----+-----+-----+
| c   | a   | b   |
+-----+-----+-----+
| 300 | 100 | 200 |
+-----+-----+-----+
```

You can also delete a field and copy another one, in a single command.

``` 
❯ glaze json misc/test-data/[123].json --jq 'del(.c) | .d = .b'
+-----+-----+-----+
| b   | d   | a   |
+-----+-----+-----+
| 2   | 2   | 1   |
| 20  | 20  | 10  |
| 200 | 200 | 100 |
+-----+-----+-----+
```

You can sum the content of each c field (which is an array of numbers originally):

``` 
❯ glaze json misc/test-data/[123].json --jq '.c = (.c | add)'
+-----+-----+-----+-----+-----+
| b   | c   | a   | d.e | d.f |
+-----+-----+-----+-----+-----+
| 2   | 12  | 1   | 6   | 7   |
| 20  | 120 | 10  | 60  | 70  |
| 200 | 300 | 100 |     |     |
+-----+-----+-----+-----+-----+
```