---
Title: Examples of how to use the pagination flags skip and limit
Slug: skip-limit
Short: |
  This document provides examples of how to use the pagination flags 'glazed-skip' and 'glazed-limit' in Glazed.
Topics:
- pagination
- command line
Flags:
- glazed-skip
- glazed-limit
IsTopLevel: false
ShowPerDefault: false
SectionType: Example
---

Glazed provides pagination options through the use of 'glazed-skip' and 'glazed-limit' flags.

## Skip flag

The 'glazed-skip' flag allows you to skip a certain number of records. 

For example, to skip the first record, you would use the following command:

``` 
❯ glaze json misc/test-data/[123].json --glazed-skip 1
+-----+-----+------------+-----------+
| a   | b   | c          | d         |
+-----+-----+------------+-----------+
| 10  | 20  | 30, 40, 50 | e:60,f:70 |
| 100 | 200 | 300        |           |
+-----+-----+------------+-----------+
```

## Limit flag

The 'glazed-limit' flag allows you to limit the number of records returned. 

For example, to limit the output to the first two records, you would use the following command:

``` 
❯ glaze json misc/test-data/[123].json --glazed-limit 2         
+----+----+------------+-----------+
| a  | b  | c          | d         |
+----+----+------------+-----------+
| 1  | 2  | 3, 4, 5    | e:6,f:7   |
| 10 | 20 | 30, 40, 50 | e:60,f:70 |
+----+----+------------+-----------+
```

## Using both flags

You can also use both flags together. For example, to skip the first two records and limit the output to one record, you would use the following command:

``` 
❯ glaze json misc/test-data/[123].json --glazed-limit 1 --glazed-skip 2
+-----+-----+-----+
| a   | b   | c   |
+-----+-----+-----+
| 100 | 200 | 300 |
+-----+-----+-----+
```