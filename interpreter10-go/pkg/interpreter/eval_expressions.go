package interpreter

import (
	"fmt"
	"math/big"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateExpression(node ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	if node == nil {
		return runtime.NilValue{}, nil
	}
	state := i.stateFromEnv(env)
	if !state.hasPlaceholderFrame() {
		if value, ok, err := i.tryBuildPlaceholderFunction(node, env); err != nil {
			return nil, err
		} else if ok {
			return value, nil
		}
	}
	switch n := node.(type) {
	case *ast.StringLiteral:
		return runtime.StringValue{Val: n.Value}, nil
	case *ast.BooleanLiteral:
		return runtime.BoolValue{Val: n.Value}, nil
	case *ast.CharLiteral:
		if len(n.Value) == 0 {
			return nil, fmt.Errorf("empty char literal")
		}
		return runtime.CharValue{Val: []rune(n.Value)[0]}, nil
	case *ast.NilLiteral:
		return runtime.NilValue{}, nil
	case *ast.IntegerLiteral:
		suffix := runtime.IntegerI32
		if n.IntegerType != nil {
			suffix = runtime.IntegerType(*n.IntegerType)
		}
		return runtime.IntegerValue{Val: runtime.CloneBigInt(bigFromLiteral(n.Value)), TypeSuffix: suffix}, nil
	case *ast.FloatLiteral:
		suffix := runtime.FloatF64
		if n.FloatType != nil {
			suffix = runtime.FloatType(*n.FloatType)
		}
		return runtime.FloatValue{Val: n.Value, TypeSuffix: suffix}, nil
	case *ast.ArrayLiteral:
		values := make([]runtime.Value, 0, len(n.Elements))
		for _, el := range n.Elements {
			val, err := i.evaluateExpression(el, env)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		return &runtime.ArrayValue{Elements: values}, nil
	case *ast.StringInterpolation:
		var builder strings.Builder
		for _, part := range n.Parts {
			if part == nil {
				return nil, fmt.Errorf("string interpolation contains nil part")
			}
			if lit, ok := part.(*ast.StringLiteral); ok {
				builder.WriteString(lit.Value)
				continue
			}
			val, err := i.evaluateExpression(part, env)
			if err != nil {
				return nil, err
			}
			str, err := i.stringifyValue(val, env)
			if err != nil {
				return nil, err
			}
			builder.WriteString(str)
		}
		return runtime.StringValue{Val: builder.String()}, nil
	case *ast.BreakpointExpression:
		return i.evaluateBreakpointExpression(n, env)
	case *ast.RangeExpression:
		start, err := i.evaluateExpression(n.Start, env)
		if err != nil {
			return nil, err
		}
		endExpr, err := i.evaluateExpression(n.End, env)
		if err != nil {
			return nil, err
		}
		if !isNumericValue(start) || !isNumericValue(endExpr) {
			return nil, fmt.Errorf("Range boundaries must be numeric")
		}
		return &runtime.RangeValue{Start: start, End: endExpr, Inclusive: n.Inclusive}, nil
	case *ast.StructLiteral:
		return i.evaluateStructLiteral(n, env)
	case *ast.MatchExpression:
		return i.evaluateMatchExpression(n, env)
	case *ast.PropagationExpression:
		return i.evaluatePropagationExpression(n, env)
	case *ast.OrElseExpression:
		return i.evaluateOrElseExpression(n, env)
	case *ast.EnsureExpression:
		return i.evaluateEnsureExpression(n, env)
	case *ast.MemberAccessExpression:
		return i.evaluateMemberAccess(n, env)
	case *ast.ImplicitMemberExpression:
		return i.evaluateImplicitMemberExpression(n, env)
	case *ast.IndexExpression:
		return i.evaluateIndexExpression(n, env)
	case *ast.UnaryExpression:
		return i.evaluateUnaryExpression(n, env)
	case *ast.TopicReferenceExpression:
		return i.evaluateTopicReferenceExpression(n, env)
	case *ast.PlaceholderExpression:
		return i.evaluatePlaceholderExpression(n, env)
	case *ast.Identifier:
		val, err := env.Get(n.Name)
		if err != nil {
			return nil, err
		}
		return val, nil
	case *ast.FunctionCall:
		return i.evaluateFunctionCall(n, env)
	case *ast.BinaryExpression:
		return i.evaluateBinaryExpression(n, env)
	case *ast.AssignmentExpression:
		return i.evaluateAssignment(n, env)
	case *ast.BlockExpression:
		return i.evaluateBlock(n, env)
	case *ast.IteratorLiteral:
		return i.evaluateIteratorLiteral(n, env)
	case *ast.IfExpression:
		return i.evaluateIfExpression(n, env)
	case *ast.RescueExpression:
		return i.evaluateRescueExpression(n, env)
	case *ast.ProcExpression:
		i.ensureConcurrencyBuiltins()
		task := i.makeAsyncTask(asyncContextProc, n.Expression, env)
		handle := i.executor.RunProc(task)
		return handle, nil
	case *ast.SpawnExpression:
		i.ensureConcurrencyBuiltins()
		task := i.makeAsyncTask(asyncContextFuture, n.Expression, env)
		future := i.executor.RunFuture(task)
		return future, nil
	case *ast.LambdaExpression:
		return i.evaluateLambdaExpression(n, env)
	default:
		return nil, fmt.Errorf("unsupported expression type: %s", n.NodeType())
	}
}

func (i *Interpreter) evaluateIfExpression(expr *ast.IfExpression, env *runtime.Environment) (runtime.Value, error) {
	cond, err := i.evaluateExpression(expr.IfCondition, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(cond) {
		return i.evaluateBlock(expr.IfBody, env)
	}
	for _, clause := range expr.OrClauses {
		if clause.Condition != nil {
			clauseCond, err := i.evaluateExpression(clause.Condition, env)
			if err != nil {
				return nil, err
			}
			if !isTruthy(clauseCond) {
				continue
			}
		}
		return i.evaluateBlock(clause.Body, env)
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateMatchExpression(expr *ast.MatchExpression, env *runtime.Environment) (runtime.Value, error) {
	subject, err := i.evaluateExpression(expr.Subject, env)
	if err != nil {
		return nil, err
	}
	for _, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		clauseEnv, matched := i.matchPattern(clause.Pattern, subject, env)
		if !matched {
			continue
		}
		if clause.Guard != nil {
			guardVal, err := i.evaluateExpression(clause.Guard, clauseEnv)
			if err != nil {
				return nil, err
			}
			if !isTruthy(guardVal) {
				continue
			}
		}
		return i.evaluateExpression(clause.Body, clauseEnv)
	}
	return nil, fmt.Errorf("Non-exhaustive match")
}

func (i *Interpreter) evaluateRescueExpression(expr *ast.RescueExpression, env *runtime.Environment) (runtime.Value, error) {
	result, err := i.evaluateExpression(expr.MonitoredExpression, env)
	if err == nil {
		return result, nil
	}
	rs, ok := err.(raiseSignal)
	if !ok {
		return nil, err
	}
	for _, clause := range expr.Clauses {
		clauseEnv, matched := i.matchPattern(clause.Pattern, rs.value, env)
		if !matched {
			continue
		}
		state := i.stateFromEnv(clauseEnv)
		state.pushRaise(rs.value)
		if clause.Guard != nil {
			guardVal, gErr := i.evaluateExpression(clause.Guard, clauseEnv)
			if gErr != nil {
				state.popRaise()
				return nil, gErr
			}
			if !isTruthy(guardVal) {
				state.popRaise()
				continue
			}
		}
		result, bodyErr := i.evaluateExpression(clause.Body, clauseEnv)
		state.popRaise()
		if bodyErr != nil {
			return nil, bodyErr
		}
		return result, nil
	}
	return nil, rs
}

func (i *Interpreter) evaluatePropagationExpression(expr *ast.PropagationExpression, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		return nil, err
	}
	if errVal, ok := val.(runtime.ErrorValue); ok {
		return nil, raiseSignal{value: errVal}
	}
	return val, nil
}

func (i *Interpreter) evaluateOrElseExpression(expr *ast.OrElseExpression, env *runtime.Environment) (runtime.Value, error) {
	val, err := i.evaluateExpression(expr.Expression, env)
	if err != nil {
		if rs, ok := err.(raiseSignal); ok {
			handlerEnv := runtime.NewEnvironment(env)
			if expr.ErrorBinding != nil {
				handlerEnv.Define(expr.ErrorBinding.Name, rs.value)
			}
			result, handlerErr := i.evaluateBlock(expr.Handler, handlerEnv)
			if handlerErr != nil {
				return nil, handlerErr
			}
			if result == nil {
				return runtime.NilValue{}, nil
			}
			return result, nil
		}
		return nil, err
	}
	return val, nil
}

func (i *Interpreter) evaluateEnsureExpression(expr *ast.EnsureExpression, env *runtime.Environment) (runtime.Value, error) {
	var (
		tryResult runtime.Value = runtime.NilValue{}
		execErr   error
	)
	val, err := i.evaluateExpression(expr.TryExpression, env)
	if err == nil {
		if val != nil {
			tryResult = val
		}
	} else {
		execErr = err
	}
	if expr.EnsureBlock != nil {
		if _, ensureErr := i.evaluateBlock(expr.EnsureBlock, env); ensureErr != nil {
			return nil, ensureErr
		}
	}
	if execErr != nil {
		return nil, execErr
	}
	if tryResult == nil {
		return runtime.NilValue{}, nil
	}
	return tryResult, nil
}

func (i *Interpreter) evaluateBreakpointExpression(expr *ast.BreakpointExpression, env *runtime.Environment) (runtime.Value, error) {
	if expr.Label == nil {
		return nil, fmt.Errorf("Breakpoint expression requires label")
	}
	label := expr.Label.Name
	state := i.stateFromEnv(env)
	state.pushBreakpoint(label)
	defer state.popBreakpoint()
	for {
		val, err := i.evaluateBlock(expr.Body, env)
		if err != nil {
			switch sig := err.(type) {
			case breakSignal:
				if sig.label == label {
					return sig.value, nil
				}
				return nil, sig
			case continueSignal:
				if sig.label == label {
					continue
				}
				return nil, sig
			default:
				return nil, err
			}
		}
		if val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}
}

func (i *Interpreter) evaluateUnaryExpression(expr *ast.UnaryExpression, env *runtime.Environment) (runtime.Value, error) {
	operand, err := i.evaluateExpression(expr.Operand, env)
	if err != nil {
		return nil, err
	}
	switch expr.Operator {
	case "-":
		switch v := operand.(type) {
		case runtime.IntegerValue:
			neg := new(big.Int).Neg(v.Val)
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, nil
		case runtime.FloatValue:
			return runtime.FloatValue{Val: -v.Val, TypeSuffix: v.TypeSuffix}, nil
		default:
			return nil, fmt.Errorf("unary '-' not supported for %T", operand)
		}
	case "!":
		if bv, ok := operand.(runtime.BoolValue); ok {
			return runtime.BoolValue{Val: !bv.Val}, nil
		}
		return nil, fmt.Errorf("unary '!' expects bool, got %T", operand)
	case "~":
		switch v := operand.(type) {
		case runtime.IntegerValue:
			if strings.HasPrefix(string(v.TypeSuffix), "u") {
				width := integerBitWidth(v.TypeSuffix)
				if width <= 0 {
					return nil, fmt.Errorf("unsupported integer width for bitwise not")
				}
				mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(width)), big.NewInt(1))
				val := new(big.Int).Set(v.Val)
				if val.Sign() < 0 {
					return nil, fmt.Errorf("bitwise not on unsigned requires non-negative operand")
				}
				result := new(big.Int).Xor(mask, val)
				return runtime.IntegerValue{Val: result, TypeSuffix: v.TypeSuffix}, nil
			}
			neg := new(big.Int).Neg(new(big.Int).Add(v.Val, big.NewInt(1)))
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, nil
		default:
			return nil, fmt.Errorf("unary '~' not supported for %T", operand)
		}
	default:
		return nil, fmt.Errorf("unsupported unary operator %s", expr.Operator)
	}
}

func (i *Interpreter) evaluateBinaryExpression(expr *ast.BinaryExpression, env *runtime.Environment) (runtime.Value, error) {
	leftVal, err := i.evaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}
	if expr.Operator == "|>" {
		return i.evaluatePipeExpression(leftVal, expr.Right, env)
	}
	switch expr.Operator {
	case "&&":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		if !lb.Val {
			return runtime.BoolValue{Val: false}, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	case "||":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		if lb.Val {
			return runtime.BoolValue{Val: true}, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("Logical operands must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	default:
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		if expr.Operator == "+" {
			if ls, ok := leftVal.(runtime.StringValue); ok {
				rs, ok := rightVal.(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("Arithmetic requires numeric operands")
				}
				return runtime.StringValue{Val: ls.Val + rs.Val}, nil
			}
			if _, ok := rightVal.(runtime.StringValue); ok {
				return nil, fmt.Errorf("Arithmetic requires numeric operands")
			}
		}
		return applyBinaryOperator(expr.Operator, leftVal, rightVal)
	}
}

func comparisonOp(op string, cmp int) bool {
	switch op {
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	default:
		return false
	}
}

func isTruthy(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.BoolValue:
		return v.Val
	case runtime.NilValue:
		return false
	default:
		return true
	}
}

func isNumericValue(val runtime.Value) bool {
	switch val.(type) {
	case runtime.IntegerValue, runtime.FloatValue:
		return true
	default:
		return false
	}
}

func numericToFloat(val runtime.Value) (float64, error) {
	switch v := val.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case runtime.IntegerValue:
		int32Val, err := int32FromIntegerValue(v)
		if err != nil {
			return 0, err
		}
		return float64(int32Val), nil
	default:
		return 0, fmt.Errorf("Arithmetic requires numeric operands")
	}
}

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
			if err := i.assignPattern(param.Name, bindArgs[idx], localEnv, true); err != nil {
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
			if err := i.assignPattern(param.Name, args[idx], localEnv, true); err != nil {
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

func (i *Interpreter) evaluatePlaceholderExpression(expr *ast.PlaceholderExpression, env *runtime.Environment) (runtime.Value, error) {
	state := i.stateFromEnv(env)
	frame, ok := state.currentPlaceholderFrame()
	if !ok {
		return nil, fmt.Errorf("Expression placeholder used outside of placeholder lambda")
	}
	if expr.Index != nil {
		idx := *expr.Index
		val, err := frame.valueAt(idx)
		if err != nil {
			return nil, err
		}
		if val == nil {
			return runtime.NilValue{}, nil
		}
		return val, nil
	}
	idx, err := frame.nextImplicitIndex()
	if err != nil {
		return nil, err
	}
	val, err := frame.valueAt(idx)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
}

func (i *Interpreter) tryBuildPlaceholderFunction(node ast.Expression, env *runtime.Environment) (runtime.Value, bool, error) {
	plan, ok, err := analyzePlaceholderExpression(node)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	if call, isCall := node.(*ast.FunctionCall); isCall {
		calleeHas := expressionContainsPlaceholder(call.Callee)
		argsHave := false
		for _, arg := range call.Arguments {
			if expressionContainsPlaceholder(arg) {
				argsHave = true
				break
			}
		}
		if calleeHas && !argsHave {
			return nil, false, nil
		}
	}
	closure := &placeholderClosure{
		interpreter: i,
		expression:  node,
		env:         env,
		plan:        plan,
	}
	fn := runtime.NativeFunctionValue{
		Name:  "<placeholder>",
		Arity: plan.paramCount,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			return closure.invoke(args)
		},
	}
	return fn, true, nil
}

type placeholderPlan struct {
	explicitIndices map[int]struct{}
	paramCount      int
}

type placeholderContext int

const (
	contextRoot placeholderContext = iota
	contextCallCallee
	contextOther
)

type placeholderAnalyzer struct {
	explicit        map[int]struct{}
	implicitCount   int
	highestExplicit int
	hasPlaceholder  bool
	relevant        bool
}

func expressionContainsPlaceholder(expr ast.Expression) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.PlaceholderExpression:
		return true
	case *ast.BinaryExpression:
		return expressionContainsPlaceholder(e.Left) || expressionContainsPlaceholder(e.Right)
	case *ast.UnaryExpression:
		return expressionContainsPlaceholder(e.Operand)
	case *ast.FunctionCall:
		if expressionContainsPlaceholder(e.Callee) {
			return true
		}
		for _, arg := range e.Arguments {
			if expressionContainsPlaceholder(arg) {
				return true
			}
		}
		return false
	case *ast.MemberAccessExpression:
		if expressionContainsPlaceholder(e.Object) {
			return true
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return expressionContainsPlaceholder(memberExpr)
		}
		return false
	case *ast.ImplicitMemberExpression:
		return false
	case *ast.IndexExpression:
		return expressionContainsPlaceholder(e.Object) || expressionContainsPlaceholder(e.Index)
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if statementContainsPlaceholder(stmt) {
				return true
			}
		}
		return false
	case *ast.AssignmentExpression:
		if expressionContainsPlaceholder(e.Right) {
			return true
		}
		if targetExpr, ok := e.Left.(ast.Expression); ok {
			return expressionContainsPlaceholder(targetExpr)
		}
		return false
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if expressionContainsPlaceholder(part) {
				return true
			}
		}
		return false
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil && expressionContainsPlaceholder(field.Value) {
				return true
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if expressionContainsPlaceholder(src) {
				return true
			}
		}
		return false
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if expressionContainsPlaceholder(el) {
				return true
			}
		}
		return false
	case *ast.RangeExpression:
		return expressionContainsPlaceholder(e.Start) || expressionContainsPlaceholder(e.End)
	case *ast.MatchExpression:
		if expressionContainsPlaceholder(e.Subject) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && expressionContainsPlaceholder(clause.Guard) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.OrElseExpression:
		return expressionContainsPlaceholder(e.Expression) || expressionContainsPlaceholder(e.Handler)
	case *ast.RescueExpression:
		if expressionContainsPlaceholder(e.MonitoredExpression) {
			return true
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil && expressionContainsPlaceholder(clause.Guard) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.EnsureExpression:
		return expressionContainsPlaceholder(e.TryExpression) || expressionContainsPlaceholder(e.EnsureBlock)
	case *ast.IfExpression:
		if expressionContainsPlaceholder(e.IfCondition) || expressionContainsPlaceholder(e.IfBody) {
			return true
		}
		for _, clause := range e.OrClauses {
			if clause == nil {
				continue
			}
			if clause.Condition != nil && expressionContainsPlaceholder(clause.Condition) {
				return true
			}
			if expressionContainsPlaceholder(clause.Body) {
				return true
			}
		}
		return false
	case *ast.IteratorLiteral:
		return false
	case *ast.LambdaExpression:
		return false
	case *ast.ProcExpression, *ast.SpawnExpression:
		return false
	case *ast.TopicReferenceExpression,
		*ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return false
	default:
		return false
	}
}

