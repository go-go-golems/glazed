package search

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/model"
	"github.com/go-go-golems/glazed/pkg/help/query"
)

// Converter converts parsed query AST to predicate functions
type Converter struct {
	// Configuration for handling different field types
	TypeMapping map[string]model.SectionType
}

// NewConverter creates a new converter with default configuration
func NewConverter() *Converter {
	return &Converter{
		TypeMapping: map[string]model.SectionType{
			"topic":       model.SectionGeneralTopic,
			"generaltopic": model.SectionGeneralTopic,
			"example":     model.SectionExample,
			"application": model.SectionApplication,
			"app":         model.SectionApplication,
			"tutorial":    model.SectionTutorial,
			"tut":         model.SectionTutorial,
		},
	}
}

// Convert converts a query AST to a predicate function
func (c *Converter) Convert(q *Query) (query.Predicate, error) {
	if q.Root == nil {
		// Empty query - return a predicate that matches everything
		return func(comp interface{}) {}, nil
	}
	
	return c.convertNode(q.Root)
}

// convertNode converts a single AST node to a predicate
func (c *Converter) convertNode(node QueryNode) (query.Predicate, error) {
	switch n := node.(type) {
	case *FilterNode:
		return c.convertFilter(n)
	case *TextNode:
		return c.convertText(n)
	case *BooleanNode:
		return c.convertBoolean(n)
	case *GroupNode:
		return c.convertNode(n.Child)
	default:
		return nil, fmt.Errorf("unsupported node type: %T", node)
	}
}

// convertFilter converts a filter node to a predicate
func (c *Converter) convertFilter(filter *FilterNode) (query.Predicate, error) {
	field := strings.ToLower(filter.Field)
	value := filter.Value
	
	var pred query.Predicate
	var err error
	
	switch field {
	case "type":
		pred, err = c.convertTypeFilter(value)
	case "topic":
		pred = query.HasTopic(value)
	case "flag":
		pred = query.HasFlag(value)
	case "command":
		pred = query.HasCommand(value)
	case "toplevel":
		pred, err = c.convertBooleanFilter(value, query.IsTopLevel)
	case "default":
		pred, err = c.convertBooleanFilter(value, query.ShownByDefault)
	case "slug":
		pred = query.SlugEquals(value)
	case "title":
		pred = query.TextSearch(value)
	case "content":
		pred = query.TextSearch(value)
	default:
		return nil, fmt.Errorf("unsupported filter field: %s", field)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Apply negation if needed
	if filter.Negated {
		pred = query.Not(pred)
	}
	
	return pred, nil
}

// convertTypeFilter converts a type filter to a predicate
func (c *Converter) convertTypeFilter(value string) (query.Predicate, error) {
	normalizedValue := strings.ToLower(value)
	
	// Check direct mapping first
	if sectionType, exists := c.TypeMapping[normalizedValue]; exists {
		return query.IsType(sectionType), nil
	}
	
	// Check exact matches
	switch normalizedValue {
	case "topic", "generaltopic":
		return query.IsType(model.SectionGeneralTopic), nil
	case "example":
		return query.IsType(model.SectionExample), nil
	case "application", "app":
		return query.IsType(model.SectionApplication), nil
	case "tutorial", "tut":
		return query.IsType(model.SectionTutorial), nil
	default:
		return nil, fmt.Errorf("unsupported section type: %s", value)
	}
}

// convertBooleanFilter converts a boolean filter value to a predicate
func (c *Converter) convertBooleanFilter(value string, truePred func() query.Predicate) (query.Predicate, error) {
	normalizedValue := strings.ToLower(value)
	
	switch normalizedValue {
	case "true", "yes", "1", "on":
		return truePred(), nil
	case "false", "no", "0", "off":
		return query.Not(truePred()), nil
	default:
		return nil, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// convertText converts a text node to a predicate
func (c *Converter) convertText(text *TextNode) (query.Predicate, error) {
	pred := query.TextSearch(text.Text)
	
	if text.Negated {
		pred = query.Not(pred)
	}
	
	return pred, nil
}

// convertBoolean converts a boolean node to a predicate
func (c *Converter) convertBoolean(boolean *BooleanNode) (query.Predicate, error) {
	switch boolean.Operator {
	case "AND":
		return c.convertAnd(boolean)
	case "OR":
		return c.convertOr(boolean)
	case "NOT":
		return c.convertNot(boolean)
	default:
		return nil, fmt.Errorf("unsupported boolean operator: %s", boolean.Operator)
	}
}

// convertAnd converts an AND node to a predicate
func (c *Converter) convertAnd(and *BooleanNode) (query.Predicate, error) {
	left, err := c.convertNode(and.Left)
	if err != nil {
		return nil, err
	}
	
	right, err := c.convertNode(and.Right)
	if err != nil {
		return nil, err
	}
	
	return query.And(left, right), nil
}

// convertOr converts an OR node to a predicate
func (c *Converter) convertOr(or *BooleanNode) (query.Predicate, error) {
	left, err := c.convertNode(or.Left)
	if err != nil {
		return nil, err
	}
	
	right, err := c.convertNode(or.Right)
	if err != nil {
		return nil, err
	}
	
	return query.Or(left, right), nil
}

// convertNot converts a NOT node to a predicate
func (c *Converter) convertNot(not *BooleanNode) (query.Predicate, error) {
	pred, err := c.convertNode(not.Left)
	if err != nil {
		return nil, err
	}
	
	return query.Not(pred), nil
}

// ConvertQuery is a convenience function that converts a query string to a predicate
func ConvertQuery(queryStr string) (query.Predicate, error) {
	q, err := ParseQuery(queryStr)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	
	err = ValidateQuery(q)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	converter := NewConverter()
	pred, err := converter.Convert(q)
	if err != nil {
		return nil, fmt.Errorf("conversion error: %w", err)
	}
	
	return pred, nil
}

// QueryBuilder provides a fluent interface for building queries
type QueryBuilder struct {
	predicates []query.Predicate
	operator   string // "AND" or "OR"
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		operator: "AND",
	}
}

// WithType adds a type filter
func (qb *QueryBuilder) WithType(sectionType model.SectionType) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.IsType(sectionType))
	return qb
}

