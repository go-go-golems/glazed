package main

import (
	"context"
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cliopatra"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"os"
	"strings"
)

//func saveArgsAndFlagsToYAML(program Program, fileName string) error {
//	data, err := yaml.Marshal(program)
//	if err != nil {
//		return err
//	}
//	f, err := os.Create(fileName)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//	_, err = f.Write(data)
//	return err
//}

func main() {
	p := &cliopatra.Program{
		Name: "ls",
		Flags: []*cliopatra.Parameter{
			{
				Name:    "l",
				Flag:    "-l",
				Type:    parameters.ParameterTypeBool,
				NoValue: true,
			},
		},
		Args: []*cliopatra.Parameter{
			{
				Name:  "path",
				Short: "p",
				Type:  parameters.ParameterTypeString,
				Value: "./",
			},
		},
	}

	buf := &strings.Builder{}

	err := p.RunIntoWriter(
		context.Background(),
		map[string]*layers.ParsedParameterLayer{},
		map[string]interface{}{},
		buf,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(buf.String())
}
