package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
type ParameterLayer interface {
	AddFlag(flag *parameters.ParameterDefinition)
	GetParameterDefinitions() map[string]*parameters.ParameterDefinition

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string
}

// ParsedParameterLayer is the result of "parsing" input data using a ParameterLayer
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query parameters.
type ParsedParameterLayer struct {
	Layer      ParameterLayer
	Parameters map[string]interface{}
}

type JSONParameterLayer interface {
	ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (map[string]interface{}, error)
}

// Clone returns a copy of the parsedParameterLayer with a fresh Parameters map.
// However, neither the Layer nor the Parameters are deep copied.
func (ppl *ParsedParameterLayer) Clone() *ParsedParameterLayer {
	ret := &ParsedParameterLayer{
		Layer:      ppl.Layer,
		Parameters: make(map[string]interface{}),
	}
	for k, v := range ppl.Parameters {
		ret.Parameters[k] = v
	}
	return ret
}

// MergeParameters merges the other ParsedParameterLayer into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *ParsedParameterLayer) MergeParameters(other *ParsedParameterLayer) {
	for k, v := range other.Parameters {
		ppl.Parameters[k] = v
	}
}

// TODO(manuel, 2023-02-27) Might be worth making a struct defaults middleware
