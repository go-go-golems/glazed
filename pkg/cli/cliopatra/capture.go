package cliopatra

import "github.com/go-go-golems/glazed/pkg/cmds"

func NewProgramFromCapture(
	description *cmds.CommandDescription,
	ps map[string]interface{},
	opts ...ProgramOption,
) (*Program, error) {
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

			ret.Flags = append(ret.Flags, param)
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

		ret.Flags = append(ret.Flags, param)
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

		ret.Args = append(ret.Args, param)
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret, nil
}
