package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

// Interpreter drives evaluation of Able v10 AST nodes.
type Interpreter struct {
	global          *runtime.Environment
	inherentMethods map[string]map[string]*runtime.FunctionValue
	interfaces      map[string]*runtime.InterfaceDefinitionValue
	implMethods     map[string][]implEntry
}

type implEntry struct {
	interfaceName string
	methods       map[string]*runtime.FunctionValue
	definition    *ast.ImplementationDefinition
	targetCount   int
	matchScore    int
}

func collectImplementationTargetNames(target ast.TypeExpression) ([]string, int, error) {
	var names []string
	score, err := gatherImplementationTargets(target, &names)
	if err != nil {
		return nil, 0, err
	}
	if score == 0 {
		score = len(names)
	}
	return names, score, nil
}

func gatherImplementationTargets(target ast.TypeExpression, out *[]string) (int, error) {
	switch t := target.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name == nil {
			return 0, fmt.Errorf("Implementation target requires identifier")
		}
		*out = append(*out, t.Name.Name)
		return 1, nil
	case *ast.UnionTypeExpression:
		score := 0
		for _, member := range t.Members {
			memberScore, err := gatherImplementationTargets(member, out)
			if err != nil {
				return 0, err
			}
			score += memberScore
		}
		return score, nil
	default:
		return 0, fmt.Errorf("Implementation target type %T is not supported", target)
	}
}

// New returns an interpreter with an empty global environment.
func New() *Interpreter {
	return &Interpreter{
		global:          runtime.NewEnvironment(nil),
		inherentMethods: make(map[string]map[string]*runtime.FunctionValue),
		interfaces:      make(map[string]*runtime.InterfaceDefinitionValue),
		implMethods:     make(map[string][]implEntry),
	}
}

// GlobalEnvironment returns the interpreter’s global environment.
func (i *Interpreter) GlobalEnvironment() *runtime.Environment {
	return i.global
}

// EvaluateModule executes a module node and returns the last evaluated value and environment.
func (i *Interpreter) EvaluateModule(module *ast.Module) (runtime.Value, *runtime.Environment, error) {
	moduleEnv := i.global
	var last runtime.Value = runtime.NilValue{}
	for _, stmt := range module.Body {
		val, err := i.evaluateStatement(stmt, moduleEnv)
		if err != nil {
			if rs, ok := err.(raiseSignal); ok {
				return nil, moduleEnv, rs
			}
			if _, ok := err.(returnSignal); ok {
				return nil, nil, fmt.Errorf("return outside function")
			}
			return nil, nil, err
		}
		last = val
	}
	return last, moduleEnv, nil
}

// evaluateStatement currently delegates to expression evaluation for expressions.
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
	case *ast.WhileLoop:
		return i.evaluateWhileLoop(n, env)
	case *ast.ForLoop:
		return i.evaluateForLoop(n, env)
	case *ast.RaiseStatement:
		return i.evaluateRaiseStatement(n, env)
	case *ast.BreakStatement:
		return i.evaluateBreakStatement(n, env)
	case *ast.ReturnStatement:
		return i.evaluateReturnStatement(n, env)
	default:
		return nil, fmt.Errorf("unsupported statement type: %s", n.NodeType())
	}
}

// evaluateExpression handles literals, identifiers, and blocks.
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
	case *ast.RangeExpression:
		start, err := i.evaluateExpression(n.Start, env)
		if err != nil {
			return nil, err
		}
		endExpr, err := i.evaluateExpression(n.End, env)
		if err != nil {
			return nil, err
		}
		return &runtime.RangeValue{Start: start, End: endExpr, Inclusive: n.Inclusive}, nil
	case *ast.StructLiteral:
		return i.evaluateStructLiteral(n, env)
	case *ast.MemberAccessExpression:
		return i.evaluateMemberAccess(n, env)
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
	case *ast.IfExpression:
		return i.evaluateIfExpression(n, env)
	case *ast.RescueExpression:
		return i.evaluateRescueExpression(n, env)
	default:
		return nil, fmt.Errorf("unsupported expression type: %s", n.NodeType())
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
					return nil, fmt.Errorf("labeled break not supported")
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("labeled continue not supported")
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
					return nil, fmt.Errorf("labeled break not supported")
				}
				return sig.value, nil
			case continueSignal:
				if sig.label != "" {
					return nil, fmt.Errorf("labeled continue not supported")
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
	}
	return nil, breakSignal{label: label, value: val}
}

