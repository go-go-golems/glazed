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

Please note that the `--template-field` flag is parsed by cobra as a slice value,
which uses the golang csv parser. If you want to use the `"` character in your field, you need to
jump through a few hoops. See the end of this file for a list of functions that can be used in 
templates.

```
❯ glaze json misc/test-data/book.json \
   --template-field '"french_author:{{ replace .author ""Poe"" ""Pouet"" }}"' \
   --use-row-templates
+-----------------+-----------+-------------------+
| author          | title     | french_author     |
+-----------------+-----------+-------------------+
| Edgar Allan Poe | The Raven | Edgar Allan Pouet |
+-----------------+-----------+-------------------+
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

## Templating functions

Glazed uses the [sprig](http://masterminds.github.io/sprig/) templating 
library to provide many useful functions.

Furthermore, there is support for a variety of legacy templating functions that
will be phased out in the future.

### Number functions

The following functions are available for computing inside template: 
- add
- sub
- mul
- div
- parseFloat(s)
- parseInt(s)


```
❯ glaze json misc/test-data/[123].json \
    --template-field 'foo:{{.a}} + {{.b}} = {{add .a .b}},bar:{{.a}} * {{ .d_f }} = {{ mul .a .d_f}}' \
    --use-row-templates --fields a,foo,bar
+-----+-----------------+-------------------------------+
| a   | foo             | bar                           |
+-----+-----------------+-------------------------------+
| 1   | 1 + 2 = 3       | 1 * 7 = 7                     |
| 10  | 10 + 20 = 30    | 10 * 70 = 700                 |
| 100 | 100 + 200 = 300 | 100 * <no value> = 0          |
+-----+-----------------+-------------------------------+
```

### String functions 

The following functions are available to manipulate strings:
- trim(s) - remove spaces
- trimRightSpace(s) - remove spaces from the right
- trimTrailingWhitespaces(s) - remove trailing whitespaces
- rpad(s, n) - right pad a string
- quote(s) - quote a string
- stripNewlines(s) - remove newlines
- toUpper(s) - convert to uppercase
- toLower(s) - convert to lowercase
- replace(s, old, new) - replace old with new in s
- replaceRegexp(s, old, new) - replace old with new in s using regexp
- padLeft(s, n) - pad s with spaces on the left
- padRight(s, n) - pad s with spaces on the right

```
❯ glaze json misc/test-data/book.json \
   --template-field '"robot_author:{{ replaceRegexp .author ""([A-Za-z])"" ""$1."" | toUpper }}"' \
   --use-row-templates
+-----------------+-----------+------------------------------+
| author          | title     | robot_author                 |
+-----------------+-----------+------------------------------+
| Edgar Allan Poe | The Raven | E.D.G.A.R. A.L.L.A.N. P.O.E. |
+-----------------+-----------+------------------------------+
```

### Markdown functions

The following templates make generating markdown easier:

- bold(s) - make text bold
- italic(s) - make text italic
- underline(s) - make text underlined
- strikethrough(s) - make text strikethrough
- code(s) - make text code
- codeBlock(s, lang) - make text a code block

### Miscellanous functions

- currency(n) - format a number as a currency (int, float, uint)