package help

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

func NewCobraHelpCommand() *cobra.Command {
	var ret *cobra.Command
	ret = &cobra.Command{
		Use:   "help [topic/command]",
		Short: "Help about any command or topic",
		Long:  `Help provides help for any command and topic in the application.`,
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// copied from cobra itself
			var completions []string
			cmd, _, e := c.Root().Find(args)
			if e != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if cmd == nil {
				// Root help command.
				cmd = c.Root()
			}
			for _, subCmd := range cmd.Commands() {
				if subCmd.IsAvailableCommand() || subCmd == ret {
					if strings.HasPrefix(subCmd.Name(), toComplete) {
						completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
					}
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		},

		Run: func(c *cobra.Command, args []string) {
			cmd, _, e := c.Root().Find(args)
			if cmd == nil || e != nil {
				c.Printf("Unknown help topic %#q\n", args)
				cobra.CheckErr(c.Root().Usage())
			} else {
				cmd.InitDefaultHelpFlag()    // make possible 'help' flag to be shown
				cmd.InitDefaultVersionFlag() // make possible 'version' flag to be shown
				cobra.CheckErr(cmd.Help())
			}
		},
	}

	return ret
}
