package main

import (
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/cmd/glaze/cmds"
	"github.com/wesen/glazed/pkg/help"
)

var rootCmd = &cobra.Command{
	Use:   "glaze",
	Short: "glaze is a tool to format structured data",
}

func main() {
	_ = rootCmd.Execute()
}

func init() {
	helpCmd := help.NewCobraHelpCommand()
	rootCmd.SetHelpCommand(helpCmd)
	rootCmd.AddCommand(cmds.JsonCmd)
	rootCmd.AddCommand(cmds.YamlCmd)
}
