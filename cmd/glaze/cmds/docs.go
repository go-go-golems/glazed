package cmds

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"glazed/pkg/cli"
	"os"
)

var DocsCmd = &cobra.Command{
	Use:   "docs [flags] file [file...]",
	Short: "Work with help documents",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gp, of, err := SetupProcessor(cmd)
		cobra.CheckErr(err)

		markdown := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
			),
		)

		for _, arg := range args {
			// read markdown file
			s, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			var buf bytes.Buffer
			context := parser.NewContext()
			err = markdown.Convert(s, &buf, parser.WithContext(context))
			cobra.CheckErr(err)

			metaData := meta.Get(context)
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
	cli.AddOutputFlags(DocsCmd)
	cli.AddTemplateFlags(DocsCmd)
	cli.AddFieldsFilterFlags(DocsCmd, "path,Title,SectionType,Slug,Commands,Flags,Topics,IsTopLevel,ShowPerDefault")
}