func (i *Interpreter) evaluateAssignment(assign *ast.AssignmentExpression, env *runtime.Environment) (runtime.Value, error) {
	value, err := i.evaluateExpression(assign.Right, env)
	if err != nil {
		return nil, err
	}
	isCompound := assign.Operator != ast.AssignmentDeclare && assign.Operator != ast.AssignmentAssign

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
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
		}
	case *ast.MemberAccessExpression:
		if assign.Operator == ast.AssignmentDeclare {
			return nil, fmt.Errorf("Cannot use := on member access")
		}
		if assign.Operator != ast.AssignmentAssign {
			return nil, fmt.Errorf("unsupported assignment operator %s", assign.Operator)
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
				if _, ok := inst.Fields[member.Name]; !ok {
					return nil, fmt.Errorf("No field named '%s'", member.Name)
				}
				inst.Fields[member.Name] = value
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
				inst.Positional[idx] = value
			default:
				return nil, fmt.Errorf("Unsupported member assignment target %s", member.NodeType())
			}
		default:
			return nil, fmt.Errorf("Member assignment requires struct instance")
		}
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
		clauseEnv := runtime.NewEnvironment(env)
		if !rescueMatches(clause.Pattern, rs.value, clauseEnv) {
			continue
		}
		if clause.Guard != nil {
			guardVal, gErr := i.evaluateExpression(clause.Guard, clauseEnv)
			if gErr != nil {
				return nil, gErr
			}
			if !isTruthy(guardVal) {
				continue
			}
		}
		return i.evaluateExpression(clause.Body, clauseEnv)
	}
	return nil, rs
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
	default:
		return nil, fmt.Errorf("unsupported unary operator %s", expr.Operator)
	}
}

func (i *Interpreter) evaluateBinaryExpression(expr *ast.BinaryExpression, env *runtime.Environment) (runtime.Value, error) {
	leftVal, err := i.evaluateExpression(expr.Left, env)
	if err != nil {
		return nil, err
	}
	rightVal, err := i.evaluateExpression(expr.Right, env)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "+", "-", "*", "/":
		return evaluateArithmetic(expr.Operator, leftVal, rightVal)
	case "<", "<=", ">", ">=":
		return evaluateComparison(expr.Operator, leftVal, rightVal)
	case "==", "!=":
		eg := valuesEqual(leftVal, rightVal)
		if expr.Operator == "!=" {
			eg = !eg
		}
		return runtime.BoolValue{Val: eg}, nil
	case "&&":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("left operand of && must be bool")
		}
		if !lb.Val {
			return runtime.BoolValue{Val: false}, nil
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("right operand of && must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	case "||":
		lb, ok := leftVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("left operand of || must be bool")
		}
		if lb.Val {
			return runtime.BoolValue{Val: true}, nil
		}
		rb, ok := rightVal.(runtime.BoolValue)
		if !ok {
			return nil, fmt.Errorf("right operand of || must be bool")
		}
		return runtime.BoolValue{Val: rb.Val}, nil
	default:
		return nil, fmt.Errorf("unsupported binary operator %s", expr.Operator)
	}
}

func (i *Interpreter) evaluateFunctionCall(call *ast.FunctionCall, env *runtime.Environment) (runtime.Value, error) {
	calleeVal, err := i.evaluateExpression(call.Callee, env)
	if err != nil {
		return nil, err
	}
	var injected []runtime.Value
	var funcValue *runtime.FunctionValue
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
		ctx := &runtime.NativeCallContext{Env: env}
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
		ctx := &runtime.NativeCallContext{Env: env}
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
		ctx := &runtime.NativeCallContext{Env: env}
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
		ctx := &runtime.NativeCallContext{Env: env}
		return fn.Method.Impl(ctx, args)
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
	return i.invokeFunction(funcValue, args)
}

func (i *Interpreter) invokeFunction(fn *runtime.FunctionValue, args []runtime.Value) (runtime.Value, error) {
	switch decl := fn.Declaration.(type) {
	case *ast.FunctionDefinition:
		if decl.Body == nil {
			return runtime.NilValue{}, nil
		}
		if len(args) != len(decl.Params) {
			name := "<anonymous>"
			if decl.ID != nil {
				name = decl.ID.Name
			}
			return nil, fmt.Errorf("Function '%s' expects %d arguments, got %d", name, len(decl.Params), len(args))
		}
		localEnv := runtime.NewEnvironment(fn.Closure)
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
	default:
		return nil, fmt.Errorf("calling unsupported function declaration %T", fn.Declaration)
	}
}

// bigFromLiteral normalizes numeric literals (number or bigint) to *big.Int.
func bigFromLiteral(val interface{}) *big.Int {
	switch v := val.(type) {
	case int:
		return big.NewInt(int64(v))
	case int64:
		return big.NewInt(v)
	case float64:
		return big.NewInt(int64(v))
	case string:
		if bi, ok := new(big.Int).SetString(v, 10); ok {
			return bi
		}
		return big.NewInt(0)
	case *big.Int:
		return runtime.CloneBigInt(v)
	default:
		return big.NewInt(0)
	}
}

