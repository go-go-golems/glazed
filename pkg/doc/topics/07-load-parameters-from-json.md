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
This works with any command that uses glazed to hook up to cobra.

This allows you to store common parameter configurations and reuse them across commands.

To load parameters from JSON, use the `--load-parameters-from-json` flag followed by the path to your JSON file:

```
command --load-parameters-from-json parameters.json [other arguments]
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

## ParseCommandFromMap implementation and usage.

The `ParseCommandFromMap` function in `cmds/map.go` is used to parse command parameters from a map structure, such as when loading parameters from JSON.

It takes a `CommandDescription`, a `map[string]interface{}` of parameters, and returns:

- A map of `ParsedParameterLayer` structs for each layer
- A combined map of all parameter values
- Any error encountered

Here is how it works:

1. It iterates through each layer in the `CommandDescription`
2. For layers that implement the `JSONParameterLayer` interface, it calls `ParseFlagsFromJSON` to parse values from the map into a `ParsedParameterLayer`
3. It adds the parsed layer to the output map
4. It also copies layer parameters into the combined parameter map
5. After parsing layers, it parses any remaining flags and arguments using the `CommandDescription` directly
6. Required arguments are checked and assigned defaults if needed

To use it in your own code:

```go
import "github.com/go-go-golems/glazed/pkg/cmds"

// cmd is your CommandDescription 

params := map[string]interface{}{
  "output": "json",
  "fields": ["id", "name"] 
}

layers, allParams, err := cmds.ParseCommandFromMap(cmd, params)
```

The returned `layers` map contains the parsed layers, while `allParams` contains all parameters combined.