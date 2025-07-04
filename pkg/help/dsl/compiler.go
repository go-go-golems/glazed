package dsl

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/help/store"
)

// Compiler compiles AST expressions to predicates
type Compiler struct {
	validFields map[string]bool
}

// NewCompiler creates a new compiler
func NewCompiler() *Compiler {
	return &Compiler{
		validFields: map[string]bool{
			"type":     true,
			"topic":    true,
			"flag":     true,
			"command":  true,
			"slug":     true,
			"toplevel": true,
			"default":  true,
			"template": true,
		},
	}
}

// Compile compiles an AST expression to a predicate
func (c *Compiler) Compile(expr Expression) (store.Predicate, error) {
	switch e := expr.(type) {
	case *BinaryExpression:
		return c.compileBinaryExpression(e)
	case *UnaryExpression:
		return c.compileUnaryExpression(e)
	case *FieldExpression:
		return c.compileFieldExpression(e)
	case *TextExpression:
		return c.compileTextExpression(e)
	case *IdentifierExpression:
		return c.compileIdentifierExpression(e)
	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// compileBinaryExpression compiles binary expressions (AND, OR)
func (c *Compiler) compileBinaryExpression(expr *BinaryExpression) (store.Predicate, error) {
	left, err := c.Compile(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := c.Compile(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "AND":
		return store.And(left, right), nil
	case "OR":
		return store.Or(left, right), nil
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", expr.Operator)
	}
}

// compileUnaryExpression compiles unary expressions (NOT)
func (c *Compiler) compileUnaryExpression(expr *UnaryExpression) (store.Predicate, error) {
	right, err := c.Compile(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "NOT":
		return store.Not(right), nil
	default:
		return nil, fmt.Errorf("unknown unary operator: %s", expr.Operator)
	}
}

// compileFieldExpression compiles field:value expressions
func (c *Compiler) compileFieldExpression(expr *FieldExpression) (store.Predicate, error) {
	field := strings.ToLower(expr.Field)
	value := strings.ToLower(expr.Value)

	if !c.validFields[field] {
		return nil, fmt.Errorf("unknown field: %s", field)
	}

	switch field {
	case "type":
		return c.compileTypeField(value)
	case "topic":
		return store.HasTopic(value), nil
	case "flag":
		// Handle flags with or without -- prefix
		if !strings.HasPrefix(value, "--") && !strings.HasPrefix(value, "-") {
			// Try both with and without -- prefix
			return store.Or(
				store.HasFlag(value),
				store.HasFlag("--"+value),
			), nil
		}
		return store.HasFlag(value), nil
	case "command":
		return store.HasCommand(value), nil
	case "slug":
		return store.SlugEquals(value), nil
	case "toplevel":
		return c.compileBooleanField(value, store.IsTopLevel())
	case "default":
		return c.compileBooleanField(value, store.ShownByDefault())
	case "template":
		return c.compileBooleanField(value, store.IsTemplate())
	default:
		return nil, fmt.Errorf("unsupported field: %s", field)
	}
}

// compileTypeField compiles type field values
func (c *Compiler) compileTypeField(value string) (store.Predicate, error) {
	switch value {
	case "example":
		return store.IsExample(), nil
	case "tutorial":
		return store.IsTutorial(), nil
	case "topic":
		return store.IsGeneralTopic(), nil
	case "application":
		return store.IsApplication(), nil
	default:
		return nil, fmt.Errorf("unknown section type: %s", value)
	}
}

// compileBooleanField compiles boolean field values
func (c *Compiler) compileBooleanField(value string, truePredicate store.Predicate) (store.Predicate, error) {
	switch value {
	case "true", "yes", "1":
		return truePredicate, nil
	case "false", "no", "0":
		return store.Not(truePredicate), nil
	default:
		return nil, fmt.Errorf("invalid boolean value: %s (use true/false)", value)
	}
}

// compileTextExpression compiles quoted text expressions
func (c *Compiler) compileTextExpression(expr *TextExpression) (store.Predicate, error) {
	// Use full-text search for quoted text
	return store.TextSearch(expr.Text), nil
}

// compileIdentifierExpression compiles shortcut identifiers
func (c *Compiler) compileIdentifierExpression(expr *IdentifierExpression) (store.Predicate, error) {
	value := strings.ToLower(expr.Value)

	switch value {
	case "examples":
		return store.IsExample(), nil
	case "tutorials":
		return store.IsTutorial(), nil
	case "topics":
		return store.IsGeneralTopic(), nil
	case "applications":
		return store.IsApplication(), nil
	case "toplevel":
		return store.IsTopLevel(), nil
	case "defaults":
		return store.ShownByDefault(), nil
	case "templates":
		return store.IsTemplate(), nil
	default:
		return nil, fmt.Errorf("unknown shortcut: %s", value)
	}
}

// ValidateField validates if a field name is supported
func (c *Compiler) ValidateField(field string) bool {
	return c.validFields[strings.ToLower(field)]
}

// GetValidFields returns all valid field names
func (c *Compiler) GetValidFields() []string {
	fields := make([]string, 0, len(c.validFields))
	for field := range c.validFields {
		fields = append(fields, field)
	}
	return fields
}

// GetValidTypeValues returns valid values for the type field
func (c *Compiler) GetValidTypeValues() []string {
	return []string{"example", "tutorial", "topic", "application"}
}

// GetValidShortcuts returns valid shortcut identifiers
func (c *Compiler) GetValidShortcuts() []string {
	return []string{"examples", "tutorials", "topics", "applications", "toplevel", "defaults", "templates"}
}
