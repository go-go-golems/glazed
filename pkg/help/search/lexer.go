package search

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	// Basic tokens
	TokenEOF TokenType = iota
	TokenError
	
	// Operators
	TokenAnd      // AND
	TokenOr       // OR
	TokenNot      // NOT
	TokenMinus    // -
	TokenLParen   // (
	TokenRParen   // )
	TokenColon    // :
	TokenEquals   // =
	TokenTilde    // ~
	
	// Values
	TokenWord     // unquoted word
	TokenString   // quoted string
	TokenFilter   // field:value or field=value
	
	// Special
	TokenSpace    // whitespace (usually ignored)
)

// Token represents a single token
type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}

func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return fmt.Sprintf("ERROR: %s", t.Value)
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenMinus:
		return "-"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenColon:
		return ":"
	case TokenEquals:
		return "="
	case TokenTilde:
		return "~"
	case TokenWord:
		return fmt.Sprintf("WORD(%s)", t.Value)
	case TokenString:
		return fmt.Sprintf("STRING(%s)", t.Value)
	case TokenFilter:
		return fmt.Sprintf("FILTER(%s)", t.Value)
	case TokenSpace:
		return "SPACE"
	default:
		return fmt.Sprintf("UNKNOWN(%s)", t.Value)
	}
}

// Lexer tokenizes query strings
type Lexer struct {
	input    string
	position int
	line     int
	column   int
	current  rune
	tokens   []Token
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
	l.advance()
	return l
}

// advance moves to the next character
func (l *Lexer) advance() {
	if l.position >= len(l.input) {
		l.current = 0 // EOF
		return
	}
	
	if l.current == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	
	l.current = rune(l.input[l.position])
	l.position++
}

// peek returns the next character without advancing
func (l *Lexer) peek() rune {
	if l.position >= len(l.input) {
		return 0
	}
	return rune(l.input[l.position])
}

// peekN returns the character n positions ahead
func (l *Lexer) peekN(n int) rune {
	pos := l.position + n - 1
	if pos >= len(l.input) {
		return 0
	}
	return rune(l.input[pos])
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.current != 0 && unicode.IsSpace(l.current) {
		l.advance()
	}
}

// readWord reads a word token
func (l *Lexer) readWord() string {
	start := l.position - 1
	for l.current != 0 && (unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' || l.current == '-' || l.current == '.' || l.current == '/') {
		l.advance()
	}
	return l.input[start : l.position-1]
}

// readString reads a quoted string
func (l *Lexer) readString() (string, error) {
	quote := l.current
	l.advance() // skip opening quote
	
	var result strings.Builder
	for l.current != 0 && l.current != quote {
		if l.current == '\\' {
			l.advance()
			if l.current == 0 {
				return "", fmt.Errorf("unterminated string at line %d, column %d", l.line, l.column)
			}
			
			switch l.current {
			case 'n':
				result.WriteRune('\n')
			case 't':
				result.WriteRune('\t')
			case 'r':
				result.WriteRune('\r')
			case '\\':
				result.WriteRune('\\')
			case '"':
				result.WriteRune('"')
			case '\'':
				result.WriteRune('\'')
			default:
				result.WriteRune(l.current)
			}
		} else {
			result.WriteRune(l.current)
		}
		l.advance()
	}
	
	if l.current == 0 {
		return "", fmt.Errorf("unterminated string at line %d, column %d", l.line, l.column)
	}
	
	l.advance() // skip closing quote
	return result.String(), nil
}

// readFilter reads a field:value or field=value filter
func (l *Lexer) readFilter() (string, error) {
	field := l.readWord()
	if field == "" {
		return "", fmt.Errorf("empty field name at line %d, column %d", l.line, l.column)
	}
	
	// Check for operator
	var operator string
	switch l.current {
	case ':':
		operator = ":"
		l.advance()
	case '=':
		operator = "="
		l.advance()
	case '~':
		operator = "~"
		l.advance()
	default:
		return "", fmt.Errorf("expected ':' or '=' after field name at line %d, column %d", l.line, l.column)
	}
	
	// Read value
	var value string
	var err error
	
	if l.current == '"' || l.current == '\'' {
		value, err = l.readString()
		if err != nil {
			return "", err
		}
	} else {
		value = l.readWord()
		if value == "" {
			return "", fmt.Errorf("empty value after '%s' at line %d, column %d", operator, l.line, l.column)
		}
	}
	
	return field + operator + value, nil
}

