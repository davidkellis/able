package interpreter

import (
	"fmt"
	"strings"

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

func inlineParamCoercionUnnecessary(layout *bytecodeFrameLayout, idx int, typeExpr ast.TypeExpression, val runtime.Value) bool {
	if layout != nil && idx >= 0 && idx < len(layout.paramSimpleTypes) {
		if typeName := layout.paramSimpleTypes[idx]; typeName != "" {
			return inlineCoercionUnnecessaryBySimpleType(typeName, val)
		}
	}
	return inlineCoercionUnnecessary(typeExpr, val)
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

func inlineCopyArgsToSlots(dst []runtime.Value, src []runtime.Value, count int) {
	if count <= 0 {
		return
	}
	copy(dst[:count], src[:count])
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
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}

	slots := vm.acquireSlotFrame(layout.slotCount)

	// Coerce parameters when the cached layout metadata says the declared type
	// can actually require runtime coercion.
	if !layout.anyParamCoercion {
		inlineCopyArgsToSlots(slots, args, paramCount)
	} else {
		for idx := 0; idx < paramCount; idx++ {
			arg := args[idx]
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						return nil, err
					}
					arg = coerced
				}
			}
			slots[idx] = arg
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	// Push implicit receiver only when the function body uses #member syntax.
	hasImplicit := paramCount > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(args[0])
	}

	// Push call frame.
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, fn.Closure)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, bytecodeInlineReturnGenericNames(fn, prog), len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)

	// Set up new frame.
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0

	return prog, nil
}

// tryInlineCallFromStack mirrors tryInlineCall but reads arguments directly
// from vm.stack[argBase:argBase+argCount]. On success it truncates the stack
// to truncateTo (dropping arguments and, for bytecodeOpCall, the callee slot).
func (vm *bytecodeVM) tryInlineCallFromStack(callee runtime.Value, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if argBase < 0 || argCount < 0 || argBase+argCount > len(vm.stack) {
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
	if argBase < 0 || argCount < 0 || argBase+argCount > len(vm.stack) {
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
	paramCount := layout.paramSlots
	expectedArgs := paramCount
	if hasInjectedReceiver {
		expectedArgs--
	}
	if expectedArgs < 0 || argCount != expectedArgs {
		return nil, nil
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 && !hasInjectedReceiver {
		return nil, nil
	}
	slots := vm.acquireSlotFrame(layout.slotCount)

	if hasInjectedReceiver {
		// Bound receivers already passed member resolution for this callable,
		// so only the explicit arguments need inline coercion work here.
		slots[0] = injectedReceiver
		if !layout.anyExplicitCoercion {
			inlineCopyArgsToSlots(slots[1:], vm.stack[argBase:argBase+argCount], argCount)
		} else {
			for idx := 1; idx < paramCount; idx++ {
				arg := vm.stack[argBase+idx-1]
				paramType := inlineParamType(layout, idx)
				if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(layout, idx, paramType, arg) {
					if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
						vm.releaseSlotFrame(slots)
						return nil, err
					} else if ok {
						arg = coerced
					} else {
						coerced, err := vm.interp.coerceValueToType(paramType, arg)
						if err != nil {
							vm.releaseSlotFrame(slots)
							return nil, err
						}
						arg = coerced
					}
				}
				slots[idx] = arg
			}
		}
	} else if !layout.anyParamCoercion {
		inlineCopyArgsToSlots(slots, vm.stack[argBase:argBase+paramCount], paramCount)
	} else {
		for idx := 0; idx < paramCount; idx++ {
			arg := vm.stack[argBase+idx]
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						return nil, err
					}
					arg = coerced
				}
			}
			slots[idx] = arg
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := paramCount > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		implicitReceiver := vm.stack[argBase]
		if hasInjectedReceiver {
			implicitReceiver = injectedReceiver
		}
		state.pushImplicitReceiver(implicitReceiver)
	}

	vm.stack = vm.stack[:truncateTo]
	selfFast := bytecodeCanUseSelfFastFrame(currentProgram, prog, vm.env, fn.Closure)
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, bytecodeInlineReturnGenericNames(fn, prog), len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0
	return prog, nil
}

