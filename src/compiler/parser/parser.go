/*
Parser takes a sequence of Tokens from the scanner and performs 2 responsibility:
1. If the sequence is valid, create an Abstract Syntax Tree (AST)
2. If the sequence is invalid, report all errors to the user in useful way

Not all input Docklett file is valid, e.g. user might accidentally commit a half-written file with incomplete tokens
We must stop all errors and edge cases before an invalid AST is passed to interpreter and do so gracefully

We design error recovery to be "panic and synchronize" mode. Essentially, discard an error token immediately and jumps the parser into a safer spot to continue parsing
1. Panic: immediately return to higher level rule
2. Synchronize: corrects the input stream such that the following tokens doesn't feed off the same error token. In other words, we skip this rule to a better parsing position

https://teaching.idallen.com/cst8152/98w/panic_mode.html

The final output should be a tree of statement nodes not evaluated yet

*/

package parser

import (
	"docklett/compiler/ast"
	compileError "docklett/compiler/error"
	"docklett/compiler/token"
	"errors"
)

// recursive descent parser
// top down parser, goes from highest rule (grammar) to lowest (terminals)
// each rule is a function
type Parser struct {
	Tokens  []token.Token
	current int
}

// consume the current token and advance to the next
func (p *Parser) advanceToken() token.Token {
	currentToken := p.Tokens[p.current]
	p.current += 1
	return currentToken
}

func (p *Parser) getCurrentToken() token.Token {
	return p.Tokens[p.current]
}

func (p *Parser) getPreviousToken() token.Token {
	return p.Tokens[p.current-1]
}

func (p *Parser) isAtEnd() bool {
	return p.getCurrentToken().Type == token.EOF
}

func (p *Parser) checkCurrentToken(tokenType token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.getCurrentToken().Type == tokenType
}

// This is a simple check whether the current token matches one of the following type
// It is mostly for branching conditions to handle different rules, and does not threaten to break the program if return false
func (p *Parser) matchCurrentToken(tokenTypes ...token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.checkCurrentToken(tokenType) {
			p.advanceToken()
			return true
		}
	}
	return false
}

// Consumes current token and returns it if it matches tokenType, otherwise returns error
func (p *Parser) consumeMatchingToken(tokenType token.TokenType, errorMessage string) (token.Token, error) {
	if p.checkCurrentToken(tokenType) {
		return p.advanceToken(), nil
	}
	return token.Token{}, compileError.NewParseError(p.getCurrentToken(), errorMessage)
}

// Synchronize discards tokens until a safe parsing boundary (newline or keyword).
func (p *Parser) synchronize() {
	// Ignore the error token. This has already been reported
	p.advanceToken()
	for !p.isAtEnd() {
		currentToken := p.getCurrentToken()
		if p.getPreviousToken().Lexeme == `\n` {
			return
		}
		// If we find a Docklett or Docker keyword, we can synchronize because these are guaranteed valid tokens
		// We dont have to consume it
		if _, isDocklett := token.DocklettTokenKeywords[currentToken.Lexeme]; isDocklett {
			return
		}
		if _, isDocker := token.DockerTokenKeywords[currentToken.Lexeme]; isDocker {
			return
		}
		p.advanceToken()
	}
}

// Parse processes the token stream into a list of statement AST nodes.
// Collects all parse errors via synchronize recovery and returns them joined at the end.
func (p *Parser) Parse(tokens []token.Token) ([]ast.Statement, error) {
	p.Tokens = tokens
	var statements []ast.Statement
	var parseErrors []error

	for !p.isAtEnd() {
		stmt, err := p.declaration()
		if err != nil {
			parseErrors = append(parseErrors, err)
			continue
		}
		statements = append(statements, stmt)
	}

	if len(parseErrors) > 0 {
		return nil, errors.Join(parseErrors...)
	}
	return statements, nil
}
