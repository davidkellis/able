package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func isVoidTypeExpr(expr ast.TypeExpression) bool {
	if expr == nil {
		return false
	}
	if simple, ok := expr.(*ast.SimpleTypeExpression); ok && simple.Name != nil {
		return simple.Name.Name == "void"
	}
	return false
}

func isVoidValue(value runtime.Value) bool {
	switch value.(type) {
	case runtime.VoidValue, *runtime.VoidValue:
		return true
	default:
		return false
	}
}

func isResultVoidType(expr ast.TypeExpression) bool {
	if expr == nil {
		return false
	}
	if res, ok := expr.(*ast.ResultTypeExpression); ok {
		return isVoidTypeExpr(res.InnerType)
	}
	return false
}

func (i *Interpreter) coerceReturnValue(returnType ast.TypeExpression, value runtime.Value, genericNames map[string]struct{}, env *runtime.Environment) (runtime.Value, error) {
	if returnType == nil {
		return value, nil
	}
	canonical := canonicalizeTypeExpression(returnType, env, i.typeAliases)
	if isVoidTypeExpr(canonical) {
		return runtime.VoidValue{}, nil
	}
	if isVoidValue(value) {
		if i.matchesType(canonical, value) || isResultVoidType(canonical) {
			return runtime.VoidValue{}, nil
		}
		expected := typeExpressionToString(canonical)
		return nil, fmt.Errorf("Return type mismatch: expected %s, got void", expected)
	}
	if typeExpressionUsesGenerics(returnType, genericNames) {
		return value, nil
	}
	if !i.matchesType(canonical, value) {
		expected := typeExpressionToString(canonical)
		actual := value.Kind().String()
		if actualExpr := i.typeExpressionFromValue(value); actualExpr != nil {
			actual = typeExpressionToString(actualExpr)
		}
		return nil, fmt.Errorf("Return type mismatch: expected %s, got %s", expected, actual)
	}
	coerced, err := i.coerceValueToType(canonical, value)
	if err != nil {
		return nil, err
	}
	return coerced, nil
}

func (i *Interpreter) evaluateFunctionCall(call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, error) {
	if member, ok := call.Callee.(*ast.MemberAccessExpression); ok {
		target, err := i.evaluateExpression(member.Object, env)
		if err != nil {
			return nil, err
		}
		if member.Safe && isNilRuntimeValue(target) {
			return runtime.NilValue{}, nil
		}
		// When a member access appears in callee position, prefer methods over fields so
		// method names that overlap with struct fields still bind to the callable.
		calleeVal, err := i.memberAccessOnValueWithOptions(target, member.Member, env, true)
		if err != nil {
			return nil, err
		}
		argValues := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argValues = append(argValues, val)
		}
		return i.callCallableValue(calleeVal, argValues, env, call)
	}
	if ident, ok := call.Callee.(*ast.Identifier); ok && ident != nil {
		calleeVal, lookupErr := env.Get(ident.Name)
		argValues := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argValues = append(argValues, val)
		}
		if lookupErr != nil {
			if dotIdx := strings.Index(ident.Name, "."); dotIdx > 0 && dotIdx < len(ident.Name)-1 {
				head := ident.Name[:dotIdx]
				tail := ident.Name[dotIdx+1:]
				receiver, err := env.Get(head)
				if err != nil {
					if def, ok := env.StructDefinition(head); ok {
						receiver = def
					} else {
						receiver = runtime.TypeRefValue{TypeName: head}
					}
				}
				member := ast.ID(tail)
				candidate, err := i.memberAccessOnValueWithOptions(receiver, member, env, true)
				if err != nil {
					return nil, err
				}
				return i.callCallableValue(candidate, argValues, env, call)
			}
			return nil, lookupErr
		}
		return i.callCallableValue(calleeVal, argValues, env, call)
	}
	calleeVal, err := i.evaluateExpression(call.Callee, env)
	if err != nil {
		return nil, err
	}
	argValues := make([]runtime.Value, 0, len(call.Arguments))
	for _, argExpr := range call.Arguments {
		val, err := i.evaluateExpression(argExpr, env)
		if err != nil {
			return nil, err
		}
		argValues = append(argValues, val)
	}
	return i.callCallableValue(calleeVal, argValues, env, call)
}

