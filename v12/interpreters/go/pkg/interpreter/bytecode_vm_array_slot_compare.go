package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (vm *bytecodeVM) execJumpIfArrayReadSlotCompareSlotFalse(instr *bytecodeInstruction, program *bytecodeProgram) error {
	if instr == nil {
		return fmt.Errorf("bytecode array-slot compare jump missing instruction")
	}
	receiverSlot, indexSlot, rightSlot := instr.argCount, instr.loopBreak, instr.loopContinue
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) ||
		rightSlot < 0 || rightSlot >= len(vm.slots) {
		return fmt.Errorf("bytecode array-slot compare slot out of range")
	}
	receiver := vm.slots[receiverSlot]
	index := vm.slots[indexSlot]
	right := vm.slots[rightSlot]
	left, err := vm.arrayReadSlotCompareValue(instr, program, receiver, index)
	if err != nil {
		return err
	}
	cond, err := vm.compareArrayReadSlotCondition(instr.operator, left, right)
	if err != nil {
		return err
	}
	if !cond {
		vm.ip = instr.target
		return nil
	}
	vm.ip++
	return nil
}

func (vm *bytecodeVM) arrayReadSlotCompareValue(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	return vm.arrayReadSlotValue(instr, program, receiver, index)
}

func (vm *bytecodeVM) arrayReadSlotValue(instr *bytecodeInstruction, program *bytecodeProgram, receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if arr, ok := receiver.(*runtime.ArrayValue); ok && arr != nil && vm.canUseCanonicalArraySlotCallCacheForArray(arr) {
		if vm.lookupCachedCanonicalArraySlotCallForArray(program, vm.ip, bytecodeMemberMethodFastPathArrayReadSlot) {
			if value, mode, handled, err := vm.readArraySlotValueFast(arr, index); handled || err != nil {
				if err == nil && vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", "read_slot", "resolved_method", mode, instr.node)
				}
				return value, err
			}
		} else if ok, err := vm.proveCanonicalArrayReadSlotCall(program, vm.ip, instr, receiver); err != nil {
			return nil, err
		} else if ok {
			if value, mode, handled, err := vm.readArraySlotValueFast(arr, index); handled || err != nil {
				if err == nil && vm.interp != nil {
					vm.interp.recordBytecodeCallTrace("call_member", "read_slot", "resolved_method", mode, instr.node)
				}
				return value, err
			}
		}
	}
	return vm.genericArrayReadSlotCompareValue(receiver, index)
}

func (vm *bytecodeVM) execArrayReadSlot(instr *bytecodeInstruction, program *bytecodeProgram) (*bytecodeProgram, error) {
	if instr == nil {
		return nil, fmt.Errorf("bytecode array read_slot missing instruction")
	}
	receiverSlot, indexSlot := instr.argCount, instr.loopBreak
	if receiverSlot < 0 || receiverSlot >= len(vm.slots) ||
		indexSlot < 0 || indexSlot >= len(vm.slots) {
		return nil, fmt.Errorf("bytecode array read_slot slot out of range")
	}
	result, err := vm.arrayReadSlotValue(instr, program, vm.slots[receiverSlot], vm.slots[indexSlot])
	var callNode *ast.FunctionCall
	if instr.node != nil {
		if call, ok := instr.node.(*ast.FunctionCall); ok {
			callNode = call
		}
	}
	return vm.finishCompletedCall(result, err, callNode, nil)
}

func (vm *bytecodeVM) proveCanonicalArrayReadSlotCall(program *bytecodeProgram, ip int, instr *bytecodeInstruction, receiver runtime.Value) (bool, error) {
	if vm == nil || vm.interp == nil || instr == nil {
		return false, nil
	}
	callable, found, err := vm.interp.resolveMethodCallableFromPool(vm.env, "read_slot", receiver, "")
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	fn, ok := bytecodeResolvedMemberFastPathFunction(callable)
	if !ok {
		return false, nil
	}
	kind := vm.resolvedMemberMethodFastPath("read_slot", receiver, fn)
	if kind != bytecodeMemberMethodFastPathArrayReadSlot {
		return false, nil
	}
	vm.storeCachedCanonicalArraySlotCall(program, ip, bytecodeInstruction{name: "read_slot", argCount: 1}, receiver, kind)
	return true, nil
}

func (vm *bytecodeVM) genericArrayReadSlotCompareValue(receiver runtime.Value, index runtime.Value) (runtime.Value, error) {
	if vm == nil || vm.interp == nil {
		return nil, fmt.Errorf("bytecode VM is nil")
	}
	callee, err := vm.interp.memberAccessOnValueWithOptions(receiver, ast.ID("read_slot"), vm.env, true)
	if err != nil {
		return nil, err
	}
	args := [1]runtime.Value{index}
	value, err := vm.interp.callCallableValueMutable(callee, args[:], vm.env, nil)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return runtime.NilValue{}, nil
	}
	return value, nil
}

func (vm *bytecodeVM) compareArrayReadSlotCondition(op string, left runtime.Value, right runtime.Value) (bool, error) {
	return vm.compareBytecodeCondition(op, left, right)
}
