package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

// Section is a type alias for layers.ParameterLayer.
// A Section represents a named group of field definitions (schema section).
type Section = layers.ParameterLayer

// Sections is a type alias for layers.ParameterLayers.
// Sections is an ordered collection of schema sections.
type Sections = layers.ParameterLayers

// SectionImpl is a type alias for layers.ParameterLayerImpl.
// SectionImpl is the common concrete implementation of Section.
type SectionImpl = layers.ParameterLayerImpl

// SectionOption is a type alias for layers.ParameterLayerOptions.
// SectionOption configures a Section during construction.
type SectionOption = layers.ParameterLayerOptions

// SectionsOption is a type alias for layers.ParameterLayersOption.
// SectionsOption configures a Sections collection during construction.
type SectionsOption = layers.ParameterLayersOption

// DefaultSlug is the default slug used for the default schema section.
const DefaultSlug = layers.DefaultSlug

// NewSection creates a new schema section with the given slug and name.
// It wraps layers.NewParameterLayer.
func NewSection(slug string, name string, options ...SectionOption) (*SectionImpl, error) {
	return layers.NewParameterLayer(slug, name, options...)
}

// NewSections creates a new collection of schema sections.
// It wraps layers.NewParameterLayers.
func NewSections(options ...SectionsOption) *Sections {
	return layers.NewParameterLayers(options...)
}

// WithSections returns a SectionsOption that adds the given sections to a Sections collection.
// It wraps layers.WithLayers.
func WithSections(sections ...Section) SectionsOption {
	return layers.WithLayers(sections...)
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