// tryInlineSelfCallFromStack is a tighter inline path for bytecodeOpCallSelf.
// It assumes a direct self-call function value (no bound-method injection).
func (vm *bytecodeVM) tryInlineSelfCallFromStack(fn *runtime.FunctionValue, argBase int, argCount int, truncateTo int, callNode *ast.FunctionCall, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if argBase < 0 || argCount < 0 || argBase+argCount > len(vm.stack) {
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
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}
	if layout.paramSlots == 1 && !layout.usesImplicitMember {
		arg := vm.stack[argBase]
		paramType := inlineParamType(layout, 0)
		if !inlineParamNeedsRuntimeCoercion(layout, 0, fn) || inlineParamCoercionUnnecessary(layout, 0, paramType, arg) {
			slots := vm.acquireSlotFrame(layout.slotCount)
			slots[0] = arg
			if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
				slots[layout.selfCallSlot] = fn
			}
			vm.stack = vm.stack[:truncateTo]
			vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, bytecodeInlineReturnGenericNames(fn, prog), len(vm.iterStack), len(vm.loopStack), false, true)
			vm.slots = slots
			vm.env = fn.Closure
			vm.ip = 0
			return prog, nil
		}
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	if !layout.anyParamCoercion {
		inlineCopyArgsToSlots(slots, vm.stack[argBase:argBase+layout.paramSlots], layout.paramSlots)
	} else {
		for idx := 0; idx < layout.paramSlots; idx++ {
			arg := vm.stack[argBase+idx]
			paramType := inlineParamType(layout, idx)
			if inlineParamNeedsRuntimeCoercion(layout, idx, fn) && !inlineParamCoercionUnnecessary(layout, idx, paramType, arg) {
				if coerced, ok, err := inlineCoerceValueBySimpleType(inlineParamSimpleType(layout, idx), arg); err != nil {
					vm.releaseSlotFrame(slots)
					return nil, err
				} else if ok {
					arg = coerced
				} else {
					coerced, err := vm.interp.coerceValueToType(paramType, arg)
					if err != nil {
						vm.releaseSlotFrame(slots)
						return nil, err
					}
					arg = coerced
				}
			}
			slots[idx] = arg
		}
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := layout.paramSlots > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(vm.stack[argBase])
	}

	vm.stack = vm.stack[:truncateTo]
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, bytecodeInlineReturnGenericNames(fn, prog), len(vm.iterStack), len(vm.loopStack), hasImplicit, true)
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0
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
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}

	paramType := inlineParamType(layout, 0)
	if inlineParamNeedsRuntimeCoercion(layout, 0, fn) {
		noCoercion := false
		if layout.firstParamSimple != "" {
			noCoercion = inlineCoercionUnnecessaryBySimpleType(layout.firstParamSimple, arg)
		} else {
			noCoercion = inlineCoercionUnnecessary(paramType, arg)
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
	slots[0] = arg
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(arg)
	}

	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, bytecodeInlineReturnGenericNames(fn, prog), len(vm.iterStack), len(vm.loopStack), hasImplicit, true)
	vm.slots = slots
	vm.env = fn.Closure
	vm.ip = 0
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
	if boxed, ok := bytecodeBoxedIntegerValue(left.TypeSuffix, diff); ok {
		return boxed, true, nil
	}
	return runtime.NewSmallInt(diff, left.TypeSuffix), true, nil
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
			return boxedOrSmallIntegerValue(lv.TypeSuffix, diff), true, nil
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
			return boxedOrSmallIntegerValue(lv.TypeSuffix, diff), true, nil
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
	if boxed, ok := bytecodeBoxedIntegerValue(left.TypeSuffix, sum); ok {
		return boxed, true, nil
	}
	return runtime.NewSmallInt(sum, left.TypeSuffix), true, nil
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

