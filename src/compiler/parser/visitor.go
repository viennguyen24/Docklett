/*
	We use Visitor Pattern for AST methods.

	The AST is not a Parser-specific domain, even though the Parser produces it.
	The Tnterpreter uses it to evaluate expressions as well, so it's an interception between the Parser and Interpreter

	However, we can't define specific methods for parsing / evaluating in each AST node. This violates single-responsibility (the methods are tied to the component calling the tree node, not the tree it self)
	and scalability (we would have to modify methods for each node everytime we want to change implementation)

	Hence the Visitor Pattern. (this sits in the Parser folder temporarily until we found a reason to move this into a its own folder)

	The Visitor is an object that defines all cases of functions to apply on different AST nodes.
	We pack all the methods that operate on the AST into the Visitor interface through visit* methods
	For each AST node, we let it choose which method to call on the visitor using an accept method.

	Example:
	A Printer implements Visitor interface, and prints out the string of the node
	A Interpreter implements Visitor interface, and evaluates the expression

	expressionNode.accept(Printer) -> "1 + 2"
	expressionNode.accept(Interpreter) -> 3
*/

package parser

type ExpressionVisitor interface {
	VisitLiteral(literal *LiteralExpression) (any, error)
	VisitBinary(binary *Binary) (any, error)
	VisitUnary(unary *Unary) (any, error)
	VisitGrouping(grouping *Grouping) (any, error)
}
