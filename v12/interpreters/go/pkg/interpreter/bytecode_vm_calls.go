package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func inlineCallFunctionValue(callee runtime.Value) (*runtime.FunctionValue, runtime.Value, bool, bool) {
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		if fn == nil {
			return nil, nil, false, false
		}
		return fn, nil, false, true
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, nil, false, false
		}
		switch method := fn.Method.(type) {
		case *runtime.FunctionValue:
			if method == nil {
				return nil, nil, false, false
			}
			return method, fn.Receiver, true, true
		}
	case runtime.BoundMethodValue:
		switch method := fn.Method.(type) {
		case *runtime.FunctionValue:
			if method == nil {
				return nil, nil, false, false
			}
			return method, fn.Receiver, true, true
		}
	}
	return nil, nil, false, false
}

func bytecodeCanUseSelfFastFrame(currentProgram *bytecodeProgram, calleeProgram *bytecodeProgram, currentEnv *runtime.Environment, calleeEnv *runtime.Environment) bool {
	return currentProgram != nil && currentProgram == calleeProgram && currentEnv != nil && currentEnv == calleeEnv
}

func inlineParamCoercionUnnecessary(interp *Interpreter, layout *bytecodeFrameLayout, idx int, typeExpr ast.TypeExpression, val runtime.Value) bool {
	if layout != nil && idx >= 0 {
		if idx < len(layout.paramExactStructDef) {
			if def := layout.paramExactStructDef[idx]; def != nil {
				return inlineExactNamedStructNoCoercionBytecodeExactDef(def, val)
			}
		}
		if idx < len(layout.paramSimpleChecks) {
			if check := layout.paramSimpleChecks[idx]; check != bytecodeSimpleTypeCheckUnknown {
				return inlineCoercionUnnecessaryBySimpleCheck(check, val)
			}
		}
		if idx < len(layout.paramSimpleTypes) {
			if typeName := layout.paramSimpleTypes[idx]; typeName != "" {
				return inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(interp, typeName, val)
			}
		}
	}
	return inlineCoercionUnnecessaryWithInterpreter(interp, typeExpr, val)
}

func inlineParamSimpleType(layout *bytecodeFrameLayout, idx int) string {
	if layout == nil || idx < 0 || idx >= len(layout.paramSimpleTypes) {
		return ""
	}
	return layout.paramSimpleTypes[idx]
}

func inlineParamType(layout *bytecodeFrameLayout, idx int) ast.TypeExpression {
	if layout == nil || idx < 0 || idx >= len(layout.paramTypes) {
		return nil
	}
	return layout.paramTypes[idx]
}

func inlineParamNeedsRuntimeCoercion(layout *bytecodeFrameLayout, idx int, fn *runtime.FunctionValue) bool {
	if layout == nil || idx < 0 || idx >= len(layout.paramNeedsCoercion) {
		return false
	}
	if !layout.paramNeedsCoercion[idx] {
		return false
	}
	if fn != nil && fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return !paramUsesGeneric(inlineParamType(layout, idx), bytecodeFunctionReturnGenericNames(fn))
	}
	return true
}

func (vm *bytecodeVM) copyInlineCallArgToSlot(dst []runtime.Value, target int, value runtime.Value) {
	if target < 0 || target >= len(dst) {
		return
	}
	switch raw := value.(type) {
	case *bytecodeRawI32StackCell:
		if raw != nil {
			dst[target] = bytecodeRawI32SlotCachedValue(raw.Val)
			return
		}
	case *bytecodeRawI64SlotCell:
		if raw != nil {
			dst[target] = bytecodeRawI64ResultValue(raw.Val)
			return
		}
	case *bytecodeRawIntegerSlotCell:
		if raw != nil {
			dst[target] = bytecodeRawIntegerResultValue(raw.TypeSuffix, raw.Raw)
			return
		}
	case *bytecodeRawIntegerReturnScratch:
		if raw != nil {
			dst[target] = bytecodeRawIntegerResultValue(raw.TypeSuffix, raw.Raw)
			return
		}
	}
	dst[target] = value
}

