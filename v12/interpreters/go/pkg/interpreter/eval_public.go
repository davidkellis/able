package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// EvaluateStatementIn evaluates a statement in the provided environment.
// When env is nil, the interpreter global environment is used.
func (i *Interpreter) EvaluateStatementIn(stmt ast.Statement, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if stmt == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = i.GlobalEnvironment()
	}
	if env == nil {
		return nil, fmt.Errorf("missing environment")
	}
	return i.evaluateStatement(stmt, env)
}

// RegisterMethodsDefinitionIn registers a methods definition in the provided environment.
// When env is nil, the interpreter global environment is used.
func (i *Interpreter) RegisterMethodsDefinitionIn(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if def == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = i.GlobalEnvironment()
	}
	if env == nil {
		return nil, fmt.Errorf("missing environment")
	}
	return i.evaluateMethodsDefinition(def, env)
}

// RegisterImplementationDefinitionIn registers an implementation definition in the provided environment.
// When env is nil, the interpreter global environment is used.
func (i *Interpreter) RegisterImplementationDefinitionIn(def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error) {
	if i == nil {
		return nil, fmt.Errorf("missing interpreter")
	}
	if def == nil {
		return runtime.NilValue{}, nil
	}
	if env == nil {
		env = i.GlobalEnvironment()
	}
	if env == nil {
		return nil, fmt.Errorf("missing environment")
	}
	return i.evaluateImplementationDefinition(def, env, false)
}