func evaluateArithmetic(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	switch lv := left.(type) {
	case runtime.IntegerValue:
		rv, ok := right.(runtime.IntegerValue)
		if !ok {
			return nil, fmt.Errorf("mixed numeric types not supported")
		}
		result := new(big.Int)
		switch op {
		case "+":
			result.Add(lv.Val, rv.Val)
		case "-":
			result.Sub(lv.Val, rv.Val)
		case "*":
			result.Mul(lv.Val, rv.Val)
		case "/":
			if rv.Val.Sign() == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			result.Quo(lv.Val, rv.Val)
		default:
			return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
		}
		return runtime.IntegerValue{Val: result, TypeSuffix: lv.TypeSuffix}, nil
	case runtime.FloatValue:
		rv, ok := right.(runtime.FloatValue)
		if !ok {
			return nil, fmt.Errorf("mixed numeric types not supported")
		}
		var val float64
		switch op {
		case "+":
			val = lv.Val + rv.Val
		case "-":
			val = lv.Val - rv.Val
		case "*":
			val = lv.Val * rv.Val
		case "/":
			if rv.Val == 0 {
				return nil, fmt.Errorf("division by zero")
			}
			val = lv.Val / rv.Val
		default:
			return nil, fmt.Errorf("unsupported arithmetic operator %s", op)
		}
		return runtime.FloatValue{Val: val, TypeSuffix: lv.TypeSuffix}, nil
	case runtime.StringValue:
		if op == "+" {
			rStr, ok := right.(runtime.StringValue)
			if !ok {
				return nil, fmt.Errorf("string concatenation requires both operands to be strings")
			}
			return runtime.StringValue{Val: lv.Val + rStr.Val}, nil
		}
		return nil, fmt.Errorf("operator %s not supported for strings", op)
	default:
		return nil, fmt.Errorf("unsupported operand types for %s", op)
	}
}

