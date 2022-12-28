---
Title: Rename columns using a YAML file
Slug: rename-yaml
Short: |
  ```
  glaze yaml misc/test-data/test.yaml --input-is-array \
  --rename-yaml misc/rename.yaml
  ```
Commands:
- yaml
- json
Flags:
- rename-yaml
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Example
---
You can specify how to rename column using a YAML file that has the following format:

```yaml
renames:
  oldColumnName: newColumnName
  otherColumnName: anotherNewColumName
regexpRenames:
  "^foo(.*}": "${1}_replacement"
```

The same rules as for the `--rename` and `--rename-regexp` flags apply.

---

```
❯ cat misc/rename.yaml                                    
renames:
  foo: bar
regexpRenames:
  "^bar$": "baz"
  ".(o.b).": "$1"%                                                                    

❯ glaze yaml misc/test-data/test.yaml --input-is-array \
  --rename-yaml misc/rename.yaml
+-----+-----+-----+-----+---------+
| bar | baz | d.e | d.f | oobr    |
+-----+-----+-----+-----+---------+
| 1   | 7   | 6   | 7   | [3 4 5] |
| 10  | 70  | 60  | 70  |         |
|     |     |     |     | [300]   |
+-----+-----+-----+-----+---------+
```