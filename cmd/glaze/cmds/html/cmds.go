package html

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	schema "github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

func NewHTMLCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "html",
		Short: "HTML commands",
	}

	parseCmd := &cobra.Command{
		Use:   "parse",
		Short: "Parse HTML",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			gp, _, err := cli.CreateGlazedProcessorFromCobra(cmd)
			cobra.CheckErr(err)

			for _, arg := range args {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer func(f *os.File) {
					_ = f.Close()
				}(f)

				doc, err := html.Parse(f)
				cobra.CheckErr(err)

				err = outputNodesDepthFirst(ctx, doc, gp)
				cobra.CheckErr(err)
			}
			err = gp.Close(ctx)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
			cobra.CheckErr(err)
		},
	}

	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	cobraSection, ok := glazedSection.(schema.CobraSection)
	if !ok {
		return nil, fmt.Errorf("glazed section is not a CobraSection")
	}

	err = cobraSection.AddSectionToCobraCommand(parseCmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(parseCmd)

	extractCmd := &cobra.Command{
		Use:   "extract",
		Short: "Extract HTML from sections",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			gp, _, err := cli.CreateGlazedProcessorFromCobra(cmd)
			cobra.CheckErr(err)

			for _, arg := range args {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer func(f *os.File) {
					_ = f.Close()
				}(f)

				doc, err := html.Parse(f)
				cobra.CheckErr(err)

				removeTags, err := cmd.Flags().GetStringSlice("remove")
				cobra.CheckErr(err)
				splitTags, err := cmd.Flags().GetStringSlice("heading")
				cobra.CheckErr(err)
				extractTitle, err := cmd.Flags().GetBool("extract-title")
				cobra.CheckErr(err)

				hsp := NewHTMLSplitParser(gp, append(removeTags, splitTags...), splitTags, extractTitle)

				_, err = hsp.ProcessNode(ctx, doc)
				cobra.CheckErr(err)
			}

			err = gp.Close(ctx)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
			cobra.CheckErr(err)
		},
	}

	extractCmd.Flags().StringSlice("heading", []string{"h1", "h2", "h3", "h4", "h5", "h6"}, "Heading tags to split on")
	extractCmd.Flags().StringSlice("remove", []string{"span"}, "Tags to remove from the output")
	extractCmd.Flags().Bool("extract-title", true, "Extract the title from the sections")

	err = cobraSection.AddSectionToCobraCommand(extractCmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(extractCmd)

	return cmd, nil
}