func valuesEqual(left runtime.Value, right runtime.Value) bool {
	switch lv := left.(type) {
	case runtime.StringValue:
		if rv, ok := right.(runtime.StringValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.BoolValue:
		if rv, ok := right.(runtime.BoolValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.CharValue:
		if rv, ok := right.(runtime.CharValue); ok {
			return lv.Val == rv.Val
		}
	case runtime.NilValue:
		_, ok := right.(runtime.NilValue)
		return ok
	case runtime.IntegerValue:
		if rv, ok := right.(runtime.IntegerValue); ok {
			return lv.Val.Cmp(rv.Val) == 0
		}
	case runtime.FloatValue:
		if rv, ok := right.(runtime.FloatValue); ok {
			return lv.Val == rv.Val
		}
	}
	return false
}

func evaluateComparison(op string, left runtime.Value, right runtime.Value) (runtime.Value, error) {
	switch lv := left.(type) {
	case runtime.IntegerValue:
		rv, ok := right.(runtime.IntegerValue)
		if !ok {
			return nil, fmt.Errorf("mixed numeric types not supported in comparison")
		}
		cmp := lv.Val.Cmp(rv.Val)
		return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
	case runtime.FloatValue:
		rv, ok := right.(runtime.FloatValue)
		if !ok {
			return nil, fmt.Errorf("mixed numeric types not supported in comparison")
		}
		var cmp int
		if lv.Val < rv.Val {
			cmp = -1
		} else if lv.Val > rv.Val {
			cmp = 1
		} else {
			cmp = 0
		}
		return runtime.BoolValue{Val: comparisonOp(op, cmp)}, nil
	default:
		return nil, fmt.Errorf("unsupported operands for comparison %s", op)
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

type breakSignal struct {
	label string
	value runtime.Value
}

func (b breakSignal) Error() string {
	if b.label != "" {
		return fmt.Sprintf("break %s", b.label)
	}
	return "break"
}

type continueSignal struct {
	label string
}

func (c continueSignal) Error() string {
	if c.label != "" {
		return fmt.Sprintf("continue %s", c.label)
	}
	return "continue"
}

type raiseSignal struct {
	value runtime.Value
}

func (r raiseSignal) Error() string {
	if errVal, ok := r.value.(runtime.ErrorValue); ok {
		return errVal.Message
	}
	return valueToString(r.value)
}

type returnSignal struct {
	value runtime.Value
}

func (r returnSignal) Error() string {
	return "return"
}

func makeErrorValue(val runtime.Value) runtime.ErrorValue {
	if errVal, ok := val.(runtime.ErrorValue); ok {
		return errVal
	}
	message := valueToString(val)
	return runtime.ErrorValue{Message: message}
}

func valueToString(val runtime.Value) string {
	switch v := val.(type) {
	case runtime.StringValue:
		return v.Val
	case runtime.BoolValue:
		if v.Val {
			return "true"
		}
		return "false"
	case runtime.IntegerValue:
		return v.Val.String()
	case runtime.FloatValue:
		return fmt.Sprintf("%g", v.Val)
	case runtime.NilValue:
		return "nil"
	default:
		return fmt.Sprintf("[%s]", v.Kind())
	}
}

func (i *Interpreter) evaluateStructDefinition(def *ast.StructDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Struct definition requires identifier")
	}
	structVal := &runtime.StructDefinitionValue{Node: def}
	env.Define(def.ID.Name, structVal)
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateInterfaceDefinition(def *ast.InterfaceDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.ID == nil {
		return nil, fmt.Errorf("Interface definition requires identifier")
	}
	ifaceVal := &runtime.InterfaceDefinitionValue{Node: def, Env: env}
	env.Define(def.ID.Name, ifaceVal)
	i.interfaces[def.ID.Name] = ifaceVal
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateImplementationDefinition(def *ast.ImplementationDefinition, env *runtime.Environment) (runtime.Value, error) {
	if def.InterfaceName == nil {
		return nil, fmt.Errorf("Implementation requires interface name")
	}
	ifaceName := def.InterfaceName.Name
	ifaceDef, ok := i.interfaces[ifaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", ifaceName)
	}
	targetNames, matchScore, err := collectImplementationTargetNames(def.TargetType)
	if err != nil {
		return nil, err
	}
	if len(targetNames) == 0 {
		return nil, fmt.Errorf("Implementation target must reference at least one concrete type")
	}
	methods := make(map[string]*runtime.FunctionValue)
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Implementation method requires identifier")
		}
		methods[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
	}
	if ifaceDef.Node != nil {
		for _, sig := range ifaceDef.Node.Signatures {
			if sig == nil || sig.Name == nil {
				continue
			}
			name := sig.Name.Name
			if _, ok := methods[name]; ok {
				continue
			}
			if sig.DefaultImpl == nil {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			methods[name] = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
		}
	}
	for _, targetName := range targetNames {
		entry := implEntry{
			interfaceName: ifaceName,
			methods:       methods,
			definition:    def,
			targetCount:   len(targetNames),
			matchScore:    matchScore,
		}
		i.implMethods[targetName] = append(i.implMethods[targetName], entry)
	}
	// Expose named impls as values (future work). For now, only register behaviourally.
	_ = ifaceDef
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateMethodsDefinition(def *ast.MethodsDefinition, env *runtime.Environment) (runtime.Value, error) {
	simpleType, ok := def.TargetType.(*ast.SimpleTypeExpression)
	if !ok || simpleType.Name == nil {
		return nil, fmt.Errorf("MethodsDefinition requires simple target type")
	}
	typeName := simpleType.Name.Name
	bucket, ok := i.inherentMethods[typeName]
	if !ok {
		bucket = make(map[string]*runtime.FunctionValue)
		i.inherentMethods[typeName] = bucket
	}
	for _, fn := range def.Definitions {
		if fn == nil || fn.ID == nil {
			return nil, fmt.Errorf("Method definition requires identifier")
		}
		bucket[fn.ID.Name] = &runtime.FunctionValue{Declaration: fn, Closure: env}
	}
	return runtime.NilValue{}, nil
}

func (i *Interpreter) evaluateStructLiteral(lit *ast.StructLiteral, env *runtime.Environment) (runtime.Value, error) {
	if lit.StructType == nil {
		return nil, fmt.Errorf("Struct literal requires explicit struct type in this milestone")
	}
	structName := lit.StructType.Name
	defValue, err := env.Get(structName)
	if err != nil {
		return nil, err
	}
	structDefVal, err := toStructDefinitionValue(defValue, structName)
	if err != nil {
		return nil, err
	}
	structDef := structDefVal.Node
	if structDef == nil {
		return nil, fmt.Errorf("struct definition '%s' unavailable", structName)
	}
	if len(lit.TypeArguments) > 0 {
		if len(structDef.GenericParams) == 0 {
			return nil, fmt.Errorf("Type '%s' does not accept type arguments", structName)
		}
		return nil, fmt.Errorf("Struct generics are not implemented in this interpreter yet")
	}
	if lit.IsPositional {
		if structDef.Kind != ast.StructKindPositional && structDef.Kind != ast.StructKindSingleton {
			return nil, fmt.Errorf("Positional struct literal not allowed for struct '%s'", structName)
		}
		if len(lit.Fields) != len(structDef.Fields) {
			return nil, fmt.Errorf("Struct '%s' expects %d fields, got %d", structName, len(structDef.Fields), len(lit.Fields))
		}
		values := make([]runtime.Value, len(lit.Fields))
		for idx, field := range lit.Fields {
			val, err := i.evaluateExpression(field.Value, env)
			if err != nil {
				return nil, err
			}
			values[idx] = val
		}
		return &runtime.StructInstanceValue{Definition: structDefVal, Positional: values}, nil
	}
	if structDef.Kind == ast.StructKindPositional {
		return nil, fmt.Errorf("Named struct literal not allowed for positional struct '%s'", structName)
	}
	if lit.FunctionalUpdateSource != nil && structDef.Kind == ast.StructKindPositional {
		return nil, fmt.Errorf("Functional update only supported for named structs")
	}
	fields := make(map[string]runtime.Value)
	if lit.FunctionalUpdateSource != nil {
		base, err := i.evaluateExpression(lit.FunctionalUpdateSource, env)
		if err != nil {
			return nil, err
		}
		baseStruct, ok := base.(*runtime.StructInstanceValue)
		if !ok {
			return nil, fmt.Errorf("Functional update source must be a struct instance")
		}
		if baseStruct.Definition == nil || baseStruct.Definition.Node == nil || baseStruct.Definition.Node.ID == nil || baseStruct.Definition.Node.ID.Name != structName {
			return nil, fmt.Errorf("Functional update source must be same struct type")
		}
		if baseStruct.Fields == nil {
			return nil, fmt.Errorf("Functional update only supported for named structs")
		}
		for k, v := range baseStruct.Fields {
			fields[k] = v
		}
	}
	for _, f := range lit.Fields {
		name := ""
		if f.Name != nil {
			name = f.Name.Name
		} else if f.IsShorthand {
			if ident, ok := f.Value.(*ast.Identifier); ok {
				name = ident.Name
			}
		}
		if name == "" {
			return nil, fmt.Errorf("Named struct field initializer must have a field name")
		}
		val, err := i.evaluateExpression(f.Value, env)
		if err != nil {
			return nil, err
		}
		fields[name] = val
	}
	if structDef.Kind == ast.StructKindNamed {
		required := make(map[string]struct{}, len(structDef.Fields))
		for _, defField := range structDef.Fields {
			if defField.Name != nil {
				required[defField.Name.Name] = struct{}{}
			}
		}
		for k := range fields {
			delete(required, k)
		}
		if len(required) > 0 {
			for missing := range required {
				return nil, fmt.Errorf("Missing field '%s' for struct '%s'", missing, structName)
			}
		}
	}
	return &runtime.StructInstanceValue{Definition: structDefVal, Fields: fields}, nil
}

func (i *Interpreter) evaluateMemberAccess(expr *ast.MemberAccessExpression, env *runtime.Environment) (runtime.Value, error) {
	obj, err := i.evaluateExpression(expr.Object, env)
	if err != nil {
		return nil, err
	}
	switch v := obj.(type) {
	case *runtime.StructDefinitionValue:
		return i.structDefinitionMember(v, expr.Member)
	case runtime.StructDefinitionValue:
		return i.structDefinitionMember(&v, expr.Member)
	case *runtime.StructInstanceValue:
		return i.structInstanceMember(v, expr.Member)
	case *runtime.InterfaceValue:
		return i.interfaceMember(v, expr.Member)
	default:
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
}

func (i *Interpreter) structInstanceMember(inst *runtime.StructInstanceValue, member ast.Expression) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
	}
	if ident, ok := member.(*ast.Identifier); ok {
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		if val, ok := inst.Fields[ident.Name]; ok {
			return val, nil
		}
		if inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil {
			return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
		}
		typeName := inst.Definition.Node.ID.Name
		if bucket, ok := i.inherentMethods[typeName]; ok {
			if method, ok := bucket[ident.Name]; ok {
				if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
					return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
				}
				return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
			}
		}
		method, err := i.selectStructMethod(typeName, ident.Name)
		if err != nil {
			return nil, err
		}
		if method != nil {
			return &runtime.BoundMethodValue{Receiver: inst, Method: method}, nil
		}
		return nil, fmt.Errorf("No field or method named '%s'", ident.Name)
	}
	if intLit, ok := member.(*ast.IntegerLiteral); ok {
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if intLit.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(intLit.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		return inst.Positional[idx], nil
	}
	return nil, fmt.Errorf("Member access only supported on structs/arrays in this milestone")
}

func (i *Interpreter) structDefinitionMember(def *runtime.StructDefinitionValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Static access expects identifier member")
	}
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing identifier")
	}
	typeName := def.Node.ID.Name
	bucket := i.inherentMethods[typeName]
	if bucket == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	method, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("No static method '%s' for %s", ident.Name, typeName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, typeName)
	}
	return method, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := ""
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		ifaceName = val.Interface.Node.ID.Name
	}
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	var method *runtime.FunctionValue
	if val.Methods != nil {
		method = val.Methods[ident.Name]
	}
	if method == nil {
		// Try to resolve lazily in case Methods map not populated
		if structName, ok := i.getTypeNameForValue(val.Underlying); ok {
			if entry, ok := i.lookupImplEntry(structName, ifaceName); ok {
				method = entry.methods[ident.Name]
			}
		}
	}
	if method == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if fnDef, ok := method.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, ifaceName)
	}
	return &runtime.BoundMethodValue{Receiver: val.Underlying, Method: method}, nil
}

