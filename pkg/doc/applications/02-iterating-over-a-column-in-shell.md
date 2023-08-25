---
Title: Iterating over a column with a shell script
Slug: iterating-over-column
Short: Shows different ways glaze can be used to iterate over columns with a shell
  script
Topics:
- glaze
- json
- templates
Commands:
- json
- yaml
Flags:
- select
- select-template
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: Application
---

It's often useful to iterate over a column with a shell script, for example
in order to print out different names.

## Iterating over a single column

If you want to quickly iterate over a single column, use the `--select` flag.
It is a shortcut for the normal output formatter and field selection.

```
❯ for i in $(glaze json misc/test-data/[123].json --select a); do echo $i; done

1
10
100
```

## Using a template to process entire rows

You can also use a go template to output more complicated values.

```
❯ glaze json misc/test-data/[123].json \
    --select-template '{{.a}} is less than {{.b}}' | 
    while read i; do 
        echo "This was the template output: '$i'"
	done
This was the template output: '1 is less than 2'
This was the template output: '10 is less than 20'
This was the template output: '100 is less than 200'
```