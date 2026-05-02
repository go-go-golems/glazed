package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		if handled, err := handleDocsctlPrintFlags(command, parsedValues, cmd.OutOrStdout()); handled || err != nil {
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

func handleDocsctlPrintFlags(command cmds.Command, parsedValues *values.Values, w io.Writer) (bool, error) {
	commandSettingsValues, ok := parsedValues.Get(cli.CommandSettingsSlug)
	if !ok {
		return false, nil
	}
	commandSettings := &cli.CommandSettings{}
	if err := commandSettingsValues.DecodeInto(commandSettings); err != nil {
		return true, err
	}
	switch {
	case commandSettings.PrintParsedFields:
		return true, printDocsctlParsedFields(w, parsedValues)
	case commandSettings.PrintYAML:
		return true, command.ToYAML(w)
	case commandSettings.PrintSchema:
		schema, err := command.Description().ToJsonSchema()
		if err != nil {
			return true, err
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return true, enc.Encode(schema)
	default:
		return false, nil
	}
}

func printDocsctlParsedFields(w io.Writer, parsedValues *values.Values) error {
	sectionsMap := map[string]map[string]interface{}{}
	parsedValues.ForEach(func(sectionName string, sectionValues *values.SectionValues) {
		fieldValues := map[string]interface{}{}
		sectionValues.Fields.ForEach(func(name string, fieldValue *fields.FieldValue) {
			serializable := fields.ToSerializableFieldValue(fieldValue)
			fieldMap := map[string]interface{}{
				"value": serializable.Value,
			}
			if len(serializable.Log) > 0 {
				logs := make([]map[string]interface{}, 0, len(serializable.Log))
				for _, l := range serializable.Log {
					logEntry := map[string]interface{}{
						"source": l.Source,
						"value":  l.Value,
					}
					if len(l.Metadata) > 0 {
						logEntry["metadata"] = l.Metadata
					}
					logs = append(logs, logEntry)
				}
				fieldMap["log"] = logs
			}
			fieldValues[name] = fieldMap
		})
		sectionsMap[sectionName] = fieldValues
	})
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer func() { _ = enc.Close() }()
	return enc.Encode(sectionsMap)
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
