package translator

import "docklett/compiler/ast"

// evaluateExpression performs evaluation of an expression for constant folding.
func (t *Translator) evaluateExpression(expr ast.Expression) (any, error) {
	// placeholder — will be expanded when ExpressionVisitor is implemented
	return expr.Accept(t)
}
