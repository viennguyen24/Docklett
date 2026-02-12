package error

import (
	"docklett/compiler/ast"
	"docklett/compiler/token"
	"fmt"
)

type RuntimeError interface {
	error
	GetLine() int
	GetExpression() ast.Expression
}

// InterpreterError represents runtime execution errors
type InterpreterError struct {
	Expression ast.Expression
	Message    string
}

func (e *InterpreterError) Error() string {
	line := getExpressionLine(e.Expression)
	if line > 0 {
		return fmt.Sprintf("Runtime Error: [line %d] %s", line, e.Message)
	}
	return fmt.Sprintf("Runtime Error: %s", e.Message)
}

func (e *InterpreterError) GetLine() int {
	return getExpressionLine(e.Expression)
}

func (e *InterpreterError) GetExpression() ast.Expression {
	return e.Expression
}

func NewInterpreterError(expr ast.Expression, message string) *InterpreterError {
	return &InterpreterError{
		Expression: expr,
		Message:    message,
	}
}

// NewRuntimeErrorFromToken creates a runtime error from a token context (for use in environment)
func NewRuntimeErrorFromToken(tok token.Token, message string) error {
	line := tok.Position.Line
	if line > 0 {
		return fmt.Errorf("Runtime Error: [line %d] %s", line, message)
	}
	return fmt.Errorf("Runtime Error: %s", message)
}

// PanicRuntimeError panics with a runtime error
func PanicRuntimeError(tok token.Token, message string) {
	panic(NewRuntimeErrorFromToken(tok, message))
}

// getExpressionLine extracts line number from expression's first token
func getExpressionLine(expr ast.Expression) int {
	switch e := expr.(type) {
	case *ast.LiteralExpression:
		return e.Token.Position.Line
	case *ast.Unary:
		return e.Operator.Position.Line
	case *ast.Binary:
		return e.Operator.Position.Line
	case *ast.Grouping:
		return getExpressionLine(e.Expression)
	case *ast.VariableExpression:
		return e.Name.Position.Line
	default:
		return 0
	}
}
