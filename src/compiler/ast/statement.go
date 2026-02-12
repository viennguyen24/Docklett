package ast

import "docklett/compiler/token"

type Statement interface {
	Accept(visitor StatementVisitor) (any, error)
}

type ExpressionStatement struct {
	Expression Expression
}

func (es *ExpressionStatement) Accept(visitor StatementVisitor) (any, error) {
	return visitor.VisitExpressionStatement(es)
}

// e.g @SET x = 10 + 20
type VariableStatement struct {
	Name        token.Token // the variable "name"
	Initializer Expression  // the expression that gives the variable a value
}

func (varStmt *VariableStatement) Accept(visitor StatementVisitor) (any, error) {
	return visitor.VisitVarStatement(varStmt)
}
