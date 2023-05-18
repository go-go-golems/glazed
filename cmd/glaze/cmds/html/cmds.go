package html

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
	"os"
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

			gp, err := cli.CreateGlazedProcessorFromCobra(cmd)
			cobra.CheckErr(err)

			for _, arg := range args {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer f.Close()

				doc, err := html.Parse(f)
				cobra.CheckErr(err)

				err = outputNodesDepthFirst(ctx, doc, gp)
				cobra.CheckErr(err)
			}

			err = gp.OutputFormatter().Output(ctx, os.Stdout)
			cobra.CheckErr(err)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
		},
	}

	g, err := settings.NewGlazedParameterLayers()
	if err != nil {
		return nil, err
	}
	err = g.AddFlagsToCobraCommand(parseCmd)
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

			gp, err := cli.CreateGlazedProcessorFromCobra(cmd)
			cobra.CheckErr(err)

			for _, arg := range args {
				if arg == "-" {
					arg = "/dev/stdin"
				}
				f, err := os.Open(arg)
				cobra.CheckErr(err)
				defer f.Close()

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

			err = gp.OutputFormatter().Output(ctx, os.Stdout)
			cobra.CheckErr(err)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
		},
	}

	extractCmd.Flags().StringSlice("heading", []string{"h1", "h2", "h3", "h4", "h5", "h6"}, "Heading tags to split on")
	extractCmd.Flags().StringSlice("remove", []string{"span"}, "Tags to remove from the output")
	extractCmd.Flags().Bool("extract-title", true, "Extract the title from the sections")

	err = g.AddFlagsToCobraCommand(extractCmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(extractCmd)

	return cmd, nil
}
