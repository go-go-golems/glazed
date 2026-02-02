package patternmapper

import (
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

// ConfigMapperBuilder provides a fluent API to assemble MappingRules and
// build a ConfigMapper that reuses NewConfigMapper validation and semantics.
type ConfigMapperBuilder struct {
	layers *schema.Schema
	rules  []MappingRule
}

// NewConfigMapperBuilder creates a new builder for pattern-based mappers.
func NewConfigMapperBuilder(l *schema.Schema) *ConfigMapperBuilder {
	return &ConfigMapperBuilder{
		layers: l,
		rules:  make([]MappingRule, 0, 8),
	}
}

// Map adds a simple leaf mapping rule.
// If required is provided and true, the rule will be marked as Required.
func (b *ConfigMapperBuilder) Map(source string, targetLayer string, targetParameter string, required ...bool) *ConfigMapperBuilder {
	r := MappingRule{
		Source:          source,
		TargetLayer:     targetLayer,
		TargetParameter: targetParameter,
	}
	if len(required) > 0 && required[0] {
		r.Required = true
	}
	b.rules = append(b.rules, r)
	return b
}

// MapObject adds a parent rule with child rules (one-level nesting supported).
func (b *ConfigMapperBuilder) MapObject(parentSource string, targetLayer string, childRules []MappingRule) *ConfigMapperBuilder {
	r := MappingRule{
		Source:      parentSource,
		TargetLayer: targetLayer,
		Rules:       childRules,
	}
	b.rules = append(b.rules, r)
	return b
}

// Build validates rules via NewConfigMapper and returns the resulting ConfigMapper.
func (b *ConfigMapperBuilder) Build() (middlewares.ConfigMapper, error) {
	return NewConfigMapper(b.layers, b.rules...)
}

// Child is a small helper to create a leaf MappingRule for MapObject children.
func Child(source string, targetParameter string) MappingRule {
	return MappingRule{Source: source, TargetParameter: targetParameter}
}
