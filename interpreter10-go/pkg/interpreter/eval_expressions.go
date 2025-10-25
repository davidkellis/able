package interpreter

import (
	"fmt"
	"math/big"
	"strings"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func (i *Interpreter) evaluateExpression(node ast.Expression, env *runtime.Environment) (runtime.Value, error) {
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
	case *ast.IndexExpression:
		return i.evaluateIndexExpression(n, env)
	case *ast.UnaryExpression:
		return i.evaluateUnaryExpression(n, env)
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
	var injected []runtime.Value
	var funcValue *runtime.FunctionValue
	var callState any
	if env != nil {
		callState = env.RuntimeData()
	}
	switch fn := calleeVal.(type) {
	case runtime.NativeFunctionValue:
		args := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Impl(ctx, args)
	case *runtime.NativeFunctionValue:
		args := make([]runtime.Value, 0, len(call.Arguments))
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Impl(ctx, args)
	case runtime.NativeBoundMethodValue:
		args := make([]runtime.Value, 0, len(call.Arguments)+1)
		args = append(args, fn.Receiver)
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Method.Impl(ctx, args)
	case *runtime.NativeBoundMethodValue:
		args := make([]runtime.Value, 0, len(call.Arguments)+1)
		args = append(args, fn.Receiver)
		for _, argExpr := range call.Arguments {
			val, err := i.evaluateExpression(argExpr, env)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
		}
		ctx := &runtime.NativeCallContext{Env: env, State: callState}
		return fn.Method.Impl(ctx, args)
	case runtime.DynRefValue:
		resolved, resErr := i.resolveDynRef(fn)
		if resErr != nil {
			return nil, resErr
		}
		funcValue = resolved
	case *runtime.DynRefValue:
		if fn == nil {
			return nil, fmt.Errorf("dyn ref is nil")
		}
		resolved, resErr := i.resolveDynRef(*fn)
		if resErr != nil {
			return nil, resErr
		}
		funcValue = resolved
	case *runtime.BoundMethodValue:
		funcValue = fn.Method
		injected = append(injected, fn.Receiver)
	case runtime.BoundMethodValue:
		funcValue = fn.Method
		injected = append(injected, fn.Receiver)
	case *runtime.FunctionValue:
		funcValue = fn
	default:
		return nil, fmt.Errorf("calling non-function value of kind %s", calleeVal.Kind())
	}
	if funcValue == nil {
		return nil, fmt.Errorf("call target missing function value")
	}
	args := make([]runtime.Value, 0, len(injected)+len(call.Arguments))
	args = append(args, injected...)
	for _, argExpr := range call.Arguments {
		val, err := i.evaluateExpression(argExpr, env)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}
	return i.invokeFunction(funcValue, args, call)
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
		if len(args) != len(decl.Params) {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, len(decl.Params), len(args))
		}
		localEnv := runtime.NewEnvironment(fn.Closure)
		if call != nil {
			i.bindTypeArgumentsIfAny(decl, call, localEnv)
		}
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("function parameter %d is nil", idx)
			}
			if err := i.assignPattern(param.Name, args[idx], localEnv, true); err != nil {
				return nil, err
			}
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
		for idx, param := range decl.Params {
			if param == nil {
				return nil, fmt.Errorf("lambda parameter %d is nil", idx)
			}
			if err := i.assignPattern(param.Name, args[idx], localEnv, true); err != nil {
				return nil, err
			}
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
			switch member := lhs.Member.(type) {
			case *ast.Identifier:
				if inst.Fields == nil {
					return nil, fmt.Errorf("Expected named struct instance")
				}
				current, ok := inst.Fields[member.Name]
				if !ok {
					return nil, fmt.Errorf("No field named '%s'", member.Name)
				}
				if assign.Operator == ast.AssignmentAssign {
					inst.Fields[member.Name] = value
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				computed, err := applyBinaryOperator(binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				inst.Fields[member.Name] = computed
				return computed, nil
			case *ast.IntegerLiteral:
				if inst.Positional == nil {
					return nil, fmt.Errorf("Expected positional struct instance")
				}
				if member.Value == nil {
					return nil, fmt.Errorf("Struct field index out of bounds")
				}
				idx := int(member.Value.Int64())
				if idx < 0 || idx >= len(inst.Positional) {
					return nil, fmt.Errorf("Struct field index out of bounds")
				}
				if assign.Operator == ast.AssignmentAssign {
					inst.Positional[idx] = value
					return value, nil
				}
				if !isCompound {
					return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
				}
				current := inst.Positional[idx]
				computed, err := applyBinaryOperator(binaryOp, current, value)
				if err != nil {
					return nil, err
				}
				inst.Positional[idx] = computed
				return computed, nil
			default:
				return nil, fmt.Errorf("Unsupported member assignment target %s", member.NodeType())
			}
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
