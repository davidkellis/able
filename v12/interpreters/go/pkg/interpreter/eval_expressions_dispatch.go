package interpreter

import (
	"fmt"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) evaluateExpression(node ast.Expression, env *runtime.Environment) (result runtime.Value, err error) {
	state := i.stateFromEnv(env)
	defer func() {
		err = i.wrapStandardRuntimeError(err)
		err = i.attachRuntimeContext(err, node, state)
	}()
	if node == nil {
		return runtime.NilValue{}, nil
	}
	var (
		serialSync *SerialExecutor
	)
	if serial, ok := i.executor.(*SerialExecutor); ok {
		var payload *asyncContextPayload
		if env != nil {
			payload = payloadFromState(env.RuntimeData())
		}
		if payload == nil {
			serialSync = serial
			serialSync.beginSynchronousSection()
		}
	}
	defer func() {
		if serialSync != nil {
			serialSync.endSynchronousSection()
		}
	}()
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
		val := runtime.CloneBigInt(bigFromLiteral(n.Value))
		info, err := getIntegerInfo(suffix)
		if err != nil {
			return nil, err
		}
		if err := ensureFitsInteger(info, val); err != nil {
			return nil, err
		}
		return runtime.IntegerValue{Val: val, TypeSuffix: suffix}, nil
	case *ast.FloatLiteral:
		suffix := runtime.FloatF64
		if n.FloatType != nil {
			suffix = runtime.FloatType(*n.FloatType)
		}
		val := n.Value
		if suffix == runtime.FloatF32 {
			val = float64(float32(val))
		}
		return runtime.FloatValue{Val: val, TypeSuffix: suffix}, nil
	case *ast.ArrayLiteral:
		values := make([]runtime.Value, 0, len(n.Elements))
		for _, el := range n.Elements {
			val, err := i.evaluateExpression(el, env)
			if err != nil {
				return nil, err
			}
			values = append(values, val)
		}
		return i.newArrayValue(values, len(values)), nil
	case *ast.TypeCastExpression:
		value, err := i.evaluateExpression(n.Expression, env)
		if err != nil {
			return nil, err
		}
		return i.castValueToType(n.TargetType, value)
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
		return i.evaluateRangeValues(start, endExpr, n.Inclusive, env)
	case *ast.StructLiteral:
		return i.evaluateStructLiteral(n, env)
	case *ast.MapLiteral:
		return i.evaluateMapLiteral(n, env)
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
	case *ast.LoopExpression:
		return i.evaluateLoopExpression(n, env)
	case *ast.IteratorLiteral:
		return i.evaluateIteratorLiteral(n, env)
	case *ast.IfExpression:
		return i.evaluateIfExpression(n, env)
	case *ast.RescueExpression:
		return i.evaluateRescueExpression(n, env)
	case *ast.SpawnExpression:
		i.ensureConcurrencyBuiltins()
		task := i.makeAsyncTask(n.Expression, env)
		future := i.executor.RunFuture(task)
		return future, nil
	case *ast.AwaitExpression:
		return i.evaluateAwaitExpression(n, env)
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
	if i.isTruthy(cond) {
		return i.evaluateBlock(expr.IfBody, env)
	}
	for _, clause := range expr.ElseIfClauses {
		clauseCond, err := i.evaluateExpression(clause.Condition, env)
		if err != nil {
			return nil, err
		}
		if i.isTruthy(clauseCond) {
			return i.evaluateBlock(clause.Body, env)
		}
	}
	if expr.ElseBody != nil {
		return i.evaluateBlock(expr.ElseBody, env)
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
			if !i.isTruthy(guardVal) {
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
			if !i.isTruthy(guardVal) {
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
	if errVal, ok := asErrorValue(val); ok {
		return nil, raiseSignal{value: errVal}
	}
	if i.matchesType(ast.Ty("Error"), val) {
		return nil, raiseSignal{value: i.makeErrorValue(val, env)}
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
	failureKind := ""
	var failureValue runtime.Value
	if val == nil {
		failureKind = "nil"
	} else if val.Kind() == runtime.KindNil {
		failureKind = "nil"
	} else if errVal, ok := asErrorValue(val); ok {
		failureKind = "error"
		failureValue = errVal
	} else if i.matchesType(ast.Ty("Error"), val) {
		failureKind = "error"
		failureValue = val
	}
	if failureKind != "" {
		handlerEnv := runtime.NewEnvironment(env)
		if expr.ErrorBinding != nil && failureKind == "error" {
			handlerEnv.Define(expr.ErrorBinding.Name, failureValue)
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
	return i.applyUnaryOperator(string(expr.Operator), operand)
}

func (i *Interpreter) applyUnaryOperator(operator string, operand runtime.Value) (runtime.Value, error) {
	rawOperand := unwrapInterfaceValue(operand)
	switch operator {
	case "-":
		switch v := rawOperand.(type) {
		case runtime.IntegerValue:
			neg := new(big.Int).Neg(v.Val)
			return runtime.IntegerValue{Val: neg, TypeSuffix: v.TypeSuffix}, nil
		case runtime.FloatValue:
			return runtime.FloatValue{Val: -v.Val, TypeSuffix: v.TypeSuffix}, nil
		default:
			if result, ok, err := i.applyUnaryInterface(operator, operand); ok {
				return result, err
			}
			return nil, fmt.Errorf("unary '-' not supported for %T", operand)
		}
	case "!":
		return runtime.BoolValue{Val: !i.isTruthy(operand)}, nil
	case "~", ".~":
		switch v := rawOperand.(type) {
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
			if result, ok, err := i.applyUnaryInterface(operator, operand); ok {
				return result, err
			}
			return nil, fmt.Errorf("unary '%s' not supported for %T", operator, operand)
		}
	default:
		return nil, fmt.Errorf("unsupported unary operator %s", operator)
	}
}

func (i *Interpreter) evaluateRangeValues(start runtime.Value, end runtime.Value, inclusive bool, env *runtime.Environment) (runtime.Value, error) {
	if result, err := i.tryInvokeRangeImplementation(start, end, inclusive, env); err != nil {
		return nil, err
	} else if result != nil {
		return result, nil
	}
	if !isIntegerValue(start) || !isIntegerValue(end) {
		return nil, fmt.Errorf("Range boundaries must be numeric")
	}
	startVal, err := rangeEndpoint(start)
	if err != nil {
		return nil, err
	}
	endVal, err := rangeEndpoint(end)
	if err != nil {
		return nil, err
	}
	step := 1
	if startVal > endVal {
		step = -1
	}
	elements := make([]runtime.Value, 0)
	for current := startVal; ; current += step {
		if step > 0 {
			if inclusive {
				if current > endVal {
					break
				}
			} else if current >= endVal {
				break
			}
		} else {
			if inclusive {
				if current < endVal {
					break
				}
			} else if current <= endVal {
				break
			}
		}
		elements = append(elements, runtime.IntegerValue{Val: big.NewInt(int64(current)), TypeSuffix: runtime.IntegerI32})
	}
	return &runtime.ArrayValue{Elements: elements}, nil
}

func (i *Interpreter) evaluateBinaryExpression(expr *ast.BinaryExpression, env *runtime.Environment) (runtime.Value, error) {
	leftVal, err := i.evaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}
	if expr.Operator == "|>" || expr.Operator == "|>>" {
		return i.evaluatePipeExpression(leftVal, expr.Right, env)
	}
	switch expr.Operator {
	case "&&":
		if !i.isTruthy(leftVal) {
			return leftVal, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		return rightVal, nil
	case "||":
		if i.isTruthy(leftVal) {
			return leftVal, nil
		}
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		return rightVal, nil
	default:
		rightVal, err := i.evaluateExpression(expr.Right, env)
		if err != nil {
			return nil, err
		}
		if expr.Operator == "+" {
			rawLeft := unwrapInterfaceValue(leftVal)
			rawRight := unwrapInterfaceValue(rightVal)
			if ls, ok := rawLeft.(runtime.StringValue); ok {
				rs, ok := rawRight.(runtime.StringValue)
				if !ok {
					return nil, fmt.Errorf("Arithmetic requires numeric operands")
				}
				return runtime.StringValue{Val: ls.Val + rs.Val}, nil
			}
			if _, ok := rawRight.(runtime.StringValue); ok {
				return nil, fmt.Errorf("Arithmetic requires numeric operands")
			}
		}
		return applyBinaryOperator(i, expr.Operator, leftVal, rightVal)
	}
}

func (p *placeholderAnalyzer) visitExpression(expr ast.Expression) error {
	if expr == nil {
		return nil
	}
	switch e := expr.(type) {
	case *ast.PlaceholderExpression:
		p.hasPlaceholder = true
		idx := 1
		if e.Index != nil {
			idx = *e.Index
		}
		if idx <= 0 {
			return fmt.Errorf("Placeholder index must be positive, found @%d", idx)
		}
		if idx > p.highestExplicit {
			p.highestExplicit = idx
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
		for _, clause := range e.ElseIfClauses {
			if clause == nil {
				continue
			}
			if err := p.visitExpression(clause.Condition); err != nil {
				return err
			}
			if err := p.visitExpression(clause.Body); err != nil {
				return err
			}
		}
		if e.ElseBody != nil {
			if err := p.visitExpression(e.ElseBody); err != nil {
				return err
			}
		}
		return nil
	case *ast.IteratorLiteral:
		return nil
	case *ast.LambdaExpression:
		return nil
	case *ast.SpawnExpression, *ast.AwaitExpression:
		return nil
	case *ast.Identifier,
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

func (p *placeholderClosure) invoke(args []runtime.Value) (runtime.Value, error) {
	if len(args) != p.plan.paramCount {
		return nil, fmt.Errorf("Placeholder lambda expects %d arguments, got %d", p.plan.paramCount, len(args))
	}
	callEnv := runtime.NewEnvironment(p.env)
	state := p.interpreter.stateFromEnv(callEnv)
	state.pushPlaceholderFrame(p.plan.paramCount, args)
	defer state.popPlaceholderFrame()
	var result runtime.Value
	var err error
	if p.bytecode != nil {
		vm := newBytecodeVM(p.interpreter, callEnv)
		result, err = vm.run(p.bytecode)
	} else {
		result, err = p.interpreter.evaluateExpression(p.expression, callEnv)
	}
	if err != nil {
		return nil, err
	}
	if result == nil {
		return runtime.NilValue{}, nil
	}
	return result, nil
}