func statementContainsPlaceholder(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return expressionContainsPlaceholder(expr)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return expressionContainsPlaceholder(s.Argument)
		}
	case *ast.RaiseStatement:
		if s.Expression != nil {
			return expressionContainsPlaceholder(s.Expression)
		}
	case *ast.ForLoop:
		if expressionContainsPlaceholder(s.Iterable) {
			return true
		}
		return expressionContainsPlaceholder(s.Body)
	case *ast.WhileLoop:
		if expressionContainsPlaceholder(s.Condition) {
			return true
		}
		return expressionContainsPlaceholder(s.Body)
	case *ast.BreakStatement:
		if s.Value != nil {
			return expressionContainsPlaceholder(s.Value)
		}
	case *ast.ContinueStatement:
		return false
	case *ast.YieldStatement:
		if s.Expression != nil {
			return expressionContainsPlaceholder(s.Expression)
		}
	case *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return false
	default:
		return false
	}
	return false
}

func analyzePlaceholderExpression(expr ast.Expression) (placeholderPlan, bool, error) {
	analyzer := &placeholderAnalyzer{
		explicit: make(map[int]struct{}),
	}
	if err := analyzer.visitExpression(expr); err != nil {
		return placeholderPlan{}, false, err
	}
	if !analyzer.hasPlaceholder {
		return placeholderPlan{}, false, nil
	}
	paramCount := analyzer.highestExplicit
	implicitTotal := len(analyzer.explicit) + analyzer.implicitCount
	if implicitTotal > paramCount {
		paramCount = implicitTotal
	}
	return placeholderPlan{
		explicitIndices: analyzer.explicit,
		paramCount:      paramCount,
	}, true, nil
}

