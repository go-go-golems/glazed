package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
//
// TODO(manuel, 2023-12-20) This is a pretty messy interface, I think it used to be a struct?
type ParameterLayer interface {
	AddFlags(flag ...*parameters.ParameterDefinition)
	GetParameterDefinitions() *parameters.ParameterDefinitions

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string

	Clone() ParameterLayer
}

const DefaultSlug = "default"

type JSONParameterLayer interface {
	ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (*parameters.ParsedParameters, error)
}