func (vm *bytecodeVM) inlineCopyArgsToSlots(dst []runtime.Value, src []runtime.Value, count int) {
	if count <= 0 {
		return
	}
	for idx := 0; idx < count; idx++ {
		vm.copyInlineCallArgToSlot(dst, idx, src[idx])
	}
}

func bytecodeFunctionNeedsCallLocalBindings(vm *bytecodeVM, fn *runtime.FunctionValue) bool {
	if vm == nil || vm.interp == nil || fn == nil {
		return false
	}
	return vm.interp.functionNeedsCallLocalTypeBindings(fn)
}

func bytecodeInlineSkipsGenericLambda(fn *runtime.FunctionValue) bool {
	if fn == nil {
		return false
	}
	lambda, ok := fn.Declaration.(*ast.LambdaExpression)
	return ok && lambda != nil && len(lambda.GenericParams) > 0
}

// tryInlineCall attempts to set up an inline call frame for a slot-enabled
// function value. Returns the new program to switch to, or nil if the
// function cannot be inlined (the caller should fall back to callCallableValue).
func (vm *bytecodeVM) tryInlineCall(callee runtime.Value, args []runtime.Value, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	fn, ok := callee.(*runtime.FunctionValue)
	if !ok || fn == nil {
		return nil, nil
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil
	}
	layout := prog.frameLayout
	// Method shorthand requires implicit receiver adjustment and extra arg.
	if layout.methodShorthand {
		return nil, nil
	}
	// Require exact arity match (skip optional params).
	paramCount := layout.paramSlots
	if len(args) != paramCount {
		return nil, nil
	}
	// Skip if call site has type arguments that need binding.
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	// Skip if function belongs to a generic method set.
	if bytecodeFunctionNeedsCallLocalBindings(vm, fn) {
		return nil, nil
	}
	if bytecodeInlineSkipsGenericLambda(fn) {
		return nil, nil
	}
	if decl, ok := fn.Declaration.(*ast.FunctionDefinition); ok {
		if receiver, hasReceiver := resolveMethodSetReceiver(decl, args); hasReceiver {
			if err := vm.interp.enforceMethodSetConstraints(fn, receiver); err != nil {
				return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
			}
		}
	}

	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)

	// Coerce parameters when the cached layout metadata says the declared type
	// can actually require runtime coercion.
	if !layout.anyParamCoercion {
		vm.inlineCopyArgsToSlots(slots, args, paramCount)
		for idx := 0; idx < paramCount; idx++ {
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, args[idx])
		}
	} else {
		for idx := 0; idx < paramCount; idx++ {
			arg := args[idx]
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, err
					}
					arg = coerced
				}
			}
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, arg)
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}
	calleeEnv := vm.bytecodeCalleeEnv(fn.Closure)

	// Push implicit receiver only when the function body uses #member syntax.
	hasImplicit := paramCount > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(calleeEnv)
		state.pushImplicitReceiver(bytecodeStackSnapshotValue(args[0]))
	}

	// Push call frame.
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, calleeEnv)
	returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, returnGenericNames, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)

	// Set up new frame.
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = calleeEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)

	return prog, nil
}

// tryInlineCallFromStack mirrors tryInlineCall but reads arguments directly
// from the operand-stack region beginning at argBase. On success it truncates
// to truncateTo (dropping arguments and, for bytecodeOpCall, the callee slot).
func (vm *bytecodeVM) tryInlineCallFromStack(callee runtime.Value, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	fn, injectedReceiver, hasInjectedReceiver, ok := inlineCallFunctionValue(callee)
	if !ok {
		return nil, nil
	}
	return vm.tryInlineResolvedCallFromStack(fn, injectedReceiver, hasInjectedReceiver, argBase, argCount, truncateTo, callNode, currentProgram)
}