func toStructDefinitionValue(val runtime.Value, name string) (*runtime.StructDefinitionValue, error) {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		return v, nil
	case runtime.StructDefinitionValue:
		return &v, nil
	default:
		return nil, fmt.Errorf("'%s' is not a struct type", name)
	}
}

func (i *Interpreter) assignPattern(pattern ast.Pattern, value runtime.Value, env *runtime.Environment, isDeclaration bool) error {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return declareOrAssign(env, p.Name, value, isDeclaration)
	case *ast.WildcardPattern:
		return nil
	case *ast.LiteralPattern:
		litExpr, ok := p.Literal.(ast.Expression)
		if !ok {
			return fmt.Errorf("invalid literal in pattern: %T", p.Literal)
		}
		litVal, err := i.evaluateExpression(litExpr, env)
		if err != nil {
			return err
		}
		if !valuesEqual(litVal, value) {
			return fmt.Errorf("pattern literal mismatch")
		}
		return nil
	case *ast.StructPattern:
		structVal, ok := value.(*runtime.StructInstanceValue)
		if !ok {
			return fmt.Errorf("cannot destructure non-struct value")
		}
		if p.StructType != nil {
			def := structVal.Definition
			if def == nil || def.Node == nil || def.Node.ID == nil || def.Node.ID.Name != p.StructType.Name {
				return fmt.Errorf("struct type mismatch in destructuring")
			}
		}
		if p.IsPositional {
			if structVal.Positional == nil {
				return fmt.Errorf("expected positional struct value")
			}
			if len(p.Fields) != len(structVal.Positional) {
				return fmt.Errorf("struct field count mismatch in destructuring")
			}
			for idx, field := range p.Fields {
				if field == nil {
					return fmt.Errorf("invalid positional struct pattern at index %d", idx)
				}
				fieldVal := structVal.Positional[idx]
				if fieldVal == nil {
					return fmt.Errorf("missing positional struct value at index %d", idx)
				}
				if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
					return err
				}
				if field.Binding != nil {
					if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
						return err
					}
				}
			}
			return nil
		}
		if structVal.Fields == nil {
			return fmt.Errorf("expected named struct value")
		}
		for _, field := range p.Fields {
			if field.FieldName == nil {
				return fmt.Errorf("named struct pattern missing field name")
			}
			fieldVal, ok := structVal.Fields[field.FieldName.Name]
			if !ok {
				return fmt.Errorf("missing field '%s' during destructuring", field.FieldName.Name)
			}
			if err := i.assignPattern(field.Pattern, fieldVal, env, isDeclaration); err != nil {
				return err
			}
			if field.Binding != nil {
				if err := declareOrAssign(env, field.Binding.Name, fieldVal, isDeclaration); err != nil {
					return err
				}
			}
		}
		return nil
	case *ast.ArrayPattern:
		var elements []runtime.Value
		switch arr := value.(type) {
		case *runtime.ArrayValue:
			elements = arr.Elements
		default:
			return fmt.Errorf("cannot destructure non-array value")
		}
		if len(elements) < len(p.Elements) {
			return fmt.Errorf("array too short for destructuring")
		}
		if p.RestPattern == nil && len(elements) != len(p.Elements) {
			return fmt.Errorf("array length mismatch in destructuring")
		}
		for idx, elemPattern := range p.Elements {
			if elemPattern == nil {
				return fmt.Errorf("invalid array pattern at index %d", idx)
			}
			elemVal := elements[idx]
			if err := i.assignPattern(elemPattern, elemVal, env, isDeclaration); err != nil {
				return err
			}
		}
		if p.RestPattern != nil {
			switch rest := p.RestPattern.(type) {
			case *ast.Identifier:
				restElems := append([]runtime.Value(nil), elements[len(p.Elements):]...)
				restVal := &runtime.ArrayValue{Elements: restElems}
				if err := declareOrAssign(env, rest.Name, restVal, isDeclaration); err != nil {
					return err
				}
			case *ast.WildcardPattern:
				// ignore remaining elements
			default:
				return fmt.Errorf("unsupported rest pattern type %s", rest.NodeType())
			}
		} else if len(elements) != len(p.Elements) {
			return fmt.Errorf("array length mismatch in destructuring")
		}
		return nil
	case *ast.TypedPattern:
		if !i.matchesType(p.TypeAnnotation, value) {
			return fmt.Errorf("Typed pattern mismatch in assignment")
		}
		coerced, err := i.coerceValueToType(p.TypeAnnotation, value)
		if err != nil {
			return err
		}
		return i.assignPattern(p.Pattern, coerced, env, isDeclaration)
	default:
		return fmt.Errorf("unsupported pattern %s", pattern.NodeType())
	}
}

