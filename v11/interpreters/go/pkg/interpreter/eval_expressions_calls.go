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
			if err := i.assignPattern(param.Name, arg, localEnv, true, nil); err != nil {
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
		result, err := i.evaluateExpression(decl.Body, localEnv)
		if err != nil {
			return nil, err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		lambdaGenerics := genericNameSet(decl.GenericParams)
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
		overloads = functionOverloads(fn.Method)
		partialTarget = fn.Method
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("bound method is nil")
		}
		injected = append(injected, fn.Receiver)
		overloads = functionOverloads(fn.Method)
		partialTarget = fn.Method
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
	return i.invokeFunction(selected, evalArgs, call)
}

func (i *Interpreter) reportOverloadMismatch(fn *runtime.FunctionValue, evalArgs []runtime.Value, call *ast.FunctionCall) error {
	if fn == nil || fn.Declaration == nil {
		return nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil {
		return nil
	}
	params := decl.Params
	paramCount := len(params)
	expectedArgs := paramCount
	if decl.IsMethodShorthand {
		expectedArgs++
	}
	optionalLast := paramCount > 0 && isNullableParam(params[paramCount-1])
	paramsForCheck := params
	argsForCheck := evalArgs
	if decl.IsMethodShorthand && len(argsForCheck) > 0 {
		argsForCheck = argsForCheck[1:]
	}
	if optionalLast && len(evalArgs) == expectedArgs-1 && len(paramsForCheck) > 0 {
		paramsForCheck = paramsForCheck[:len(paramsForCheck)-1]
	}
	if len(paramsForCheck) != len(argsForCheck) {
		return nil
	}
	generics := functionGenericNameSet(fn, decl)
	for idx, param := range paramsForCheck {
		if param == nil || param.ParamType == nil {
			continue
		}
		if paramUsesGeneric(param.ParamType, generics) {
			continue
		}
		if i.matchesParamTypeForOverload(fn, param.ParamType, argsForCheck[idx]) {
			continue
		}
		name := fmt.Sprintf("param_%d", idx)
		if id, ok := param.Name.(*ast.Identifier); ok && id != nil {
			name = id.Name
		}
		expected := typeExpressionToString(param.ParamType)
		actual := describeRuntimeValue(argsForCheck[idx])
		return fmt.Errorf("Parameter type mismatch for '%s': expected %s, got %s", name, expected, actual)
	}
	return nil
}

func (i *Interpreter) matchesParamTypeForOverload(fn *runtime.FunctionValue, param ast.TypeExpression, value runtime.Value) bool {
	if param == nil {
		return true
	}
	if simple, ok := param.(*ast.SimpleTypeExpression); ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self" {
		if fn != nil && fn.MethodSet != nil && fn.MethodSet.TargetType != nil {
			return i.matchesType(fn.MethodSet.TargetType, value)
		}
	}
	return i.matchesType(param, value)
}

func (i *Interpreter) selectRuntimeOverload(overloads []*runtime.FunctionValue, evalArgs []runtime.Value, call *ast.FunctionCall) (*runtime.FunctionValue, error) {
	type candidate struct {
		fn       *runtime.FunctionValue
		score    float64
		priority float64
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
			generics := functionGenericNameSet(fn, decl)
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
					if paramUsesGeneric(param.ParamType, generics) {
						continue
					}
					if !i.matchesParamTypeForOverload(fn, param.ParamType, argsForCheck[idx]) {
						compatible = false
						break
					}
					score += float64(parameterSpecificity(param.ParamType, generics))
				}
			}
			if compatible {
				candidates = append(candidates, candidate{fn: fn, score: score, priority: fn.MethodPriority})
			}
		case *ast.LambdaExpression:
			if len(decl.Params) != len(evalArgs) {
				continue
			}
			candidates = append(candidates, candidate{fn: fn, score: 0, priority: fn.MethodPriority})
		default:
			candidates = append(candidates, candidate{fn: fn, score: 0, priority: fn.MethodPriority})
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	best := candidates[0]
	ties := []candidate{best}
	for _, cand := range candidates[1:] {
		if cand.score > best.score || (cand.score == best.score && cand.priority > best.priority) {
			best = cand
			ties = []candidate{cand}
		} else if cand.score == best.score && cand.priority == best.priority {
			ties = append(ties, cand)
		}
	}
	if len(ties) > 1 {
		return nil, fmt.Errorf("Ambiguous overload for %s", overloadName(call))
	}
	return best.fn, nil
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
	return &runtime.FunctionValue{Declaration: expr, Closure: env}, nil
}
