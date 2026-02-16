/*
The interpreter traverses the AST produced by the parser and executes each statement
to produce side effects (variable assignments, output, etc.).

EVALUATION vs EXECUTION:
  - Expressions are EVALUATED to produce values
  - Statements are EXECUTED to produce side effects

EXAMPLE EXECUTION FLOW:
  Source: @SET x = 5\n  Parse: VariableStatement(Name="x", Initializer=Literal(5))
  Execute: VisitVarDeclarationStatement → Evaluate initializer → Define in environment
  Result: Environment.Map["x"] = 5

*/

package interpreter

import (
	"docklett/compiler/ast"
)

// Compile-time check to ensure Interpreter implements StatementVisitor
var _ ast.StatementVisitor = (*Interpreter)(nil)

// Main entry point to start interpreting
func (i *Interpreter) interpret(statements []ast.Statement) error {
	for _, stmt := range statements {
		// statements produce an effect rather than value, so for now we ignore the return value and perform the instruction in place
		_, err := i.execute(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// Call the corresponding handling logic base on type of statement
// This is just an entry point for the interpreter to implement its specific logic for a statement
func (i *Interpreter) execute(statement ast.Statement) (any, error) {
	return statement.Accept(i)
}

// placeholder for the base Statement interface.
func (i *Interpreter) VisitStatement(statement *ast.Statement) (any, error) {
	return nil, nil
}

// VisitExpressionStatement evaluates an expression for its side effects (like assignment).
// The expression's return value is discarded - we only care about state changes.
//
// Functionality:
//   - Evaluates the wrapped expression
//   - Discards the resulting value
//   - Captures side effects (assignments, function calls)
//
// This allows expressions to stand alone as statements. Without this, you couldn't write:
//   - Standalone assignments: x = 5
//   - Function calls: print("hello")
//
// Examples:
//
//	Source: x = 10
//	AST: ExpressionStatement(AssignmentExpression("x", Literal(10)))
//	Execute: Evaluate assignment → Environment.Assign("x", 10) → Discard return value (10)
//
//	Source: 5 + 3  (useless but valid)
//	AST: ExpressionStatement(BinaryExpression(5, +, 3))
//	Execute: Evaluate → 8 → Discard result
func (i *Interpreter) VisitExpressionStatement(expressionStatement *ast.ExpressionStatement) (any, error) {
	return i.evaluate(expressionStatement.Expression)
}

// VisitVarDeclarationStatement creates a new variable binding in the environment.
// Evaluates the initializer (initial value expression) if present and binds the value to the variable name.
//
//   - Evaluates Initializer expression (if not nil)
//
//   - Defaults to nil if no initializer provided
//
//   - Creates binding in environment via Define()
//
//   - Define() never fails (overwrites existing bindings)
//
//     @SET x = 5    → VisitVarDeclarationStatement (creates NEW binding)
//
//     x = 10        → VisitAssignmentExpr (updates EXISTING binding, fails if undefined)
func (i *Interpreter) VisitVarDeclarationStatement(varStatement *ast.VariableDeclarationStatement) (any, error) {
	var value any = nil

	if varStatement.Initializer != nil {
		val, err := i.evaluate(varStatement.Initializer)
		if err != nil {
			return nil, err
		}
		value = val
	}

	i.Environment.Define(varStatement.Name.Lexeme, value)
	return nil, nil
}

// VisitBlockStatement creates a new child scope and executes every statements within that block.
// After execution, the scope is discarded and the previous scope is restored.
func (i *Interpreter) VisitBlockStatement(blStatement *ast.BlockStatement) (any, error) {
	blockEnvironment := Environment{
		Map:       make(map[string]any),
		Enclosing: &i.Environment,
	}
	return i.executeBlock(blStatement.Statements, blockEnvironment)
}

func (i *Interpreter) executeBlock(statements []ast.Statement, environment Environment) (any, error) {
	// restore environment after executing block
	previous := i.Environment
	defer func() { i.Environment = previous }()

	i.Environment = environment

	for _, statement := range statements {
		_, err := i.execute(statement)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}