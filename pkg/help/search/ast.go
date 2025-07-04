package search

import (
	"fmt"
	"strings"
)

// QueryNode represents a node in the query AST
type QueryNode interface {
	String() string
}

// TermNode represents a terminal node (a single filter or search term)
type TermNode interface {
	QueryNode
	IsNegated() bool
}

// FilterNode represents a field:value filter
type FilterNode struct {
	Field    string
	Value    string
	Negated  bool
	Operator string // ":" for exact match, "~" for contains, etc.
}

func (f *FilterNode) String() string {
	neg := ""
	if f.Negated {
		neg = "-"
	}
	return fmt.Sprintf("%s%s%s%s", neg, f.Field, f.Operator, f.Value)
}

func (f *FilterNode) IsNegated() bool {
	return f.Negated
}

// TextNode represents a text search term
type TextNode struct {
	Text     string
	Negated  bool
	Quoted   bool // Whether the text was quoted
}

func (t *TextNode) String() string {
	neg := ""
	if t.Negated {
		neg = "-"
	}
	if t.Quoted {
		return fmt.Sprintf("%s\"%s\"", neg, t.Text)
	}
	return fmt.Sprintf("%s%s", neg, t.Text)
}

func (t *TextNode) IsNegated() bool {
	return t.Negated
}

// BooleanNode represents a boolean operation (AND, OR, NOT)
type BooleanNode struct {
	Operator string      // "AND", "OR", "NOT"
	Left     QueryNode
	Right    QueryNode
}

func (b *BooleanNode) String() string {
	if b.Operator == "NOT" {
		return fmt.Sprintf("NOT %s", b.Left.String())
	}
	return fmt.Sprintf("(%s %s %s)", b.Left.String(), b.Operator, b.Right.String())
}

// GroupNode represents a parenthesized group
type GroupNode struct {
	Child QueryNode
}

func (g *GroupNode) String() string {
	return fmt.Sprintf("(%s)", g.Child.String())
}

// Query represents a complete parsed query
type Query struct {
	Root QueryNode
	Raw  string
}

func (q *Query) String() string {
	if q.Root == nil {
		return ""
	}
	return q.Root.String()
}

// Helper functions for building the AST
func NewFilterNode(field, value string, negated bool) *FilterNode {
	return &FilterNode{
		Field:    field,
		Value:    value,
		Negated:  negated,
		Operator: ":",
	}
}

func NewTextNode(text string, negated, quoted bool) *TextNode {
	return &TextNode{
		Text:    text,
		Negated: negated,
		Quoted:  quoted,
	}
}

func NewBooleanNode(operator string, left, right QueryNode) *BooleanNode {
	return &BooleanNode{
		Operator: operator,
		Left:     left,
		Right:    right,
	}
}

func NewGroupNode(child QueryNode) *GroupNode {
	return &GroupNode{
		Child: child,
	}
}

// Convenience functions for common patterns
func IsAndNode(node QueryNode) bool {
	if bn, ok := node.(*BooleanNode); ok {
		return bn.Operator == "AND"
	}
	return false
}

func IsOrNode(node QueryNode) bool {
	if bn, ok := node.(*BooleanNode); ok {
		return bn.Operator == "OR"
	}
	return false
}

func IsNotNode(node QueryNode) bool {
	if bn, ok := node.(*BooleanNode); ok {
		return bn.Operator == "NOT"
	}
	return false
}

// WalkAST traverses the AST depth-first, calling the visitor function for each node
func WalkAST(node QueryNode, visitor func(QueryNode)) {
	if node == nil {
		return
	}
	
	visitor(node)
	
	switch n := node.(type) {
	case *BooleanNode:
		WalkAST(n.Left, visitor)
		WalkAST(n.Right, visitor)
	case *GroupNode:
		WalkAST(n.Child, visitor)
	}
}

// CollectTerms collects all terminal nodes (filters and text) from the AST
func CollectTerms(node QueryNode) []TermNode {
	var terms []TermNode
	
	WalkAST(node, func(n QueryNode) {
		if term, ok := n.(TermNode); ok {
			terms = append(terms, term)
		}
	})
	
	return terms
}

// CollectFilters collects all filter nodes from the AST
func CollectFilters(node QueryNode) []*FilterNode {
	var filters []*FilterNode
	
	WalkAST(node, func(n QueryNode) {
		if filter, ok := n.(*FilterNode); ok {
			filters = append(filters, filter)
		}
	})
	
	return filters
}

// CollectTextSearches collects all text search nodes from the AST
func CollectTextSearches(node QueryNode) []*TextNode {
	var texts []*TextNode
	
	WalkAST(node, func(n QueryNode) {
		if text, ok := n.(*TextNode); ok {
			texts = append(texts, text)
		}
	})
	
	return texts
}

// SimplifyAST removes unnecessary group nodes and normalizes the tree
func SimplifyAST(node QueryNode) QueryNode {
	if node == nil {
		return nil
	}
	
	switch n := node.(type) {
	case *GroupNode:
		// Remove unnecessary grouping
		simplified := SimplifyAST(n.Child)
		return simplified
	case *BooleanNode:
		left := SimplifyAST(n.Left)
		right := SimplifyAST(n.Right)
		
		if n.Operator == "NOT" && right == nil {
			return &BooleanNode{
				Operator: "NOT",
				Left:     left,
				Right:    nil,
			}
		}
		
		return &BooleanNode{
			Operator: n.Operator,
			Left:     left,
			Right:    right,
		}
	default:
		return node
	}
}

// GetAllFields returns all unique field names used in filter nodes
func GetAllFields(node QueryNode) []string {
	fieldMap := make(map[string]bool)
	
	WalkAST(node, func(n QueryNode) {
		if filter, ok := n.(*FilterNode); ok {
			fieldMap[filter.Field] = true
		}
	})
	
	var fields []string
	for field := range fieldMap {
		fields = append(fields, field)
	}
	
	return fields
}

// ContainsField checks if the AST contains a filter for the given field
func ContainsField(node QueryNode, field string) bool {
	found := false
	WalkAST(node, func(n QueryNode) {
		if filter, ok := n.(*FilterNode); ok && filter.Field == field {
			found = true
		}
	})
	return found
}

// GetFieldValues returns all values for filters with the given field name
func GetFieldValues(node QueryNode, field string) []string {
	var values []string
	
	WalkAST(node, func(n QueryNode) {
		if filter, ok := n.(*FilterNode); ok && filter.Field == field {
			values = append(values, filter.Value)
		}
	})
	
	return values
}

// HasTextSearch checks if the AST contains any text search terms
func HasTextSearch(node QueryNode) bool {
	hasText := false
	WalkAST(node, func(n QueryNode) {
		if _, ok := n.(*TextNode); ok {
			hasText = true
		}
	})
	return hasText
}

// GetTextSearchTerms returns all text search terms combined
func GetTextSearchTerms(node QueryNode) []string {
	var terms []string
	
	WalkAST(node, func(n QueryNode) {
		if text, ok := n.(*TextNode); ok {
			terms = append(terms, text.Text)
		}
	})
	
	return terms
}

// CombineTextSearchTerms combines all text search terms into a single query string
func CombineTextSearchTerms(node QueryNode) string {
	terms := GetTextSearchTerms(node)
	if len(terms) == 0 {
		return ""
	}
	
	// Join terms with AND for FTS
	return strings.Join(terms, " AND ")
}
