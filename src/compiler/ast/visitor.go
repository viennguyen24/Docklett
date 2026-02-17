/*
	We use Visitor Pattern for AST methods.

	The AST is not a Parser-specific domain, even though the Parser produces it.
	The Tnterpreter uses it to evaluate expressions as well, so it's an interception between the Parser and Interpreter
	However, we can't define specific methods for parsing / evaluating in each AST node. This violates single-responsibility (the methods are tied to the component calling the tree node, not the tree it self)
    and scalability (we would have to modify methods for each node everytime we want to change implementation)
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
	VisitVariableExpr(variable *VariableExpression) (any, error)
	VisitLiteralExpr(literal *LiteralExpression) (any, error)
	VisitBinaryExpr(binary *BinaryExpression) (any, error)
	VisitUnaryExpr(unary *UnaryExpression) (any, error)
	VisitGroupingExpr(grouping *GroupingExpression) (any, error)
	VisitLogicalExpr(logical *LogicalExpression) (any, error)
	VisitAssignmentExpr(assignment *AssignmentExpression) (any, error)
}

type StatementVisitor interface {
	VisitStatement(statement *Statement) (any, error)
	VisitExpressionStatement(expressionStatement *ExpressionStatement) (any, error)
	VisitVarDeclarationStatement(varDeclareStatement *VariableDeclarationStatement) (any, error)
	VisitBlockStatement(blockStatement *BlockStatement) (any, error)
	VisitIfStatement(ifStatement *IfStatement) (any, error)
}
