package typechecker

// Environment represents a lexical scope used during typechecking.
type Environment struct {
	parent  *Environment
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

// ForEach walks all bindings in lexical order (outer scopes first).
func (e *Environment) ForEach(fn func(string, Type)) {
	if e == nil || fn == nil {
		return
	}
	if e.parent != nil {
		e.parent.ForEach(fn)
	}
	for name, typ := range e.symbols {
		fn(name, typ)
	}
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

// Clone creates a shallow copy of the environment without preserving the parent chain.
func (e *Environment) Clone() *Environment {
	if e == nil {
		return nil
	}
	clone := NewEnvironment(nil)
	e.ForEach(func(name string, typ Type) {
		clone.Define(name, typ)
	})
	return clone
}

// Extend returns a child environment.
func (e *Environment) Extend() *Environment {
	return NewEnvironment(e)
}
