package cli

import (
	"github.com/spf13/cobra"
	"glazed/pkg/types"
	"strings"
)

// Helpers for cobra commands

func AddOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format (table, csv, tsv, json, sqlite)")
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
	}, nil
}

// TODO(manuel, 2022-11-20) Make it easy for the developer to configure which flag they want
// and which they don't

// AddFieldsFilterFlags adds the flags for the following middlewares to the cmd:
// - FieldsFilterMiddleware
// - SortColumnsMiddleware
// - ReorderColumnOrderMiddleware
func AddFieldsFilterFlags(cmd *cobra.Command) {
	cmd.Flags().String("fields", "", "Fields to include in the output, default: all")
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
