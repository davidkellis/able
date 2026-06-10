package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type forLoopIteratorBytecodeCall struct {
	fn               *runtime.FunctionValue
	program          *bytecodeProgram
	layout           *bytecodeFrameLayout
	env              *runtime.Environment
	receiver         runtime.Value
	implicitReceiver runtime.Value
	hasReceiver      bool
	hasImplicit      bool
}

func (vm *bytecodeVM) resolveForLoopIteratorFrame(iterable runtime.Value) (forLoopIterator, error) {
	if frame, ok, err := vm.tryAdaptForLoopIteratorFrame(iterable); ok || err != nil {
		return frame, err
	}

	ident := ast.NewIdentifier("iterator")
	switch it := iterable.(type) {
	case *runtime.StructInstanceValue:
		member, err := vm.interp.structInstanceMember(it, ident, vm.env, true)
		if err != nil {
			if vm.interp.interfaceMethodResolver != nil {
				if resolved, found := vm.interp.interfaceMethodResolver(it, "Iterable", "iterator"); found && resolved != nil {
					value, callErr := vm.interp.CallFunction(resolved, []runtime.Value{it})
					if callErr != nil {
						return forLoopIterator{}, callErr
					}
					return vm.frameFromIteratorMethodResult(iterable, value)
				}
			}
			return forLoopIterator{}, err
		}
		return vm.callForLoopIteratorMethod(iterable, member)
	case *runtime.InterfaceValue:
		member, err := vm.interp.interfaceMember(it, ident)
		if err != nil {
			return forLoopIterator{}, err
		}
		return vm.callForLoopIteratorMethod(iterable, member)
	default:
		member, err := vm.interp.memberAccessOnValueWithOptions(iterable, ident, vm.env, true)
		if err == nil && member != nil {
			return vm.callForLoopIteratorMethod(iterable, member)
		}
		return forLoopIterator{}, fmt.Errorf("for-loop iterable of kind %s is not Iterable", iterable.Kind())
	}
}

func (vm *bytecodeVM) callForLoopIteratorMethod(iterable runtime.Value, member runtime.Value) (forLoopIterator, error) {
	value, err := vm.interp.CallFunction(member, nil)
	if err != nil {
		return forLoopIterator{}, err
	}
	return vm.frameFromIteratorMethodResult(iterable, value)
}

func (vm *bytecodeVM) frameFromIteratorMethodResult(iterable runtime.Value, value runtime.Value) (forLoopIterator, error) {
	if frame, ok, err := vm.tryAdaptForLoopIteratorFrame(value); ok || err != nil {
		return frame, err
	}
	if iterator, ok := value.(*runtime.IteratorValue); ok {
		return forLoopIterator{iter: iterator}, nil
	}
	return forLoopIterator{}, fmt.Errorf("iterator() on %s did not return Iterator", iterable.Kind())
}

func (vm *bytecodeVM) tryAdaptForLoopIteratorFrame(candidate runtime.Value) (forLoopIterator, bool, error) {
	switch it := candidate.(type) {
	case *runtime.IteratorValue:
		if it == nil {
			return forLoopIterator{}, false, nil
		}
		return forLoopIterator{iter: it}, true, nil
	case *runtime.InterfaceValue:
		if it == nil {
			return forLoopIterator{}, false, nil
		}
		return vm.tryAdaptForLoopIteratorFrame(it.Underlying)
	case runtime.InterfaceValue:
		return vm.tryAdaptForLoopIteratorFrame(it.Underlying)
	case *runtime.StructInstanceValue:
		if it == nil {
			return forLoopIterator{}, false, nil
		}
		nextVal, err := vm.interp.memberAccessOnValueWithOptions(it, ast.NewIdentifier("next"), vm.env, true)
		if err != nil || nextVal == nil {
			return forLoopIterator{}, false, nil
		}
		frame := forLoopIterator{}
		vm.configureForLoopIteratorNext(&frame, nextVal)
		if closeVal, err := vm.interp.memberAccessOnValueWithOptions(it, ast.NewIdentifier("close"), vm.env, true); err == nil {
			frame.closeCallable = closeVal
		}
		return frame, true, nil
	default:
		return forLoopIterator{}, false, nil
	}
}

