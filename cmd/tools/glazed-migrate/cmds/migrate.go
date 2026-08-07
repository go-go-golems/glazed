// Package cmds holds the Glazed commands for the glazed-migrate binary.
package cmds

import (
	"context"
	"sort"

	"github.com/go-go-golems/glazed/pkg/analysis/glazedmigration"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// ScanSettings decodes the shared positional arguments for check and fix.
type ScanSettings struct {
	Paths []string `glazed:"paths"`
}

func pathsArgument() *fields.Definition {
	return fields.New(
		"paths",
		fields.TypeStringList,
		fields.WithIsArgument(true),
		fields.WithDefault([]string{"."}),
		fields.WithHelp("Directories (walked recursively) or .go files to scan"),
	)
}

// CheckCommand reports migration findings without editing files.
type CheckCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*CheckCommand)(nil)

// NewCheckCommand creates the check subcommand.
func NewCheckCommand() *CheckCommand {
	return &CheckCommand{CommandDescription: cmds.NewCommandDescription(
		"check",
		cmds.WithShort("Report usages of removed Glazed APIs"),
		cmds.WithLong(`Scan Go source for usages of Glazed APIs removed by the structured-output
cleanup (NewGlazedSchema/NewGlazedSection, With*SectionOptions, the
"output" default key, GlazedSlug, removed feature sections, and more).

Each emitted row is one finding: file, line, column, message, and the number
of automatic fixes available. No files are modified; run "glazed-migrate fix"
to apply the automatic fixes.

The scanner parses source directly and needs no type-checking, so it works on
code that no longer compiles against the current Glazed release.`),
		cmds.WithArguments(pathsArgument()),
	)}
}

// RunIntoGlazeProcessor implements cmds.GlazeCommand.
func (c *CheckCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsed *values.Values,
	processor middlewares.Processor,
) error {
	settings := &ScanSettings{}
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	diagnostics, err := glazedmigration.Scan(settings.Paths)
	if err != nil {
		return err
	}
	for _, d := range diagnostics {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := processor.AddRow(ctx, types.NewRow(
			types.MRP("file", d.File),
			types.MRP("line", d.Line),
			types.MRP("column", d.Column),
			types.MRP("message", d.Message),
			types.MRP("fixes_available", d.FixCount),
		)); err != nil {
			return err
		}
	}
	return nil
}

// FixCommand applies the automatic migrations in place.
type FixCommand struct {
	*cmds.CommandDescription
}

var _ cmds.GlazeCommand = (*FixCommand)(nil)

// NewFixCommand creates the fix subcommand.
func NewFixCommand() *FixCommand {
	return &FixCommand{CommandDescription: cmds.NewCommandDescription(
		"fix",
		cmds.WithShort("Apply automatic Glazed API migrations in place"),
		cmds.WithLong(`Scan Go source for usages of removed Glazed APIs and rewrite every usage
that has an automatic migration, editing files in place. Edits within a file
are applied back to front so offsets stay valid; conflicting edits are
skipped and reported.

Each emitted row is one modified file with the number of edits applied.
Findings without an automatic fix (removed feature sections, Setup* runtime
helpers) are reported as rows with zero edits so they can be redesigned by
hand. Review the diff afterwards; the tool deliberately performs no
type-checking.`),
		cmds.WithArguments(pathsArgument()),
	)}
}

// RunIntoGlazeProcessor implements cmds.GlazeCommand.
func (c *FixCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	parsed *values.Values,
	processor middlewares.Processor,
) error {
	settings := &ScanSettings{}
	if err := parsed.DecodeSectionInto(schema.DefaultSlug, settings); err != nil {
		return err
	}

	diagnostics, err := glazedmigration.Scan(settings.Paths)
	if err != nil {
		return err
	}

	appliedPerFile, skipped, err := glazedmigration.ApplyFixes(diagnostics)
	if err != nil {
		return err
	}

	files := make([]string, 0, len(appliedPerFile))
	for f := range appliedPerFile {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := processor.AddRow(ctx, types.NewRow(
			types.MRP("file", f),
			types.MRP("line", 0),
			types.MRP("message", "applied automatic migrations"),
			types.MRP("edits_applied", appliedPerFile[f]),
		)); err != nil {
			return err
		}
	}

	if skipped > 0 {
		if err := processor.AddRow(ctx, types.NewRow(
			types.MRP("file", ""),
			types.MRP("line", 0),
			types.MRP("message", "conflicting edits skipped (review manually)"),
			types.MRP("edits_applied", skipped),
		)); err != nil {
			return err
		}
	}

	// Report-only findings (no automatic fix) remain after fixing; surface
	// them so the user knows manual work is left.
	for _, d := range diagnostics {
		if d.FixCount > 0 {
			continue
		}
		if err := processor.AddRow(ctx, types.NewRow(
			types.MRP("file", d.File),
			types.MRP("line", d.Line),
			types.MRP("message", "manual migration required: "+d.Message),
			types.MRP("edits_applied", 0),
		)); err != nil {
			return err
		}
	}
	return nil
}
