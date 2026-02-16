/*
Evaluates expression AST nodes to produce runtime values using recursive descent.
Each Visit method handles one expression type, delegating to child expressions as needed.

TYPE COERCION RULES:
  1. Numeric operations: promote int → float64 (3 + 2.5 → 5.5)
  2. String concatenation: + operator only ("hello" + " world")
  3. Equality: works across all types (5 == 5.0 → true)
  4. Comparison: numbers and strings only ("a" < "b" uses lexicographic order)
  5. Logical operators: booleans only (! negation, future: &&, ||)
*/

package interpreter

import (
	"docklett/compiler/ast"
	runtimeError "docklett/compiler/error"
	"docklett/compiler/token"
	"fmt"
)

// Compile-time check to ensure Interpreter implements ExpressionVisitor
var _ ast.ExpressionVisitor = (*Interpreter)(nil)

type Interpreter struct {
	Environment Environment
}

// isTruthy determines the boolean value of any runtime value (truthiness).
// Used for conditional logic and boolean coercion.
//   - bool: returns the boolean value itself
//   - nil: false
//   - int/float64: false if zero, true otherwise
//   - string: false if empty "", true otherwise
//   - unknown types: true if not nil
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

func (i *Interpreter) evaluate(expr ast.Expression) (any, error) {
	return expr.Accept(i)
}

// VisitLiteralExpr evaluates a literal expression by returning its stored value.
// This is the terminal case in expression evaluation - no further recursion needed.
func (i *Interpreter) VisitLiteralExpr(literal *ast.LiteralExpression) (any, error) {
	return literal.Value, nil
}

// VisitUnaryExpr evaluates unary operators (prefix operators) by evaluating operand then applying operator.
//
// Supported Operators:
//
//	! (NEGATE): boolean negation, requires boolean operand
//	- (SUBTRACT): numeric negation, requires int or float64 operand
//
// Examples:
//
//	Source: !true
//	Evaluate: Right = true → Apply ! → Result: false
//
//	Source: -5
//	Evaluate: Right = 5 → Apply - → Result: -5
//
//	Source: -"hello"  (type error)
//	Error: "subtraction operation requires number, got string"
func (i *Interpreter) VisitUnaryExpr(unary *ast.UnaryExpression) (any, error) {
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

// VisitGroupingExpr evaluates a grouped expression by evaluating its wrapped expression.
// Grouping exists only for precedence control during parsing - evaluation just unwraps it.
//
// Example:
//
//	Source: (1 + 2) * 3
//	Evaluate: Unwrap grouping → Evaluate 1 + 2 → 3 → Continue with * 3
func (i *Interpreter) VisitGroupingExpr(grouping *ast.GroupingExpression) (any, error) {
	return i.evaluate(grouping.Expression)
}

// VisitBinaryExpr evaluates binary operators by evaluating both operands then applying the operator.
// Implements type-specific behavior for numeric, string, and nil operations.
//
// Type Dispatch Priority:
//  1. Both operands numeric → executeNumeric() (promotes to float64)
//  2. Both operands string → executeString()
//  3. Either operand nil → executeNil() (equality only)
//  4. Otherwise → Error: "mismatched or unsupported types"
//
// Examples:
//
//	Source: 3 + 2
//	Evaluate: 3, 2 → Both numeric → executeNumeric(3.0, 2.0, +) → 5.0
//
//	Source: "hello" + " world"
//	Evaluate: "hello", " world" → Both strings → executeString(..., +) → "hello world"
//
//	Source: 5 + "hello"  (type error)
//	Error: "mismatched or unsupported types: int and string"
func (i *Interpreter) VisitBinaryExpr(binary *ast.BinaryExpression) (any, error) {
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

// VisitVariableExpr evaluates a variable reference by looking up its value in the environment.
//   - Looks up variable name in environment's symbol table
//   - Returns the bound value if found
//   - Panics with RuntimeError if undefined
//
// Example:
//
//	Source: x + 5  (where @SET x = 10 was executed earlier)
//	Evaluate: Get("x") → 10
//
//	Source: y  (y was never declared)
//	Panic: "Runtime Error: [line 10] undefined variable 'y'"
func (i *Interpreter) VisitVariableExpr(variable *ast.VariableExpression) (any, error) {
	return i.Environment.Get(variable.Name), nil
}

// VisitAssignmentExpr evaluates an assignment by evaluating the value and updating the environment.
// Returns the assigned value (enables chained assignments: a = b = c).
//
// Examples:
//
//	Source: x = 20  (where @SET x = 10 was executed earlier)
//	Evaluate: Value = 20 → Assign("x", 20) → Return 20
//
//	Source: a = b = 10  (chained assignment)
//	Evaluate: b = 10 returns 10 → a = 10 returns 10
func (i *Interpreter) VisitAssignmentExpr(assignment *ast.AssignmentExpression) (any, error) {
	val, err := i.evaluate(assignment.Value)
	if err != nil {
		return nil, err
	}
	i.Environment.Assign(assignment.Name, val)
	return val, nil
}