// createToken creates a token at the current position
func (l *Lexer) createToken(tokenType TokenType, value string) Token {
	return Token{
		Type:     tokenType,
		Value:    value,
		Position: l.position - len(value),
		Line:     l.line,
		Column:   l.column - len(value),
	}
}

// nextToken returns the next token
func (l *Lexer) nextToken() Token {
	for {
		l.skipWhitespace()
		
		if l.current == 0 {
			return l.createToken(TokenEOF, "")
		}
		
		// Save position for token creation
		line := l.line
		column := l.column
		
		switch l.current {
		case '(':
			l.advance()
			return Token{Type: TokenLParen, Value: "(", Position: l.position - 1, Line: line, Column: column}
		case ')':
			l.advance()
			return Token{Type: TokenRParen, Value: ")", Position: l.position - 1, Line: line, Column: column}
		case '-':
			l.advance()
			return Token{Type: TokenMinus, Value: "-", Position: l.position - 1, Line: line, Column: column}
		case '"', '\'':
			str, err := l.readString()
			if err != nil {
				return Token{Type: TokenError, Value: err.Error(), Position: l.position - 1, Line: line, Column: column}
			}
			return Token{Type: TokenString, Value: str, Position: l.position - len(str) - 2, Line: line, Column: column}
		default:
			if unicode.IsLetter(l.current) || l.current == '_' {
				word := l.readWord()
				
				// Check if it's a keyword
				switch strings.ToUpper(word) {
				case "AND":
					return Token{Type: TokenAnd, Value: word, Position: l.position - len(word), Line: line, Column: column}
				case "OR":
					return Token{Type: TokenOr, Value: word, Position: l.position - len(word), Line: line, Column: column}
				case "NOT":
					return Token{Type: TokenNot, Value: word, Position: l.position - len(word), Line: line, Column: column}
				}
				
				// Check if it's a filter (word followed by : or =)
				if l.current == ':' || l.current == '=' || l.current == '~' {
					// Reset position to re-read as filter
					l.position = l.position - len(word)
					l.column = column
					l.current = rune(l.input[l.position-1])
					
					filter, err := l.readFilter()
					if err != nil {
						return Token{Type: TokenError, Value: err.Error(), Position: l.position - 1, Line: line, Column: column}
					}
					return Token{Type: TokenFilter, Value: filter, Position: l.position - len(filter), Line: line, Column: column}
				}
				
				// Regular word
				return Token{Type: TokenWord, Value: word, Position: l.position - len(word), Line: line, Column: column}
			}
			
			// Unknown character
			char := l.current
			l.advance()
			return Token{
				Type:     TokenError,
				Value:    fmt.Sprintf("unexpected character '%c'", char),
				Position: l.position - 1,
				Line:     line,
				Column:   column,
			}
		}
	}
}

// Tokenize tokenizes the entire input and returns all tokens
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	
	for {
		token := l.nextToken()
		tokens = append(tokens, token)
		
		if token.Type == TokenEOF {
			break
		}
		
		if token.Type == TokenError {
			return nil, fmt.Errorf("lexer error at line %d, column %d: %s", token.Line, token.Column, token.Value)
		}
	}
	
	return tokens, nil
}

// ParseFilter parses a filter string into field, operator, and value
func ParseFilter(filter string) (field, operator, value string, err error) {
	// Find the operator
	for i, ch := range filter {
		if ch == ':' || ch == '=' || ch == '~' {
			field = filter[:i]
			operator = string(ch)
			value = filter[i+1:]
			return
		}
	}
	
	return "", "", "", fmt.Errorf("invalid filter format: %s", filter)
}

// IsValidFieldName checks if a field name is valid
func IsValidFieldName(field string) bool {
	if field == "" {
		return false
	}
	
	validFields := map[string]bool{
		"type":     true,
		"topic":    true,
		"flag":     true,
		"command":  true,
		"toplevel": true,
		"default":  true,
		"slug":     true,
		"title":    true,
		"content":  true,
	}
	
	return validFields[strings.ToLower(field)]
}

// NormalizeFieldName normalizes field names to lowercase
func NormalizeFieldName(field string) string {
	return strings.ToLower(field)
}

// IsKeyword checks if a word is a reserved keyword
func IsKeyword(word string) bool {
	keywords := map[string]bool{
		"and": true,
		"or":  true,
		"not": true,
	}
	
	return keywords[strings.ToLower(word)]
}
