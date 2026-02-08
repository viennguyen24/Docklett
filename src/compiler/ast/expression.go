package ast

import (
	"docklett/compiler/token"
)

type Expression interface {
	Accept(visitor ExpressionVisitor) (any, error)
}

type LiteralExpression struct {
	Value any
	Token token.Token
}

func (le *LiteralExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitLiteral(le)
}

type Binary struct {
	Left     Expression
	Right    Expression
	Operator token.Token
}

func (b *Binary) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitBinary(b)
}

type Unary struct {
	Operator token.Token
	Right    Expression
}

func (u *Unary) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitUnary(u)
}

type Grouping struct {
	Expression
}

func (g *Grouping) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitGrouping(g)
}
