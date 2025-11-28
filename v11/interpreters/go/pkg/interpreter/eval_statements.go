package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateStatement(node ast.Statement, env *runtime.Environment) (runtime.Value, error) {
	switch n := node.(type) {
	case ast.Expression:
		return i.evaluateExpression(n, env)
	case *ast.StructDefinition:
		return i.evaluateStructDefinition(n, env)
	case *ast.UnionDefinition:
		return i.evaluateUnionDefinition(n, env)
	case *ast.TypeAliasDefinition:
		return runtime.NilValue{}, nil
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
	case *ast.YieldStatement:
		gen := i.currentGenerator()
		if gen == nil {
			return nil, fmt.Errorf("yield may only appear inside iterator literal")
		}
		var value runtime.Value = runtime.NilValue{}
		if n.Expression != nil {
			val, err := i.evaluateExpression(n.Expression, env)
			if err != nil {
				return nil, err
			}
			value = val
		}
		if err := gen.emit(value); err != nil {
			return nil, err
		}
		return runtime.NilValue{}, nil
	case *ast.RethrowStatement:
		return i.evaluateRethrowStatement(n, env)
	case *ast.PackageStatement:
		return runtime.NilValue{}, nil
	case *ast.ImportStatement:
		return i.evaluateImportStatement(n, env)
	case *ast.DynImportStatement:
		return i.evaluateDynImportStatement(n, env)
	case *ast.PreludeStatement:
		return runtime.NilValue{}, nil
	case *ast.ExternFunctionBody:
		return runtime.NilValue{}, nil
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
	for {
		cond, err := i.evaluateExpression(loop.Condition, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(cond) {
			return runtime.NilValue{}, nil
		}
		if _, err := i.evaluateBlock(loop.Body, env); err != nil {
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
	}
}

func (i *Interpreter) evaluateLoopExpression(loop *ast.LoopExpression, env *runtime.Environment) (runtime.Value, error) {
	if loop == nil || loop.Body == nil {
		return runtime.NilValue{}, nil
	}
	for {
		_, err := i.evaluateBlock(loop.Body, env)
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
	baseEnv := runtime.NewEnvironment(env)
	switch it := iterable.(type) {
	case *runtime.ArrayValue:
		state, err := i.ensureArrayState(it, 0)
		if err != nil {
			return nil, err
		}
		return i.iterateStaticValues(loop, baseEnv, state.values)
	case *runtime.IteratorValue:
		return i.iterateDynamicIterator(loop, baseEnv, it)
	default:
		iterator, err := i.resolveIteratorValue(iterable, env)
		if err != nil {
			return nil, err
		}
		return i.iterateDynamicIterator(loop, baseEnv, iterator)
	}
}

func (i *Interpreter) iterateStaticValues(loop *ast.ForLoop, baseEnv *runtime.Environment, values []runtime.Value) (runtime.Value, error) {
	for _, el := range values {
		val, continueLoop, err := i.runForLoopBody(loop, baseEnv, el)
		if err != nil {
			return nil, err
		}
		if !continueLoop {
			return val, nil
		}
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) iterateDynamicIterator(loop *ast.ForLoop, baseEnv *runtime.Environment, iterator *runtime.IteratorValue) (runtime.Value, error) {
	if iterator == nil {
		return nil, fmt.Errorf("iterator is nil")
	}
	defer iterator.Close()
	for {
		value, done, err := iterator.Next()
		if err != nil {
			return nil, err
		}
		if done {
			return runtime.NilValue{}, nil
		}
		val, continueLoop, err := i.runForLoopBody(loop, baseEnv, value)
		if err != nil {
			return nil, err
		}
		if !continueLoop {
			return val, nil
		}
	}
}

func (i *Interpreter) runForLoopBody(loop *ast.ForLoop, baseEnv *runtime.Environment, element runtime.Value) (runtime.Value, bool, error) {
	iterEnv := runtime.NewEnvironment(baseEnv)
	if err := i.assignPattern(loop.Pattern, element, iterEnv, true, nil); err != nil {
		return nil, false, err
	}
	val, err := i.evaluateBlock(loop.Body, iterEnv)
	if err != nil {
		switch sig := err.(type) {
		case breakSignal:
			if sig.label != "" {
				return nil, false, sig
			}
			return sig.value, false, nil
		case continueSignal:
			if sig.label != "" {
				return nil, false, fmt.Errorf("Labeled continue not supported")
			}
			return runtime.NilValue{}, true, nil
		case raiseSignal:
			return nil, false, sig
		case returnSignal:
			return nil, false, sig
		default:
			return nil, false, err
		}
	}
	return val, true, nil
}

func (i *Interpreter) resolveIteratorValue(iterable runtime.Value, env *runtime.Environment) (*runtime.IteratorValue, error) {
	ident := ast.NewIdentifier("iterator")
	if adapted, err := i.adaptIteratorValue(iterable, env); err != nil {
		return nil, err
	} else if adapted != nil {
		return adapted, nil
	}
	switch it := iterable.(type) {
	case *runtime.StructInstanceValue:
		member, err := i.structInstanceMember(it, ident, env, true)
		if err != nil {
			return nil, err
		}
		value, err := i.CallFunction(member, nil)
		if err != nil {
			return nil, err
		}
		if adapted, err := i.adaptIteratorValue(value, env); err != nil {
			return nil, err
		} else if adapted != nil {
			return adapted, nil
		}
		iterator, ok := value.(*runtime.IteratorValue)
		if !ok {
			return nil, fmt.Errorf("iterator() on %s did not return Iterator", iterable.Kind())
		}
		return iterator, nil
	case *runtime.InterfaceValue:
		member, err := i.interfaceMember(it, ident)
		if err != nil {
			return nil, err
		}
		value, err := i.CallFunction(member, nil)
		if err != nil {
			return nil, err
		}
		if adapted, err := i.adaptIteratorValue(value, env); err != nil {
			return nil, err
		} else if adapted != nil {
			return adapted, nil
		}
		iterator, ok := value.(*runtime.IteratorValue)
		if !ok {
			return nil, fmt.Errorf("iterator() on %s did not return Iterator", iterable.Kind())
		}
		return iterator, nil
	default:
		member, err := i.memberAccessOnValueWithOptions(iterable, ident, env, true)
		if err == nil && member != nil {
			value, callErr := i.CallFunction(member, nil)
			if callErr != nil {
				return nil, callErr
			}
			if adapted, err := i.adaptIteratorValue(value, env); err != nil {
				return nil, err
			} else if adapted != nil {
				return adapted, nil
			}
			if iterator, ok := value.(*runtime.IteratorValue); ok {
				return iterator, nil
			}
			return nil, fmt.Errorf("iterator() on %s did not return Iterator", iterable.Kind())
		}
		return nil, fmt.Errorf("for-loop iterable of kind %s is not Iterable", iterable.Kind())
	}
}

func (i *Interpreter) adaptIteratorValue(candidate runtime.Value, env *runtime.Environment) (*runtime.IteratorValue, error) {
	if iface, ok := candidate.(*runtime.InterfaceValue); ok && iface != nil {
		return i.adaptIteratorValue(iface.Underlying, env)
	}
	inst, ok := candidate.(*runtime.StructInstanceValue)
	if !ok || inst == nil {
		return nil, nil
	}
	nextVal, err := i.memberAccessOnValueWithOptions(inst, ast.NewIdentifier("next"), env, true)
	if err != nil || nextVal == nil {
		return nil, nil
	}
	var closeVal runtime.Value
	if closeCandidate, err := i.memberAccessOnValueWithOptions(inst, ast.NewIdentifier("close"), env, true); err == nil {
		closeVal = closeCandidate
	}
	step := func() (runtime.Value, bool, error) {
		res, err := i.CallFunction(nextVal, nil)
		if err != nil {
			return nil, true, err
		}
		if i.isIteratorEnd(res) {
			return runtime.IteratorEnd, true, nil
		}
		if res == nil {
			return runtime.NilValue{}, false, nil
		}
		return res, false, nil
	}
	var closer func()
	if closeVal != nil {
		closer = func() {
			_, _ = i.CallFunction(closeVal, nil)
		}
	}
	return runtime.NewIteratorValue(step, closer), nil
}

func (i *Interpreter) isIteratorEnd(val runtime.Value) bool {
	if val == nil {
		return false
	}
	switch v := val.(type) {
	case runtime.IteratorEndValue:
		return true
	case *runtime.InterfaceValue:
		return i.isIteratorEnd(v.Underlying)
	case *runtime.StructInstanceValue:
		return v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil && v.Definition.Node.ID.Name == "IteratorEnd"
	default:
		return val.Kind() == runtime.KindIteratorEnd
	}
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
