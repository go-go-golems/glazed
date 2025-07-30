package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type ParameterTypesSettings struct {
	// Basic types
	StringParam  string    `glazed.parameter:"string-param"`
	SecretParam  string    `glazed.parameter:"secret-param"`
	IntegerParam int       `glazed.parameter:"integer-param"`
	FloatParam   float64   `glazed.parameter:"float-param"`
	BoolParam    bool      `glazed.parameter:"bool-param"`
	DateParam    time.Time `glazed.parameter:"date-param"`
	ChoiceParam  string    `glazed.parameter:"choice-param"`

	// List types
	StringListParam  []string  `glazed.parameter:"string-list-param"`
	IntegerListParam []int     `glazed.parameter:"integer-list-param"`
	FloatListParam   []float64 `glazed.parameter:"float-list-param"`
	ChoiceListParam  []string  `glazed.parameter:"choice-list-param"`

	// File types
	FileParam                *parameters.FileData     `glazed.parameter:"file-param"`
	FileListParam            []*parameters.FileData   `glazed.parameter:"file-list-param"`
	StringFromFileParam      string                   `glazed.parameter:"string-from-file-param"`
	StringFromFilesParam     string                   `glazed.parameter:"string-from-files-param"`
	StringListFromFileParam  []string                 `glazed.parameter:"string-list-from-file-param"`
	StringListFromFilesParam []string                 `glazed.parameter:"string-list-from-files-param"`
	ObjectFromFileParam      map[string]interface{}   `glazed.parameter:"object-from-file-param"`
	ObjectListFromFileParam  []map[string]interface{} `glazed.parameter:"object-list-from-file-param"`
	ObjectListFromFilesParam []map[string]interface{} `glazed.parameter:"object-list-from-files-param"`

	// Key-value type
	KeyValueParam map[string]string `glazed.parameter:"key-value-param"`
}

type ParameterTypesCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*ParameterTypesCommand)(nil)

