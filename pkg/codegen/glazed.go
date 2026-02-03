package codegen

import (
	"reflect"

	"github.com/dave/jennifer/jen"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
)

const GlazedCommandsPath = "github.com/go-go-golems/glazed/pkg/cmds"
const GlazedMiddlewaresPath = "github.com/go-go-golems/glazed/pkg/middlewares"
const GlazedFieldsPath = "github.com/go-go-golems/glazed/pkg/cmds/fields"
const ClaySqlPath = "github.com/go-go-golems/clay/pkg/sql"
const MapsHelpersPath = "github.com/go-go-golems/glazed/pkg/helpers/maps"

func ParameterDefinitionToDict(p *fields.Definition) (jen.Code, error) {
	ret := jen.Dict{
		jen.Id("Name"): jen.Lit(p.Name),
		jen.Id("Type"): jen.Lit(string(p.Type)),
		jen.Id("Help"): jen.Lit(p.Help),
	}

	if p.Default != nil {
		var err error
		v, err := FlagValueToJen(p)
		if err != nil {
			return nil, err
		}
		ret[jen.Id("Default")] = jen.Qual("github.com/go-go-golems/glazed/pkg/helpers/cast", "InterfaceAddr").Call(v)
	}
	if p.Choices != nil {
		ret[jen.Id("Choices")] = jen.Index().String().ValuesFunc(func(g *jen.Group) {
			for _, c := range p.Choices {
				g.Lit(c)
			}
		})
	}

	return ret, nil
}

func FlagValueToJen(p *fields.Definition) (jen.Code, error) {
	d, err := p.CheckDefaultValueValidity()
	if err != nil {
		return nil, err
	}

	return LiteralToJen(reflect.ValueOf(d))
}

func FlagTypeToGoType(s *jen.Statement, parameterType fields.Type) *jen.Statement {
	switch parameterType {
	case fields.TypeFloat:
		return s.Id("float64")
	case fields.TypeFloatList:
		return s.Index().Id("float64")
	case fields.TypeInteger:
		return s.Id("int")
	case fields.TypeIntegerList:
		return s.Index().Id("int")
	case fields.TypeBool:
		return s.Id("bool")
	case fields.TypeDate:
		return s.Qual("time", "Time")
	case fields.TypeStringFromFile,
		fields.TypeStringFromFiles,
		fields.TypeChoice,
		fields.TypeString,
		fields.TypeSecret:
		return s.Id("string")
	case fields.TypeStringList,
		fields.TypeStringListFromFile,
		fields.TypeStringListFromFiles,
		fields.TypeChoiceList:
		return s.Index().Id("string")
	case fields.TypeFile:
		return s.Qual(GlazedFieldsPath, "FileData")
	case fields.TypeFileList:
		return s.Index().Qual(GlazedFieldsPath, "FileData")
	case fields.TypeObjectFromFile:
		return s.Map(jen.Id("string")).Id("interface{}")
	case fields.TypeObjectListFromFile, fields.TypeObjectListFromFiles:
		return s.Index().Map(jen.Id("string")).Id("interface{}")
	case fields.TypeKeyValue:
		return s.Map(jen.Id("string")).Id("string")
	default:
		return s.Id(string(parameterType))
	}
}
