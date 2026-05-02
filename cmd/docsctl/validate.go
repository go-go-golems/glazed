package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/help/publish"
)

type ValidateCommand struct {
	*cmds.CommandDescription
}

type validateOptions struct {
	PackageName string `glazed:"package"`
	Version     string `glazed:"version"`
	File        string `glazed:"file"`
	JSONOutput  bool   `glazed:"json"`
}

var _ cmds.WriterCommand = (*ValidateCommand)(nil)

func NewValidateCommand() (*ValidateCommand, error) {
	return &ValidateCommand{CommandDescription: cmds.NewCommandDescription(
		"validate",
		cmds.WithShort("Validate a Glazed help SQLite database before publishing"),
		cmds.WithLong(`Validate opens a Glazed help SQLite export read-only and checks that it
contains a usable sections table, non-empty sections, non-empty slugs, and no
duplicate slugs. It also validates the package/version identity that will be
used by the registry publication path.`),
		cmds.WithFlags(
			fields.New("package", fields.TypeString, fields.WithHelp("Package name to publish, for example pinocchio"), fields.WithRequired(true)),
			fields.New("version", fields.TypeString, fields.WithHelp("Package version to publish, for example v1.2.3"), fields.WithRequired(true)),
			fields.New("file", fields.TypeString, fields.WithHelp("Path to the SQLite help export to validate"), fields.WithRequired(true)),
			fields.New("json", fields.TypeBool, fields.WithHelp("Emit validation result as JSON"), fields.WithDefault(false)),
		),
	)}, nil
}

func (c *ValidateCommand) RunIntoWriter(ctx context.Context, parsedValues *values.Values, w io.Writer) error {
	opts := &validateOptions{}
	if err := parsedValues.DecodeSectionInto(schema.DefaultSlug, opts); err != nil {
		return err
	}
	return runValidate(ctx, w, opts)
}

func runValidate(ctx context.Context, w io.Writer, opts *validateOptions) error {
	result, err := publish.ValidateSQLiteHelpDB(ctx, opts.File, publish.SQLiteValidationOptions{
		PackageName: opts.PackageName,
		Version:     opts.Version,
	})
	if err != nil {
		return err
	}

	if opts.JSONOutput {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	_, err = fmt.Fprintf(w, "OK: %s is a valid Glazed help database for %s@%s (%d sections, %d slugs)\n", result.Path, opts.PackageName, opts.Version, result.SectionCount, result.SlugCount)
	if err != nil {
		return err
	}
	for _, warning := range result.Warnings {
		if _, err := fmt.Fprintf(w, "WARN: %s\n", warning); err != nil {
			return err
		}
	}
	return nil
}