func (vm *bytecodeVM) tryInlineResolvedCallFromStack(fn *runtime.FunctionValue, injectedReceiver runtime.Value, hasInjectedReceiver bool, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if fn == nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissNoBytecode)
		}
		return nil, nil
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissNoBytecode)
		}
		return nil, nil
	}
	if prog.frameLayout == nil {
		return vm.tryInlineSlotlessResolvedCallFromStack(fn, prog, injectedReceiver, hasInjectedReceiver, argBase, argCount, truncateTo, callNode, currentProgram)
	}
	layout := prog.frameLayout
	paramCount := layout.paramSlots
	expectedArgs := paramCount
	sourceArgBase := argBase
	var implicitReceiver runtime.Value
	injectReceiverIntoSlot0 := hasInjectedReceiver && !layout.methodShorthand
	if layout.methodShorthand {
		if hasInjectedReceiver {
			implicitReceiver = injectedReceiver
		} else {
			implicitReceiver = vm.stackValue(argBase)
			expectedArgs++
			sourceArgBase++
		}
	} else if hasInjectedReceiver {
		expectedArgs--
		implicitReceiver = injectedReceiver
	} else if paramCount > 0 {
		implicitReceiver = vm.stackValue(argBase)
	}
	if expectedArgs < 0 || argCount != expectedArgs {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissArity)
		}
		return nil, nil
	}
	if bytecodeInlineSkipsGenericLambda(fn) {
		if vm.interp != nil {
			vm.interp.recordBytecodeInlineResolvedMiss(bytecodeInlineResolvedMissGenericLambda)
		}
		return nil, nil
	}
	localEnv, err := vm.inlineResolvedCallEnvForBindings(fn, prog, layout, injectedReceiver, hasInjectedReceiver, argBase, argCount, callNode)
	if err != nil {
		return nil, err
	}
	if localEnv == nil {
		localEnv = fn.Closure
	}
	localEnv = vm.bytecodeCalleeEnv(localEnv)
	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)

	if injectReceiverIntoSlot0 {
		// Bound receivers already passed member resolution for this callable,
		// so only the explicit arguments need inline coercion work here.
		vm.copyInlineCallArgToSlot(slots, 0, injectedReceiver)
		seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, 0, injectedReceiver)
		if !layout.anyExplicitCoercion {
			vm.inlineCopyArgsToSlots(slots[1:], vm.stackValues(sourceArgBase, sourceArgBase+argCount), argCount)
			for idx := 1; idx < paramCount; idx++ {
				seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, vm.stackValue(sourceArgBase+idx-1))
			}
		} else {
			for idx := 1; idx < paramCount; idx++ {
				arg := vm.stackValue(sourceArgBase + idx - 1)
				paramType := inlineParamType(layout, idx)
				if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
					if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, err
					} else if ok {
						arg = coerced
					} else {
						coerced, err := vm.interp.coerceValueToType(paramType, arg)
						if err != nil {
							vm.releaseSlotFrame(slots)
							vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
							return nil, err
						}
						arg = coerced
					}
				}
				vm.copyInlineCallArgToSlot(slots, idx, arg)
				seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, arg)
			}
		}
	} else if !layout.anyParamCoercion {
		vm.inlineCopyArgsToSlots(slots, vm.stackValues(sourceArgBase, sourceArgBase+paramCount), paramCount)
		for idx := 0; idx < paramCount; idx++ {
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, vm.stackValue(sourceArgBase+idx))
		}
	} else {
		for idx := 0; idx < paramCount; idx++ {
			arg := vm.stackValue(sourceArgBase + idx)
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, err
					}
					arg = coerced
				}
			}
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, arg)
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := layout.usesImplicitMember && (layout.methodShorthand || paramCount > 0)
	if hasImplicit {
		state := vm.interp.stateFromEnv(localEnv)
		state.pushImplicitReceiver(bytecodeStackSnapshotValue(implicitReceiver))
	}

	vm.truncateStack(truncateTo)
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, localEnv)
	returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, returnGenericNames, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = localEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
	return prog, nil
}

