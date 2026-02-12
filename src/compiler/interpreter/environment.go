/*
Environment stores variable states.
A variable declaration binds a name (identifier) to a value (expression)
Once done, a variable expression access that binding later to get the value
The bindings that associate variables to values need to be stored in an environment
*/
package interpreter

import (
	runtimeError "docklett/compiler/error"
	"docklett/compiler/token"
	"fmt"
)

type Environment struct {
	Map map[string]any
}

// Define doesn't need the Token because it never fails, we adds/overwrites the binding
func (env *Environment) Define(name string, value any) {
	env.Map[name] = value
}

// Get needs a Token because it might fail and need to report an error with line information
func (env *Environment) Get(name token.Token) any {
	val, ok := env.Map[name.Lexeme]
	if !ok {
		runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
	}
	return val
}

func (env *Environment) Assign(name token.Token, value any) {
	_, ok := env.Map[name.Lexeme]
	if ok {
		env.Map[name.Lexeme] = value
		return
	}
	runtimeError.PanicRuntimeError(name, fmt.Sprintf("undefined variable '%s'", name.Lexeme))
}
