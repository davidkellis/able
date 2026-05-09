package interpreter

import (
	"fmt"
	"math"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) finishInlineReturn(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) error {
	activeProgram := *program

	if vm != nil && vm.selfFastMinimalSuffix > 0 {
		if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.returnType != nil {
			layout := activeProgram.frameLayout
			noCoercion := false
			if knownReturnSimple != bytecodeSimpleTypeCheckUnknown && knownReturnSimple == layout.returnSimpleCheck {
				noCoercion = true
			} else if instr != nil && instr.op == bytecodeOpReturnConstIfIntLessEqualSlotConst && layout.returnSimpleCheck == bytecodeSimpleTypeCheckI32 {
				noCoercion = true
			} else if layout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
				noCoercion = inlineCoercionUnnecessaryBySimpleCheck(layout.returnSimpleCheck, val)
			} else if layout.returnSimpleType != "" {
				noCoercion = inlineCoercionUnnecessaryBySimpleType(layout.returnSimpleType, val)
			} else {
				noCoercion = inlineCoercionUnnecessary(layout.returnType, val)
			}
			if !noCoercion {
				coerced, coerceErr := vm.interp.coerceReturnValue(layout.returnType, val, nil, vm.env)
				if coerceErr != nil {
					if instr != nil && instr.node != nil {
						coerceErr = vm.interp.attachRuntimeContext(coerceErr, instr.node, vm.interp.stateFromEnv(vm.env))
					}
					return coerceErr
				}
				val = coerced
			}
		}

		if len(vm.selfFastMinimal) == 0 {
			return fmt.Errorf("bytecode call frame underflow")
		}
		idx := len(vm.selfFastMinimal) - 1
		frame := &vm.selfFastMinimal[idx]
		returnIP := frame.returnIP
		returnSlots := frame.slots
		reusesSlots := frame.reusesSlots
		returnSlot0 := frame.slot0
		frame.slot0 = nil
		frame.reusesSlots = false
		vm.restoreSelfFastSlot0I32(frame)
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		vm.selfFastMinimalSuffix--
		calleeSlots := vm.slots
		vm.ip = returnIP
		vm.slots = returnSlots
		if reusesSlots {
			if len(returnSlots) > 0 {
				returnSlots[0] = returnSlot0
			}
		} else {
			if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.slotCount == 2 {
				vm.releaseSlotFrame2(calleeSlots)
			} else {
				vm.releaseSlotFrame(calleeSlots)
			}
		}
		vm.stack = append(vm.stack, val)
		return nil
	}

	returnGenericNames := vm.peekReturnGenericNames()
	if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.returnType != nil {
		noCoercion := false
		if knownReturnSimple != bytecodeSimpleTypeCheckUnknown && knownReturnSimple == activeProgram.frameLayout.returnSimpleCheck {
			noCoercion = true
		} else if activeProgram.frameLayout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
			noCoercion = inlineCoercionUnnecessaryBySimpleCheck(activeProgram.frameLayout.returnSimpleCheck, val)
		} else if activeProgram.frameLayout.returnSimpleType != "" {
			noCoercion = inlineCoercionUnnecessaryBySimpleType(activeProgram.frameLayout.returnSimpleType, val)
		} else {
			noCoercion = inlineCoercionUnnecessary(activeProgram.frameLayout.returnType, val)
		}
		if !noCoercion {
			coerced, coerceErr := vm.interp.coerceReturnValue(activeProgram.frameLayout.returnType, val, returnGenericNames, vm.env)
			if coerceErr != nil {
				if instr != nil && instr.node != nil {
					coerceErr = vm.interp.attachRuntimeContext(coerceErr, instr.node, vm.interp.stateFromEnv(vm.env))
				}
				return coerceErr
			}
			val = coerced
		}
	}

	returnIP, returnProgram, returnSlots, returnEnv, iterBase, loopBase, hasImplicitReceiver, selfFast, ok := vm.popCallFrameFields()
	if !ok {
		return fmt.Errorf("bytecode call frame underflow")
	}
	calleeSlots := vm.slots
	if hasImplicitReceiver {
		state := vm.interp.stateFromEnv(vm.env)
		state.popImplicitReceiver()
	}
	vm.ip = returnIP
	vm.slots = returnSlots
	if !selfFast {
		vm.env = returnEnv
		vm.switchRunProgram(program, instructions, validatedIntConsts, slotConstIntImmTable, returnProgram)
	}
	if len(vm.iterStack) > iterBase {
		for idx := len(vm.iterStack) - 1; idx >= iterBase; idx-- {
			if iter := vm.iterStack[idx].iter; iter != nil {
				iter.Close()
			}
		}
		vm.iterStack = vm.iterStack[:iterBase]
	}
	if len(vm.loopStack) > loopBase {
		vm.loopStack = vm.loopStack[:loopBase]
	}
	if !sameSlotFrame(calleeSlots, returnSlots) {
		vm.releaseSlotFrame(calleeSlots)
	}
	vm.stack = append(vm.stack, val)
	return nil
}

