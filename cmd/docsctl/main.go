package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/spf13/cobra"
)

var version = "dev"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docsctl",
		Short:   "Publish Glazed help databases to a shared docs registry",
		Version: version,
		Long: `docsctl validates and publishes Glazed help SQLite databases.

It is intended for package release workflows that publish versioned help exports
to a shared docs.yolo.scapegoat.dev registry. The first implementation phase
adds local validation and direct registry upload using package-scoped publish
tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
	}
	if err := logging.AddLoggingSectionToRootCommand(cmd, "docsctl"); err != nil {
		cobra.CheckErr(err)
	}
	for _, command := range []cmds.Command{mustCommand(NewValidateCommand()), mustCommand(NewPublishCommand())} {
		cobraCmd, err := buildDocsctlCobraCommand(command)
		if err != nil {
			cobra.CheckErr(err)
		}
		cmd.AddCommand(cobraCmd)
	}
	return cmd
}

func buildDocsctlCobraCommand(command cmds.Command) (*cobra.Command, error) {
	description := command.Description()
	cmd := cli.NewCobraCommandFromCommandDescription(description)
	parser, err := cli.NewCobraParserFromSections(description.Schema, &cli.CobraParserConfig{})
	if err != nil {
		return nil, err
	}
	if err := parser.AddToCobraCommand(cmd); err != nil {
		return nil, err
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		parsedValues, err := parser.Parse(cmd, args)
		if err != nil {
			return err
		}
		if writerCommand, ok := command.(cmds.WriterCommand); ok {
			return writerCommand.RunIntoWriter(cmd.Context(), parsedValues, cmd.OutOrStdout())
		}
		if bareCommand, ok := command.(cmds.BareCommand); ok {
			return bareCommand.Run(cmd.Context(), parsedValues)
		}
		return fmt.Errorf("unsupported docsctl command type %T", command)
	}
	return cmd, nil
}

func mustCommand(command cmds.Command, err error) cmds.Command {
	if err != nil {
		cobra.CheckErr(err)
	}
	return command
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
