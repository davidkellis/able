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

func canReuseFunctionClosureEnvForBytecode(slotProgram *bytecodeProgram, decl *ast.FunctionDefinition, call *ast.FunctionCall, closure *runtime.Environment) bool {
	if decl == nil {
		return false
	}
	return canReuseCallableClosureEnvForBytecode(slotProgram, callableNeedsExplicitRuntimeTypeBindings(decl), closure)
}

func (i *Interpreter) coerceReturnValue(returnType ast.TypeExpression, value runtime.Value, genericNames map[string]struct{}, env *runtime.Environment) (runtime.Value, error) {
	if returnType == nil {
		return value, nil
	}
	if typeExpressionUsesGenerics(returnType, genericNames) {
		return value, nil
	}
	returnTypeHasAlias := i.typeExpressionReferencesAliasCached(returnType)
	if !returnTypeHasAlias && isVoidTypeExpr(returnType) {
		return runtime.VoidValue{}, nil
	}
	if !returnTypeHasAlias {
		if coerced, ok, err := i.tryFastSimpleTypeCoercion(returnType, value); ok {
			return coerced, err
		}
	}
	canonical := i.canonicalizeTypeExpressionCached(returnType, env, returnTypeHasAlias)
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
	if simple, ok := canonical.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil {
		name := normalizeKernelAliasName(simple.Name.Name)
		if !fastNamedStructTypeNameIsNonNominal(i, name) {
			if coerced, ok := exactNamedStructCoercionValueForName(value, name); ok {
				return coerced, nil
			}
		}
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

func (i *Interpreter) coerceCallableReturnValue(fn *runtime.FunctionValue, returnType ast.TypeExpression, value runtime.Value, env *runtime.Environment) (runtime.Value, error) {
	if returnType == nil {
		return value, nil
	}
	if i.functionReturnTypeUsesGenerics(fn, returnType) {
		return value, nil
	}
	return i.coerceReturnValue(returnType, value, nil, env)
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
		if ident, ok := member.Member.(*ast.Identifier); ok && ident != nil {
			receiverTypeHint := i.staticReceiverTypeForCall(call, env)
			if callable, injectedReceiver, hasInjectedReceiver, found, err := i.resolveDirectCallMemberCallable(env, target, ident.Name, receiverTypeHint); err != nil {
				return nil, err
			} else if found {
				argValues := make([]runtime.Value, 0, len(call.Arguments))
				for _, argExpr := range call.Arguments {
					val, err := i.evaluateExpression(argExpr, env)
					if err != nil {
						return nil, err
					}
					argValues = append(argValues, val)
				}
				if hasInjectedReceiver {
					return i.callCallableValueWithInjectedReceiver(callable, injectedReceiver, argValues, env, call, false)
				}
				return i.callCallableValue(callable, argValues, env, call)
			}
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
		calleeVal, found := env.Lookup(ident.Name)
		argValues := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			argValues = append(argValues, val)
		}
		if !found {
			if dotIdx := strings.Index(ident.Name, "."); dotIdx > 0 && dotIdx < len(ident.Name)-1 {
				head := ident.Name[:dotIdx]
				tail := ident.Name[dotIdx+1:]
				receiver, headFound := env.Lookup(head)
				if !headFound {
					if def, ok := env.StructDefinition(head); ok {
						receiver = def
					} else {
						receiver = runtime.TypeRefValue{TypeName: head}
					}
				}
				if callable, injectedReceiver, hasInjectedReceiver, found, err := i.resolveDirectCallMemberCallable(env, receiver, tail, nil); err != nil {
					return nil, err
				} else if found {
					if hasInjectedReceiver {
						return i.callCallableValueWithInjectedReceiver(callable, injectedReceiver, argValues, env, call, false)
					}
					return i.callCallableValue(callable, argValues, env, call)
				}
				member := ast.ID(tail)
				candidate, err := i.memberAccessOnValueWithOptions(receiver, member, env, true)
				if err != nil {
					return nil, err
				}
				return i.callCallableValue(candidate, argValues, env, call)
			}
			return nil, fmt.Errorf("Undefined variable '%s'", ident.Name)
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

func (i *Interpreter) invokeFunction(fn *runtime.FunctionValue, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall, argsMutable bool) (runtime.Value, error) {
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.Body == nil {
			return runtime.NilValue{}, nil
		}
		if call != nil && len(decl.GenericParams) > 0 {
			if err := i.populateCallTypeArguments(decl, call, args); err != nil {
				return nil, err
			}
		}
		callPlan := i.functionRuntimeGenericBindingPlan(fn)
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
		receiver, hasReceiver := resolveMethodSetReceiver(decl, args)
		if hasReceiver {
			if err := i.enforceMethodSetConstraints(fn, receiver); err != nil {
				return nil, err
			}
		}
		needsCallLocalTypeBindings := hasReceiver && callPlan.callLocalUsed
		if call != nil && len(decl.GenericParams) > 0 && callPlan.hasGenericConstraints {
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		// Check for slot-enabled bytecode program.
		var slotProgram *bytecodeProgram
		var slotLayout *bytecodeFrameLayout
		if i.execMode == execModeBytecode {
			if p, ok := fn.Bytecode.(*bytecodeProgram); ok && p != nil && p.frameLayout != nil {
				slotProgram = p
				slotLayout = p.frameLayout
			}
		}
		// When slot execution never writes call-local env bindings, reusing the
		// closure env avoids per-call Environment allocation churn.
		reuseClosureEnv := canReuseCallableClosureEnvForBytecode(slotProgram, callPlan.explicitUsed, fn.Closure)
		if reuseClosureEnv && needsCallLocalTypeBindings {
			reuseClosureEnv = false
		}
		localEnv := fn.Closure
		var transientCallEnv *runtime.Environment
		callTypeBindings := functionCallTypeBindingSet{}
		if !reuseClosureEnv {
			callTypeBindings = i.functionCallTypeBindingSetWithPlanAndEnv(fn, decl, call, receiver, needsCallLocalTypeBindings, callPlan, env)
			if reusableEnv, ok := i.reusableBytecodeCallEnvForResolvedBindings(fn, decl, call, slotProgram, callTypeBindings); ok {
				localEnv = reusableEnv
			} else if i.callableAllowsTransientCallEnvReuse(decl) {
				localEnv = i.acquireTransientCallEnvForBindingSets(
					fn.Closure,
					callTypeBindings.envValueCapacity(functionLocalBindingCapacityForLayout(decl, call, slotLayout)),
					callTypeBindings.explicit,
					callTypeBindings.callLocal,
				)
				transientCallEnv = localEnv
			} else {
				localEnv = runtime.NewEnvironmentWithBindingSets(
					fn.Closure,
					callTypeBindings.envValueCapacity(functionLocalBindingCapacityForLayout(decl, call, slotLayout)),
					callTypeBindings.explicit,
					callTypeBindings.callLocal,
				)
			}
		}
		if transientCallEnv != nil {
			defer i.releaseTransientCallEnv(transientCallEnv)
		}
		bindArgs := args
		var mutableBindArgs []runtime.Value
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
			if mutableBindArgs == nil {
				if argsMutable {
					mutableBindArgs = bindArgs
				} else {
					mutableBindArgs = append([]runtime.Value(nil), bindArgs...)
					argsMutable = true
				}
				bindArgs = mutableBindArgs
			}
			bindArgs = append(bindArgs, runtime.NilValue{})
			mutableBindArgs = bindArgs
		}
		if len(bindArgs) != paramCount {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, paramCount, len(bindArgs))
		}
		if slotLayout != nil {
			coercedArgs, mutated, err := invokeFunctionBindArgsForSlotLayout(i, fn, slotLayout, bindArgs, argsMutable)
			if err != nil {
				return nil, err
			}
			bindArgs = coercedArgs
			argsMutable = mutated
		} else {
			for idx, param := range decl.Params {
				if param == nil {
					return nil, fmt.Errorf("function parameter %d is nil", idx)
				}
				arg := bindArgs[idx]
				paramType := i.canonicalizeTypeExpressionCached(param.ParamType, fn.Closure, i.typeExpressionReferencesAliasCached(param.ParamType))
				if paramType != nil && !callPlan.paramUsesGeneric(idx) && !i.coerceValueToTypeWouldBeNoOp(paramType) && !inlineCoercionUnnecessaryWithInterpreter(i, paramType, arg) {
					coerced, err := i.coerceValueToType(paramType, arg)
					if err != nil {
						return nil, err
					}
					if mutableBindArgs == nil {
						if argsMutable {
							mutableBindArgs = bindArgs
						} else {
							mutableBindArgs = append([]runtime.Value(nil), bindArgs...)
							argsMutable = true
						}
						bindArgs = mutableBindArgs
					}
					arg = coerced
					bindArgs[idx] = coerced
				}
				if bindSimpleIdentifierPatternIntoEnv(localEnv, param.Name, arg) {
					continue
				}
				if err := i.assignPattern(param.Name, arg, localEnv, true, nil); err != nil {
					return nil, err
				}
			}
		}
		localEnv = i.asyncCallableEnv(localEnv, env)
		if slotLayoutUsesImplicitReceiver(slotLayout, hasImplicit) {
			state := i.stateFromEnv(localEnv)
			state.pushImplicitReceiver(implicitReceiver)
			defer state.popImplicitReceiver()
		}
		if thunk, ok := fn.Bytecode.(CompiledThunk); ok && thunk != nil {
			var serialSync *SerialExecutor
			if serial, ok := i.executor.(*SerialExecutor); ok {
				var payload *asyncContextPayload
				if env != nil {
					payload = payloadFromState(i.runtimeDataFromEnv(env))
				}
				if payload == nil {
					payload = payloadFromState(i.runtimeDataFromEnv(localEnv))
				}
				if payload == nil {
					serialSync = serial
					serialSync.beginSynchronousSection()
				}
			}
			if serialSync != nil {
				defer serialSync.endSynchronousSection()
			}
			thunkArgs := args
			if bytecodeCallArgsNeedMaterialization(thunkArgs) {
				if argsMutable {
					bytecodeMaterializeRawFloatArgs(thunkArgs)
				} else {
					thunkArgs = bytecodeMaterializedCallArgs(thunkArgs)
				}
			}
			result, err := thunk(localEnv, thunkArgs)
			if err != nil {
				return nil, err
			}
			if result == nil {
				result = runtime.NilValue{}
			}
			return i.coerceCallableReturnValue(fn, decl.ReturnType, result, localEnv)
		}
		if i.execMode == execModeBytecode {
			if program, ok := fn.Bytecode.(*bytecodeProgram); ok && program != nil {
				vm := i.acquireBytecodeVM(localEnv)
				defer i.releaseBytecodeVM(vm)
				if slotProgram != nil {
					layout := slotProgram.frameLayout
					slots := vm.acquireSlotFrame(layout.slotCount)
					for idx := 0; idx < len(bindArgs) && idx < layout.paramSlots; idx++ {
						slots[idx] = bindArgs[idx]
					}
					if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
						slots[layout.selfCallSlot] = fn
					}
					vm.slots = slots
				}
				result, err := vm.runDetached(program)
				if err != nil {
					if ret, ok := err.(returnSignal); ok {
						retVal := ret.value
						if retVal == nil {
							retVal = runtime.NilValue{}
						}
						if slotProgram != nil {
							return retVal, nil
						}
						coerced, err := i.coerceCallableReturnValue(fn, decl.ReturnType, retVal, localEnv)
						if err != nil {
							if ret.node != nil {
								return nil, i.attachRuntimeContext(err, ret.node, i.stateFromEnv(localEnv))
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
				if slotProgram != nil {
					return result, nil
				}
				return i.coerceCallableReturnValue(fn, decl.ReturnType, result, localEnv)
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
				coerced, err := i.coerceCallableReturnValue(fn, decl.ReturnType, retVal, localEnv)
				if err != nil {
					if ret.node != nil {
						return nil, i.attachRuntimeContext(err, ret.node, i.stateFromEnv(localEnv))
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
		return i.coerceCallableReturnValue(fn, decl.ReturnType, result, localEnv)
	case *ast.LambdaExpression:
		callPlan := i.functionRuntimeGenericBindingPlan(fn)
		if call != nil {
			if len(decl.GenericParams) > 0 {
				if err := i.populateCallTypeArguments(decl, call, args); err != nil {
					return nil, err
				}
			}
		}
		if call != nil && len(decl.GenericParams) > 0 && callPlan.hasGenericConstraints {
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		if len(args) != len(decl.Params) {
			return nil, fmt.Errorf("Lambda expects %d arguments, got %d", len(decl.Params), len(args))
		}
		var slotProgram *bytecodeProgram
		var slotLayout *bytecodeFrameLayout
		if i.execMode == execModeBytecode {
			if p, ok := fn.Bytecode.(*bytecodeProgram); ok && p != nil && p.frameLayout != nil {
				slotProgram = p
				slotLayout = p.frameLayout
			}
		}
		reuseClosureEnv := canReuseCallableClosureEnvForBytecode(slotProgram, callPlan.explicitUsed, fn.Closure)
		localEnv := fn.Closure
		var transientCallEnv *runtime.Environment
		if !reuseClosureEnv {
			explicitBindings := i.explicitCallTypeBindingValuesIfAny(decl, call)
			if i.callableAllowsTransientCallEnvReuse(decl) {
				localEnv = i.acquireTransientCallEnvForBindingSets(
					fn.Closure,
					lambdaLocalBindingCapacityForLayout(decl, call, slotLayout),
					explicitBindings,
					nil,
				)
				transientCallEnv = localEnv
			} else {
				localEnv = runtime.NewEnvironmentWithBindings(
					fn.Closure,
					lambdaLocalBindingCapacityForLayout(decl, call, slotLayout),
					explicitBindings,
				)
			}
		}
		if transientCallEnv != nil {
			defer i.releaseTransientCallEnv(transientCallEnv)
		}
		bindArgs := args
		var implicitReceiver runtime.Value
		hasImplicit := false
		if len(decl.Params) > 0 && len(args) > 0 {
			implicitReceiver = args[0]
			hasImplicit = true
		}
		if slotLayout != nil {
			coercedArgs, _, err := invokeFunctionBindArgsForSlotLayout(i, fn, slotLayout, bindArgs, argsMutable)
			if err != nil {
				return nil, err
			}
			bindArgs = coercedArgs
		} else {
			for idx, param := range decl.Params {
				if param == nil {
					return nil, fmt.Errorf("lambda parameter %d is nil", idx)
				}
				if bindSimpleIdentifierPatternIntoEnv(localEnv, param.Name, bindArgs[idx]) {
					continue
				}
				if err := i.assignPattern(param.Name, bindArgs[idx], localEnv, true, nil); err != nil {
					return nil, err
				}
			}
		}
		localEnv = i.asyncCallableEnv(localEnv, env)
		if slotLayoutUsesImplicitReceiver(slotLayout, hasImplicit) {
			state := i.stateFromEnv(localEnv)
			state.pushImplicitReceiver(implicitReceiver)
			defer state.popImplicitReceiver()
		}
		if i.execMode == execModeBytecode {
			if program, ok := fn.Bytecode.(*bytecodeProgram); ok && program != nil {
				vm := i.acquireBytecodeVM(localEnv)
				defer i.releaseBytecodeVM(vm)
				if slotProgram != nil {
					slots := vm.acquireSlotFrame(slotLayout.slotCount)
					for idx := 0; idx < len(bindArgs) && idx < slotLayout.paramSlots; idx++ {
						slots[idx] = bindArgs[idx]
					}
					vm.slots = slots
				}
				result, err := vm.runDetached(program)
				if err != nil {
					if ret, ok := err.(returnSignal); ok {
						retVal := ret.value
						if retVal == nil {
							retVal = runtime.NilValue{}
						}
						if slotProgram != nil {
							return retVal, nil
						}
						coerced, err := i.coerceCallableReturnValue(fn, decl.ReturnType, retVal, localEnv)
						if err != nil {
							if ret.node != nil {
								return nil, i.attachRuntimeContext(err, ret.node, i.stateFromEnv(localEnv))
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
				if slotProgram != nil {
					return result, nil
				}
				return i.coerceCallableReturnValue(fn, decl.ReturnType, result, localEnv)
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
				coerced, err := i.coerceCallableReturnValue(fn, decl.ReturnType, retVal, localEnv)
				if err != nil {
					if ret.node != nil {
						return nil, i.attachRuntimeContext(err, ret.node, i.stateFromEnv(localEnv))
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
		return i.coerceCallableReturnValue(fn, decl.ReturnType, result, localEnv)
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

func mergePartialCallArgs(bound []runtime.Value, args []runtime.Value) []runtime.Value {
	if len(bound) == 0 {
		return args
	}
	if len(args) == 0 {
		return bound
	}
	total := len(bound) + len(args)
	merged := make([]runtime.Value, total)
	copy(merged, bound)
	copy(merged[len(bound):], args)
	return merged
}

func (i *Interpreter) callCallableValue(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	return i.callCallableValueWithMutability(callee, args, env, call, false)
}

func (i *Interpreter) callCallableValueMutable(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall) (runtime.Value, error) {
	return i.callCallableValueWithMutability(callee, args, env, call, true)
}

func (i *Interpreter) callCallableValueWithMutability(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall, argsMutable bool) (runtime.Value, error) {
	return i.callCallableValueWithOptionalInjectedReceiver(callee, args, env, call, argsMutable, nil, false)
}

func (i *Interpreter) callCallableValueWithInjectedReceiver(callee runtime.Value, receiver runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall, argsMutable bool) (runtime.Value, error) {
	return i.callCallableValueWithOptionalInjectedReceiver(callee, args, env, call, argsMutable, receiver, true)
}

func (i *Interpreter) callCallableValueWithOptionalInjectedReceiver(callee runtime.Value, args []runtime.Value, env *runtime.Environment, call *ast.FunctionCall, argsMutable bool, injectedReceiver runtime.Value, hasInjectedReceiver bool) (runtime.Value, error) {
	if callee == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	switch fn := callee.(type) {
	case runtime.PartialFunctionValue:
		merged := mergePartialCallArgs(fn.BoundArgs, args)
		return i.callCallableValueWithOptionalInjectedReceiver(fn.Target, merged, env, call, false, injectedReceiver, hasInjectedReceiver)
	case *runtime.PartialFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("partial function is nil")
		}
		merged := mergePartialCallArgs(fn.BoundArgs, args)
		return i.callCallableValueWithOptionalInjectedReceiver(fn.Target, merged, env, call, false, injectedReceiver, hasInjectedReceiver)
	}
	if call != nil {
		state := i.stateFromEnv(env)
		state.pushCallFrame(call)
		defer state.popCallFrame()
	}
	var native runtime.NativeFunctionValue
	hasNative := false
	var directFunction *runtime.FunctionValue
	var overloads []*runtime.FunctionValue
	partialTarget := callee

	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("function is nil")
		}
		directFunction = fn
	case *runtime.FunctionOverloadValue:
		if fn == nil {
			return nil, fmt.Errorf("function overload is nil")
		}
		overloads = functionOverloadsView(fn)
	case runtime.NativeFunctionValue:
		native = fn
		hasNative = true
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("native function is nil")
		}
		native = *fn
		hasNative = true
	case runtime.NativeBoundMethodValue:
		native = fn.Method
		hasNative = true
		injectedReceiver = fn.Receiver
		hasInjectedReceiver = true
		partialTarget = fn.Method
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native bound method is nil")
		}
		native = fn.Method
		hasNative = true
		injectedReceiver = fn.Receiver
		hasInjectedReceiver = true
		partialTarget = fn.Method
	case runtime.DynRefValue:
		resolved, err := i.resolveDynRef(fn)
		if err != nil {
			return nil, err
		}
		return i.callCallableValueWithOptionalInjectedReceiver(resolved, args, env, call, argsMutable, injectedReceiver, hasInjectedReceiver)
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, err := i.resolveDynRef(*fn)
		if err != nil {
			return nil, err
		}
		return i.callCallableValueWithOptionalInjectedReceiver(resolved, args, env, call, argsMutable, injectedReceiver, hasInjectedReceiver)
	case runtime.BoundMethodValue:
		injectedReceiver = fn.Receiver
		hasInjectedReceiver = true
		if nfv, ok := fn.Method.(*runtime.NativeFunctionValue); ok && nfv != nil {
			native = *nfv
			hasNative = true
			partialTarget = fn.Method
		} else if nfv, ok := fn.Method.(runtime.NativeFunctionValue); ok {
			native = nfv
			hasNative = true
			partialTarget = fn.Method
		} else if methodFn, ok := fn.Method.(*runtime.FunctionValue); ok && methodFn != nil {
			directFunction = methodFn
			partialTarget = fn.Method
		} else {
			overloads = functionOverloadsView(fn.Method)
			partialTarget = fn.Method
		}
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("bound method is nil")
		}
		injectedReceiver = fn.Receiver
		hasInjectedReceiver = true
		if nfv, ok := fn.Method.(*runtime.NativeFunctionValue); ok && nfv != nil {
			native = *nfv
			hasNative = true
			partialTarget = fn.Method
		} else if nfv, ok := fn.Method.(runtime.NativeFunctionValue); ok {
			native = nfv
			hasNative = true
			partialTarget = fn.Method
		} else if methodFn, ok := fn.Method.(*runtime.FunctionValue); ok && methodFn != nil {
			directFunction = methodFn
			partialTarget = fn.Method
		} else {
			overloads = functionOverloadsView(fn.Method)
			partialTarget = fn.Method
		}
	default:
		overloads = functionOverloadsView(callee)
	}

	evalArgs := args
	if hasInjectedReceiver {
		evalArgs = prependReceiverCallArgs(injectedReceiver, args, argsMutable)
		argsMutable = true
	}

	if hasNative {
		if native.Arity >= 0 {
			provided := len(args)
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
		var callState any
		if !native.SkipContext && env != nil {
			callState = i.runtimeDataFromEnv(env)
		}
		return i.invokeNativeFunctionValue(native, env, callState, evalArgs)
	}
	if directFunction != nil {
		return i.callResolvedFunctionValue(directFunction, partialTarget, evalArgs, env, call, argsMutable)
	}

	if len(overloads) == 0 {
		if applyMethod, err := i.findApplyMethod(callee); err == nil && applyMethod != nil {
			bound := runtime.BoundMethodValue{Receiver: callee, Method: applyMethod}
			return i.callCallableValueWithOptionalInjectedReceiver(bound, args, env, call, false, nil, false)
		} else if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("calling non-function value of kind %s (%T)", callee.Kind(), callee)
	}

	if len(overloads) == 1 {
		only := overloads[0]
		minRequired := minArgsForFunctionValue(only)
		if len(evalArgs) < minRequired {
			return makePartialFunctionValue(partialTarget, evalArgs, call), nil
		}
		if i.matchesSingleRuntimeOverload(only, evalArgs) {
			return i.invokeFunction(only, evalArgs, env, call, argsMutable)
		}
		if mismatchErr := i.reportOverloadMismatch(only, evalArgs, call); mismatchErr != nil {
			return nil, mismatchErr
		}
		return nil, fmt.Errorf("No overloads of %s match provided arguments", overloadName(call))
	}

	if len(overloads) > 1 {
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
		return nil, fmt.Errorf("No overloads of %s match provided arguments", overloadName(call))
	}
	return i.invokeFunction(selected, evalArgs, env, call, argsMutable)
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

func (i *Interpreter) evaluateIteratorLiteral(expr *ast.IteratorLiteral, env *runtime.Environment) (runtime.Value, error) {
	iterCapacity := 1
	bindingName := "gen"
	if expr.Binding != nil && expr.Binding.Name != "" {
		bindingName = expr.Binding.Name
		if bindingName != "gen" {
			iterCapacity = 2
		}
	}
	iterEnv := runtime.NewEnvironmentWithValueCapacity(env, iterCapacity)
	instance := newGeneratorInstance(i, iterEnv, expr.Body)
	controller := instance.controllerValue()
	iterEnv.DefineWithoutMerge(bindingName, controller)
	if bindingName != "gen" {
		iterEnv.DefineWithoutMerge("gen", controller)
	}
	return runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return instance.nextRaw()
	}, instance.close), nil
}

func (i *Interpreter) evaluateLambdaExpression(expr *ast.LambdaExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr == nil {
		return nil, fmt.Errorf("lambda expression is nil")
	}
	fnVal := &runtime.FunctionValue{Declaration: expr, Closure: env}
	if expr.Body != nil {
		program, err := i.lowerLambdaExpressionBytecodeWithEnv(expr, env)
		if err != nil {
			if i.execMode == execModeBytecode {
				return nil, err
			}
		} else {
			setFunctionBytecodeProgram(fnVal, program)
		}
	}
	return fnVal, nil
}
