/*
Translator walks the AST produced by the parser and constructs a BuildKit LLB state graph.

COMPILATION FLOW:

	Source → Scanner → Parser → [Translator] → LLB State Graph → Marshal → BuildKit

STATEMENT DISPATCH:

	Each AST statement type is handled by a visitor method that either:
	- Mutates the LLB state (DockerStatement: FROM, RUN, COPY, etc.)
	- Binds compile-time variables (VariableDeclarationStatement)
	- Controls code generation flow (IfStatement, ForStatement)

IMMUTABILITY RULE:

	llb.State is a value type. Every method returns a NEW state, never mutates in place.
*/
package translator

import (
	"docklett/compiler/ast"
	"fmt"
)

type Translator struct {
	env         *Environment // variable scope
	maxLoopIter int          // guard against infinite loop unrolling (default: 10000)
	errors      []error      // collected translation errors
}

func NewTranslator() *Translator {
	return &Translator{
		env:         NewEnvironment(nil),
		maxLoopIter: 10000,
	}
}

// Translate processes the full AST and produces an LLB state graph.
// Returns collected errors if any statement fails translation.
func (t *Translator) Translate(statements []ast.Statement) error {
	for _, stmt := range statements {
		_, err := t.execute(stmt)
		if err != nil {
			t.errors = append(t.errors, err)
		}
	}
	if len(t.errors) > 0 {
		return fmt.Errorf("translation failed with %d error(s): %v", len(t.errors), t.errors)
	}
	return nil
}

// execute dispatches a statement to its corresponding visitor method
func (t *Translator) execute(statement ast.Statement) (any, error) {
	return statement.Accept(t)
}
