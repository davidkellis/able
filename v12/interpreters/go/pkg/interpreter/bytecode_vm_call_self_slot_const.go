package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) callSelfIntSubSlotConstArg(instr *bytecodeInstruction, right runtime.IntegerValue, hasImmediate bool) (runtime.Value, error) {
	if instr.argCount < 0 || instr.argCount >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode slot out of range")
	}
	if !hasImmediate {
		return nil, fmt.Errorf("bytecode self call immediate must be integer")
	}
	left := vm.slots[instr.argCount]
	if fast, handled, err := bytecodeSubtractIntegerImmediateFast(left, right); handled {
		return fast, err
	}
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

func (vm *bytecodeVM) execCallSelfIntSubSlotConstCompact(instr *bytecodeInstruction, currentProgram *bytecodeProgram) (*bytecodeProgram, bool, error) {
	if instr == nil || currentProgram == nil || !instr.hasIntRaw || !instr.hasIntImmediate || instr.target != 1 || instr.argCount != 0 {
		return nil, false, nil
	}
	if len(vm.slots) < 2 || len(vm.iterStack) != 0 || len(vm.loopStack) != 0 {
		return nil, false, nil
	}
	layout := currentProgram.frameLayout
	if layout == nil || !layout.selfCallOneArgFast || layout.slotCount != 2 || layout.selfCallSlot != 1 || layout.usesImplicitMember {
		return nil, false, nil
	}
	if !currentProgram.returnGenericNamesCached || currentProgram.returnGenericNames != nil {
		return nil, false, nil
	}
	fn, ok := vm.slots[1].(*runtime.FunctionValue)
	if !ok || fn == nil || fn.Bytecode != currentProgram {
		return nil, false, nil
	}

	var diff int64
	if vm.selfFastSlot0I32Valid {
		diff = int64(vm.selfFastSlot0I32Raw) - instr.intImmediateRaw
	} else {
		switch left := vm.slots[0].(type) {
		case runtime.IntegerValue:
			if left.TypeSuffix != runtime.IntegerI32 {
				return nil, false, nil
			}
			leftRef := &left
			if !leftRef.IsSmallRef() {
				return nil, false, nil
			}
			diff = leftRef.Int64FastRef() - instr.intImmediateRaw
		case *runtime.IntegerValue:
			if left == nil || left.TypeSuffix != runtime.IntegerI32 || !left.IsSmallRef() {
				return nil, false, nil
			}
			diff = left.Int64FastRef() - instr.intImmediateRaw
		default:
			return nil, false, nil
		}
	}
	if diff < math.MinInt32 || diff > math.MaxInt32 {
		err := vm.interp.wrapStandardRuntimeError(newOverflowError("integer overflow"))
		if instr.node != nil {
			err = vm.attachBytecodeRuntimeContext(err, instr.node, nil)
		}
		return nil, true, err
	}
	if cap(vm.selfFastMinimal) == 0 {
		vm.selfFastMinimal = make([]bytecodeSelfFastMinimalCallFrame, 0, 32)
	}
	idx := len(vm.selfFastMinimal)
	if idx < cap(vm.selfFastMinimal) {
		vm.selfFastMinimal = vm.selfFastMinimal[:idx+1]
	} else {
		vm.selfFastMinimal = append(vm.selfFastMinimal, bytecodeSelfFastMinimalCallFrame{})
	}
	frame := &vm.selfFastMinimal[idx]
	frame.returnIP = vm.ip + 1
	frame.slots = vm.slots
	frame.slot0 = vm.slots[0]
	vm.saveSelfFastSlot0I32(frame)
	frame.reusesSlots = true
	vm.selfFastMinimalSuffix++
	vm.slots[0] = bytecodeBoxedIntegerI32Value(diff)
	vm.setSelfFastSlot0I32Raw(int32(diff))
	vm.env = fn.Closure
	vm.ip = 0
	if vm.interp != nil && vm.interp.bytecodeStatsEnabled {
		vm.interp.recordBytecodeInlineCallHit()
	}
	return currentProgram, true, nil
}

