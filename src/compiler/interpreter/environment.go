/*
Environment manages the runtime state of variables during program execution.
Each variable name (identifier) is bound to a value in the environment's symbol table.

EXAMPLES:
@SET x = 5       → Define("x", 5)
x                → Get("x") returns 5
x = 10           → Assign("x", 10)

Currently implements a single global scope. All variables are accessible from anywhere.
Future: Will support nested scopes (functions, blocks) using parent environment links.
*/
package interpreter

import (
	runtimeError "docklett/compiler/error"
	"docklett/compiler/token"
	"fmt"
)

// Environment stores variable bindings (name → value mappings) for the interpreter.
// Implements a symbol table that tracks variable state during program execution.
type Environment struct {
	Map       map[string]any // Symbol table mapping variable names to runtime values
}

// Define creates a new variable binding or overwrites an existing one.
// This operation never fails - it always succeeds in creating/updating the binding.
func (env *Environment) Define(name string, value any) {
	env.Map[name] = value
}

// Get retrieves the value bound to a variable name.
// Creates RuntimeError if trying to retrieve an undefined variable.
func (env *Environment) Get(name token.Token) any {
	val, ok := env.Map[name.Lexeme]
	if !ok {
		runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
	}
	return val
}

// Assign updates the value of an existing variable.
// Creates RuntimeError if trying to update an undefined variable.
// This distinction prevents typos: x = 5 fails if x wasn't declared.
func (env *Environment) Assign(name token.Token, value any) {
	_, ok := env.Map[name.Lexeme]
	if ok {
		env.Map[name.Lexeme] = value
		return
	}
	runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}