// tryInlineSelfCallFromStack is a tighter inline path for bytecodeOpCallSelf.
// It assumes a direct self-call function value (no bound-method injection).
func (vm *bytecodeVM) tryInlineSelfCallFromStack(fn *runtime.FunctionValue, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if argBase < 0 || argCount < 0 || argBase+argCount > vm.stackDepth() {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if truncateTo < 0 || truncateTo > argBase {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if fn == nil {
		return nil, nil
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil
	}
	layout := prog.frameLayout
	if layout.methodShorthand {
		return nil, nil
	}
	if argCount != layout.paramSlots {
		return nil, nil
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	if bytecodeFunctionNeedsCallLocalBindings(vm, fn) {
		return nil, nil
	}
	if layout.paramSlots == 1 && !layout.usesImplicitMember {
		arg := vm.stackValue(argBase)
		paramType := inlineParamType(layout, 0)
		if !inlineParamNeedsRuntimeCoercion(layout, 0, fn) || inlineParamCoercionUnnecessary(vm.interp, layout, 0, paramType, arg) {
			slots := vm.acquireSlotFrame(layout.slotCount)
			calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)
			vm.copyInlineCallArgToSlot(slots, 0, arg)
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, 0, arg)
			if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
				slots[layout.selfCallSlot] = fn
			}
			vm.truncateStack(truncateTo)
			returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
			vm.pushInlineSelfFastFrame(vm.ip+1, currentProgram, vm.slots, returnGenericNames, len(vm.iterStack), len(vm.loopStack), false)
			vm.slots = slots
			vm.prepareValueSlotI32Frame(prog)
			vm.env = vm.bytecodeCalleeEnv(fn.Closure)
			vm.ip = 0
			vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
			return prog, nil
		}
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)
	if !layout.anyParamCoercion {
		vm.inlineCopyArgsToSlots(slots, vm.stackValues(argBase, argBase+layout.paramSlots), layout.paramSlots)
		for idx := 0; idx < layout.paramSlots; idx++ {
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, vm.stackValue(argBase+idx))
		}
	} else {
		for idx := 0; idx < layout.paramSlots; idx++ {
			arg := vm.stackValue(argBase + idx)
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(vm.interp, layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						vm.releaseI32RegisterFrame(calleeI32Values, calleeI32Valid)
						return nil, err
					}
					arg = coerced
				}
			}
			vm.copyInlineCallArgToSlot(slots, idx, arg)
			seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, idx, arg)
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}
	calleeEnv := vm.bytecodeCalleeEnv(fn.Closure)

	hasImplicit := layout.paramSlots > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(calleeEnv)
		state.pushImplicitReceiver(bytecodeStackSnapshotValue(vm.stackValue(argBase)))
	}

	vm.truncateStack(truncateTo)
	returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
	vm.pushInlineSelfFastFrame(vm.ip+1, currentProgram, vm.slots, returnGenericNames, len(vm.iterStack), len(vm.loopStack), hasImplicit)
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = calleeEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
	return prog, nil
}

