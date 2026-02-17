/*
  The following defines all node interfaces for an instruction and Recursive Descent Parser implementation to verify that an instruction
  follows Docklett's grammar rules.
  The syntax rules is defined in the root design folder

	Expressions: Code that evaluates to a value
	Statements: Code that performs an action/side effect
*/

package parser

import (
	"docklett/compiler/ast"
	"docklett/compiler/token"
)

// declaration wraps statement parsing with panic-mode error recovery.
// On encountering error token, synchronize to skip to a safe token boundary, then returns the error up to Parse() which collects it.
func (p *Parser) declaration() (ast.Statement, error) {
	var stmt ast.Statement
	var err error

	if p.matchCurrentToken(token.SET) {
		stmt, err = p.variableDeclaration()
	} else {
		stmt, err = p.statement()
	}

	if err != nil {
		p.synchronize()
		return nil, err
	}
	return stmt, nil
}

// varDecl parses variable declarations in the form "var IDENTIFIER = expression;".
// We need a separate rule for variable declarations (rather than treating them as
// expressions) because:
//
// 1. Declarations are statements that create side effects (adding to symbol table)
//    without producing a value to be consumed, unlike expressions which always yield a value.
//
// 2. Variable declarations introduce NEW bindings in the environment, which is
//    fundamentally different from expressions that compute values from existing bindings
//
// 3. A variable must be given a scope clearly to define where it can be used. Something like this causes chaos because we dont know when does x become registered and how
//    long will it live:
//    print (var x = 5) + x;

func (p *Parser) variableDeclaration() (ast.Statement, error) {
	identifier, errIdentifier := p.consumeMatchingToken(token.IDENTIFIER, "Expect identifier after SET variable declaration")
	if errIdentifier != nil {
		return nil, errIdentifier
	}

	var expression ast.Expression = nil
	if p.matchCurrentToken(token.ASSIGN) {
		var err error
		expression, err = p.expression()
		if err != nil {
			return nil, err
		}
	}

	return &ast.VariableDeclarationStatement{Name: identifier, Initializer: expression}, nil

}

func (p *Parser) statement() (ast.Statement, error) {
	// Todo: add other statements here
	// Block statements
	if p.matchCurrentToken(token.IF) {
		return p.ifStatement()
	}
	if p.matchCurrentToken(token.FOR) {
		return p.blockStatement()
	}
	return p.expressionStatement()
}

// We need a separate expressionStatement to wrap expression, because some operations are expressions that we want to execute as standalone statements.
// We effectively allow expressions to stand alone

// for example, without expressionStatement:
// 1.
// i++; // Error! Expression, not statement

// 2. You have an expression that returns a value, but you don't care about it
// x = calculate() + doSomethingElse();
// Both calculate() and doSomethingElse() return values
// But you might want to call them just for side effects

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

// statementList parses a sequence of declarations until it encounters one of the
// provided 'terminator' tokens. This centralizes our "Greedy Collection" logic
// used by IF, ELSE, and FOR blocks.
func (p *Parser) collectStatements(terminators ...token.TokenType) ([]ast.Statement, error) {
	var statements []ast.Statement

	// Continue parsing as long as we haven't hit a terminator or the EOF
	for !p.isAtEnd() && !p.checkCurrentToken(terminators...) {
		stmt, err := p.declaration()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

func (p *Parser) blockStatement() (ast.Statement, error) {
	statements, err := p.collectStatements()
	if err != nil {
		return nil, err
	}
	_, err = p.consumeMatchingToken(token.END, "END directive expected after block declaration.")
	if err != nil {
		return nil, err
	}
	return &ast.BlockStatement{Statements: statements}, nil
}

func (p *Parser) ifStatement() (ast.Statement, error) {
	var elseBranch ast.Statement
	// Parse the boolean guard that decides true vs false path
	// The IF or ELIF token is already consumed.
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}

	_, err = p.consumeMatchingToken(token.NLINE, "Expected newline after if condition.")
	if err != nil {
		return nil, err
	}

	// the IfStatement must proactively collect statements until it hits
	// a control-flow keyword.
	thenStatements, err := p.collectStatements(token.ELIF, token.ELSE, token.END)
	if err != nil {
		return nil, err
	}
	thenBranch := &ast.BlockStatement{Statements: thenStatements}

	// If an ELIF is found, we recurse. This creates the 'ElseBranch' link
	// to a new IfStatement, continuing the chain.
	if p.matchCurrentToken(token.ELIF) {
		elseBranch, err = p.ifStatement()
		if err != nil {
			return nil, err
		}
		// The deeply nested call already consumed the END so we can immediately return.
		return &ast.IfStatement{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
	}

	// ELSE: collect all statements in the false-path and group them into a single block
	// If an ELSE is found, we collect the final "catch-all" block.
	if p.matchCurrentToken(token.ELSE) {
		_, err = p.consumeMatchingToken(token.NLINE, "Expected newline after ELSE.")
		if err != nil {
			return nil, err
		}

		elseStatements, err := p.collectStatements(token.END)
		if err != nil {
			return nil, err
		}
		elseBranch = &ast.BlockStatement{Statements: elseStatements}
	}

	// Every conditional chain (no matter how many ELIFs) must end with exactly one 'END'.
	_, err = p.consumeMatchingToken(token.END, "END directive expected after if block.")
	if err != nil {
		return nil, err
	}

	return &ast.IfStatement{Condition: condition, ThenBranch: thenBranch, ElseBranch: elseBranch}, nil
}
