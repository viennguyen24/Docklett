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

*/

package parser

import (
	"docklett/compiler/token"
)

// recursive descent parser
// top down parser, goes from highest rule (grammar) to lowest (terminals)
// each rule is a function
type Parser struct {
	Tokens     []token.Token
	current    int
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

func (p *Parser) matchCurrentToken(tokenTypes ...token.TokenType) bool {
	for _, tokenType := range tokenTypes {
		if p.checkCurrentToken(tokenType) {
			p.advanceToken()
			return true
		}
	}
	return false
}

func (p *Parser) consumeMatchingToken(tokenType token.TokenType, errorMessage string) (token.Token, error) {
	if p.checkCurrentToken(tokenType) {
		return p.advanceToken(), nil
	}
	return token.Token{}, p.error(p.getCurrentToken(), errorMessage)
}

// Synchronize is only called in the event we encounter an error token. 
func (p *Parser) synchronize() {
	// Ignore the error token. This has already been reported from Parser.error()
	p.advanceToken()
	for!p.isAtEnd(){
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