func declareOrAssign(env *runtime.Environment, name string, value runtime.Value, isDeclaration bool) error {
	if isDeclaration {
		env.Define(name, value)
		return nil
	}
	return env.Assign(name, value)
}

func (i *Interpreter) getTypeNameForValue(value runtime.Value) (string, bool) {
	switch v := value.(type) {
	case *runtime.StructInstanceValue:
		if v.Definition != nil && v.Definition.Node != nil && v.Definition.Node.ID != nil {
			return v.Definition.Node.ID.Name, true
		}
	case *runtime.InterfaceValue:
		return i.getTypeNameForValue(v.Underlying)
	}
	return "", false
}

func (i *Interpreter) lookupImplEntry(structName, interfaceName string) (*implEntry, bool) {
	entries := i.implMethods[structName]
	bestIdx := -1
	for idx := range entries {
		entry := entries[idx]
		if entry.interfaceName != interfaceName {
			continue
		}
		score := entry.matchScore
		if score == 0 {
			score = entry.targetCount
		}
		if bestIdx == -1 {
			bestIdx = idx
			entries[bestIdx].matchScore = score
			continue
		}
		currentScore := entries[bestIdx].matchScore
		if currentScore == 0 {
			currentScore = entries[bestIdx].targetCount
		}
		if score < currentScore {
			bestIdx = idx
			entries[bestIdx].matchScore = score
		} else if score == currentScore {
			// ambiguous; prefer to signal absence so callers can raise an error later
			return nil, false
		}
	}
	if bestIdx == -1 {
		return nil, false
	}
	i.implMethods[structName] = entries
	return &entries[bestIdx], true
}

