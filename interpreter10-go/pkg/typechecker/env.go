package typechecker

// Environment represents a lexical scope used during typechecking.
type Environment struct {
	parent *Environment
	symbols map[string]Type
}

// NewEnvironment creates a new environment with an optional parent.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		parent:  parent,
		symbols: make(map[string]Type),
	}
}

// Define binds a name to a type in the current scope.
func (e *Environment) Define(name string, typ Type) {
	e.symbols[name] = typ
}

// Lookup searches for a name in the current scope chain.
func (e *Environment) Lookup(name string) (Type, bool) {
	if typ, ok := e.symbols[name]; ok {
		return typ, true
	}
	if e.parent != nil {
		return e.parent.Lookup(name)
	}
	return nil, false
}

// Extend returns a child environment.
func (e *Environment) Extend() *Environment {
	return NewEnvironment(e)
}
