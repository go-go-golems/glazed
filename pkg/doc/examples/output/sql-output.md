---
Title: SQL Output Flags
Slug: sql-output-flags
Command: glaze
Short: |
  Learn how to use the flags that control SQL output in the `glaze` program.
Topics:
- sql
- output
Commands:
- json
Flags:
- sql-table-name
- with-upsert
- sql-split-by-rows
IsTemplate: false
IsTopLevel: true
ShowPerDefault: false
SectionType: Example
---

The `glaze` program provides flags that allow you to control the SQL output format.

Here are the descriptions of the flags:

- `sql-table-name`: Specifies the table name for SQL output. The default value is "output".
- `with-upsert`: Uses upsert instead of insert for SQL output. The default value is false.
- `sql-split-by-rows`: Splits SQL output by the specified number of rows. The default value is 1000.

Note that only the columns of the first row are used for the output statements.
Missing columns are replaced by NULL.
Integer, string, and boolean values are encoded as SQL,
while other values are encoded to JSON and inserted as strings.

## Use the default table name and insert statement:

```
❯ glaze json misc/test-data/[123].json --output sql
INSERT INTO output (a, b, c, d) VALUES
(1, 2, '[3,4,5]', '{"e":6,"f":7}')
, (10, 20, '[30,40,50]', '{"e":60,"f":70}')
, (100, 200, '[300]', NULL)
;
```

This example uses the default table name "output" and generates an insert statement for each row in the JSON input.

## Specify a custom table name for the SQL output:

```
❯ glaze json misc/test-data/[123].json --output sql --sql-table-name foobar
INSERT INTO foobar (a, b, c, d) VALUES
(1, 2, '[3,4,5]', '{"e":6,"f":7}')
, (10, 20, '[30,40,50]', '{"e":60,"f":70}')
, (100, 200, '[300]', NULL)
;
```

In this example, the `--sql-table-name` flag is used to specify a custom table name "foobar" for the SQL output.

## Use upsert instead of insert for SQL output:

```
❯ glaze json misc/test-data/[123].json --output sql --sql-table-name foobar --with-upsert
INSERT INTO foobar (a, b, c, d) VALUES
(1, 2, '[3,4,5]', '{"e":6,"f":7}')
, (10, 20, '[30,40,50]', '{"e":60,"f":70}')
, (100, 200, '[300]', NULL)
ON DUPLICATE KEY UPDATE
a = VALUES(a),
b = VALUES(b),
c = VALUES(c),
d = VALUES(d);
```

By using the `--with-upsert` flag, the SQL output will use upsert statements instead of insert statements.

## Split SQL output by a specified number of rows:

```
❯ glaze json misc/test-data/[123].json --output sql --sql-split-by-rows 2
INSERT INTO output (a, b, c, d) VALUES
(1, 2, '[3,4,5]', '{"e":6,"f":7}')
, (10, 20, '[30,40,50]', '{"e":60,"f":70}')
;
INSERT INTO output (a, b, c, d) VALUES
(100, 200, '[300]', NULL)
;
```

In this example, the `--sql-split-by-rows` flag is used to split the SQL output into multiple statements,
with each statement containing the specified number of rows.

