package schema

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/spf13/cobra"
)

// Section is a type alias for layers.ParameterLayer.
// A Section represents a named group of field definitions (schema section).
type Section = layers.ParameterLayer

// Schema is a type alias for layers.ParameterLayers.
// Schema is an ordered collection of schema sections.
type Schema = layers.ParameterLayers

// FlagGroupUsage is a type alias for layers.FlagGroupUsage.
// FlagGroupUsage describes a formatted cobra flag group.
type FlagGroupUsage = layers.FlagGroupUsage

// CommandFlagGroupUsage is a type alias for layers.CommandFlagGroupUsage.
// CommandFlagGroupUsage aggregates flag group usages for a cobra command.
type CommandFlagGroupUsage = layers.CommandFlagGroupUsage

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

// NewSectionFromYAML creates a new schema section from YAML definitions.
// It wraps layers.NewParameterLayerFromYAML.
func NewSectionFromYAML(s []byte, options ...SectionOption) (*SectionImpl, error) {
	return layers.NewParameterLayerFromYAML(s, options...)
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

// ComputeCommandFlagGroupUsage computes flag group usage information for a cobra command.
// It wraps layers.ComputeCommandFlagGroupUsage.
func ComputeCommandFlagGroupUsage(c *cobra.Command) *CommandFlagGroupUsage {
	return layers.ComputeCommandFlagGroupUsage(c)
}

// Re-export common section options for convenience
var (
	WithPrefix      = layers.WithPrefix
	WithName        = layers.WithName
	WithDescription = layers.WithDescription
	WithDefaults    = layers.WithDefaults
	// WithFields attaches field definitions to a section.
	// It is a clearer alias for the historical layers.WithParameterDefinitions.
	WithFields    = layers.WithParameterDefinitions
	WithArguments = layers.WithArguments
)