func NewParameterTypesCommand() (*ParameterTypesCommand, error) {
	glazedParameterLayer, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed parameter layer")
	}

	return &ParameterTypesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"parameter-types",
			cmds.WithShort("Showcase all parameter types available in glazed"),
			cmds.WithLong(`This command demonstrates all the different parameter types available in the glazed framework.
It shows how to define and use each type, and displays the parsed values.

Parameter types demonstrated:
- Basic types: string, secret, integer, float, bool, date, choice
- List types: string-list, integer-list, float-list, choice-list  
- File types: file, file-list, string-from-file, object-from-file, etc.
- Key-value type: key-value mappings

Use --help to see all available parameters and their descriptions.`),
			cmds.WithFlags(
				// Basic types
				parameters.NewParameterDefinition(
					"string-param",
					parameters.ParameterTypeString,
					parameters.WithHelp("A simple string parameter"),
					parameters.WithDefault("default-string"),
				),
				parameters.NewParameterDefinition(
					"secret-param",
					parameters.ParameterTypeSecret,
					parameters.WithHelp("A secret parameter (will be masked when displayed)"),
					parameters.WithDefault("secret-value"),
				),
				parameters.NewParameterDefinition(
					"integer-param",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("An integer parameter"),
					parameters.WithDefault(42),
				),
				parameters.NewParameterDefinition(
					"float-param",
					parameters.ParameterTypeFloat,
					parameters.WithHelp("A floating point parameter"),
					parameters.WithDefault(3.14),
				),
				parameters.NewParameterDefinition(
					"bool-param",
					parameters.ParameterTypeBool,
					parameters.WithHelp("A boolean parameter"),
					parameters.WithDefault(true),
				),
				parameters.NewParameterDefinition(
					"date-param",
					parameters.ParameterTypeDate,
					parameters.WithHelp("A date parameter (RFC3339 format or natural language)"),
					parameters.WithDefault("2024-01-01T00:00:00Z"),
				),
				parameters.NewParameterDefinition(
					"choice-param",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("A choice parameter with predefined options"),
					parameters.WithChoices("option1", "option2", "option3"),
					parameters.WithDefault("option1"),
				),

				// List types
				parameters.NewParameterDefinition(
					"string-list-param",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("A list of strings"),
					parameters.WithDefault([]string{"item1", "item2"}),
				),
				parameters.NewParameterDefinition(
					"integer-list-param",
					parameters.ParameterTypeIntegerList,
					parameters.WithHelp("A list of integers"),
					parameters.WithDefault([]int{1, 2, 3}),
				),
				parameters.NewParameterDefinition(
					"float-list-param",
					parameters.ParameterTypeFloatList,
					parameters.WithHelp("A list of floating point numbers"),
					parameters.WithDefault([]float64{1.1, 2.2, 3.3}),
				),
				parameters.NewParameterDefinition(
					"choice-list-param",
					parameters.ParameterTypeChoiceList,
					parameters.WithHelp("A list of choices from predefined options"),
					parameters.WithChoices("red", "green", "blue"),
					parameters.WithDefault([]string{"red", "blue"}),
				),

				// File types
				parameters.NewParameterDefinition(
					"file-param",
					parameters.ParameterTypeFile,
					parameters.WithHelp("A file parameter that loads file metadata"),
				),
				parameters.NewParameterDefinition(
					"file-list-param",
					parameters.ParameterTypeFileList,
					parameters.WithHelp("A list of files with metadata"),
				),
				parameters.NewParameterDefinition(
					"string-from-file-param",
					parameters.ParameterTypeStringFromFile,
					parameters.WithHelp("Load string content from a file"),
				),
				parameters.NewParameterDefinition(
					"string-from-files-param",
					parameters.ParameterTypeStringFromFiles,
					parameters.WithHelp("Load and concatenate string content from multiple files"),
				),
				parameters.NewParameterDefinition(
					"string-list-from-file-param",
					parameters.ParameterTypeStringListFromFile,
					parameters.WithHelp("Load lines from a file as a string list"),
				),
				parameters.NewParameterDefinition(
					"string-list-from-files-param",
					parameters.ParameterTypeStringListFromFiles,
					parameters.WithHelp("Load lines from multiple files as a string list"),
				),
				parameters.NewParameterDefinition(
					"object-from-file-param",
					parameters.ParameterTypeObjectFromFile,
					parameters.WithHelp("Load a JSON/YAML object from a file"),
				),
				parameters.NewParameterDefinition(
					"object-list-from-file-param",
					parameters.ParameterTypeObjectListFromFile,
					parameters.WithHelp("Load a list of objects from a file"),
				),
				parameters.NewParameterDefinition(
					"object-list-from-files-param",
					parameters.ParameterTypeObjectListFromFiles,
					parameters.WithHelp("Load and merge object lists from multiple files"),
				),

				// Key-value type
				parameters.NewParameterDefinition(
					"key-value-param",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("Key-value pairs (format: key:value or @filename for JSON/YAML file)"),
					parameters.WithDefault(map[string]string{"default-key": "default-value"}),
				),
			),
			cmds.WithLayersList(
				glazedParameterLayer,
			),
		),
	}, nil
}

