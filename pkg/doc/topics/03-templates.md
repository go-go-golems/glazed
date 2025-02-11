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

Additional data can be provided using the `--template-data` argument
which should point to a JSON/CSV/YAML file containing the data to be 
loaded.

``` 
❯ glaze json misc/test-data/[123].json \
    --template-data misc/test-data/book.json \
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

## Templating functions

Glazed uses the [sprig](http://masterminds.github.io/sprig/) templating 
library to provide many useful functions.

Furthermore, there is support for a variety of additional templating functions.

### Number functions

The following functions are available for computing inside templates:

- `add(a, b interface{})` - Add two numbers (supports int, uint, float)
- `sub(a, b interface{})` - Subtract b from a (supports int, uint, float)
- `mul(a, b interface{})` - Multiply two numbers (supports int, uint, float)
- `div(a, b interface{})` - Divide a by b (supports int, uint, float)
- `parseFloat(s string) float64` - Parse string to float64
- `parseInt(s string) int64` - Parse string to int64
- `currency(n interface{}) string` - Format number as currency with 2 decimal places

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

- `trim(s string) string` - Remove spaces from both ends
- `trimRightSpace(s string) string` - Remove spaces from the right
- `trimTrailingWhitespaces(s string) string` - Remove trailing whitespaces
- `rpad(s string, n int) string` - Right pad a string with spaces
- `quote(s string) string` - Quote a string with backticks
- `stripNewlines(s string) string` - Remove newlines
- `quoteNewlines(s string) string` - Replace newlines with \n
- `toUpper(s string) string` - Convert to uppercase
- `toLower(s string) string` - Convert to lowercase
- `replaceRegexp(s, old, new string) string` - Replace using regexp
- `padLeft(s string, n int) string` - Pad with spaces on the left
- `padRight(s string, n int) string` - Pad with spaces on the right
- `padCenter(s string, n int) string` - Center text with spaces
- `toUrlParameter(v interface{}) string` - Convert value to URL parameter format
- `toYaml(v interface{}) string` - Convert value to YAML format
- `indentBlock(indent int, value string) string` - Indent text block

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

- `bold(s string) string` - Make text bold with **
- `italic(s string) string` - Make text italic with *
- `underline(s string) string` - Make text underlined with __
- `strikethrough(s string) string` - Make text strikethrough with ~~
- `code(s string) string` - Make text code with backticks
- `codeBlock(s, lang string) string` - Make text a code block with language

### Date functions

- `toDate(v interface{}) string` - Convert value to YYYY-MM-DD format (supports string and time.Time)

### Random functions

The following functions are available for randomization:

- `randomChoice(list interface{}) interface{}` - Return random element from list
- `randomSubset(list interface{}, n int) interface{}` - Return n random elements from list
- `randomPermute(list interface{}) interface{}` - Return random permutation of list
- `randomInt(min, max interface{}) int` - Return random integer between min and max (inclusive)
- `randomFloat(min, max interface{}) float64` - Return random float between min and max
- `randomBool() bool` - Return random boolean
- `randomString(length int) string` - Return random alphanumeric string of given length
- `randomStringList(count, minLength, maxLength int) []string` - Return list of random strings

```
❯ glaze json misc/test-data/[123].json \
    --template-field 'random:{{ randomChoice .c }},subset:{{ randomSubset .c 2 }}' \
    --use-row-templates
+-----+--------+-------------+
| a   | random | subset      |
+-----+--------+-------------+
| 1   | 4      | [3 5]       |
| 10  | 30     | [40 50]     |
| 100 | 300    | [300]       |
+-----+--------+-------------+
```

### Style functions

- `styleBold(s string) string` - Make text bold with ANSI escape codes

### Sprig functions

In addition to the functions above, all functions from the [sprig](http://masterminds.github.io/sprig/) library are available.
These include many useful functions for string manipulation, math, lists, dictionaries, and more.

