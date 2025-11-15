package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateFunctionCall(call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, error) {
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
			if err := i.enforceGenericConstraintsIfAny(decl, call); err != nil {
				return nil, err
			}
		}
		paramCount := len(decl.Params)
		expectedArgs := paramCount
		if decl.IsMethodShorthand {
			expectedArgs++
		}
		if len(args) != expectedArgs {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, expectedArgs, len(args))
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
	switch fn := callee.(type) {
	case runtime.NativeFunctionValue:
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Impl(ctx, args)
	case *runtime.NativeFunctionValue:
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Impl(ctx, args)
	case runtime.NativeBoundMethodValue:
		combined := append([]runtime.Value{fn.Receiver}, args...)
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Method.Impl(ctx, combined)
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native bound method is nil")
		}
		combined := append([]runtime.Value{fn.Receiver}, args...)
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Method.Impl(ctx, combined)
	case runtime.DynRefValue:
		resolved, err := i.resolveDynRef(fn)
		if err != nil {
			return nil, err
		}
		return i.invokeFunction(resolved, args, call)
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, err := i.resolveDynRef(*fn)
		if err != nil {
			return nil, err
		}
		return i.invokeFunction(resolved, args, call)
	case runtime.BoundMethodValue:
		combined := append([]runtime.Value{fn.Receiver}, args...)
		return i.invokeFunction(fn.Method, combined, call)
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("bound method is nil")
		}
		combined := append([]runtime.Value{fn.Receiver}, args...)
		return i.invokeFunction(fn.Method, combined, call)
	case *runtime.FunctionValue:
		return i.invokeFunction(fn, args, call)
	default:
		return nil, fmt.Errorf("calling non-function value of kind %s", callee.Kind())
	}
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
			intMember, ok := lhs.Member.(*ast.IntegerLiteral)
			if !ok {
				return nil, fmt.Errorf("Array member assignment requires integer member")
			}
			if intMember.Value == nil {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			idx := int(intMember.Value.Int64())
			if idx < 0 || idx >= len(arrayVal.Elements) {
				return nil, fmt.Errorf("Array index out of bounds")
			}
			if assign.Operator == ast.AssignmentAssign {
				arrayVal.Elements[idx] = value
				return value, nil
			}
			if !isCompound {
				return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
			}
			current := arrayVal.Elements[idx]
			computed, err := applyBinaryOperator(binaryOp, current, value)
			if err != nil {
				return nil, err
			}
			arrayVal.Elements[idx] = computed
			return computed, nil
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
		arr, err := toArrayValue(arrObj)
		if err != nil {
			return nil, err
		}
		idxVal, err := i.evaluateExpression(lhs.Index, env)
		if err != nil {
			return nil, err
		}
		idx, err := indexFromValue(idxVal)
		if err != nil {
			return nil, err
		}
		if idx < 0 || idx >= len(arr.Elements) {
			return nil, fmt.Errorf("Array index out of bounds")
		}
		if assign.Operator == ast.AssignmentAssign {
			arr.Elements[idx] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		current := arr.Elements[idx]
		computed, err := applyBinaryOperator(binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		arr.Elements[idx] = computed
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
