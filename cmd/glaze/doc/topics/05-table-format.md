---
Title: Table Output Format
Slug: table-format
Short: |
  Glazed supports a wide variety of table output formats. Besides HTML and Markdown, 
  it leverages the go-pretty library to provide a rich set of styled terminal output.
Topics:
- output
IsTemplate: false
IsTopLevel: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Glazed supports a wide variety of table output formats.

Besides HTML and Markdown, it leverages the go-pretty library to provide a rich set of styled terminal output.

`--output table` is the default and so wouldn't need to be explicitly specified in most cases.

## Markdown output

Use `--output table --table-format markdown` to output a table in markdown format.

``` 
❯ glaze json misc/test-data/[123].json --table-format markdown
| b   | c          | a   | d.e | d.f |
| --- | ---------- | --- | --- | --- |
| 2   | 3, 4, 5    | 1   | 6   | 7   |
| 20  | 30, 40, 50 | 10  | 60  | 70  |
| 200 | 300        | 100 |     |     |
```

By piping it into [glow](https://github.com/charmbracelet/glow), you can get a nice looking markdown table:

```
❯ glaze json misc/test-data/[123].json --table-format markdown | glow
                                                                                                                    
     A  │  B  │     C      │ D E │ D F                                                                                
  ──────┼─────┼────────────┼─────┼──────                                                                              
      1 │   2 │ 3, 4, 5    │   6 │   7                                                                                
     10 │  20 │ 30, 40, 50 │  60 │  70                                                                                
    100 │ 200 │        300 │     │         
```

## HTML output

Use `--output table --table-format html` to output a table in HTML format.

```
❯ glaze json misc/test-data/[123].json --table-format html           
<table class="termtable">
<thead>
<tr><th>a</th><th>b</th><th>c</th><th>d.e</th><th>d.f</th></tr>
</thead>
<tbody>
<tr><td>1</td><td>2</td><td>3, 4, 5</td><td>6</td><td>7</td></tr>
<tr><td>10</td><td>20</td><td>30, 40, 50</td><td>60</td><td>70</td></tr>
<tr><td>100</td><td>200</td><td>300</td><td></td><td></td></tr>
</tbody>
</table>
```

## Pretty styles

The go-pretty library supports a wide variety of styles. You can use the `--table-style` flag to select a style.

```
❯ glaze json misc/test-data/[123].json --table-style double
╔═════╦═════╦════════════╦═════╦═════╗
║ A   ║ B   ║ C          ║ D.E ║ D.F ║
╠═════╬═════╬════════════╬═════╬═════╣
║ 1   ║ 2   ║ 3, 4, 5    ║ 6   ║ 7   ║
║ 10  ║ 20  ║ 30, 40, 50 ║ 60  ║ 70  ║
║ 100 ║ 200 ║ 300        ║     ║     ║
╚═════╩═════╩════════════╩═════╩═════╝
```

```
❯ glaze json misc/test-data/[123].json --table-style light
┌─────┬────────────┬─────┬─────┬─────┐
│ B   │ C          │ A   │ D.F │ D.E │
├─────┼────────────┼─────┼─────┼─────┤
│ 2   │ 3, 4, 5    │ 1   │ 7   │ 6   │
│ 20  │ 30, 40, 50 │ 10  │ 70  │ 60  │
│ 200 │ 300        │ 100 │     │     │
└─────┴────────────┴─────┴─────┴─────┘
```

```
❯ glaze json misc/test-data/[123].json --table-style rounded
╭─────┬─────┬────────────┬─────┬─────╮
│ A   │ B   │ C          │ D.E │ D.F │
├─────┼─────┼────────────┼─────┼─────┤
│ 1   │ 2   │ 3, 4, 5    │ 6   │ 7   │
│ 10  │ 20  │ 30, 40, 50 │ 60  │ 70  │
│ 100 │ 200 │ 300        │     │     │
╰─────┴─────┴────────────┴─────┴─────╯
```

Pretty styles can be loaded from YAML files. To get a starting point, use the --print-table-style option:

``` 
❯ glaze json --table-style light --print-table-style misc/test-data/1.json | tee /tmp/light.yaml
name: StyleLight
box:
    bottom-left: └
    bottom-right: ┘
    bottom-separator: ┴
    left: │
    left-separator: ├
    middle-horizontal: ─
    middle-separator: ┼
    middle-vertical: │
    padding-left: ' '
    padding-right: ' '
    page-separator: "\n"
    right: │
    right-separator: ┤
    top-left: ┌
    top-right: ┐
    top-separator: ┬
    unfinished-row: ' ≈'
color: {}
format:
    footer: upper
    header: upper
    row: default
options:
    draw-border: true
    separate-columns: true
    separate-footer: true
    separate-header: true
title:
    align: default
    format: default
```

You can then reuse that style:

```
❯ glaze json --table-style-file /tmp/light.yaml misc/test-data/1.json                     
┌───┬───┬─────────┬─────┬─────┐
│ A │ B │ C       │ D.E │ D.F │
├───┼───┼─────────┼─────┼─────┤
│ 1 │ 2 │ 3, 4, 5 │ 6   │ 7   │
└───┴───┴─────────┴─────┴─────┘
```