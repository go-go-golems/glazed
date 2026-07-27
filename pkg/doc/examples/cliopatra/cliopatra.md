---
Title: Create cliopatra YAML
Slug: cliopatra-capture
Short: |
  ```
    glaze yaml misc/test-data/test.yaml \
        --input-is-array --format json \
        --create-cliopatra | tee /tmp/yaml-json.yaml
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

You can create a skeleton for use with [cliopatra](https://github.com/go-go-golems/cliopatra),
which executes programs from a small YAML description containing arguments and default flag values.

Passing `--create-cliopatra` to a Glazed command captures its current flags and arguments, maps them to
the command's field definitions, and emits a reusable YAML invocation.

```console
$ glaze yaml misc/test-data/test.yaml \
    --input-is-array --format json \
    --create-cliopatra > /tmp/yaml-json.yaml

$ cliopatra run /tmp/yaml-json.yaml
[
  {
    "foo": 1,
    "baz": 2
  }
]
```

Structured commands expose `--format`, `--output-fields`, and `--max-output-rows`; all three can be
captured like ordinary typed command fields.
