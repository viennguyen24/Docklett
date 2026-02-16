/*
STATEMENTS execute actions and produce side effects (unlike expressions which produce values).
Each statement performs an operation on program state but doesn't return a usable value (sometimes code return something to adhere to Golang interface but don't actually need it)

Statement Types:
  - ExpressionStatement: Wraps an expression to execute it for side effects
  - VariableDeclarationStatement: Creates a new variable binding in the current scope

EXAMPLES:

	@SET x = 5              → VariableDeclarationStatement (creates binding)
	x = x + 1               → ExpressionStatement containing AssignmentExpression
	5 + 3                   → ExpressionStatement (discards result)

USAGE:

	Statements are produced by the Parser and executed sequentially by the Interpreter.
	Each statement modifies program state (environment, output, etc.) as a side effect.
*/
package ast

import "docklett/compiler/token"

type Statement interface {
	Accept(visitor StatementVisitor) (any, error)
}

// ExpressionStatement wraps an expression to execute it as a statement.
// This allows expressions with side effects to stand alone as complete statements by evaluating the expression and discard the resulting value if not needed (we only care about side effects)
//
// Without ExpressionStatement, you couldn't use expressions as standalone statements.
// Assignments, function calls, and other side-effecting expressions need this wrapper.
//
// Example:
//
//	Source: x = 5        (assignment expression executed for side effect)
//	AST: ExpressionStatement(AssignmentExpression("x", LiteralExpression(5)))
//	Execution: Evaluates assignment → Updates x → Discards returned value
//
//	Source: i + 1        (expression with no side effect - result discarded)
//	AST: ExpressionStatement(BinaryExpression(VariableExpression(i), +, LiteralExpression(1)))
//	Execution: Evaluates to value → Value discarded (useless but valid)
//
// Used in: All standalone expressions, especially assignments and function calls
type ExpressionStatement struct {
	Expression Expression // The expression to evaluate for side effects
}

func (es *ExpressionStatement) Accept(visitor StatementVisitor) (any, error) {
	return visitor.VisitExpressionStatement(es)
}

// VariableDeclarationStatement creates a new variable binding in the current scope.
// This is a declaration that introduces a NEW identifier (unlike AssignmentExpression which updates existing).
// Declarations create bindings without producing usable values and variables must be declared separately before they can be used (assigned).
// This prevents chaos like: print((var x = 5) + x)  // When does x become available?
//   - Creates a new entry in the environment's symbol table
//   - Binds Name to the result of evaluating Initializer
//   - If Initializer is nil, binds to nil (uninitialized variable)
//
// DECLARATION vs ASSIGNMENT:
//
//	@SET x = 5    → VariableDeclarationStatement (creates NEW binding)
//	x = 10        → AssignmentExpression (updates EXISTING binding)
//	@SET x = 5; x = 10;  → First line creates, second line updates
//
// Example:
//
//	Source: @SET x = 10 + 20
//	AST: VariableDeclarationStatement(Name=\"x\", Initializer=BinaryExpression(10, +, 20))
//	Execution: Evaluate 10 + 20 = 30 → Create binding x → Assign 30 to x
//
//	Source: @SET y
//	AST: VariableDeclarationStatement(Name=\"y\", Initializer=nil)
//	Execution: Create binding y → Assign nil to y
type VariableDeclarationStatement struct {
	Name        token.Token // The identifier token for the new variable
	Initializer Expression  // Expression to evaluate for initial value (nil for uninitialized)
}

func (varStmt *VariableDeclarationStatement) Accept(visitor StatementVisitor) (any, error) {
	return visitor.VisitVarDeclarationStatement(varStmt)
}