func (vm *bytecodeVM) configureForLoopIteratorNext(frame *forLoopIterator, next runtime.Value) {
	frame.nextCallable = next
	if vm == nil || vm.interp == nil || next == nil {
		return
	}
	fn, partial, receiver, hasReceiver, ok := bytecodeResolveDirectFunctionCallTarget(next)
	if !ok || fn == nil {
		return
	}
	args := frame.nextArgs[:0]
	if hasReceiver {
		frame.nextArgs[0] = bytecodeSlotReadValue(receiver)
		args = frame.nextArgs[:1]
	}
	if len(args) < minArgsForFunctionValue(fn) || !vm.interp.matchesSingleRuntimeOverload(fn, args) {
		return
	}
	frame.nextFn = fn
	frame.nextPartial = partial
	frame.nextReceiver = receiver
	frame.nextHasReceiver = hasReceiver
	frame.nextBytecode = vm.prepareForLoopIteratorBytecodeCall(fn, receiver, hasReceiver)
}

func (vm *bytecodeVM) nextForLoopIteratorFrame(frame *forLoopIterator) (runtime.Value, bool, error) {
	res, err := vm.callForLoopIteratorNext(frame)
	if err != nil {
		return nil, true, err
	}
	if vm.interp.isIteratorEnd(res) {
		return runtime.IteratorEnd, true, nil
	}
	if res == nil {
		return runtime.NilValue{}, false, nil
	}
	return res, false, nil
}

func (vm *bytecodeVM) callForLoopIteratorNext(frame *forLoopIterator) (runtime.Value, error) {
	if frame.nextBytecode != nil {
		return vm.callForLoopIteratorBytecodeNext(frame.nextBytecode)
	}
	if frame.nextFn != nil {
		args := frame.nextArgs[:0]
		if frame.nextHasReceiver {
			frame.nextArgs[0] = bytecodeSlotReadValue(frame.nextReceiver)
			args = frame.nextArgs[:1]
		}
		return vm.interp.callResolvedFunctionValue(frame.nextFn, frame.nextPartial, args, vm.env, nil, true)
	}
	return vm.interp.CallFunction(frame.nextCallable, nil)
}

