package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) finishInlineReturn(program **bytecodeProgram, instructions *[]bytecodeInstruction, validatedIntConsts *[]bool, slotConstIntImmTable **bytecodeSlotConstIntImmediateTable, instr *bytecodeInstruction, val runtime.Value) error {
	activeProgram := *program

	if vm != nil && vm.selfFastMinimalSuffix > 0 {
		if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.returnType != nil {
			noCoercion := false
			if activeProgram.frameLayout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
				noCoercion = inlineCoercionUnnecessaryBySimpleCheck(activeProgram.frameLayout.returnSimpleCheck, val)
			} else if activeProgram.frameLayout.returnSimpleType != "" {
				noCoercion = inlineCoercionUnnecessaryBySimpleType(activeProgram.frameLayout.returnSimpleType, val)
			} else {
				noCoercion = inlineCoercionUnnecessary(activeProgram.frameLayout.returnType, val)
			}
			if !noCoercion {
				coerced, coerceErr := vm.interp.coerceReturnValue(activeProgram.frameLayout.returnType, val, nil, vm.env)
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
		vm.selfFastMinimal = vm.selfFastMinimal[:idx]
		vm.selfFastMinimalSuffix--
		calleeSlots := vm.slots
		vm.ip = returnIP
		vm.slots = returnSlots
		if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.slotCount == 2 {
			vm.releaseSlotFrame2(calleeSlots)
		} else {
			vm.releaseSlotFrame(calleeSlots)
		}
		vm.stack = append(vm.stack, val)
		return nil
	}

	returnGenericNames := vm.peekReturnGenericNames()
	if activeProgram != nil && activeProgram.frameLayout != nil && activeProgram.frameLayout.returnType != nil {
		noCoercion := false
		if activeProgram.frameLayout.returnSimpleCheck != bytecodeSimpleTypeCheckUnknown {
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
	vm.releaseSlotFrame(calleeSlots)
	vm.stack = append(vm.stack, val)
	return nil
}

func (vm *bytecodeVM) execReturnBinaryIntAdd(instr *bytecodeInstruction) (runtime.Value, error) {
	if len(vm.stack) < 2 {
		return nil, fmt.Errorf("bytecode stack underflow")
	}
	rightIdx := len(vm.stack) - 1
	right := vm.stack[rightIdx]
	leftIdx := rightIdx - 1
	left := vm.stack[leftIdx]
	if instr.op == bytecodeOpReturnBinaryIntAddI32 {
		if val, handled, err := bytecodeAddSmallI32PairFast(left, right); handled {
			vm.stack = vm.stack[:leftIdx]
			return val, err
		}
	}
	val, handled, err := vm.execBinarySpecializedOpcode(instr, left, right)
	if !handled && err == nil {
		err = fmt.Errorf("bytecode return-add opcode missing add handler")
	}
	if err != nil {
		return nil, err
	}
	vm.stack = vm.stack[:leftIdx]
	return val, nil
}
