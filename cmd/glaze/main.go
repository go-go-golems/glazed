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

func main() {
	_ = rootCmd.Execute()
}

//go:embed doc/*
var docFS embed.FS

func init() {
	helpSystem := help.NewHelpSystem()
	err := helpSystem.LoadSectionsFromFS(docFS, ".")
	if err != nil {
		panic(err)
	}

	helpFunc, usageFunc := help.GetCobraHelpUsageFuncs(helpSystem)
	helpTemplate, usageTemplate := help.GetCobraHelpUsageTemplates(helpSystem)

	_ = usageFunc
	_ = usageTemplate

	rootCmd.SetHelpFunc(helpFunc)
	rootCmd.SetUsageFunc(usageFunc)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	helpCmd := help.NewCobraHelpCommand(helpSystem)
	rootCmd.SetHelpCommand(helpCmd)

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
}