func (c *ParameterTypesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	gp middlewares.Processor,
) error {
	s := &ParameterTypesSettings{}
	err := parsedLayers.InitializeStruct(layers.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize settings from parameters")
	}

	// We'll use hardcoded metadata since layer access is complex

	// Create a result row for each parameter
	parameterData := []struct {
		name         string
		paramType    parameters.ParameterType
		realValue    interface{}
		help         string
		required     bool
		choices      []string
		defaultValue interface{}
	}{
		{"string-param", parameters.ParameterTypeString, s.StringParam, "A simple string parameter", false, nil, "default-string"},
		{"secret-param", parameters.ParameterTypeSecret, s.SecretParam, "A secret parameter (will be masked when displayed)", false, nil, "secret-value"},
		{"integer-param", parameters.ParameterTypeInteger, s.IntegerParam, "An integer parameter", false, nil, 42},
		{"float-param", parameters.ParameterTypeFloat, s.FloatParam, "A floating point parameter", false, nil, 3.14},
		{"bool-param", parameters.ParameterTypeBool, s.BoolParam, "A boolean parameter", false, nil, true},
		{"date-param", parameters.ParameterTypeDate, s.DateParam, "A date parameter (RFC3339 format or natural language)", false, nil, "2024-01-01T00:00:00Z"},
		{"choice-param", parameters.ParameterTypeChoice, s.ChoiceParam, "A choice parameter with predefined options", false, []string{"option1", "option2", "option3"}, "option1"},
		{"string-list-param", parameters.ParameterTypeStringList, s.StringListParam, "A list of strings", false, nil, []string{"item1", "item2"}},
		{"integer-list-param", parameters.ParameterTypeIntegerList, s.IntegerListParam, "A list of integers", false, nil, []int{1, 2, 3}},
		{"float-list-param", parameters.ParameterTypeFloatList, s.FloatListParam, "A list of floating point numbers", false, nil, []float64{1.1, 2.2, 3.3}},
		{"choice-list-param", parameters.ParameterTypeChoiceList, s.ChoiceListParam, "A list of choices from predefined options", false, []string{"red", "green", "blue"}, []string{"red", "blue"}},
		{"file-param", parameters.ParameterTypeFile, s.FileParam, "A file parameter that loads file metadata", false, nil, nil},
		{"file-list-param", parameters.ParameterTypeFileList, s.FileListParam, "A list of files with metadata", false, nil, nil},
		{"string-from-file-param", parameters.ParameterTypeStringFromFile, s.StringFromFileParam, "Load string content from a file", false, nil, nil},
		{"string-from-files-param", parameters.ParameterTypeStringFromFiles, s.StringFromFilesParam, "Load and concatenate string content from multiple files", false, nil, nil},
		{"string-list-from-file-param", parameters.ParameterTypeStringListFromFile, s.StringListFromFileParam, "Load lines from a file as a string list", false, nil, nil},
		{"string-list-from-files-param", parameters.ParameterTypeStringListFromFiles, s.StringListFromFilesParam, "Load lines from multiple files as a string list", false, nil, nil},
		{"object-from-file-param", parameters.ParameterTypeObjectFromFile, s.ObjectFromFileParam, "Load a JSON/YAML object from a file", false, nil, nil},
		{"object-list-from-file-param", parameters.ParameterTypeObjectListFromFile, s.ObjectListFromFileParam, "Load a list of objects from a file", false, nil, nil},
		{"object-list-from-files-param", parameters.ParameterTypeObjectListFromFiles, s.ObjectListFromFilesParam, "Load and merge object lists from multiple files", false, nil, nil},
		{"key-value-param", parameters.ParameterTypeKeyValue, s.KeyValueParam, "Key-value pairs (format: key:value or @filename for JSON/YAML file)", false, nil, map[string]string{"default-key": "default-value"}},
	}

	for _, param := range parameterData {
		// Get rendered value (what would be displayed to user)
		renderedValue := "<nil>"
		if param.realValue != nil {
			// Check for nil pointers using reflection
			v := reflect.ValueOf(param.realValue)
			isNil := false
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Map {
				isNil = v.IsNil()
			}

			if !isNil {
				var err error
				renderedValue, err = parameters.RenderValue(param.paramType, param.realValue)
				if err != nil {
					renderedValue = fmt.Sprintf("ERROR: %v", err)
				}
			}
		}

		// Format real value for display
		realValueStr := fmt.Sprintf("%v", param.realValue)
		if param.realValue == nil {
			realValueStr = "<nil>"
		}

		// Format default value
		defaultValueStr := fmt.Sprintf("%v", param.defaultValue)
		if param.defaultValue == nil {
			defaultValueStr = "<nil>"
		}

		// Format choices
		choicesStr := ""
		if len(param.choices) > 0 {
			choicesStr = fmt.Sprintf("[%s]", strings.Join(param.choices, ", "))
		}

		result := types.NewRow(
			types.MRP("parameter_name", param.name),
			types.MRP("parameter_type", string(param.paramType)),
			types.MRP("real_value", realValueStr),
			types.MRP("rendered_value", renderedValue),
			types.MRP("default_value", defaultValueStr),
			types.MRP("required", param.required),
			types.MRP("choices", choicesStr),
			types.MRP("help", param.help),
		)

		err = gp.AddRow(ctx, result)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	helpSystem := help.NewHelpSystem()

	cmd, err := NewParameterTypesCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create command: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "parameter-types",
		Short: "Showcase all glazed parameter types",
	}

	cobraCommand, err := cli.BuildCobraCommand(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build cobra command: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(cobraCommand)
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
