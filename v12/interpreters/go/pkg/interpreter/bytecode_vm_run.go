package interpreter

import (
	"errors"
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

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
	var serialSync *SerialExecutor
	if serial, ok := vm.interp.executor.(*SerialExecutor); ok {
		var payload *asyncContextPayload
		if vm.env != nil {
			payload = payloadFromState(vm.env.RuntimeData())
		}
		if payload == nil {
			if serial.beginSynchronousSectionIfNeeded() {
				serialSync = serial
			}
		}
	}
	if serialSync != nil {
		defer serialSync.endSynchronousSection()
	}
	program = vm.prepareRunProgram(program, resume)
	defer vm.finishRunResumable(&err)
	instructions := program.instructions
	validatedIntConsts := vm.validatedIntegerConstSlots(program)
	slotConstIntImmTable := vm.slotConstImmediateTable(program)
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	for vm.ip < len(instructions) {
		instr := &instructions[vm.ip]
		if statsEnabled {
			vm.interp.recordBytecodeOp(instr.op)
		}
		switch instr.op {
		case bytecodeOpConst:
			value := instr.value
			if intVal, ok := value.(runtime.IntegerValue); ok {
				needsValidation := vm.ip < 0 || vm.ip >= len(validatedIntConsts) || !validatedIntConsts[vm.ip]
				if needsValidation {
					info, err := getIntegerInfo(intVal.TypeSuffix)
					if err != nil {
						if instr.node != nil {
							err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						}
						return nil, err
					}
					if err := ensureFitsInteger(info, intVal.BigInt()); err != nil {
						err = vm.interp.wrapStandardRuntimeError(err)
						if instr.node != nil {
							err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						}
						return nil, err
					}
					if vm.ip >= 0 && vm.ip < len(validatedIntConsts) {
						validatedIntConsts[vm.ip] = true
					}
				}
			}
			vm.stack = append(vm.stack, value)
			vm.ip++
		case bytecodeOpLoadName:
			if statsEnabled {
				vm.interp.recordBytecodeLoadNameLookup()
			}
			var (
				val runtime.Value
				err error
			)
			if instr.nameSimple {
				val, err = vm.resolveCachedIdentifierName(program, vm.ip, instr.name)
			} else {
				val, err = vm.resolveCachedName(program, vm.ip, instr.name)
			}
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
			if err := vm.execAssignPattern(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpAssignNameCompound:
			if err := vm.execAssignNameCompound(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDup:
			if err := vm.execDup(); err != nil {
				return nil, err
			}
		case bytecodeOpPop:
			if err := vm.execPop(); err != nil {
				return nil, err
			}
		case bytecodeOpBinary,
			bytecodeOpBinaryIntAdd,
			bytecodeOpBinaryIntSub,
			bytecodeOpBinaryIntLessEqual,
			bytecodeOpBinaryIntDivCast,
			bytecodeOpBinaryIntAddSlotConst,
			bytecodeOpBinaryIntSubSlotConst,
			bytecodeOpBinaryIntLessEqualSlotConst:
			{
				handled, err := vm.execBinary(instr, slotConstIntImmTable)
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
				if err := vm.execStringInterpolation(instr); err != nil {
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						if vm.handleLoopSignal(err) {
							continue
						}
					}
					return nil, err
				}
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
				if vm.interp.matchesType(cachedSimpleTypeExpression("Error"), val) {
					return nil, raiseSignal{value: vm.interp.makeErrorValue(val, vm.env)}
				}
				vm.stack = append(vm.stack, val)
				vm.ip++
			}
		case bytecodeOpOrElse:
			{
				handled, err := vm.execOrElse(*instr)
				if err != nil {
					return nil, err
				}
				if handled {
					continue
				}
			}
		case bytecodeOpSpawn:
			if err := vm.execSpawn(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpAwait:
			{
				awaitExpr, ok := instr.node.(*ast.AwaitExpression)
				if !ok || awaitExpr == nil {
					return nil, fmt.Errorf("bytecode await expects node")
				}
				iterable, err := vm.pop()
				if err != nil {
					return nil, err
				}
				val, err := vm.runAwaitExpression(awaitExpr, iterable)
				if err != nil {
					if errors.Is(err, errSerialYield) {
						vm.stack = append(vm.stack, iterable)
						return nil, err
					}
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
			if err := vm.execImplicitMemberSet(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpIteratorLiteral:
			{
				iterExpr, ok := instr.node.(*ast.IteratorLiteral)
				if !ok || iterExpr == nil {
					return nil, fmt.Errorf("bytecode iterator literal expects node")
				}
				val, err := vm.runIteratorLiteral(iterExpr, instr.program)
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
			if err := vm.execPlaceholderLambda(instr); err != nil {
				return nil, err
			}
		case bytecodeOpPlaceholderValue:
			if err := vm.execPlaceholderValue(instr); err != nil {
				return nil, err
			}
		case bytecodeOpMakeFunction:
			{
				fn := &runtime.FunctionValue{Declaration: instr.node, Closure: vm.env}
				if lambda, ok := instr.node.(*ast.LambdaExpression); ok && lambda != nil && lambda.Body != nil {
					program, err := vm.interp.lowerExpressionToBytecodeWithOptions(lambda.Body, true)
					if err != nil {
						return nil, err
					}
					setFunctionBytecodeProgram(fn, program)
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
			if err := vm.execDefineUnion(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDefineTypeAlias:
			if err := vm.execDefineTypeAlias(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDefineMethods:
			if err := vm.execDefineMethods(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDefineInterface:
			if err := vm.execDefineInterface(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDefineImplementation:
			if err := vm.execDefineImplementation(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDefineExtern:
			if err := vm.execDefineExtern(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpImport:
			if err := vm.execImport(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpDynImport:
			if err := vm.execDynImport(*instr); err != nil {
				return nil, err
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
				if err := vm.execArrayLiteral(instr); err != nil {
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						if vm.handleLoopSignal(err) {
							continue
						}
					}
					return nil, err
				}
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
				var pattern ast.Pattern
				contextNode := instr.node
				isForLoop := false
				switch node := instr.node.(type) {
				case ast.Pattern:
					pattern = node
				case *ast.ForLoop:
					if node != nil {
						pattern = node.Pattern
						contextNode = node
						isForLoop = true
					}
				}
				if pattern == nil {
					return nil, fmt.Errorf("bytecode bind pattern expects pattern node")
				}
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				if isForLoop {
					assigned, err := vm.interp.assignPatternForLoop(pattern, val, vm.env)
					if err != nil {
						err = vm.interp.attachRuntimeContext(err, contextNode, vm.interp.stateFromEnv(vm.env))
						if vm.handleLoopSignal(err) {
							continue
						}
						return nil, err
					}
					if errVal, ok := asErrorValue(assigned); ok {
						if len(vm.loopStack) == 0 {
							return nil, fmt.Errorf("bytecode loop frame missing for pattern mismatch")
						}
						frame := vm.loopStack[len(vm.loopStack)-1]
						vm.env = frame.env
						vm.stack = append(vm.stack, errVal)
						vm.ip = frame.breakTarget
						continue
					}
					vm.ip++
					continue
				}
				if err := vm.interp.assignPattern(pattern, val, vm.env, true, nil); err != nil {
					err = vm.interp.attachRuntimeContext(err, contextNode, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
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
			if err := vm.execIndexGet(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpIndexSet:
			if err := vm.execIndexSet(*instr); err != nil {
				return nil, err
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
				newProg, err := vm.execCall(*instr, program)
				if err != nil {
					return nil, err
				}
				if newProg != nil {
					vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
					continue
				}
			}
		case bytecodeOpCallName:
			{
				newProg, err := vm.execCallName(*instr, program)
				if err != nil {
					return nil, err
				}
				if newProg != nil {
					vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
					continue
				}
			}
		case bytecodeOpCallMember:
			{
				newProg, err := vm.execCallMember(*instr, program)
				if err != nil {
					return nil, err
				}
				if newProg != nil {
					vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
					continue
				}
			}
		case bytecodeOpCallSelf:
			{
				newProg, err := vm.execCallSelf(*instr, program)
				if err != nil {
					return nil, err
				}
				if newProg != nil {
					vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
					continue
				}
			}
		case bytecodeOpCallSelfIntSubSlotConst:
			{
				newProg, err := vm.execCallSelfIntSubSlotConst(instr, slotConstIntImmTable, program)
				if err != nil {
					return nil, err
				}
				if newProg != nil {
					vm.switchRunProgram(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, newProg)
					continue
				}
			}
		case bytecodeOpMemberAccess:
			if err := vm.execMemberAccess(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpMemberSet:
			if err := vm.execMemberSet(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpMatch:
			{
				matchExpr, ok := instr.node.(*ast.MatchExpression)
				if !ok || matchExpr == nil {
					return nil, fmt.Errorf("bytecode match expects node")
				}
				subject, err := vm.pop()
				if err != nil {
					return nil, err
				}
				val, err := vm.runMatchExpression(matchExpr, subject)
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
			if err := vm.execRescue(*instr); err != nil {
				if vm.handleLoopSignal(err) {
					continue
				}
				return nil, err
			}
		case bytecodeOpRaise:
			{
				raiseStmt, ok := instr.node.(*ast.RaiseStatement)
				if !ok || raiseStmt == nil {
					return nil, fmt.Errorf("bytecode raise expects node")
				}
				val, err := vm.evalExpressionBytecode(raiseStmt.Expression, vm.env)
				if err != nil {
					err = vm.interp.attachRuntimeContext(err, raiseStmt, vm.interp.stateFromEnv(vm.env))
					if vm.handleLoopSignal(err) {
						continue
					}
					return nil, err
				}
				raiseErr := raiseSignal{value: vm.interp.makeErrorValue(val, vm.env)}
				err = vm.interp.attachRuntimeContext(raiseErr, raiseStmt, vm.interp.stateFromEnv(vm.env))
				if vm.handleLoopSignal(err) {
					continue
				}
				return nil, err
			}
		case bytecodeOpEnsure:
			if err := vm.execEnsureStart(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpEnsureEnd:
			if err := vm.execEnsureEnd(*instr); err != nil {
				if vm.handleLoopSignal(err) {
					continue
				}
				return nil, err
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
		case bytecodeOpBreakLabel:
			{
				if instr.name == "" {
					return nil, fmt.Errorf("bytecode break label missing")
				}
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				state := vm.interp.stateFromEnv(vm.env)
				if !state.hasBreakpoint(instr.name) {
					err := fmt.Errorf("Unknown break label '%s'", instr.name)
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, state)
					}
					return nil, err
				}
				return nil, breakSignal{label: instr.name, value: val}
			}
		case bytecodeOpBreakSignal:
			{
				val, err := vm.pop()
				if err != nil {
					return nil, err
				}
				return nil, breakSignal{value: val}
			}
		case bytecodeOpContinueSignal:
			{
				return nil, continueSignal{}
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
		case bytecodeOpJumpIfIntLessEqualSlotConstFalse:
			{
				if err := vm.execJumpIfIntLessEqualSlotConstFalse(instr, slotConstIntImmTable); err != nil {
					err = vm.interp.wrapStandardRuntimeError(err)
					if instr.node != nil {
						err = vm.interp.attachRuntimeContext(err, instr.node, vm.interp.stateFromEnv(vm.env))
						if vm.handleLoopSignal(err) {
							continue
						}
					}
					return nil, err
				}
			}
		case bytecodeOpReturnIfIntLessEqualSlotConst, bytecodeOpReturnConstIfIntLessEqualSlotConst:
			{
				var (
					val      runtime.Value
					returned bool
					err      error
				)
				if instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst {
					val, returned, err = vm.execReturnConstIfIntLessEqualSlotConst(instr, slotConstIntImmTable)
				} else {
					val, returned, err = vm.execReturnIfIntLessEqualSlotConst(instr, slotConstIntImmTable)
				}
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
				if !returned {
					continue
				}
				if vm.hasCallFrames() {
					if err := vm.finishInlineReturn(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, instr, val); err != nil {
						return nil, err
					}
					continue
				}
				if instr.node != nil {
					return nil, returnSignal{value: val, node: instr.node}
				}
				return val, nil
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
			if program.frameLayout == nil || program.frameLayout.needsEnvScopes {
				vm.env = runtime.NewEnvironment(vm.env)
			}
			vm.ip++
		case bytecodeOpExitScope:
			if program.frameLayout == nil || program.frameLayout.needsEnvScopes {
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
			}
			vm.ip++
		case bytecodeOpReturnBinaryIntAdd, bytecodeOpReturnBinaryIntAddI32:
			{
				val, err := vm.execReturnBinaryIntAdd(instr)
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
				if vm.hasCallFrames() {
					if err := vm.finishInlineReturn(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, nil, val); err != nil {
						return nil, err
					}
					continue
				}
				return val, nil
			}
		case bytecodeOpLoadSlot:
			vm.stack = append(vm.stack, vm.slots[instr.target])
			vm.ip++
		case bytecodeOpStoreSlot:
			if err := vm.execStoreSlot(instr); err != nil {
				return nil, err
			}
		case bytecodeOpStoreSlotNew:
			if err := vm.execStoreSlot(instr); err != nil {
				return nil, err
			}
		case bytecodeOpCompoundAssignSlot:
			if err := vm.execCompoundAssignSlot(*instr); err != nil {
				return nil, err
			}
		case bytecodeOpReturn:
			if len(vm.stack) == 0 {
				return nil, fmt.Errorf("bytecode stack underflow")
			}
			valIdx := len(vm.stack) - 1
			val := vm.stack[valIdx]
			vm.stack = vm.stack[:valIdx]
			if vm.hasCallFrames() {
				if err := vm.finishInlineReturn(&program, &instructions, &validatedIntConsts, &slotConstIntImmTable, instr, val); err != nil {
					return nil, err
				}
				continue
			}
			if instr.node != nil {
				return nil, returnSignal{value: val, node: instr.node}
			}
			return val, nil
		default:
			return nil, fmt.Errorf("bytecode opcode %d not implemented", instr.op)
		}
	}
	return runtime.NilValue{}, err
}