func (i *Interpreter) invokeFunction(fn *runtime.FunctionValue, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.Body == nil {
			return runtime.NilValue{}, nil
		}
		if call != nil {
			if err := i.populateCallTypeArguments(decl, call, args); err != nil {
				return nil, err
			}
		}
		paramCount := len(decl.Params)
		optionalLast := paramCount > 0 && isNullableParam(decl.Params[paramCount-1])
		expectedArgs := paramCount
		if decl.IsMethodShorthand {
			expectedArgs++
		}
		if !arityMatchesRuntime(expectedArgs, len(args), optionalLast) {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, expectedArgs, len(args))
		}
		missingOptional := optionalLast && len(args) == expectedArgs-1
		if receiver, ok := resolveMethodSetReceiver(decl, args); ok {
			if err := i.enforceMethodSetConstraints(fn, receiver); err != nil {
				return nil, err
			}
		}
		if call != nil {
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		localEnv := runtime.NewEnvironment(fn.Closure)
		if call != nil {
			i.bindTypeArgumentsIfAny(decl, call, localEnv)
		}
		bindArgs := args
		var implicitReceiver runtime.Value
		hasImplicit := false
		if decl.IsMethodShorthand {
			implicitReceiver = args[0]
			hasImplicit = true
			if len(args) > 1 {
				bindArgs = args[1:]
			} else {
				bindArgs = nil
			}
		} else {
			if paramCount > 0 && len(args) > 0 {
				implicitReceiver = args[0]
				hasImplicit = true
			}
		}
		if missingOptional {
			bindArgs = append(bindArgs, runtime.NilValue{})
		}
		if len(bindArgs) != paramCount {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, paramCount, len(bindArgs))
		}
		generics := functionGenericNameSet(fn, decl)
		// Check for slot-enabled bytecode program.
		var slotProgram *bytecodeProgram
		if i.execMode == execModeBytecode {
			if p, ok := fn.Bytecode.(*bytecodeProgram); ok && p != nil && p.frameLayout != nil {
				slotProgram = p
			}
		}
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("function parameter %d is nil", idx)
			}
			arg := bindArgs[idx]
			if param.ParamType != nil && !paramUsesGeneric(param.ParamType, generics) {
				coerced, err := i.coerceValueToType(param.ParamType, arg)
				if err != nil {
					return nil, err
				}
				arg = coerced
				bindArgs[idx] = coerced
			}
			if slotProgram == nil {
				if err := i.assignPattern(param.Name, arg, localEnv, true, nil); err != nil {
					return nil, err
				}
			}
		}
		state := i.stateFromEnv(localEnv)
		if hasImplicit {
			state.pushImplicitReceiver(implicitReceiver)
			defer state.popImplicitReceiver()
		}
		if thunk, ok := fn.Bytecode.(CompiledThunk); ok && thunk != nil {
			var serialSync *SerialExecutor
			if serial, ok := i.executor.(*SerialExecutor); ok {
				var payload *asyncContextPayload
				if env != nil {
					payload = payloadFromState(env.RuntimeData())
				}
				if payload == nil {
					payload = payloadFromState(localEnv.RuntimeData())
				}
				if payload == nil {
					serialSync = serial
					serialSync.beginSynchronousSection()
				}
			}
			if serialSync != nil {
				defer serialSync.endSynchronousSection()
			}
			result, err := thunk(localEnv, args)
			if err != nil {
				return nil, err
			}
			if result == nil {
				result = runtime.NilValue{}
			}
			return i.coerceReturnValue(decl.ReturnType, result, generics, localEnv)
		}
		if i.execMode == execModeBytecode {
			if program, ok := fn.Bytecode.(*bytecodeProgram); ok && program != nil {
				vm := newBytecodeVM(i, localEnv)
				if slotProgram != nil {
					layout := slotProgram.frameLayout
					slots := make([]runtime.Value, layout.slotCount)
					for idx := 0; idx < len(bindArgs) && idx < layout.paramSlots; idx++ {
						slots[idx] = bindArgs[idx]
					}
					vm.slots = slots
				}
				result, err := vm.run(program)
				if err != nil {
					if ret, ok := err.(returnSignal); ok {
						retVal := ret.value
						if retVal == nil {
							retVal = runtime.NilValue{}
						}
						coerced, err := i.coerceReturnValue(decl.ReturnType, retVal, generics, localEnv)
						if err != nil {
							if ret.context != nil {
								return nil, runtimeDiagnosticError{err: err, context: ret.context}
							}
							return nil, err
						}
						return coerced, nil
					}
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				return i.coerceReturnValue(decl.ReturnType, result, generics, localEnv)
			}
			name := "<anonymous>"
			if decl.ID != nil && decl.ID.Name != "" {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("bytecode missing for function %s", name)
		}
		result, err := i.evaluateBlock(decl.Body, localEnv)
		if err != nil {
			if ret, ok := err.(returnSignal); ok {
				retVal := ret.value
				if retVal == nil {
					retVal = runtime.NilValue{}
				}
				coerced, err := i.coerceReturnValue(decl.ReturnType, retVal, generics, localEnv)
				if err != nil {
					if ret.context != nil {
						return nil, runtimeDiagnosticError{err: err, context: ret.context}
					}
					return nil, err
				}
				return coerced, nil
			}
			return nil, err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		return i.coerceReturnValue(decl.ReturnType, result, generics, localEnv)
	case *ast.LambdaExpression:
		if call != nil {
			if err := i.populateCallTypeArguments(decl, call, args); err != nil {
				return nil, err
			}
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		if len(args) != len(decl.Params) {
			return nil, fmt.Errorf("Lambda expects %d arguments, got %d", len(decl.Params), len(args))
		}
		localEnv := runtime.NewEnvironment(fn.Closure)
		if call != nil {
			i.bindTypeArgumentsIfAny(decl, call, localEnv)
		}
		var implicitReceiver runtime.Value
		hasImplicit := false
		if len(decl.Params) > 0 && len(args) > 0 {
			implicitReceiver = args[0]
			hasImplicit = true
		}
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("lambda parameter %d is nil", idx)
			}
			if err := i.assignPattern(param.Name, args[idx], localEnv, true, nil); err != nil {
				return nil, err
			}
		}
		state := i.stateFromEnv(localEnv)
		if hasImplicit {
			state.pushImplicitReceiver(implicitReceiver)
			defer state.popImplicitReceiver()
		}
		lambdaGenerics := genericNameSet(decl.GenericParams)
		if i.execMode == execModeBytecode {
			if program, ok := fn.Bytecode.(*bytecodeProgram); ok && program != nil {
				vm := newBytecodeVM(i, localEnv)
				result, err := vm.run(program)
				if err != nil {
					if ret, ok := err.(returnSignal); ok {
						retVal := ret.value
						if retVal == nil {
							retVal = runtime.NilValue{}
						}
						coerced, err := i.coerceReturnValue(decl.ReturnType, retVal, lambdaGenerics, localEnv)
						if err != nil {
							if ret.context != nil {
								return nil, runtimeDiagnosticError{err: err, context: ret.context}
							}
							return nil, err
						}
						return coerced, nil
					}
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				return i.coerceReturnValue(decl.ReturnType, result, lambdaGenerics, localEnv)
			}
			return nil, fmt.Errorf("bytecode missing for lambda")
		}
		result, err := i.evaluateExpression(decl.Body, localEnv)
		if err != nil {
			if ret, ok := err.(returnSignal); ok {
				retVal := ret.value
				if retVal == nil {
					retVal = runtime.NilValue{}
				}
				coerced, err := i.coerceReturnValue(decl.ReturnType, retVal, lambdaGenerics, localEnv)
				if err != nil {
					if ret.context != nil {
						return nil, runtimeDiagnosticError{err: err, context: ret.context}
					}
					return nil, err
				}
				return coerced, nil
			}
			return nil, err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		return i.coerceReturnValue(decl.ReturnType, result, lambdaGenerics, localEnv)
	default:
		return nil, fmt.Errorf("calling unsupported function declaration %T", fn.Declaration)
	}
}

func makePartialFunctionValue(target runtime.Value, bound []runtime.Value, call *ast.FunctionCall) runtime.Value {
	argsCopy := make([]runtime.Value, len(bound))
	copy(argsCopy, bound)
	return &runtime.PartialFunctionValue{
		Target:    target,
		BoundArgs: argsCopy,
		Call:      call,
	}
}

func (i *Interpreter) callCallableValue(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	if callee == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	switch fn := callee.(type) {
	case runtime.PartialFunctionValue:
		merged := append([]runtime.Value{}, fn.BoundArgs...)
		merged = append(merged, args...)
		return i.callCallableValue(fn.Target, merged, env, call)
	case *runtime.PartialFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("partial function is nil")
		}
		merged := append([]runtime.Value{}, fn.BoundArgs...)
		merged = append(merged, args...)
		return i.callCallableValue(fn.Target, merged, env, call)
	}
	if state := i.stateFromEnv(env); call != nil {
		state.pushCallFrame(call)
		defer state.popCallFrame()
	}
	var callState any
	if env != nil {
		callState = env.RuntimeData()
	}
	var native *runtime.NativeFunctionValue
	var injected []runtime.Value
	var overloads []*runtime.FunctionValue
	partialTarget := callee

	switch fn := callee.(type) {
	case runtime.NativeFunctionValue:
		native = &fn
		partialTarget = native
	case *runtime.NativeFunctionValue:
		native = fn
		partialTarget = fn
	case runtime.NativeBoundMethodValue:
		native = &fn.Method
		injected = append(injected, fn.Receiver)
		partialTarget = native
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native bound method is nil")
		}
		native = &fn.Method
		injected = append(injected, fn.Receiver)
		partialTarget = native
	case runtime.DynRefValue:
		resolved, err := i.resolveDynRef(fn)
		if err != nil {
			return nil, err
		}
		return i.callCallableValue(resolved, args, env, call)
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, err := i.resolveDynRef(*fn)
		if err != nil {
			return nil, err
		}
		return i.callCallableValue(resolved, args, env, call)
	case runtime.BoundMethodValue:
		injected = append(injected, fn.Receiver)
		if nfv, ok := fn.Method.(*runtime.NativeFunctionValue); ok && nfv != nil {
			native = nfv
			partialTarget = nfv
		} else if nfv, ok := fn.Method.(runtime.NativeFunctionValue); ok {
			native = &nfv
			partialTarget = native
		} else {
			overloads = functionOverloads(fn.Method)
			partialTarget = fn.Method
		}
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("bound method is nil")
		}
		injected = append(injected, fn.Receiver)
		if nfv, ok := fn.Method.(*runtime.NativeFunctionValue); ok && nfv != nil {
			native = nfv
			partialTarget = nfv
		} else if nfv, ok := fn.Method.(runtime.NativeFunctionValue); ok {
			native = &nfv
			partialTarget = native
		} else {
			overloads = functionOverloads(fn.Method)
			partialTarget = fn.Method
		}
	default:
		overloads = functionOverloads(callee)
	}

	evalArgs := append(injected, args...)

	if native != nil {
		if native.Arity >= 0 {
			provided := len(evalArgs) - len(injected)
			if provided < 0 {
				provided = 0
			}
			if provided > native.Arity {
				name := native.Name
				if name == "" {
					name = "(native)"
				}
				return nil, fmt.Errorf("Arity mismatch calling %s: expected %d, got %d", name, native.Arity, provided)
			}
			if provided < native.Arity {
				return makePartialFunctionValue(partialTarget, evalArgs, call), nil
			}
		}
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return native.Impl(ctx, evalArgs)
	}

	if len(overloads) == 0 {
		if applyMethod, err := i.findApplyMethod(callee); err == nil && applyMethod != nil {
			bound := runtime.BoundMethodValue{Receiver: callee, Method: applyMethod}
			return i.callCallableValue(bound, args, env, call)
		} else if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("calling non-function value of kind %s (%T)", callee.Kind(), callee)
	}

	if len(overloads) > 0 {
		minRequired := minArgsForOverloads(overloads)
		if len(evalArgs) < minRequired {
			return makePartialFunctionValue(partialTarget, evalArgs, call), nil
		}
	}

	selected, err := i.selectRuntimeOverload(overloads, evalArgs, call)
	if err != nil {
		return nil, err
	}
	if selected == nil {
		if len(overloads) == 1 {
			if mismatchErr := i.reportOverloadMismatch(overloads[0], evalArgs, call); mismatchErr != nil {
				return nil, mismatchErr
			}
		}
		return nil, fmt.Errorf("No overloads of %s match provided arguments", overloadName(call))
	}
	return i.invokeFunction(selected, evalArgs, env, call)
}

