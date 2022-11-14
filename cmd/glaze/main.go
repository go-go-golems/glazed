package main

import (
	"dd-cli/pkg/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "glaze",
	Short: "glaze is a tool to format structured data",
}

var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Format JSON data",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func main() {
	_ = rootCmd.Execute()

	cli.AddOutputFlags(jsonCmd)
	rootCmd.AddCommand(jsonCmd)

}
