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
- template-file
- template-data
- output
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

## Replacing a row with a single new field

You can use go templates to create a new field (called _0 per default).
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

## Adding a new field for each with row templates

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

## Adding/replacing multiple fields with templates

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

## Rendering a single template output 

You can also use a template output formatter that takes a `--template-file` argument.
The template is rendered with an object that has a `rows` field.

``` 
❯ cat misc/template-file-example.tmpl.md 
# Counts Rows {{ len .rows }}
{{ range $row := .rows }}
## Row {{.b}}

- c: {{.c}}
- a: {{.a}}
{{ end }}%                                                                                                    

❯ glaze json misc/test-data/[123].json \
     --output template \
     --template-file misc/template-file-example.tmpl.md 
# Counts Rows 3

## Row 2

- c: [3 4 5]
- a: 1

## Row 20

- c: [30 40 50]
- a: 10

## Row 200

- c: [300]
- a: 100

```

Additional data can be provided using the `--template-data` argument,
either as `:` separated pairs, themselves separated by commas,
or by providing a json or yaml file using a path preceded by a `@`.

``` 
❯ glaze json misc/test-data/[123].json \
    --template-data @misc/test-data/book.json \
    --template-file misc/template-file-example2.tmpl.md \
    --output template
# The Raven

Author: Edgar Allan Poe


## Row 2

- c: [3 4 5]
- a: 1
  
## Row 20

- c: [30 40 50]
- a: 10
  
## Row 200

- c: [300]
- a: 100
```

or 

``` 
❯ glaze json misc/test-data/[123].json \
     --template-data "author:J.R.R Tolkien,title:The Lord of the Rings" \
     --template-file misc/template-file-example2.tmpl.md \
     --output template
# The Lord of the Rings

Author: J.R.R Tolkien


## Row 2

- c: [3 4 5]
- a: 1
  
## Row 20

- c: [30 40 50]
- a: 10
  
## Row 200

- c: [300]
- a: 100
  %       
  ```