// WithTopic adds a topic filter
func (qb *QueryBuilder) WithTopic(topic string) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.HasTopic(topic))
	return qb
}

// WithFlag adds a flag filter
func (qb *QueryBuilder) WithFlag(flag string) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.HasFlag(flag))
	return qb
}

// WithCommand adds a command filter
func (qb *QueryBuilder) WithCommand(command string) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.HasCommand(command))
	return qb
}

// WithTopLevel adds a top-level filter
func (qb *QueryBuilder) WithTopLevel(topLevel bool) *QueryBuilder {
	if topLevel {
		qb.predicates = append(qb.predicates, query.IsTopLevel())
	} else {
		qb.predicates = append(qb.predicates, query.Not(query.IsTopLevel()))
	}
	return qb
}

// WithDefault adds a default filter
func (qb *QueryBuilder) WithDefault(showDefault bool) *QueryBuilder {
	if showDefault {
		qb.predicates = append(qb.predicates, query.ShownByDefault())
	} else {
		qb.predicates = append(qb.predicates, query.Not(query.ShownByDefault()))
	}
	return qb
}

// WithSlug adds a slug filter
func (qb *QueryBuilder) WithSlug(slug string) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.SlugEquals(slug))
	return qb
}

// WithTextSearch adds a text search filter
func (qb *QueryBuilder) WithTextSearch(text string) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.TextSearch(text))
	return qb
}

// WithNot adds a negated predicate
func (qb *QueryBuilder) WithNot(pred query.Predicate) *QueryBuilder {
	qb.predicates = append(qb.predicates, query.Not(pred))
	return qb
}

// WithPredicate adds a custom predicate
func (qb *QueryBuilder) WithPredicate(pred query.Predicate) *QueryBuilder {
	qb.predicates = append(qb.predicates, pred)
	return qb
}

