---
Title: Using GatherFlagsFromStringList Function
Slug: using-gatherflagsfromstringlist
Short: |
  Learn how to use the GatherFlagsFromStringList function for parsing command line flags programmatically. This function does not require the use of cobra bindings, but instead just requires a list of strings.
Topics:
- Command Line
- Flags
- Parsing
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

# Using GatherFlagsFromStringList Function

The `GatherFlagsFromStringList` function allows you to parse command line flags programmatically. This function
does not require the use of cobra bindings, but instead just requires a list of strings.

## Function Signature

```go
func GatherFlagsFromStringList(
	args []string,
	params []*ParameterDefinition,
	onlyProvided bool,
	ignoreRequired bool,
	prefix string,
) (map[string]interface{}, []string, error)
```

## Parameters

- `args`: a slice of strings representing the command line arguments.
- `params`: a slice of `*ParameterDefinition` representing the parameter definitions.
- `onlyProvided`: a boolean indicating whether to only include flags that were provided in the command line arguments.
- `ignoreRequired`: a boolean indicating whether to ignore required flags that were not provided in the command line arguments.
- `prefix`: a string to be added as a prefix to all flag names.

## Return Values

The function returns a map where the keys are the parameter names and the values are the parsed values. If a flag is not recognized or its value cannot be parsed, an error is returned.

## Usage

Here is an example of how to use the function:

```go
params := []*ParameterDefinition{
   {Name: "verbose", ShortFlag: "v", Type: ParameterTypeBool},
   {Name: "output", ShortFlag: "o", Type: ParameterTypeString},
}

args := []string{"--verbose", "-o", "file.txt"}
result, args, err := GatherFlagsFromStringList(args, params, false, false, "")

if err != nil {
   log.Fatal(err)
}

fmt.Println(result) // prints: map[verbose:true output:file.txt]
```

In this example, the function parses the `--verbose` and `-o` flags according to the provided parameter definitions. The `--verbose` flag is a boolean flag and is set to "true". The `-o` flag is a string flag and its value is "file.txt".

## Examples

Here are some examples of command line inputs and the corresponding output of the function:

- Input: `--verbose -o file.txt`
    - Output:
        - Flags:
            - verbose: true
            - output: file.txt

- Input: `--debug --log-level info`
    - Output:
        - Flags:
            - debug: true
            - log-level: info

- Input: `--output=file.txt --verbose=true`
    - Output:
        - Flags:
            - output: file.txt
            - verbose: true

- Input: `--size 100 --color red`
    - Output:
        - Flags:
            - size: 100
            - color: red

- Input: `--int-list=1,2 --int-list 3,4 --int-list=5`
    - Output:
        - Flags:
            - int-list: [1, 2, 3, 4, 5]

- Input: `--string-list item1,item2,item3`
    - Output:
        - Flags:
            - string-list: [item1, item2, item3]

- Input: `--verbose foobar --output file.txt another argument`
    - Output:
        - Flags:
            - verbose: true
            - output: file.txt
        - Args:
            - foobar
            - another
            - argument
