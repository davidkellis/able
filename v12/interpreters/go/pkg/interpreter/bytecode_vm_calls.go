package interpreter

import (
	"errors"
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
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		return nil, nil
	}
	// Method shorthand requires implicit receiver adjustment and extra arg.
	if decl.IsMethodShorthand {
		return nil, nil
	}
	// Require exact arity match (skip optional params).
	paramCount := len(decl.Params)
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

	layout := prog.frameLayout
	slots := vm.acquireSlotFrame(layout.slotCount)

	// Coerce parameters (skip for params that use generic type names,
	// matching the behavior in invokeFunction). Use fast path when the
	// value already trivially matches the declared type.
	var generics map[string]struct{}
	if len(decl.GenericParams) > 0 {
		generics = functionGenericNameSet(fn, decl)
	}
	for idx, param := range decl.Params {
		arg := args[idx]
		if param.ParamType != nil && !paramUsesGeneric(param.ParamType, generics) && !inlineCoercionUnnecessary(param.ParamType, arg) {
			coerced, err := vm.interp.coerceValueToType(param.ParamType, arg)
			if err != nil {
				vm.releaseSlotFrame(slots)
				return nil, err
			}
			arg = coerced
		}
		slots[idx] = arg
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
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)

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
	prog, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || prog == nil || prog.frameLayout == nil {
		return nil, nil
	}
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		return nil, nil
	}
	if decl.IsMethodShorthand {
		return nil, nil
	}
	paramCount := len(decl.Params)
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
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}

	layout := prog.frameLayout
	slots := vm.acquireSlotFrame(layout.slotCount)
	var generics map[string]struct{}
	if len(decl.GenericParams) > 0 {
		generics = functionGenericNameSet(fn, decl)
	}
	for idx, param := range decl.Params {
		var arg runtime.Value
		if hasInjectedReceiver {
			if idx == 0 {
				arg = injectedReceiver
			} else {
				arg = vm.stack[argBase+idx-1]
			}
		} else {
			arg = vm.stack[argBase+idx]
		}
		if param.ParamType != nil && !paramUsesGeneric(param.ParamType, generics) && !inlineCoercionUnnecessary(param.ParamType, arg) {
			coerced, err := vm.interp.coerceValueToType(param.ParamType, arg)
			if err != nil {
				vm.releaseSlotFrame(slots)
				return nil, err
			}
			arg = coerced
		}
		slots[idx] = arg
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
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), hasImplicit, selfFast)
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
	decl, ok := fn.Declaration.(*ast.FunctionDefinition)
	if !ok || decl == nil || decl.Body == nil {
		return nil, nil
	}
	if decl.IsMethodShorthand {
		return nil, nil
	}
	if argCount != len(decl.Params) {
		return nil, nil
	}
	if callNode != nil && len(callNode.TypeArguments) > 0 {
		return nil, nil
	}
	if fn.MethodSet != nil && len(fn.MethodSet.GenericParams) > 0 {
		return nil, nil
	}

	layout := prog.frameLayout
	if len(decl.Params) == 1 && len(decl.GenericParams) == 0 && !layout.usesImplicitMember {
		param := decl.Params[0]
		arg := vm.stack[argBase]
		if param != nil && (param.ParamType == nil || inlineCoercionUnnecessary(param.ParamType, arg)) {
			slots := vm.acquireSlotFrame(layout.slotCount)
			slots[0] = arg
			if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
				slots[layout.selfCallSlot] = fn
			}
			vm.stack = vm.stack[:truncateTo]
			vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), false, true)
			vm.slots = slots
			vm.env = fn.Closure
			vm.ip = 0
			return prog, nil
		}
	}
	slots := vm.acquireSlotFrame(layout.slotCount)
	var generics map[string]struct{}
	if len(decl.GenericParams) > 0 {
		generics = functionGenericNameSet(fn, decl)
	}
	for idx, param := range decl.Params {
		arg := vm.stack[argBase+idx]
		if param.ParamType != nil && !paramUsesGeneric(param.ParamType, generics) && !inlineCoercionUnnecessary(param.ParamType, arg) {
			coerced, err := vm.interp.coerceValueToType(param.ParamType, arg)
			if err != nil {
				vm.releaseSlotFrame(slots)
				return nil, err
			}
			arg = coerced
		}
		slots[idx] = arg
	}
	if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
		slots[layout.selfCallSlot] = fn
	}

	hasImplicit := len(decl.Params) > 0 && layout.usesImplicitMember
	if hasImplicit {
		state := vm.interp.stateFromEnv(fn.Closure)
		state.pushImplicitReceiver(vm.stack[argBase])
	}

	vm.stack = vm.stack[:truncateTo]
	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), hasImplicit, true)
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

	if layout.firstParamType != nil {
		noCoercion := false
		if layout.firstParamSimple != "" {
			noCoercion = inlineCoercionUnnecessaryBySimpleType(layout.firstParamSimple, arg)
		} else {
			noCoercion = inlineCoercionUnnecessary(layout.firstParamType, arg)
		}
		if !noCoercion {
			coerced, err := vm.interp.coerceValueToType(layout.firstParamType, arg)
			if err != nil {
				return nil, err
			}
			arg = coerced
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

	vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), hasImplicit, true)
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

