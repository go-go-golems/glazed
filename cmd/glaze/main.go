package main

import (
	"context"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds/html"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "glaze",
	Short:   "glaze is a tool to format structured data",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := logging.InitLoggerFromViper()
		cobra.CheckErr(err)
	},
}

func main() {
	err := logging.AddLoggingLayerToRootCommand(rootCmd, "glaze")
	cobra.CheckErr(err)

	err = viper.BindPFlags(rootCmd.PersistentFlags())
	cobra.CheckErr(err)

	err = logging.InitLoggerFromViper()
	cobra.CheckErr(err)

	ctx := context.Background()
	helpSystem, err := store.NewInMemoryHelpSystem()
	cobra.CheckErr(err)
	defer helpSystem.Close()

	err = doc.AddDocToHelpSystem(ctx, helpSystem)
	cobra.CheckErr(err)

	helpSystem.SetupCobraRootCommand(rootCmd)

	jsonCmd, err := cmds.NewJsonCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(jsonCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	yamlCmd, err := cmds.NewYamlCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(yamlCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)
	rootCmd.AddCommand(cmds.DocsCmd)
	rootCmd.AddCommand(cmds.MarkdownCmd)

	exampleCmd, err := cmds.NewExampleCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(exampleCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	csvCmd, err := cmds.NewCsvCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommandFromGlazeCommand(csvCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	htmlCommand, err := html.NewHTMLCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(htmlCommand)

	_ = rootCmd.Execute()
}
