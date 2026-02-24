/*
Environment manages compile-time variable bindings for the Translator.
Same scope chain semantics as the interpreter's Environment:
 1. Check current environment
 2. If not found, recursively check parent (Enclosing)
 3. If not found in any scope, panic — undefined variable is a fatal compile error
*/
package translator

import (
	compileError "docklett/compiler/error"
	"fmt"
)

// Environment stores compile-time variable bindings for one scope level.
// Forms a linked list via Enclosing pointer to parent scope.
type Environment struct {
	Bindings  map[string]any // variables defined in THIS scope only
	Enclosing *Environment   // parent scope (nil for global scope)
}

func NewEnvironment(enclosing *Environment) *Environment {
	return &Environment{
		Bindings:  make(map[string]any),
		Enclosing: enclosing,
	}
}

// Define creates a new variable in the current scope.
// Allows shadowing of outer scope variables.
func (env *Environment) Define(name string, value any) {
	env.Bindings[name] = value
}

// Get retrieves a variable's value by walking the scope chain.
// Panics if variable is undefined — this is a fatal compile-time error.
func (env *Environment) Get(name string) any {
	val, ok := env.Bindings[name]
	if ok {
		return val
	}
	if env.Enclosing != nil {
		return env.Enclosing.Get(name)
	}
	compileError.PanicTranslatorError(0, fmt.Sprintf("undefined variable '%s'", name))
	return nil // unreachable
}

// Assign updates an existing variable by walking the scope chain.
// Panics if variable does not exist in any reachable scope.
func (env *Environment) Assign(name string, value any) {
	_, ok := env.Bindings[name]
	if ok {
		env.Bindings[name] = value
		return
	}
	if env.Enclosing != nil {
		env.Enclosing.Assign(name, value)
		return
	}
	compileError.PanicTranslatorError(0, fmt.Sprintf("undefined variable '%s'", name))
}

// Delete removes a variable from the current scope only.
// Used to clean up loop variables after ForStatement completes.
func (env *Environment) Delete(name string) {
	delete(env.Bindings, name)
}
