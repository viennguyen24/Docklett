package ast

import (
	"docklett/compiler/token"
)

type Expression interface {
	Accept(visitor ExpressionVisitor) (any, error)
}

// An AST node that wraps a token representing a variable name.
// Created during parsing when the parser encounters an identifier in expression context.
type VariableExpression struct {
	Name token.Token
}

func (ve *VariableExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitVariableExpr(ve)
}

// the most atomic unit that actually returns a fundamental value
type LiteralExpression struct {
	Value any
	Token token.Token
}

func (le *LiteralExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitLiteralExpr(le)
}

// a post-order value compensated by an operator of type - or !
type Unary struct {
	Operator token.Token
	Right    Expression
}

func (u *Unary) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitUnaryExpr(u)
}

type Grouping struct {
	Expression
}

func (g *Grouping) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitGroupingExpr(g)
}

// two arbitrary values combined through + - * /
type Binary struct {
	Left     Expression
	Right    Expression
	Operator token.Token
}

func (b *Binary) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitBinaryExpr(b)
}

type Assignment struct {
	Name  token.Token
	Value Expression
}

func (a *Assignment) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitAssignmentExpr(a)
}
