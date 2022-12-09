package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/pkg/cli"
	"os"
)

var JsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Format JSON data",
	Long:  "Format JSON data LONG LONG LONG",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No input file specified")
			os.Exit(1)
		}

		gp, of, err := SetupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create glaze  procersors: %v\n", err)
			os.Exit(1)
		}

		inputIsArray, err := cmd.Flags().GetBool("input-is-array")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing input-is-array flag: %v\n", err)
			os.Exit(1)
		}

		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", arg, err)
				os.Exit(1)
			}

			if inputIsArray {
				data := make([]map[string]interface{}, 0)
				err = json.NewDecoder(f).Decode(&data)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error decoding file %s as array: %s\n", arg, err)
					os.Exit(1)
				}

				i := 1
				for _, d := range data {
					err = gp.ProcessInputObject(d)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "Error processing row %d of file %s as object: %s\n", i, arg, err)
						os.Exit(1)
					}
					i++
				}
			} else {
				// read json file
				data := make(map[string]interface{})
				err = json.NewDecoder(f).Decode(&data)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error decoding file %s as object: %s\n", arg, err)
					os.Exit(1)
				}
				err = gp.ProcessInputObject(data)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error processing file %s as object: %s\n", arg, err)
					os.Exit(1)
				}
			}
		}

		s, err := of.Output()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error rendering output: %s\n", err)
			os.Exit(1)
		}
		fmt.Println(s)
	},
}

func init() {
	// TODO(manuel, 2022-11-20) We should make it possible to specify the names of the flags
	// if the defaults don't make use happy. Potentially with a builder interface
	JsonCmd.Flags().SortFlags = false
	cli.AddOutputFlags(JsonCmd)
	cli.AddTemplateFlags(JsonCmd)
	cli.AddFieldsFilterFlags(JsonCmd, "")

	// json input options
	JsonCmd.Flags().Bool("input-is-array", false, "Input is an array of objects")
}
