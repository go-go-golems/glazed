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
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
	glazedSection, err := schema.NewGlazedSchema()
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
				fields.New(
					"string-param",
					fields.TypeString,
					fields.WithHelp("A simple string parameter"),
					fields.WithDefault("default-string"),
				),
				fields.New(
					"secret-param",
					fields.TypeSecret,
					fields.WithHelp("A secret parameter (will be masked when displayed)"),
					fields.WithDefault("secret-value"),
				),
				fields.New(
					"integer-param",
					fields.TypeInteger,
					fields.WithHelp("An integer parameter"),
					fields.WithDefault(42),
				),
				fields.New(
					"float-param",
					fields.TypeFloat,
					fields.WithHelp("A floating point parameter"),
					fields.WithDefault(3.14),
				),
				fields.New(
					"bool-param",
					fields.TypeBool,
					fields.WithHelp("A boolean parameter"),
					fields.WithDefault(true),
				),
				fields.New(
					"date-param",
					fields.TypeDate,
					fields.WithHelp("A date parameter (RFC3339 format or natural language)"),
					fields.WithDefault("2024-01-01T00:00:00Z"),
				),
				fields.New(
					"choice-param",
					fields.TypeChoice,
					fields.WithHelp("A choice parameter with predefined options"),
					fields.WithChoices("option1", "option2", "option3"),
					fields.WithDefault("option1"),
				),

				// List types
				fields.New(
					"string-list-param",
					fields.TypeStringList,
					fields.WithHelp("A list of strings"),
					fields.WithDefault([]string{"item1", "item2"}),
				),
				fields.New(
					"integer-list-param",
					fields.TypeIntegerList,
					fields.WithHelp("A list of integers"),
					fields.WithDefault([]int{1, 2, 3}),
				),
				fields.New(
					"float-list-param",
					fields.TypeFloatList,
					fields.WithHelp("A list of floating point numbers"),
					fields.WithDefault([]float64{1.1, 2.2, 3.3}),
				),
				fields.New(
					"choice-list-param",
					fields.TypeChoiceList,
					fields.WithHelp("A list of choices from predefined options"),
					fields.WithChoices("red", "green", "blue"),
					fields.WithDefault([]string{"red", "blue"}),
				),

				// File types
				fields.New(
					"file-param",
					fields.TypeFile,
					fields.WithHelp("A file parameter that loads file metadata"),
				),
				fields.New(
					"file-list-param",
					fields.TypeFileList,
					fields.WithHelp("A list of files with metadata"),
				),
				fields.New(
					"string-from-file-param",
					fields.TypeStringFromFile,
					fields.WithHelp("Load string content from a file"),
				),
				fields.New(
					"string-from-files-param",
					fields.TypeStringFromFiles,
					fields.WithHelp("Load and concatenate string content from multiple files"),
				),
				fields.New(
					"string-list-from-file-param",
					fields.TypeStringListFromFile,
					fields.WithHelp("Load lines from a file as a string list"),
				),
				fields.New(
					"string-list-from-files-param",
					fields.TypeStringListFromFiles,
					fields.WithHelp("Load lines from multiple files as a string list"),
				),
				fields.New(
					"object-from-file-param",
					fields.TypeObjectFromFile,
					fields.WithHelp("Load a JSON/YAML object from a file"),
				),
				fields.New(
					"object-list-from-file-param",
					fields.TypeObjectListFromFile,
					fields.WithHelp("Load a list of objects from a file"),
				),
				fields.New(
					"object-list-from-files-param",
					fields.TypeObjectListFromFiles,
					fields.WithHelp("Load and merge object lists from multiple files"),
				),

				// Key-value type
				fields.New(
					"key-value-param",
					fields.TypeKeyValue,
					fields.WithHelp("Key-value pairs (format: key:value or @filename for JSON/YAML file)"),
					fields.WithDefault(map[string]string{"default-key": "default-value"}),
				),
			),
			cmds.WithLayersList(
				glazedSection,
			),
		),
	}, nil
}

func (c *ParameterTypesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &ParameterTypesSettings{}
	err := values.DecodeSectionInto(vals, schema.DefaultSlug, s)
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
