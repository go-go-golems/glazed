package search

import (
	"fmt"
	"strings"
)

// Parser parses tokenized query strings into AST
type Parser struct {
	tokens   []Token
	position int
	current  Token
}

// NewParser creates a new parser for the given tokens
func NewParser(tokens []Token) *Parser {
	p := &Parser{
		tokens: tokens,
	}
	p.advance()
	return p
}

// advance moves to the next token
func (p *Parser) advance() {
	if p.position >= len(p.tokens) {
		p.current = Token{Type: TokenEOF}
		return
	}
	
	p.current = p.tokens[p.position]
	p.position++
}

// peek returns the next token without advancing
func (p *Parser) peek() Token {
	if p.position >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.position]
}

// expect consumes a token of the expected type or returns an error
func (p *Parser) expect(tokenType TokenType) (Token, error) {
	if p.current.Type != tokenType {
		return Token{}, fmt.Errorf("expected %v, got %v at line %d, column %d", 
			tokenType, p.current.Type, p.current.Line, p.current.Column)
	}
	
	token := p.current
	p.advance()
	return token, nil
}

// match checks if the current token matches any of the given types
func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.current.Type == t {
			return true
		}
	}
	return false
}

// Parse parses the tokens into an AST
func (p *Parser) Parse() (*Query, error) {
	if p.current.Type == TokenEOF {
		return &Query{Root: nil, Raw: ""}, nil
	}
	
	root, err := p.parseOrExpression()
	if err != nil {
		return nil, err
	}
	
	if p.current.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token after query: %v at line %d, column %d", 
			p.current.Type, p.current.Line, p.current.Column)
	}
	
	return &Query{Root: root}, nil
}

// parseOrExpression parses OR expressions (lowest precedence)
func (p *Parser) parseOrExpression() (QueryNode, error) {
	left, err := p.parseAndExpression()
	if err != nil {
		return nil, err
	}
	
	for p.match(TokenOr) {
		p.advance()
		right, err := p.parseAndExpression()
		if err != nil {
			return nil, err
		}
		left = NewBooleanNode("OR", left, right)
	}
	
	return left, nil
}

// parseAndExpression parses AND expressions (higher precedence than OR)
func (p *Parser) parseAndExpression() (QueryNode, error) {
	left, err := p.parseNotExpression()
	if err != nil {
		return nil, err
	}
	
	// Handle implicit AND (space-separated terms)
	for p.match(TokenWord, TokenString, TokenFilter, TokenMinus, TokenLParen, TokenNot) {
		// Explicit AND
		if p.match(TokenAnd) {
			p.advance()
		}
		
		right, err := p.parseNotExpression()
		if err != nil {
			return nil, err
		}
		left = NewBooleanNode("AND", left, right)
	}
	
	return left, nil
}

// parseNotExpression parses NOT expressions (highest precedence)
func (p *Parser) parseNotExpression() (QueryNode, error) {
	if p.match(TokenNot) {
		p.advance()
		expr, err := p.parseNotExpression()
		if err != nil {
			return nil, err
		}
		return NewBooleanNode("NOT", expr, nil), nil
	}
	
	return p.parsePrimary()
}

// parsePrimary parses primary expressions (terms, filters, groups)
func (p *Parser) parsePrimary() (QueryNode, error) {
	switch p.current.Type {
	case TokenLParen:
		return p.parseGroup()
	case TokenMinus:
		return p.parseNegatedTerm()
	case TokenFilter:
		return p.parseFilter()
	case TokenWord:
		return p.parseWord()
	case TokenString:
		return p.parseString()
	default:
		return nil, fmt.Errorf("unexpected token: %v at line %d, column %d", 
			p.current.Type, p.current.Line, p.current.Column)
	}
}

// parseGroup parses parenthesized expressions
func (p *Parser) parseGroup() (QueryNode, error) {
	_, err := p.expect(TokenLParen)
	if err != nil {
		return nil, err
	}
	
	expr, err := p.parseOrExpression()
	if err != nil {
		return nil, err
	}
	
	_, err = p.expect(TokenRParen)
	if err != nil {
		return nil, err
	}
	
	return NewGroupNode(expr), nil
}

// parseNegatedTerm parses negated terms (starting with -)
func (p *Parser) parseNegatedTerm() (QueryNode, error) {
	_, err := p.expect(TokenMinus)
	if err != nil {
		return nil, err
	}
	
	switch p.current.Type {
	case TokenFilter:
		filter, err := p.parseFilter()
		if err != nil {
			return nil, err
		}
		if f, ok := filter.(*FilterNode); ok {
			f.Negated = true
		}
		return filter, nil
	case TokenWord:
		word, err := p.parseWord()
		if err != nil {
			return nil, err
		}
		if w, ok := word.(*TextNode); ok {
			w.Negated = true
		}
		return word, nil
	case TokenString:
		str, err := p.parseString()
		if err != nil {
			return nil, err
		}
		if s, ok := str.(*TextNode); ok {
			s.Negated = true
		}
		return str, nil
	default:
		return nil, fmt.Errorf("expected filter, word, or string after '-' at line %d, column %d", 
			p.current.Line, p.current.Column)
	}
}

