---
Title: Rename one or multiple columns
Slug: rename-column
Short: |
  ```
  glaze yaml misc/test-data/test.yaml --input-is-array --rename baz:blop
  ```
Commands:
- yaml
- json
Flags:
- rename
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can easily rename a column (or multiple columns) by providing a list of `:` 
separated pairs using the `--rename` flag.

---

```
❯ glaze yaml misc/test-data/test.yaml --input-is-array --rename baz:blop
+------+-----+-----+-----+---------+
| blop | d.e | d.f | foo | foobar  |
+------+-----+-----+-----+---------+
| 2    | 6   | 7   | 1   | [3 4 5] |
| 20   | 60  | 70  | 10  |         |
| 200  |     |     |     | [300]   |
+------+-----+-----+-----+---------+
```

To rename multiple columns, use:

```
❯ glaze yaml misc/test-data/test.yaml --input-is-array --rename baz:blop,d.e:dang
+------+-----+------+-----+---------+
| blop | d.f | dang | foo | foobar  |
+------+-----+------+-----+---------+
| 2    | 7   | 6    | 1   | [3 4 5] |
| 20   | 70  | 60   | 10  |         |
| 200  |     |      |     | [300]   |
+------+-----+------+-----+---------+
```