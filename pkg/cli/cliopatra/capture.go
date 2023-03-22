package cliopatra

import "github.com/go-go-golems/glazed/pkg/cmds"

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

	for _, layer := range description.Layers {
		for _, p := range layer.GetParameterDefinitions() {
			name := layer.GetPrefix() + p.Name
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
				Flag:  flag,
				Short: p.Help,
				Type:  p.Type,
				Value: v,
			}

			// TODO(manuel, 2023-03-21) This would be easier if we knew why and from where something is set
			//
			// Right now we can only kind of guess, by doing some comparison.
			//
			// See https://github.com/go-go-golems/glazed/issues/239
			if !p.IsEqualToDefault(v) {
				ret.Flags = append(ret.Flags, param)
			}
		}
	}

	for _, p := range description.Flags {
		name := p.Name
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

		param := &Parameter{
			Name:  name,
			Flag:  flag,
			Short: p.Help,
			Type:  p.Type,
			Value: v,
		}

		if !p.IsEqualToDefault(v) {
			ret.Flags = append(ret.Flags, param)
		}
	}

	for _, p := range description.Arguments {
		name := p.Name
		flag := name

		v, ok := ps[name]
		if !ok {
			continue
		}

		param := &Parameter{
			Name:  name,
			Flag:  flag,
			Short: p.Help,
			Type:  p.Type,
			Value: v,
		}

		if !p.IsEqualToDefault(v) {
			ret.Args = append(ret.Args, param)
		}
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}
