package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func main() {
	// Create the YAML content directly
	yamlContent := `name: greeting
short: Renders a greeting template
flags:
  - name: name
    type: string
    help: Name to greet
    default: "World"
  - name: hobbies
    type: stringList
    help: List of hobbies
template: |
  Hello {{.name}}!
  {{if .hobbies}}
  Your hobbies are:
  {{range .hobbies}}  - {{.}}
  {{end}}{{end}}`

	// Load the TemplateCommand from the YAML content
	reader := strings.NewReader(yamlContent)
	tcl := &cmds.TemplateCommandLoader{}
	commands, err := tcl.LoadCommandFromYAML(reader)
	if err != nil {
		fmt.Printf("Error loading command from YAML: %v\n", err)
		return
	}
	if len(commands) != 1 {
		fmt.Printf("Expected exactly one command in the YAML, got %d\n", len(commands))
		return
	}
	cmd, ok := commands[0].(*cmds.TemplateCommand)
	if !ok {
		fmt.Printf("Command is not a TemplateCommand, got %T\n", commands[0])
		return
	}

	// Test 1: Basic usage without hobbies
	fmt.Println("=== Test 1: Basic greeting ===")
	runTemplateCommand(cmd, map[string]interface{}{"name": "Alice"})

	// Test 2: With hobbies
	fmt.Println("\n=== Test 2: With hobbies ===")
	runTemplateCommand(cmd, map[string]interface{}{
		"name":    "Bob",
		"hobbies": []string{"reading", "coding", "gaming"},
	})

	// Test 3: Default name
	fmt.Println("\n=== Test 3: Default name ===")
	runTemplateCommand(cmd, map[string]interface{}{})
}

func runTemplateCommand(cmd *cmds.TemplateCommand, inputValues map[string]interface{}) {
	// Get the default layer
	defaultLayer, ok := cmd.Description().Schema.Get(schema.DefaultSlug)
	if !ok {
		fmt.Printf("Default layer not found\n")
		return
	}

	// Prepare options for creating parsed layer
	var options []values.SectionValuesOption
	for k, v := range inputValues {
		if _, ok := defaultLayer.GetDefinitions().Get(k); ok {
			options = append(options, values.WithFieldValue(k, v))
		}
	}

	// Create a parsed layer with the values
	parsedLayer, err := values.NewSectionValues(defaultLayer, options...)
	if err != nil {
		fmt.Printf("Error creating parsed layer: %v\n", err)
		return
	}

	// Create parsed layers container
	parsedLayers := values.New()
	parsedLayers.Set(schema.DefaultSlug, parsedLayer)

	// Run the command
	buf := &strings.Builder{}
	err = cmd.RunIntoWriter(context.Background(), parsedLayers, buf)
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return
	}

	// Print the output
	fmt.Printf("Output:\n%s", buf.String())
}
