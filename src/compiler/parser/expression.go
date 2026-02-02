package parser

import (
	"docklett/compiler/token"
)

type Expression interface {
	exprNode()
}

type LiteralExpression struct {
	Value any
}

func (le LiteralExpression) exprNode() {}

type Binary struct {
	Left     Expression
	Right    Expression
	Operator token.Token
}

func (b Binary) exprNode() {}

type Unary struct {
	Operator token.Token
	Right    Expression
}

func (u Unary) exprNode() {}

type Grouping struct {
	Expression
}

// This is the most base line rule and return an immediate value
// It can be a literal (true, false, string, number) or a grouped expression, which calls back to expression() to parse it completely top level down again
// We add the group expression because parenthesis have the highest precedence, and we want to treat it as a single unit of value like literals and identifiers
func (p *Parser) primary() (Expression, error) {
	if p.matchCurrentToken(token.TRUE) {
		return LiteralExpression{Value: true}, nil
	}
	if p.matchCurrentToken(token.FALSE) {
		return LiteralExpression{Value: false}, nil
	}
	if p.matchCurrentToken(token.STRING) {
		return LiteralExpression{Value: p.getPreviousToken().Literal}, nil
	}
	if p.matchCurrentToken(token.NUMBER) {
		return LiteralExpression{Value: p.getPreviousToken().Literal}, nil
	}

	// If token is an opening parenthesis, the next tokens must form a new expression followed by a closing parenthesis token
	if p.matchCurrentToken(token.LPAREN) {
		expression, _ := p.expression()
		_, err := p.consumeMatchingToken(token.RPAREN, "Expected ')' after expression.")
		if err != nil {
			return nil, err
		}
		return Grouping{Expression: expression}, nil
	}
	return nil, p.error(p.getCurrentToken(), "Unexpected token "+p.getCurrentToken().Lexeme)
}

// An unary just takes the immediate value returned from primary and mutate that
// There can be an arbitrary number of unary operators before getting to the actual value
func (p *Parser) unary() (Expression, error) {
	if p.matchCurrentToken(token.NEGATE, token.SUBTRACT) {
		prev := p.getPreviousToken()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return Unary{Operator: prev, Right: right}, nil
	}
	primaryValue, err := p.primary()
	if err != nil {
		return nil, err
	}
	return primaryValue, nil
}

// A factor rule is defined as an unary (now a single unit of actual value) followed by
// an arbitray number of this structure: (MUL|DIV) unary
// We implement a "running epxression" approach by continuously consumming next tokens and add that to our current expression
func (p *Parser) factor() (Expression, error) {
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
		expr = Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// A term is just like factor, but for ADD and SUBTRACT operators after we have completed MULTIPLY and DIVIDE operations in an expresison
func (p *Parser) term() (Expression, error) {
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
		expr = Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) comparison() (Expression, error) {
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
		expr = Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) equality() (Expression, error) {
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
		expr = Binary{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

func (p *Parser) expression() (Expression, error) {
	var equ Expression
	var err error
	for !p.isAtEnd() {
		equ, err = p.equality()
		if err != nil {
			p.synchronize()
		}
	}
	return equ, nil
}

func (p *Parser) Parse() (Expression, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	return expr, nil
}