// tryInlineSelfCallWithArg is a no-stack inline setup path for self calls
// that already computed a single argument value.
func (vm *bytecodeVM) tryInlineSelfCallWithArg(fn *runtime.FunctionValue, arg runtime.Value, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if fn == nil {
		return nil, nil
	}
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil
	}
	layout := prog.frameLayout
	if !layout.selfCallOneArgFast {
		return nil, nil
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	if bytecodeFunctionNeedsCallLocalBindings(vm, fn) {
		return nil, nil
	}

	paramType := inlineParamType(layout, 0)
	if inlineParamNeedsRuntimeCoercion(layout, 0, fn) {
		noCoercion := false
		if len(layout.paramExactStructDef) > 0 && layout.paramExactStructDef[0] != nil {
			noCoercion = inlineExactNamedStructNoCoercionBytecodeExactDef(layout.paramExactStructDef[0], arg)
		} else if layout.firstParamSimple != "" {
			noCoercion = inlineCoercionUnnecessaryBySimpleTypeWithInterpreter(vm.interp, layout.firstParamSimple, arg)
		} else {
			noCoercion = inlineCoercionUnnecessaryWithInterpreter(vm.interp, paramType, arg)
		}
		if !noCoercion {
			if coerced, ok, err := inlineCoerceValueBySimpleType(layout.firstParamSimple, arg); err != nil {
				return nil, err
			} else if ok {
				arg = coerced
			} else {
				coerced, err := vm.interp.coerceValueToType(paramType, arg)
				if err != nil {
					return nil, err
				}
				arg = coerced
			}
		}
	}

	slots := vm.acquireSlotFrame(layout.slotCount)
	calleeI32Values, calleeI32Valid := vm.acquireInlineCalleeI32RegisterFrame(layout)
	vm.copyInlineCallArgToSlot(slots, 0, arg)
	seedInlineCalleeI32RegisterSlot(layout, calleeI32Values, calleeI32Valid, 0, arg)
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}
	calleeEnv := vm.bytecodeCalleeEnv(fn.Closure)

	hasImplicit := layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(calleeEnv)
		state.pushImplicitReceiver(bytecodeStackSnapshotValue(arg))
	}

	returnGenericNames := bytecodeInlineReturnGenericNames(fn, prog)
	vm.pushInlineSelfFastFrame(vm.ip+1, currentProgram, vm.slots, returnGenericNames, len(vm.iterStack), len(vm.loopStack), hasImplicit)
	vm.slots = slots
	vm.prepareValueSlotI32Frame(prog)
	vm.env = calleeEnv
	vm.ip = 0
	vm.installInlineCalleeI32RegisterFrame(prog, calleeI32Values, calleeI32Valid)
	return prog, nil
}

func subtractIntegerSameTypeFast(left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	lv, lok := left.ToInt64()
	rv, rok := right.ToInt64()
	if !lok || !rok {
		return nil, false, nil
	}
	diff, overflow := subInt64Overflow(lv, rv)
	if overflow {
		return nil, false, nil
	}
	if err := ensureFitsInt64Type(left.TypeSuffix, diff); err != nil {
		return nil, true, err
	}
	return bytecodeRawIntegerResultValue(left.TypeSuffix, diff), true, nil
}

func bytecodeSubtractIntegerImmediateFast(left runtime.Value, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if fast, handled, err := bytecodeSubtractIntegerImmediateI32Fast(left, right); handled {
		return fast, true, err
	}
	rightRef := &right
	if !rightRef.IsSmallRef() {
		return nil, false, nil
	}
	rightVal := rightRef.Int64FastRef()
	switch lv := left.(type) {
	case runtime.IntegerValue:
		lvRef := &lv
		if lvRef.IsSmallRef() {
			if lv.TypeSuffix != right.TypeSuffix {
				return nil, false, nil
			}
			diff, overflow := subInt64Overflow(lvRef.Int64FastRef(), rightVal)
			if overflow {
				return nil, false, nil
			}
			if err := ensureFitsInt64Type(lv.TypeSuffix, diff); err != nil {
				return nil, true, err
			}
			return bytecodeRawIntegerResultValue(lv.TypeSuffix, diff), true, nil
		}
	case *runtime.IntegerValue:
		if lv != nil && lv.IsSmallRef() {
			if lv.TypeSuffix != right.TypeSuffix {
				return nil, false, nil
			}
			diff, overflow := subInt64Overflow(lv.Int64FastRef(), rightVal)
			if overflow {
				return nil, false, nil
			}
			if err := ensureFitsInt64Type(lv.TypeSuffix, diff); err != nil {
				return nil, true, err
			}
			return bytecodeRawIntegerResultValue(lv.TypeSuffix, diff), true, nil
		}
	}
	return nil, false, nil
}

