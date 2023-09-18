---
Title: Loading Parameters from JSON
Slug: load-parameters-json
Short: Explains how to load parameters from a JSON file.
Topics:
- User Guide
- Parameters
Flags: 
- load-parameters-from-json
IsTemplate: false
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

In addition to specifying parameters via command line flags, you can also load parameters from a JSON file.

This allows you to store common parameter configurations and reuse them across commands.

To load parameters from JSON, use the `--load-parameters-from-json` flag followed by the path to your JSON file:

```
glazed [command] --load-parameters-from-json parameters.json [other arguments]
```

The JSON file should contain a JSON object where the keys are parameter names and the values are the parameter values you want to set.

For example:

```json
{
   "fields": ["id", "name"],
   "output": "json"
}

```

This will set the `fields` and `output` parameters as if they had been passed via the command line.
However, flags passed on the command line will overwrite values in the JSON file.

## Example

```
❯ glaze json misc/test-data/[123].json 
+-----+-----+------------+-----------+
| a   | b   | c          | d         |
+-----+-----+------------+-----------+
| 1   | 2   | 3, 4, 5    | e:6,f:7   |
| 10  | 20  | 30, 40, 50 | e:60,f:70 |
| 100 | 200 | 300        |           |
+-----+-----+------------+-----------+

❯ cat /tmp/test-json.json 
{
	"fields": [ "a", "b" ],
	"output": "json"
}

❯ glaze json --load-parameters-from-json /tmp/test-json.json misc/test-data/[123].json 
[
{
  "a": 1,
  "b": 2
}
{
  "a": 10,
  "b": 20
}
{
  "a": 100,
  "b": 200
}
]
```
