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
