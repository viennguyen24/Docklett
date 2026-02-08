/*
  The following defines all node interfaces for an instruction and Recursive Descent Parser implementation to verify that an instruction
  follows Docklett's grammar rules.
  The syntax rules is defined in the root design folder
*/

package parser

import (
	"docklett/compiler/ast"
	"docklett/compiler/token"
)

// Todo: add synchronize
func (p *Parser) declaration() (ast.Statement, error) {
	isVarDeclaration := p.matchCurrentToken(token.SET)

	if isVarDeclaration {
		return p.variableDeclaration()
	}

	return p.statement()
}

func (p *Parser) variableDeclaration() (ast.Statement, error) {
	identifier, errIdentifier := p.consumeMatchingToken(token.IDENTIFIER, "Expect identifier after SET variable declaration")
	if errIdentifier != nil {
		return nil, errIdentifier
	}

	var expression ast.Expression
	if p.matchCurrentToken(token.ASSIGN) {
		var err error
		expression, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	return &ast.VarDeclareStatement{Identifier: identifier, Expression: expression}, nil

}

func (p *Parser) statement() (ast.Statement, error) {
	// Todo: add other statements here
	return p.expressionStatement()
}

func (p *Parser) expressionStatement() (ast.Statement, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	// a single expression must end the instruction (reminder instructions can be multi line)
	_, err = p.consumeMatchingToken(token.NLINE, "Expected newline after expression to signal end of instruction.")
	if err != nil {
		return nil, err
	}
	return &ast.ExpressionStatement{Expression: expr}, nil
}
