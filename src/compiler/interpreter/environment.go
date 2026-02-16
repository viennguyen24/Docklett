/*
Environment manages variable bindings using a parent-pointer tree (scope chain).
Each environment represents one scope and links to its immediately enclosing scope.

SCOPE CHAIN LOOKUP:
Variables are resolved by walking the chain from innermost to outermost scope:
 1. Check current environment
 2. If not found, recursively check parent (Enclosing)
 3. If not found in any scope, panic with RuntimeError

EXAMPLES:

	@SET x = 5       → Define("x", 5) in current scope
	x                → Get("x") walks chain: current → parent → ... → global
	x = 10           → Assign("x", 10) walks chain to find existing binding

SHADOWING:
Inner scopes can shadow outer variables with the same name:

	var x = "outer";
	{ var x = "inner"; }  // shadows outer x
*/
package interpreter

import (
	runtimeError "docklett/compiler/error"
	"docklett/compiler/token"
	"fmt"
)

// Environment stores variable bindings for one scope level.
// Forms a linked list (scope chain) via Enclosing pointer to parent scope.
type Environment struct {
	Map       map[string]any // Variables defined in THIS scope only
	Enclosing *Environment   // Parent scope (nil for global scope)
}

// Define creates a new variable in the CURRENT scope (does not walk chain).
// Allows shadowing: defining a variable that exists in parent scope creates a NEW binding in current scope.
func (env *Environment) Define(name string, value any) {
	env.Map[name] = value
}

// Get retrieves a variable's value by walking the scope chain.
// Searches current scope first, then recursively searches parent scopes.
func (env *Environment) Get(name token.Token) any {
	// Check current scope first
	val, ok := env.Map[name.Lexeme]
	if ok {
		return val
	}

	// Variable not in current scope - check parent scope
	if env.Enclosing != nil {
		return env.Enclosing.Get(name)
	}

	runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
	return nil // unreachable
}

// Assign updates an EXISTING variable by walking the scope chain.
// Unlike Define, this requires the variable to already exist somewhere in the chain.
func (env *Environment) Assign(name token.Token, value any) {
	// Check if variable exists in current scope
	_, ok := env.Map[name.Lexeme]
	if ok {
		env.Map[name.Lexeme] = value
		return
	}

	// Variable not in current scope - try parent scope
	if env.Enclosing != nil {
		env.Enclosing.Assign(name, value)
		return
	}
	runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}