func (vm *bytecodeVM) callSelfIntSubSlotConstArg(instr *bytecodeInstruction, right runtime.IntegerValue, hasImmediate bool) (runtime.Value, error) {
	if instr.argCount < 0 || instr.argCount >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode slot out of range")
	}
	if !hasImmediate {
		return nil, fmt.Errorf("bytecode self call immediate must be integer")
	}
	left := vm.slots[instr.argCount]
	switch lv := left.(type) {
	case runtime.IntegerValue:
		if fast, handled, err := subtractIntegerSameTypeFast(lv, right); handled {
			return fast, err
		}
		return evaluateIntegerArithmeticFast("-", lv, right)
	case *runtime.IntegerValue:
		if lv != nil {
			if fast, handled, err := subtractIntegerSameTypeFast(*lv, right); handled {
				return fast, err
			}
			return evaluateIntegerArithmeticFast("-", *lv, right)
		}
	default:
		if leftInt, ok := bytecodeIntegerValue(left); ok {
			return evaluateIntegerArithmeticFast("-", leftInt, right)
		}
	}
	return applyBinaryOperator(vm.interp, "-", left, right)
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
	vm.stack = vm.stack[:argBase]
	callee, err := vm.pop()
	if err != nil {
		return nil, err
	}
	if bytecodeCallTargetNeedsStableArgs(callee) {
		args = copyCallArgs(args)
	}
	// Normal call.
	result, err := vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil, nil
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
	if bytecodeCallTargetNeedsStableArgs(callee) {
		args = copyCallArgs(args)
	}
	result, err := vm.interp.callCallableValueMutable(callee, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil, nil
}

