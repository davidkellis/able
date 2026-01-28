package interpreter

import (
	"context"
	"errors"
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
	bytecodeOpAssignPattern
	bytecodeOpAssignNameCompound
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
	bytecodeOpImplicitMemberSet
	bytecodeOpIteratorLiteral
	bytecodeOpBreakpoint
	bytecodeOpPlaceholderLambda
	bytecodeOpPlaceholderValue
	bytecodeOpIterInit
	bytecodeOpIterNext
	bytecodeOpIterClose
	bytecodeOpBindPattern
	bytecodeOpYield
	bytecodeOpMakeFunction
	bytecodeOpDefineFunction
	bytecodeOpDefineStruct
	bytecodeOpDefineUnion
	bytecodeOpDefineTypeAlias
	bytecodeOpDefineMethods
	bytecodeOpDefineInterface
	bytecodeOpDefineImplementation
	bytecodeOpDefineExtern
	bytecodeOpStructLiteral
	bytecodeOpMapLiteral
	bytecodeOpArrayLiteral
	bytecodeOpIndexGet
	bytecodeOpIndexSet
	bytecodeOpForLoop
	bytecodeOpCall
	bytecodeOpCallName
	bytecodeOpMemberAccess
	bytecodeOpMemberSet
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
	bytecodeOpLoopEnter
	bytecodeOpLoopExit
	bytecodeOpEnterScope
	bytecodeOpExitScope
	bytecodeOpReturn
)

func (vm *bytecodeVM) run(program *bytecodeProgram) (runtime.Value, error) {
	return vm.runResumable(program, false)
}

