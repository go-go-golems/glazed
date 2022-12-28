package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/pkg/formatters"
	"github.com/wesen/glazed/pkg/middlewares"
	"github.com/wesen/glazed/pkg/types"
	"regexp"
	"strings"
)

// Helpers for cobra commands

func AddOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format (table, csv, tsv, json, yaml, sqlite)")
	cmd.Flags().StringP("output-file", "f", "", "Output file")

	cmd.Flags().String("table-format", "ascii", "Table format (ascii, markdown, html, csv, tsv)")
	cmd.Flags().Bool("with-headers", true, "Include headers in output (CSV, TSV)")
	cmd.Flags().String("csv-separator", ",", "CSV separator")

	// json output flags
	cmd.Flags().Bool("output-as-objects", false, "Output as individual objects instead of JSON array")

	// output processing
	cmd.Flags().Bool("flatten", false, "Flatten nested fields (after templating)")
}

func ParseOutputFlags(cmd *cobra.Command) (*OutputFormatterSettings, error) {
	output := cmd.Flag("output").Value.String()
	// TODO(manuel, 2022-11-21) Add support for output file / directory
	_ = cmd.Flag("output-file").Value.String()
	tableFormat := cmd.Flag("table-format").Value.String()
	flattenInput, _ := cmd.Flags().GetBool("flatten")
	outputAsObjects, _ := cmd.Flags().GetBool("output-as-objects")
	withHeaders, _ := cmd.Flags().GetBool("with-headers")
	csvSeparator, _ := cmd.Flags().GetString("csv-separator")

	return &OutputFormatterSettings{
		Output:          output,
		TableFormat:     tableFormat,
		WithHeaders:     withHeaders,
		OutputAsObjects: outputAsObjects,
		FlattenObjects:  flattenInput,
		CsvSeparator:    csvSeparator,
	}, nil
}

func AddSelectFlags(cmd *cobra.Command) {
	cmd.Flags().String("select", "", "Select a single field and output as a single line")
	cmd.Flags().String("select-template", "", "Output a single templated value for each row, on a single line")
}

func ParseSelectFlags(cmd *cobra.Command) (*SelectSettings, error) {
	selectField, _ := cmd.Flags().GetString("select")
	selectTemplate, _ := cmd.Flags().GetString("select-template")

	return &SelectSettings{
		SelectField:    selectField,
		SelectTemplate: selectTemplate,
	}, nil
}

func AddRenameFlags(cmd *cobra.Command) {
	cmd.Flags().StringSlice("rename", []string{}, "Rename fields (list of oldName:newName)")
	cmd.Flags().StringSlice("rename-regexp", []string{}, "Rename fields using regular expressions (list of regex:newName)")
	cmd.Flags().String("rename-yaml", "", "Rename fields using a yaml file")
}

func ParseRenameFlags(cmd *cobra.Command) (*RenameSettings, error) {
	renameFields, _ := cmd.Flags().GetStringSlice("rename")
	renameRegexpFields, _ := cmd.Flags().GetStringSlice("rename-regexp")
	renameYaml, _ := cmd.Flags().GetString("rename-yaml")

	renamesFieldsMap := map[types.FieldName]types.FieldName{}
	for _, renameField := range renameFields {
		parts := strings.Split(renameField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename field: %s", renameField)
		}
		renamesFieldsMap[types.FieldName(parts[0])] = types.FieldName(parts[1])
	}

	regexpReplacements := middlewares.RegexpReplacements{}
	for _, renameRegexpField := range renameRegexpFields {
		parts := strings.Split(renameRegexpField, ":")
		if len(parts) != 2 {
			return nil, errors.Errorf("Invalid rename-regexp field: %s", renameRegexpField)
		}
		re, err := regexp.Compile(parts[0])
		if err != nil {
			return nil, errors.Wrapf(err, "Invalid regexp: %s", parts[0])
		}
		regexpReplacements = append(regexpReplacements,
			&middlewares.RegexpReplacement{Regexp: re, Replacement: parts[1]})
	}

	return &RenameSettings{
		RenameFields:  renamesFieldsMap,
		RenameRegexps: regexpReplacements,
		YamlFile:      renameYaml,
	}, nil
}

