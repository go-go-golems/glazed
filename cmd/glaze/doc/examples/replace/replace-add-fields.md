---
Title: Add constant columns to your output
Slug: replace-add-field
Short: |
  ```
  glaze yaml misc/test-data/test.yaml --input-is-array \
       --add-fields name:Hello
  ```
Topics:
- replace
Commands:
- yaml
- json
- csv
Flags:
- add-fields
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---

You can easily add constant fields to your output, which can be useful for tagging
your data in a shell script, for example.

```

❯ glaze yaml misc/test-data/test.yaml --input-is-array \
      --add-fields name:Hello,yoyo:bar
+---------+-----+-----+-----+-----+-----+------+-------+
| foobar  | foo | bar | baz | d.e | d.f | yoyo | name  |
+---------+-----+-----+-----+-----+-----+------+-------+
| 3, 4, 5 | 1   | 7   | 2   | 6   | 7   | bar  | Hello |
|         | 10  | 70  | 20  | 60  | 70  | bar  | Hello |
| 300     |     |     | 200 |     |     | bar  | Hello |
+---------+-----+-----+-----+-----+-----+------+-------+
```