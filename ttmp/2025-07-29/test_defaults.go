package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

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

	fmt.Print(output.String())
}

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

	// Test 1: With explicit value
	fmt.Print("With value: ")
	runTemplateCommand(cmd, map[string]interface{}{
		"name": "Alice",
	})
	fmt.Println()

	// Test 2: Without value (should use default)
	fmt.Print("Without value: ")
	runTemplateCommand(cmd, map[string]interface{}{})
	fmt.Println()
}
