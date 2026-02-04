---
Title: Create cliopatra YAML
Slug: cliopatra-capture
Short: |
  ```
    glaze yaml misc/test-data/test.yaml \
        --input-is-array --rename baz:blop \
        --create-cliopatra | tee /tmp/yaml-rename.yaml
  ```
Topics:
- cliopatra
Commands:
- yaml
- json
- csv
Flags:
- create-cliopatra
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: Example
---

You can easily create a skeleton for use with [cliopatra](https://github.com/go-go-golems/cliopatra),
which is a tool to execute programs by providing a short YAML with default flag values.

It also allows overriding the environment, providing expected outputs for golden testing, rendering out
CLI commands from within templates, and more.

Passing `--create-cliopatra` to any glazed Command will capture the given flags and argments passed to it,
map it to the Definition the Command uses, and create a YAML file with the default values.

``` 
❯ glaze yaml misc/test-data/test.yaml \
     --input-is-array --rename baz:blop \
     --create-cliopatra | tee /tmp/yaml-rename.yaml
name: glaze
verbs:
    - yaml
description: Format YAML data
flags:
    - name: rename
      flag: rename
      short: Rename fields (list of oldName:newName)
      type: stringList
      value:
        - baz:blop
    - name: input-is-array
      flag: input-is-array
      short: Input is an array of objects
      type: bool
      value: true
args:
    - name: input-files
      flag: input-files
      short: ""
      type: stringList
      value:
        - misc/test-data/test.yaml

❯ cliopatra run /tmp/yaml-rename.yaml
+------+---------+-----+-----+-----+-----+
| blop | foobar  | foo | bar | d.f | d.e |
+------+---------+-----+-----+-----+-----+
| 2    | 3, 4, 5 | 1   | 7   | 7   | 6   |
| 20   |         | 10  | 70  | 70  | 60  |
| 200  | 300     |     |     |     |     |
+------+---------+-----+-----+-----+-----+
```