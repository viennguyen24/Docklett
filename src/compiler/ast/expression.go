/*
EXPRESSIONS represent code that evaluates to a value. They can be composed and nested.
The expression hierarchy from atomic to complex:

	atomic:   Literal (5, "hello", true)
	variable: VariableExpression (x, userId)
	prefix:   UnaryExpression (-, !)
	grouped:  GroupingExpression ((expression))
	binary:   BinaryExpression (+, -, *, /, ==, !=, <, >, <=, >=)
	logic:    LogicalExpression(or / and)
	assign:   AssignmentExpression (x = value)

EXAMPLES:

	5                    → LiteralExpression
	-5                   → UnaryExpression(-, LiteralExpression(5))
	x + 3                → BinaryExpression(VariableExpression(x), +, LiteralExpression(3))
	(1 + 2) * 3          → BinaryExpression(GroupingExpression(BinaryExpression(...)), *, LiteralExpression(3))
	x = y + 1            → AssignmentExpression(x, BinaryExpression(VariableExpression(y), +, LiteralExpression(1)))

USAGE:

	AST nodes are produced by the Parser from tokens, then consumed by interpreter to evaluate expressions to produce runtime values
*/
package ast

import (
	"docklett/compiler/token"
)

type Expression interface {
	Accept(visitor ExpressionVisitor) (any, error)
}

// VariableExpression represents a variable reference in an expression context.
// This is an l-value (can appear on left side of assignment) that looks up a value from the environment.
type VariableExpression struct {
	Name token.Token // Identifier token containing the variable name and position
}

func (ve *VariableExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitVariableExpr(ve)
}

// LiteralExpression represents atomic constant values that directly produce a runtime value.
// This is the terminal node in expression evaluation - it requires no further computation.
type LiteralExpression struct {
	Value any         // The actual runtime value (bool, int, float64, string)
	Token token.Token // Token for error reporting and position tracking
}

func (le *LiteralExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitLiteralExpr(le)
}

// UnaryExpression represents a prefix operator applied to a single operand.
// Evaluation occurs right-to-left: first evaluate Right, then apply Operator.
// Supported Operators:
//   - (negation): inverts boolean value (!true → false)
//     ! (minus): negates numeric value (-5 → -5, -(-5) → 5)
type UnaryExpression struct {
	Operator token.Token // The unary operator (NEGATE or SUBTRACT)
	Right    Expression  // The operand expression to transform
}

func (u *UnaryExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitUnaryExpr(u)
}

// GroupingExpression represents a parenthesized expression that overrides operator precedence.
// Parentheses force inner expression to evaluate first before outer operations.
type GroupingExpression struct {
	Expression // The wrapped expression to evaluate
}

func (g *GroupingExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitGroupingExpr(g)
}

// two arbitrary values combined through + - * /
type BinaryExpression struct {
	Left     Expression  // Left operand expression
	Right    Expression  // Right operand expression
	Operator token.Token // Infix operator token
}

func (b *BinaryExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitBinaryExpr(b)
}

type LogicalExpression struct {
	Left     Expression  // Left operand, always evaluated
	Right    Expression  // Right operand, conditionally evaluated
	Operator token.Token // AND or OR token
}

func (l *LogicalExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitLogicalExpr(l)
}

// AssignmentExpression represents binding a value to an existing variable (not declaration).
// This is a statement-like expression that produces a side effect AND returns a value. This returns value because AssignmentExpression lives inside an ExpressionStatement, and it's returned value would be brought to an effect
// e.g ExpressionStatement(AssignmentExpression("x", 5)).
// Functionality:
//   - Evaluates Value expression
//   - Updates the binding in the environment for Name
//   - Returns the assigned value (enables chained assignments)
//
// Important Distinction:
//   - AssignmentExpression: x = 5     (variable must already exist)
//   - Declaration: @SET x = 5  (creates new variable)
//
// Example:
//
//	Source: x = y + 1
//	AST: AssignmentExpression("x", BinaryExpression(VariableExpression(y), +, LiteralExpression(1)))
//	Evaluation: Look up y → Add 1 → Update x binding → Return result
type AssignmentExpression struct {
	Name  token.Token // Identifier for the target variable
	Value Expression  // Expression to evaluate and assign
}

func (a *AssignmentExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitAssignmentExpr(a)
}

// ArrayLiteralExpression represents an inline list of expressions.
// Used as a ForStatement iterable: @FOR pkg IN ["curl", "git", "vim"]
//
// Example:
//
//	Source:  ["curl", "git"]
//	AST:    ArrayLiteralExpression{Elements: [LiteralExpr("curl"), LiteralExpr("git")]}
type ArrayLiteralExpression struct {
	Bracket  token.Token  // opening [ token for error reporting
	Elements []Expression // ordered list of element expressions
}

func (a *ArrayLiteralExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitArrayLiteralExpr(a)
}

// RangeExpression represents a range() call for generating integer sequences at compile time.
// Used as a ForStatement iterable: @FOR i IN range(0, 5)
//
// Supported forms:
//   - range(start, end)        → [start, start+1, ..., end-1]
//   - range(start, end, step)  → [start, start+step, ..., <end]
//
// All arguments must be integer expressions evaluable at compile time.
// Step defaults to 1 when omitted (nil). Step of 0 is a compile-time error.
//
// Examples:
//
//	range(0, 5)      → [0, 1, 2, 3, 4]
//	range(0, 10, 2)  → [0, 2, 4, 6, 8]
//	range(5, 0, -1)  → [5, 4, 3, 2, 1]
type RangeExpression struct {
	Token token.Token // RANGE keyword token for error reporting
	Start Expression  // inclusive lower bound
	End   Expression  // exclusive upper bound
	Step  Expression  // nil → step of 1
}

func (r *RangeExpression) Accept(visitor ExpressionVisitor) (any, error) {
	return visitor.VisitRangeExpr(r)
}
