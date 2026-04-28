package interpreter

import (
	"fmt"

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

// execCallSelfIntSubSlotConst handles fused recursive calls of the form
// self(slot - const), computing the argument directly from slot/immediate.
func (vm *bytecodeVM) execCallSelfIntSubSlotConst(instr *bytecodeInstruction, slotConstIntImmTable *bytecodeSlotConstIntImmediateTable, currentProgram *bytecodeProgram) (*bytecodeProgram, error) {
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
					arg, handled, argErr = bytecodeSelfCallSubtractIntegerImmediateI32RawFast(vm.slots[instr.argCount], instr.intImmediateRaw)
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
					hasImplicit := layout.usesImplicitMember
					if hasImplicit {
						state := vm.interp.stateFromEnv(fn.Closure)
						state.pushImplicitReceiver(arg)
					}
					iterBase := len(vm.iterStack)
					loopBase := len(vm.loopStack)
					returnGenericNames := currentProgram.returnGenericNames
					if !currentProgram.returnGenericNamesCached {
						returnGenericNames = bytecodeInlineReturnGenericNames(fn, currentProgram)
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