// execCallSelfIntSubSlotConst handles fused recursive calls of the form
// self(slot - const), computing the argument directly from slot/immediate.
func (vm *bytecodeVM) execCallSelfIntSubSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
	if newProgram, handled, err := vm.execCallSelfIntSubSlotConstCompact(instr, currentProgram); handled || err != nil {
		return newProgram, err
	}
	vm.clearSelfFastSlot0I32()
	if instr.target < 0 || instr.target >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode self call slot out of range")
	}
	rightImmediate, hasImmediate := instr.intImmediate, instr.hasIntImmediate
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeImmediateIntegerValue(instr.value)
	}
	if !hasImmediate {
		rightImmediate, hasImmediate = bytecodeSlotConstImmediateAtIP(vm.ip, slotConstIntImmTable)
	}
	statsEnabled := vm.interp != nil && vm.interp.bytecodeStatsEnabled
	if currentProgram != nil && hasImmediate {
		if layout := currentProgram.frameLayout; layout != nil && layout.selfCallOneArgFast && instr.argCount >= 0 && instr.argCount < len(vm.slots) && layout.selfCallSlot == instr.target {
			if fn, ok := vm.slots[instr.target].(*runtime.FunctionValue); ok && fn != nil && fn.Bytecode == currentProgram {
				var (
					arg     runtime.Value
					handled bool
					argErr  error
				)
				if instr.hasIntRaw {
					switch lv := vm.slots[instr.argCount].(type) {
					case runtime.IntegerValue:
						if lv.TypeSuffix == runtime.IntegerI32 {
							lvRef := &lv
							if lvRef.IsSmallRef() {
								diff := lvRef.Int64FastRef() - instr.intImmediateRaw
								handled = true
								if diff < math.MinInt32 || diff > math.MaxInt32 {
									argErr = newOverflowError("integer overflow")
								} else {
									arg = bytecodeBoxedIntegerI32Value(diff)
								}
							}
						}
					case *runtime.IntegerValue:
						if lv != nil && lv.TypeSuffix == runtime.IntegerI32 && lv.IsSmallRef() {
							diff := lv.Int64FastRef() - instr.intImmediateRaw
							handled = true
							if diff < math.MinInt32 || diff > math.MaxInt32 {
								argErr = newOverflowError("integer overflow")
							} else {
								arg = bytecodeBoxedIntegerI32Value(diff)
							}
						}
					}
				} else {
					arg, handled, argErr = bytecodeSelfCallSubtractIntegerImmediateI32Fast(vm.slots[instr.argCount], rightImmediate)
				}
				if !handled {
					arg, handled, argErr = bytecodeSubtractIntegerImmediateFast(vm.slots[instr.argCount], rightImmediate)
				}
				if handled {
					if argErr != nil {
						argErr = vm.interp.wrapStandardRuntimeError(argErr)
						if instr.node != nil {
							argErr = vm.attachBytecodeRuntimeContext(argErr, instr.node, nil)
						}
						return nil, argErr
					}
					hasImplicit := layout.usesImplicitMember
					iterBase := len(vm.iterStack)
					loopBase := len(vm.loopStack)
					returnGenericNames := currentProgram.returnGenericNames
					if !currentProgram.returnGenericNamesCached {
						returnGenericNames = bytecodeInlineReturnGenericNames(fn, currentProgram)
					}
					if layout.slotCount == 2 && layout.selfCallSlot == 1 && instr.argCount == 0 && returnGenericNames == nil && iterBase == 0 && loopBase == 0 && !hasImplicit {
						if vm.pushSelfFastSlot0CallFrame(vm.ip + 1) {
							vm.slots[0] = arg
							vm.setSelfFastSlot0I32Value(arg)
							vm.env = fn.Closure
							vm.ip = 0
							if statsEnabled {
								vm.interp.recordBytecodeInlineCallHit()
							}
							return currentProgram, nil
						}
					}
					var slots []runtime.Value
					if layout.slotCount == 2 {
						slots = vm.acquireSlotFrame2()
					} else {
						slots = vm.acquireSlotFrame(layout.slotCount)
					}
					slots[0] = arg
					if layout.selfCallSlot >= 0 && layout.selfCallSlot < len(slots) {
						slots[layout.selfCallSlot] = fn
					}
					if hasImplicit {
						state := vm.interp.stateFromEnv(fn.Closure)
						state.pushImplicitReceiver(arg)
					}
					if returnGenericNames == nil && iterBase == 0 && loopBase == 0 && !hasImplicit {
						vm.pushSelfFastMinimalCallFrame(vm.ip+1, vm.slots)
					} else {
						vm.pushCallFrame(vm.ip+1, currentProgram, vm.slots, vm.env, returnGenericNames, iterBase, loopBase, hasImplicit, true)
					}
					vm.slots = slots
					vm.env = fn.Closure
					vm.ip = 0
					if statsEnabled {
						vm.interp.recordBytecodeInlineCallHit()
					}
					return currentProgram, nil
				}
			}
		}
	}

	return vm.execCallSelfIntSubSlotConstFallback(instr, rightImmediate, hasImmediate, currentProgram, statsEnabled)
}

func (vm *bytecodeVM) execCallSelfIntSubSlotConstFallback(instr *bytecodeInstruction, rightImmediate runtime.IntegerValue, hasImmediate bool, currentProgram *bytecodeProgram, statsEnabled bool) (*bytecodeProgram, error) {
	callee := vm.slots[instr.target]
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	switch fn := callee.(type) {
	case *runtime.FunctionValue:
		arg, argErr := vm.callSelfIntSubSlotConstArg(instr, rightImmediate, hasImmediate)
		if argErr != nil {
			argErr = vm.interp.wrapStandardRuntimeError(argErr)
			if instr.node != nil {
				argErr = vm.attachBytecodeRuntimeContext(argErr, instr.node, nil)
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
		if result, handled, err := vm.tryExecExactNativeCall(callee, args[:], callNode); handled {
			return vm.finishCompletedCall(result, err, callNode, nil)
		}
		result, err := vm.interp.callCallableValueMutable(callee, args[:], vm.env, callNode)
		return vm.finishCompletedCall(result, err, callNode, nil)
	default:
		if statsEnabled {
			vm.interp.recordBytecodeInlineCallMiss()
		}
	}

	arg, argErr := vm.callSelfIntSubSlotConstArg(instr, rightImmediate, hasImmediate)
	if argErr != nil {
		argErr = vm.interp.wrapStandardRuntimeError(argErr)
		if instr.node != nil {
			argErr = vm.attachBytecodeRuntimeContext(argErr, instr.node, nil)
		}
		return nil, argErr
	}

	args := [1]runtime.Value{arg}
	if result, handled, err := vm.tryExecExactNativeCall(callee, args[:], callNode); handled {
		return vm.finishCompletedCall(result, err, callNode, nil)
	}
	result, err := vm.interp.callCallableValueMutable(callee, args[:], vm.env, callNode)
	return vm.finishCompletedCall(result, err, callNode, nil)
}
