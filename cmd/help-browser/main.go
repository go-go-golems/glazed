package main

import (
	"os"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/server"
	"github.com/go-go-golems/glazed/pkg/web"
)

func main() {
	helpSystem := help.NewHelpSystem()
	rootCmd := server.NewServeCommand(helpSystem, web.FS)
	rootCmd.Use = "help-browser [flags] <path> [<path>...]"
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
