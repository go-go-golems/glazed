package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/settings"
)

// Section is a type alias for layers.ParameterLayer.
// A Section represents a named group of field definitions (schema section).
type Section = layers.ParameterLayer

// Schema is a type alias for layers.ParameterLayers.
// Schema is an ordered collection of schema sections.
type Schema = layers.ParameterLayers

// SectionImpl is a type alias for layers.ParameterLayerImpl.
// SectionImpl is the common concrete implementation of Section.
type SectionImpl = layers.ParameterLayerImpl

// SectionOption is a type alias for layers.ParameterLayerOptions.
// SectionOption configures a Section during construction.
type SectionOption = layers.ParameterLayerOptions

// SchemaOption is a type alias for layers.ParameterLayersOption.
// SchemaOption configures a Schema collection during construction.
type SchemaOption = layers.ParameterLayersOption

// DefaultSlug is the default slug used for the default schema section.
const DefaultSlug = layers.DefaultSlug

// NewSection creates a new schema section with the given slug and name.
// It wraps layers.NewParameterLayer.
func NewSection(slug string, name string, options ...SectionOption) (*SectionImpl, error) {
	return layers.NewParameterLayer(slug, name, options...)
}

// NewSchema creates a new collection of schema sections.
// It wraps layers.NewParameterLayers.
func NewSchema(options ...SchemaOption) *Schema {
	return layers.NewParameterLayers(options...)
}

// WithSections returns a SchemaOption that adds the given sections to a Schema collection.
// It wraps layers.WithLayers.
func WithSections(sections ...Section) SchemaOption {
	return layers.WithLayers(sections...)
}

// NewGlazedSchema creates a new glazed schema section containing all glazed output/formatting settings.
// It wraps settings.NewGlazedParameterLayers.
func NewGlazedSchema(options ...settings.GlazeParameterLayerOption) (Section, error) {
	return settings.NewGlazedParameterLayers(options...)
}

// Re-export common section options for convenience
var (
	WithPrefix               = layers.WithPrefix
	WithName                 = layers.WithName
	WithDescription          = layers.WithDescription
	WithDefaults             = layers.WithDefaults
	WithParameterDefinitions = layers.WithParameterDefinitions
	WithArguments            = layers.WithArguments
)
