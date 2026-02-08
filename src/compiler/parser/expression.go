/*
  The following defines all node interfaces for an expression and Recursive Descent Parser implementation to verify that an expression
  follows Docklett's grammar rules.
  The syntax rules is defined in the root design folder
*/

package parser

import (
	"docklett/compiler/ast"
	"docklett/compiler/token"
)

// This is the most base line rule and return an immediate value
// It can be a literal (true, false, string, number) or a grouped expression, which calls back to expression() to parse it completely top level down again
// We add the group expression because parenthesis have the highest precedence, and we want to treat it as a single unit of value like literals and identifiers
func (p *Parser) primary() (ast.Expression, error) {
	if p.matchCurrentToken(token.TRUE) {
		return &ast.LiteralExpression{Value: true, Token: p.getPreviousToken()}, nil
	}
	if p.matchCurrentToken(token.FALSE) {
		return &ast.LiteralExpression{Value: false, Token: p.getPreviousToken()}, nil
	}
	if p.matchCurrentToken(token.STRING, token.NUMBER) {
		prev := p.getPreviousToken()
		return &ast.LiteralExpression{Value: prev.Literal, Token: prev}, nil
	}
	if p.matchCurrentToken(token.IDENTIFIER) {
		return &ast.LiteralExpression{Value: p.getPreviousToken().Lexeme, Token: p.getPreviousToken()}, nil
	}

	// If token is an opening parenthesis, the next tokens must form a new expression followed by a closing parenthesis token
	if p.matchCurrentToken(token.LPAREN) {
		expression, err := p.expression()
		if err != nil {
			return nil, err
		}
		_, err = p.consumeMatchingToken(token.RPAREN, "Expected ')' after expression.")
		if err != nil {
			return nil, err
		}
		return &ast.Grouping{Expression: expression}, nil
	}
	return nil, p.error(p.getCurrentToken(), "Unexpected token "+p.getCurrentToken().Lexeme)
}

// An unary just takes the immediate value returned from primary and mutate that
// There can be an arbitrary number of unary operators before getting to the actual value
func (p *Parser) unary() (ast.Expression, error) {
	if p.matchCurrentToken(token.NEGATE, token.SUBTRACT) {
		prev := p.getPreviousToken()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.Unary{Operator: prev, Right: right}, nil
	}
	return p.primary()
}

// A factor rule is defined as an unary (now a single unit of actual value) followed by
// an arbitray number of this structure: (MUL|DIV) unary
// We implement a "running epxression" approach by continuously consumming next tokens and add that to our current expression
func (p *Parser) factor() (ast.Expression, error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.matchCurrentToken(token.MULTI, token.DIVIDE) {
		operator := p.getPreviousToken()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// A term is just like factor, but for ADD and SUBTRACT operators after we have completed MULTIPLY and DIVIDE operations in an expresison
func (p *Parser) term() (ast.Expression, error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}
	for p.matchCurrentToken(token.ADD, token.SUBTRACT) {
		operator := p.getPreviousToken()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) comparison() (ast.Expression, error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.matchCurrentToken(token.GTE, token.GREATER, token.LTE, token.LESS) {
		operator := p.getPreviousToken()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) equality() (ast.Expression, error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}
	for p.matchCurrentToken(token.EQUAL, token.UNEQUAL) {
		operator := p.getPreviousToken()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = &ast.Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) expression() (ast.Expression, error) {
	return p.equality()
}
