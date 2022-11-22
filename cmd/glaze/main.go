package main

import (
	"dd-cli/pkg/cli"
	"dd-cli/pkg/formatters"
	"dd-cli/pkg/middlewares"
	"dd-cli/pkg/types"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "glaze",
	Short: "glaze is a tool to format structured data",
}

type GlazeProcessor struct {
	of  formatters.OutputFormatter
	oms []middlewares.ObjectMiddleware
}

func NewGlazeProcessor(of formatters.OutputFormatter, oms []middlewares.ObjectMiddleware) *GlazeProcessor {
	return &GlazeProcessor{
		of:  of,
		oms: oms,
	}
}

func (gp *GlazeProcessor) processInputObject(obj map[string]interface{}) error {
	for _, om := range gp.oms {
		obj2, err := om.Process(obj)
		if err != nil {
			return err
		}
		obj = obj2
	}

	gp.of.AddRow(&types.SimpleRow{Hash: obj})
	return nil
}

var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Format JSON data",
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			fmt.Println("No input file specified")
			os.Exit(1)
		}

		outputSettings, err := cli.ParseOutputFlags(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing output flags: %v\n", err)
			os.Exit(1)
		}

		templateSettings, err := cli.ParseTemplateFlags(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing template flags: %v\n", err)
			os.Exit(1)
		}

		fieldsFilterSettings, err := cli.ParseFieldsFilterFlags(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing fields filter flags: %v\n", err)
			os.Exit(1)
		}

		of := outputSettings.OutputFormatter

		err = templateSettings.AddMiddlewares(of)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error adding template middlewares: %v\n", err)
			os.Exit(1)
		}
		fieldsFilterSettings.AddMiddlewares(of)

		inputIsArray, err := cmd.Flags().GetBool("input-is-array")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error parsing input-is-array flag: %v\n", err)
			os.Exit(1)
		}

		var middlewares_ []middlewares.ObjectMiddleware
		if templateSettings.UseRowTemplates {
			ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error creating object template middleware: %v\n", err)
				os.Exit(1)
			}
			middlewares_ = append(middlewares_, ogtm)
		}

		gp := NewGlazeProcessor(of, middlewares_)

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
					err = gp.processInputObject(d)
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
				err = gp.processInputObject(data)
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

func main() {
	_ = rootCmd.Execute()
}

func init() {
	// TODO(manuel, 2022-11-20) We should make it possible to specify the names of the flags
	// if the defaults don't make use happy. Potentially with a builder interface
	cli.AddOutputFlags(jsonCmd)
	cli.AddTemplateFlags(jsonCmd)
	cli.AddFieldsFilterFlags(jsonCmd)

	// json input options
	jsonCmd.Flags().Bool("input-is-array", false, "Input is an array of objects")

	rootCmd.AddCommand(jsonCmd)
}