func (vm *bytecodeVM) runResumable(program *bytecodeProgram, resume bool) (result runtime.Value, err error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode vm missing interpreter")
	}
	if program == nil {
		return nil, fmt.Errorf("bytecode program is nil")
	}
	if vm.env == nil {
		return nil, fmt.Errorf("bytecode vm missing environment")
	}

	if !resume {
		vm.stack = vm.stack[:0]
		vm.iterStack = vm.iterStack[:0]
		vm.loopStack = vm.loopStack[:0]
		vm.ip = 0
	}
	defer func() {
		if err == nil || !errors.Is(err, errSerialYield) {
			vm.closeAllIterators()
			vm.loopStack = vm.loopStack[:0]
		}
	}()

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
				if instr.node != nil {
					err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
				}
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
					err := fmt.Errorf(":= requires at least one new binding")
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
					}
					return nil, err
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
		case bytecodeOpAssignPattern:
			{
				if err := vm.execAssignPattern(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpAssignNameCompound:
			{
				if err := vm.execAssignNameCompound(instr); err != nil {
					return nil, err
				}
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
				handled, err := vm.execBinary(instr)
				if err != nil {
					return nil, err
				}
				if handled {
					continue
				}
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
					err = vm.interp.wrapStandardRuntimeError(err)
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						if vm.handleLoopSignal(err) {
							continue
						}
					}
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
					err = vm.interp.wrapStandardRuntimeError(err)
					err = vm.interp.attachRuntimeContext(err, rangeExpr, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
					err = vm.interp.attachRuntimeContext(err, castExpr, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
				val, err := vm.runOrElseExpression(orElseExpr)
				if err != nil {
					if vm.handleLoopSignal(err) {
						continue
					}
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
				val, err := vm.runAwaitExpression(awaitExpr)
				if err != nil {
					if vm.handleLoopSignal(err) {
						continue
					}
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
				if implicitExpr.Member == nil {
					return nil, fmt.Errorf("Implicit member requires identifier")
				}
				state := vm.interp.stateFromEnv(vm.env)
				receiver, ok := state.currentImplicitReceiver()
				if !ok {
					return nil, fmt.Errorf("Implicit member '#%s' requires enclosing function with a first parameter", implicitExpr.Member.Name)
				}
				val, err := vm.interp.memberAccessOnValue(receiver, implicitExpr.Member, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpImplicitMemberSet:
			{
				if err := vm.execImplicitMemberSet(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpIteratorLiteral:
			{
				iterExpr, ok := instr.node.(*ast.IteratorLiteral)
				if !ok || iterExpr == nil {
					return nil, fmt.Errorf("bytecode iterator literal expects node")
				}
				val, err := vm.runIteratorLiteral(iterExpr)
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
				val, err := vm.runBreakpointExpression(breakExpr)
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
				var program *bytecodeProgram
				if lowered, err := vm.interp.lowerExpressionToBytecodeWithOptions(expr, false); err == nil {
					program = lowered
				} else if !errors.Is(err, errBytecodeUnsupported) {
					return nil, err
				}
				state := vm.interp.stateFromEnv(vm.env)
				if state.hasPlaceholderFrame() {
					var val runtime.Value
					var err error
					if program != nil {
						innerVM := newBytecodeVM(vm.interp, vm.env)
						val, err = innerVM.run(program)
					} else {
						val, err = vm.interp.evaluateExpression(expr, vm.env)
					}
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
					bytecode:    program,
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
		case bytecodeOpPlaceholderValue:
			{
				placeholderExpr, ok := instr.node.(*ast.PlaceholderExpression)
				if !ok || placeholderExpr == nil {
					return nil, fmt.Errorf("bytecode placeholder value expects placeholder expression")
				}
				val, err := vm.interp.evaluatePlaceholderExpression(placeholderExpr, vm.env)
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpMakeFunction:
			{
				fn := &runtime.FunctionValue{Declaration: instr.node, Closure: vm.env}
				if lambda, ok := instr.node.(*ast.LambdaExpression); ok && lambda != nil && lambda.Body != nil {
					if program, err := vm.interp.lowerExpressionToBytecodeWithOptions(lambda.Body, true); err == nil {
						fn.Bytecode = program
					}
				}
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
					err = vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
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
					err = vm.interp.attachRuntimeContext(err, def, vm.interp.stateFromEnv(vm.env))
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpDefineUnion:
			{
				if err := vm.execDefineUnion(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpDefineTypeAlias:
			{
				if err := vm.execDefineTypeAlias(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpDefineMethods:
			{
				if err := vm.execDefineMethods(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpDefineInterface:
			{
				if err := vm.execDefineInterface(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpDefineImplementation:
			{
				if err := vm.execDefineImplementation(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpDefineExtern:
			{
				if err := vm.execDefineExtern(instr); err != nil {
					return nil, err
				}
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
		case bytecodeOpIterInit:
			{
				iterable, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if err := vm.pushForIterator(iterable); err != nil {
					return nil, err
				}
				vm.ip++
			}
		case bytecodeOpIterNext:
			{
				val, done, err := vm.nextForIterator()
				if err != nil {
					return nil, err
				}
				vm.stack = append(vm.stack, val, runtime.BoolValue{Val: done})
				vm.ip++
			}
		case bytecodeOpIterClose:
			{
				if err := vm.closeForIterator(); err != nil {
					return nil, err
				}
				vm.ip++
			}
		case bytecodeOpBindPattern:
			{
				pattern, ok := instr.node.(ast.Pattern)
				if !ok || pattern == nil {
					return nil, fmt.Errorf("bytecode bind pattern expects pattern node")
				}
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if err := vm.interp.assignPattern(pattern, val, vm.env, true, nil); err != nil {
					err = vm.interp.attachRuntimeContext(err, pattern, vm.interp.stateFromEnv(vm.env))
					return nil, err
				}
				vm.ip++
			}
		case bytecodeOpYield:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if val == nil {
					val = runtime.NilValue{}
				}
				gen := vm.interp.currentGenerator()
				if gen == nil {
					return nil, fmt.Errorf("yield may only appear inside iterator literal")
				}
				if err := gen.emit(val); err != nil {
					return nil, err
				}
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
		case bytecodeOpIndexGet:
			{
				if err := vm.execIndexGet(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpIndexSet:
			{
				if err := vm.execIndexSet(instr); err != nil {
					return nil, err
				}
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
				if err := vm.execCall(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpCallName:
			{
				if err := vm.execCallName(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpMemberAccess:
			{
				if err := vm.execMemberAccess(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpMemberSet:
			{
				if err := vm.execMemberSet(instr); err != nil {
					return nil, err
				}
			}
		case bytecodeOpMatch:
			{
				matchExpr, ok := instr.node.(*ast.MatchExpression)
				if !ok || matchExpr == nil {
					return nil, fmt.Errorf("bytecode match expects node")
				}
				val, err := vm.runMatchExpression(matchExpr)
				if err != nil {
					err = vm.interp.attachRuntimeContext(err, matchExpr, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
				val, err := vm.runRescueExpression(rescueExpr)
				if err != nil {
					if vm.handleLoopSignal(err) {
						continue
					}
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
					err = vm.interp.attachRuntimeContext(err, raiseStmt, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
				val, err := vm.runEnsureExpression(ensureExpr)
				if err != nil {
					if vm.handleLoopSignal(err) {
						continue
					}
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
					err = vm.interp.attachRuntimeContext(err, rethrowStmt, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
					if vm.handleLoopSignal(err) {
						continue
					}
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
					if vm.handleLoopSignal(err) {
						continue
					}
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
					if vm.handleLoopSignal(err) {
						continue
					}
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
		case bytecodeOpLoopEnter:
			{
				if err := vm.pushLoopFrame(instr.loopBreak, instr.loopContinue); err != nil {
					return nil, err
				}
				vm.ip++
			}
		case bytecodeOpLoopExit:
			{
				if err := vm.popLoopFrame(); err != nil {
					return nil, err
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
			val, err := vm.pop()
			if err != nil {
				return nil, err
			}
			if instr.node != nil {
				state := vm.interp.stateFromEnv(vm.env)
				var context *runtimeDiagnosticContext
				if state != nil {
					context = &runtimeDiagnosticContext{
						node:      instr.node,
						callStack: state.snapshotCallStack(),
					}
				}
				return nil, returnSignal{value: val, context: context}
			}
			return val, nil
		default:
			return nil, fmt.Errorf("bytecode opcode %d not implemented", instr.op)
		}
	}
	return runtime.NilValue{}, err
}
