package interpreter

import (
	"context"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type bytecodeOp int

const (
	bytecodeOpConst bytecodeOp = iota
	bytecodeOpLoadName
	bytecodeOpDeclareName
	bytecodeOpAssignName
	bytecodeOpDup
	bytecodeOpPop
	bytecodeOpBinary
	bytecodeOpUnary
	bytecodeOpRange
	bytecodeOpCast
	bytecodeOpStringInterpolation
	bytecodeOpPropagation
	bytecodeOpOrElse
	bytecodeOpSpawn
	bytecodeOpAwait
	bytecodeOpImplicitMember
	bytecodeOpIteratorLiteral
	bytecodeOpBreakpoint
	bytecodeOpPlaceholderLambda
	bytecodeOpMakeFunction
	bytecodeOpDefineFunction
	bytecodeOpDefineStruct
	bytecodeOpStructLiteral
	bytecodeOpMapLiteral
	bytecodeOpArrayLiteral
	bytecodeOpIndexGet
	bytecodeOpIndexSet
	bytecodeOpForLoop
	bytecodeOpCall
	bytecodeOpCallName
	bytecodeOpMemberAccess
	bytecodeOpMatch
	bytecodeOpRescue
	bytecodeOpRaise
	bytecodeOpEnsure
	bytecodeOpRethrow
	bytecodeOpEvalExpression
	bytecodeOpEvalStatement
	bytecodeOpPipe
	bytecodeOpJump
	bytecodeOpJumpIfFalse
	bytecodeOpJumpIfNil
	bytecodeOpEnterScope
	bytecodeOpExitScope
	bytecodeOpReturn
)

type bytecodeInstruction struct {
	op            bytecodeOp
	name          string
	operator      string
	value         runtime.Value
	target        int
	argCount      int
	node          ast.Node
	safe          bool
	preferMethods bool
}

type bytecodeProgram struct {
	instructions []bytecodeInstruction
}

type bytecodeVM struct {
	interp *Interpreter
	stack  []runtime.Value
	env    *runtime.Environment
	ip     int
}

func newBytecodeVM(interp *Interpreter, env *runtime.Environment) *bytecodeVM {
	return &bytecodeVM{
		interp: interp,
		env:    env,
		stack:  make([]runtime.Value, 0, 8),
	}
}

func (vm *bytecodeVM) run(program *bytecodeProgram) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode vm missing interpreter")
	}
	if program == nil {
		return nil, fmt.Errorf("bytecode program is nil")
	}
	if vm.env == nil {
		return nil, fmt.Errorf("bytecode vm missing environment")
	}

	vm.stack = vm.stack[:0]
	vm.ip = 0

	instructions := program.instructions
	for vm.ip < len(instructions) {
		instr := instructions[vm.ip]
		switch instr.op {
		case bytecodeOpConst:
			vm.stack = append(vm.stack, instr.value)
			vm.ip++
		case bytecodeOpLoadName:
			val, err := vm.env.Get(instr.name)
			if err != nil {
				return nil, err
			}
			vm.stack = append(vm.stack, val)
			vm.ip++
		case bytecodeOpDeclareName:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if vm.env.HasInCurrentScope(instr.name) {
					return nil, fmt.Errorf(":= requires at least one new binding")
				}
				vm.env.Define(instr.name, val)
				if vm.interp.currentPackage != "" && vm.env.Parent() == vm.interp.global {
					vm.interp.registerSymbol(instr.name, val)
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpAssignName:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if !vm.env.AssignExisting(instr.name, val) {
					vm.env.Define(instr.name, val)
					if vm.interp.currentPackage != "" && vm.env.Parent() == vm.interp.global {
						vm.interp.registerSymbol(instr.name, val)
					}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpDup:
			{
				if len(vm.stack) == 0 {
					return nil, fmt.Errorf("bytecode stack underflow")
				}
				vm.stack = append(vm.stack, vm.stack[len(vm.stack)-1])
				vm.ip++
			}
		case bytecodeOpPop:
			if _, err := vm.pop(); err != nil {
				return nil, err
			}
			vm.ip++
		case bytecodeOpBinary:
			{
				right, err := vm.pop()
				if err != nil {
					return nil, err
				}
				left, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if instr.operator == "+" {
					rawLeft := unwrapInterfaceValue(left)
					rawRight := unwrapInterfaceValue(right)
					if ls, ok := rawLeft.(runtime.StringValue); ok {
						rs, ok := rawRight.(runtime.StringValue)
						if !ok {
							return nil, fmt.Errorf("Arithmetic requires numeric operands")
						}
						vm.stack = append(vm.stack, runtime.StringValue{Val: ls.Val + rs.Val})
						vm.ip++
						break
					}
					if _, ok := rawRight.(runtime.StringValue); ok {
						return nil, fmt.Errorf("Arithmetic requires numeric operands")
					}
				}
				result, err := applyBinaryOperator(vm.interp, instr.operator, left, right)
				if err != nil {
					return nil, err
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpUnary:
			{
				operand, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if instr.operator == "" {
					return nil, fmt.Errorf("bytecode unary missing operator")
				}
				result, err := vm.interp.applyUnaryOperator(instr.operator, operand)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpRange:
			{
				endVal, err := vm.pop()
				if err != nil {
					return nil, err
				}
				startVal, err := vm.pop()
				if err != nil {
					return nil, err
				}
				rangeExpr, ok := instr.node.(*ast.RangeExpression)
				if !ok || rangeExpr == nil {
					return nil, fmt.Errorf("bytecode range expects node")
				}
				result, err := vm.interp.evaluateRangeValues(startVal, endVal, rangeExpr.Inclusive, vm.env)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpCast:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				castExpr, ok := instr.node.(*ast.TypeCastExpression)
				if !ok || castExpr == nil {
					return nil, fmt.Errorf("bytecode cast expects node")
				}
				result, err := vm.interp.castValueToType(castExpr.TargetType, val)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpStringInterpolation:
			{
				if instr.argCount < 0 {
					return nil, fmt.Errorf("bytecode string interpolation count invalid")
				}
				parts := make([]runtime.Value, instr.argCount)
				for idx := instr.argCount - 1; idx >= 0; idx-- {
					val, err := vm.pop()
					if err != nil {
						return nil, err
					}
					parts[idx] = val
				}
				var builder strings.Builder
				for _, part := range parts {
					str, err := vm.interp.stringifyValue(part, vm.env)
					if err != nil {
						return nil, err
					}
					builder.WriteString(str)
				}
				vm.stack = append(vm.stack, runtime.StringValue{Val: builder.String()})
				vm.ip++
			}
		case bytecodeOpPropagation:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if errVal, ok := asErrorValue(val); ok {
					return nil, raiseSignal{value: errVal}
				}
				if vm.interp.matchesType(ast.Ty("Error"), val) {
					return nil, raiseSignal{value: vm.interp.makeErrorValue(val, vm.env)}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpOrElse:
			{
				orElseExpr, ok := instr.node.(*ast.OrElseExpression)
				if !ok || orElseExpr == nil {
					return nil, fmt.Errorf("bytecode or-else expects node")
				}
				val, err := vm.interp.evaluateOrElseExpression(orElseExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpSpawn:
			{
				spawnExpr, ok := instr.node.(*ast.SpawnExpression)
				if !ok || spawnExpr == nil {
					return nil, fmt.Errorf("bytecode spawn expects node")
				}
				vm.interp.ensureConcurrencyBuiltins()
				capturedEnv := runtime.NewEnvironment(vm.env)
				task := ProcTask(nil)
				if vm.interp.execMode == execModeBytecode {
					if program, err := vm.interp.lowerExpressionToBytecode(spawnExpr.Expression); err == nil {
						task = func(ctx context.Context) (runtime.Value, error) {
							payload := payloadFromContext(ctx)
							if payload == nil {
								payload = &asyncContextPayload{kind: asyncContextFuture}
							} else {
								payload.kind = asyncContextFuture
							}
							return vm.interp.runAsyncBytecodeProgram(payload, program, capturedEnv)
						}
					}
				}
				if task == nil {
					task = vm.interp.makeAsyncTask(spawnExpr.Expression, vm.env)
				}
				future := vm.interp.executor.RunFuture(task)
				if future == nil {
					vm.stack = append(vm.stack, runtime.NilValue{})
				} else {
					vm.stack = append(vm.stack, future)
				}
				vm.ip++
			}
		case bytecodeOpAwait:
			{
				awaitExpr, ok := instr.node.(*ast.AwaitExpression)
				if !ok || awaitExpr == nil {
					return nil, fmt.Errorf("bytecode await expects node")
				}
				val, err := vm.interp.evaluateAwaitExpression(awaitExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpImplicitMember:
			{
				implicitExpr, ok := instr.node.(*ast.ImplicitMemberExpression)
				if !ok || implicitExpr == nil {
					return nil, fmt.Errorf("bytecode implicit member expects node")
				}
				val, err := vm.interp.evaluateImplicitMemberExpression(implicitExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpIteratorLiteral:
			{
				iterExpr, ok := instr.node.(*ast.IteratorLiteral)
				if !ok || iterExpr == nil {
					return nil, fmt.Errorf("bytecode iterator literal expects node")
				}
				val, err := vm.interp.evaluateIteratorLiteral(iterExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpBreakpoint:
			{
				breakExpr, ok := instr.node.(*ast.BreakpointExpression)
				if !ok || breakExpr == nil {
					return nil, fmt.Errorf("bytecode breakpoint expects node")
				}
				val, err := vm.interp.evaluateBreakpointExpression(breakExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpPlaceholderLambda:
			{
				expr, ok := instr.node.(ast.Expression)
				if !ok || expr == nil {
					return nil, fmt.Errorf("bytecode placeholder lambda expects expression node")
				}
				state := vm.interp.stateFromEnv(vm.env)
				if state.hasPlaceholderFrame() {
					val, err := vm.interp.evaluateExpression(expr, vm.env)
					if err != nil {
						return nil, err
					}
					if val == nil {
						val = runtime.NilValue{}
					}
					vm.stack = append(vm.stack, val)
					vm.ip++
					break
				}
				if instr.argCount <= 0 {
					return nil, fmt.Errorf("bytecode placeholder lambda missing arity")
				}
				closure := &placeholderClosure{
					interpreter: vm.interp,
					expression:  expr,
					env:         vm.env,
					plan:        placeholderPlan{paramCount: instr.argCount},
				}
				fn := runtime.NativeFunctionValue{
					Name:  "<placeholder>",
					Arity: instr.argCount,
					Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
						return closure.invoke(args)
					},
				}
				vm.stack = append(vm.stack, fn)
				vm.ip++
			}
		case bytecodeOpMakeFunction:
			{
				fn := &runtime.FunctionValue{Declaration: instr.node, Closure: vm.env}
				vm.stack = append(vm.stack, fn)
				vm.ip++
			}
		case bytecodeOpDefineFunction:
			{
				def, ok := instr.node.(*ast.FunctionDefinition)
				if !ok || def == nil {
					return nil, fmt.Errorf("bytecode define expects function definition")
				}
				val, err := vm.interp.evaluateFunctionDefinition(def, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpDefineStruct:
			{
				def, ok := instr.node.(*ast.StructDefinition)
				if !ok || def == nil {
					return nil, fmt.Errorf("bytecode define expects struct definition")
				}
				val, err := vm.interp.evaluateStructDefinition(def, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpStructLiteral:
			{
				lit, ok := instr.node.(*ast.StructLiteral)
				if !ok || lit == nil {
					return nil, fmt.Errorf("bytecode struct literal expects node")
				}
				val, err := vm.interp.evaluateStructLiteral(lit, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpMapLiteral:
			{
				lit, ok := instr.node.(*ast.MapLiteral)
				if !ok || lit == nil {
					return nil, fmt.Errorf("bytecode map literal expects node")
				}
				val, err := vm.interp.evaluateMapLiteral(lit, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpArrayLiteral:
			{
				if instr.argCount < 0 {
					return nil, fmt.Errorf("bytecode array literal count invalid")
				}
				values := make([]runtime.Value, instr.argCount)
				for idx := instr.argCount - 1; idx >= 0; idx-- {
					val, err := vm.pop()
					if err != nil {
						return nil, err
					}
					values[idx] = val
				}
				arr := vm.interp.newArrayValue(values, len(values))
				vm.stack = append(vm.stack, arr)
				vm.ip++
			}
		case bytecodeOpIndexGet:
			{
				idxVal, err := vm.pop()
				if err != nil {
					return nil, err
				}
				obj, err := vm.pop()
				if err != nil {
					return nil, err
				}
				result, err := vm.interp.indexGet(obj, idxVal)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpIndexSet:
			{
				idxVal, err := vm.pop()
				if err != nil {
					return nil, err
				}
				obj, err := vm.pop()
				if err != nil {
					return nil, err
				}
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if instr.operator == "" {
					return nil, fmt.Errorf("bytecode index set missing operator")
				}
				op := ast.AssignmentOperator(instr.operator)
				binaryOp, isCompound := binaryOpForAssignment(op)
				result, err := vm.interp.assignIndex(obj, idxVal, val, op, binaryOp, isCompound)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpForLoop:
			{
				loop, ok := instr.node.(*ast.ForLoop)
				if !ok || loop == nil {
					return nil, fmt.Errorf("bytecode for loop expects node")
				}
				val, err := vm.interp.evaluateForLoop(loop, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpCall:
			{
				if instr.argCount < 0 {
					return nil, fmt.Errorf("bytecode call arg count invalid")
				}
				args := make([]runtime.Value, instr.argCount)
				for idx := instr.argCount - 1; idx >= 0; idx-- {
					arg, err := vm.pop()
					if err != nil {
						return nil, err
					}
					args[idx] = arg
				}
				callee, err := vm.pop()
				if err != nil {
					return nil, err
				}
				var callNode *ast.FunctionCall
				if instr.node != nil {
					if call, ok := instr.node.(*ast.FunctionCall); ok {
						callNode = call
					}
				}
				result, err := vm.interp.callCallableValue(callee, args, vm.env, callNode)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpCallName:
			{
				if instr.argCount < 0 {
					return nil, fmt.Errorf("bytecode call arg count invalid")
				}
				args := make([]runtime.Value, instr.argCount)
				for idx := instr.argCount - 1; idx >= 0; idx-- {
					arg, err := vm.pop()
					if err != nil {
						return nil, err
					}
					args[idx] = arg
				}
				if instr.name == "" {
					return nil, fmt.Errorf("bytecode call missing target name")
				}
				calleeVal, err := vm.env.Get(instr.name)
				if err != nil {
					if dotIdx := strings.Index(instr.name, "."); dotIdx > 0 && dotIdx < len(instr.name)-1 {
						head := instr.name[:dotIdx]
						tail := instr.name[dotIdx+1:]
						receiver, recvErr := vm.env.Get(head)
						if recvErr != nil {
							if def, ok := vm.env.StructDefinition(head); ok {
								receiver = def
							} else {
								receiver = runtime.TypeRefValue{TypeName: head}
							}
						}
						member := ast.ID(tail)
						candidate, err := vm.interp.memberAccessOnValueWithOptions(receiver, member, vm.env, true)
						if err != nil {
							return nil, err
						}
						calleeVal = candidate
					} else {
						return nil, err
					}
				}
				var callNode *ast.FunctionCall
				if instr.node != nil {
					if call, ok := instr.node.(*ast.FunctionCall); ok {
						callNode = call
					}
				}
				result, err := vm.interp.callCallableValue(calleeVal, args, vm.env, callNode)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpMemberAccess:
			{
				obj, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if instr.safe && isNilRuntimeValue(obj) {
					vm.stack = append(vm.stack, runtime.NilValue{})
					vm.ip++
					break
				}
				memberExpr := ast.Expression(nil)
				if instr.node != nil {
					if memberNode, ok := instr.node.(*ast.MemberAccessExpression); ok {
						memberExpr = memberNode.Member
					} else if expr, ok := instr.node.(ast.Expression); ok {
						memberExpr = expr
					}
				}
				if memberExpr == nil {
					return nil, fmt.Errorf("bytecode member access missing member")
				}
				result, err := vm.interp.memberAccessOnValueWithOptions(obj, memberExpr, vm.env, instr.preferMethods)
				if err != nil {
					return nil, err
				}
				if result == nil {
					result = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, result)
				vm.ip++
			}
		case bytecodeOpMatch:
			{
				matchExpr, ok := instr.node.(*ast.MatchExpression)
				if !ok || matchExpr == nil {
					return nil, fmt.Errorf("bytecode match expects node")
				}
				val, err := vm.interp.evaluateMatchExpression(matchExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpRescue:
			{
				rescueExpr, ok := instr.node.(*ast.RescueExpression)
				if !ok || rescueExpr == nil {
					return nil, fmt.Errorf("bytecode rescue expects node")
				}
				val, err := vm.interp.evaluateRescueExpression(rescueExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpRaise:
			{
				raiseStmt, ok := instr.node.(*ast.RaiseStatement)
				if !ok || raiseStmt == nil {
					return nil, fmt.Errorf("bytecode raise expects node")
				}
				val, err := vm.interp.evaluateRaiseStatement(raiseStmt, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpEnsure:
			{
				ensureExpr, ok := instr.node.(*ast.EnsureExpression)
				if !ok || ensureExpr == nil {
					return nil, fmt.Errorf("bytecode ensure expects node")
				}
				val, err := vm.interp.evaluateEnsureExpression(ensureExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpRethrow:
			{
				rethrowStmt, ok := instr.node.(*ast.RethrowStatement)
				if !ok || rethrowStmt == nil {
					return nil, fmt.Errorf("bytecode rethrow expects node")
				}
				val, err := vm.interp.evaluateRethrowStatement(rethrowStmt, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpEvalExpression:
			{
				expr, ok := instr.node.(ast.Expression)
				if !ok || expr == nil {
					return nil, fmt.Errorf("bytecode eval expects expression node")
				}
				val, err := vm.interp.evaluateExpression(expr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpEvalStatement:
			{
				stmt, ok := instr.node.(ast.Statement)
				if !ok || stmt == nil {
					return nil, fmt.Errorf("bytecode eval expects statement node")
				}
				val, err := vm.interp.evaluateStatement(stmt, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpPipe:
			{
				rhs, ok := instr.node.(ast.Expression)
				if !ok || rhs == nil {
					return nil, fmt.Errorf("bytecode pipe expects expression node")
				}
				subject, err := vm.pop()
				if err != nil {
					return nil, err
				}
				val, err := vm.interp.evaluatePipeExpression(subject, rhs, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpJump:
			vm.ip = instr.target
		case bytecodeOpJumpIfFalse:
			{
				cond, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if !vm.interp.isTruthy(cond) {
					vm.ip = instr.target
					break
				}
				vm.ip++
			}
		case bytecodeOpJumpIfNil:
			{
				cond, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if isNilRuntimeValue(cond) {
					vm.ip = instr.target
					break
				}
				vm.ip++
			}
		case bytecodeOpEnterScope:
			vm.env = runtime.NewEnvironment(vm.env)
			vm.ip++
		case bytecodeOpExitScope:
			count := instr.argCount
			if count <= 0 {
				count = 1
			}
			for idx := 0; idx < count; idx++ {
				if vm.env.Parent() == nil {
					return nil, fmt.Errorf("bytecode scope underflow")
				}
				vm.env = vm.env.Parent()
			}
			vm.ip++
		case bytecodeOpReturn:
			return vm.pop()
		default:
			return nil, fmt.Errorf("bytecode opcode %d not implemented", instr.op)
		}
	}
	return runtime.NilValue{}, nil
}

func (vm *bytecodeVM) pop() (runtime.Value, error) {
	if len(vm.stack) == 0 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	last := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	return last, nil
}
