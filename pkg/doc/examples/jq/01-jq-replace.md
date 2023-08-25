---
Title: Use jq to replace one column with another
Slug: jq-replace
Short: |
  ```
  glaze json misc/test-data/[123].json --jq '.b = .c'
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
 
You can use `--jq` expressions to replace one column with another.

The `--jq` flag is a way to use the [jq](https://stedolan.github.io/jq/) command line tool to manipulate
each row before it is passed on.

When the jq expressions returns an array, each element of that array is treated as a new row.

```
❯ glaze json misc/test-data/[123].json --jq '.b = .c'
+------------+------------+-----+-----+-----+
| b          | c          | a   | d.e | d.f |
+------------+------------+-----+-----+-----+
| 3, 4, 5    | 3, 4, 5    | 1   | 6   | 7   |
| 30, 40, 50 | 30, 40, 50 | 10  | 60  | 70  |
| 300        | 300        | 100 |     |     |
+------------+------------+-----+-----+-----+
```

You can add 50 to the a column:

``` 
❯ glaze json misc/test-data/[123].json --jq '.a += 50'
+-----+------------+-----+-----+-----+
| b   | c          | a   | d.e | d.f |
+-----+------------+-----+-----+-----+
| 2   | 3, 4, 5    | 51  | 6   | 7   |
| 20  | 30, 40, 50 | 60  | 60  | 70  |
| 200 | 300        | 150 |     |     |
+-----+------------+-----+-----+-----+
```

Replace c with the first element followed by 700:

```
❯ glaze json misc/test-data/[123].json --jq '.c = [.c[0], 700]'
+-----+-----+----------+-----+-----+
| a   | b   | c        | d.e | d.f |
+-----+-----+----------+-----+-----+
| 1   | 2   | 3, 700   | 6   | 7   |
| 10  | 20  | 30, 700  | 60  | 70  |
| 100 | 200 | 300, 700 |     |     |
+-----+-----+----------+-----+-----+
```