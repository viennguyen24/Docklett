/*
	Visitor Pattern for AST Traversal

	The AST is a shared domain between Parser, Interpreter, and other tools.
	The Visitor Pattern allows different operations on the AST without modifying AST node definitions.

	Operations are separated from data structures for better Single Responsibility Principle.
	Each visitor implementation defines HOW to process nodes, while nodes define WHAT data they hold.

	Example implementations:
	- TreePrinter (in parser): Prints AST structure
	- Interpreter (in interpreter): Evaluates expressions and executes statements

	Usage:
		expressionNode.Accept(printer) -> "1 + 2"
		expressionNode.Accept(interpreter) -> 3
*/

package ast

type ExpressionVisitor interface {
	VisitLiteral(literal *LiteralExpression) (any, error)
	VisitBinary(binary *Binary) (any, error)
	VisitUnary(unary *Unary) (any, error)
	VisitGrouping(grouping *Grouping) (any, error)
}

type StatementVisitor interface {
	VisitStatement(statement *Statement) (any, error)
	VisitExpressionStatement(expressionStatement *ExpressionStatement) (any, error)
	VisitVarDeclareStatement(varDeclareStatement *VarDeclareStatement) (any, error)
}
