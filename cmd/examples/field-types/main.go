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
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type FieldTypesSettings struct {
	// Basic types
	StringField  string    `glazed:"string-field"`
	SecretField  string    `glazed:"secret-field"`
	IntegerField int       `glazed:"integer-field"`
	FloatField   float64   `glazed:"float-field"`
	BoolField    bool      `glazed:"bool-field"`
	DateField    time.Time `glazed:"date-field"`
	ChoiceField  string    `glazed:"choice-field"`

	// List types
	StringListField  []string  `glazed:"string-list-field"`
	IntegerListField []int     `glazed:"integer-list-field"`
	FloatListField   []float64 `glazed:"float-list-field"`
	ChoiceListField  []string  `glazed:"choice-list-field"`

	// File types
	FileField                *fields.FileData         `glazed:"file-field"`
	FileListField            []*fields.FileData       `glazed:"file-list-field"`
	StringFromFileField      string                   `glazed:"string-from-file-field"`
	StringFromFilesField     string                   `glazed:"string-from-files-field"`
	StringListFromFileField  []string                 `glazed:"string-list-from-file-field"`
	StringListFromFilesField []string                 `glazed:"string-list-from-files-field"`
	ObjectFromFileField      map[string]interface{}   `glazed:"object-from-file-field"`
	ObjectListFromFileField  []map[string]interface{} `glazed:"object-list-from-file-field"`
	ObjectListFromFilesField []map[string]interface{} `glazed:"object-list-from-files-field"`

	// Key-value type
	KeyValueField map[string]string `glazed:"key-value-field"`
}

type FieldTypesCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FieldTypesCommand)(nil)

func NewFieldTypesCommand() (*FieldTypesCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, errors.Wrap(err, "could not create Glazed section")
	}

	return &FieldTypesCommand{
		CommandDescription: cmds.NewCommandDescription(
			"field-types",
			cmds.WithShort("Showcase all field types available in glazed"),
			cmds.WithLong(`This command demonstrates all the different field types available in the glazed framework.
It shows how to define and use each type, and displays the parsed values.

Field types demonstrated:
- Basic types: string, secret, integer, float, bool, date, choice
- List types: string-list, integer-list, float-list, choice-list  
- File types: file, file-list, string-from-file, object-from-file, etc.
- Key-value type: key-value mappings

Use --help to see all available fields and their descriptions.`),
			cmds.WithFlags(
				// Basic types
				fields.New(
					"string-field",
					fields.TypeString,
					fields.WithHelp("A simple string field"),
					fields.WithDefault("default-string"),
				),
				fields.New(
					"secret-field",
					fields.TypeSecret,
					fields.WithHelp("A secret field (will be masked when displayed)"),
					fields.WithDefault("secret-value"),
				),
				fields.New(
					"integer-field",
					fields.TypeInteger,
					fields.WithHelp("An integer field"),
					fields.WithDefault(42),
				),
				fields.New(
					"float-field",
					fields.TypeFloat,
					fields.WithHelp("A floating point field"),
					fields.WithDefault(3.14),
				),
				fields.New(
					"bool-field",
					fields.TypeBool,
					fields.WithHelp("A boolean field"),
					fields.WithDefault(true),
				),
				fields.New(
					"date-field",
					fields.TypeDate,
					fields.WithHelp("A date field (RFC3339 format or natural language)"),
					fields.WithDefault("2024-01-01T00:00:00Z"),
				),
				fields.New(
					"choice-field",
					fields.TypeChoice,
					fields.WithHelp("A choice field with predefined options"),
					fields.WithChoices("option1", "option2", "option3"),
					fields.WithDefault("option1"),
				),

				// List types
				fields.New(
					"string-list-field",
					fields.TypeStringList,
					fields.WithHelp("A list of strings"),
					fields.WithDefault([]string{"item1", "item2"}),
				),
				fields.New(
					"integer-list-field",
					fields.TypeIntegerList,
					fields.WithHelp("A list of integers"),
					fields.WithDefault([]int{1, 2, 3}),
				),
				fields.New(
					"float-list-field",
					fields.TypeFloatList,
					fields.WithHelp("A list of floating point numbers"),
					fields.WithDefault([]float64{1.1, 2.2, 3.3}),
				),
				fields.New(
					"choice-list-field",
					fields.TypeChoiceList,
					fields.WithHelp("A list of choices from predefined options"),
					fields.WithChoices("red", "green", "blue"),
					fields.WithDefault([]string{"red", "blue"}),
				),

				// File types
				fields.New(
					"file-field",
					fields.TypeFile,
					fields.WithHelp("A file field that loads file metadata"),
				),
				fields.New(
					"file-list-field",
					fields.TypeFileList,
					fields.WithHelp("A list of files with metadata"),
				),
				fields.New(
					"string-from-file-field",
					fields.TypeStringFromFile,
					fields.WithHelp("Load string content from a file"),
				),
				fields.New(
					"string-from-files-field",
					fields.TypeStringFromFiles,
					fields.WithHelp("Load and concatenate string content from multiple files"),
				),
				fields.New(
					"string-list-from-file-field",
					fields.TypeStringListFromFile,
					fields.WithHelp("Load lines from a file as a string list"),
				),
				fields.New(
					"string-list-from-files-field",
					fields.TypeStringListFromFiles,
					fields.WithHelp("Load lines from multiple files as a string list"),
				),
				fields.New(
					"object-from-file-field",
					fields.TypeObjectFromFile,
					fields.WithHelp("Load a JSON/YAML object from a file"),
				),
				fields.New(
					"object-list-from-file-field",
					fields.TypeObjectListFromFile,
					fields.WithHelp("Load a list of objects from a file"),
				),
				fields.New(
					"object-list-from-files-field",
					fields.TypeObjectListFromFiles,
					fields.WithHelp("Load and merge object lists from multiple files"),
				),

				// Key-value type
				fields.New(
					"key-value-field",
					fields.TypeKeyValue,
					fields.WithHelp("Key-value pairs (format: key:value or @filename for JSON/YAML file)"),
					fields.WithDefault(map[string]string{"default-key": "default-value"}),
				),
			),
			cmds.WithSections(
				glazedSection,
			),
		),
	}, nil
}

