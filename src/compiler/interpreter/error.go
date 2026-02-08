package interpreter

import (
	"docklett/compiler/ast"
	"fmt"
	"os"
)

type InterpreterError struct {
	Expression ast.Expression
	Message    string
}

func (e *InterpreterError) Error() string {
	line := getExpressionLine(e.Expression)
	if line > 0 {
		return fmt.Sprintf("line %d: %s", line, e.Message)
	}
	return e.Message
}

func (i *Interpreter) error(expr ast.Expression, message string) *InterpreterError {
	err := &InterpreterError{
		Message:    message,
		Expression: expr,
	}
	i.reportError(err)
	return err
}

// Extract line number from expression's first token
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
	default:
		// just a defensive check. this is normally dead code assuming we handle all expression types above
		return 0
	}
}

func (i *Interpreter) reportError(err *InterpreterError) {
	line := getExpressionLine(err.Expression)
	if line > 0 {
		fmt.Fprintf(os.Stderr, "Error at line %d: %s\n", line, err.Message)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Message)
	}
}