func (i *Interpreter) evaluatePipeExpression(subject runtime.Value, rhs ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	state.pushImplicitReceiver(subject)
	defer state.popImplicitReceiver()

	if placeholderCallable, ok, err := i.tryBuildPlaceholderFunction(rhs, env); err != nil {
		return nil, err
	} else if ok {
		result, callErr := i.callCallableValue(placeholderCallable, []runtime.Value{subject}, env, nil)
		if callErr != nil {
			return nil, fmt.Errorf("pipe RHS must be callable: %w", callErr)
		}
		if result == nil {
			return runtime.NilValue{}, nil
		}
		return result, nil
	}

	if call, isCall := rhs.(*ast.FunctionCall); isCall {
		calleeVal, err := i.evaluateExpression(call.Callee, env)
		if err != nil {
			return nil, err
		}
		argValues := make([]runtime.Value, 0, len(call.Arguments))
		for _, arg := range call.Arguments {
			val, evalErr := i.evaluateExpression(arg, env)
			if evalErr != nil {
				return nil, evalErr
			}
			argValues = append(argValues, val)
		}
		callArgs := argValues
		switch calleeVal.(type) {
		case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:
			// Bound methods already capture the receiver.
		default:
			callArgs = append([]runtime.Value{subject}, argValues...)
		}
		result, callErr := i.callCallableValue(calleeVal, callArgs, env, call)
		if callErr != nil {
			return nil, fmt.Errorf("pipe RHS must be callable: %w", callErr)
		}
		if result == nil {
			return runtime.NilValue{}, nil
		}
		return result, nil
	}

	rhsVal, err := i.evaluateExpression(rhs, env)
	if err != nil {
		return nil, err
	}
	callArgs := []runtime.Value{subject}
	switch rhsVal.(type) {
	case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:
		callArgs = nil
	}
	result, err := i.callCallableValue(rhsVal, callArgs, env, nil)
	if err != nil {
		return nil, fmt.Errorf("pipe RHS must be callable: %w", err)
	}
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func (i *Interpreter) populateCallTypeArguments(funcNode ast.Node, call *ast.FunctionCall, args []runtime.Value) error {
	if funcNode == nil || call == nil {
		return nil
	}
	generics, _ := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return nil
	}
	if len(call.TypeArguments) > 0 {
		if len(call.TypeArguments) != len(generics) {
			name := functionNameForErrors(funcNode)
			return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", name, len(generics), len(call.TypeArguments))
		}
		return nil
	}
	bindings := make(map[string]ast.TypeExpression)
	genericNames := genericNameSet(generics)
	params := extractFunctionParams(funcNode)
	bindArgs := args
	if def, ok := funcNode.(*ast.FunctionDefinition); ok && def.IsMethodShorthand && len(bindArgs) > 0 {
		bindArgs = bindArgs[1:]
	}
	max := len(params)
	if len(bindArgs) < max {
		max = len(bindArgs)
	}
	for idx := 0; idx < max; idx++ {
		param := params[idx]
		if param == nil || param.ParamType == nil {
			continue
		}
		actual := i.typeExpressionFromValue(bindArgs[idx])
		if actual == nil {
			continue
		}
		matchTypeExpressionTemplate(param.ParamType, actual, genericNames, bindings)
	}
	typeArgs := make([]ast.TypeExpression, len(generics))
	for idx, gp := range generics {
		if gp != nil && gp.Name != nil {
			if bound, ok := bindings[gp.Name.Name]; ok {
				typeArgs[idx] = bound
				continue
			}
		}
		typeArgs[idx] = ast.NewWildcardTypeExpression()
	}
	call.TypeArguments = typeArgs
	return nil
}

