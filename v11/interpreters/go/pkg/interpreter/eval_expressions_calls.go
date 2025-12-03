package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateFunctionCall(call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, error) {
	if member, ok := call.Callee.(*ast.MemberAccessExpression); ok {
		target, err := i.evaluateExpression(member.Object, env)
		if err != nil {
			return nil, err
		}
		if member.Safe && isNilRuntimeValue(target) {
			return runtime.NilValue{}, nil
		}
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
			if len(argValues) > 0 {
				if bound, ok := i.tryUfcs(env, ident.Name, argValues[0]); ok {
					return i.callCallableValue(bound, argValues[1:], env, call)
				}
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

func (i *Interpreter) invokeFunction(fn *runtime.FunctionValue, args []runtime.Value, call *ast.FunctionCall) (runtime.Value, error) {
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.Body == nil {
			return runtime.NilValue{}, nil
		}
		if call != nil {
			if err := i.populateCallTypeArguments(decl, call, args); err != nil {
				return nil, err
			}
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
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
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("function parameter %d is nil", idx)
			}
			if err := i.assignPattern(param.Name, bindArgs[idx], localEnv, true, nil); err != nil {
				return nil, err
			}
		}
		state := i.stateFromEnv(localEnv)
		if hasImplicit {
			state.pushImplicitReceiver(implicitReceiver)
			defer state.popImplicitReceiver()
		}
		result, err := i.evaluateBlock(decl.Body, localEnv)
		if err != nil {
			if ret, ok := err.(returnSignal); ok {
				if ret.value == nil {
					return runtime.NilValue{}, nil
				}
				return ret.value, nil
			}
			return nil, err
		}
		if result == nil {
			return runtime.NilValue{}, nil
		}
		return result, nil
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
		result, err := i.evaluateExpression(decl.Body, localEnv)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return runtime.NilValue{}, nil
		}
		return result, nil
	default:
		return nil, fmt.Errorf("calling unsupported function declaration %T", fn.Declaration)
	}
}

