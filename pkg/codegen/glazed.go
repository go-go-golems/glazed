package codegen

import (
	"reflect"

	"github.com/dave/jennifer/jen"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const GlazedCommandsPath = "github.com/go-go-golems/glazed/pkg/cmds"
const GlazedMiddlewaresPath = "github.com/go-go-golems/glazed/pkg/middlewares"
const GlazedParametersPath = "github.com/go-go-golems/glazed/pkg/cmds/parameters"
const ClaySqlPath = "github.com/go-go-golems/clay/pkg/sql"
const MapsHelpersPath = "github.com/go-go-golems/glazed/pkg/helpers/maps"

func ParameterDefinitionToDict(p *parameters.ParameterDefinition) (jen.Code, error) {
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

func FlagValueToJen(p *parameters.ParameterDefinition) (jen.Code, error) {
	d, err := p.CheckParameterDefaultValueValidity()
	if err != nil {
		return nil, err
	}

	return LiteralToJen(reflect.ValueOf(d))
}

func FlagTypeToGoType(s *jen.Statement, parameterType parameters.ParameterType) *jen.Statement {
	switch parameterType {
	case parameters.ParameterTypeFloat:
		return s.Id("float64")
	case parameters.ParameterTypeFloatList:
		return s.Index().Id("float64")
	case parameters.ParameterTypeInteger:
		return s.Id("int")
	case parameters.ParameterTypeIntegerList:
		return s.Index().Id("int")
	case parameters.ParameterTypeBool:
		return s.Id("bool")
	case parameters.ParameterTypeDate:
		return s.Qual("time", "Time")
	case parameters.ParameterTypeStringFromFile,
		parameters.ParameterTypeStringFromFiles,
		parameters.ParameterTypeChoice,
		parameters.ParameterTypeString,
		parameters.ParameterTypeSecret:
		return s.Id("string")
	case parameters.ParameterTypeStringList,
		parameters.ParameterTypeStringListFromFile,
		parameters.ParameterTypeStringListFromFiles,
		parameters.ParameterTypeChoiceList:
		return s.Index().Id("string")
	case parameters.ParameterTypeFile:
		return s.Qual(GlazedParametersPath, "FileData")
	case parameters.ParameterTypeFileList:
		return s.Index().Qual(GlazedParametersPath, "FileData")
	case parameters.ParameterTypeObjectFromFile:
		return s.Map(jen.Id("string")).Id("interface{}")
	case parameters.ParameterTypeObjectListFromFile, parameters.ParameterTypeObjectListFromFiles:
		return s.Index().Map(jen.Id("string")).Id("interface{}")
	case parameters.ParameterTypeKeyValue:
		return s.Map(jen.Id("string")).Id("string")
	default:
		return s.Id(string(parameterType))
	}
}
