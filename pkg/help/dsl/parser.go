package dsl

import (
	"fmt"
	"strings"
)

// Node represents a node in the AST
type Node interface {
	String() string
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// BinaryExpression represents a binary operation (AND, OR)
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode() {}
func (be *BinaryExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", be.Left.String(), be.Operator, be.Right.String())
}

// UnaryExpression represents a unary operation (NOT)
type UnaryExpression struct {
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) expressionNode() {}
func (ue *UnaryExpression) String() string {
	return fmt.Sprintf("(%s %s)", ue.Operator, ue.Right.String())
}

// FieldExpression represents a field:value expression
type FieldExpression struct {
	Field string
	Value string
}

func (fe *FieldExpression) expressionNode() {}
func (fe *FieldExpression) String() string {
	return fmt.Sprintf("%s:%s", fe.Field, fe.Value)
}

// TextExpression represents a quoted text search
type TextExpression struct {
	Text string
}

func (te *TextExpression) expressionNode() {}
func (te *TextExpression) String() string {
	return fmt.Sprintf("\"%s\"", te.Text)
}

// Parser parses tokens into an AST
type Parser struct {
	lexer *Lexer

	curToken  Token
	peekToken Token

	errors []string
}

// NewParser creates a new parser
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []string{},
	}

	// Read two tokens so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances the parser tokens
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

// ParseExpression parses the entire expression
func (p *Parser) ParseExpression() Expression {
	return p.parseOrExpression()
}

// parseOrExpression parses OR expressions (lowest precedence)
func (p *Parser) parseOrExpression() Expression {
	expr := p.parseAndExpression()

	for p.curToken.Type == TokenOr {
		operator := p.curToken.Value
		p.nextToken()
		right := p.parseAndExpression()
		expr = &BinaryExpression{
			Left:     expr,
			Operator: strings.ToUpper(operator),
			Right:    right,
		}
	}

	return expr
}

// parseAndExpression parses AND expressions (middle precedence)
func (p *Parser) parseAndExpression() Expression {
	expr := p.parseNotExpression()

	for p.curToken.Type == TokenAnd {
		operator := p.curToken.Value
		p.nextToken()
		right := p.parseNotExpression()
		expr = &BinaryExpression{
			Left:     expr,
			Operator: strings.ToUpper(operator),
			Right:    right,
		}
	}

	return expr
}

// parseNotExpression parses NOT expressions (higher precedence)
func (p *Parser) parseNotExpression() Expression {
	if p.curToken.Type == TokenNot {
		operator := p.curToken.Value
		p.nextToken()
		right := p.parseNotExpression()
		return &UnaryExpression{
			Operator: strings.ToUpper(operator),
			Right:    right,
		}
	}

	return p.parsePrimaryExpression()
}

// parsePrimaryExpression parses primary expressions (highest precedence)
func (p *Parser) parsePrimaryExpression() Expression {
	switch p.curToken.Type {
	case TokenLeftParen:
		return p.parseGroupedExpression()
	case TokenString:
		return p.parseTextExpression()
	case TokenIdent:
		// Check if next token is colon for field expression
		if p.peekToken.Type == TokenColon {
			return p.parseFieldExpression()
		}
		return p.parseImplicitTextExpression()
	case TokenIllegal, TokenEOF, TokenError, TokenFieldName, TokenColon, TokenAnd, TokenOr, TokenNot, TokenRightParen:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

// parseGroupedExpression parses expressions in parentheses
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken() // consume '('

	expr := p.parseOrExpression()

	if p.curToken.Type != TokenRightParen {
		p.addError("expected ')' after grouped expression")
		return nil
	}

	p.nextToken() // consume ')'
	return expr
}

// parseTextExpression parses quoted text expressions
func (p *Parser) parseTextExpression() Expression {
	text := p.curToken.Value
	p.nextToken()
	return &TextExpression{Text: text}
}

// parseFieldExpression parses field:value expressions
func (p *Parser) parseFieldExpression() Expression {
	field := p.curToken.Value
	p.nextToken() // consume field

	if p.curToken.Type != TokenColon {
		p.addError("expected ':' after field name")
		return nil
	}
	p.nextToken() // consume ':'

	if p.curToken.Type != TokenIdent {
		p.addError("expected value after ':'")
		return nil
	}

	value := p.curToken.Value
	p.nextToken()

	return &FieldExpression{
		Field: strings.ToLower(field),
		Value: strings.ToLower(value),
	}
}

// parseImplicitTextExpression parses unquoted text as implicit text search
func (p *Parser) parseImplicitTextExpression() Expression {
	var textParts []string

	// Collect all consecutive identifiers until we hit a boolean operator or EOF
	for p.curToken.Type == TokenIdent {
		textParts = append(textParts, p.curToken.Value)
		p.nextToken()

		// Stop if we hit a boolean operator, colon, or parenthesis
		if p.curToken.Type == TokenAnd || p.curToken.Type == TokenOr ||
			p.curToken.Type == TokenNot || p.curToken.Type == TokenColon ||
			p.curToken.Type == TokenLeftParen || p.curToken.Type == TokenRightParen ||
			p.curToken.Type == TokenEOF {
			break
		}
	}

	// Join all collected text parts
	text := strings.Join(textParts, " ")
	return &TextExpression{Text: text}
}

// addError adds an error to the parser
func (p *Parser) addError(msg string) {
	errorMsg := fmt.Sprintf("parse error at line %d, column %d: %s",
		p.curToken.Line, p.curToken.Column, msg)
	p.errors = append(p.errors, errorMsg)
}

// Errors returns all parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

// HasErrors returns true if there are any errors
func (p *Parser) HasErrors() bool {
	return len(p.errors) > 0
}

// Parse parses the input and returns the AST
func Parse(input string) (Expression, error) {
	lexer := NewLexer(input)
	parser := NewParser(lexer)

	expr := parser.ParseExpression()

	if parser.HasErrors() {
		return nil, fmt.Errorf("parse errors: %s", strings.Join(parser.Errors(), "; "))
	}

	// Check for unexpected tokens after parsing
	if parser.curToken.Type != TokenEOF {
		return nil, fmt.Errorf("unexpected token after expression: %s", parser.curToken.Type)
	}

	return expr, nil
}
