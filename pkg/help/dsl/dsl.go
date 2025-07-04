package dsl

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/store"
)

// ParseQuery parses a query string and returns a predicate
func ParseQuery(query string) (store.Predicate, error) {
	// Handle empty query
	if strings.TrimSpace(query) == "" {
		return func(qc *store.QueryCompiler) {
			// Empty predicate that matches all
		}, nil
	}

	// Parse the query into an AST
	expr, err := Parse(query)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Compile the AST to a predicate
	compiler := NewCompiler()
	predicate, err := compiler.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	return predicate, nil
}

// ValidateQuery validates a query string without executing it
func ValidateQuery(query string) error {
	_, err := ParseQuery(query)
	return err
}

// QueryInfo provides information about query syntax
type QueryInfo struct {
	ValidFields []string
	ValidTypes  []string
	Examples    []string
}

// GetQueryInfo returns information about the query DSL
func GetQueryInfo() *QueryInfo {
	compiler := NewCompiler()
	return &QueryInfo{
		ValidFields: compiler.GetValidFields(),
		ValidTypes:  compiler.GetValidTypeValues(),
		Examples: []string{
			"type:example",
			"topic:database",
			"flag:--output",
			"command:json",
			"\"full text search\"",
			"'single quoted text'",
			"unquoted text search",
			"database tutorial AND type:example",
			"type:tutorial OR type:example",
			"NOT type:application",
			"(database OR sql) AND type:tutorial",
			"toplevel:true AND default:true",
			"\"SQLite\" AND type:tutorial",
			"performance optimization",
		},
	}
}

// FormatError formats a parsing or compilation error with context
func FormatError(query string, err error) string {
	if err == nil {
		return ""
	}

	// Add the original query for context
	errorMsg := fmt.Sprintf("Query error: %s\n", err.Error())
	errorMsg += fmt.Sprintf("Query: %s\n", query)

	// Add helpful information
	info := GetQueryInfo()
	errorMsg += "\nValid fields: " + strings.Join(info.ValidFields, ", ") + "\n"
	errorMsg += "Valid types: " + strings.Join(info.ValidTypes, ", ") + "\n"
	errorMsg += "\nExample queries:\n"
	for _, example := range info.Examples {
		errorMsg += fmt.Sprintf("  %s\n", example)
	}

	return errorMsg
}

// SuggestCorrection suggests corrections for common mistakes
func SuggestCorrection(query string, err error) string {
	if err == nil {
		return ""
	}

	errStr := strings.ToLower(err.Error())

	// Suggest corrections based on common mistakes
	suggestions := []string{}

	if strings.Contains(errStr, "unknown field") {
		suggestions = append(suggestions, "Check field name spelling. Valid fields: type, topic, flag, command, slug, toplevel, default, template")
	}

	if strings.Contains(errStr, "unknown section type") {
		suggestions = append(suggestions, "Valid section types: example, tutorial, topic, application")
	}

	if strings.Contains(errStr, "expected") {
		suggestions = append(suggestions, "Check for missing quotes, colons, or parentheses")
	}

	if strings.Contains(errStr, "unexpected token") {
		suggestions = append(suggestions, "Check for extra or misplaced characters")
	}

	if len(suggestions) == 0 {
		return ""
	}

	return "Suggestions:\n" + strings.Join(suggestions, "\n")
}

// IsValidField checks if a field name is valid
func IsValidField(field string) bool {
	compiler := NewCompiler()
	return compiler.ValidateField(field)
}

// IsValidType checks if a type value is valid
func IsValidType(typeValue string) bool {
	validTypes := map[string]bool{
		"example":     true,
		"tutorial":    true,
		"topic":       true,
		"application": true,
	}
	return validTypes[strings.ToLower(typeValue)]
}
