package bridge

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func ApplyBinaryOperator(rt *Runtime, op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		if value, handled, err := applyStaticPrimitiveBinaryOperator(op, left, right); handled {
			return value, err
		}
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	value, err := rt.interp.ApplyBinaryOperator(op, materializeBoundaryValue(left), materializeBoundaryValue(right))
	return materializeBoundaryValue(value), err
}

func ApplyUnaryOperator(rt *Runtime, op string, operand runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	value, err := rt.interp.ApplyUnaryOperator(op, materializeBoundaryValue(operand))
	return materializeBoundaryValue(value), err
}

func Range(rt *Runtime, start runtime.Value, end runtime.Value, inclusive bool) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	value, err := rt.interp.EvaluateRangeValues(materializeBoundaryValue(start), materializeBoundaryValue(end), inclusive, env)
	return materializeBoundaryValue(value), err
}

func ResolveIterator(rt *Runtime, iterable runtime.Value) (*runtime.IteratorValue, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	env := rt.currentEnv()
	return rt.interp.ResolveIteratorValue(iterable, env)
}

func Spawn(rt *Runtime, task func(*runtime.Environment) (runtime.Value, error)) (*runtime.FutureValue, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if task == nil {
		return nil, fmt.Errorf("compiler bridge: missing task")
	}
	env := rt.currentEnv()
	future := rt.interp.RunCompiledFuture(env, func(taskEnv *runtime.Environment) (runtime.Value, error) {
		if prev, swapped := SwapEnvIfNeeded(rt, taskEnv); swapped {
			defer RestoreEnvIfNeeded(rt, prev, swapped)
		}
		return task(taskEnv)
	})
	if future == nil {
		return nil, fmt.Errorf("compiler bridge: spawn failed")
	}
	return future, nil
}

func Await(rt *Runtime, expr *ast.AwaitExpression, iterable runtime.Value) (runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if expr == nil {
		return nil, fmt.Errorf("compiler bridge: missing await expression")
	}
	env := rt.currentEnv()
	value, err := rt.interp.AwaitIterable(expr, materializeBoundaryValue(iterable), env)
	return materializeBoundaryValue(value), err
}

func ArrayElements(rt *Runtime, arr *runtime.ArrayValue) ([]runtime.Value, error) {
	if rt == nil || rt.interp == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	values, err := rt.interp.ArrayElements(arr)
	return materializeBoundaryValues(values), err
}

func Cast(rt *Runtime, typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		coerced, ok := matchTypeWithoutInterpreter(typeExpr, value)
		if !ok {
			return nil, fmt.Errorf("cannot cast value to requested type")
		}
		if coerced == nil {
			return runtime.NilValue{}, nil
		}
		return coerced, nil
	}
	coerced, err := rt.interp.CastValueToType(typeExpr, materializeBoundaryValue(value))
	return materializeBoundaryValue(coerced), err
}

// MatchType checks whether a value matches a type expression and returns the coerced value when it does.
func MatchType(rt *Runtime, typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, bool, error) {
	if rt == nil {
		return nil, false, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		coerced, ok := matchTypeWithoutInterpreter(typeExpr, value)
		if !ok {
			return nil, false, nil
		}
		if coerced == nil {
			coerced = runtime.NilValue{}
		}
		return coerced, true, nil
	}
	value = materializeBoundaryValue(value)
	if !rt.interp.MatchesType(typeExpr, value) {
		return nil, false, nil
	}
	coerced, err := rt.interp.CoerceValueToType(typeExpr, value)
	if err != nil {
		return nil, false, err
	}
	if coerced == nil {
		coerced = runtime.NilValue{}
	}
	return materializeBoundaryValue(coerced), true, nil
}

// TypeExpressionFromValue exposes runtime type expression inference for compiler helpers.
func TypeExpressionFromValue(rt *Runtime, value runtime.Value) (ast.TypeExpression, error) {
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		return staticTypeExpressionFromValue(value), nil
	}
	return rt.interp.TypeExpressionFromValue(value), nil
}

// ExpandTypeAliases expands type aliases using the interpreter alias table.
func ExpandTypeAliases(rt *Runtime, expr ast.TypeExpression) (ast.TypeExpression, error) {
	if rt == nil {
		return nil, fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		return expr, nil
	}
	return rt.interp.ExpandTypeAliases(expr), nil
}

// EnsureTypeSatisfiesInterface checks interface constraints using the interpreter.
func EnsureTypeSatisfiesInterface(rt *Runtime, subject ast.TypeExpression, iface ast.TypeExpression, context string) error {
	if rt == nil {
		return fmt.Errorf("compiler bridge: missing interpreter")
	}
	if rt.interp == nil {
		// Static no-bootstrap mode cannot enforce dynamic interface constraints at runtime.
		return nil
	}
	return rt.interp.EnsureTypeSatisfiesInterface(subject, iface, context)
}

// IsKnownConstraintTypeName reports if a type name is known for constraint enforcement.
func IsKnownConstraintTypeName(rt *Runtime, name string) bool {
	if rt == nil || rt.interp == nil {
		return false
	}
	return rt.interp.IsKnownConstraintTypeName(name)
}
