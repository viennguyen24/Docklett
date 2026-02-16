/*
Expression are parsed using recursive descent parsing.

RECURSIVE DESCENT PARSING:
Each grammar rule becomes a method. Methods call "higher" precedence rules (lower in the call chain).
Precedence from lowest to highest (call order):
  expression → assignment → equality → comparison → term → factor → unary → primary

GRAMMAR RULES (from Crafting Interpreters):
  expression     → assignment
  assignment     → IDENTIFIER "=" assignment | equality
  equality       → comparison ( ("==" | "!=") comparison )*
  comparison     → term ( (">" | ">=" | "<" | "<=") term )*
  term           → factor ( ("+" | "-") factor )*
  factor         → unary ( ("*" | "/") unary )*
  unary          → ("!" | "-") unary | primary
  primary        → NUMBER | STRING | "true" | "false" | IDENTIFIER | "(" expression ")"

PARSING STRATEGY:
Each function parses its level and delegates to higher-precedence rules.
Left-associative operators use iteration: a + b + c → (a + b) + c
Right-associative operators use recursion: a = b = c → a = (b = c)

EXAMPLE PARSE TREE:
  Source: (1 + 2) * 3
  Call chain: expression → ... → term → factor → unary → primary
  Result: BinaryExpression(GroupingExpression(BinaryExpression(Literal(1), +, Literal(2))), *, Literal(3))

USED BY: Parser.Parse() to build expression AST nodes from token sequences.
*/

package parser

import (
	"docklett/compiler/ast"
	compileError "docklett/compiler/error"
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
		return &ast.GroupingExpression{Expression: expression}, nil
	}
	return nil, compileError.NewParseError(p.getCurrentToken(), "Unexpected token "+p.getCurrentToken().Lexeme)
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
		return &ast.UnaryExpression{Operator: prev, Right: right}, nil
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
		expr = &ast.BinaryExpression{Left: expr, Operator: operator, Right: right}
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
		expr = &ast.BinaryExpression{Left: expr, Operator: operator, Right: right}
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
		expr = &ast.BinaryExpression{Left: expr, Operator: operator, Right: right}
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
		expr = &ast.BinaryExpression{Left: expr, Operator: operator, Right: right}
	}
	return expr, nil
}

// In an assignment, the left side is just an identifier that needs to be binded to a value, so we don't consider it an epxression.
// But parser can't know if an identifier, let's say "x" in "x + ...", is an assignment target or expression until it sees "=".
// So we parse as expression first, then convert to assignment target if "=" found.
func (p *Parser) assignment() (ast.Expression, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	// assign token means signal for assignment
	if p.matchCurrentToken(token.ASSIGN) {
		equals := p.getPreviousToken()

		// recursively parse the right-hand side (r-value)
		// This allows chained assignment: a = b = c = 10;
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		// The left-hand side (expr) must be a valid assignment target (l-value) that can hold a binding to a value
		// For now, only Variable expressions are valid targets.
		if variable, ok := expr.(*ast.VariableExpression); ok {
			name := variable.Name
			return &ast.AssignmentExpression{Name: name, Value: value}, nil
		}

		return nil, compileError.NewParseError(equals, "Unable to perform assignment on expression "+equals.Lexeme)
	}

	return expr, nil
}

// expression is the entry point for parsing any expression.
func (p *Parser) expression() (ast.Expression, error) {
	return p.assignment()
}
