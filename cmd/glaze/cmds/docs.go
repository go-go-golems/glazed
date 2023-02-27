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
		gp, of, err := cli.CreateGlazedProcessorFromCobra(cmd)
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
	// This is an example of selective use of glazed parameter layers.
	// If we extracted out the docts command into a cmds.Command, which we should
	// in order to expose it as a REST API, all of this would not even be necessary,
	// I think.
	gpl, err := cli.NewGlazedParameterLayers()
	if err != nil {
		panic(err)
	}
	err = gpl.OutputParameterLayer.AddFlagsToCobraCommand(DocsCmd)
	if err != nil {
		panic(err)
	}
	err = gpl.TemplateParameterLayer.AddFlagsToCobraCommand(DocsCmd)
	if err != nil {
		panic(err)
	}

	// TODO(2023-02-12, manuel) Overload settings could be loaded from YAML too
	//
	// See https://github.com/go-go-golems/glazed/issues/133
	defaults := &cli.FieldsFilterFlagsDefaults{
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
		Filter:      []string{},
		SortColumns: false,
	}
	err = gpl.FieldsFiltersParameterLayer.InitializeParameterDefaultsFromStruct(defaults)
	if err != nil {
		panic(err)
	}
	err = gpl.FieldsFiltersParameterLayer.AddFlagsToCobraCommand(DocsCmd)
	if err != nil {
		panic(err)
	}
}
