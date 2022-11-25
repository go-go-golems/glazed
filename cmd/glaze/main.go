package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"glazed/pkg/cli"
	"glazed/pkg/formatters"
	"glazed/pkg/middlewares"
	"glazed/pkg/types"
	"gopkg.in/yaml.v3"
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

var yamlCmd = &cobra.Command{
	Use:   "yaml",
	Short: "Format YAML data",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No input file specified")
			os.Exit(1)
		}

		gp, of, err := setupProcessor(cmd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Could not create Glaze processor: %v\n", err)
			os.Exit(1)
		}

		for _, arg := range args {
			inputBytes, err := os.ReadFile(arg)

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
			err = gp.processInputObject(data2)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error processing file %s as object: %s\n", arg, err)
				os.Exit(1)
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

var jsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Format JSON data",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("No input file specified")
			os.Exit(1)
		}

		gp, of, err := setupProcessor(cmd)
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

func setupProcessor(cmd *cobra.Command) (*GlazeProcessor, formatters.OutputFormatter, error) {
	outputSettings, err := cli.ParseOutputFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing output flags")
	}

	templateSettings, err := cli.ParseTemplateFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing template flags")
	}

	fieldsFilterSettings, err := cli.ParseFieldsFilterFlags(cmd)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error parsing fields filter flags")
	}

	of, err := outputSettings.CreateOutputFormatter()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error creating output formatter")
	}

	err = templateSettings.AddMiddlewares(of)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error adding template middlewares")
	}

	if (outputSettings.Output == "json" || outputSettings.Output == "yaml") && outputSettings.FlattenObjects {
		mw := middlewares.NewFlattenObjectMiddleware()
		of.AddTableMiddleware(mw)
	}
	fieldsFilterSettings.AddMiddlewares(of)

	var middlewares_ []middlewares.ObjectMiddleware
	if templateSettings.UseRowTemplates {
		ogtm, err := middlewares.NewObjectGoTemplateMiddleware(templateSettings.Templates)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Could not process template argument")
		}
		middlewares_ = append(middlewares_, ogtm)
	}

	gp := NewGlazeProcessor(of, middlewares_)
	return gp, of, nil
}

func main() {
	_ = rootCmd.Execute()
}

func init() {
	// TODO(manuel, 2022-11-20) We should make it possible to specify the names of the flags
	// if the defaults don't make use happy. Potentially with a builder interface
	jsonCmd.Flags().SortFlags = false
	cli.AddOutputFlags(jsonCmd)
	cli.AddTemplateFlags(jsonCmd)
	cli.AddFieldsFilterFlags(jsonCmd)

	// json input options
	jsonCmd.Flags().Bool("input-is-array", false, "Input is an array of objects")

	yamlCmd.Flags().SortFlags = false
	cli.AddOutputFlags(yamlCmd)
	cli.AddTemplateFlags(yamlCmd)
	cli.AddFieldsFilterFlags(yamlCmd)

	rootCmd.AddCommand(jsonCmd)
	rootCmd.AddCommand(yamlCmd)
}
