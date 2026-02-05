package interpreter

import (
	"docklett/compiler/parser"
	"fmt"
	"os"
)

type InterpreterError struct {
	Expression parser.Expression
	Message    string
}

func (e *InterpreterError) Error() string {
	line := getExpressionLine(e.Expression)
	if line > 0 {
		return fmt.Sprintf("line %d: %s", line, e.Message)
	}
	return e.Message
}

func (i *Interpreter) error(expr parser.Expression, message string) *InterpreterError {
	err := &InterpreterError{
		Message:    message,
		Expression: expr,
	}
	i.reportError(err)
	return err
}

// Extract line number from expression's first token
func getExpressionLine(expr parser.Expression) int {
	switch e := expr.(type) {
	case *parser.LiteralExpression:
		return e.Token.Position.Line
	case *parser.Unary:
		return e.Operator.Position.Line
	case *parser.Binary:
		return e.Operator.Position.Line
	case *parser.Grouping:
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