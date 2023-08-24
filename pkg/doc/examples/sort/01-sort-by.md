---
Title: Sort by columns
Slug: sort-by
Short: |
  ```
  glaze json misc/test-data/sort.json --input-is-array --sort-by city,-name
  ```
Topics:
  - sort
Commands:
  - yaml
  - json
  - csv
Flags:
  - sort-by
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: Example
---

You can sort an output table that has columns of consistent type (int, float or string are allowed).
The list of columns is given as a comma-separated list of names, with a `-` prefix for descending order.

```
❯ glaze json misc/test-data/sort.json --input-is-array --sort-by city,-name
+---------+-----+----------+----+
| name    | age | city     | id |
+---------+-----+----------+----+
| Peter   | 40  | Boston   | 2  |
| Hannah  | 60  | Chicago  | 4  |
| Amy     | 50  | Chicago  | 3  |
| Michael | 70  | Houston  | 5  |
| John    | 30  | New York | 1  |
+---------+-----+----------+----+
```