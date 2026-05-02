package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "docsctl",
		Short:   "Publish Glazed help databases to a shared docs registry",
		Version: version,
		Long: `docsctl validates and publishes Glazed help SQLite databases.

It is intended for package release workflows that publish versioned help exports
to a shared docs.yolo.scapegoat.dev registry. The first implementation phase
adds local validation and direct registry upload using package-scoped publish
tokens.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newValidateCommand())
	return cmd
}

func main() {
	if err := newRootCommand().Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
