package layers

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type ErrInvalidParameterLayer struct {
	Name     string
	Expected string
}

func (e ErrInvalidParameterLayer) Error() string {
	if e.Expected == "" {
		return fmt.Sprintf("invalid parameter layer: %s", e.Name)
	}
	return fmt.Sprintf("invalid parameter layer: %s (expected %s)", e.Name, e.Expected)
}

// ParameterLayer is a struct that is used by one specific functionality layer
// to group and describe all the parameter definitions that it uses.
// It also provides a location for a name, slug and description to be used in help
// pages.
//
// TODO(manuel, 2023-12-20) This is a pretty messy interface, I think it used to be a struct?
type ParameterLayer interface {
	AddFlags(flag ...*parameters.ParameterDefinition)
	GetParameterDefinitions() parameters.ParameterDefinitions

	InitializeParameterDefaultsFromStruct(s interface{}) error

	GetName() string
	GetSlug() string
	GetDescription() string
	GetPrefix() string

	Clone() ParameterLayer
}

// ParsedParameterLayer is the result of "parsing" input data using a ParameterLayer
// specification. For example, it could be the result of parsing cobra command flags,
// or a JSON body, or HTTP query parameters.
type ParsedParameterLayer struct {
	Layer      ParameterLayer
	Parameters *parameters.ParsedParameters
}

type JSONParameterLayer interface {
	ParseFlagsFromJSON(m map[string]interface{}, onlyProvided bool) (*parameters.ParsedParameters, error)
}

// Clone returns a copy of the parsedParameterLayer with a fresh Parameters map.
// However, neither the Layer nor the Parameters are deep copied.
func (ppl *ParsedParameterLayer) Clone() *ParsedParameterLayer {
	ret := &ParsedParameterLayer{
		Layer:      ppl.Layer,
		Parameters: parameters.NewParsedParameters().Merge(ppl.Parameters),
	}
	ppl.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
		ret.Parameters.Set(k, v)
	})
	return ret
}

// MergeParameters merges the other ParsedParameterLayer into this one, overwriting any
// existing values. This doesn't replace the actual Layer pointer.
func (ppl *ParsedParameterLayer) MergeParameters(other *ParsedParameterLayer) {
	ppl.Parameters.Merge(other.Parameters)
}

func GetAllParsedParameters(layers map[string]*ParsedParameterLayer) map[string]interface{} {
	ret := make(map[string]interface{})
	for _, l := range layers {
		l.Parameters.ForEach(func(k string, v *parameters.ParsedParameter) {
			ret[k] = v
		})
	}

	return ret
}
