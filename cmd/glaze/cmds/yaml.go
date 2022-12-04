package cmds

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wesen/glazed/pkg/cli"
	"gopkg.in/yaml.v3"
	"os"
)

var YamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "Format YAML data",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No input file specified")
			os.Exit(1)
		}

		gp, of, err := SetupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create Glaze processor: %v\n", err)
			os.Exit(1)
		}

		inputIsArray, err := cmd.Flags().GetBool("input-is-array")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing input-is-array)lag: %v\n", err)
			os.Exit(1)
		}

		for _, arg := range args {

			inputBytes, err := os.ReadFile(arg)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading file %s: %s\n", arg, err)
				os.Exit(1)
			}

			if inputIsArray {
				data := make([]map[string]interface{}, 0)
				err = yaml.Unmarshal(inputBytes, &data)
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
				data := make(map[interface{}]interface{})
				data2 := make(map[string]interface{})

				err = yaml.Unmarshal(inputBytes, data)
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error decoding file %s as object: %s\n", arg, err)
					os.Exit(1)
				}
				for k, v := range data {
					data2[fmt.Sprintf("%v", k)] = v
				}
				err = gp.ProcessInputObject(data2)
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
	YamlCmd.Flags().SortFlags = false
	cli.AddOutputFlags(YamlCmd)
	cli.AddTemplateFlags(YamlCmd)
	cli.AddFieldsFilterFlags(YamlCmd, "")

	// json input options
	YamlCmd.Flags().Bool("input-is-array", false, "Input is an array of objects")

}
