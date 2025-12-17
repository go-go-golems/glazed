package values

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
)

// SectionValues is a type alias for layers.ParsedLayer.
// SectionValues contains resolved values for a single schema section.
type SectionValues = layers.ParsedLayer

// Values is a type alias for layers.ParsedLayers.
// Values contains resolved values for all schema sections.
type Values = layers.ParsedLayers

// ValuesOption is a type alias for layers.ParsedLayersOption.
// ValuesOption configures a Values collection during construction.
type ValuesOption = layers.ParsedLayersOption

// New creates a new collection of resolved values.
// It wraps layers.NewParsedLayers.
func New(options ...ValuesOption) *Values {
	return layers.NewParsedLayers(options...)
}

// WithSectionValues returns a ValuesOption that adds a section's values to a Values collection.
// It wraps layers.WithParsedLayer.
func WithSectionValues(slug string, v *SectionValues) ValuesOption {
	return layers.WithParsedLayer(slug, v)
}

// DecodeInto decodes resolved values from a single section into the destination struct.
// It wraps v.InitializeStruct(dst).
func DecodeInto(v *SectionValues, dst any) error {
	return v.InitializeStruct(dst)
}

// DecodeSectionInto decodes resolved values from a specific section (by slug) into the destination struct.
// It wraps vs.InitializeStruct(sectionSlug, dst).
func DecodeSectionInto(vs *Values, sectionSlug string, dst any) error {
	return vs.InitializeStruct(sectionSlug, dst)
}

// AsMap returns a flat map of all resolved values across all sections.
// It wraps vs.GetDataMap().
func AsMap(vs *Values) map[string]any {
	return vs.GetDataMap()
}