func (i *Interpreter) callCallableValue(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	if callee == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	var callState any
	if env != nil {
		callState = env.RuntimeData()
	}
	var native *runtime.NativeFunctionValue
	var injected []runtime.Value
	var overloads []*runtime.FunctionValue

	switch fn := callee.(type) {
	case runtime.NativeFunctionValue:
		native = &fn
	case *runtime.NativeFunctionValue:
		native = fn
	case runtime.NativeBoundMethodValue:
		native = &fn.Method
		injected = append(injected, fn.Receiver)
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native bound method is nil")
		}
		native = &fn.Method
		injected = append(injected, fn.Receiver)
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
		overloads = functionOverloads(fn.Method)
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("bound method is nil")
		}
		injected = append(injected, fn.Receiver)
		overloads = functionOverloads(fn.Method)
	default:
		overloads = functionOverloads(callee)
	}

	evalArgs := append(injected, args...)

	if native != nil {
		if native.Arity >= 0 && len(injected) == 0 && len(evalArgs) != native.Arity {
			name := native.Name
			if name == "" {
				name = "(native)"
			}
			return nil, fmt.Errorf("Arity mismatch calling %s: expected %d, got %d", name, native.Arity, len(evalArgs))
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

	selected, err := i.selectRuntimeOverload(overloads, evalArgs, call)
	if err != nil {
		return nil, err
	}
	if selected == nil && len(overloads) == 1 {
		selected = overloads[0]
	}
	if selected == nil {
		return nil, fmt.Errorf("No overloads of %s match provided arguments", overloadName(call))
	}
	return i.invokeFunction(selected, evalArgs, call)
}

func (i *Interpreter) selectRuntimeOverload(overloads []*runtime.FunctionValue, evalArgs []runtime.Value, call *ast.FunctionCall) (*runtime.FunctionValue, error) {
	type candidate struct {
		fn    *runtime.FunctionValue
		score float64
	}
	candidates := make([]candidate, 0, len(overloads))
	for _, fn := range overloads {
		if fn == nil || fn.Declaration == nil {
			continue
		}
		switch decl := fn.Declaration.(type) {
		case *ast.FunctionDefinition:
			params := decl.Params
			paramCount := len(params)
			optionalLast := paramCount > 0 && isNullableParam(params[paramCount-1])
			expectedArgs := paramCount
			if decl.IsMethodShorthand {
				expectedArgs++
			}
			if !arityMatchesRuntime(expectedArgs, len(evalArgs), optionalLast) {
				continue
			}
			paramsForCheck := params
			argsForCheck := evalArgs
			if optionalLast && len(evalArgs) == expectedArgs-1 {
				paramsForCheck = params[:paramCount-1]
			}
			if decl.IsMethodShorthand && len(argsForCheck) > 0 {
				argsForCheck = argsForCheck[1:]
			}
			if len(argsForCheck) != len(paramsForCheck) {
				continue
			}
			generics := genericNameSet(decl.GenericParams)
			score := 0.0
			if optionalLast && len(evalArgs) == expectedArgs-1 {
				score -= 0.5
			}
			compatible := true
			for idx, param := range paramsForCheck {
				if param == nil {
					compatible = false
					break
				}
				if param.ParamType != nil {
					if !i.matchesType(param.ParamType, argsForCheck[idx]) {
						compatible = false
						break
					}
					score += float64(parameterSpecificity(param.ParamType, generics))
				}
			}
			if compatible {
				candidates = append(candidates, candidate{fn: fn, score: score})
			}
		case *ast.LambdaExpression:
			if len(decl.Params) != len(evalArgs) {
				continue
			}
			candidates = append(candidates, candidate{fn: fn, score: 0})
		default:
			candidates = append(candidates, candidate{fn: fn, score: 0})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	best := candidates[0]
	ties := []candidate{best}
	for _, cand := range candidates[1:] {
		if cand.score > best.score {
			best = cand
			ties = []candidate{cand}
		} else if cand.score == best.score {
			ties = append(ties, cand)
		}
	}
	if len(ties) > 1 {
		return nil, fmt.Errorf("Ambiguous overload for %s", overloadName(call))
	}
	return best.fn, nil
}

func overloadName(call *ast.FunctionCall) string {
	if call == nil {
		return "(function)"
	}
	switch cal := call.Callee.(type) {
	case *ast.Identifier:
		return cal.Name
	case *ast.MemberAccessExpression:
		if id, ok := cal.Member.(*ast.Identifier); ok {
			return id.Name
		}
	}
	return "(function)"
}

func parameterSpecificity(typeExpr ast.TypeExpression, generics map[string]struct{}) int {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return 0
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0
		}
		if _, ok := generics[t.Name.Name]; ok {
			return 1
		}
		return 3
	case *ast.NullableTypeExpression:
		return 1 + parameterSpecificity(t.InnerType, generics)
	case *ast.GenericTypeExpression:
		score := 2 + parameterSpecificity(t.Base, generics)
		for _, arg := range t.Arguments {
			score += parameterSpecificity(arg, generics)
		}
		return score
	case *ast.FunctionTypeExpression, *ast.UnionTypeExpression:
		return 2
	default:
		return 1
	}
}

func arityMatchesRuntime(expected, actual int, optionalLast bool) bool {
	return actual == expected || (optionalLast && actual == expected-1)
}

func isNullableParam(param *ast.FunctionParameter) bool {
	if param == nil {
		return false
	}
	if param.ParamType == nil {
		return false
	}
	_, ok := param.ParamType.(*ast.NullableTypeExpression)
	return ok
}

func (i *Interpreter) evaluatePipeExpression(subject runtime.Value, rhs ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	state.pushTopic(subject)
	state.pushImplicitReceiver(subject)
	rhsVal, err := i.evaluateExpression(rhs, env)
	state.popImplicitReceiver()
	used := state.topicWasUsed()
	state.popTopic()
	if err != nil {
		return nil, err
	}
	if used {
		if rhsVal == nil {
			return runtime.NilValue{}, nil
		}
		return rhsVal, nil
	}
	callArgs := []runtime.Value{subject}
	switch rhsVal.(type) {
	case runtime.BoundMethodValue, *runtime.BoundMethodValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue:
		callArgs = nil
	}
	result, err := i.callCallableValue(rhsVal, callArgs, env, nil)
	if err != nil {
		return nil, fmt.Errorf("pipe RHS must be callable when '%%' is not used: %w", err)
	}
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}

func (i *Interpreter) evaluateTopicReferenceExpression(_ *ast.TopicReferenceExpression, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	val, ok := state.currentTopic()
	if !ok {
		return nil, fmt.Errorf("Topic reference '%%' used outside of pipe expression")
	}
	state.markTopicUsed()
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
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
	}
}

