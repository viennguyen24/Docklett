package parser

import (
	"docklett/compiler/token"
	"fmt"
	"os"
)

// Specific error for Parser
type ParseError struct {
	Token   token.Token
	Message string
}
func (e *ParseError) Error() string {
	return e.Message
}

// Parser's report 
func (p *Parser) error(tok token.Token, message string) *ParseError {
	p.reportError(tok, message)
	return &ParseError{
		Token:   tok,
		Message: message,
	}
}

func (p *Parser) reportError(tok token.Token, message string) {
	var location string

	if tok.Type == token.EOF {
		location = " at end"
	} else {
		location = fmt.Sprintf(" at '%s'", tok.Lexeme)
	}

	fmt.Fprintf(os.Stderr, "[line %d] Error%s: %s\n", tok.Line, location, message)
}