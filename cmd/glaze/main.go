package main

import (
	"embed"
	"github.com/go-go-golems/glazed/cmd/glaze/cmds"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cli"
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
	command, err := glazed_cmds.BuildCobraCommandFromGlazeCommand(jsonCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	yamlCmd, err := cmds.NewYamlCommand()
	cobra.CheckErr(err)
	command, err = glazed_cmds.BuildCobraCommandFromGlazeCommand(yamlCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)
	rootCmd.AddCommand(cmds.DocsCmd)
	rootCmd.AddCommand(cmds.MarkdownCmd)

	csvCmd, err := cmds.NewCsvCommand()
	cobra.CheckErr(err)
	command, err = glazed_cmds.BuildCobraCommandFromGlazeCommand(csvCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)

	htmlCommand, err := cmds.NewHTMLCommand()
	cobra.CheckErr(err)
	rootCmd.AddCommand(htmlCommand)

	_ = rootCmd.Execute()
}
