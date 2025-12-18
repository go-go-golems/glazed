package fields

import (
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// Definition is a type alias for parameters.ParameterDefinition.
// A Definition specifies a single field in a schema section (name, type, default, help, etc.).
type Definition = parameters.ParameterDefinition

// Definitions is a type alias for parameters.ParameterDefinitions.
// Definitions is an ordered collection of field definitions.
type Definitions = parameters.ParameterDefinitions

// Type is a type alias for parameters.ParameterType.
// Type represents the type of a field (string, int, bool, etc.).
type Type = parameters.ParameterType

// Option is a type alias for parameters.ParameterDefinitionOption.
// Option configures a Definition during construction.
type Option = parameters.ParameterDefinitionOption

// New creates a new field definition with the given name and type.
// It wraps parameters.NewParameterDefinition.
func New(name string, t Type, options ...Option) *Definition {
	return parameters.NewParameterDefinition(name, t, options...)
}

// NewDefinitions creates a new collection of field definitions.
// It wraps parameters.NewParameterDefinitions.
func NewDefinitions(options ...parameters.ParameterDefinitionsOption) *Definitions {
	return parameters.NewParameterDefinitions(options...)
}

// Re-export common field definition options
var (
	WithHelp       = parameters.WithHelp
	WithShortFlag  = parameters.WithShortFlag
	WithDefault    = parameters.WithDefault
	WithChoices    = parameters.WithChoices
	WithRequired   = parameters.WithRequired
	WithIsArgument = parameters.WithIsArgument
)

// Re-export common field types
const (
	TypeString = parameters.ParameterTypeString
	TypeSecret = parameters.ParameterTypeSecret

	TypeStringFromFile  = parameters.ParameterTypeStringFromFile
	TypeStringFromFiles = parameters.ParameterTypeStringFromFiles

	TypeFile     = parameters.ParameterTypeFile
	TypeFileList = parameters.ParameterTypeFileList

	TypeObjectListFromFile  = parameters.ParameterTypeObjectListFromFile
	TypeObjectListFromFiles = parameters.ParameterTypeObjectListFromFiles
	TypeObjectFromFile      = parameters.ParameterTypeObjectFromFile
	TypeStringListFromFile  = parameters.ParameterTypeStringListFromFile
	TypeStringListFromFiles = parameters.ParameterTypeStringListFromFiles

	TypeKeyValue = parameters.ParameterTypeKeyValue

	TypeInteger     = parameters.ParameterTypeInteger
	TypeFloat       = parameters.ParameterTypeFloat
	TypeBool        = parameters.ParameterTypeBool
	TypeDate        = parameters.ParameterTypeDate
	TypeStringList  = parameters.ParameterTypeStringList
	TypeIntegerList = parameters.ParameterTypeIntegerList
	TypeFloatList   = parameters.ParameterTypeFloatList
	TypeChoice      = parameters.ParameterTypeChoice
	TypeChoiceList  = parameters.ParameterTypeChoiceList
)
