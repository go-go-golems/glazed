package cliopatra

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/maps"
)

func getCliopatraFlag(
	definitions []*parameters.ParameterDefinition,
	ps map[string]interface{},
	prefix string,
) []*Parameter {
	ret := []*Parameter{}

	for _, p := range definitions {
		name := prefix + p.Name
		shortFlag := p.ShortFlag
		flag := name

		v, ok := ps[name]
		if !ok {
			flag = shortFlag
			v, ok = ps[shortFlag]
			if !ok {
				continue
			}
		}

		// NOTE(manuel, 2023-03-21) do we need to check for FromFile types? or is that covered by the value representation?
		param := &Parameter{
			Name:  name,
			Short: p.Help,
			Type:  p.Type,
			Value: v,
		}
		if p.Type == parameters.ParameterTypeBool {
			param.NoValue = true
		}

		if flag != name {
			param.Flag = flag
		}

		// TODO(manuel, 2023-03-21) This would be easier if we knew why and from where something is set
		//
		// Right now we can only kind of guess, by doing some comparison.
		//
		// See https://github.com/go-go-golems/glazed/issues/239
		if !p.IsEqualToDefault(v) {
			ret = append(ret, param)
		}
	}

	return ret
}

// NewProgramFromCapture is a helper function to help capture a cliopatra Program from
// the description and the parameters map of a glazed command.
//
// It will go over all the ParameterDefinition (from the layers, flags and arguments)
// and will try to create the best cliopatra map it can. It tries to resolve the prefixes
// of layered parameters.
//
// Values in the parameter map that are not present under the form of a ParameterDefinition
// will not be added to the command, and should be added separately using the WithRawFlags
// option.
func NewProgramFromCapture(
	description *cmds.CommandDescription,
	ps map[string]interface{},
	opts ...ProgramOption,
) *Program {
	ret := &Program{
		Name:        description.Name,
		Description: description.Short,
	}

	// NOTE(manuel, 2023-03-21) Maybe we should add layers to the program capture too, to expose all the parameters
	//
	// See https://github.com/go-go-golems/cliopatra/issues/6
	for _, layer := range description.Layers {
		ret.Flags = append(ret.Flags, getCliopatraFlag(maps.GetValues(layer.GetParameterDefinitions()), ps, layer.GetPrefix())...)
	}

	ret.Flags = append(ret.Flags, getCliopatraFlag(description.Flags, ps, "")...)
	ret.Args = getCliopatraFlag(description.Arguments, ps, "")

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}
