package cmds

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// JsonSchemaProperty represents a property in the JSON Schema
type JsonSchemaProperty struct {
	Type                 string                         `json:"type"`
	Description          string                         `json:"description,omitempty"`
	Enum                 []string                       `json:"enum,omitempty"`
	Default              interface{}                    `json:"default,omitempty"`
	Items                *JsonSchemaProperty            `json:"items,omitempty"`
	Required             bool                           `json:"-"`
	Properties           map[string]*JsonSchemaProperty `json:"properties,omitempty"`
	AdditionalProperties *JsonSchemaProperty            `json:"additionalProperties,omitempty"`
}

// CommandJsonSchema represents the root JSON Schema for a command
type CommandJsonSchema struct {
	Type        string                         `json:"type"`
	Description string                         `json:"description,omitempty"`
	Properties  map[string]*JsonSchemaProperty `json:"properties"`
	Required    []string                       `json:"required,omitempty"`
}

// parameterTypeToJsonSchema converts a parameter definition to a JSON schema property
func parameterTypeToJsonSchema(param *parameters.ParameterDefinition) (*JsonSchemaProperty, error) {
	prop := &JsonSchemaProperty{
		Description: param.Help,
		Required:    param.Required,
	}

	if param.Default != nil {
		prop.Default = *param.Default
	}

	switch param.Type {
	// Basic types
	case parameters.ParameterTypeString:
		prop.Type = "string"

	case parameters.ParameterTypeInteger:
		prop.Type = "integer"

	case parameters.ParameterTypeFloat:
		prop.Type = "number"

	case parameters.ParameterTypeBool:
		prop.Type = "boolean"

	case parameters.ParameterTypeDate:
		prop.Type = "string"
		// Add format for date strings
		prop.Properties = map[string]*JsonSchemaProperty{
			"format": {Type: "string", Default: "date"},
		}

	// List types
	case parameters.ParameterTypeStringList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeIntegerList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "integer"}

	case parameters.ParameterTypeFloatList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "number"}

	// Choice types
	case parameters.ParameterTypeChoice:
		prop.Type = "string"
		prop.Enum = param.Choices

	case parameters.ParameterTypeChoiceList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}
		prop.Items.Enum = param.Choices

	// File types
	case parameters.ParameterTypeFile:
		prop.Type = "object"
		prop.Properties = map[string]*JsonSchemaProperty{
			"path":    {Type: "string", Description: "Path to the file"},
			"content": {Type: "string", Description: "File content"},
		}

	case parameters.ParameterTypeFileList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type: "object",
			Properties: map[string]*JsonSchemaProperty{
				"path":    {Type: "string", Description: "Path to the file"},
				"content": {Type: "string", Description: "File content"},
			},
		}

	// Key-value type
	case parameters.ParameterTypeKeyValue:
		prop.Type = "object"
		prop.Properties = map[string]*JsonSchemaProperty{
			"key":   {Type: "string"},
			"value": {Type: "string"},
		}

	// File-based parameter types
	case parameters.ParameterTypeStringFromFile:
		prop.Type = "string"

	case parameters.ParameterTypeStringFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeObjectFromFile:
		prop.Type = "object"
		prop.AdditionalProperties = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeObjectListFromFile:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type:                 "object",
			AdditionalProperties: &JsonSchemaProperty{Type: "string"},
		}

	case parameters.ParameterTypeObjectListFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type:                 "object",
			AdditionalProperties: &JsonSchemaProperty{Type: "string"},
		}

	case parameters.ParameterTypeStringListFromFile:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeStringListFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", param.Type)
	}

	return prop, nil
}

// ToJsonSchema converts a ShellCommand to a JSON Schema representation
func (c *CommandDescription) ToJsonSchema() (*CommandJsonSchema, error) {
	schema := &CommandJsonSchema{
		Type:        "object",
		Description: fmt.Sprintf("%s\n\n%s", c.Short, c.Long),
		Properties:  make(map[string]*JsonSchemaProperty),
		Required:    []string{},
	}

	// Process flags
	err := c.GetDefaultFlags().ForEachE(func(flag *parameters.ParameterDefinition) error {
		prop, err := parameterTypeToJsonSchema(flag)
		if err != nil {
			return fmt.Errorf("error processing flag %s: %w", flag.Name, err)
		}
		schema.Properties[flag.Name] = prop
		if flag.Required {
			schema.Required = append(schema.Required, flag.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Process arguments
	err = c.GetDefaultArguments().ForEachE(func(arg *parameters.ParameterDefinition) error {
		prop, err := parameterTypeToJsonSchema(arg)
		if err != nil {
			return fmt.Errorf("error processing argument %s: %w", arg.Name, err)
		}
		schema.Properties[arg.Name] = prop
		if arg.Required {
			schema.Required = append(schema.Required, arg.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return schema, nil
}
