package cli

import (
	_ "embed"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func initFlagsFromYaml(yamlContent []byte) (map[string]*cmds.ParameterDefinition, []*cmds.ParameterDefinition) {
	flags := make(map[string]*cmds.ParameterDefinition)
	flagList := make([]*cmds.ParameterDefinition, 0)

	var err error
	var parameters []*cmds.ParameterDefinition

	err = yaml.Unmarshal(yamlContent, &parameters)
	if err != nil {
		panic(errors.Wrap(err, "Failed to unmarshal output flags yaml"))
	}

	for _, p := range parameters {
		err := p.CheckParameterDefaultValueValidity()
		if err != nil {
			panic(errors.Wrap(err, "Failed to check parameter default value validity"))
		}
		flags[p.Name] = p
		flagList = append(flagList, p)
	}

	return flags, flagList
}
