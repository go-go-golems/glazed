package cmds

import (
	"bytes"
	"github.com/adrg/frontmatter"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/spf13/cobra"
	"os"
)

var DocsCmd = &cobra.Command{
	Use:   "docs [flags] file [file...]",
	Short: "Work with help documents",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		gp, err := cli.CreateGlazedProcessorFromCobra(cmd)
		cobra.CheckErr(err)

		for _, arg := range args {
			// read markdown file
			s, err := os.ReadFile(arg)
			cobra.CheckErr(err)

			var metaData types.Row
			inputReader := bytes.NewReader(s)
			_, err = frontmatter.Parse(inputReader, &metaData)
			cobra.CheckErr(err)

			metaData.Set("path", arg)

			// TODO(manuel, 2023-06-25) It would be nice to unmarshal the YAML to an orderedmap
			// See https://github.com/go-go-golems/glazed/issues/305
			err = gp.AddRow(ctx, metaData)
			cobra.CheckErr(err)
		}

		err = gp.Close(ctx)
		if _, ok := err.(*cmds.ExitWithoutGlazeError); ok {
			os.Exit(0)
		}
		cobra.CheckErr(err)
	},
}

func init() {
	DocsCmd.Flags().SortFlags = false
	// This is an example of selective use of glazed parameter layers.
	// If we extracted out the docs command into a cmds.GlazeCommand, which we should
	// in order to expose it as a REST API, all of this would not even be necessary,
	// I think.
	gpl, err := settings.NewGlazedParameterLayers(
		settings.WithFieldsFiltersParameterLayerOptions(
			layers.WithDefaults(
				&settings.FieldsFilterFlagsDefaults{
					Fields: []string{
						"path",
						"Title",
						"SectionType",
						"Slug",
						"Commands",
						"Flags",
						"Topics",
						"IsTopLevel",
						"ShowPerDefault",
					},
				},
			),
		),
	)
	if err != nil {
		panic(err)
	}

	err = gpl.AddFlagsToCobraCommand(DocsCmd)
	if err != nil {
		panic(err)
	}
}
