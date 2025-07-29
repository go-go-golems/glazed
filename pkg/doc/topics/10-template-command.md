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
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

func main() {
	yamlContent := `name: greeting
short: Generate personalized greetings
flags:
  - name: name
    type: string
    default: "World"
template: "Hello {{.name}}!"`

	// Load command from YAML
	reader := strings.NewReader(yamlContent)
	loader := &cmds.TemplateCommandLoader{}
	commands, err := loader.LoadCommandFromYAML(reader)
	if err != nil {
		panic(err)
	}
	
	cmd := commands[0].(*cmds.TemplateCommand)
	
	// Execute with parameters
	runTemplateCommand(cmd, map[string]interface{}{
		"name": "Alice",
	})
}

func runTemplateCommand(cmd *cmds.TemplateCommand, values map[string]interface{}) {
	// Get default parameter layer
	defaultLayer, ok := cmd.Description().Layers.Get(layers.DefaultSlug)
	if !ok {
		panic("default layer not found")
	}
	
	// Create parsed layer with parameter values
	var options []layers.ParsedLayerOption
	for k, v := range values {
		if _, exists := defaultLayer.GetParameterDefinitions().Get(k); exists {
			options = append(options, layers.WithParsedParameterValue(k, v))
		}
	}
	
	parsedLayer, err := layers.NewParsedLayer(defaultLayer, options...)
	if err != nil {
		panic(err)
	}
	
	// Execute template
	parsedLayers := layers.NewParsedLayers()
	parsedLayers.Set(layers.DefaultSlug, parsedLayer)
	
	var output strings.Builder
	err = cmd.RunIntoWriter(context.Background(), parsedLayers, &output)
	if err != nil {
		panic(err)
	}
	
	// Output: Hello Alice!
	fmt.Print(output.String())
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

Template commands can be loaded dynamically using the command loader system. This enables building applications that discover and load template commands from directories.

For information about setting up command loaders:
```
glaze help command-loaders
```
