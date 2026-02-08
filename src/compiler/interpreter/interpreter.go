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

/*
Main entry point to start interpreting
*/
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

func (i *Interpreter) VisitExpressionStatement(expressionStatement *ast.ExpressionStatement) (any, error) {
	return i.evaluate(expressionStatement.Expression)
}

func (i *Interpreter) VisitVarDeclareStatement(varDeclareStatement *ast.VarDeclareStatement) (any, error) {
	return nil, nil
}
