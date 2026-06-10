package interpreter

import (
	"context"
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type forLoopIterator struct {
	values          []runtime.Value
	index           int
	iter            *runtime.IteratorValue
	nextCallable    runtime.Value
	closeCallable   runtime.Value
	nextFn          *runtime.FunctionValue
	nextPartial     runtime.Value
	nextReceiver    runtime.Value
	nextHasReceiver bool
	nextArgs        [1]runtime.Value
	nextBytecode    *forLoopIteratorBytecodeCall
}

func (vm *bytecodeVM) pushForIterator(value runtime.Value) error {
	switch it := value.(type) {
	case *runtime.ArrayValue:
		state, err := vm.interp.ensureArrayState(it, 0)
		if err != nil {
			return err
		}
		vm.iterStack = append(vm.iterStack, forLoopIterator{values: state.Values})
		return nil
	case *runtime.IteratorValue:
		vm.iterStack = append(vm.iterStack, forLoopIterator{iter: it})
		return nil
	default:
		frame, err := vm.resolveForLoopIteratorFrame(value)
		if err != nil {
			return err
		}
		vm.iterStack = append(vm.iterStack, frame)
		return nil
	}
}

func (vm *bytecodeVM) nextForIterator() (runtime.Value, bool, error) {
	if len(vm.iterStack) == 0 {
		return nil, true, fmt.Errorf("bytecode iterator stack underflow")
	}
	frame := &vm.iterStack[len(vm.iterStack)-1]
	if frame.values != nil {
		if frame.index >= len(frame.values) {
			return runtime.NilValue{}, true, nil
		}
		val := frame.values[frame.index]
		frame.index++
		if val == nil {
			val = runtime.NilValue{}
		}
		return val, false, nil
	}
	if frame.iter == nil {
		if frame.nextCallable != nil {
			return vm.nextForLoopIteratorFrame(frame)
		}
		return runtime.NilValue{}, true, nil
	}
	raw, done, err := frame.iter.NextRaw()
	if err != nil {
		return nil, true, err
	}
	val := bytecodeValueFromRuntimeRawValue(raw)
	if val == nil {
		val = runtime.NilValue{}
	}
	return val, done, nil
}

func (vm *bytecodeVM) closeForIterator() error {
	if len(vm.iterStack) == 0 {
		return fmt.Errorf("bytecode iterator stack underflow")
	}
	last := vm.iterStack[len(vm.iterStack)-1]
	vm.iterStack = vm.iterStack[:len(vm.iterStack)-1]
	vm.closeForLoopIterator(&last)
	return nil
}

func (vm *bytecodeVM) closeAllIterators() {
	for idx := len(vm.iterStack) - 1; idx >= 0; idx-- {
		vm.closeForLoopIterator(&vm.iterStack[idx])
	}
	vm.iterStack = vm.iterStack[:0]
}

func (vm *bytecodeVM) execSpawn(instr bytecodeInstruction) error {
	spawnExpr, ok := instr.node.(*ast.SpawnExpression)
	if !ok || spawnExpr == nil {
		return fmt.Errorf("bytecode spawn expects node")
	}
	vm.markAllBytecodeArrayOwnershipEscaped(bytecodeArrayOwnershipEscapeFuture)
	vm.interp.ensureConcurrencyBuiltins()
	vm.interp.ensureMultiThread()
	capturedEnv := runtime.NewEnvironment(vm.env)
	program := instr.program
	if program == nil {
		var err error
		program, err = vm.interp.lowerExpressionToBytecode(spawnExpr.Expression)
		if err != nil {
			return err
		}
	}
	task := func(ctx context.Context) (runtime.Value, error) {
		payload := payloadFromContext(ctx)
		if payload == nil {
			payload = &asyncContextPayload{kind: asyncContextFuture}
		} else {
			payload.kind = asyncContextFuture
		}
		return vm.interp.runAsyncBytecodeProgram(payload, program, capturedEnv)
	}
	future := vm.interp.executor.RunFuture(task)
	if future == nil {
		vm.appendStackValue(runtime.NilValue{})
	} else {
		vm.appendStackValue(future)
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) evalExpressionBytecode(expr ast.Expression, env *runtime.Environment) (runtime.Value, error) {
	return vm.evalExpressionBytecodeWithOptions(expr, env, true)
}

func (vm *bytecodeVM) evalExpressionBytecodeWithOptions(expr ast.Expression, env *runtime.Environment, allowPlaceholderLambda bool) (runtime.Value, error) {
	if expr == nil {
		return runtime.NilValue{}, nil
	}
	program, err := vm.interp.lowerExpressionToBytecodeWithOptions(expr, allowPlaceholderLambda)
	if err != nil {
		return nil, err
	}
	return vm.evalExpressionBytecodeProgram(program, env)
}

func (vm *bytecodeVM) evalExpressionBytecodeProgram(program *bytecodeProgram, env *runtime.Environment) (runtime.Value, error) {
	if program == nil {
		return runtime.NilValue{}, nil
	}
	innerVM := vm.interp.acquireBytecodeVM(env)
	defer vm.interp.releaseBytecodeVM(innerVM)
	val, err := innerVM.run(program)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return runtime.NilValue{}, nil
	}
	return val, nil
}

func (vm *bytecodeVM) runMatchExpression(expr *ast.MatchExpression, subject runtime.Value) (runtime.Value, error) {
	plans := vm.interp.matchExpressionClausePlans(expr)
	programs := vm.interp.matchExpressionBytecodePrograms(expr)
	for idx, clause := range expr.Clauses {
		if clause == nil {
			continue
		}
		plan := plans[idx]
		clauseEnv, matched, transientEnv, transientBindings := vm.interp.matchPatternForClauseTransient(clause.Pattern, subject, vm.env, plan)
		if !matched {
			vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
			continue
		}
		if clause.Guard != nil {
			guardProgram, err := vm.interp.matchExpressionClauseBytecodeProgram(programs, idx, clause.Guard, bytecodeMatchClauseProgramGuard)
			if err != nil {
				vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
				return nil, err
			}
			guardVal, err := vm.evalExpressionBytecodeProgram(guardProgram, clauseEnv)
			if err != nil {
				vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
				return nil, err
			}
			if !vm.interp.isTruthy(guardVal) {
				vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
				continue
			}
		}
		bodyProgram, err := vm.interp.matchExpressionClauseBytecodeProgram(programs, idx, clause.Body, bytecodeMatchClauseProgramBody)
		if err != nil {
			vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
			return nil, err
		}
		result, err := vm.evalExpressionBytecodeProgram(bodyProgram, clauseEnv)
		vm.interp.releaseTransientClauseMatch(transientEnv, transientBindings)
		return result, err
	}
	return nil, fmt.Errorf("Non-exhaustive match")
}

func (vm *bytecodeVM) runBreakpointExpression(expr *ast.BreakpointExpression) (runtime.Value, error) {
	if expr.Label == nil {
		return nil, fmt.Errorf("Breakpoint expression requires label")
	}
	label := expr.Label.Name
	state := vm.interp.stateFromEnv(vm.env)
	state.pushBreakpoint(label)
	defer state.popBreakpoint()
	for {
		val, err := vm.evalExpressionBytecode(expr.Body, vm.env)
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

func (vm *bytecodeVM) runIteratorLiteral(expr *ast.IteratorLiteral, program *bytecodeProgram) (runtime.Value, error) {
	bindingNames := iteratorLiteralBindingNames(expr)
	iterCapacity := len(bindingNames)
	iterEnv := runtime.NewEnvironmentWithValueCapacity(vm.env, iterCapacity)
	if program == nil && expr != nil {
		lowered, loweredBindingNames, err := vm.interp.lowerIteratorLiteralBodyToBytecode(expr, vm.env)
		if err != nil {
			return nil, err
		}
		program = lowered
		bindingNames = loweredBindingNames
	}
	instance := newGeneratorInstanceWithBytecode(vm.interp, iterEnv, expr.Body, program)
	controller := instance.controllerValue()
	if len(bindingNames) > 0 {
		instance.bytecodeSlotArg = controller
		instance.bytecodeSlotArgCount = len(bindingNames)
	}
	for _, name := range bindingNames {
		iterEnv.DefineWithoutMerge(name, controller)
	}
	return runtime.NewIteratorValueWithRaw(func() (runtime.RawValue, bool, error) {
		return instance.nextRaw()
	}, instance.close), nil
}

func (vm *bytecodeVM) runAwaitExpression(expr *ast.AwaitExpression, iterable runtime.Value) (runtime.Value, error) {
	payload := payloadFromState(vm.runtimeData())
	if payload == nil || payload.kind != asyncContextFuture {
		return nil, fmt.Errorf("await expressions must run inside an asynchronous task")
	}

	state := payload.getAwaitState(expr)
	if state == nil {
		arms, err := vm.interp.collectAwaitArms(iterable, vm.env)
		if err != nil {
			return nil, err
		}
		if len(arms) == 0 {
			return nil, fmt.Errorf("await requires at least one arm")
		}
		var defaultArm *awaitArmState
		for _, arm := range arms {
			if arm != nil && arm.isDefault {
				if defaultArm != nil {
					return nil, fmt.Errorf("await accepts at most one default arm")
				}
				defaultArm = arm
			}
		}
		state = &awaitEvalState{
			env:        vm.env,
			arms:       arms,
			defaultArm: defaultArm,
			payload:    payload,
		}
		state.ensureWaitCh()
		vm.interp.ensureConcurrencyBuiltins()
		if vm.interp.awaitWakerStruct == nil {
			return nil, fmt.Errorf("Await waker builtins are not initialized")
		}
		waker, err := vm.interp.makeAwaitWaker(payload, state)
		if err != nil {
			return nil, err
		}
		state.waker = waker
		payload.setAwaitState(expr, state)
	}

	for {
		winner, err := vm.interp.selectReadyAwaitArm(state, vm.env)
		if err != nil {
			return nil, err
		}
		if winner != nil {
			return vm.interp.completeAwait(payload, expr, state, winner, vm.env)
		}
		if state.defaultArm != nil {
			return vm.interp.completeAwait(payload, expr, state, state.defaultArm, vm.env)
		}
		if payload.handle != nil && payload.handle.CancelRequested() {
			vm.interp.cleanupAwaitState(payload, expr, state, vm.env)
			return nil, context.Canceled
		}
		if state.consumeWakePending() {
			vm.interp.clearAwaitRegistrations(state, vm.env)
			continue
		}
		if state.beginWaiting() {
			if err := vm.interp.registerAwaitState(state, vm.env); err != nil {
				state.clearWaiting()
				return nil, err
			}
		}

		waitCh := state.ensureWaitCh()
		payload.setAwaitBlocked(true)

		if _, ok := vm.interp.executor.(*SerialExecutor); ok {
			return nil, errSerialYield
		}

		var handle *runtime.FutureValue
		if payload != nil {
			handle = payload.handle
		}
		vm.interp.markBlocked(handle)
		ctx := payload.handle.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		select {
		case <-waitCh:
		case <-ctx.Done():
			vm.interp.markUnblocked(handle)
			payload.setAwaitBlocked(false)
			vm.interp.cleanupAwaitState(payload, expr, state, vm.env)
			return nil, ctx.Err()
		}
		vm.interp.markUnblocked(handle)
		payload.setAwaitBlocked(false)
		vm.interp.clearAwaitRegistrations(state, vm.env)
		state.clearWaiting()
	}
}
