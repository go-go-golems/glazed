package cmds

import (
	"bytes"
	"fmt"
	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/spf13/cobra"
	"os"
)

var DocsCmd = &cobra.Command{
	Use:   "docs [flags] file [file...]",
	Short: "Work with help documents",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gp, of, err := cli.SetupProcessor(cmd)
		cobra.CheckErr(err)

		for _, arg := range args {
			// read markdown file
			s, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			var metaData map[string]interface{}
			inputReader := bytes.NewReader(s)
			_, err = frontmatter.Parse(inputReader, &metaData)
			cobra.CheckErr(err)

			metaData["path"] = arg

			err = gp.ProcessInputObject(metaData)
			cobra.CheckErr(err)
		}

		s, err := of.Output()
		cobra.CheckErr(err)
		fmt.Print(s)
	},
}

func init() {
	DocsCmd.Flags().SortFlags = false
	cli.AddOutputFlags(DocsCmd, cli.NewOutputFlagsDefaults())
	cli.AddTemplateFlags(DocsCmd, cli.NewTemplateFlagsDefaults())
	cli.AddFieldsFilterFlags(DocsCmd, &cli.FieldsFilterFlagsDefaults{
		Fields:      "path,Title,SectionType,Slug,Commands,Flags,Topics,IsTopLevel,ShowPerDefault",
		Filter:      "",
		SortColumns: false,
	})
}
