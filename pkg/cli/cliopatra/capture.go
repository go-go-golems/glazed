package cliopatra

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
)

func getCliopatraFields(
	definitions *fields.Definitions,
	ps *fields.FieldValues,
	prefix string,
) []*Field {
	ret := []*Field{}

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
		param := &Field{
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
// the description and the resolved values for a glazed command.
//
// It will go over all the field definitions (from all sections, which now also include the default section),
// and will try to create the best cliopatra map it can. It tries to resolve the prefixes
// of sectioned fields.
//
// Values in the field map that are not present under the form of a field definition
// will not be added to the command, and should be added separately using the WithRawFlags
// option.
func NewProgramFromCapture(
	description *cmds.CommandDescription,
	parsedValues *values.Values,
	opts ...ProgramOption,
) *Program {
	ret := &Program{
		Name:        description.Name,
		Description: description.Short,
	}

	description.Schema.ForEach(func(_ string, section schema.Section) {
		if section.GetSlug() == "glazed-command" {
			// skip the meta glazed command flags, which also contain the create-cliopatra flag
			return
		}
		sectionValues, ok := parsedValues.Get(section.GetSlug())
		if !ok {
			return
		}

		// TODO(manuel, 2023-03-21) This is broken I think, there's no need to use the prefix here
		fields_ := getCliopatraFields(
			section.GetDefinitions(),
			sectionValues.Fields,
			section.GetPrefix())
		flags := []*Field{}
		arguments := []*Field{}

		for _, field := range fields_ {
			if field.IsArgument {
				arguments = append(arguments, field)
			} else {
				flags = append(flags, field)
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
