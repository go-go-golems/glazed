package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"os"
)

type JsonCommand struct {
	description *cmds.CommandDescription
}

func NewJsonCommand() *JsonCommand {
	glazedParameterLayer, err := cli.NewGlazedParameterLayers()
	if err != nil {
		panic(err)
	}

	return &JsonCommand{
		description: cmds.NewCommandDescription(
			"json",
			cmds.WithShort("Format JSON data"),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"input-is-array",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Input is an array of objects"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input-files",
					parameters.ParameterTypeStringList,
					parameters.WithRequired(true),
				),
			),
			cmds.WithLayers(
				glazedParameterLayer,
			),
		),
	}
}

func (j *JsonCommand) Run(ps map[string]interface{}, gp *cmds.GlazeProcessor) error {
	inputIsArray, ok := ps["input-is-array"].(bool)
	if !ok {
		return fmt.Errorf("input-is-array flag is not a bool")
	}

	inputFiles, ok := ps["input-files"].([]string)
	if !ok {
		return fmt.Errorf("input-files is not a string list")
	}

	for _, arg := range inputFiles {
		if arg == "-" {
			arg = "/dev/stdin"
		}
		f, err := os.Open(arg)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error opening file %s: %s\n", arg, err)
			os.Exit(1)
		}

		if inputIsArray {
			data := make([]map[string]interface{}, 0)
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as array", arg)
			}

			i := 1
			for _, d := range data {
				err = gp.ProcessInputObject(d)
				if err != nil {
					return errors.Wrapf(err, "Error processing row %d of file %s as object", i, arg)
				}
				i++
			}
		} else {
			// read json file
			data := make(map[string]interface{})
			err = json.NewDecoder(f).Decode(&data)
			if err != nil {
				return errors.Wrapf(err, "Error decoding file %s as object", arg)
			}
			err = gp.ProcessInputObject(data)
			if err != nil {
				return errors.Wrapf(err, "Error processing file %s as object", arg)
			}
		}
	}

	return nil
}

func (j *JsonCommand) Description() *cmds.CommandDescription {
	return j.description
}

func (j *JsonCommand) BuildCobraCommand() *cobra.Command {
	ret := &cobra.Command{
		Use:   j.description.Name,
		Short: j.description.Short,
		Run: func(cmd *cobra.Command, args []string) {
			flags := j.description.Flags
			ps, err := parameters.GatherFlagsFromCobraCommand(cmd, flags, false)
			cobra.CheckErr(err)

			arguments := j.description.Arguments
			arguments_, err := parameters.GatherArguments(args, arguments, false)
			cobra.CheckErr(err)

			for k, v := range arguments_ {
				ps[k] = v
			}

			layers := j.description.Layers
			for _, layer := range layers {
				layerFlags, err := layer.ParseFlagsFromCobraCommand(cmd)
				cobra.CheckErr(err)

				for k, v := range layerFlags {
					ps[k] = v
				}
			}

			gp, of, err := cli.SetupProcessor(cmd)
			cobra.CheckErr(err)

			err = j.Run(ps, gp)
			cobra.CheckErr(err)

			s, err := of.Output()
			cobra.CheckErr(err)

			fmt.Println(s)
		},
	}

	ret.Flags().SortFlags = false
	err := parameters.AddFlagsToCobraCommand(ret.PersistentFlags(), j.description.Flags)
	if err != nil {
		panic(err)
	}

	for _, layer := range j.description.Layers {
		err = layer.AddFlagsToCobraCommand(ret, nil)
		if err != nil {
			panic(err)
		}
	}

	return ret
}
