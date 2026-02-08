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

type VarDeclareStatement struct {
	Identifier token.Token
	Expression Expression
}

func (vds *VarDeclareStatement) Accept(visitor StatementVisitor) (any, error) {
	return visitor.VisitVarDeclareStatement(vds)
}