// execCallSelfIntSubSlotConst handles fused recursive calls of the form
// self(slot - const), computing the argument directly from slot/immediate.
func (vm *bytecodeVM) execCallSelfIntSubSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode self call slot out of range")
	}
	rightImmediate, hasImmediate := bytecodeImmediateIntegerValue(instr.value)
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	callee := vm.slots[instr.target]
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		calleeProgram, _ := fn.Bytecode.(*bytecodeProgram)
		if instr.argCount >= 0 && instr.argCount < len(vm.slots) && currentProgram != nil && calleeProgram == currentProgram && currentProgram.frameLayout != nil && currentProgram.frameLayout.selfCallOneArgFast {
			if hasImmediate {
				right := rightImmediate
				var leftInt runtime.IntegerValue
				leftIntOK := false
				switch lv := vm.slots[instr.argCount].(type) {
				case runtime.IntegerValue:
					leftInt = lv
					leftIntOK = true
				case *runtime.IntegerValue:
					if lv != nil {
						leftInt = *lv
						leftIntOK = true
					}
				}
				if leftIntOK {
					if arg, handled, argErr := subtractIntegerSameTypeFast(leftInt, right); handled {
						if argErr != nil {
							argErr = vm.interp.wrapStandardRuntimeError(argErr)
							if instr.node != nil {
								argErr = vm.interp.attachRuntimeContext(argErr, instr.node, vm.interp.stateFromEnv(vm.env))
							}
							return nil, argErr
						}
						layout := currentProgram.frameLayout
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
						vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, len(vm.iterStack), len(vm.loopStack), hasImplicit, true)
						vm.slots = slots
						vm.env = fn.Closure
						vm.ip = 0
						if statsEnabled {
							vm.interp.recordBytecodeInlineCallHit()
						}
						return calleeProgram, nil
					}
				}
			}
		}

		arg, argErr := vm.callSelfIntSubSlotConstArg(instr, rightImmediate, hasImmediate)
		if argErr != nil {
			argErr = vm.interp.wrapStandardRuntimeError(argErr)
			if instr.node != nil {
				argErr = vm.interp.attachRuntimeContext(argErr, instr.node, vm.interp.stateFromEnv(vm.env))
			}
			return nil, argErr
		}
		if newProg, err := vm.tryInlineSelfCallWithArg(fn, arg, callNode, currentProgram); err != nil {
			return nil, err
		} else if newProg != nil {
			if statsEnabled {
				vm.interp.recordBytecodeInlineCallHit()
			}
			return newProg, nil
		} else if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
		args := [1]runtime.Value{arg}
		result, err := vm.interp.callCallableValueMutable(callee, args[:], vm.env, callNode)
		if err != nil {
			if errors.Is(err, errSerialYield) {
				payload := payloadFromState(vm.env.RuntimeData())
				if payload == nil || !payload.awaitBlocked {
					vm.stack = append(vm.stack, runtime.NilValue{})
					vm.ip++
				}
				return nil, err
			}
			err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
			if vm.handleLoopSignal(err) {
				return nil, nil
			}
			return nil, err
		}
		if result == nil {
			result = runtime.NilValue{}
		}
		vm.stack = append(vm.stack, result)
		vm.ip++
		return nil, nil
	default:
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	}

	arg, argErr := vm.callSelfIntSubSlotConstArg(instr, rightImmediate, hasImmediate)
	if argErr != nil {
		argErr = vm.interp.wrapStandardRuntimeError(argErr)
		if instr.node != nil {
			argErr = vm.interp.attachRuntimeContext(argErr, instr.node, vm.interp.stateFromEnv(vm.env))
		}
		return nil, argErr
	}

	args := [1]runtime.Value{arg}
	result, err := vm.interp.callCallableValueMutable(callee, args[:], vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, vm.interp.stateFromEnv(vm.env))
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil, nil
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
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	state := vm.interp.stateFromEnv(vm.env)
	if statsEnabled {
		vm.interp.recordBytecodeCallNameLookup()
	}
	calleeVal, found := vm.lookupCachedName(currentProgram, vm.ip, instr.name)
	if !found {
		if dotIdx := strings.Index(instr.name, "."); dotIdx > 0 && dotIdx < len(instr.name)-1 {
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
					return nil, vm.interp.attachRuntimeContext(err, callNode, state)
				}
				calleeVal = candidate
				vm.storeCachedMemberMethod(currentProgram, vm.ip, tail, true, receiver, candidate)
			}
		} else {
			err := fmt.Errorf("Undefined variable '%s'", instr.name)
			return nil, vm.interp.attachRuntimeContext(err, callNode, state)
		}
	}
	argBase := len(vm.stack) - instr.argCount
	// Fast path: inline without allocating an argument slice.
	if newProg, err := vm.tryInlineCallFromStack(calleeVal, argBase, instr.argCount, argBase, callNode, currentProgram); err != nil {
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
	vm.stack = vm.stack[:argBase]
	if bytecodeCallTargetNeedsStableArgs(calleeVal) {
		args = copyCallArgs(args)
	}
	// Normal call.
	result, err := vm.interp.callCallableValueMutable(calleeVal, args, vm.env, callNode)
	if err != nil {
		if errors.Is(err, errSerialYield) {
			payload := payloadFromState(vm.env.RuntimeData())
			if payload == nil || !payload.awaitBlocked {
				vm.stack = append(vm.stack, runtime.NilValue{})
				vm.ip++
			}
			return nil, err
		}
		err = vm.interp.attachRuntimeContext(err, callNode, state)
		if vm.handleLoopSignal(err) {
			return nil, nil
		}
		return nil, err
	}
	if result == nil {
		result = runtime.NilValue{}
	}
	vm.stack = append(vm.stack, result)
	vm.ip++
	return nil, nil
}