func addIntegerSameTypeFast(left runtime.IntegerValue, right runtime.IntegerValue) (runtime.Value, bool, error) {
	if left.TypeSuffix != right.TypeSuffix {
		return nil, false, nil
	}
	lv, lok := left.ToInt64()
	rv, rok := right.ToInt64()
	if !lok || !rok {
		return nil, false, nil
	}
	sum, overflow := addInt64Overflow(lv, rv)
	if overflow {
		return nil, false, nil
	}
	if err := ensureFitsInt64Type(left.TypeSuffix, sum); err != nil {
		return nil, true, err
	}
	return bytecodeRawIntegerResultValue(left.TypeSuffix, sum), true, nil
}

func nativeCallNeedsStableArgs(fn runtime.NativeFunctionValue) bool {
	return !fn.BorrowArgs
}

func bytecodeCallTargetNeedsStableArgs(callee runtime.Value) bool {
	switch v := callee.(type) {
	case runtime.NativeFunctionValue:
		return nativeCallNeedsStableArgs(v)
	case *runtime.NativeFunctionValue:
		if v == nil {
			return false
		}
		return nativeCallNeedsStableArgs(*v)
	case runtime.NativeBoundMethodValue:
		return nativeCallNeedsStableArgs(v.Method)
	case *runtime.NativeBoundMethodValue:
		if v == nil {
			return false
		}
		return nativeCallNeedsStableArgs(v.Method)
	case runtime.DynRefValue, *runtime.DynRefValue:
		return true
	case runtime.BoundMethodValue:
		switch method := v.Method.(type) {
		case runtime.NativeFunctionValue:
			return nativeCallNeedsStableArgs(method)
		case *runtime.NativeFunctionValue:
			if method == nil {
				return false
			}
			return nativeCallNeedsStableArgs(*method)
		case runtime.NativeBoundMethodValue:
			return nativeCallNeedsStableArgs(method.Method)
		case *runtime.NativeBoundMethodValue:
			if method == nil {
				return false
			}
			return nativeCallNeedsStableArgs(method.Method)
		case runtime.DynRefValue, *runtime.DynRefValue:
			return true
		}
		return false
	case *runtime.BoundMethodValue:
		if v == nil {
			return false
		}
		switch method := v.Method.(type) {
		case runtime.NativeFunctionValue:
			return nativeCallNeedsStableArgs(method)
		case *runtime.NativeFunctionValue:
			if method == nil {
				return false
			}
			return nativeCallNeedsStableArgs(*method)
		case runtime.NativeBoundMethodValue:
			return nativeCallNeedsStableArgs(method.Method)
		case *runtime.NativeBoundMethodValue:
			if method == nil {
				return false
			}
			return nativeCallNeedsStableArgs(method.Method)
		case runtime.DynRefValue, *runtime.DynRefValue:
			return true
		}
		return false
	case runtime.PartialFunctionValue:
		return bytecodeCallTargetNeedsStableArgs(v.Target)
	case *runtime.PartialFunctionValue:
		if v == nil {
			return false
		}
		return bytecodeCallTargetNeedsStableArgs(v.Target)
	default:
		return false
	}
}

func copyCallArgs(args []runtime.Value) []runtime.Value {
	if len(args) == 0 {
		return args
	}
	cloned := make([]runtime.Value, len(args))
	copy(cloned, args)
	return cloned
}

func bytecodeCallArgsNeedMaterialization(args []runtime.Value) bool {
	for _, arg := range args {
		if bytecodeIsRawIntegerCarrier(arg) {
			return true
		}
		if _, _, ok := bytecodeDirectRawFloatValue(arg); ok {
			return true
		}
	}
	return false
}

func bytecodePrepareCallArgs(args []runtime.Value, needsStableCopy bool) []runtime.Value {
	if len(args) == 0 {
		return args
	}
	if needsStableCopy {
		args = copyCallArgs(args)
	}
	bytecodeMaterializeRawFloatArgs(args)
	return args
}

func (vm *bytecodeVM) prepareMaterializedCallArgs(args []runtime.Value, needsStableCopy bool, class, reason string) []runtime.Value {
	vm.recordPrimitiveMaterializationValues(class, reason, args)
	return bytecodePrepareCallArgs(args, needsStableCopy)
}