func (vm *bytecodeVM) prepareForLoopIteratorBytecodeCall(fn *runtime.FunctionValue, receiver runtime.Value, hasReceiver bool) *forLoopIteratorBytecodeCall {
	if vm == nil || vm.interp == nil || fn == nil {
		return nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		return nil
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil || program.frameLayout == nil {
		return nil
	}
	layout := program.frameLayout
	switch {
	case layout.methodShorthand:
		if !hasReceiver || layout.paramSlots != 0 {
			return nil
		}
	case hasReceiver:
		if layout.paramSlots != 1 {
			return nil
		}
	default:
		if layout.paramSlots != 0 {
			return nil
		}
	}
	if hasReceiver {
		if err := vm.interp.enforceMethodSetConstraints(fn, receiver); err != nil {
			return nil
		}
	}

	callPlan := vm.interp.functionRuntimeGenericBindingPlan(fn)
	if callPlan == nil || callPlan.hasGenericConstraints {
		return nil
	}
	needsCallLocalTypeBindings := hasReceiver && callPlan.callLocalUsed
	if callPlan.explicitUsed && (!needsCallLocalTypeBindings || !forLoopIteratorMethodSetCoversFunctionGenerics(fn, decl)) {
		return nil
	}
	reuseClosureEnv := canReuseCallableClosureEnvForBytecode(program, callPlan.explicitUsed, fn.Closure)
	if reuseClosureEnv && needsCallLocalTypeBindings {
		reuseClosureEnv = false
	}

	localEnv := fn.Closure
	if !reuseClosureEnv {
		callTypeBindings := functionCallTypeBindingSet{}
		if needsCallLocalTypeBindings {
			callTypeBindings.callLocal, callTypeBindings.receiverType = vm.interp.callLocalTypeBindingValuesAndReceiverTypeIfAny(fn, receiver)
		}
		if callTypeBindings.empty() {
			return nil
		}
		reusableEnv, ok := vm.interp.reusableBytecodeCallEnvForResolvedBindings(fn, decl, nil, program, callTypeBindings)
		if !ok {
			return nil
		}
		localEnv = reusableEnv
	}

	argReceiver := receiver
	if hasReceiver && !layout.methodShorthand {
		argReceiver = bytecodeSlotReadValue(receiver)
		paramType := inlineParamType(layout, 0)
		if inlineParamNeedsRuntimeCoercion(layout, 0, fn) &&
			!forLoopIteratorResolvedReceiverParamNoCoercion(fn, paramType) &&
			!inlineParamCoercionUnnecessary(vm.interp, layout, 0, paramType, argReceiver) {
			return nil
		}
	}
	hasImplicit := layout.usesImplicitMember && (layout.methodShorthand || layout.paramSlots > 0)
	return &forLoopIteratorBytecodeCall{
		fn:               fn,
		program:          program,
		layout:           layout,
		env:              localEnv,
		receiver:         argReceiver,
		implicitReceiver: receiver,
		hasReceiver:      hasReceiver && !layout.methodShorthand,
		hasImplicit:      hasImplicit,
	}
}

func forLoopIteratorMethodSetCoversFunctionGenerics(fn *runtime.FunctionValue, decl *ast.FunctionDefinition) bool {
	if fn == nil || decl == nil || len(decl.GenericParams) == 0 {
		return true
	}
	if fn.MethodSet == nil || len(fn.MethodSet.GenericParams) == 0 {
		return false
	}
	names := genericNameSet(fn.MethodSet.GenericParams)
	for _, param := range decl.GenericParams {
		if param == nil || param.Name == nil || param.Name.Name == "" {
			continue
		}
		if _, ok := names[param.Name.Name]; !ok {
			return false
		}
	}
	return true
}

func forLoopIteratorResolvedReceiverParamNoCoercion(fn *runtime.FunctionValue, paramType ast.TypeExpression) bool {
	if fn == nil || fn.MethodSet == nil || paramType == nil {
		return false
	}
	simple, ok := paramType.(*ast.SimpleTypeExpression)
	if !ok || simple == nil || simple.Name == nil {
		return false
	}
	switch simple.Name.Name {
	case "Self", "SelfType":
		return true
	default:
		return false
	}
}

func (vm *bytecodeVM) callForLoopIteratorBytecodeNext(plan *forLoopIteratorBytecodeCall) (runtime.Value, error) {
	if vm == nil || vm.interp == nil || plan == nil || plan.program == nil || plan.layout == nil || plan.env == nil {
		return nil, fmt.Errorf("bytecode iterator next call is not initialized")
	}
	inner := vm.interp.acquireBytecodeVM(plan.env)

	var implicitState *evalState
	if plan.hasImplicit {
		implicitState = vm.interp.stateFromEnv(plan.env)
		implicitState.pushImplicitReceiver(plan.implicitReceiver)
	}

	slots := inner.acquireSlotFrame(plan.layout.slotCount)
	if plan.hasReceiver && len(slots) > 0 {
		slots[0] = plan.receiver
	}
	if plan.layout.selfCallSlot >= 0 && plan.layout.selfCallSlot < len(slots) {
		slots[plan.layout.selfCallSlot] = plan.fn
	}
	inner.slots = slots

	result, err := inner.run(plan.program)
	if implicitState != nil {
		implicitState.popImplicitReceiver()
	}
	vm.interp.releaseBytecodeVM(inner)
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
		result = runtime.NilValue{}
	}
	return result, nil
}

func (vm *bytecodeVM) closeForLoopIterator(frame *forLoopIterator) {
	if frame == nil {
		return
	}
	if frame.iter != nil {
		frame.iter.Close()
	}
	if frame.closeCallable != nil {
		_, _ = vm.interp.CallFunction(frame.closeCallable, nil)
	}
}