// execCall handles bytecodeOpCall. It returns a non-nil program when an
// inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCall(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if len(vm.stack) < instr.argCount+1 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	argBase := len(vm.stack) - instr.argCount
	calleeIndex := argBase - 1
	callee := vm.stack[calleeIndex]
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if target, ok := bytecodeResolveExactNativeCallTarget(callee, instr.argCount); ok {
		args := vm.stack[argBase:]
		vm.stack = vm.stack[:calleeIndex]
		result, _, err := vm.execExactNativeCall(target, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
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
	args := vm.stack[argBase:]
	vm.stack = vm.stack[:calleeIndex]
	if bytecodeCallTargetNeedsStableArgs(callee) {
		args = copyCallArgs(args)
	}
	// Normal call.
	result, err := vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

// execCallSelf handles bytecodeOpCallSelf for self-recursive slot calls.
// The callee is read from instr.target in the active slot frame.
func (vm *bytecodeVM) execCallSelf(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if len(vm.stack) < instr.argCount {
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
	argBase := len(vm.stack) - instr.argCount
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
	args := vm.stack[argBase:]
	vm.stack = vm.stack[:argBase]
	if result, handled, err := vm.tryExecExactNativeCall(callee, args, callNode); handled {
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	if bytecodeCallTargetNeedsStableArgs(callee) {
		args = copyCallArgs(args)
	}
	result, err := vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}

// execCallName handles bytecodeOpCallName. It returns a non-nil program when
// an inline call frame was set up (the caller must switch to the new program).
// A nil program with nil error means the call completed normally.
func (vm *bytecodeVM) execCallName(instr bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.argCount < 0 {
		return nil, fmt.Errorf("bytecode call arg count invalid")
	}
	if len(vm.stack) < instr.argCount {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	if instr.name == "" {
		return nil, fmt.Errorf("bytecode call missing target name")
	}
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	traceNode := instr.node
	if callNode != nil {
		traceNode = callNode
	}
	traceLookup := "name"
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if statsEnabled {
		vm.interp.recordBytecodeCallNameLookup()
	}
	argBase := len(vm.stack) - instr.argCount
	if instr.nameSimple {
		if cached, ok := vm.lookupCachedCallName(currentProgram, vm.ip, instr.name); ok {
			return vm.execCachedCallName(cached, argBase, instr.argCount, callNode, currentProgram)
		}
	}
	var (
		calleeVal runtime.Value
		found     bool
		lookup    bytecodeResolvedIdentifierLookup
	)
	if instr.nameSimple {
		lookup, found = vm.lookupCachedIdentifierNameEntry(currentProgram, vm.ip, instr.name)
		calleeVal = lookup.value
	} else {
		calleeVal, found = vm.lookupCachedName(currentProgram, vm.ip, instr.name)
	}
	if !found {
		if !instr.nameSimple {
			dotIdx := strings.Index(instr.name, ".")
			if dotIdx <= 0 || dotIdx >= len(instr.name)-1 {
				err := fmt.Errorf("Undefined variable '%s'", instr.name)
				return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
			}
			traceLookup = "dot_fallback"
			if statsEnabled {
				vm.interp.recordBytecodeCallNameDotFallback()
			}
			head := instr.name[:dotIdx]
			tail := instr.name[dotIdx+1:]
			receiver, recvFound := vm.lookupCachedName(currentProgram, vm.ip, head)
			if !recvFound {
				if def, ok := vm.env.StructDefinition(head); ok {
					receiver = def
				} else {
					receiver = runtime.TypeRefValue{TypeName: head}
				}
			}
			if cached, ok := vm.lookupCachedMemberMethod(currentProgram, vm.ip, tail, true, receiver); ok {
				calleeVal = cached
			} else {
				member := ast.ID(tail)
				candidate, err := vm.interp.memberAccessOnValueWithOptions(receiver, member, vm.env, true)
				if err != nil {
					return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
				}
				calleeVal = candidate
				vm.storeCachedMemberMethod(currentProgram, vm.ip, tail, true, receiver, candidate)
			}
		} else {
			err := fmt.Errorf("Undefined variable '%s'", instr.name)
			return nil, vm.attachBytecodeRuntimeContext(err, callNode, nil)
		}
	}
	if instr.nameSimple && found {
		entry := bytecodeBuildCallNameCacheEntry(instr.name, lookup, calleeVal, instr.argCount)
		if cached := vm.storeCachedCallName(currentProgram, vm.ip, entry); cached != nil {
			return vm.execCachedCallName(cached, argBase, instr.argCount, callNode, currentProgram)
		}
		return nil, fmt.Errorf("bytecode call-name cache store failed")
	}
	if target, ok := bytecodeResolveExactNativeCallTarget(calleeVal, instr.argCount); ok {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "exact_native", traceNode)
		}
		args := vm.stack[argBase:]
		vm.stack = vm.stack[:argBase]
		result, _, err := vm.execExactNativeCall(target, args, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	// Fast path: inline without allocating an argument slice.
	if newProg, err := vm.tryInlineCallFromStack(calleeVal, argBase, instr.argCount, argBase, callNode, currentProgram); err != nil {
		return nil, err
	} else if newProg != nil {
		if vm.interp != nil {
			vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "inline", traceNode)
		}
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallHit()
		}
		return newProg, nil
	} else if statsEnabled {
		vm.interp.recordBytecodeInlineCallMiss()
	}
	args := vm.stack[argBase:]
	vm.stack = vm.stack[:argBase]
	if bytecodeCallTargetNeedsStableArgs(calleeVal) {
		args = copyCallArgs(args)
	}
	// Normal call.
	if vm.interp != nil {
		vm.interp.recordBytecodeCallTrace("call_name", instr.name, traceLookup, "generic", traceNode)
	}
	result, err := vm.interp.callCallableValueMutable(calleeVal, args, vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
