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
	if err != nil {
		panic(err)
	}
	command, err := glazed_cmds.BuildCobraCommandFromGlazeCommand(jsonCmd)
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(command)

	yamlCmd, err := cmds.NewYamlCommand()
	if err != nil {
		panic(err)
	}
	command, err = glazed_cmds.BuildCobraCommandFromGlazeCommand(yamlCmd)
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(command)
	rootCmd.AddCommand(cmds.DocsCmd)
	rootCmd.AddCommand(cmds.MarkdownCmd)

	csvCmd, err := cmds.NewCsvCommand()
	if err != nil {
		panic(err)
	}
	command, err = glazed_cmds.BuildCobraCommandFromGlazeCommand(csvCmd)
	if err != nil {
		panic(err)
	}
	rootCmd.AddCommand(command)
	_ = rootCmd.Execute()
}