func (i *Interpreter) interfaceMatches(val *runtime.InterfaceValue, interfaceName string) bool {
	if val == nil {
		return false
	}
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		if val.Interface.Node.ID.Name == interfaceName {
			return true
		}
	}
	if structName, ok := i.getTypeNameForValue(val.Underlying); ok {
		_, ok := i.lookupImplEntry(structName, interfaceName)
		return ok
	}
	return false
}

func (i *Interpreter) selectStructMethod(typeName, methodName string) (*runtime.FunctionValue, error) {
	entries := i.implMethods[typeName]
	bestScore := int(^uint(0) >> 1)
	var selected *runtime.FunctionValue
	ambiguous := false
	for idx := range entries {
		entry := &entries[idx]
		method := entry.methods[methodName]
		if method == nil {
			if ifaceDef, ok := i.interfaces[entry.interfaceName]; ok && ifaceDef.Node != nil {
				for _, sig := range ifaceDef.Node.Signatures {
					if sig == nil || sig.Name == nil || sig.Name.Name != methodName || sig.DefaultImpl == nil {
						continue
					}
					defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
					method = &runtime.FunctionValue{Declaration: defaultDef, Closure: ifaceDef.Env}
					if entry.methods == nil {
						entry.methods = make(map[string]*runtime.FunctionValue)
					}
					entry.methods[methodName] = method
					break
				}
			}
		}
		if method == nil {
			continue
		}
		score := entry.matchScore
		if score == 0 {
			score = entry.targetCount
		}
		if score < bestScore {
			bestScore = score
			selected = method
			ambiguous = false
		} else if score == bestScore {
			ambiguous = true
		}
	}
	i.implMethods[typeName] = entries
	if selected == nil {
		return nil, nil
	}
	if ambiguous {
		return nil, fmt.Errorf("Ambiguous method '%s' for type '%s'", methodName, typeName)
	}
	if fnDef, ok := selected.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
		return nil, fmt.Errorf("Method '%s' on %s is private", methodName, typeName)
	}
	return selected, nil
}