func (i *Interpreter) enforceGenericConstraintsIfAny(funcNode ast.Node, call *ast.FunctionCall) error {
	if funcNode == nil || call == nil {
		return nil
	}
	generics, whereClause := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return nil
	}
	name := functionNameForErrors(funcNode)
	if len(call.TypeArguments) != len(generics) {
		return fmt.Errorf("Type arguments count mismatch calling %s: expected %d, got %d", name, len(generics), len(call.TypeArguments))
	}
	constraints := collectConstraintSpecs(generics, whereClause)
	if len(constraints) == 0 {
		return nil
	}
	typeArgMap, err := mapTypeArguments(generics, call.TypeArguments, fmt.Sprintf("calling %s", name))
	if err != nil {
		return err
	}
	return i.enforceConstraintSpecs(constraints, typeArgMap)
}

func (i *Interpreter) bindTypeArgumentsIfAny(funcNode ast.Node, call *ast.FunctionCall, env *runtime.Environment) {
	if funcNode == nil || call == nil {
		return
	}
	generics, _ := extractFunctionGenerics(funcNode)
	if len(generics) == 0 {
		return
	}
	count := len(generics)
	if len(call.TypeArguments) < count {
		count = len(call.TypeArguments)
	}
	for idx := 0; idx < count; idx++ {
		gp := generics[idx]
		if gp == nil || gp.Name == nil {
			continue
		}
		ta := call.TypeArguments[idx]
		if ta == nil {
			continue
		}
		name := gp.Name.Name + "_type"
		value := runtime.StringValue{Val: typeExpressionToString(ta)}
		env.Define(name, value)
		if info, ok := parseTypeExpression(ta); ok {
			env.Define(gp.Name.Name, runtime.TypeRefValue{TypeName: info.name, TypeArgs: info.typeArgs})
		}
	}
}

func (i *Interpreter) evaluateIteratorLiteral(expr *ast.IteratorLiteral, env *runtime.Environment) (runtime.Value, error) {
	iterEnv := runtime.NewEnvironment(env)
	instance := newGeneratorInstance(i, iterEnv, expr.Body)
	controller := instance.controllerValue()
	bindingName := "gen"
	if expr.Binding != nil && expr.Binding.Name != "" {
		bindingName = expr.Binding.Name
	}
	iterEnv.Define(bindingName, controller)
	if bindingName != "gen" {
		iterEnv.Define("gen", controller)
	}
	return runtime.NewIteratorValue(func() (runtime.Value, bool, error) {
		return instance.next()
	}, instance.close), nil
}

func (i *Interpreter) evaluateLambdaExpression(expr *ast.LambdaExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr == nil {
		return nil, fmt.Errorf("lambda expression is nil")
	}
	fnVal := &runtime.FunctionValue{Declaration: expr, Closure: env}
	if expr.Body != nil {
		program, err := i.lowerExpressionToBytecodeWithOptions(expr.Body, true)
		if err != nil {
			if i.execMode == execModeBytecode {
				return nil, err
			}
		} else {
			fnVal.Bytecode = program
		}
	}
	return fnVal, nil
}
