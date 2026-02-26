package bridge

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// EvaluateStatement executes a statement through the interpreter in the provided environment.
func EvaluateStatement(rt *Runtime, stmt ast.Statement, env *runtime.Environment) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if stmt == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = rt.Env()
	}
	return rt.interp.EvaluateStatementIn(stmt, env)
}

// RegisterMethodsDefinition registers a methods definition through the interpreter in the provided environment.
func RegisterMethodsDefinition(rt *Runtime, def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if def == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = rt.Env()
	}
	return rt.interp.RegisterMethodsDefinitionIn(def, env)
}

// RegisterImplementationDefinition registers an implementation definition through the interpreter in the provided environment.
func RegisterImplementationDefinition(rt *Runtime, def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if def == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = rt.Env()
	}
	return rt.interp.RegisterImplementationDefinitionIn(def, env)
}