func bytecodeMaterializedCallArgs(args []runtime.Value) []runtime.Value {
	if len(args) == 0 {
		return args
	}
	cloned := copyCallArgs(args)
	bytecodeMaterializeRawFloatArgs(cloned)
	return cloned
}

func (vm *bytecodeVM) copyMaterializedCallArgs(dst []runtime.Value, src []runtime.Value, class, reason string) {
	vm.recordPrimitiveMaterializationValues(class, reason, src)
	bytecodeCopyMaterializedCallArgs(dst, src)
}

func bytecodeCopyMaterializedCallArgs(dst []runtime.Value, src []runtime.Value) {
	for idx, arg := range src {
		dst[idx] = bytecodeMaterializeRawValue(arg)
	}
}

func bytecodeMaterializeRawFloatArgs(args []runtime.Value) {
	for idx, arg := range args {
		args[idx] = bytecodeMaterializeRawValue(arg)
	}
}

// execCall handles bytecodeOpCall. It returns a non-nil program when an
// inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCall(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if vm.stackDepth() < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	argBase := vm.stackDepth() - instr.argCount
	calleeIndex := argBase - 1
	callee := vm.stackValue(calleeIndex)
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		args := vm.stackValuesFrom(argBase)
		vm.truncateStack(calleeIndex)
		return vm.execAndFinishExactNativeCall(target, args, callNode)
	}
	// Fast path: inline without allocating an argument slice.
	if newProg, err := vm.tryInlineCallFromStack(callee, argBase, instr.argCount, calleeIndex, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		return newProg, nil
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	if result, handled, err := vm.tryCallDirectFunctionValueFromStack(callee, argBase, instr.argCount, calleeIndex, callNode); handled || err != nil {
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(calleeIndex)
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(callee)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
	}
	// Normal call.
	result, err := vm.callCallableValueMutable(callee, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

// execCallSelf handles bytecodeOpCallSelf for self-recursive slot calls.
// The callee is read from instr.target in the active slot frame.
func (vm *bytecodeVM) execCallSelf(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if vm.stackDepth() < instr.argCount {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode self call slot out of range")
	}
	callee := vm.slots[instr.target]
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	argBase := vm.stackDepth() - instr.argCount
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	// Fast path: inline without allocating an argument slice.
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		if newProg, err := vm.tryInlineSelfCallFromStack(fn, argBase, instr.argCount, argBase, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if statsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		} else if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	default:
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	}
	args := vm.stackValuesFrom(argBase)
	vm.truncateStack(argBase)
	if result, handled, err := vm.tryExecExactNativeCall(callee, args, callNode); handled {
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	needsStableArgsCopy := bytecodeCallTargetNeedsStableArgs(callee)
	if needsStableArgsCopy {
		var inline [bytecodeInlinePreparedCallArgStorage]runtime.Value
		if prepared, ok := bytecodePrepareCallArgsIntoBuffer(inline[:], args, true); ok {
			args = prepared
		} else {
			args = vm.prepareMaterializedCallArgs(args, true, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
		}
	} else {
		args = vm.prepareMaterializedCallArgs(args, false, bytecodeMaterializationCandidateStatic, bytecodeMaterializationReasonStaticCall)
	}
	result, err := vm.callCallableValueMutable(callee, args, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) pushCallNameSlotArgs(instr bytecodeInstruction) error {
	if instr.argCount < 0 || instr.argCount > 3 {
		return fmt.Errorf("bytecode slot-arg call count invalid")
	}
	argSlots := [3]int{instr.target, instr.loopBreak, instr.loopContinue}
	for idx := 0; idx < instr.argCount; idx++ {
		slot := argSlots[idx]
		if slot < 0 || slot >= len(vm.slots) {
			return fmt.Errorf("bytecode slot-arg call slot out of range")
		}
		vm.appendStackValue(vm.slotRuntimeValue(slot))
	}
	return nil
}