// UseOr sets the operator to OR (default is AND)
func (qb *QueryBuilder) UseOr() *QueryBuilder {
	qb.operator = "OR"
	return qb
}

// UseAnd sets the operator to AND
func (qb *QueryBuilder) UseAnd() *QueryBuilder {
	qb.operator = "AND"
	return qb
}

// Build builds the final predicate
func (qb *QueryBuilder) Build() query.Predicate {
	if len(qb.predicates) == 0 {
		return func(comp interface{}) {} // Empty predicate
	}
	
	if len(qb.predicates) == 1 {
		return qb.predicates[0]
	}
	
	if qb.operator == "OR" {
		return query.Or(qb.predicates...)
	}
	
	return query.And(qb.predicates...)
}

// QueryOptimizer optimizes queries for better performance
type QueryOptimizer struct {
	// Configuration for optimization
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{}
}

// Optimize optimizes a query AST for better performance
func (o *QueryOptimizer) Optimize(q *Query) *Query {
	if q.Root == nil {
		return q
	}
	
	optimized := o.optimizeNode(q.Root)
	return &Query{
		Root: optimized,
		Raw:  q.Raw,
	}
}

// optimizeNode optimizes a single AST node
func (o *QueryOptimizer) optimizeNode(node QueryNode) QueryNode {
	switch n := node.(type) {
	case *BooleanNode:
		return o.optimizeBooleanNode(n)
	case *GroupNode:
		child := o.optimizeNode(n.Child)
		// Remove unnecessary grouping
		if _, needsGroup := child.(*BooleanNode); !needsGroup {
			return child
		}
		return &GroupNode{Child: child}
	default:
		return node
	}
}

// optimizeBooleanNode optimizes boolean nodes
func (o *QueryOptimizer) optimizeBooleanNode(node *BooleanNode) QueryNode {
	left := o.optimizeNode(node.Left)
	right := o.optimizeNode(node.Right)
	
	// Handle NOT optimization
	if node.Operator == "NOT" {
		return &BooleanNode{
			Operator: "NOT",
			Left:     left,
			Right:    nil,
		}
	}
	
	// Flatten nested boolean operations of the same type
	if leftBoolean, ok := left.(*BooleanNode); ok && leftBoolean.Operator == node.Operator {
		if rightBoolean, ok := right.(*BooleanNode); ok && rightBoolean.Operator == node.Operator {
			// Both sides are the same operator - flatten
			return &BooleanNode{
				Operator: node.Operator,
				Left: &BooleanNode{
					Operator: node.Operator,
					Left:     leftBoolean.Left,
					Right:    leftBoolean.Right,
				},
				Right: &BooleanNode{
					Operator: node.Operator,
					Left:     rightBoolean.Left,
					Right:    rightBoolean.Right,
				},
			}
		}
	}
	
	return &BooleanNode{
		Operator: node.Operator,
		Left:     left,
		Right:    right,
	}
}

// GetSupportedFields returns a list of supported filter fields
func GetSupportedFields() []string {
	return []string{
		"type",
		"topic", 
		"flag",
		"command",
		"toplevel",
		"default",
		"slug",
		"title",
		"content",
	}
}

// GetSupportedTypes returns a list of supported section types
func GetSupportedTypes() []string {
	return []string{
		"topic",
		"generaltopic",
		"example",
		"application",
		"app",
		"tutorial",
		"tut",
	}
}

// GetFieldDescription returns a description of a filter field
func GetFieldDescription(field string) string {
	descriptions := map[string]string{
		"type":     "Section type (topic, example, application, tutorial)",
		"topic":    "Topic tags associated with the section",
		"flag":     "Command-line flags associated with the section",
		"command":  "Commands associated with the section",
		"toplevel": "Whether the section is shown at top level (true/false)",
		"default":  "Whether the section is shown by default (true/false)",
		"slug":     "Exact slug match for the section",
		"title":    "Text search in section title",
		"content":  "Text search in section content",
	}
	
	if desc, exists := descriptions[strings.ToLower(field)]; exists {
		return desc
	}
	
	return "Unknown field"
}
