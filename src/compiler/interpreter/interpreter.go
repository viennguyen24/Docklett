/*
Interpreter walks the statement trees produced by the parser and then "operates" on the instruction thus actually doing something
to the data structure we have been building

For expression, we need to evaluate them to produce a final value
For a statement, we need to execute them to produce an effect

*/

package interpreter

import (
	"docklett/compiler/ast"
)

// Compile time check to ensure that Interpreter implements StatementVisitor
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

func (i *Interpreter) VisitStatement(statement *ast.Statement) (any, error) {
	return nil, nil
}

// VisitExpressionStatement evaluates an expression to perform a certain effect on the program's state.
// The expression itself (like an assignment x = 5) may reference variables that must exist,
// e.g: x = 5, i = i + 1, functionCall()
func (i *Interpreter) VisitExpressionStatement(expressionStatement *ast.ExpressionStatement) (any, error) {
	return i.evaluate(expressionStatement.Expression)
}

// VisitVarStatement creates a new variable in the environment and adds a binding of an identifier to a value.
// Used for variable declaration, so the variable doesn't need to exist already,
// e.g: var x = 5, var y (initialized to nil)
func (i *Interpreter) VisitVarStatement(varStatement *ast.VariableStatement) (any, error) {
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
