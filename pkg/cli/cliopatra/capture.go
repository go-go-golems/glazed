package cliopatra

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func getCliopatraParameters(
	definitions *fields.Definitions,
	ps *fields.FieldValues,
	prefix string,
) []*Parameter {
	ret := []*Parameter{}

	definitions.ForEach(func(p *fields.Definition) {
		name := prefix + p.Name
		shortFlag := p.ShortFlag
		flag := name

		v, ok := ps.Get(name)
		if !ok {
			flag = shortFlag
			v, ok = ps.Get(shortFlag)
			if !ok {
				return
			}
		}

		// NOTE(manuel, 2023-03-21) do we need to check for FromFile types? or is that covered by the value representation?
		param := &Parameter{
			Name:  name,
			Short: p.Help,
			Type:  p.Type,
			Value: v.Value,
			Log:   v.Log,
		}
		if p.Type == fields.TypeBool {
			param.NoValue = true
		}

		if flag != name {
			param.Flag = flag
		}

		param.IsArgument = p.IsArgument

		isFromDefault := true
		for _, l := range v.Log {
			if l.Source != fields.SourceDefaults {
				isFromDefault = false
				break
			}
		}
		if !isFromDefault || !p.IsEqualToDefault(v.Value) {
			ret = append(ret, param)
		}
	})

	return ret
}

// NewProgramFromCapture is a helper function to help capture a cliopatra Program from
// the description and the parsed layers a glazed command.
//
// It will go over all the ParameterDefinition (from all layers, which now also include the default layers).
// and will try to create the best cliopatra map it can. It tries to resolve the prefixes
// of layered fields.
//
// Values in the parameter map that are not present under the form of a ParameterDefinition
// will not be added to the command, and should be added separately using the WithRawFlags
// option.
func NewProgramFromCapture(
	description *cmds.CommandDescription,
	parsedLayers *values.Values,
	opts ...ProgramOption,
) *Program {
	ret := &Program{
		Name:        description.Name,
		Description: description.Short,
	}

	description.Layers.ForEach(func(_ string, layer schema.Section) {
		if layer.GetSlug() == "glazed-command" {
			// skip the meta glazed command flags, which also contain the create-cliopatra flag
			return
		}
		parsedLayer, ok := parsedLayers.Get(layer.GetSlug())
		if !ok {
			return
		}

		// TODO(manuel, 2023-03-21) This is broken I think, there's no need to use the prefix here
		parameters_ := getCliopatraParameters(
			layer.GetDefinitions(),
			parsedLayer.Fields,
			layer.GetPrefix())
		flags := []*Parameter{}
		arguments := []*Parameter{}

		for _, p := range parameters_ {
			if p.IsArgument {
				arguments = append(arguments, p)
			} else {
				flags = append(flags, p)
			}
		}
		ret.Flags = append(ret.Flags, flags...)
		ret.Args = append(ret.Args, arguments...)
	})

	for _, opt := range opts {
		opt(ret)
	}

	return ret
}
