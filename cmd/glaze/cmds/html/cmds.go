package html

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
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
			gp, of, err := cli.CreateGlazedProcessorFromCobra(cmd)
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

				err = outputNodesDepthFirst(doc, gp)
				cobra.CheckErr(err)
			}

			s, err := of.Output()
			cobra.CheckErr(err)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
			fmt.Print(s)
		},
	}

	g, err := cli.NewGlazedParameterLayers()
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
			gp, of, err := cli.CreateGlazedProcessorFromCobra(cmd)
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

				hsp := NewHTMLHeadingSplitParser(gp, []string{"span"})

				_, err = hsp.ProcessNode(doc)
				cobra.CheckErr(err)
			}

			s, err := of.Output()
			cobra.CheckErr(err)
			if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
				os.Exit(0)
			}
			fmt.Print(s)

		},
	}

	err = g.AddFlagsToCobraCommand(extractCmd)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(extractCmd)

	return cmd, nil
}
