---
Title: TemplateCommand
Slug: template-command
Short: |
  Create commands that render Go template text using command-line flags and arguments as template variables.
Topics:
  - command
  - template
  - TemplateCommand
IsTopLevel: false
ShowPerDefault: false
SectionType: GeneralTopic
---

# TemplateCommand

A TemplateCommand allows you to define commands that render Go template text using command-line parameters as template variables. This enables rapid prototyping of text generation tools without writing Go code—simply define parameters in YAML and write a template that uses those parameters.

## Creating Template Commands

Template commands are defined in YAML files with a `template` field containing Go template syntax. The template receives all parsed parameters as variables accessible through the standard `{{.variable}}` syntax.

**Example YAML definition:**

```yaml
name: greeting
short: Generate personalized greetings
flags:
  - name: name
    type: string
    help: Name to greet
    default: "World"
  - name: language
    type: choice
    help: Greeting language
    choices: [english, spanish, french]
    default: "english"
  - name: hobbies
    type: stringList
    help: List of hobbies to mention
template: |
  {{if eq .language "spanish"}}¡Hola{{else if eq .language "french"}}Bonjour{{else}}Hello{{end}} {{.name}}!
  {{if .hobbies}}
  Your hobbies are:
  {{range .hobbies}}  - {{.}}
  {{end}}{{end}}
```

**Use cases:**
- Quick text generation utilities
- Configuration file templates
- Report generators
- Code scaffolding tools

## Loading and Running Template Commands

Template commands implement the `WriterCommand` interface and can be loaded from YAML using the `TemplateCommandLoader`.

**Loading from YAML:**

```go
package main

import (
	"context"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
)

func main() {
	yamlContent := `name: greeting
short: A simple greeting command
flags:
  - name: name
    type: string
    default: "World"
template: "Hello, {{.name}}!"`

	// 1. Load the command from a YAML string
	loader := &cmds.TemplateCommandLoader{}
	commands, err := loader.LoadCommandFromYAML(strings.NewReader(yamlContent))
	if err != nil || len(commands) == 0 {
		panic(err)
	}

	cmd := commands[0]

	// 2. Run the command programmatically
	// The runner handles parsing and execution.
	// For more complex execution, see `glaze help programmatic-execution`.
	err = runner.RunCommand(
		context.Background(),
		cmd,
		// Provide parameter values for the "default" layer
		runner.WithValues(map[string]interface{}{
			"name": "Alice",
		}),
		// Direct output to stdout
		runner.WithWriter(os.Stdout),
	)
	if err != nil {
		panic(err)
	}

	// Expected output: Hello, Alice!
}
```

## Template Syntax and Variables

Template commands use Go's `text/template` package syntax. All parsed parameters are available as variables in the template context.

**Common template patterns:**

```yaml
template: |
  # Conditional content
  {{if .debug}}Debug mode enabled{{end}}
  
  # Range over lists
  {{range .items}}
  - {{.}}
  {{end}}
  
  # String comparison
  {{if eq .environment "production"}}
  Production configuration
  {{else}}
  Development configuration
  {{end}}
  
  # Using defaults for optional parameters
  Name: {{.name | default "unnamed"}}
```

## Parameter Types

Template commands can leverage the full range of Glazed parameter types, allowing for rich and validated inputs. This means you can create templates that accept everything from simple strings and booleans to lists, choices, and even content from files.

Template commands support all standard Glazed parameter types:

- `string`, `stringList` - Text values
- `int`, `intList`, `float`, `floatList` - Numeric values  
- `bool` - Boolean flags
- `choice`, `choiceList` - Constrained selections
- `stringFromFile`, `objectFromFile` - File-based inputs

For more details on parameter types, see:
```
glaze help parameter-types
```

## Integration with Command Loaders

While `TemplateCommand` can be loaded manually, its real power is unlocked when used with the command loader system. This allows you to build applications that discover and load commands from a directory of YAML files, enabling you to add new functionality without recompiling your application.

Template commands can be loaded dynamically using the command loader system. This enables building applications that discover and load template commands from directories.

For information about setting up command loaders:
```
glaze help command-loaders
```
