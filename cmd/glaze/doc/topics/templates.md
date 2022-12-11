---
Title: Templates
Slug: templates
Short: Use go templates to customize output
Topics:
- templates
Commands:
- json
- yaml
Flags:
- template
- template-field
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

You can use go templates to either create a new field (called _0 per default).
Per default, the templates are applied at the input level, when rows
are actually still full blown objects (if reading in from JSON for example).

```
❯ glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d.f}}'

+---------------------+
| _0                  |
+---------------------+
| 1-2: 7              |
| 10-20: 70           |
| 100-200: <no value> |
+---------------------+
```

You can also apply templates at the row level, once the input has been flattened.
In this case, because flattened columns contain the symbol `.`, fields get renamed
to use the symbol `_` as a separator.

``` 
❯ glaze json misc/test-data/[123].json --template '{{.a}}-{{.b}}: {{.d_f}}' \
  --use-row-templates --fields a,_0 \
  --output csv
a,_0
1,1-2: 7
10,10-20: 70
100,100-200: <no value>
```

Instead of just adding / replacing everything with a single field `_0`, you
can also specify multiple templates using the `--template-field` argument, which has
the form `COLNAME:TEMPLATE`.

``` 
❯ glaze json misc/test-data/[123].json \
    --template-field 'foo:{{.a}}-{{.b}},bar:{{.d_f}}' \
    --use-row-templates --fields a,foo,bar
+-----+---------+------------+
| a   | foo     | bar        |
+-----+---------+------------+
| 1   | 1-2     | 7          |
| 10  | 10-20   | 70         |
| 100 | 100-200 | <no value> |
+-----+---------+------------+
```

To make things a bit more readable, especially when doing a lot of template transformations,
you can also load field templates from a yaml file using the `@` symbol.

``` 
❯ glaze json misc/test-data/[123].json \
    --template-field '@misc/template-field-object.yaml' \
    --output json
[
  {
    "barbaz": "6 - 7",
    "foobar": "1"
  },
  {
    "barbaz": "60 - 70",
    "foobar": "10"
  },
  {
    "barbaz": "\u003cno value\u003e - \u003cno value\u003e",
    "foobar": "100"
  }
]
```

