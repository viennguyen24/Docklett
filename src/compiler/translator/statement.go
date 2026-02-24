/*
Statement visitor methods for the Translator.
Each method handles one AST statement type and produces LLB state mutations.
*/
package translator

import (
	"docklett/compiler/ast"
	"fmt"
)

// Compile-time check to ensure Translator implements StatementVisitor
var _ ast.StatementVisitor = (*Translator)(nil)

// isTruthy determines boolean value of any value.
func (t *Translator) isTruthy(value any) bool {
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
		return value != nil
	}
}

// VisitStatement is a placeholder for the base Statement interface
func (t *Translator) VisitStatement(statement *ast.Statement) (any, error) {
	return nil, nil
}

// VisitExpressionStatement evaluates the expression for side effects.
func (t *Translator) VisitExpressionStatement(stmt *ast.ExpressionStatement) (any, error) {
	return nil, nil
}

// VisitVarDeclarationStatement binds a variable in the translator's environment.
// The initializer is evaluated and the result stored for later interpolation.
func (t *Translator) VisitVarDeclarationStatement(stmt *ast.VariableDeclarationStatement) (any, error) {
	var value any = nil
	if stmt.Initializer != nil {
		val, err := t.evaluateExpression(stmt.Initializer)
		if err != nil {
			return nil, err
		}
		value = val
	}
	t.env.Define(stmt.Name.Lexeme, value)
	return nil, nil
}

// VisitBlockStatement creates a child scope and translates all statements within it.
// The child scope is discarded after the block completes.
func (t *Translator) VisitBlockStatement(stmt *ast.BlockStatement) (any, error) {
	childEnv := NewEnvironment(t.env)
	previousEnv := t.env
	defer func() { t.env = previousEnv }()
	t.env = childEnv

	for _, s := range stmt.Statements {
		if _, err := t.execute(s); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// VisitIfStatement evaluates the condition and only translates the taken branch,
// producing LLB nodes for that path only.
func (t *Translator) VisitIfStatement(stmt *ast.IfStatement) (any, error) {
	condVal, err := t.evaluateExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}
	if t.isTruthy(condVal) {
		return t.execute(stmt.ThenBranch)
	}
	if stmt.ElseBranch != nil {
		return t.execute(stmt.ElseBranch)
	}
	return nil, nil
}

// VisitDockerStatement translates a Docker instruction into LLB state operations.
// Delegates to translateDocker for keyword-specific LLB graph construction.
func (t *Translator) VisitDockerStatement(stmt *ast.DockerStatement) (any, error) {
	return nil, t.translateDocker(stmt)
}

// VisitForStatement unrolls the loop at compile time.
// This is just a placeholder implementation, TODO is look up compiler design for looping
// Evaluates the iterable (array literal or range), then for each element:
//  1. Binds the target variable to the element value
//  2. Executes the body (producing LLB nodes)
//  3. Unbinds the target after the loop completes
func (t *Translator) VisitForStatement(stmt *ast.ForStatement) (any, error) {
	iterVal, err := t.evaluateExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	// the iterable must evaluate to a []any slice
	elements, ok := iterVal.([]any)
	if !ok {
		return nil, fmt.Errorf("[line %d] for loop iterable must be an array or range, got %T",
			stmt.Target.Line, iterVal)
	}

	for i, elem := range elements {
		if i >= t.maxLoopIter {
			return nil, fmt.Errorf("[line %d] for loop exceeded maximum iteration limit (%d)",
				stmt.Target.Line, t.maxLoopIter)
		}
		t.env.Define(stmt.Target.Lexeme, elem)
		if _, err := t.execute(stmt.Body); err != nil {
			return nil, err
		}
	}
	t.env.Delete(stmt.Target.Lexeme)
	return nil, nil
}
