---
Title: Remove duplicates
Slug: remove-duplicates
Short: |
  ```
  glaze json misc/test-data/duplicates.json --input-is-array --remove-duplicates a,b,c --fields a,b,c
  ```
Topics:
- remove-duplicates
Commands:
- yaml
- json
- csv
Flags:
- remove-duplicates
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: Example
---
If your data is sorted, you can remove duplicates by specifying a list of columns.
For each row, glazed will compare the value of those columns to the values of the previous row, 
and skip the row if they are identical.

---

```
❯ glaze json misc/test-data/duplicates.json --input-is-array --remove-duplicates a,b,c --fields a,b,c
+---+---+---+
| a | b | c |
+---+---+---+
| 1 | 2 | 3 |
| 7 | 5 | 9 |
| 1 | 5 | 6 |
| 1 | 2 | 3 |
+---+---+---+
```

Or, only on a single column:

```
❯ glaze json misc/test-data/duplicates.json --input-is-array --remove-duplicates a --fields a
+---+---+---+
| a | b | c |
+---+---+---+
| 1 | 2 | 3 |
| 7 | 5 | 9 |
| 1 | 5 | 6 |
+---+---+---+
```