// parseFilter parses field:value filters
func (p *Parser) parseFilter() (QueryNode, error) {
	token, err := p.expect(TokenFilter)
	if err != nil {
		return nil, err
	}
	
	field, operator, value, err := ParseFilter(token.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid filter '%s' at line %d, column %d: %v", 
			token.Value, token.Line, token.Column, err)
	}
	
	// Validate field name
	if !IsValidFieldName(field) {
		return nil, fmt.Errorf("invalid field name '%s' at line %d, column %d", 
			field, token.Line, token.Column)
	}
	
	return &FilterNode{
		Field:    NormalizeFieldName(field),
		Value:    value,
		Operator: operator,
		Negated:  false,
	}, nil
}

// parseWord parses word tokens as text search terms
func (p *Parser) parseWord() (QueryNode, error) {
	token, err := p.expect(TokenWord)
	if err != nil {
		return nil, err
	}
	
	return NewTextNode(token.Value, false, false), nil
}

// parseString parses quoted strings as text search terms
func (p *Parser) parseString() (QueryNode, error) {
	token, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}
	
	return NewTextNode(token.Value, false, true), nil
}

// ParseQuery is a convenience function that lexes and parses a query string
func ParseQuery(input string) (*Query, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	
	parser := NewParser(tokens)
	query, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	
	query.Raw = input
	return query, nil
}

// ValidateQuery performs semantic validation on a parsed query
func ValidateQuery(query *Query) error {
	if query.Root == nil {
		return nil
	}
	
	var errors []string
	
	// Check for valid field names
	filters := CollectFilters(query.Root)
	for _, filter := range filters {
		if !IsValidFieldName(filter.Field) {
			errors = append(errors, fmt.Sprintf("invalid field name: %s", filter.Field))
		}
	}
	
	// Check for empty values
	for _, filter := range filters {
		if strings.TrimSpace(filter.Value) == "" {
			errors = append(errors, fmt.Sprintf("empty value for field: %s", filter.Field))
		}
	}
	
	// Check for empty text searches
	texts := CollectTextSearches(query.Root)
	for _, text := range texts {
		if strings.TrimSpace(text.Text) == "" {
			errors = append(errors, "empty text search term")
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("query validation errors: %s", strings.Join(errors, "; "))
	}
	
	return nil
}

// QueryInfo provides information about a parsed query
type QueryInfo struct {
	HasFilters    bool
	HasTextSearch bool
	Fields        []string
	TextTerms     []string
	IsSimple      bool // true if query has no boolean operators
}

// AnalyzeQuery analyzes a parsed query and returns information about it
func AnalyzeQuery(query *Query) *QueryInfo {
	if query.Root == nil {
		return &QueryInfo{IsSimple: true}
	}
	
	info := &QueryInfo{}
	
	// Check for filters
	filters := CollectFilters(query.Root)
	info.HasFilters = len(filters) > 0
	
	// Collect field names
	fieldMap := make(map[string]bool)
	for _, filter := range filters {
		fieldMap[filter.Field] = true
	}
	for field := range fieldMap {
		info.Fields = append(info.Fields, field)
	}
	
	// Check for text searches
	texts := CollectTextSearches(query.Root)
	info.HasTextSearch = len(texts) > 0
	
	// Collect text terms
	for _, text := range texts {
		info.TextTerms = append(info.TextTerms, text.Text)
	}
	
	// Check if query is simple (no boolean operators)
	info.IsSimple = !containsBooleanOperators(query.Root)
	
	return info
}

// containsBooleanOperators checks if the AST contains any boolean operators
func containsBooleanOperators(node QueryNode) bool {
	if node == nil {
		return false
	}
	
	switch n := node.(type) {
	case *BooleanNode:
		return true
	case *GroupNode:
		return containsBooleanOperators(n.Child)
	default:
		return false
	}
}

// FormatQuery formats a query for display
func FormatQuery(query *Query) string {
	if query.Root == nil {
		return ""
	}
	return formatNode(query.Root, 0)
}

// formatNode formats a query node with proper indentation
func formatNode(node QueryNode, indent int) string {
	if node == nil {
		return ""
	}
	
	prefix := strings.Repeat("  ", indent)
	
	switch n := node.(type) {
	case *FilterNode:
		neg := ""
		if n.Negated {
			neg = "NOT "
		}
		return fmt.Sprintf("%s%s%s%s%s", prefix, neg, n.Field, n.Operator, n.Value)
	case *TextNode:
		neg := ""
		if n.Negated {
			neg = "NOT "
		}
		if n.Quoted {
			return fmt.Sprintf("%s%s\"%s\"", prefix, neg, n.Text)
		}
		return fmt.Sprintf("%s%s%s", prefix, neg, n.Text)
	case *BooleanNode:
		if n.Operator == "NOT" {
			return fmt.Sprintf("%sNOT\n%s", prefix, formatNode(n.Left, indent+1))
		}
		return fmt.Sprintf("%s%s\n%s\n%s", prefix, n.Operator, 
			formatNode(n.Left, indent+1), formatNode(n.Right, indent+1))
	case *GroupNode:
		return fmt.Sprintf("%s(\n%s\n%s)", prefix, formatNode(n.Child, indent+1), prefix)
	default:
		return fmt.Sprintf("%s<unknown node type>", prefix)
	}
}
