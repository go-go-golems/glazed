package main

import (
	"github.com/go-go-golems/glazed/cmd/glaze/cmds"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds/html"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/doc"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/go-go-golems/glazed/pkg/help/server"
	"github.com/go-go-golems/glazed/pkg/help/site"
	"github.com/go-go-golems/glazed/pkg/web"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "glaze",
	Short:   "glaze is a tool to format structured data",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromCobra(cmd)
	},
}

func main() {
	err := logging.AddLoggingSectionToRootCommand(rootCmd, "glaze")
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	err = doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	// Add the export subcommand to help.
	helpCmd, _, err := rootCmd.Find([]string{"help"})
	cobra.CheckErr(err)
	err = help_cmd.AddExportCommand(helpCmd, helpSystem)
	cobra.CheckErr(err)

	// Add the serve command (embeds the React SPA via pkg/web).
	spaHandler, err := web.NewSPAHandler()
	cobra.CheckErr(err)
	serveCmd, err := server.NewServeCommand(helpSystem, spaHandler)
	cobra.CheckErr(err)
	serveCobraCmd, err := cli.BuildCobraCommand(serveCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(serveCobraCmd)

	renderSiteCmd, err := site.NewRenderSiteCommand(helpSystem)
	cobra.CheckErr(err)
	renderSiteCobraCmd, err := cli.BuildCobraCommand(renderSiteCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(renderSiteCobraCmd)

	// JSON command
	jsonCmd, err := cmds.NewJsonCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommand(jsonCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	// YAML command
	yamlCmd, err := cmds.NewYamlCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommand(yamlCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)
	rootCmd.AddCommand(cmds.DocsCmd)
	rootCmd.AddCommand(cmds.MarkdownCmd)

	exampleCmd, err := cmds.NewExampleCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommand(exampleCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	csvCmd, err := cmds.NewCsvCommand()
	cobra.CheckErr(err)
	command, err = cli.BuildCobraCommand(csvCmd,
		cli.WithParserConfig(cli.CobraParserConfig{AppName: "glaze"}),
	)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	htmlCommand, err := html.NewHTMLCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(htmlCommand)

	_ = rootCmd.Execute()
}