func (i *Interpreter) evaluateAssignment(assign *ast.AssignmentExpression, env *runtime.Environment) (runtime.Value, error) {
	value, err := i.evaluateExpression(assign.Right, env)
	if err != nil {
		return nil, err
	}
	binaryOp, isCompound := binaryOpForAssignment(assign.Operator)

	switch lhs := assign.Left.(type) {
	case *ast.Identifier:
		switch assign.Operator {
		case ast.AssignmentDeclare:
			if env.HasInCurrentScope(lhs.Name) {
				return nil, fmt.Errorf(":= requires at least one new binding")
			}
			env.Define(lhs.Name, value)
		case ast.AssignmentAssign:
			if !env.AssignExisting(lhs.Name, value) {
				env.Define(lhs.Name, value)
			}
		default:
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			current, err := env.Get(lhs.Name)
			if err != nil {
				return nil, err
			}
			computed, err := applyBinaryOperator(binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			if err := env.Assign(lhs.Name, computed); err != nil {
				return nil, err
			}
			return computed, nil
		}
	case *ast.MemberAccessExpression:
		if lhs.Safe {
			return nil, fmt.Errorf("Cannot assign through safe navigation")
		}
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on member access")
		}
		target, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		switch inst := target.(type) {
		case *runtime.StructInstanceValue:
			return assignStructMember(inst, lhs.Member, value, assign.Operator, binaryOp, isCompound)
		case *runtime.ArrayValue:
			arrayVal := inst
			switch member := lhs.Member.(type) {
			case *ast.IntegerLiteral:
				if member.Value == nil {
					return nil, fmt.Errorf("Array index out of bounds")
				}
				idx := int(member.Value.Int64())
				state, err := i.ensureArrayState(arrayVal, 0)
				if err != nil {
					return nil, err
				}
				if idx < 0 || idx >= len(state.values) {
					return nil, fmt.Errorf("Array index out of bounds")
				}
				if assign.Operator == ast.AssignmentAssign {
					state.values[idx] = value
					i.syncArrayValues(arrayVal.Handle, state)
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				current := state.values[idx]
				computed, err := applyBinaryOperator(binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				state.values[idx] = computed
				i.syncArrayValues(arrayVal.Handle, state)
				return computed, nil
			case *ast.Identifier:
				if assign.Operator != ast.AssignmentAssign {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				state, err := i.ensureArrayState(arrayVal, 0)
				if err != nil {
					return nil, err
				}
				switch member.Name {
				case "storage_handle":
					intVal, ok := value.(runtime.IntegerValue)
					if !ok || intVal.Val == nil || !intVal.Val.IsInt64() {
						return nil, fmt.Errorf("array storage_handle must be an integer")
					}
					handle := intVal.Val.Int64()
					if handle <= 0 {
						return nil, fmt.Errorf("array storage_handle must be positive")
					}
					newState, ok := i.arrayStates[handle]
					if !ok {
						newState = &arrayState{values: make([]runtime.Value, 0), capacity: 0}
						i.arrayStates[handle] = newState
					}
					i.trackArrayValue(handle, arrayVal)
					arrayVal.Elements = newState.values
					i.syncArrayValues(handle, newState)
					return value, nil
				case "length":
					newLen, err := arrayIndexFromValue(value)
					if err != nil {
						return nil, fmt.Errorf("array length must be a non-negative integer")
					}
					setArrayLength(state, newLen)
					i.syncArrayValues(arrayVal.Handle, state)
					return value, nil
				case "capacity":
					newCap, err := arrayIndexFromValue(value)
					if err != nil {
						return nil, fmt.Errorf("array capacity must be a non-negative integer")
					}
					if newCap < len(state.values) {
						newCap = len(state.values)
					}
					if ensureArrayCapacity(state, newCap) {
						// ensureArrayCapacity already sets capacity and sync handles reallocations
					} else if newCap > state.capacity {
						state.capacity = newCap
					}
					i.syncArrayValues(arrayVal.Handle, state)
					return value, nil
				default:
					return nil, fmt.Errorf("Array has no member '%s'", member.Name)
				}
			default:
				return nil, fmt.Errorf("Array member assignment requires integer member")
			}
		default:
			return nil, fmt.Errorf("Member assignment requires struct or array")
		}
	case *ast.ImplicitMemberExpression:
		if assign.Operator == ast.AssignmentDeclare {
			if lhs.Member != nil {
				return nil, fmt.Errorf("Cannot use := on implicit member '#%s'", lhs.Member.Name)
			}
			return nil, fmt.Errorf("Cannot use := on implicit member")
		}
		state := i.stateFromEnv(env)
		receiver, ok := state.currentImplicitReceiver()
		if !ok || receiver == nil {
			if lhs.Member != nil {
				return nil, fmt.Errorf("Implicit member '#%s' used outside of function with implicit receiver", lhs.Member.Name)
			}
			return nil, fmt.Errorf("Implicit member used outside of function with implicit receiver")
		}
		switch inst := receiver.(type) {
		case *runtime.StructInstanceValue:
			return assignStructMember(inst, lhs.Member, value, assign.Operator, binaryOp, isCompound)
		default:
			return nil, fmt.Errorf("Implicit member assignments supported only on struct instances")
		}
	case *ast.IndexExpression:
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on index assignment")
		}
		arrObj, err := i.evaluateExpression(lhs.Object, env)
		if err != nil {
			return nil, err
		}
		idxVal, err := i.evaluateExpression(lhs.Index, env)
		if err != nil {
			return nil, err
		}
		if setMethod, err := i.findIndexMethod(arrObj, "set", "IndexMut"); err != nil {
			return nil, err
		} else if setMethod != nil {
			if assign.Operator == ast.AssignmentAssign {
				if _, err := i.CallFunction(setMethod, []runtime.Value{arrObj, idxVal, value}); err != nil {
					return nil, err
				}
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			getMethod, err := i.findIndexMethod(arrObj, "get", "Index")
			if err != nil {
				return nil, err
			}
			if getMethod == nil {
				return nil, fmt.Errorf("Compound index assignment requires readable Index implementation")
			}
			current, err := i.CallFunction(getMethod, []runtime.Value{arrObj, idxVal})
			if err != nil {
				return nil, err
			}
			computed, err := applyBinaryOperator(binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			if _, err := i.CallFunction(setMethod, []runtime.Value{arrObj, idxVal, computed}); err != nil {
				return nil, err
			}
			return computed, nil
		}
		arr, err := i.toArrayValue(arrObj)
		if err != nil {
			return nil, err
		}
		idx, err := indexFromValue(idxVal)
		if err != nil {
			return nil, err
		}
		state, err := i.ensureArrayState(arr, 0)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx >= len(state.values) {
			return nil, fmt.Errorf("Array index out of bounds")
		}
		if assign.Operator == ast.AssignmentAssign {
			state.values[idx] = value
			i.syncArrayValues(arr.Handle, state)
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		current := state.values[idx]
		computed, err := applyBinaryOperator(binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		state.values[idx] = computed
		i.syncArrayValues(arr.Handle, state)
		return computed, nil
	case ast.Pattern:
		if isCompound {
			return nil, fmt.Errorf("compound assignment not supported with patterns")
		}
		switch assign.Operator {
		case ast.AssignmentDeclare:
			newNames, hasAny := analyzePatternDeclarationNames(env, lhs)
			if !hasAny || len(newNames) == 0 {
				return nil, fmt.Errorf(":= requires at least one new binding")
			}
			intent := &bindingIntent{declarationNames: newNames}
			if err := i.assignPattern(lhs, value, env, true, intent); err != nil {
				return nil, err
			}
		case ast.AssignmentAssign:
			intent := &bindingIntent{allowFallback: true}
			if err := i.assignPattern(lhs, value, env, false, intent); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
	default:
		return nil, fmt.Errorf("unsupported assignment target %s", lhs.NodeType())
	}

	return value, nil
}

func analyzePatternDeclarationNames(env *runtime.Environment, pattern ast.Pattern) (map[string]struct{}, bool) {
	names := make(map[string]struct{})
	collectPatternIdentifiers(pattern, names)
	newNames := make(map[string]struct{})
	for name := range names {
		if !env.HasInCurrentScope(name) {
			newNames[name] = struct{}{}
		}
	}
	return newNames, len(names) > 0
}

func collectPatternIdentifiers(pattern ast.Pattern, into map[string]struct{}) {
	switch p := pattern.(type) {
	case *ast.Identifier:
		if p.Name != "" {
			into[p.Name] = struct{}{}
		}
	case *ast.StructPattern:
		for _, field := range p.Fields {
			if field == nil {
				continue
			}
			if field.Binding != nil && field.Binding.Name != "" {
				into[field.Binding.Name] = struct{}{}
			}
			if inner, ok := field.Pattern.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			}
		}
	case *ast.ArrayPattern:
		for _, elem := range p.Elements {
			if elem == nil {
				continue
			}
			if inner, ok := elem.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			}
		}
		if rest := p.RestPattern; rest != nil {
			if inner, ok := rest.(ast.Pattern); ok {
				collectPatternIdentifiers(inner, into)
			} else if ident, ok := rest.(*ast.Identifier); ok && ident.Name != "" {
				into[ident.Name] = struct{}{}
			}
		}
	case *ast.TypedPattern:
		if inner, ok := p.Pattern.(ast.Pattern); ok {
			collectPatternIdentifiers(inner, into)
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
	return &runtime.FunctionValue{Declaration: expr, Closure: env}, nil
}
