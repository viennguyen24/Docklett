/*
Expression evaluation stubs for the Translator.
Placeholder — will be expanded when expression evaluation is needed
for constant folding and variable interpolation.
*/
package translator

import (
	"docklett/compiler/ast"
)

// Compile-time check to ensure Translator implements ExpressionVisitor
var _ ast.ExpressionVisitor = (*Translator)(nil)

func (t *Translator) VisitLiteralExpr(literal *ast.LiteralExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitVariableExpr(variable *ast.VariableExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitUnaryExpr(unary *ast.UnaryExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitBinaryExpr(binary *ast.BinaryExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitGroupingExpr(grouping *ast.GroupingExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitLogicalExpr(logical *ast.LogicalExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitAssignmentExpr(assignment *ast.AssignmentExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitArrayLiteralExpr(array *ast.ArrayLiteralExpression) (any, error) {
	return nil, nil
}

func (t *Translator) VisitRangeExpr(rangeExpr *ast.RangeExpression) (any, error) {
	return nil, nil
}
