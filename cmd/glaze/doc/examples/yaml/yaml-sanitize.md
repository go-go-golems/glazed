---
Title: Sanitize YAML 
Slug: yaml-sanitize
Short: |
  ```
  glaze yaml --sanitize misc/broken.yaml --output yaml
  ```
Commands:
- yaml
Flags:
- sanitize
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---

Often when interacting with LLMs generating YAML, various little imperfections
make their way through. The most common are:

- Interleaved markdown and YAML.
- Header or footer text that is not YAML
- Not quoting strings that contain special characters

With the `--sanitize` flag, you can attempt to cleanup YAML before processing it further.
This is quite hacky, and may not work for all cases.

```
❯ glaze yaml --sanitize misc/broken.yaml --output yaml
- a: 1
  b:
    - 2
    - 3
    - d: 2
      e: 'fourbe: courbevoie'
      f: |
        - 4
        - 5: 234234
        foobar
      g: 2
```

It can also handle multiline strings:

```
❯ glaze yaml --sanitize misc/broken2.yaml --output yaml
- a: 1
  b: 'foobar: blabla'
  c: |
    foobar: foobar
  d: '&lkjsld'
```
