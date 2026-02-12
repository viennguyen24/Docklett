/*
   Type Coercion Rules:
   1. All numeric operations promote to float64
   2. String concatenation with + only
   3. Equality works across all types
   4. Comparison (>, <) only for numbers and strings (lexicographic)
   5. Logical operators (!, &&, ||) only for booleans
*/

package interpreter

import (
	"docklett/compiler/ast"
	runtimeError "docklett/compiler/error"
	"docklett/compiler/token"
	"fmt"
)

// compile time check to ensure that Interpreter implements ExpressionVisitor
var _ ast.ExpressionVisitor = (*Interpreter)(nil)

type Interpreter struct {
	Environment Environment
}

func (i *Interpreter) isTruthy(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case nil:
		return false
	case int:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	default:
		// Unknown types are truthy if not nil
		return value != nil
	}
}

/*
A generic method to
*/
func (i *Interpreter) evaluate(expr ast.Expression) (any, error) {
	return expr.Accept(i)
}

func (i *Interpreter) VisitLiteralExpr(literal *ast.LiteralExpression) (any, error) {
	return literal.Value, nil
}

func (i *Interpreter) VisitUnaryExpr(unary *ast.Unary) (any, error) {
	right, err := i.evaluate(unary.Right)
	if err != nil {
		return nil, err
	}

	switch unary.Operator.Type {
	case token.NEGATE:
		b, ok := right.(bool)
		if !ok {
			return nil, runtimeError.NewInterpreterError(unary, fmt.Sprintf("negate operation requires boolean, got %T", right))
		}
		return !b, nil

	case token.SUBTRACT:
		switch v := right.(type) {
		case int:
			return -v, nil
		case float64:
			return -v, nil
		default:
			return nil, runtimeError.NewInterpreterError(unary, fmt.Sprintf("subtraction operation requires number, got %T", right))
		}

	default:
		return nil, runtimeError.NewInterpreterError(unary, fmt.Sprintf("unknown unary operator: %v", unary.Operator.Lexeme))
	}
}

func (i *Interpreter) VisitGroupingExpr(grouping *ast.Grouping) (any, error) {
	return i.evaluate(grouping.Expression)
}

func (i *Interpreter) VisitBinaryExpr(binary *ast.Binary) (any, error) {
	left, lErr := binary.Left.Accept(i)
	if lErr != nil {
		return nil, lErr
	}

	right, rErr := binary.Right.Accept(i)
	if rErr != nil {
		return nil, rErr
	}

	op := binary.Operator.Type

	lNum, lErr := toFloat(left)
	rNum, rErr := toFloat(right)
	// if either is float, implicitly cast result to float
	if lErr == nil && rErr == nil {
		return i.executeNumeric(binary, lNum, rNum, op)
	}

	// only operate on both string operands
	lStr, lOk := left.(string)
	rStr, rOk := right.(string)
	if lOk && rOk {
		return i.executeString(binary, lStr, rStr, op)
	}

	// ony support equality on nil
	if left == nil || right == nil {
		return i.executeNil(binary, op)
	}

	return nil, runtimeError.NewInterpreterError(binary, fmt.Sprintf("mismatched or unsupported types: %T and %T", left, right))
}

// Only allow operations on numeric types (float, number or float and number)
func (i *Interpreter) executeNumeric(expr ast.Expression, l float64, r float64, op token.TokenType) (any, error) {
	switch op {
	case token.ADD:
		return l + r, nil
	case token.SUBTRACT:
		return l - r, nil
	case token.MULTI:
		return l * r, nil
	case token.DIVIDE:
		if r == 0.0 {
			return nil, runtimeError.NewInterpreterError(expr, "division by zero")
		}
		return l / r, nil
	case token.EQUAL:
		return l == r, nil
	case token.UNEQUAL:
		return l != r, nil
	case token.GREATER:
		return l > r, nil
	case token.GTE:
		return l >= r, nil
	case token.LESS:
		return l < r, nil
	case token.LTE:
		return l <= r, nil
	}
	return nil, runtimeError.NewInterpreterError(expr, fmt.Sprintf("unrecognized numeric operator %v", op))
}

func (i *Interpreter) executeString(expr ast.Expression, l string, r string, op token.TokenType) (any, error) {
	switch op {
	case token.ADD:
		return l + r, nil // Concatenation
	case token.EQUAL:
		return l == r, nil
	case token.UNEQUAL:
		return l != r, nil
	// comparing string base on lexicographic order
	case token.GREATER:
		return l > r, nil
	case token.LESS:
		return l < r, nil
	}
	return nil, runtimeError.NewInterpreterError(expr, fmt.Sprintf("invalid string operator: %v", op))
}

func (i *Interpreter) executeNil(expr ast.Expression, op token.TokenType) (any, error) {
	switch op {
	case token.EQUAL:
		return true, nil
	case token.UNEQUAL:
		return false, nil
	}
	return nil, runtimeError.NewInterpreterError(expr, "nil only supports equality checks")
}

func toFloat(val any) (float64, error) {
	switch v := val.(type) {
	case int:
		return float64(v), nil
	case float64:
		return v, nil
	default:
		return 0, fmt.Errorf("type error: cannot convert %T to float64", val)
	}
}

func (i *Interpreter) Interpret(expr ast.Expression) (any, error) {
	return i.evaluate(expr)
}

func (i *Interpreter) VisitVariableExpr(variable *ast.VariableExpression) (any, error) {
	return i.Environment.Get(variable.Name), nil
}

func (i *Interpreter) VisitAssignmentExpr(assignment *ast.Assignment) (any, error) {
	val, err := i.evaluate(assignment.Value)
	if err != nil {
		return nil, err
	}
	i.Environment.Assign(assignment.Name, val)
	return val, nil
}
