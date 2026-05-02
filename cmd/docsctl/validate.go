package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/help/publish"
	"github.com/spf13/cobra"
)

type validateOptions struct {
	packageName string
	version     string
	file        string
	jsonOutput  bool
}

func newValidateCommand() *cobra.Command {
	opts := &validateOptions{}
	cmd := &cobra.Command{
		Use:   "validate --package <name> --version <version> --file <help.db>",
		Short: "Validate a Glazed help SQLite database before publishing",
		Long: `Validate opens a Glazed help SQLite export read-only and checks that it
contains a usable sections table, non-empty sections, non-empty slugs, and no
duplicate slugs. It also validates the package/version identity that will be
used by the registry publication path.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.packageName, "package", "", "Package name to publish, for example pinocchio")
	cmd.Flags().StringVar(&opts.version, "version", "", "Package version to publish, for example v1.2.3")
	cmd.Flags().StringVar(&opts.file, "file", "", "Path to the SQLite help export to validate")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit validation result as JSON")
	_ = cmd.MarkFlagRequired("package")
	_ = cmd.MarkFlagRequired("version")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func runValidate(cmd *cobra.Command, opts *validateOptions) error {
	result, err := publish.ValidateSQLiteHelpDB(context.Background(), opts.file, publish.SQLiteValidationOptions{
		PackageName: opts.packageName,
		Version:     opts.version,
	})
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, err = fmt.Fprintf(cmd.OutOrStdout(), "OK: %s is a valid Glazed help database for %s@%s (%d sections, %d slugs)\n", result.Path, opts.packageName, opts.version, result.SectionCount, result.SlugCount)
	if err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		if _, err := fmt.Fprintf(cmd.OutOrStdout(), "WARN: %s\n", warning); err != nil {
			return err
		}
	}
	return nil
}
