package main

import (
	"embed"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds/html"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "glaze",
	Short: "glaze is a tool to format structured data",
}

//go:embed doc/*
var docFS embed.FS

func main() {
	helpSystem := help.NewHelpSystem()
	err := helpSystem.LoadSectionsFromFS(docFS, ".")
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