func AddTemplateFlags(cmd *cobra.Command) {
	cmd.Flags().String("template", "", "Go Template to use for single string")
	cmd.Flags().StringSlice("template-field", nil, "For table output, fieldName:template to create new fields, or @fileName to read field templates from a yaml dictionary")
	cmd.Flags().Bool("use-row-templates", false, "Use row templates instead of column templates")
}

func ParseTemplateFlags(cmd *cobra.Command) (*TemplateSettings, error) {
	// templates get applied before flattening
	var templates map[types.FieldName]string
	var err error

	templateArgument, _ := cmd.Flags().GetString("template")
	if templateArgument != "" {
		templates = map[types.FieldName]string{}
		templates["_0"] = templateArgument
	} else {
		templateFields, _ := cmd.Flags().GetStringSlice("template-field")
		templates, err = ParseTemplateFieldArguments(templateFields)
		if err != nil {
			return nil, err
		}
	}

	useRowTemplates, _ := cmd.Flags().GetBool("use-row-templates")

	return &TemplateSettings{
		Templates:       templates,
		UseRowTemplates: useRowTemplates,
		RenameSeparator: "_",
	}, nil
}

// TODO(manuel, 2022-11-20) Make it easy for the developer to configure which flag they want
// and which they don't

// AddFieldsFilterFlags adds the flags for the following middlewares to the cmd:
// - FieldsFilterMiddleware
// - SortColumnsMiddleware
// - ReorderColumnOrderMiddleware
func AddFieldsFilterFlags(cmd *cobra.Command, defaultFields string) {
	defaultFieldHelp := defaultFields
	if defaultFieldHelp == "" {
		defaultFieldHelp = "all"
	}
	cmd.Flags().String("fields", defaultFields, "Fields to include in the output, default: "+defaultFieldHelp)
	cmd.Flags().String("filter", "", "Fields to remove from output")
	cmd.Flags().Bool("sort-columns", true, "Sort columns alphabetically")
}

func ParseFieldsFilterFlags(cmd *cobra.Command) (*FieldsFilterSettings, error) {
	fieldStr := cmd.Flag("fields").Value.String()
	filters := []string{}
	fields := []string{}
	if fieldStr != "" {
		fields = strings.Split(fieldStr, ",")
	}
	filterStr := cmd.Flag("filter").Value.String()
	if filterStr != "" {
		filters = strings.Split(filterStr, ",")
	}

	sortColumns, err := cmd.Flags().GetBool("sort-columns")
	if err != nil {
		return nil, err
	}

	return &FieldsFilterSettings{
		Fields:         fields,
		Filters:        filters,
		SortColumns:    sortColumns,
		ReorderColumns: fields,
	}, nil
}

func SetupProcessor(cmd *cobra.Command) (*GlazeProcessor, formatters.OutputFormatter, error) {
	outputSettings, err := ParseOutputFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	fieldsFilterSettings, err := ParseFieldsFilterFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	selectSettings, err := ParseSelectFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing select flags")
	}
	outputSettings.UpdateWithSelectSettings(selectSettings)
	fieldsFilterSettings.UpdateWithSelectSettings(selectSettings)
	templateSettings.UpdateWithSelectSettings(selectSettings)

	renameSettings, err := ParseRenameFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing rename flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	// rename middlewares run first because they are used to clean up column names
	// for the following middlewares too.
	// these following middlewares can create proper column names on their own
	// when needed
	err = renameSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding rename middlewares")
	}

	err = templateSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := middlewares.NewFlattenObjectMiddleware()
		of.AddTableMiddleware(mw)
	}
	fieldsFilterSettings.AddMiddlewares(of)

	var middlewares_ []middlewares.ObjectMiddleware
	if !templateSettings.UseRowTemplates && len(templateSettings.Templates) > 0 {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}
