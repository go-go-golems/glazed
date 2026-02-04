package patternmapper

import (
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	sources "github.com/go-go-golems/glazed/pkg/cmds/sources"
)

// ConfigMapperBuilder provides a fluent API to assemble MappingRules and
// build a ConfigMapper that reuses NewConfigMapper validation and semantics.
type ConfigMapperBuilder struct {
	sectionSchema *schema.Schema
	rules         []MappingRule
}

// NewConfigMapperBuilder creates a new builder for pattern-based mappers.
func NewConfigMapperBuilder(sectionSchema *schema.Schema) *ConfigMapperBuilder {
	return &ConfigMapperBuilder{
		sectionSchema: sectionSchema,
		rules:         make([]MappingRule, 0, 8),
	}
}

// Map adds a simple leaf mapping rule.
// If required is provided and true, the rule will be marked as Required.
func (b *ConfigMapperBuilder) Map(source string, targetSection string, targetField string, required ...bool) *ConfigMapperBuilder {
	r := MappingRule{
		Source:        source,
		TargetSection: targetSection,
		TargetField:   targetField,
	}
	if len(required) > 0 && required[0] {
		r.Required = true
	}
	b.rules = append(b.rules, r)
	return b
}

// MapObject adds a parent rule with child rules (one-level nesting supported).
func (b *ConfigMapperBuilder) MapObject(parentSource string, targetSection string, childRules []MappingRule) *ConfigMapperBuilder {
	r := MappingRule{
		Source:        parentSource,
		TargetSection: targetSection,
		Rules:         childRules,
	}
	b.rules = append(b.rules, r)
	return b
}

// Build validates rules via NewConfigMapper and returns the resulting ConfigMapper.
func (b *ConfigMapperBuilder) Build() (sources.ConfigMapper, error) {
	return NewConfigMapper(b.sectionSchema, b.rules...)
}

// Child is a small helper to create a leaf MappingRule for MapObject children.
func Child(source string, targetField string) MappingRule {
	return MappingRule{Source: source, TargetField: targetField}
}
