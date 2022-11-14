package cli

import "github.com/spf13/cobra"

// Helpers for cobra commands

func AddOutputFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "table", "Output format (table, json, sqlite)")
	cmd.Flags().StringP("output-file", "f", "", "Output file")
	cmd.Flags().String("table-format", "ascii", "Table format (ascii, markdown, html, csv, tsv)")
	cmd.Flags().Bool("with-headers", true, "Include headers in output (CSV, TSV)")
	cmd.Flags().String("csv-separator", ",", "CSV separator")
}

func AddTemplateFlags(cmd *cobra.Command) {
	cmd.Flags().String("template", "", "Go Template to use for single string")
	cmd.Flags().StringSlice("template-field", nil, "For table output, fieldName:template to create new fields, or @fileName to read field templates from a yaml dictionary")
}

func AddFieldsFilterFlags(cmd *cobra.Command) {
	cmd.Flags().String("fields", "", "Fields to include in the output, default: all")
	cmd.Flags().String("filter", "", "Fields to remove from output")
}
