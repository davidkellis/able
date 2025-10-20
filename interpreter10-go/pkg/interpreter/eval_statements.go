package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateStatement(node ast.Statement, env *runtime.Environment) (runtime.Value, error) {
	switch n := node.(type) {
	case ast.Expression:
		return i.evaluateExpression(n, env)
	case *ast.StructDefinition:
		return i.evaluateStructDefinition(n, env)
	case *ast.MethodsDefinition:
		return i.evaluateMethodsDefinition(n, env)
	case *ast.InterfaceDefinition:
		return i.evaluateInterfaceDefinition(n, env)
	case *ast.ImplementationDefinition:
		return i.evaluateImplementationDefinition(n, env)
	case *ast.FunctionDefinition:
		return i.evaluateFunctionDefinition(n, env)
	case *ast.WhileLoop:
		return i.evaluateWhileLoop(n, env)
	case *ast.ForLoop:
		return i.evaluateForLoop(n, env)
	case *ast.RaiseStatement:
		return i.evaluateRaiseStatement(n, env)
	case *ast.BreakStatement:
		return i.evaluateBreakStatement(n, env)
	case *ast.ContinueStatement:
		return i.evaluateContinueStatement(n, env)
	case *ast.ReturnStatement:
		return i.evaluateReturnStatement(n, env)
	case *ast.RethrowStatement:
		return i.evaluateRethrowStatement(n, env)
	case *ast.PackageStatement:
		return runtime.NilValue{}, nil
	case *ast.ImportStatement:
		return i.evaluateImportStatement(n, env)
	case *ast.DynImportStatement:
		return i.evaluateDynImportStatement(n, env)
	default:
		return nil, fmt.Errorf("unsupported statement type: %s", n.NodeType())
	}
}

func (i *Interpreter) evaluateBlock(block *ast.BlockExpression, env *runtime.Environment) (runtime.Value, error) {
	scope := runtime.NewEnvironment(env)
	var result runtime.Value = runtime.NilValue{}
	for _, stmt := range block.Body {
		val, err := i.evaluateStatement(stmt, scope)
		if err != nil {
			if _, ok := err.(returnSignal); ok {
				return nil, err
			}
			return nil, err
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) evaluateWhileLoop(loop *ast.WhileLoop, env *runtime.Environment) (runtime.Value, error) {
	var result runtime.Value = runtime.NilValue{}
	for {
		cond, err := i.evaluateExpression(loop.Condition, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			return result, nil
		}
		val, err := i.evaluateBlock(loop.Body, env)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label != "" {
					return nil, sig
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("Labeled continue not supported")
				}
				continue
			case raiseSignal:
				return nil, sig
			case returnSignal:
				return nil, sig
			default:
				return nil, err
			}
		}
		result = val
	}
}

func (i *Interpreter) evaluateRaiseStatement(stmt *ast.RaiseStatement, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(stmt.Expression, env)
	if err != nil {
		return nil, err
	}
	errVal := makeErrorValue(val)
	return nil, raiseSignal{value: errVal}
}

func (i *Interpreter) evaluateReturnStatement(stmt *ast.ReturnStatement, env *runtime.Environment) (runtime.Value, error) {
	var result runtime.Value = runtime.NilValue{}
	if stmt.Argument != nil {
		val, err := i.evaluateExpression(stmt.Argument, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return nil, returnSignal{value: result}
}

func (i *Interpreter) evaluateForLoop(loop *ast.ForLoop, env *runtime.Environment) (runtime.Value, error) {
	iterable, err := i.evaluateExpression(loop.Iterable, env)
	if err != nil {
		return nil, err
	}
	bodyEnvBase := runtime.NewEnvironment(env)

	var values []runtime.Value
	switch it := iterable.(type) {
	case *runtime.ArrayValue:
		values = it.Elements
	case *runtime.RangeValue:
		startVal, err := rangeEndpoint(it.Start)
		if err != nil {
			return nil, err
		}
		endVal, err := rangeEndpoint(it.End)
		if err != nil {
			return nil, err
		}
		step := 1
		if endVal < startVal {
			step = -1
		}
		values = make([]runtime.Value, 0)
		for v := startVal; ; v += step {
			if step > 0 {
				if it.Inclusive {
					if v > endVal {
						break
					}
				} else if v >= endVal {
					break
				}
			} else {
				if it.Inclusive {
					if v < endVal {
						break
					}
				} else if v <= endVal {
					break
				}
			}
			values = append(values, runtime.IntegerValue{Val: big.NewInt(int64(v)), TypeSuffix: runtime.IntegerI32})
		}
	default:
		return nil, fmt.Errorf("for-loop iterable must be array or range, got %s", iterable.Kind())
	}

	var result runtime.Value = runtime.NilValue{}
	for _, el := range values {
		iterEnv := runtime.NewEnvironment(bodyEnvBase)
		if err := i.assignPattern(loop.Pattern, el, iterEnv, true); err != nil {
			return nil, err
		}
		val, err := i.evaluateBlock(loop.Body, iterEnv)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label != "" {
					return nil, sig
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("Labeled continue not supported")
				}
				continue
			case raiseSignal:
				return nil, sig
			case returnSignal:
				return nil, sig
			default:
				return nil, err
			}
		}
		result = val
	}
	return result, nil
}

func (i *Interpreter) evaluateBreakStatement(stmt *ast.BreakStatement, env *runtime.Environment) (runtime.Value, error) {
	var val runtime.Value = runtime.NilValue{}
	if stmt.Value != nil {
		var err error
		val, err = i.evaluateExpression(stmt.Value, env)
		if err != nil {
			return nil, err
		}
	}
	label := ""
	if stmt.Label != nil {
		label = stmt.Label.Name
		state := i.stateFromEnv(env)
		if !state.hasBreakpoint(label) {
			return nil, fmt.Errorf("Unknown break label '%s'", label)
		}
	}
	return nil, breakSignal{label: label, value: val}
}

func (i *Interpreter) evaluateContinueStatement(stmt *ast.ContinueStatement, env *runtime.Environment) (runtime.Value, error) {
	label := ""
	if stmt.Label != nil {
		label = stmt.Label.Name
		return nil, fmt.Errorf("Labeled continue not supported")
	}
	return nil, continueSignal{label: label}
}
