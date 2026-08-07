// glazed-migrate finds and fixes usages of removed Glazed APIs.
//
// It is the command-framework front end for the glazedmigration analyzer:
// "check" reports findings as structured rows, "fix" applies the automatic
// migrations in place. Embedded help is available via:
//
//	glazed-migrate help glazed-migrate-guide
//
// The analyzer also remains usable as a vettool through the companion
// singlechecker entry point; see the help pages for both workflows.
package main

import (
	"embed"

	migratecmds "github.com/go-go-golems/glazed/cmd/tools/glazed-migrate/cmds"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"
)

//go:embed doc
var docFS embed.FS

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "glazed-migrate",
	Short:   "Find and fix usages of removed Glazed APIs",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return logging.InitLoggerFromCobra(cmd)
	},
}

func main() {
	err := logging.AddLoggingSectionToRootCommand(rootCmd, "glazed-migrate")
	cobra.CheckErr(err)

	helpSystem := help.NewHelpSystem()
	err = helpSystem.LoadSectionsFromFS(docFS, "doc")
	cobra.CheckErr(err)
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	checkCmd := migratecmds.NewCheckCommand()
	fixCmd := migratecmds.NewFixCommand()
	err = cli.AddCommandsToRootCommand(rootCmd, []cmds.Command{checkCmd, fixCmd}, nil)
	cobra.CheckErr(err)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
