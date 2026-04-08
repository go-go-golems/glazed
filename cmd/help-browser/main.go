package main

import (
	"os"

	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/glazed/pkg/help/server"
	"github.com/go-go-golems/glazed/pkg/web"
)

func main() {
	helpSystem := help.NewHelpSystem()
	spaHandler, err := web.NewSPAHandler()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
	rootCmd := server.NewServeCommand(helpSystem, spaHandler)
	rootCmd.Use = "help-browser [flags] <path> [<path>...]"
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
