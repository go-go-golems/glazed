---
Title: Rename one or multiple columns using a regular expression
Slug: rename-regexps
Short: |
  ```
  glaze yaml misc/test-data/test.yaml --input-is-array \
  --rename-regexp '^(.*)bar:${1}blop'
  ```
Commands:
- yaml
- json
Flags:
- rename-regexp
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can use one (or multiple) regexps to do column renames as well. 

You can even use capture groups and specify `$1` or `${1}` in the replacement 
string to expand the resulting column name with the expanded group.

Regexp replacements are performed after regular renames (because these 
are considered to be "exact" matches). 

Regular expression replacement are performed in order, and stop after the 
first match that actually modifies the column name.

---

```
❯ glaze yaml misc/test-data/test.yaml --input-is-array \
  --rename-regexp '^(.*)bar:${1}blop'
+-----+------+-----+-----+-----+---------+
| baz | blop | d.e | d.f | foo | fooblop |
+-----+------+-----+-----+-----+---------+
| 2   | 7    | 6   | 7   | 1   | [3 4 5] |
| 20  | 70   | 60  | 70  | 10  |         |
| 200 |      |     |     |     | [300]   |
+-----+------+-----+-----+-----+---------+
```

```
❯ glaze yaml misc/test-data/test.yaml --input-is-array \
  --rename-regexp '^(.*)bar:${1}blop','b..:blip'
+------+------+-----+-----+-----+---------+
| blip | blop | d.e | d.f | foo | fooblop |
+------+------+-----+-----+-----+---------+
| 2    | 7    | 6   | 7   | 1   | [3 4 5] |
| 20   | 70   | 60  | 70  | 10  |         |
| 200  |      |     |     |     | [300]   |
+------+------+-----+-----+-----+---------+
```