func (c *FieldTypesCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &FieldTypesSettings{}
	err := vals.DecodeSectionInto(schema.DefaultSlug, s)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize settings from fields")
	}

	// We'll use hardcoded metadata since section access is complex

	// Create a result row for each field
	fieldData := []struct {
		name         string
		fieldType    fields.Type
		realValue    interface{}
		help         string
		required     bool
		choices      []string
		defaultValue interface{}
	}{
		{"string-field", fields.TypeString, s.StringField, "A simple string field", false, nil, "default-string"},
		{"secret-field", fields.TypeSecret, s.SecretField, "A secret field (will be masked when displayed)", false, nil, "secret-value"},
		{"integer-field", fields.TypeInteger, s.IntegerField, "An integer field", false, nil, 42},
		{"float-field", fields.TypeFloat, s.FloatField, "A floating point field", false, nil, 3.14},
		{"bool-field", fields.TypeBool, s.BoolField, "A boolean field", false, nil, true},
		{"date-field", fields.TypeDate, s.DateField, "A date field (RFC3339 format or natural language)", false, nil, "2024-01-01T00:00:00Z"},
		{"choice-field", fields.TypeChoice, s.ChoiceField, "A choice field with predefined options", false, []string{"option1", "option2", "option3"}, "option1"},
		{"string-list-field", fields.TypeStringList, s.StringListField, "A list of strings", false, nil, []string{"item1", "item2"}},
		{"integer-list-field", fields.TypeIntegerList, s.IntegerListField, "A list of integers", false, nil, []int{1, 2, 3}},
		{"float-list-field", fields.TypeFloatList, s.FloatListField, "A list of floating point numbers", false, nil, []float64{1.1, 2.2, 3.3}},
		{"choice-list-field", fields.TypeChoiceList, s.ChoiceListField, "A list of choices from predefined options", false, []string{"red", "green", "blue"}, []string{"red", "blue"}},
		{"file-field", fields.TypeFile, s.FileField, "A file field that loads file metadata", false, nil, nil},
		{"file-list-field", fields.TypeFileList, s.FileListField, "A list of files with metadata", false, nil, nil},
		{"string-from-file-field", fields.TypeStringFromFile, s.StringFromFileField, "Load string content from a file", false, nil, nil},
		{"string-from-files-field", fields.TypeStringFromFiles, s.StringFromFilesField, "Load and concatenate string content from multiple files", false, nil, nil},
		{"string-list-from-file-field", fields.TypeStringListFromFile, s.StringListFromFileField, "Load lines from a file as a string list", false, nil, nil},
		{"string-list-from-files-field", fields.TypeStringListFromFiles, s.StringListFromFilesField, "Load lines from multiple files as a string list", false, nil, nil},
		{"object-from-file-field", fields.TypeObjectFromFile, s.ObjectFromFileField, "Load a JSON/YAML object from a file", false, nil, nil},
		{"object-list-from-file-field", fields.TypeObjectListFromFile, s.ObjectListFromFileField, "Load a list of objects from a file", false, nil, nil},
		{"object-list-from-files-field", fields.TypeObjectListFromFiles, s.ObjectListFromFilesField, "Load and merge object lists from multiple files", false, nil, nil},
		{"key-value-field", fields.TypeKeyValue, s.KeyValueField, "Key-value pairs (format: key:value or @filename for JSON/YAML file)", false, nil, map[string]string{"default-key": "default-value"}},
	}

	for _, field := range fieldData {
		// Get rendered value (what would be displayed to user)
		renderedValue := "<nil>"
		if field.realValue != nil {
			// Check for nil pointers using reflection
			v := reflect.ValueOf(field.realValue)
			isNil := false
			if v.Kind() == reflect.Ptr || v.Kind() == reflect.Slice || v.Kind() == reflect.Map {
				isNil = v.IsNil()
			}

			if !isNil {
				var err error
				renderedValue, err = fields.RenderValue(field.fieldType, field.realValue)
				if err != nil {
					renderedValue = fmt.Sprintf("ERROR: %v", err)
				}
			}
		}

		// Format real value for display
		realValueStr := fmt.Sprintf("%v", field.realValue)
		if field.realValue == nil {
			realValueStr = "<nil>"
		}

		// Format default value
		defaultValueStr := fmt.Sprintf("%v", field.defaultValue)
		if field.defaultValue == nil {
			defaultValueStr = "<nil>"
		}

		// Format choices
		choicesStr := ""
		if len(field.choices) > 0 {
			choicesStr = fmt.Sprintf("[%s]", strings.Join(field.choices, ", "))
		}

		result := types.NewRow(
			types.MRP("field_name", field.name),
			types.MRP("field_type", string(field.fieldType)),
			types.MRP("real_value", realValueStr),
			types.MRP("rendered_value", renderedValue),
			types.MRP("default_value", defaultValueStr),
			types.MRP("required", field.required),
			types.MRP("choices", choicesStr),
			types.MRP("help", field.help),
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

	cmd, err := NewFieldTypesCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create command: %v\n", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "field-types",
		Short: "Showcase all glazed field types",
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
