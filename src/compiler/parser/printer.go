package parser

import (
	"docklett/compiler/ast"
	"docklett/compiler/token"
	"fmt"
	"strings"
)

type TreePrinter struct {
	isLastChild []bool
}

func NewTreePrinter() *TreePrinter {
	return &TreePrinter{
		isLastChild: []bool{},
	}
}

func (tp *TreePrinter) getIndent(isLast bool, isField bool) string {
	var result strings.Builder

	for i := 0; i < len(tp.isLastChild); i++ {
		if tp.isLastChild[i] {
			result.WriteString("  ")
		} else {
			result.WriteString("│ ")
		}
	}

	if isField {
		if isLast {
			result.WriteString("└─")
		} else {
			result.WriteString("├─")
		}
	}

	return result.String()
}

func (tp *TreePrinter) formatToken(tok token.Token) string {
	typeName := token.TokenTypeNames[tok.Type]
	if typeName == "" {
		typeName = fmt.Sprintf("UNKNOWN(%d)", tok.Type)
	}

	base := fmt.Sprintf("%s [%s] @Line:%d,Col:%d",
		typeName, tok.Lexeme, tok.Position.Line, tok.Position.Col)

	if tok.Literal != nil {
		base += fmt.Sprintf(" (Literal: %v)", tok.Literal)
	}

	return base
}

func (tp *TreePrinter) VisitLiteralExpr(literal *ast.LiteralExpression) (any, error) {
	result := "LiteralExpression\n"
	prefix := tp.getIndent(true, true)
	result += prefix + fmt.Sprintf("Value: %v\n", literal.Value)
	return result, nil
}

func (tp *TreePrinter) VisitBinaryExpr(binary *ast.BinaryExpression) (any, error) {
	result := "Binary\n"

	prefix := tp.getIndent(false, true)
	result += prefix + "Left: "
	tp.isLastChild = append(tp.isLastChild, false)
	leftResult, err := binary.Left.Accept(tp)
	if err != nil {
		return nil, err
	}
	result += leftResult.(string)
	tp.isLastChild = tp.isLastChild[:len(tp.isLastChild)-1]

	prefix = tp.getIndent(false, true)
	result += prefix + "Operator: " + tp.formatToken(binary.Operator) + "\n"

	prefix = tp.getIndent(true, true)
	result += prefix + "Right: "
	tp.isLastChild = append(tp.isLastChild, true)
	rightResult, err := binary.Right.Accept(tp)
	if err != nil {
		return nil, err
	}
	result += rightResult.(string)
	tp.isLastChild = tp.isLastChild[:len(tp.isLastChild)-1]

	return result, nil
}

func (tp *TreePrinter) VisitUnaryExpr(unary *ast.UnaryExpression) (any, error) {
	result := "Unary\n"

	prefix := tp.getIndent(false, true)
	result += prefix + "Operator: " + tp.formatToken(unary.Operator) + "\n"

	prefix = tp.getIndent(true, true)
	result += prefix + "Right: "
	tp.isLastChild = append(tp.isLastChild, true)
	rightResult, err := unary.Right.Accept(tp)
	if err != nil {
		return nil, err
	}
	result += rightResult.(string)
	tp.isLastChild = tp.isLastChild[:len(tp.isLastChild)-1]

	return result, nil
}

func (tp *TreePrinter) VisitGroupingExpr(grouping *ast.GroupingExpression) (any, error) {
	result := "Grouping\n"

	prefix := tp.getIndent(true, true)
	result += prefix + "Expression: "
	tp.isLastChild = append(tp.isLastChild, true)
	exprResult, err := grouping.Expression.Accept(tp)
	if err != nil {
		return nil, err
	}
	result += exprResult.(string)
	tp.isLastChild = tp.isLastChild[:len(tp.isLastChild)-1]

	return result, nil
}

func (tp *TreePrinter) VisitVariableExpr(variable *ast.VariableExpression) (any, error) {
	return nil, nil
}

func (tp *TreePrinter) VisitAssignmentExpr(assignment *ast.AssignmentExpression) (any, error) {
	return nil, nil
}

func PrintAST(expr ast.Expression) {
	printer := NewTreePrinter()
	result, err := expr.Accept(printer)
	if err != nil {
		fmt.Printf("Error printing AST: %v\n", err)
		return
	}
	fmt.Println(result.(string))
}

func DemoPrinter() {
	// (5 + 3) * 2
	expr := &ast.BinaryExpression{
		Left: &ast.GroupingExpression{
			Expression: &ast.BinaryExpression{
				Left: &ast.LiteralExpression{Value: 5},
				Operator: token.Token{
					Type:     token.ADD,
					Lexeme:   "+",
					Position: token.Position{Line: 1, Col: 3},
				},
				Right: &ast.LiteralExpression{Value: 3},
			},
		},
		Operator: token.Token{
			Type:     token.MULTI,
			Lexeme:   "*",
			Position: token.Position{Line: 1, Col: 8},
		},
		Right: &ast.LiteralExpression{Value: 2},
	}

	PrintAST(expr)
}
