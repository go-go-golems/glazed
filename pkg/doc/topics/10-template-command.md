---
Title: TemplateCommand
Slug: template-command
Short: |
  TemplateCommand is a feature that allows you to define a command that renders template text. 
  The template text is defined in the `template` field of the YAML file, and flags/params are 
  used to render the final template output.
Topics:
  - command
  - template
  - TemplateCommand
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

# TemplateCommand

TemplateCommand is a feature that allows you to define a command that renders template text.
The template text is defined in the `template` field of the YAML file,
and flags/params are used to render the final template output.

## Example

Here is an example of a TemplateCommand in YAML format:

  ```yaml
name: greeting
short: Renders a greeting template
flags:
  - name: name
    type: string
    help: Name to greet
  - name: hobbies
    type: stringList
    help: Hobbies
template: |
  Hello {{.name}}!

  Your hobbies are:
  {{ range .hobbies }}
    - {{ . }}
  {{ end }}
  ```
  
This command will render a greeting based on the name and hobbies provided as flags.

## Usage

Here is an example of how to use a TemplateCommand.

```go
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

func main() {
	// Load the TemplateCommand from the YAML file
	f, err := os.Open("/tmp/test.yaml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	tcl := &cmds.TemplateCommandLoader{}
	commands, err := tcl.LoadCommandFromYAML(f)
	if err != nil {
		fmt.Println("Error loading command from YAML:", err)
		return
	}
	if len(commands) != 1 {
		fmt.Println("Expected exactly one command in the YAML file")
		return
	}
	cmd, ok := commands[0].(*cmds.TemplateCommand)
	if !ok {
		fmt.Println("Command is not a TemplateCommand")
		return
	}

	// Define the flags
	flags := []string{"--flag1=value1", "--flag2=value2"}

	// Run the command
	buf := &strings.Builder{}
	parsedLayers := map[string]*layers.ParsedParameterLayer{}
	ps, args, err := parameters.GatherFlagsFromStringList(
		flags, cmd.Flags,
		false, false,
		"")
	if err != nil {
		fmt.Println("Error gathering flags:", err)
		return
	}
	arguments, err := parameters.GatherArguments(args, cmd.Arguments, false, false)
	if err != nil {
		fmt.Println("Error gathering arguments:", err)
		return
	}
	for p := arguments.Oldest(); p != nil; p = p.Next() {
		k, v := p.Key, p.Value
		ps[k] = v
	}
	err = cmd.RunIntoWriter(context.Background(), parsedLayers, ps, buf)
	if err != nil {
		fmt.Println("Error running command:", err)
		return
	}

	// Print the output
	fmt.Println(buf.String())
}
```
  