func (i *Interpreter) matchesType(typeExpr ast.TypeExpression, value runtime.Value) bool {
	switch t := typeExpr.(type) {
	case *ast.WildcardTypeExpression:
		return true
	case *ast.SimpleTypeExpression:
		name := t.Name.Name
		switch name {
		case "string":
			_, ok := value.(runtime.StringValue)
			return ok
		case "bool":
			_, ok := value.(runtime.BoolValue)
			return ok
		case "char":
			_, ok := value.(runtime.CharValue)
			return ok
		case "nil":
			_, ok := value.(runtime.NilValue)
			return ok
		case "i8", "i16", "i32", "i64", "i128", "u8", "u16", "u32", "u64", "u128":
			iv, ok := value.(runtime.IntegerValue)
			if !ok {
				return false
			}
			return string(iv.TypeSuffix) == name
		case "f32", "f64":
			fv, ok := value.(runtime.FloatValue)
			if !ok {
				return false
			}
			return string(fv.TypeSuffix) == name
		case "Error":
			_, ok := value.(runtime.ErrorValue)
			return ok
		default:
			if _, ok := i.interfaces[name]; ok {
				switch v := value.(type) {
				case *runtime.InterfaceValue:
					return i.interfaceMatches(v, name)
				case runtime.InterfaceValue:
					return i.interfaceMatches(&v, name)
				default:
					if structName, ok := i.getTypeNameForValue(value); ok {
						if _, ok := i.lookupImplEntry(structName, name); ok {
							return true
						}
					}
					return false
				}
			}
			if structVal, ok := value.(*runtime.StructInstanceValue); ok {
				if structVal.Definition != nil && structVal.Definition.Node != nil && structVal.Definition.Node.ID != nil {
					return structVal.Definition.Node.ID.Name == name
				}
			}
			return false
		}
	case *ast.GenericTypeExpression:
		if base, ok := t.Base.(*ast.SimpleTypeExpression); ok && base.Name.Name == "Array" {
			arr, ok := value.(*runtime.ArrayValue)
			if !ok {
				return false
			}
			if len(t.Arguments) == 0 {
				return true
			}
			elemType := t.Arguments[0]
			for _, el := range arr.Elements {
				if !i.matchesType(elemType, el) {
					return false
				}
			}
			return true
		}
		return true
	case *ast.FunctionTypeExpression:
		_, ok := value.(*runtime.FunctionValue)
		return ok
	case *ast.NullableTypeExpression:
		if _, ok := value.(runtime.NilValue); ok {
			return true
		}
		return i.matchesType(t.InnerType, value)
	case *ast.ResultTypeExpression:
		return i.matchesType(t.InnerType, value)
	case *ast.UnionTypeExpression:
		for _, member := range t.Members {
			if i.matchesType(member, value) {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func (i *Interpreter) coerceValueToType(typeExpr ast.TypeExpression, value runtime.Value) (runtime.Value, error) {
	switch t := typeExpr.(type) {
	case *ast.SimpleTypeExpression:
		if t.Name != nil {
			name := t.Name.Name
			if _, ok := i.interfaces[name]; ok {
				return i.coerceToInterfaceValue(name, value)
			}
		}
	}
	return value, nil
}

func (i *Interpreter) coerceToInterfaceValue(interfaceName string, value runtime.Value) (runtime.Value, error) {
	if ifaceVal, ok := value.(*runtime.InterfaceValue); ok {
		if i.interfaceMatches(ifaceVal, interfaceName) {
			return value, nil
		}
	}
	if ifaceVal, ok := value.(runtime.InterfaceValue); ok {
		if i.interfaceMatches(&ifaceVal, interfaceName) {
			return value, nil
		}
	}
	ifaceDef, ok := i.interfaces[interfaceName]
	if !ok {
		return nil, fmt.Errorf("Interface '%s' is not defined", interfaceName)
	}
	structName, ok := i.getTypeNameForValue(value)
	if !ok {
		return nil, fmt.Errorf("Value does not implement interface %s", interfaceName)
	}
	entry, ok := i.lookupImplEntry(structName, interfaceName)
	if !ok {
		return nil, fmt.Errorf("Type '%s' does not implement interface %s", structName, interfaceName)
	}
	methods := make(map[string]*runtime.FunctionValue, len(entry.methods))
	for name, fn := range entry.methods {
		methods[name] = fn
	}
	return &runtime.InterfaceValue{Interface: ifaceDef, Underlying: value, Methods: methods}, nil
}

func rescueMatches(pattern ast.Pattern, err runtime.Value, env *runtime.Environment) bool {
	switch p := pattern.(type) {
	case *ast.Identifier:
		env.Define(p.Name, err)
		return true
	case *ast.WildcardPattern:
		return true
	case *ast.LiteralPattern:
		if strLit, ok := p.Literal.(*ast.StringLiteral); ok {
			return valueToString(err) == strLit.Value
		}
		return false
	default:
		return false
	}
}

func rangeEndpoint(val runtime.Value) (int, error) {
	switch v := val.(type) {
	case runtime.IntegerValue:
		return int(v.Val.Int64()), nil
	case runtime.FloatValue:
		return int(v.Val), nil
	default:
		return 0, fmt.Errorf("range endpoint must be numeric, got %s", val.Kind())
	}
}
