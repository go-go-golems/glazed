package dsl

import (
	"fmt"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	TokenIllegal TokenType = iota
	TokenEOF
	TokenError

	// Literals
	TokenIdent     // identifiers like field names, values
	TokenString    // quoted strings like "hello world"
	TokenFieldName // field names before colon

	// Operators
	TokenColon      // :
	TokenAnd        // AND
	TokenOr         // OR
	TokenNot        // NOT
	TokenLeftParen  // (
	TokenRightParen // )
)

// Token represents a single token
type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Type.String(), t.Value)
}

// String returns a string representation of the token type
func (tt TokenType) String() string {
	switch tt {
	case TokenIllegal:
		return "ILLEGAL"
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	case TokenIdent:
		return "IDENT"
	case TokenString:
		return "STRING"
	case TokenFieldName:
		return "FIELD"
	case TokenColon:
		return "COLON"
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenLeftParen:
		return "LPAREN"
	case TokenRightParen:
		return "RPAREN"
	default:
		return "UNKNOWN"
	}
}

// Lexer tokenizes input text
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// NewLexer creates a new lexer
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL character represents EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// NextToken returns the next token
func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	tok.Position = l.position
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case ':':
		tok = Token{Type: TokenColon, Value: string(l.ch), Position: l.position, Line: l.line, Column: l.column}
	case '(':
		tok = Token{Type: TokenLeftParen, Value: string(l.ch), Position: l.position, Line: l.line, Column: l.column}
	case ')':
		tok = Token{Type: TokenRightParen, Value: string(l.ch), Position: l.position, Line: l.line, Column: l.column}
	case '"':
		tok.Type = TokenString
		tok.Value = l.readString('"')
		return tok // readString advances position, so we don't call readChar
	case '\'':
		tok.Type = TokenString
		tok.Value = l.readString('\'')
		return tok // readString advances position, so we don't call readChar
	case 0:
		tok.Type = TokenEOF
		tok.Value = ""
	default:
		if isLetter(l.ch) || l.ch == '-' {
			tok.Position = l.position
			tok.Line = l.line
			tok.Column = l.column
			tok.Value = l.readIdentifier()
			tok.Type = l.lookupIdent(tok.Value)
			return tok // readIdentifier advances position, so we don't call readChar
		} else {
			tok = Token{Type: TokenIllegal, Value: string(l.ch), Position: l.position, Line: l.line, Column: l.column}
		}
	}

	l.readChar()
	return tok
}

// readString reads a quoted string
func (l *Lexer) readString(quote byte) string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == quote || l.ch == 0 {
			break
		}
	}
	value := l.input[position:l.position]
	// Advance past the closing quote
	if l.ch == quote {
		l.readChar()
	}
	return value
}

// readIdentifier reads an identifier or keyword
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '-' || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// lookupIdent determines if an identifier is a keyword
func (l *Lexer) lookupIdent(ident string) TokenType {
	switch ident {
	case "AND", "and":
		return TokenAnd
	case "OR", "or":
		return TokenOr
	case "NOT", "not":
		return TokenNot
	default:
		return TokenIdent
	}
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// isLetter checks if character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// GetAllTokens returns all tokens from the input (useful for testing)
func (l *Lexer) GetAllTokens() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

// Error creates an error token
func (l *Lexer) Error(msg string) Token {
	return Token{
		Type:     TokenError,
		Value:    msg,
		Position: l.position,
		Line:     l.line,
		Column:   l.column,
	}
}

// IsAtEnd checks if lexer is at end of input
func (l *Lexer) IsAtEnd() bool {
	return l.ch == 0
}

// CurrentPosition returns current position information
func (l *Lexer) CurrentPosition() (int, int, int) {
	return l.position, l.line, l.column
}