func (p *placeholderAnalyzer) visitExpression(expr ast.Expression) error {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.PlaceholderExpression:
		p.hasPlaceholder = true
		if e.Index != nil {
			idx := *e.Index
			if idx <= 0 {
				return fmt.Errorf("Placeholder index must be positive, found @%d", idx)
			}
			p.explicit[idx] = struct{}{}
			if idx > p.highestExplicit {
				p.highestExplicit = idx
			}
		} else {
			p.implicitCount++
		}
	case *ast.BinaryExpression:
		if err := p.visitExpression(e.Left); err != nil {
			return err
		}
		return p.visitExpression(e.Right)
	case *ast.UnaryExpression:
		return p.visitExpression(e.Operand)
	case *ast.FunctionCall:
		if err := p.visitExpression(e.Callee); err != nil {
			return err
		}
		for _, arg := range e.Arguments {
			if err := p.visitExpression(arg); err != nil {
				return err
			}
		}
		return nil
	case *ast.MemberAccessExpression:
		if err := p.visitExpression(e.Object); err != nil {
			return err
		}
		if memberExpr, ok := e.Member.(ast.Expression); ok {
			return p.visitExpression(memberExpr)
		}
		return nil
	case *ast.ImplicitMemberExpression:
		return nil
	case *ast.IndexExpression:
		if err := p.visitExpression(e.Object); err != nil {
			return err
		}
		return p.visitExpression(e.Index)
	case *ast.BlockExpression:
		for _, stmt := range e.Body {
			if err := p.visitStatement(stmt); err != nil {
				return err
			}
		}
		return nil
	case *ast.AssignmentExpression:
		if err := p.visitExpression(e.Right); err != nil {
			return err
		}
		if targetExpr, ok := e.Left.(ast.Expression); ok {
			return p.visitExpression(targetExpr)
		}
		return nil
	case *ast.StringInterpolation:
		for _, part := range e.Parts {
			if err := p.visitExpression(part); err != nil {
				return err
			}
		}
		return nil
	case *ast.StructLiteral:
		for _, field := range e.Fields {
			if field != nil {
				if err := p.visitExpression(field.Value); err != nil {
					return err
				}
			}
		}
		for _, src := range e.FunctionalUpdateSources {
			if err := p.visitExpression(src); err != nil {
				return err
			}
		}
		return nil
	case *ast.ArrayLiteral:
		for _, el := range e.Elements {
			if err := p.visitExpression(el); err != nil {
				return err
			}
		}
		return nil
	case *ast.RangeExpression:
		if err := p.visitExpression(e.Start); err != nil {
			return err
		}
		return p.visitExpression(e.End)
	case *ast.MatchExpression:
		if err := p.visitExpression(e.Subject); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.OrElseExpression:
		if err := p.visitExpression(e.Expression); err != nil {
			return err
		}
		return p.visitExpression(e.Handler)
	case *ast.RescueExpression:
		if err := p.visitExpression(e.MonitoredExpression); err != nil {
			return err
		}
		for _, clause := range e.Clauses {
			if clause == nil {
				continue
			}
			if clause.Guard != nil {
				if err := p.visitExpression(clause.Guard); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.EnsureExpression:
		if err := p.visitExpression(e.TryExpression); err != nil {
			return err
		}
		return p.visitExpression(e.EnsureBlock)
	case *ast.IfExpression:
		if err := p.visitExpression(e.IfCondition); err != nil {
			return err
		}
		if err := p.visitExpression(e.IfBody); err != nil {
			return err
		}
		for _, clause := range e.OrClauses {
			if clause == nil {
				continue
			}
			if clause.Condition != nil {
				if err := p.visitExpression(clause.Condition); err != nil {
					return err
				}
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		return nil
	case *ast.IteratorLiteral:
		return nil
	case *ast.LambdaExpression:
		return nil
	case *ast.ProcExpression, *ast.SpawnExpression:
		return nil
	case *ast.TopicReferenceExpression,
		*ast.Identifier,
		*ast.IntegerLiteral,
		*ast.FloatLiteral,
		*ast.BooleanLiteral,
		*ast.StringLiteral,
		*ast.CharLiteral,
		*ast.NilLiteral:
		return nil
	default:
		return nil
	}
	return nil
}

func (p *placeholderAnalyzer) visitStatement(stmt ast.Statement) error {
	if stmt == nil {
		return nil
	}
	if expr, ok := stmt.(ast.Expression); ok {
		return p.visitExpression(expr)
	}
	switch s := stmt.(type) {
	case *ast.ReturnStatement:
		if s.Argument != nil {
			return p.visitExpression(s.Argument)
		}
	case *ast.RaiseStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression)
		}
	case *ast.ForLoop:
		if err := p.visitExpression(s.Iterable); err != nil {
			return err
		}
		return p.visitExpression(s.Body)
	case *ast.WhileLoop:
		if err := p.visitExpression(s.Condition); err != nil {
			return err
		}
		return p.visitExpression(s.Body)
	case *ast.BreakStatement:
		if s.Value != nil {
			return p.visitExpression(s.Value)
		}
	case *ast.ContinueStatement:
		return nil
	case *ast.YieldStatement:
		if s.Expression != nil {
			return p.visitExpression(s.Expression)
		}
	case *ast.PreludeStatement, *ast.ExternFunctionBody, *ast.ImportStatement, *ast.DynImportStatement, *ast.PackageStatement:
		return nil
	default:
		return nil
	}
	return nil
}

type placeholderClosure struct {
	interpreter *Interpreter
	expression  ast.Expression
	env         *runtime.Environment
	plan        placeholderPlan
}

func (p *placeholderClosure) invoke(args []runtime.Value) (runtime.Value, error) {
	if len(args) != p.plan.paramCount {
		return nil, fmt.Errorf("Placeholder lambda expects %d arguments, got %d", p.plan.paramCount, len(args))
	}
	callEnv := runtime.NewEnvironment(p.env)
	state := p.interpreter.stateFromEnv(callEnv)
	state.pushPlaceholderFrame(p.plan.explicitIndices, p.plan.paramCount, args)
	defer state.popPlaceholderFrame()
	result, err := p.interpreter.evaluateExpression(p.expression, callEnv)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
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

func extractFunctionGenerics(funcNode ast.Node) ([]*ast.GenericParameter, []*ast.WhereClauseConstraint) {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		return fn.GenericParams, fn.WhereClause
	case *ast.LambdaExpression:
		return fn.GenericParams, fn.WhereClause
	default:
		return nil, nil
	}
}

func functionNameForErrors(funcNode ast.Node) string {
	switch fn := funcNode.(type) {
	case *ast.FunctionDefinition:
		if fn.ID != nil && fn.ID.Name != "" {
			return fn.ID.Name
		}
	case *ast.LambdaExpression:
		return "(lambda)"
	}
	return "(lambda)"
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
			env.Define(lhs.Name, value)
		case ast.AssignmentAssign:
			if err := env.Assign(lhs.Name, value); err != nil {
				return nil, err
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
		isDeclaration := assign.Operator == ast.AssignmentDeclare
		if !isDeclaration && assign.Operator != ast.AssignmentAssign {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
		if err := i.assignPattern(lhs, value, env, isDeclaration); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported assignment target %s", lhs.NodeType())
	}

	return value, nil
}

func assignStructMember(inst *runtime.StructInstanceValue, member ast.Expression, value runtime.Value, operator ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("struct instance is nil")
	}
	switch mem := member.(type) {
	case *ast.Identifier:
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		current, ok := inst.Fields[mem.Name]
		if !ok {
			return nil, fmt.Errorf("No field named '%s'", mem.Name)
		}
		if operator == ast.AssignmentAssign {
			inst.Fields[mem.Name] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", operator)
		}
		computed, err := applyBinaryOperator(binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		inst.Fields[mem.Name] = computed
		return computed, nil
	case *ast.IntegerLiteral:
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if mem.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(mem.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		if operator == ast.AssignmentAssign {
			inst.Positional[idx] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", operator)
		}
		current := inst.Positional[idx]
		computed, err := applyBinaryOperator(binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		inst.Positional[idx] = computed
		return computed, nil
	default:
		return nil, fmt.Errorf("Unsupported member assignment target %s", mem.NodeType())
	}
}

func (i *Interpreter) evaluateIteratorLiteral(expr *ast.IteratorLiteral, env *runtime.Environment) (runtime.Value, error) {
	iterEnv := runtime.NewEnvironment(env)
	instance := newGeneratorInstance(i, iterEnv, expr.Body)
	iterEnv.Define("gen", instance.controllerValue())
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

func integerBitWidth(t runtime.IntegerType) int {
	switch t {
	case runtime.IntegerI8, runtime.IntegerU8:
		return 8
	case runtime.IntegerI16, runtime.IntegerU16:
		return 16
	case runtime.IntegerI32, runtime.IntegerU32:
		return 32
	case runtime.IntegerI64, runtime.IntegerU64:
		return 64
	case runtime.IntegerI128, runtime.IntegerU128:
		return 128
	default:
		return 0
	}
}