func bytecodeCanFinishMinimalReturnNoCoerce(program *bytecodeProgram, instr *bytecodeInstruction, knownReturnSimple bytecodeSimpleTypeCheck) bool {
	if program == nil || instr == nil {
		return false
	}
	layout := program.frameLayout
	if layout == nil || layout.returnSimpleCheck != bytecodeSimpleTypeCheckI32 {
		return false
	}
	if knownReturnSimple == bytecodeSimpleTypeCheckI32 {
		return true
	}
	switch instr.op {
	case bytecodeOpReturnConstIfIntLessEqualSlotConst:
		return true
	case bytecodeOpReturnIfIntLessEqualSlotConst:
		return instr.target >= 0 && instr.target < len(layout.slotKinds) && layout.slotKinds[instr.target] == bytecodeCellKindI32
	default:
		return false
	}
}

func (vm *bytecodeVM) finishMinimalSelfFastReturnNoCoerce(val runtime.Value) bool {
	if vm == nil || vm.selfFastMinimalSuffix <= 0 || len(vm.selfFastMinimal) == 0 {
		return false
	}
	idx := len(vm.selfFastMinimal) - 1
	frame := &vm.selfFastMinimal[idx]
	if !frame.reusesSlots || len(frame.slots) == 0 {
		return false
	}
	returnIP := frame.returnIP
	returnSlots := frame.slots
	returnSlot0 := frame.slot0
	frame.slot0 = nil
	frame.reusesSlots = false
	vm.restoreSelfFastSlot0I32(frame)
	vm.selfFastMinimal = vm.selfFastMinimal[:idx]
	vm.selfFastMinimalSuffix--
	vm.ip = returnIP
	vm.slots = returnSlots
	returnSlots[0] = returnSlot0
	vm.stack = append(vm.stack, val)
	return true
}

func (vm *bytecodeVM) tryFinishMinimalSelfFastReturnNoCoerce(program *bytecodeProgram, instr *bytecodeInstruction, val runtime.Value, knownReturnSimple bytecodeSimpleTypeCheck) bool {
	if !bytecodeCanFinishMinimalReturnNoCoerce(program, instr, knownReturnSimple) {
		return false
	}
	return vm.finishMinimalSelfFastReturnNoCoerce(val)
}

func (vm *bytecodeVM) execReturnBinaryIntAdd(instr *bytecodeInstruction) (runtime.Value, bytecodeSimpleTypeCheck, error) {
	if len(vm.stack) < 2 {
		return nil, bytecodeSimpleTypeCheckUnknown, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := len(vm.stack) - 1
	leftIdx := rightIdx - 1
	right := vm.stack[rightIdx]
	left := vm.stack[leftIdx]
	if instr.op == bytecodeOpReturnBinaryIntAddI32 {
		if lv, ok := left.(runtime.IntegerValue); ok && lv.TypeSuffix == runtime.IntegerI32 {
			if rv, ok := right.(runtime.IntegerValue); ok && rv.TypeSuffix == runtime.IntegerI32 {
				lvRef := &lv
				rvRef := &rv
				if lvRef.IsSmallRef() && rvRef.IsSmallRef() {
					l := lvRef.Int64FastRef()
					r := rvRef.Int64FastRef()
					if l >= math.MinInt32 && l <= math.MaxInt32 && r >= math.MinInt32 && r <= math.MaxInt32 {
						sum := l + r
						vm.stack = vm.stack[:leftIdx]
						if sum < math.MinInt32 || sum > math.MaxInt32 {
							return nil, bytecodeSimpleTypeCheckI32, newOverflowError("integer overflow")
						}
						return bytecodeBoxedIntegerI32Value(sum), bytecodeSimpleTypeCheckI32, nil
					}
				}
			}
		}
		if val, handled, err := bytecodeAddSmallI32PairFast(left, right); handled {
			vm.stack = vm.stack[:leftIdx]
			return val, bytecodeSimpleTypeCheckI32, err
		}
	}
	val, handled, err := vm.execBinarySpecializedOpcode(instr, left, right)
	if !handled && err == nil {
		err = fmt.Errorf("bytecode return-add opcode missing add handler")
	}
	if err != nil {
		return nil, bytecodeSimpleTypeCheckUnknown, err
	}
	vm.stack = vm.stack[:leftIdx]
	return val, bytecodeSimpleTypeCheckUnknown, nil
}
